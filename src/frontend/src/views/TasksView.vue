<template>
  <div class="tasks-page">
    <h3>Tasks</h3>
    <el-row :gutter="16">
      <el-col :xs="24" :md="10" :lg="8">
        <el-card shadow="never">
          <template #header>任务列表</template>
          <el-table
            ref="tableRef"
            :data="tasks"
            row-key="task_id"
            style="width: 100%"
            highlight-current-row
            @row-click="(row) => select(row)"
          >
            <el-table-column prop="task_id" label="Task ID" min-width="220" />
            <el-table-column prop="task_type" label="类型" min-width="110" />
            <el-table-column label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'FAILED' ? 'danger' : row.status === 'SUCCEEDED' ? 'success' : 'info'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="14" :lg="16">
        <el-card shadow="never">
          <template #header>
            日志 {{ selected?.task_id ? `(Task: ${selected.task_id})` : '' }}
          </template>
          <el-space class="toolbar">
            <el-checkbox v-model="autoScroll">自动滚动</el-checkbox>
            <el-input v-model="keyword" placeholder="搜索日志" clearable />
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
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const tasks = ref<any[]>([])
const selected = ref<any>(null)
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
  const { data } = await axios.get('/api/v1/tasks')
  tasks.value = Array.isArray(data?.tasks) ? data.tasks : []
  await focusTaskFromQuery()
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

async function select(t: any, syncQuery = true) {
  selected.value = t
  logText.value = ''
  closeStream()

  if (syncQuery) {
    router.replace({ path: '/tasks', query: { task_id: String(t.task_id || '') } })
  }

  const token = String(auth.accessToken || '').trim()
  if (!token) {
    appendLog('未找到 access token，无法订阅任务日志。')
    return
  }

  const url = `/api/v1/tasks/${encodeURIComponent(t.task_id)}/events?token=${encodeURIComponent(token)}`
  eventSource = new EventSource(url)
  eventSource.onmessage = (ev) => {
    try {
      const payload = JSON.parse(ev.data)
      if (payload?.line) {
        appendLog(String(payload.line))
      } else if (payload?.type === 'done') {
        appendLog(`done: status=${payload.status} exit_code=${payload.exit_code}`)
      } else {
        appendLog(ev.data)
      }
    } catch {
      appendLog(ev.data)
    }
  }
  eventSource.onerror = () => {
    appendLog('日志流已结束。')
    closeStream()
  }
}

watch(() => route.query.task_id, () => {
  void focusTaskFromQuery()
})

onMounted(() => {
  load().catch(() => ElMessage.error('加载任务列表失败'))
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
