<template>
  <div class="skills-page">
    <div class="topbar">
      <h3>{{ t('skills.title') }}</h3>
      <el-space>
        <el-button :loading="loading" @click="loadSkills">{{ t('common.actions.refresh') }}</el-button>
      </el-space>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12" class="stats-row">
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">{{ t('skills.installedCount', { count: skills.length }) }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">{{ t('skills.totalSize', { size: formatBytes(totalBytes) }) }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">{{ t('skills.currentView', { scope: scopeLabel }) }}</el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="display: none;">
      <template #header>{{ t('skills.installTitle') }}</template>
      <el-form label-position="top">
        <el-row :gutter="12">
          <el-col :xs="24" :sm="8">
            <el-form-item :label="t('skills.installScope')">
              <el-select v-model="installForm.scope" style="width: 100%">
                <el-option :label="t('skills.scopeGlobal')" value="global" />
                <el-option :label="t('skills.scopeAgent')" value="agent" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col v-if="installForm.scope === 'agent'" :xs="24" :sm="8">
            <el-form-item :label="t('skills.columns.agentId')">
              <el-select v-model="installForm.agent_id" filterable clearable style="width: 100%">
                <el-option v-for="a in agents" :key="a.agent_id" :label="a.agent_id" :value="a.agent_id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-form-item :label="t('skills.skillNameOptional')">
              <el-input v-model="installForm.skill_name" :placeholder="t('skills.skillNamePlaceholder')" clearable />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="t('skills.uploadPackage')">
          <input ref="fileInputRef" type="file" accept=".zip,.tar.gz" @change="onFileChange" />
          <el-text type="info" class="file-hint">{{ t('skills.currentFile', { fileName: selectedFileName || t('skills.notSelected') }) }}</el-text>
        </el-form-item>

        <el-button
          type="primary"
          :loading="installing"
          :disabled="!canInstall"
          @click="installSkill"
        >
          {{ t('skills.install') }}
        </el-button>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="list-header">
          <span>{{ t('skills.listTitle') }}</span>
          <el-space>
            <el-select v-model="viewScope" style="width: 180px" @change="loadSkills">
              <el-option :label="t('skills.scopeGlobal')" value="global" />
              <el-option :label="t('skills.scopeAgent')" value="agent" />
            </el-select>
            <el-select
              v-if="viewScope === 'agent'"
              v-model="viewAgentID"
              filterable
              clearable
              :placeholder="t('skills.selectAgent')"
              style="width: 220px"
              @change="loadSkills"
            >
              <el-option v-for="a in agents" :key="a.agent_id" :label="a.agent_id" :value="a.agent_id" />
            </el-select>
          </el-space>
        </div>
      </template>

      <el-table v-loading="loading" :data="skills" row-key="name" style="width: 100%">
        <el-table-column prop="name" :label="t('skills.columns.name')" min-width="220" />
        <el-table-column :label="t('skills.columns.scope')" width="160">
          <template #default="{ row }">
            <el-tag type="info">{{ row.scope }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('skills.columns.agentId')" width="180">
          <template #default="{ row }">
            <el-text>{{ row.agent_id || t('common.emptyValue') }}</el-text>
          </template>
        </el-table-column>
        <el-table-column :label="t('skills.columns.meta')" width="120">
          <template #default="{ row }">
            <el-tag :type="row.has_meta ? 'success' : 'warning'">
              {{ row.has_meta ? t('skills.metaComplete') : t('skills.metaMissing') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('skills.columns.size')" width="160">
          <template #default="{ row }">{{ formatBytes(row.size_bytes) }}</template>
        </el-table-column>
        <el-table-column :label="t('skills.columns.actions')" width="120">
          <template #default="{ row }">
            <el-button type="danger" link :loading="deletingName === row.name" @click="deleteSkill(row)">{{ t('common.actions.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && skills.length === 0"
        :description="t('skills.empty')"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

type SkillItem = {
  name: string
  scope: string
  agent_id?: string
  size_bytes: number
  has_meta: boolean
}

type AgentItem = {
  agent_id: string
}

const loading = ref(false)
const installing = ref(false)
const deletingName = ref('')
const errorMessage = ref('')
const skills = ref<SkillItem[]>([])
const agents = ref<AgentItem[]>([])
const selectedFile = ref<File | null>(null)
const selectedFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const { t } = useI18n()

const installForm = ref({
  scope: 'global',
  agent_id: '',
  skill_name: '',
})

const viewScope = ref<'global' | 'agent'>('global')
const viewAgentID = ref('')

const totalBytes = computed(() => skills.value.reduce((sum, it) => sum + (it.size_bytes || 0), 0))
const scopeLabel = computed(() => viewScope.value === 'global' ? 'global' : `agent:${viewAgentID.value || t('common.emptyValue')}`)
const canInstall = computed(() => {
  if (!selectedFile.value) return false
  if (installForm.value.scope === 'agent' && !installForm.value.agent_id.trim()) return false
  return true
})

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / Math.pow(1024, exp)
  return `${value.toFixed(value >= 10 || exp === 0 ? 0 : 1)} ${units[exp]}`
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function loadAgents() {
  try {
    const { data } = await axios.get('/api/v1/agents')
    agents.value = Array.isArray(data?.agents) ? data.agents : []
  } catch {
    agents.value = []
  }
}

async function loadSkills() {
  loading.value = true
  errorMessage.value = ''
  try {
    const params: Record<string, string> = { scope: viewScope.value }
    if (viewScope.value === 'agent') {
      if (!viewAgentID.value.trim()) {
        skills.value = []
        errorMessage.value = t('skills.messages.selectAgentForScope')
        return
      }
      params.agent_id = viewAgentID.value.trim()
    }
    const { data } = await axios.get('/api/v1/skills', { params })
    skills.value = Array.isArray(data?.skills) ? data.skills : []
  } catch (err) {
    skills.value = []
    errorMessage.value = parseError(err, t('skills.messages.loadFailedHint'))
    ElMessage.error(t('skills.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

function onFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0] || null
  selectedFile.value = file
  selectedFileName.value = file?.name || ''
}

async function installSkill() {
  if (!selectedFile.value) {
    ElMessage.warning(t('skills.messages.needPackage'))
    return
  }
  if (installForm.value.scope === 'agent' && !installForm.value.agent_id.trim()) {
    ElMessage.warning(t('skills.messages.needAgent'))
    return
  }

  const fd = new FormData()
  fd.append('file', selectedFile.value)
  fd.append('scope', installForm.value.scope)
  if (installForm.value.agent_id.trim()) fd.append('agent_id', installForm.value.agent_id.trim())
  if (installForm.value.skill_name.trim()) fd.append('skill_name', installForm.value.skill_name.trim())

  installing.value = true
  try {
    await axios.post('/api/v1/skills/install', fd)
    ElMessage.success(t('skills.messages.installSubmitted'))
    installForm.value.skill_name = ''
    selectedFile.value = null
    selectedFileName.value = ''
    if (fileInputRef.value) fileInputRef.value.value = ''

    if (installForm.value.scope === 'agent') {
      viewScope.value = 'agent'
      viewAgentID.value = installForm.value.agent_id.trim()
    } else {
      viewScope.value = 'global'
    }
    await loadSkills()
  } catch (err) {
    ElMessage.error(parseError(err, t('skills.messages.installFailed')))
  } finally {
    installing.value = false
  }
}

async function deleteSkill(row: SkillItem) {
  try {
    await ElMessageBox.confirm(
      t('skills.messages.confirmDelete', { name: row.name }),
      t('skills.messages.deleteConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }
  deletingName.value = row.name
  try {
    const params: Record<string, string> = { scope: row.scope || viewScope.value }
    if (params.scope === 'agent' && row.agent_id) params.agent_id = row.agent_id
    await axios.delete(`/api/v1/skills/${encodeURIComponent(row.name)}`, { params })
    ElMessage.success(t('skills.messages.deleteSuccess'))
    await loadSkills()
  } catch (err) {
    ElMessage.error(parseError(err, t('skills.messages.deleteFailed')))
  } finally {
    deletingName.value = ''
  }
}

onMounted(async () => {
  await loadAgents()
  await loadSkills()
})
</script>

<style scoped>
.skills-page {
  display: grid;
  gap: 12px;
}
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.topbar h3 {
  margin: 0;
}
.stats-row {
  margin: 0;
}
.file-hint {
  margin-left: 8px;
}
.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
</style>
