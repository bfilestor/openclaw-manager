<template>
  <div class="tasks-page">
    <div class="topbar">
      <h3>{{ t('tasks.title') }}</h3>
      <el-space>
        <el-button :loading="loading" @click="load">{{ t('common.actions.refresh') }}</el-button>
        <el-button type="danger" :loading="clearing" :disabled="tasks.length===0" @click="clearTasks">{{ t('tasks.clearTasks') }}</el-button>
      </el-space>
    </div>
    <el-row :gutter="16">
      <el-col :xs="24" :md="10" :lg="8">
        <el-card shadow="never">
          <template #header>{{ t('tasks.taskList') }}</template>
          <el-table
            ref="tableRef"
            :data="tasks"
            row-key="task_id"
            style="width: 100%"
            highlight-current-row
            @row-click="(row) => select(row)"
          >
            <el-table-column prop="task_id" :label="t('tasks.columns.taskId')" min-width="220" />
            <el-table-column prop="task_type" :label="t('tasks.columns.type')" min-width="110" />
            <el-table-column :label="t('tasks.columns.status')" min-width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'FAILED' ? 'danger' : row.status === 'SUCCEEDED' ? 'success' : 'info'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('tasks.columns.actions')" min-width="90" fixed="right">
              <template #default="{ row }">
                <el-button
                  type="danger"
                  link
                  :loading="deletingTaskID===row.task_id"
                  @click.stop="deleteTask(row)"
                >
                  {{ t('common.actions.delete') }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="14" :lg="16">
        <el-card shadow="never">
          <template #header>
            {{ logHeader }}
          </template>
          <el-space class="toolbar">
            <el-checkbox v-model="autoScroll">{{ t('tasks.autoScroll') }}</el-checkbox>
            <el-input v-model="keyword" :placeholder="t('tasks.searchLogs')" clearable />
          </el-space>
          <pre ref="logBox" class="log-box">{{ filteredLog }}</pre>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import axios from 'axios'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const { t } = useI18n()

const tasks = ref<any[]>([])
const selected = ref<any>(null)
const loading = ref(false)
const clearing = ref(false)
const deletingTaskID = ref('')
const logText = ref('')
const autoScroll = ref(true)
const keyword = ref('')
const logBox = ref<HTMLElement | null>(null)
const tableRef = ref<any>(null)
let eventSource: EventSource | null = null
let refreshTimer: ReturnType<typeof setInterval> | null = null

const filteredLog = computed(() =>
  keyword.value ? logText.value.split('\n').filter((l) => l.includes(keyword.value)).join('\n') : logText.value
)
const logHeader = computed(() =>
  selected.value?.task_id
    ? t('tasks.logsWithTask', { taskId: selected.value.task_id })
    : t('tasks.logs')
)

function appendLog(line: string) {
  logText.value += `${line}\n`
  if (autoScroll.value && logBox.value) {
    logBox.value.scrollTop = logBox.value.scrollHeight
  }
}

function closeStream() {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
}

async function load() {
  loading.value = true
  try {
    const { data } = await axios.get('/api/v1/tasks')
    tasks.value = Array.isArray(data?.tasks) ? data.tasks : []
    await focusTaskFromQuery()
  } finally {
    loading.value = false
  }
}

async function focusTaskFromQuery() {
  const taskID = String(route.query.task_id || '').trim()
  if (!taskID) return
  if (selected.value?.task_id === taskID) return
  const hit = tasks.value.find((t) => t.task_id === taskID)
  if (!hit) return
  await select(hit, false)
  tableRef.value?.setCurrentRow?.(hit)
}

async function select(task: any, syncQuery = true) {
  selected.value = task
  logText.value = ''
  closeStream()

  if (syncQuery) {
    router.replace({ path: '/tasks', query: { task_id: String(task.task_id || '') } })
  }

  const token = String(auth.accessToken || '').trim()
  if (!token) {
    appendLog(t('tasks.messages.noAccessToken'))
    return
  }

  const url = `/api/v1/tasks/${encodeURIComponent(task.task_id)}/events?token=${encodeURIComponent(token)}`
  eventSource = new EventSource(url)
  eventSource.onmessage = (ev) => {
    try {
      const payload = JSON.parse(ev.data)
      if (payload?.line) {
        appendLog(String(payload.line))
      } else if (payload?.type === 'done') {
        appendLog(t('tasks.messages.done', { status: payload.status, exitCode: payload.exit_code }))
      } else {
        appendLog(ev.data)
      }
    } catch {
      appendLog(ev.data)
    }
  }
  eventSource.onerror = () => {
    appendLog(t('tasks.messages.streamClosed'))
    closeStream()
  }
}

watch(() => route.query.task_id, () => {
  void focusTaskFromQuery()
})

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function deleteTask(task: any) {
  try {
    await ElMessageBox.confirm(
      t('tasks.messages.confirmDeleteTask', { taskId: task.task_id }),
      t('tasks.messages.deleteConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }
  deletingTaskID.value = String(task.task_id || '')
  try {
    await axios.delete(`/api/v1/tasks/${encodeURIComponent(task.task_id)}`)
    ElMessage.success(t('tasks.messages.deleteSuccess'))
    if (selected.value?.task_id === task.task_id) {
      selected.value = null
      logText.value = ''
      closeStream()
      router.replace({ path: '/tasks', query: {} })
    }
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, t('tasks.messages.deleteFailed')))
  } finally {
    deletingTaskID.value = ''
  }
}

async function clearTasks() {
  try {
    await ElMessageBox.confirm(
      t('tasks.messages.confirmClearTasks'),
      t('tasks.messages.clearConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }
  clearing.value = true
  try {
    const { data } = await axios.delete('/api/v1/tasks')
    const n = Number(data?.deleted || 0)
    ElMessage.success(t('tasks.messages.clearSuccess', { count: n }))
    selected.value = null
    logText.value = ''
    closeStream()
    router.replace({ path: '/tasks', query: {} })
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, t('tasks.messages.clearFailed')))
  } finally {
    clearing.value = false
  }
}

onMounted(() => {
  load().catch(() => ElMessage.error(t('tasks.messages.loadFailed')))
  refreshTimer = setInterval(() => {
    load().catch(() => {})
  }, 5000)
})
onUnmounted(() => {
  closeStream()
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>
<style scoped>
.tasks-page { display: grid; gap: 12px; }
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.topbar h3 { margin: 0; }
.toolbar { margin-bottom: 8px; }
.log-box {
  height: 320px;
  overflow: auto;
  background: #111;
  color: #ddd;
  padding: 8px;
  border-radius: 6px;
}
</style>
