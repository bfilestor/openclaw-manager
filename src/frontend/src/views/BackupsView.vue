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
      <template #header>备份计划</template>
      <el-form inline>
        <el-form-item label="名称">
          <el-input v-model="planForm.name" placeholder="例如：每日配置备份" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="planForm.label" placeholder="可选" />
        </el-form-item>
        <el-form-item label="调度类型">
          <el-select v-model="planForm.schedule_kind" style="width: 140px">
            <el-option label="每天" value="daily" />
            <el-option label="每月" value="monthly" />
            <el-option label="固定间隔" value="interval" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="planForm.schedule_kind === 'daily' || planForm.schedule_kind === 'monthly'" label="执行时间(HH:MM:SS)">
          <el-input v-model="planForm.daily_time" placeholder="02:00:00" style="width: 140px" />
        </el-form-item>
        <el-form-item v-if="planForm.schedule_kind === 'monthly'" label="每月几号">
          <el-input-number v-model="planForm.monthly_day" :min="1" :max="31" />
        </el-form-item>
        <el-form-item v-if="planForm.schedule_kind === 'interval'" label="间隔(分钟)">
          <el-input-number v-model="planForm.interval_minutes" :min="1" :max="10080" />
        </el-form-item>
        <el-form-item label="最多保留份数">
          <el-input-number v-model="planForm.retention_count" :min="1" :max="999" />
        </el-form-item>
      </el-form>
      <el-form-item label="备份范围">
        <el-checkbox-group v-model="planForm.scope">
          <el-checkbox v-for="opt in scopeOptions" :key="`plan-${opt.value}`" :label="opt.value">
            {{ t(`backups.scopeOptions.${opt.value}`) }}
          </el-checkbox>
        </el-checkbox-group>
      </el-form-item>
      <el-button type="primary" :disabled="!canCreate" @click="createPlan">新增计划</el-button>

      <el-table :data="plans" row-key="plan_id" style="width: 100%; margin-top: 12px">
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column label="调度" min-width="180">
          <template #default="{ row }">{{ formatPlanSchedule(row) }}</template>
        </el-table-column>
        <el-table-column label="保留份数" width="100">
          <template #default="{ row }">{{ Number(row.retention_count) > 0 ? row.retention_count : 30 }}</template>
        </el-table-column>
        <el-table-column prop="next_run_at" label="下次执行" min-width="180">
          <template #default="{ row }">{{ formatDateTime(row.next_run_at) }}</template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">{{ row.enabled ? '启用' : '停用' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="220">
          <template #default="{ row }">
            <el-button link type="primary" @click="runPlanNow(row.plan_id)">立即执行</el-button>
            <el-button link :type="row.enabled ? 'warning' : 'success'" @click="togglePlan(row)">{{ row.enabled ? '停用' : '启用' }}</el-button>
            <el-button link type="danger" @click="deletePlan(row.plan_id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
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

