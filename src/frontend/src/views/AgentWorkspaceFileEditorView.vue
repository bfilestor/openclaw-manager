<template>
  <div class="config-page">
    <div class="topbar">
      <h3>{{ t('workspaceEditor.title') }}</h3>
      <el-space>
        <el-button @click="goBack">{{ t('workspaceEditor.backToFiles') }}</el-button>
        <el-button :loading="loading" @click="loadAll">{{ t('common.actions.refresh') }}</el-button>
        <el-button type="success" :loading="saving" :disabled="!canEdit" @click="saveFile">
          {{ t('workspaceEditor.save') }}
        </el-button>
      </el-space>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12">
      <el-col :xs="24" :lg="15">
        <el-card shadow="never">
          <template #header>{{ filePath || t('common.emptyValue') }}</template>
          <el-space class="meta-row">
            <el-tag type="info">{{ t('workspaceEditor.size', { size: formatBytes(sizeBytes) }) }}</el-tag>
            <el-tag type="info">{{ t('workspaceEditor.updatedAt', { time: formatDateTime(modifiedAt) }) }}</el-tag>
          </el-space>
          <el-input
            v-model="content"
            type="textarea"
            :rows="20"
            spellcheck="false"
            :autosize="{ minRows: 20, maxRows: 28 }"
            class="editor"
            :placeholder="t('workspaceEditor.markdownPlaceholder')"
          />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="9">
        <el-card shadow="never">
          <template #header>
            <div class="revision-header">
              <span>{{ t('workspaceEditor.revisions.title') }}</span>
              <el-space>
                <el-text type="info">{{ t('workspaceEditor.revisions.selectedCount', { count: selectedRevisions.length }) }}</el-text>
                <el-button size="small" :disabled="selectedRevisions.length !== 2" @click="compareSelectedRevisions">
                  {{ t('workspaceEditor.revisions.compareSelected') }}
                </el-button>
              </el-space>
            </div>
          </template>
          <el-table
            v-loading="loadingRevisions"
            :data="revisions"
            row-key="revision_id"
            style="width: 100%"
            @selection-change="onRevisionSelectionChange"
          >
            <el-table-column type="selection" width="44" />
            <el-table-column prop="created_at" :label="t('workspaceEditor.revisions.columns.time')" min-width="170">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column :label="t('workspaceEditor.revisions.columns.sha')" min-width="120">
              <template #default="{ row }">
                <code>{{ shortSHA(row.sha256) }}</code>
              </template>
            </el-table-column>
            <el-table-column :label="t('workspaceEditor.revisions.columns.actions')" width="180">
              <template #default="{ row }">
                <el-space>
                  <el-button type="info" link @click="previewRevision(row)">{{ t('workspaceEditor.revisions.view') }}</el-button>
                  <el-button type="success" link @click="compareWithCurrent(row)">{{ t('workspaceEditor.revisions.compareCurrent') }}</el-button>
                  <el-button
                    type="warning"
                    link
                    :disabled="!canEdit"
                    :loading="restoringID === row.revision_id"
                    @click="restoreRevision(row)"
                  >
                    {{ t('workspaceEditor.revisions.restore') }}
                  </el-button>
                  <el-button
                    type="danger"
                    link
                    :disabled="!canEdit"
                    :loading="deletingID === row.revision_id"
                    @click="deleteRevision(row)"
                  >
                    {{ t('common.actions.delete') }}
                  </el-button>
                </el-space>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!loadingRevisions && revisions.length === 0" :description="t('workspaceEditor.revisions.empty')" />
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="revisionDialogVisible" width="760px" :title="t('workspaceEditor.revisions.previewTitle', { id: currentRevisionID || t('common.emptyValue') })">
      <el-scrollbar height="420px">
        <pre class="revision-content">{{ currentRevisionContent }}</pre>
      </el-scrollbar>
      <template #footer>
        <el-button @click="revisionDialogVisible = false">{{ t('common.actions.close') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="diffDialogVisible" width="1080px" :title="t('workspaceEditor.revisions.diffTitle', { id: currentRevisionID || t('common.emptyValue') })">
      <DiffViewer :from-text="diffFromText" :to-text="diffToText" :height="460" />
      <template #footer>
        <el-button @click="diffDialogVisible = false">{{ t('common.actions.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useRoute, useRouter } from 'vue-router'
import DiffViewer from '../components/DiffViewer.vue'

type Revision = {
  revision_id: string
  content: string
  sha256: string
  created_at: string
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const loadingRevisions = ref(false)
const saving = ref(false)
const restoringID = ref('')
const deletingID = ref('')
const errorMessage = ref('')

const content = ref('')
const sizeBytes = ref(0)
const modifiedAt = ref('')
const revisions = ref<Revision[]>([])

const revisionDialogVisible = ref(false)
const diffDialogVisible = ref(false)
const currentRevisionID = ref('')
const currentRevisionContent = ref('')
const diffFromText = ref('')
const diffToText = ref('')
const selectedRevisions = ref<Revision[]>([])

const agentID = computed(() => String(route.params.id || '').trim())
const filePath = computed(() => String(route.query.path || '').trim())

const canEdit = computed(() => {
  const role = auth.user?.role || 'Viewer'
  return role === 'Operator' || role === 'Admin'
})

function goBack() {
  router.push(`/agents/${encodeURIComponent(agentID.value)}/workspace-files`)
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

function shortSHA(sha: string): string {
  return String(sha || '').slice(0, 10) || t('common.emptyValue')
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function fileApiParams() {
  return { path: filePath.value }
}

async function loadFile() {
  const { data } = await axios.get(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/file`, {
    params: fileApiParams(),
  })
  content.value = typeof data?.content === 'string' ? data.content : ''
  sizeBytes.value = Number(data?.size || new Blob([content.value]).size || 0)
  modifiedAt.value = String(data?.modified_at || '')
}

async function loadRevisions() {
  loadingRevisions.value = true
  try {
    const { data } = await axios.get(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/revisions`, {
      params: { ...fileApiParams(), limit: 50 },
    })
    revisions.value = Array.isArray(data?.revisions) ? data.revisions : []
    selectedRevisions.value = []
  } finally {
    loadingRevisions.value = false
  }
}

async function loadAll() {
  if (!agentID.value || !filePath.value) {
    errorMessage.value = t('workspaceEditor.messages.missingParams')
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    await Promise.all([loadFile(), loadRevisions()])
  } catch (err) {
    errorMessage.value = parseError(err, t('workspaceEditor.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function saveFile() {
  if (!canEdit.value) {
    ElMessage.warning(t('workspaceEditor.messages.noEditPermission'))
    return
  }
  saving.value = true
  try {
    await axios.put(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/file`, {
      content: content.value,
    }, {
      params: fileApiParams(),
    })
    ElMessage.success(t('workspaceEditor.messages.saveSuccess'))
    await loadAll()
  } catch (err) {
    ElMessage.error(parseError(err, t('workspaceEditor.messages.saveFailed')))
  } finally {
    saving.value = false
  }
}

function previewRevision(rev: Revision) {
  currentRevisionID.value = rev.revision_id
  currentRevisionContent.value = rev.content || ''
  revisionDialogVisible.value = true
}

function compareWithCurrent(rev: Revision) {
  currentRevisionID.value = t('workspaceEditor.revisions.toCurrent', { id: rev.revision_id })
  diffFromText.value = rev.content || ''
  diffToText.value = content.value || ''
  diffDialogVisible.value = true
}

function onRevisionSelectionChange(rows: Revision[]) {
  selectedRevisions.value = Array.isArray(rows) ? rows : []
}

function compareSelectedRevisions() {
  if (selectedRevisions.value.length !== 2) {
    ElMessage.warning(t('workspaceEditor.messages.selectTwoRevisions'))
    return
  }
  const [a, b] = selectedRevisions.value
  const olderFirst = new Date(a.created_at).getTime() <= new Date(b.created_at).getTime()
  const fromRev = olderFirst ? a : b
  const toRev = olderFirst ? b : a
  currentRevisionID.value = `${fromRev.revision_id} -> ${toRev.revision_id}`
  diffFromText.value = fromRev.content || ''
  diffToText.value = toRev.content || ''
  diffDialogVisible.value = true
}

async function restoreRevision(rev: Revision) {
  if (!canEdit.value) {
    ElMessage.warning(t('workspaceEditor.messages.noEditPermission'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('workspaceEditor.messages.confirmRestore'),
      t('workspaceEditor.messages.restoreConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }
  restoringID.value = rev.revision_id
  try {
    await axios.post(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/revisions/${rev.revision_id}/restore`, {}, {
      params: fileApiParams(),
    })
    ElMessage.success(t('workspaceEditor.messages.restoreSuccess'))
    await loadAll()
  } catch (err) {
    ElMessage.error(parseError(err, t('workspaceEditor.messages.restoreFailed')))
  } finally {
    restoringID.value = ''
  }
}

async function deleteRevision(rev: Revision) {
  if (!canEdit.value) {
    ElMessage.warning(t('workspaceEditor.messages.noEditPermission'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('workspaceEditor.messages.confirmDeleteRevision'),
      t('workspaceEditor.messages.deleteConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }
  deletingID.value = rev.revision_id
  try {
    await axios.delete(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/revisions/${rev.revision_id}`, {
      params: fileApiParams(),
    })
    ElMessage.success(t('workspaceEditor.messages.deleteSuccess'))
    await loadRevisions()
  } catch (err) {
    ElMessage.error(parseError(err, t('workspaceEditor.messages.deleteFailed')))
  } finally {
    deletingID.value = ''
  }
}

onMounted(loadAll)
</script>

<style scoped>
.config-page {
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
.meta-row {
  margin-bottom: 8px;
}
.revision-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.editor :deep(textarea) {
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}
.revision-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}
</style>
