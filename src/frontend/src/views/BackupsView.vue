<template>
  <div class="backups-page">
    <div class="topbar">
      <h3>Backups</h3>
      <el-button :loading="loading" @click="loadBackups">刷新</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-card shadow="never">
      <template #header>创建备份</template>
      <el-form label-position="top">
        <el-form-item label="备份标签">
          <el-input
            v-model="createForm.label"
            placeholder="例如：before-upgrade-2026-03-05"
            maxlength="120"
            clearable
          />
        </el-form-item>
        <el-form-item label="备份范围">
          <el-checkbox-group v-model="createForm.scope">
            <el-checkbox v-for="opt in scopeOptions" :key="opt.value" :label="opt.value">
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-button
          type="primary"
          :loading="creating"
          :disabled="!canCreate || createForm.scope.length === 0"
          @click="createBackup"
        >
          创建备份
        </el-button>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>备份列表</template>
      <el-table v-loading="loading" :data="backups" row-key="backup_id" style="width: 100%">
        <el-table-column prop="label" label="标签" min-width="220" />
        <el-table-column prop="backup_id" label="Backup ID" min-width="280" />
        <el-table-column label="大小" width="130">
          <template #default="{ row }">{{ formatBytes(row.size_bytes) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="200">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="320">
          <template #default="{ row }">
            <el-space>
              <el-button type="info" link @click="viewBackupDetail(row.backup_id)">
                详情
              </el-button>
              <el-button
                type="success"
                link
                :loading="downloadingID === row.backup_id"
                :disabled="!canDownload"
                @click="downloadBackup(row.backup_id)"
              >
                下载
              </el-button>
              <el-button
                type="primary"
                link
                :disabled="!canRestore"
                @click="previewRestore(row.backup_id)"
              >
                还原
              </el-button>
              <el-button
                type="danger"
                link
                :disabled="!canDelete"
                @click="deleteBackup(row.backup_id)"
              >
                删除
              </el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && backups.length === 0" description="暂无备份记录" />
    </el-card>

    <el-dialog
      v-model="restoreDialogVisible"
      width="700px"
      :title="`还原预演 - ${pendingRestoreId || '-'}`"
    >
      <el-alert
        title="将先查看覆盖清单，确认后才会执行真正还原。"
        type="warning"
        show-icon
        :closable="false"
      />
      <el-checkbox v-model="restoreRestartGateway" class="restart-opt">
        还原后重启 Gateway
      </el-checkbox>
      <el-card shadow="never">
        <template #header>
          覆盖文件清单（{{ restorePreview.length }}）
        </template>
        <el-scrollbar height="260px">
          <pre class="preview-box">{{ restorePreview.join('\n') || '（空）' }}</pre>
        </el-scrollbar>
      </el-card>
      <template #footer>
        <el-space>
          <el-button @click="restoreDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="restoring" @click="confirmRestore">
            确认还原
          </el-button>
        </el-space>
      </template>
    </el-dialog>

    <el-dialog
      v-model="manifestDialogVisible"
      width="760px"
      :title="`备份详情 - ${manifestBackupID || '-'}`"
    >
      <el-card shadow="never" v-loading="manifestLoading" class="manifest-summary-card">
        <template #header>本次纳入的 Workspace（{{ workspacePaths.length }}）</template>
        <el-empty v-if="workspacePaths.length === 0" description="该备份未包含 workspace 目录" />
        <el-scrollbar v-else height="120px">
          <div class="workspace-list">
            <el-tag v-for="p in workspacePaths" :key="p" type="success" effect="plain">{{ p }}</el-tag>
          </div>
        </el-scrollbar>

        <el-alert
          v-if="missingWorkspacePaths.length > 0"
          class="workspace-alert"
          type="warning"
          show-icon
          :closable="false"
          :title="`检测到 ${missingWorkspacePaths.length} 个当前 Agent Workspace 未出现在该备份中`"
        />
        <el-scrollbar v-if="missingWorkspacePaths.length > 0" height="100px">
          <div class="workspace-list missing">
            <el-tag v-for="p in missingWorkspacePaths" :key="p" type="danger" effect="plain">{{ p }}</el-tag>
          </div>
        </el-scrollbar>
      </el-card>

      <el-card shadow="never" v-loading="manifestLoading">
        <template #header>Manifest 原文</template>
        <el-scrollbar height="240px">
          <pre class="manifest-box">{{ manifestContent }}</pre>
        </el-scrollbar>
      </el-card>
      <template #footer>
        <el-button @click="manifestDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

