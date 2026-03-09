<template>
  <div class="qqbot-page">
    <OpenclawSaveActions
      :title="t('qqbot.pageTitle')"
      :loading="loading"
      :saving="saving"
      :can-edit="canEdit"
      :show-preview="hasBotRows"
      :show-save="hasBotRows"
      @refresh="loadConfig"
      @preview="previewDiff"
      @save="saveConfig"
    />

    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>{{ t('qqbot.listTitle') }}</span>
          <el-tag type="info">{{ t('qqbot.totalCount', { count: bots.length }) }}</el-tag>
        </div>
      </template>

      <el-empty v-if="!loading && bots.length === 0" :description="t('qqbot.empty')" />

      <el-table v-else :data="bots" row-key="key" style="width: 100%">
        <el-table-column prop="name" :label="t('qqbot.botName')" min-width="140" />
        <el-table-column :label="t('qqbot.appId')" min-width="220">
          <template #default="{ row }">
            <el-input v-model="row.appId" :disabled="!canEdit" :placeholder="t('qqbot.firstForm.appIdPlaceholder')" />
          </template>
        </el-table-column>
        <el-table-column :label="t('qqbot.appSecret')" min-width="260">
          <template #default="{ row }">
            <el-input
              v-model="row.clientSecret"
              :disabled="!canEdit"
              :placeholder="t('qqbot.firstForm.appSecretPlaceholder')"
              show-password
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('qqbot.operations')" width="120">
          <template #default="{ row }">
            <el-button type="danger" link :disabled="!canEdit" @click="removeBot(row)">
              {{ t('common.actions.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-alert
        v-if="bots.length === 0"
        class="tips"
        type="info"
        :closable="false"
        show-icon
        :title="t('qqbot.firstAccessTips')"
      >
      </el-alert>

      <div v-if="bots.length === 0" class="first-access">
        <el-form label-position="top">
          <el-row :gutter="12">
            <el-col :xs="24" :md="12">
              <el-form-item :label="t('qqbot.firstForm.appIdLabel')">
                <el-input
                  v-model="firstBot.appId"
                  :disabled="!canEdit"
                  :placeholder="t('qqbot.firstForm.appIdPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :xs="24" :md="12">
              <el-form-item :label="t('qqbot.firstForm.appSecretLabel')">
                <el-input
                  v-model="firstBot.clientSecret"
                  :disabled="!canEdit"
                  show-password
                  :placeholder="t('qqbot.firstForm.appSecretPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>

        <div class="cmd-box">
          <div class="cmd-title-row">
            <div class="cmd-title">{{ t('qqbot.firstAccessCommandTitle') }}</div>
            <el-button type="primary" :disabled="!canEdit || !canExecuteFirstAccess" @click="goExecuteFirstAccessCommands">
              {{ t('common.actions.goExecute') }}
            </el-button>
          </div>
          <pre>{{ firstAccessCommand }}</pre>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" v-if="bots.length !== 0">
      <template #header>{{ t('qqbot.addTitle') }}</template>

      <el-alert
        type="warning"
        show-icon
        :closable="false"
        :title="t('qqbot.addTips')"
      />

      <div class="add-list">
        <div v-for="(item, idx) in addRows" :key="item.id" class="add-row">
          <el-row :gutter="10">
            <el-col :xs="24" :md="6">
              <el-input
                v-model="item.name"
                :disabled="!canEdit"
                :placeholder="t('qqbot.addForm.botNamePlaceholder')"
              />
            </el-col>
            <el-col :xs="24" :md="8">
              <el-input v-model="item.appId" :disabled="!canEdit" :placeholder="t('qqbot.addForm.appIdPlaceholder')" />
            </el-col>
            <el-col :xs="24" :md="8">
              <el-input
                v-model="item.clientSecret"
                :disabled="!canEdit"
                :placeholder="t('qqbot.addForm.appSecretPlaceholder')"
                show-password
              />
            </el-col>
            <el-col :xs="24" :md="2" class="row-op">
              <el-button type="danger" :disabled="!canEdit" @click="removeAddRow(idx)">
                {{ t('common.actions.delete') }}
              </el-button>
            </el-col>
          </el-row>
        </div>
      </div>

      <el-space>
        <el-button :disabled="!canEdit" @click="addRow">{{ t('common.actions.addRow') }}</el-button>
        <el-button type="primary" :disabled="!canEdit" @click="appendRowsToConfig">
          {{ t('common.actions.appendDraft') }}
        </el-button>
      </el-space>
    </el-card>

    <el-dialog v-model="diffDialogVisible" width="1080px" :title="t('qqbot.diffPreviewTitle')">
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
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import DiffViewer from '../components/DiffViewer.vue'
import OpenclawSaveActions from '../components/OpenclawSaveActions.vue'
import { buildOpenclawDiff, getOpenclawConfig, saveOpenclawConfig } from '../services/openclawConfig'

type BotRow = {
  key: string
  name: string
  isPrimary: boolean
  appId: string
  clientSecret: string
}

type AddRow = {
  id: string
  name: string
  appId: string
  clientSecret: string
}

type ShellStashItem = {
  id: string
  command: string
}

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const SHELL_STASH_STORAGE_KEY = 'openclaw_manager_shell_stash_v1'

const rawConfig = ref<any>({})
const originalConfigText = ref('{}')
const bots = ref<BotRow[]>([])

function newRowID(): string {
  const c = (globalThis as any)?.crypto
  if (c && typeof c.randomUUID === 'function') return c.randomUUID()
  return `row_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

const firstBot = ref({ appId: '', clientSecret: '' })
const addRows = ref<AddRow[]>([{ id: newRowID(), name: '', appId: '', clientSecret: '' }])

const diffDialogVisible = ref(false)
const diffFromText = ref('')
const diffToText = ref('')

const canEdit = computed(() => {
  const role = auth.user?.role || 'Viewer'
  return role === 'Operator' || role === 'Admin'
})

const hasBotRows = computed(() => bots.value.length > 0)

const firstAccessCommand = computed(() => {
  const appId = firstBot.value.appId || '[yourAppId]'
  const clientSecret = firstBot.value.clientSecret || '[yourAppSecret]'

  return `openclaw plugins install @sliverp/qqbot@latest
openclaw channels add --channel qqbot --token "${appId}:${clientSecret}"
openclaw gateway restart`
})

const canExecuteFirstAccess = computed(() => {
  return Boolean(firstBot.value.appId.trim() && firstBot.value.clientSecret.trim())
})

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function parseLines(raw: string): string[] {
  return raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
}

function listBotsFromConfig(cfg: any): BotRow[] {
  const out: BotRow[] = []
  const qqbot = cfg?.channels?.qqbot
  if (!qqbot || typeof qqbot !== 'object') return out

  out.push({
    key: 'primary',
    name: 'qqbot',
    isPrimary: true,
    appId: String(qqbot.appId || ''),
    clientSecret: String(qqbot.clientSecret || ''),
  })

  const accounts = qqbot.accounts
  if (accounts && typeof accounts === 'object') {
    Object.keys(accounts).forEach((name) => {
      const item = accounts[name] || {}
      out.push({
        key: `acc:${name}`,
        name,
        isPrimary: false,
        appId: String(item.appId || ''),
        clientSecret: String(item.clientSecret || ''),
      })
    })
  }
  return out
}

function ensureQQBotRoot(cfg: any) {
  if (!cfg.channels || typeof cfg.channels !== 'object') cfg.channels = {}
  if (!cfg.channels.qqbot || typeof cfg.channels.qqbot !== 'object') {
    cfg.channels.qqbot = {
      enabled: true,
      allowFrom: ['*'],
      appId: '',
      clientSecret: '',
    }
  }
}

async function loadConfig() {
  loading.value = true
  errorMessage.value = ''
  try {
    const payload = await getOpenclawConfig()
    const text = String(payload.content || '{}')
    originalConfigText.value = text
    rawConfig.value = JSON.parse(text)
    bots.value = listBotsFromConfig(rawConfig.value)
  } catch (err) {
    errorMessage.value = parseError(err, t('qqbot.messages.loadConfigFailed'))
  } finally {
    loading.value = false
  }
}

function removeBot(row: BotRow) {
  if (row.isPrimary) {
    const qqbot = rawConfig.value?.channels?.qqbot
    if (qqbot && typeof qqbot === 'object') {
      delete rawConfig.value.channels.qqbot
    }
  } else {
    const accounts = rawConfig.value?.channels?.qqbot?.accounts
    if (accounts && typeof accounts === 'object') {
      delete accounts[row.name]
    }
  }
  bots.value = listBotsFromConfig(rawConfig.value)
}

function addRow() {
  addRows.value.push({ id: newRowID(), name: '', appId: '', clientSecret: '' })
}

function removeAddRow(idx: number) {
  addRows.value.splice(idx, 1)
  if (addRows.value.length === 0) addRow()
}

function ensureValidBotName(name: string): boolean {
  return /^[A-Za-z0-9]+$/.test(name)
}

function appendRowsToConfig() {
  if (!canEdit.value) return

  if (bots.value.length === 0) {
    if (!firstBot.value.appId.trim() || !firstBot.value.clientSecret.trim()) {
      ElMessage.warning(t('qqbot.messages.needPrimaryCredentials'))
      return
    }
    ensureQQBotRoot(rawConfig.value)
    rawConfig.value.channels.qqbot.enabled = true
    rawConfig.value.channels.qqbot.allowFrom = Array.isArray(rawConfig.value.channels.qqbot.allowFrom)
      ? rawConfig.value.channels.qqbot.allowFrom
      : ['*']
    rawConfig.value.channels.qqbot.appId = firstBot.value.appId.trim()
    rawConfig.value.channels.qqbot.clientSecret = firstBot.value.clientSecret.trim()
    bots.value = listBotsFromConfig(rawConfig.value)
    ElMessage.success(t('qqbot.messages.primaryAdded'))
    return
  }

  ensureQQBotRoot(rawConfig.value)
  const qqbot = rawConfig.value.channels.qqbot
  if (!qqbot.accounts || typeof qqbot.accounts !== 'object') qqbot.accounts = {}

  let added = 0
  for (const row of addRows.value) {
    const name = row.name.trim()
    const appId = row.appId.trim()
    const clientSecret = row.clientSecret.trim()
    if (!name && !appId && !clientSecret) continue

    if (!ensureValidBotName(name)) {
      ElMessage.error(t('qqbot.messages.invalidBotName', { name: name || t('qqbot.messages.emptyName') }))
      return
    }
    if (!appId || !clientSecret) {
      ElMessage.error(t('qqbot.messages.missingBotCredentials', { name }))
      return
    }

    qqbot.accounts[name] = {
      enabled: true,
      appId,
      clientSecret,
      allowFrom: ['*'],
    }
    added += 1
  }

  if (added === 0) {
    ElMessage.warning(t('qqbot.messages.noRowsToAdd'))
    return
  }

  bots.value = listBotsFromConfig(rawConfig.value)
  addRows.value = [{ id: newRowID(), name: '', appId: '', clientSecret: '' }]
  ElMessage.success(t('qqbot.messages.addedBots', { count: added }))
}

function buildNormalizedConfigText(): string {
  // 同步页面已编辑的 bot 列表到配置对象
  if (bots.value.length > 0) {
    ensureQQBotRoot(rawConfig.value)
    const qqbot = rawConfig.value.channels.qqbot
    const primary = bots.value.find((x) => x.isPrimary)
    if (primary) {
      qqbot.enabled = true
      qqbot.allowFrom = Array.isArray(qqbot.allowFrom) ? qqbot.allowFrom : ['*']
      qqbot.appId = primary.appId.trim()
      qqbot.clientSecret = primary.clientSecret.trim()
    }

    const accounts: Record<string, any> = {}
    bots.value.filter((x) => !x.isPrimary).forEach((row) => {
      accounts[row.name] = {
        enabled: true,
        appId: row.appId.trim(),
        clientSecret: row.clientSecret.trim(),
        allowFrom: ['*'],
      }
    })
    if (Object.keys(accounts).length > 0) qqbot.accounts = accounts
    else delete qqbot.accounts
  }

  return JSON.stringify(rawConfig.value, null, 2)
}

function goExecuteFirstAccessCommands() {
  const appId = firstBot.value.appId.trim()
  const clientSecret = firstBot.value.clientSecret.trim()
  if (!appId && !clientSecret) {
    ElMessage.warning(t('qqbot.messages.needBothCredentials'))
    return
  }
  if (!appId) {
    ElMessage.warning(t('qqbot.messages.needAppId'))
    return
  }
  if (!clientSecret) {
    ElMessage.warning(t('qqbot.messages.needAppSecret'))
    return
  }

  const commands = parseLines(firstAccessCommand.value)
  if (commands.length === 0) {
    ElMessage.warning(t('qqbot.messages.noCommands'))
    return
  }

  try {
    const raw = localStorage.getItem(SHELL_STASH_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    const stash: ShellStashItem[] = Array.isArray(parsed)
      ? parsed
          .map((item: any) => ({
            id: String(item?.id || newRowID()),
            command: String(item?.command || '').trim(),
          }))
          .filter((item) => item.command)
      : []

    const existing = new Set(stash.map((item) => item.command))
    let added = 0
    commands.forEach((command) => {
      if (existing.has(command)) return
      stash.push({ id: newRowID(), command })
      existing.add(command)
      added += 1
    })

    localStorage.setItem(SHELL_STASH_STORAGE_KEY, JSON.stringify(stash))
    if (added > 0) {
      ElMessage.success(t('qqbot.messages.addedCommandsAndRedirect', { count: added }))
    } else {
      ElMessage.info(t('qqbot.messages.commandsAlreadyExist'))
    }
    router.push('/shell')
  } catch {
    ElMessage.error(t('qqbot.messages.writeStashFailed'))
  }
}

function previewDiff() {
  try {
    const diff = buildOpenclawDiff(originalConfigText.value, buildNormalizedConfigText())
    diffFromText.value = diff.fromText
    diffToText.value = diff.toText
    diffDialogVisible.value = true
  } catch {
    ElMessage.error(t('qqbot.messages.buildPreviewFailed'))
  }
}

async function saveConfig() {
  if (!canEdit.value) {
    ElMessage.warning(t('qqbot.messages.noEditPermission'))
    return
  }

  let normalized = ''
  try {
    normalized = buildNormalizedConfigText()
  } catch {
    ElMessage.error(t('qqbot.messages.serializeFailed'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('qqbot.messages.saveConfirmContent'),
      t('qqbot.messages.saveConfirmTitle'),
      { type: 'warning' },
    )
  } catch {
    return
  }

  saving.value = true
  try {
    await saveOpenclawConfig(normalized)
    ElMessage.success(t('qqbot.messages.saveSuccess'))
    await loadConfig()
  } catch (err) {
    ElMessage.error(parseError(err, t('qqbot.messages.saveFailed')))
  } finally {
    saving.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.qqbot-page {
  display: grid;
  gap: 12px;
}
.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.tips {
  margin-top: 10px;
}
.first-access {
  margin-top: 12px;
}
.cmd-box {
  border: 1px dashed var(--oc-border-strong);
  border-radius: 8px;
  padding: 10px;
  background: var(--oc-surface-muted);
}
.cmd-title {
  font-weight: 600;
  color: var(--oc-text);
}
.cmd-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}
.cmd-box pre {
  margin: 0;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
}
.add-list {
  margin: 10px 0;
  display: grid;
  gap: 8px;
}
.add-row {
  border: 1px solid var(--oc-border);
  border-radius: 8px;
  padding: 8px;
  background: var(--oc-surface);
}
.row-op {
  display: flex;
  justify-content: flex-end;
}
</style>
