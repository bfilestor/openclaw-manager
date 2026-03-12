<template>
  <div class="api-provider-page">
    <OpenclawSaveActions
      :title="t('apiProviders.pageTitle')"
      :loading="loading"
      :saving="saving"
      :can-edit="canEdit"
      :show-preview="providers.length > 0"
      :show-save="providers.length > 0"
      @refresh="loadConfig"
      @preview="previewDiff"
      @save="saveConfig"
    />

    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>{{ t('apiProviders.listTitle') }}</span>
          <el-tag type="info">{{ t('apiProviders.totalCount', { count: providers.length }) }}</el-tag>
        </div>
      </template>

      <el-empty v-if="!loading && providers.length === 0" :description="t('apiProviders.empty')" />

      <el-table v-else :data="providers" row-key="id" style="width: 100%">
        <el-table-column prop="id" :label="t('apiProviders.columns.provider')" min-width="140" />
        <el-table-column :label="t('apiProviders.columns.baseUrl')" min-width="240">
          <template #default="{ row }">
            <el-input v-model="row.baseUrl" :disabled="!canEdit" :placeholder="t('apiProviders.placeholders.baseUrl')" />
          </template>
        </el-table-column>
        <el-table-column :label="t('apiProviders.columns.key')" min-width="240">
          <template #default="{ row }">
            <el-input
              v-model="row.apiKey"
              :disabled="!canEdit"
              :placeholder="t('apiProviders.placeholders.key')"
              show-password
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('apiProviders.columns.accessMethod')" min-width="180">
          <template #default="{ row }">
            <el-input
              v-model="row.accessMethod"
              :disabled="!canEdit"
              :placeholder="t('apiProviders.placeholders.accessMethod')"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('apiProviders.columns.costPer1k')" min-width="160">
          <template #default="{ row }">
            <el-input-number
              v-model="row.costPer1k"
              :disabled="!canEdit"
              :placeholder="t('apiProviders.placeholders.costPer1k')"
              :min="0"
              :precision="6"
              :step="0.001"
              controls-position="right"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('apiProviders.columns.actions')" width="120">
          <template #default="{ row }">
            <el-button type="danger" link :disabled="!canEdit" @click="removeProvider(row.id)">
              {{ t('common.actions.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card shadow="never">
      <template #header>{{ t('apiProviders.addTitle') }}</template>
      <el-row :gutter="10" class="add-row">
        <el-col :xs="24" :md="4">
          <el-input v-model="newProvider.id" :disabled="!canEdit" :placeholder="t('apiProviders.placeholders.provider')" />
        </el-col>
        <el-col :xs="24" :md="5">
          <el-input v-model="newProvider.baseUrl" :disabled="!canEdit" :placeholder="t('apiProviders.placeholders.baseUrl')" />
        </el-col>
        <el-col :xs="24" :md="5">
          <el-input
            v-model="newProvider.apiKey"
            :disabled="!canEdit"
            :placeholder="t('apiProviders.placeholders.key')"
            show-password
          />
        </el-col>
        <el-col :xs="24" :md="4">
          <el-input
            v-model="newProvider.accessMethod"
            :disabled="!canEdit"
            :placeholder="t('apiProviders.placeholders.accessMethod')"
          />
        </el-col>
        <el-col :xs="24" :md="4">
          <el-input-number
            v-model="newProvider.costPer1k"
            :disabled="!canEdit"
            :placeholder="t('apiProviders.placeholders.costPer1k')"
            :min="0"
            :precision="6"
            :step="0.001"
            controls-position="right"
          />
        </el-col>
        <el-col :xs="24" :md="2" class="row-op">
          <el-button type="success" :disabled="!canEdit" @click="addProvider">{{ t('common.actions.addRow') }}</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-dialog v-model="diffDialogVisible" width="1080px" :title="t('apiProviders.diffPreviewTitle')">
      <DiffViewer :from-text="diffFromText" :to-text="diffToText" :height="460" />
      <template #footer>
        <el-button @click="diffDialogVisible = false">{{ t('common.actions.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import DiffViewer from '../components/DiffViewer.vue'
import OpenclawSaveActions from '../components/OpenclawSaveActions.vue'
import { buildOpenclawDiff, getOpenclawConfig, saveOpenclawConfig } from '../services/openclawConfig'

type ProviderRow = {
  id: string
  baseUrl: string
  apiKey: string
  accessMethod: string
  costPer1k: number | null
}

const auth = useAuthStore()
const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const rawConfig = ref<any>({})
const originalConfigText = ref('{}')
const providers = ref<ProviderRow[]>([])
const newProvider = ref<ProviderRow>({ id: '', baseUrl: '', apiKey: '', accessMethod: '', costPer1k: null })

const diffDialogVisible = ref(false)
const diffFromText = ref('')
const diffToText = ref('')

const canEdit = computed(() => {
  const role = auth.user?.role || 'Viewer'
  return role === 'Operator' || role === 'Admin'
})

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function parseCostPer1k(provider: any): number | null {
  const raw = provider?.xManager?.costPer1kToken
  if (typeof raw !== 'number' || Number.isNaN(raw) || raw < 0) return null
  return raw
}

function listProvidersFromConfig(cfg: any): ProviderRow[] {
  const entries = cfg?.models?.providers
  if (!entries || typeof entries !== 'object') return []
  return Object.entries(entries).map(([id, provider]: [string, any]) => ({
    id,
    baseUrl: String(provider?.baseUrl || provider?.baseURL || ''),
    apiKey: String(provider?.apiKey || ''),
    accessMethod: String(provider?.api || ''),
    costPer1k: parseCostPer1k(provider),
  }))
}

function ensureProviderRoot(cfg: any) {
  if (!cfg.models || typeof cfg.models !== 'object') cfg.models = {}
  if (!cfg.models.providers || typeof cfg.models.providers !== 'object') cfg.models.providers = {}
}

function resetNewProvider() {
  newProvider.value = { id: '', baseUrl: '', apiKey: '', accessMethod: '', costPer1k: null }
}

async function loadConfig() {
  loading.value = true
  errorMessage.value = ''
  try {
    const payload = await getOpenclawConfig()
    const text = String(payload.content || '{}')
    originalConfigText.value = text
    rawConfig.value = JSON.parse(text)
    providers.value = listProvidersFromConfig(rawConfig.value)
  } catch (err) {
    errorMessage.value = parseError(err, t('apiProviders.messages.loadConfigFailed'))
  } finally {
    loading.value = false
  }
}

function removeProvider(providerID: string) {
  if (!rawConfig.value?.models?.providers || typeof rawConfig.value.models.providers !== 'object') return
  delete rawConfig.value.models.providers[providerID]
  providers.value = listProvidersFromConfig(rawConfig.value)
}

function addProvider() {
  if (!canEdit.value) return
  const id = newProvider.value.id.trim()
  if (!id) {
    ElMessage.warning(t('apiProviders.messages.needProviderName'))
    return
  }
  if (!/^[A-Za-z0-9_-]+$/.test(id)) {
    ElMessage.error(t('apiProviders.messages.invalidProviderName'))
    return
  }
  if (providers.value.some((item) => item.id === id)) {
    ElMessage.error(t('apiProviders.messages.providerExists', { id }))
    return
  }

  ensureProviderRoot(rawConfig.value)
  rawConfig.value.models.providers[id] = {
    baseUrl: newProvider.value.baseUrl.trim(),
    apiKey: newProvider.value.apiKey.trim(),
    api: newProvider.value.accessMethod.trim(),
  }
  if (typeof newProvider.value.costPer1k === 'number' && newProvider.value.costPer1k >= 0) {
    rawConfig.value.models.providers[id].xManager = { costPer1kToken: newProvider.value.costPer1k }
  }

  providers.value = listProvidersFromConfig(rawConfig.value)
  resetNewProvider()
  ElMessage.success(t('apiProviders.messages.providerAdded'))
}

function buildNormalizedConfigText(): string {
  ensureProviderRoot(rawConfig.value)

  const nextProviders: Record<string, any> = {}
  for (const row of providers.value) {
    const current = rawConfig.value.models.providers[row.id] || {}
    current.baseUrl = row.baseUrl.trim()
    current.apiKey = row.apiKey.trim()
    current.api = row.accessMethod.trim()

    if (typeof row.costPer1k === 'number' && row.costPer1k >= 0) {
      if (!current.xManager || typeof current.xManager !== 'object') current.xManager = {}
      current.xManager.costPer1kToken = row.costPer1k
    } else if (current.xManager && typeof current.xManager === 'object') {
      delete current.xManager.costPer1kToken
      if (Object.keys(current.xManager).length === 0) delete current.xManager
    }

    nextProviders[row.id] = current
  }

  rawConfig.value.models.providers = nextProviders
  return JSON.stringify(rawConfig.value, null, 2)
}

function previewDiff() {
  try {
    const diff = buildOpenclawDiff(originalConfigText.value, buildNormalizedConfigText())
    diffFromText.value = diff.fromText
    diffToText.value = diff.toText
    diffDialogVisible.value = true
  } catch {
    ElMessage.error(t('apiProviders.messages.buildPreviewFailed'))
  }
}

async function saveConfig() {
  if (!canEdit.value) {
    ElMessage.warning(t('apiProviders.messages.noEditPermission'))
    return
  }

  let normalized = ''
  try {
    normalized = buildNormalizedConfigText()
  } catch {
    ElMessage.error(t('apiProviders.messages.serializeFailed'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('apiProviders.messages.saveConfirmContent'),
      t('apiProviders.messages.saveConfirmTitle'),
      { type: 'warning' },
    )
  } catch {
    return
  }

  saving.value = true
  try {
    await saveOpenclawConfig(normalized)
    ElMessage.success(t('apiProviders.messages.saveSuccess'))
    await loadConfig()
  } catch (err) {
    ElMessage.error(parseError(err, t('apiProviders.messages.saveFailed')))
  } finally {
    saving.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.api-provider-page {
  display: grid;
  gap: 12px;
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.add-row {
  align-items: center;
}

.row-op {
  display: flex;
  justify-content: flex-end;
}
</style>
