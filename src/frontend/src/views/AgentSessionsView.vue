<template>
  <div class="sessions-page">
    <div class="topbar">
      <h3>Agent Sessions</h3>
      <div class="topbar-actions">
        <el-input
          v-model="agentFilter"
          clearable
          placeholder="按 Agent ID 过滤（如 researcher）"
          class="agent-filter"
          @keyup.enter="loadSessions"
        />
        <el-button :loading="loading" @click="loadSessions">刷新</el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12" class="stats-row">
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">总会话: {{ sessions.length }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">运行中: {{ runningCount }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">等待中: {{ waitingCount }}</el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!loading && sessions.length === 0" description="当前没有会话数据" />

    <div v-else class="session-grid">
      <el-card
        v-for="session in sessions"
        :key="session.id"
        shadow="hover"
        class="session-card"
        :class="`status-${normalizeStatus(session.status)}`"
      >
        <div class="session-header">
          <div class="session-title-row">
            <span class="status-emoji" :class="`emoji-${normalizeStatus(session.status)}`" aria-hidden="true">
              {{ statusMeta(session.status).emoji }}
            </span>
            <div class="session-title-wrap">
              <div class="session-id" :title="session.id">{{ session.id }}</div>
              <div class="session-agent">Agent: {{ session.agentId || 'unknown' }}</div>
            </div>
          </div>
          <el-tag :type="statusMeta(session.status).tagType">{{ normalizeStatus(session.status) }}</el-tag>
        </div>

        <div class="session-body">
          <div class="status-line">{{ statusMeta(session.status).text }}</div>
          <div class="time-line">创建时间: {{ formatTime(session.createdAt) }}</div>
          <div class="time-line">最近活动: {{ formatTime(session.lastActivity) }}</div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

type SessionItem = {
  id: string
  agentId: string
  status: string
  createdAt: string
  lastActivity: string
}

const loading = ref(false)
const errorMessage = ref('')
const sessions = ref<SessionItem[]>([])
const agentFilter = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const runningCount = computed(() => sessions.value.filter((item) => normalizeStatus(item.status) === 'running').length)
const waitingCount = computed(() => sessions.value.filter((item) => normalizeStatus(item.status) === 'waiting').length)

function normalizeStatus(status: string) {
  const raw = String(status || '').toLowerCase().trim()
  if (raw === 'running') return 'running'
  if (raw === 'waiting') return 'waiting'
  if (raw === 'completed') return 'completed'
  if (raw === 'failed') return 'failed'
  return 'unknown'
}

function statusMeta(status: string) {
  const normalized = normalizeStatus(status)
  if (normalized === 'running') return { emoji: '🏃', text: 'Agent 正在火力全开搬砖中...', tagType: 'success' as const }
  if (normalized === 'waiting') return { emoji: '⏳', text: 'Agent 在排队等任务，先摸会儿鱼~', tagType: 'warning' as const }
  if (normalized === 'completed') return { emoji: '🎉', text: '任务搞定，Agent 正在等夸夸！', tagType: 'info' as const }
  if (normalized === 'failed') return { emoji: '💥', text: '翻车了，Agent 申请一次复活机会。', tagType: 'danger' as const }
  return { emoji: '🫥', text: '状态未知，Agent 看起来有点迷茫。', tagType: 'info' as const }
}

function formatTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

async function loadSessions() {
  loading.value = true
  errorMessage.value = ''
  try {
    const params = agentFilter.value.trim() ? { agentId: agentFilter.value.trim() } : undefined
    const { data } = await axios.get('/api/sessions', { params })
    sessions.value = Array.isArray(data?.sessions) ? data.sessions : []
  } catch {
    sessions.value = []
    errorMessage.value = '加载 Agent Session 失败，请检查 API 状态'
    ElMessage.error('加载 Agent Session 失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSessions()
  timer = setInterval(loadSessions, 15000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.sessions-page {
  display: grid;
  gap: 12px;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.topbar h3 {
  margin: 0;
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.agent-filter {
  width: 280px;
}

.stats-row {
  margin: 0;
}

.session-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
}

.session-card {
  border: 1px solid #e5e9f2;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.session-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(31, 64, 122, 0.12);
}

.session-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.session-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.status-emoji {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  background: #f1f4fb;
}

.session-title-wrap {
  min-width: 0;
}

.session-id {
  font-weight: 600;
  color: #273041;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 210px;
}

.session-agent {
  color: #5f6b7a;
  font-size: 12px;
}

.session-body {
  margin-top: 10px;
  display: grid;
  gap: 6px;
}

.status-line {
  color: #1f2f46;
  font-size: 14px;
}

.time-line {
  color: #657286;
  font-size: 12px;
}

.status-running .status-emoji {
  background: linear-gradient(135deg, #def8e9, #e7fff5);
}

.status-waiting .status-emoji {
  background: linear-gradient(135deg, #fff3dc, #fff9eb);
}

.status-completed .status-emoji {
  background: linear-gradient(135deg, #e8f1ff, #f2f7ff);
}

.status-failed .status-emoji {
  background: linear-gradient(135deg, #ffe4e6, #fff0f1);
}

.emoji-running {
  animation: bounceRun 0.9s ease-in-out infinite;
}

.emoji-waiting {
  animation: swingWait 1.4s ease-in-out infinite;
}

.emoji-completed {
  animation: popDone 1.4s ease-in-out infinite;
}

.emoji-failed {
  animation: shakeFail 0.7s ease-in-out infinite;
}

@keyframes bounceRun {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

@keyframes swingWait {
  0%,
  100% {
    transform: rotate(-7deg);
  }
  50% {
    transform: rotate(7deg);
  }
}

@keyframes popDone {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.14);
  }
}

@keyframes shakeFail {
  0%,
  100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-2px);
  }
  75% {
    transform: translateX(2px);
  }
}

@media (max-width: 900px) {
  .topbar {
    flex-direction: column;
    align-items: stretch;
  }

  .topbar-actions {
    width: 100%;
  }

  .agent-filter {
    width: 100%;
  }
}
</style>
