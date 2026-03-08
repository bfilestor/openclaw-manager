<template>
  <div class="config-page">
    <div class="topbar">
      <h3>Config Editor</h3>
      <el-space>
        <el-button :loading="loading" @click="loadAll">刷新</el-button>
        <el-button @click="formatJSON">格式化</el-button>
        <el-button type="primary" :loading="saving" :disabled="!canEdit" @click="saveConfig">
          保存配置
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
          <template #header>openclaw.json</template>
          <el-space class="meta-row">
            <el-tag type="info">大小: {{ formatBytes(sizeBytes) }}</el-tag>
            <el-tag type="info">更新时间: {{ formatDateTime(modifiedAt) }}</el-tag>
          </el-space>
          <el-input
            v-model="content"
            type="textarea"
            :rows="20"
            spellcheck="false"
            :autosize="{ minRows: 20, maxRows: 28 }"
            class="editor"
            placeholder='请输入合法 JSON，例如 {"gateway":{"port":18790}}'
          />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="9">
        <el-card shadow="never">
          <template #header>
            <div class="revision-header">
              <span>Revisions</span>
              <el-space wrap>
                <el-text type="info">已选 {{ selectedRevisions.length }}/2</el-text>
                <el-button type="primary" size="small" :disabled="selectedRevisions.length !== 2" @click="compareSelectedRevisions">
                  版本比较
                </el-button>
              </el-space>
            </div>
          </template>
          <el-table
            ref="revisionTableRef"
            v-loading="loadingRevisions"
            :data="revisions"
            row-key="revision_id"
            style="width: 100%"
            @selection-change="onRevisionSelectionChange"
          >
            <el-table-column type="selection" width="44" />
            <el-table-column prop="created_at" label="时间" min-width="170">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="SHA" min-width="120">
              <template #default="{ row }">
                <code>{{ shortSHA(row.sha256) }}</code>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-space>
                  <el-button type="info" link @click="previewRevision(row)">查看</el-button>
                  <el-button type="primary" link @click="compareWithCurrent(row)">对比当前</el-button>
                  <el-button
                    type="warning"
                    link
                    :disabled="!canEdit"
                    :loading="restoringID === row.revision_id"
                    @click="restoreRevision(row)"
                  >
                    回滚
                  </el-button>
                  <el-button
                    type="danger"
                    link
                    :disabled="!canEdit"
                    :loading="deletingID === row.revision_id"
                    @click="deleteRevision(row)"
                  >
                    删除
                  </el-button>
                </el-space>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!loadingRevisions && revisions.length === 0" description="暂无历史版本" />
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="revisionDialogVisible" width="760px" :title="`Revision - ${currentRevisionID || '-'}`">
      <el-scrollbar height="420px">
        <pre class="revision-content">{{ currentRevisionContent }}</pre>
      </el-scrollbar>
      <template #footer>
        <el-button @click="revisionDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="diffDialogVisible" width="1080px" :title="`Diff - ${currentRevisionID || '-'}`">
      <el-space class="diff-toolbar" wrap>
        <el-text type="info">显示模式</el-text>
        <el-radio-group v-model="diffViewMode" size="small">
          <el-radio-button label="unified">统一视图</el-radio-button>
          <el-radio-button label="split">左右分栏</el-radio-button>
        </el-radio-group>
      </el-space>

      <el-scrollbar height="460px">
        <pre v-if="diffViewMode === 'unified'" class="revision-content diff-content"><template v-for="(line, idx) in diffLines" :key="idx"><span :class="line.type">{{ line.text }}
</span></template></pre>

        <table v-else class="split-diff-table">
          <thead>
            <tr>
              <th>旧版本</th>
              <th>新版本</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in splitDiffRows" :key="idx">
              <td :class="['left', row.leftType]">{{ row.left }}</td>
              <td :class="['right', row.rightType]">{{ row.right }}</td>
            </tr>
          </tbody>
        </table>
      </el-scrollbar>
      <template #footer>
        <el-button @click="diffDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { diffLines as calcDiffLines } from 'diff'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'

type Revision = {
  revision_id: string
  target_type: string
  target_id: string
  content: string
  sha256: string
  created_at: string
  created_by: string
}

const auth = useAuthStore()
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
const revisionTableRef = ref<any>(null)