type PlanItem = {
  plan_id: string
  name: string
  label: string
  scope: string[]
  schedule_kind: 'interval' | 'daily' | 'monthly'
  daily_time?: string
  monthly_day?: number
  interval_minutes?: number
  retention_count: number
  enabled: boolean
  next_run_at: string
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

const plans = ref<PlanItem[]>([])
const planForm = ref({
  name: '',
  label: '',
  scope: ['openclaw_json', 'global_skills'],
  schedule_kind: 'daily',
  daily_time: '02:00:00',
  monthly_day: 1,
  interval_minutes: 1440,
  retention_count: 30
})

const scopeOptions = [
  { value: 'openclaw_json' },
  { value: 'global_skills' },
  { value: 'workspaces' },
  { value: 'user_systemd_unit' },
  { value: 'manager_db' }
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

function formatPlanSchedule(row: PlanItem): string {
  if (row.schedule_kind === 'daily') {
    return `每天 ${row.daily_time || '00:00:00'}`
  }
  if (row.schedule_kind === 'monthly') {
    return `每月${row.monthly_day || 1}号 ${row.daily_time || '00:00:00'}`
  }
  return `每${row.interval_minutes || 0}分钟`
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

async function loadPreference() {
  try {
    const { data } = await axios.get('/api/v1/backup-preferences/me')
    const scope = Array.isArray(data?.scope) && data.scope.length > 0 ? data.scope : ['openclaw_json', 'global_skills']
    createForm.value.scope = scope
    planForm.value.scope = [...scope]
    if (typeof data?.label === 'string') createForm.value.label = data.label
  } catch {
    // Ignore preference fetch failure and keep defaults.
  }
}

async function loadPlans() {
  try {
    const { data } = await axios.get('/api/v1/backup-plans')
    const raw = Array.isArray(data?.plans) ? data.plans : []
    plans.value = raw.map((it: any) => ({
      ...it,
      retention_count: Number(it?.retention_count ?? it?.retentionCount ?? 0) > 0
        ? Number(it?.retention_count ?? it?.retentionCount)
        : 30
    }))
  } catch {
    plans.value = []
  }
}

async function createPlan() {
  if (!canCreate.value) return
  const mode = planForm.value.schedule_kind
  const needTime = mode === 'daily' || mode === 'monthly'
  if (!planForm.value.name.trim() || planForm.value.scope.length === 0) {
    ElMessage.warning('请填写完整的计划信息')
    return
  }
  if (mode === 'interval' && planForm.value.interval_minutes <= 0) {
    ElMessage.warning('间隔分钟需要大于 0')
    return
  }
  if (needTime && !/^\d{2}:\d{2}:\d{2}$/.test(planForm.value.daily_time)) {
    ElMessage.warning('时间格式必须是 HH:MM:SS')
    return
  }
  if (mode === 'monthly' && (planForm.value.monthly_day < 1 || planForm.value.monthly_day > 31)) {
    ElMessage.warning('每月几号必须在 1..31')
    return
  }
  try {
    await axios.post('/api/v1/backup-plans', {
      name: planForm.value.name.trim(),
      label: planForm.value.label.trim(),
      scope: planForm.value.scope,
      schedule_kind: planForm.value.schedule_kind,
      daily_time: planForm.value.daily_time,
      monthly_day: planForm.value.monthly_day,
      interval_minutes: planForm.value.interval_minutes,
      retention_count: planForm.value.retention_count
    })
    ElMessage.success('计划创建成功')
    planForm.value.name = ''
    await loadPlans()
  } catch (err) {
    ElMessage.error(parseError(err, '计划创建失败'))
  }
}

async function togglePlan(row: PlanItem) {
  try {
    await axios.post(`/api/v1/backup-plans/${row.plan_id}/${row.enabled ? 'disable' : 'enable'}`)
    await loadPlans()
  } catch (err) {
    ElMessage.error(parseError(err, '操作失败'))
  }
}

async function deletePlan(planID: string) {
  try {
    await axios.delete(`/api/v1/backup-plans/${planID}`)
    await loadPlans()
  } catch (err) {
    ElMessage.error(parseError(err, '删除失败'))
  }
}

async function runPlanNow(planID: string) {
  try {
    const { data } = await axios.post(`/api/v1/backup-plans/${planID}/run`)
    ElMessage.success(`执行成功，backup_id=${data?.backup_id || ''}`)
    await Promise.all([loadPlans(), loadBackups()])
  } catch (err) {
    ElMessage.error(parseError(err, '执行失败'))
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

onMounted(async () => {
  await Promise.all([loadPreference(), loadPlans(), loadBackups()])
})
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

/* Improve button readability in dark theme for this page. */
.backups-page :deep(.el-button:not(.el-button--primary)) {
  color: var(--oc-text);
  background: var(--oc-surface-muted);
  border-color: var(--oc-border);
}

.backups-page :deep(.el-button--primary) {
  color: var(--oc-accent-contrast);
}

.backups-page :deep(.el-button.is-link:not(.el-button--primary)) {
  background: transparent;
}
</style>