type BackupItem = {
  backup_id: string
  label: string
  size_bytes: number
  sha256: string
  created_at: string
}

type AgentItem = {
  workspace_path?: string
}

const auth = useAuthStore()
const router = useRouter()
const loading = ref(false)
const creating = ref(false)
const restoring = ref(false)
const manifestLoading = ref(false)
const errorMessage = ref('')
const backups = ref<BackupItem[]>([])
const downloadingID = ref('')

const createForm = ref({
  label: '',
  scope: ['openclaw_json', 'global_skills']
})

const scopeOptions = [
  { value: 'openclaw_json', label: 'openclaw.json 配置' },
  { value: 'global_skills', label: '全局 skills' },
  { value: 'workspaces', label: 'agents workspaces' },
  { value: 'user_systemd_unit', label: 'user systemd unit' },
  { value: 'manager_revisions', label: 'manager revisions' }
]

const role = computed(() => auth.user?.role || 'Viewer')
const canCreate = computed(() => role.value === 'Operator' || role.value === 'Admin')
const canRestore = computed(() => role.value === 'Admin')
const canDelete = computed(() => role.value === 'Admin')
const canDownload = computed(() => role.value === 'Operator' || role.value === 'Admin')

const restoreDialogVisible = ref(false)
const pendingRestoreId = ref('')
const restorePreview = ref<string[]>([])
const restoreRestartGateway = ref(true)
const manifestDialogVisible = ref(false)
const manifestBackupID = ref('')
const manifestContent = ref('{}')
const workspacePaths = ref<string[]>([])
const missingWorkspacePaths = ref<string[]>([])

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const exp = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / Math.pow(1024, exp)
  return `${value.toFixed(value >= 10 || exp === 0 ? 0 : 1)} ${units[exp]}`
}

function formatDateTime(v: string): string {
  if (!v) return '-'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString()
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function asPrettyJSON(v: any): string {
  if (typeof v === 'string') {
    try {
      return JSON.stringify(JSON.parse(v), null, 2)
    } catch {
      return v
    }
  }
  return JSON.stringify(v ?? {}, null, 2)
}

function extractWorkspacePaths(manifest: any): string[] {
  const scopes = Array.isArray(manifest?.scope) ? manifest.scope.map((x: any) => String(x)) : []
  if (!scopes.includes('workspaces')) return []
  const paths = Array.isArray(manifest?.paths) ? manifest.paths : []
  return Array.from(new Set(paths
    .map((x: any) => String(x || '').trim())
    .filter((p: string) => p.includes('/workspace'))))
}

async function detectMissingWorkspaces(includedPaths: string[]) {
  missingWorkspacePaths.value = []
  if (includedPaths.length === 0) return
  try {
    const { data } = await axios.get('/api/v1/agents')
    const list = Array.isArray(data?.agents) ? data.agents as AgentItem[] : []
    const expected = Array.from(new Set(list
      .map((it) => String(it.workspace_path || '').trim())
      .filter((p) => p.includes('/workspace'))))
    if (expected.length === 0) return
    const included = new Set(includedPaths)
    missingWorkspacePaths.value = expected.filter((p) => !included.has(p))
  } catch {
    missingWorkspacePaths.value = []
  }
}

async function loadBackups() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get('/api/v1/backups')
    backups.value = Array.isArray(data?.backups) ? data.backups : []
  } catch (err) {
    backups.value = []
    errorMessage.value = parseError(err, '加载备份列表失败')
  } finally {
    loading.value = false
  }
}

