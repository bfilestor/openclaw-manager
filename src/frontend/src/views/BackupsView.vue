<template>
  <div class="backups-page">
    <div class="topbar">
      <h3>{{ t('backups.title') }}</h3>
      <el-button :loading="loading" @click="loadBackups">{{ t('common.actions.refresh') }}</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-card shadow="never">
      <template #header>{{ t('backups.createTitle') }}</template>
      <el-form label-position="top">
        <el-form-item :label="t('backups.backupLabel')">
          <el-input
            v-model="createForm.label"
            :placeholder="t('backups.backupLabelPlaceholder')"
            maxlength="120"
            clearable
          />
        </el-form-item>
        <el-form-item :label="t('backups.backupScope')">
          <el-checkbox-group v-model="createForm.scope">
            <el-checkbox v-for="opt in scopeOptions" :key="opt.value" :label="opt.value">
              {{ t(`backups.scopeOptions.${opt.value}`) }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-button
          type="primary"
          :loading="creating"
          :disabled="!canCreate || createForm.scope.length === 0"
          @click="createBackup"
        >
          {{ t('backups.createAction') }}
        </el-button>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>{{ t('backups.listTitle') }}</template>
      <el-table v-loading="loading" :data="backups" row-key="backup_id" style="width: 100%">
        <el-table-column prop="label" :label="t('backups.columns.label')" min-width="220" />
        <el-table-column prop="backup_id" :label="t('backups.columns.backupId')" min-width="280" />
        <el-table-column :label="t('backups.columns.size')" width="130">
          <template #default="{ row }">{{ formatBytes(row.size_bytes) }}</template>
        </el-table-column>
        <el-table-column :label="t('backups.columns.createdAt')" min-width="200">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column :label="t('backups.columns.actions')" width="320">
          <template #default="{ row }">
            <el-space>
              <el-button type="info" link @click="viewBackupDetail(row.backup_id)">
                {{ t('backups.actions.detail') }}
              </el-button>
              <el-button
                type="success"
                link
                :loading="downloadingID === row.backup_id"
                :disabled="!canDownload"
                @click="downloadBackup(row.backup_id)"
              >
                {{ t('backups.actions.download') }}
              </el-button>
              <el-button
                type="primary"
                link
                :disabled="!canRestore"
                @click="previewRestore(row.backup_id)"
              >
                {{ t('backups.actions.restore') }}
              </el-button>
              <el-button
                type="danger"
                link
                :disabled="!canDelete"
                @click="deleteBackup(row.backup_id)"
              >
                {{ t('common.actions.delete') }}
              </el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && backups.length === 0" :description="t('backups.empty')" />
    </el-card>

    <el-dialog
      v-model="restoreDialogVisible"
      width="700px"
      :title="t('backups.restoreDialogTitle', { id: pendingRestoreId || t('common.emptyValue') })"
    >
      <el-alert
        :title="t('backups.restorePreviewTip')"
        type="warning"
        show-icon
        :closable="false"
      />
      <el-checkbox v-model="restoreRestartGateway" class="restart-opt">
        {{ t('backups.restartGatewayAfterRestore') }}
      </el-checkbox>
      <el-card shadow="never">
        <template #header>
          {{ t('backups.overwriteList', { count: restorePreview.length }) }}
        </template>
        <el-scrollbar height="260px">
          <pre class="preview-box">{{ restorePreview.join('\n') || t('backups.emptyList') }}</pre>
        </el-scrollbar>
      </el-card>
      <template #footer>
        <el-space>
          <el-button @click="restoreDialogVisible = false">{{ t('common.actions.cancel') }}</el-button>
          <el-button type="primary" :loading="restoring" @click="confirmRestore">
            {{ t('backups.confirmRestore') }}
          </el-button>
        </el-space>
      </template>
    </el-dialog>

    <el-dialog
      v-model="manifestDialogVisible"
      width="760px"
      :title="t('backups.manifestDialogTitle', { id: manifestBackupID || t('common.emptyValue') })"
    >
      <el-card shadow="never" v-loading="manifestLoading" class="manifest-summary-card">
        <template #header>{{ t('backups.workspaceIncluded', { count: workspacePaths.length }) }}</template>
        <el-empty v-if="workspacePaths.length === 0" :description="t('backups.noWorkspaceIncluded')" />
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
          :title="t('backups.missingWorkspaceWarning', { count: missingWorkspacePaths.length })"
        />
        <el-scrollbar v-if="missingWorkspacePaths.length > 0" height="100px">
          <div class="workspace-list missing">
            <el-tag v-for="p in missingWorkspacePaths" :key="p" type="danger" effect="plain">{{ p }}</el-tag>
          </div>
        </el-scrollbar>
      </el-card>

      <el-card shadow="never" v-loading="manifestLoading">
        <template #header>{{ t('backups.manifestRaw') }}</template>
        <el-scrollbar height="240px">
          <pre class="manifest-box">{{ manifestContent }}</pre>
        </el-scrollbar>
      </el-card>
      <template #footer>
        <el-button @click="manifestDialogVisible = false">{{ t('common.actions.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
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
const { t } = useI18n()
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
  { value: 'openclaw_json' },
  { value: 'global_skills' },
  { value: 'workspaces' },
  { value: 'user_systemd_unit' },
  { value: 'manager_revisions' }
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
  if (!v) return t('common.emptyValue')
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
    errorMessage.value = parseError(err, t('backups.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function createBackup() {
  if (!canCreate.value) {
    ElMessage.warning(t('backups.messages.noCreatePermission'))
    return
  }
  if (createForm.value.scope.length === 0) {
    ElMessage.warning(t('backups.messages.needScope'))
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
      ElMessage.success(t('backups.messages.createTaskSubmitted', { taskId: taskID }))
      await router.push({ path: '/tasks', query: { task_id: taskID } })
      return
    }
    ElMessage.success(t('backups.messages.createSubmitted'))
    createForm.value.label = ''
    await loadBackups()
  } catch (err) {
    ElMessage.error(parseError(err, t('backups.messages.createFailed')))
  } finally {
    creating.value = false
  }
}

async function previewRestore(backupID: string) {
  if (!canRestore.value) {
    ElMessage.warning(t('backups.messages.noRestorePermission'))
    return
  }
  restoring.value = true
  try {
    const { data } = await axios.post(`/api/v1/backups/${backupID}/restore`, { dry_run: true })
    restorePreview.value = Array.isArray(data?.will_overwrite) ? data.will_overwrite : []
    pendingRestoreId.value = backupID
    restoreDialogVisible.value = true
  } catch (err) {
    ElMessage.error(parseError(err, t('backups.messages.previewRestoreFailed')))
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
      ElMessage.success(t('backups.messages.restoreTaskSubmitted', { taskId: taskID }))
      await router.push({ path: '/tasks', query: { task_id: taskID } })
      return
    }
    ElMessage.success(t('backups.messages.restoreSubmitted'))
    await loadBackups()
  } catch (err) {
    ElMessage.error(parseError(err, t('backups.messages.restoreFailed')))
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
    ElMessage.error(parseError(err, t('backups.messages.detailFailed')))
  } finally {
    manifestLoading.value = false
  }
}

async function downloadBackup(backupID: string) {
  if (!canDownload.value) {
    ElMessage.warning(t('backups.messages.noDownloadPermission'))
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
    ElMessage.success(t('backups.messages.downloadStarted'))
  } catch (err) {
    ElMessage.error(parseError(err, t('backups.messages.downloadFailed')))
  } finally {
    downloadingID.value = ''
  }
}

async function deleteBackup(backupID: string) {
  if (!canDelete.value) {
    ElMessage.warning(t('backups.messages.noDeletePermission'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('backups.messages.confirmDelete', { id: backupID }),
      t('backups.messages.deleteConfirmTitle'),
      { type: 'warning' }
    )
    await axios.delete(`/api/v1/backups/${backupID}`)
    ElMessage.success(t('backups.messages.deleteSuccess'))
    await loadBackups()
  } catch (err: any) {
    if (err === 'cancel' || err === 'close') return
    ElMessage.error(parseError(err, t('backups.messages.deleteFailed')))
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
