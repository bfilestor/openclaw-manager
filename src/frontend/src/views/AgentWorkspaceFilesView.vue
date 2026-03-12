<template>
  <div class="workspace-files-page">
    <div class="topbar">
      <h3>{{ t('workspaceFiles.title') }}</h3>
      <el-space>
        <el-button @click="goBack">{{ t('workspaceFiles.backToAgents') }}</el-button>
        <el-button :loading="loading" @click="loadAll">{{ t('common.actions.refresh') }}</el-button>
      </el-space>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>{{ t('workspaceFiles.agentId', { id: agentID }) }}</span>
          <el-text type="info">{{ t('workspaceFiles.workspacePath', { path: workspacePath || t('common.emptyValue') }) }}</el-text>
        </div>
      </template>

      <el-table v-loading="loading" :data="files" row-key="path" style="width: 100%">
        <el-table-column prop="path" :label="t('workspaceFiles.columns.path')" min-width="360" />
        <el-table-column :label="t('workspaceFiles.columns.size')" width="120">
          <template #default="{ row }">{{ formatBytes(row.size) }}</template>
        </el-table-column>
        <el-table-column :label="t('workspaceFiles.columns.updatedAt')" min-width="180">
          <template #default="{ row }">{{ formatDateTime(row.modified_at) }}</template>
        </el-table-column>
        <el-table-column :label="t('workspaceFiles.columns.actions')" width="180">
          <template #default="{ row }">
            <el-space>
              <el-button type="info" link @click="viewFile(row.path)">{{ t('workspaceFiles.actions.view') }}</el-button>
              <el-button type="success" link @click="goEdit(row.path)">{{ t('workspaceFiles.actions.edit') }}</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && files.length === 0" :description="t('workspaceFiles.empty')" />
    </el-card>

    <el-dialog v-model="previewVisible" width="860px" :title="t('workspaceFiles.previewTitle', { path: previewPath || t('common.emptyValue') })">
      <el-scrollbar height="460px" v-loading="previewLoading">
        <pre class="preview-content">{{ previewContent }}</pre>
      </el-scrollbar>
      <template #footer>
        <el-space>
          <el-button @click="previewVisible = false">{{ t('common.actions.close') }}</el-button>
          <el-button type="success" @click="goEdit(previewPath)">{{ t('workspaceFiles.actions.goEdit') }}</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { useI18n } from 'vue-i18n'

type AgentItem = {
  agent_id: string
  workspace_path: string
}

type WorkspaceFile = {
  path: string
  size: number
  modified_at: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const errorMessage = ref('')
const files = ref<WorkspaceFile[]>([])
const workspacePath = ref('')
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewPath = ref('')
const previewContent = ref('')
const { t } = useI18n()

const agentID = computed(() => String(route.params.id || '').trim())

function goBack() {
  router.push('/agents')
}

function goEdit(path: string) {
  if (!path) return
  router.push({
    path: `/agents/${encodeURIComponent(agentID.value)}/workspace-files/edit`,
    query: { path },
  })
}

async function viewFile(path: string) {
  if (!path || !agentID.value) return
  previewVisible.value = true
  previewLoading.value = true
  previewPath.value = path
  previewContent.value = ''
  try {
    const { data } = await axios.get(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/file`, {
      params: { path },
    })
    previewContent.value = typeof data?.content === 'string' ? data.content : ''
  } catch (err) {
    previewContent.value = parseError(err, t('workspaceFiles.messages.readFailed'))
  } finally {
    previewLoading.value = false
  }
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function formatDateTime(v: string): string {
  if (!v) return t('common.emptyValue')
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString()
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const val = bytes / Math.pow(1024, exp)
  return `${val.toFixed(val >= 10 || exp === 0 ? 0 : 1)} ${units[exp]}`
}

async function loadAll() {
  if (!agentID.value) {
    errorMessage.value = t('workspaceFiles.messages.missingAgentId')
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    const [agentResp, filesResp] = await Promise.all([
      axios.get<AgentItem>(`/api/v1/agents/${encodeURIComponent(agentID.value)}`),
      axios.get(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/files`),
    ])
    workspacePath.value = String(agentResp.data?.workspace_path || '')
    files.value = Array.isArray(filesResp.data?.files) ? filesResp.data.files : []
  } catch (err) {
    errorMessage.value = parseError(err, t('workspaceFiles.messages.loadFailed'))
    files.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadAll)
</script>

<style scoped>
.workspace-files-page {
  display: grid;
  gap: 12px;
}
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.topbar h3 {
  margin: 0;
}
.card-header {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}
.preview-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.65;
}
</style>