async function createBackup() {
  if (!canCreate.value) {
    ElMessage.warning('当前角色无创建备份权限')
    return
  }
  if (createForm.value.scope.length === 0) {
    ElMessage.warning('请至少选择一个备份范围')
    return
  }
  creating.value = true
  try {
    const { data } = await axios.post('/api/v1/backups', {
      label: createForm.value.label.trim(),
      scope: createForm.value.scope
    })
    const taskID = String(data?.task_id || '').trim()
    if (taskID) {
      ElMessage.success(`备份任务已提交，task_id: ${taskID}`)
      await router.push({ path: '/tasks', query: { task_id: taskID } })
      return
    }
    ElMessage.success('备份创建请求已提交')
    createForm.value.label = ''
    await loadBackups()
  } catch (err) {
    ElMessage.error(parseError(err, '创建备份失败'))
  } finally {
    creating.value = false
  }
}

async function previewRestore(backupID: string) {
  if (!canRestore.value) {
    ElMessage.warning('当前角色无还原权限')
    return
  }
  restoring.value = true
  try {
    const { data } = await axios.post(`/api/v1/backups/${backupID}/restore`, { dry_run: true })
    restorePreview.value = Array.isArray(data?.will_overwrite) ? data.will_overwrite : []
    pendingRestoreId.value = backupID
    restoreDialogVisible.value = true
  } catch (err) {
    ElMessage.error(parseError(err, '还原预演失败'))
  } finally {
    restoring.value = false
  }
}

async function confirmRestore() {
  if (!pendingRestoreId.value) return
  restoring.value = true
  try {
    const { data } = await axios.post(`/api/v1/backups/${pendingRestoreId.value}/restore`, {
      dry_run: false,
      restart_gateway: restoreRestartGateway.value
    })
    restoreDialogVisible.value = false
    const taskID = String(data?.task_id || '').trim()
    if (taskID) {
      ElMessage.success(`还原任务已提交，task_id: ${taskID}`)
      await router.push({ path: '/tasks', query: { task_id: taskID } })
      return
    }
    ElMessage.success('还原请求已提交')
    await loadBackups()
  } catch (err) {
    ElMessage.error(parseError(err, '执行还原失败'))
  } finally {
    restoring.value = false
  }
}

async function viewBackupDetail(backupID: string) {
  manifestLoading.value = true
  workspacePaths.value = []
  missingWorkspacePaths.value = []
  try {
    const { data } = await axios.get(`/api/v1/backups/${backupID}`)
    manifestBackupID.value = backupID
    manifestContent.value = asPrettyJSON(data)
    workspacePaths.value = extractWorkspacePaths(data)
    await detectMissingWorkspaces(workspacePaths.value)
    manifestDialogVisible.value = true
  } catch (err) {
    ElMessage.error(parseError(err, '读取备份详情失败'))
  } finally {
    manifestLoading.value = false
  }
}

async function downloadBackup(backupID: string) {
  if (!canDownload.value) {
    ElMessage.warning('当前角色无下载权限')
    return
  }
  downloadingID.value = backupID
  try {
    const res = await axios.get(`/api/v1/backups/${backupID}/download`, {
      responseType: 'blob'
    })
    const blob = res.data instanceof Blob
      ? res.data
      : new Blob([res.data], { type: 'application/gzip' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${backupID}.tar.gz`
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    ElMessage.success('备份下载已开始')
  } catch (err) {
    ElMessage.error(parseError(err, '下载备份失败'))
  } finally {
    downloadingID.value = ''
  }
}

async function deleteBackup(backupID: string) {
  if (!canDelete.value) {
    ElMessage.warning('当前角色无删除权限')
    return
  }
  try {
    await ElMessageBox.confirm(`确认删除备份 ${backupID} ？`, '删除确认', { type: 'warning' })
    await axios.delete(`/api/v1/backups/${backupID}`)
    ElMessage.success('备份已删除')
    await loadBackups()
  } catch (err: any) {
    if (err === 'cancel' || err === 'close') return
    ElMessage.error(parseError(err, '删除备份失败'))
  }
}

onMounted(loadBackups)
</script>

<style scoped>
.backups-page {
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
.restart-opt {
  margin: 12px 0;
}
.preview-box {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}
.manifest-summary-card {
  margin-bottom: 12px;
}
.workspace-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.workspace-list.missing {
  margin-top: 8px;
}
.workspace-alert {
  margin-top: 12px;
}
.manifest-box {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}
</style>
