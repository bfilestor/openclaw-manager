<template>
  <div class="config-page">
    <div class="topbar">
      <h3>Markdown 编辑器</h3>
      <el-space>
        <el-button @click="goBack">返回文件列表</el-button>
        <el-button :loading="loading" @click="loadAll">刷新</el-button>
        <el-button type="primary" :loading="saving" :disabled="!canEdit" @click="saveFile">
          保存
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
          <template #header>{{ filePath || '-' }}</template>
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
            placeholder="请输入 Markdown 内容"
          />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="9">
        <el-card shadow="never">
          <template #header>
            <div class="revision-header">
              <span>Revisions</span>
              <el-space>
                <el-text type="info">已选 {{ selectedRevisions.length }}</el-text>
                <el-button size="small" :disabled="selectedRevisions.length !== 2" @click="compareSelectedRevisions">
                  对比所选版本
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
import { useRoute, useRouter } from 'vue-router'

type Revision = {
  revision_id: string
  content: string
  sha256: string
  created_at: string
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
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
const diffLines = ref<{ text: string; type: 'same' | 'add' | 'remove' }[]>([])
const splitDiffRows = ref<{ left: string; right: string; leftType: 'same' | 'add' | 'remove'; rightType: 'same' | 'add' | 'remove' }[]>([])
const diffViewMode = ref<'unified' | 'split'>('unified')
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
    errorMessage.value = '缺少 agent_id 或 path 参数'
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    await Promise.all([loadFile(), loadRevisions()])
  } catch (err) {
    errorMessage.value = parseError(err, '加载文件失败')
  } finally {
    loading.value = false
  }
}

async function saveFile() {
  if (!canEdit.value) {
    ElMessage.warning('当前角色无编辑权限')
    return
  }
  saving.value = true
  try {
    await axios.put(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/file`, {
      content: content.value,
    }, {
      params: fileApiParams(),
    })
    ElMessage.success('保存成功')
    await loadAll()
  } catch (err) {
    ElMessage.error(parseError(err, '保存失败'))
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
  selectedRevisions.value = Array.isArray(rows) ? rows : []
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
    await axios.post(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/revisions/${rev.revision_id}/restore`, {}, {
      params: fileApiParams(),
    })
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
    await axios.delete(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/markdown/revisions/${rev.revision_id}`, {
      params: fileApiParams(),
    })
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