const revisionDialogVisible = ref(false)
const diffDialogVisible = ref(false)
const currentRevisionID = ref('')
const currentRevisionContent = ref('')
const diffLines = ref<{ text: string; type: 'same' | 'add' | 'remove' }[]>([])
const splitDiffRows = ref<{ left: string; right: string; leftType: 'same' | 'add' | 'remove'; rightType: 'same' | 'add' | 'remove' }[]>([])
const diffViewMode = ref<'unified' | 'split'>('unified')
const selectedRevisions = ref<Revision[]>([])

const canEdit = computed(() => {
  const role = auth.user?.role || 'Viewer'
  return role === 'Operator' || role === 'Admin'
})

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

function shortSHA(sha: string): string {
  return String(sha || '').slice(0, 10) || '-'
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

function normalizeJSON(raw: string): string {
  const parsed = JSON.parse(raw)
  return JSON.stringify(parsed, null, 2)
}

async function loadConfig() {
  const { data } = await axios.get('/api/v1/config/openclaw')
  const raw = data?.content
  content.value = typeof raw === 'string' ? raw : '{}'
  sizeBytes.value = Number(data?.size || new Blob([content.value]).size || 0)
  modifiedAt.value = String(data?.modified_at || '')
}

async function loadRevisions() {
  loadingRevisions.value = true
  try {
    const { data } = await axios.get('/api/v1/config/openclaw/revisions', { params: { limit: 50 } })
    revisions.value = Array.isArray(data?.revisions) ? data.revisions : []
    selectedRevisions.value = []
  } finally {
    loadingRevisions.value = false
  }
}

async function loadAll() {
  loading.value = true
  errorMessage.value = ''
  try {
    await Promise.all([loadConfig(), loadRevisions()])
  } catch (err) {
    errorMessage.value = parseError(err, '加载配置失败')
  } finally {
    loading.value = false
  }
}

function formatJSON() {
  try {
    content.value = normalizeJSON(content.value)
    ElMessage.success('JSON 格式化完成')
  } catch {
    ElMessage.error('当前内容不是合法 JSON，无法格式化')
  }
}

async function saveConfig() {
  if (!canEdit.value) {
    ElMessage.warning('当前角色无编辑权限')
    return
  }
  let normalized = ''
  try {
    normalized = normalizeJSON(content.value)
  } catch {
    ElMessage.error('JSON 格式不合法，请修复后再保存')
    return
  }
  saving.value = true
  try {
    await axios.put('/api/v1/config/openclaw', { content: normalized })
    content.value = normalized
    ElMessage.success('配置保存成功')
    await loadAll()
  } catch (err) {
    ElMessage.error(parseError(err, '保存配置失败'))
  } finally {
    saving.value = false
  }
}

function previewRevision(rev: Revision) {
  currentRevisionID.value = rev.revision_id
  currentRevisionContent.value = rev.content || ''
  revisionDialogVisible.value = true
}

function buildDiffLines(fromText: string, toText: string) {
  const parts = calcDiffLines(fromText || '', toText || '')
  return parts.flatMap((part) => {
    const rows = String(part.value || '').split('\n')
    if (rows.length > 0 && rows[rows.length - 1] === '') rows.pop()
    const type: 'same' | 'add' | 'remove' = part.added ? 'add' : part.removed ? 'remove' : 'same'
    return rows.map((row) => ({
      type,
      text: `${type === 'add' ? '+' : type === 'remove' ? '-' : ' '} ${row}`,
    }))
  })
}

function buildSplitDiffRows(fromText: string, toText: string) {
  const parts = calcDiffLines(fromText || '', toText || '')
  const rows: { left: string; right: string; leftType: 'same' | 'add' | 'remove'; rightType: 'same' | 'add' | 'remove' }[] = []

  for (let i = 0; i < parts.length; i++) {
    const part = parts[i]
    if (part.removed && parts[i + 1]?.added) {
      const removedLines = String(part.value || '').split('\n')
      const addedLines = String(parts[i + 1].value || '').split('\n')
      if (removedLines.length > 0 && removedLines[removedLines.length - 1] === '') removedLines.pop()
      if (addedLines.length > 0 && addedLines[addedLines.length - 1] === '') addedLines.pop()
      const maxLen = Math.max(removedLines.length, addedLines.length)
      for (let j = 0; j < maxLen; j++) {
        rows.push({
          left: removedLines[j] ?? '',
          right: addedLines[j] ?? '',
          leftType: 'remove',
          rightType: 'add',
        })
      }
      i += 1
      continue
    }

    const lines = String(part.value || '').split('\n')
    if (lines.length > 0 && lines[lines.length - 1] === '') lines.pop()
    if (part.added) {
      lines.forEach((line) => rows.push({ left: '', right: line, leftType: 'same', rightType: 'add' }))
    } else if (part.removed) {
      lines.forEach((line) => rows.push({ left: line, right: '', leftType: 'remove', rightType: 'same' }))
    } else {
      lines.forEach((line) => rows.push({ left: line, right: line, leftType: 'same', rightType: 'same' }))
    }
  }

  return rows
}

function compareWithCurrent(rev: Revision) {
  currentRevisionID.value = `${rev.revision_id} -> CURRENT`
  diffLines.value = buildDiffLines(rev.content || '', content.value || '')
  splitDiffRows.value = buildSplitDiffRows(rev.content || '', content.value || '')
  diffViewMode.value = 'unified'
  diffDialogVisible.value = true
}

function onRevisionSelectionChange(rows: Revision[]) {
  const list = Array.isArray(rows) ? rows : []
  if (list.length <= 2) {
    selectedRevisions.value = list
    return
  }

  ElMessage.warning('最多选择 2 个版本进行比较')
  const keep = list.slice(-2)
  selectedRevisions.value = keep
  if (revisionTableRef.value) {
    revisionTableRef.value.clearSelection()
    keep.forEach((row) => revisionTableRef.value.toggleRowSelection(row, true))
  }
}

function compareSelectedRevisions() {
  if (selectedRevisions.value.length !== 2) {
    ElMessage.warning('请先勾选 2 个版本再对比')
    return
  }
  const [a, b] = selectedRevisions.value
  const olderFirst = new Date(a.created_at).getTime() <= new Date(b.created_at).getTime()
  const fromRev = olderFirst ? a : b
  const toRev = olderFirst ? b : a
  currentRevisionID.value = `${fromRev.revision_id} -> ${toRev.revision_id}`
  diffLines.value = buildDiffLines(fromRev.content || '', toRev.content || '')
  splitDiffRows.value = buildSplitDiffRows(fromRev.content || '', toRev.content || '')
  diffViewMode.value = 'unified'
  diffDialogVisible.value = true
}

async function restoreRevision(rev: Revision) {
  if (!canEdit.value) {
    ElMessage.warning('当前角色无编辑权限')
    return
  }
  try {
    await ElMessageBox.confirm('确认回滚到该历史版本？', '回滚确认', { type: 'warning' })
  } catch {
    return
  }
  restoringID.value = rev.revision_id
  try {
    await axios.post(`/api/v1/config/openclaw/revisions/${rev.revision_id}/restore`)
    ElMessage.success('回滚成功')
    await loadAll()
  } catch (err) {
    ElMessage.error(parseError(err, '回滚失败'))
  } finally {
    restoringID.value = ''
  }
}

async function deleteRevision(rev: Revision) {
  if (!canEdit.value) {
    ElMessage.warning('当前角色无编辑权限')
    return
  }
  try {
    await ElMessageBox.confirm('确认删除该历史版本？删除后不可恢复。', '删除确认', { type: 'warning' })
  } catch {
    return
  }
  deletingID.value = rev.revision_id
  try {
    await axios.delete(`/api/v1/config/openclaw/revisions/${rev.revision_id}`)
    ElMessage.success('删除成功')
    await loadRevisions()
  } catch (err) {
    ElMessage.error(parseError(err, '删除失败'))
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
.diff-content .add {
  background: #ecfdf3;
  color: #1b5e20;
}
.diff-content .remove {
  background: #fff1f0;
  color: #b42318;
}
.diff-content .same {
  color: #667085;
}
.diff-toolbar {
  margin-bottom: 8px;
}
.split-diff-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
}
.split-diff-table th,
.split-diff-table td {
  border: 1px solid #eaecf0;
  padding: 4px 8px;
  vertical-align: top;
  white-space: pre-wrap;
  word-break: break-word;
}
.split-diff-table td.add {
  background: #ecfdf3;
  color: #1b5e20;
}
.split-diff-table td.remove {
  background: #fff1f0;
  color: #b42318;
}
.split-diff-table td.same {
  color: #667085;
}
</style>
