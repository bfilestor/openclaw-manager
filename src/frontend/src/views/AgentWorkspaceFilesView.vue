<template>
  <div class="workspace-files-page">
    <div class="topbar">
      <h3>Workspace Markdown 文件</h3>
      <el-space>
        <el-button @click="goBack">返回 Agents</el-button>
        <el-button :loading="loading" @click="loadAll">刷新</el-button>
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
          <span>Agent: {{ agentID }}</span>
          <el-text type="info">Workspace: {{ workspacePath || '-' }}</el-text>
        </div>
      </template>

      <el-table v-loading="loading" :data="files" row-key="path" style="width: 100%">
        <el-table-column prop="path" label="文件路径" min-width="360" />
        <el-table-column label="大小" width="120">
          <template #default="{ row }">{{ formatBytes(row.size) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="180">
          <template #default="{ row }">{{ formatDateTime(row.modified_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link @click="goEdit(row.path)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && files.length === 0" description="该 Agent Workspace 下暂无 .md 文件" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'

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

const agentID = computed(() => String(route.params.id || '').trim())

function goBack() {
  router.push('/agents')
}

function goEdit(path: string) {
  router.push({
    path: `/agents/${encodeURIComponent(agentID.value)}/workspace-files/edit`,
    query: { path },
  })
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function formatDateTime(v: string): string {
  if (!v) return '-'
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
    errorMessage.value = '缺少 agent_id 参数'
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
    errorMessage.value = parseError(err, '加载 Workspace 文件列表失败')
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
</style>
