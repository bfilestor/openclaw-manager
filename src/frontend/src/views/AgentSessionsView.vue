<template>
  <div class="sessions-page">
    <div class="topbar">
      <h3>Agent Office Live</h3>
      <div class="topbar-actions">
        <el-input
          v-model="agentFilter"
          clearable
          placeholder="按 Agent ID 过滤（如 main / xcoder）"
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
        <el-card shadow="never">总 Agent: {{ officeAgents.length }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">在线中: {{ onlineCount }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card shadow="never">最后刷新: {{ lastRefreshText }}</el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="office-board">
      <template #header>
        <div class="office-title">像素办公室 · 状态实时可视化</div>
      </template>

      <div class="office-grid">
        <section
          v-for="zone in officeZones"
          :key="zone.state"
          class="zone"
          :class="`zone-${zone.state}`"
        >
          <div class="zone-header">
            <span class="zone-emoji" aria-hidden="true">{{ zone.emoji }}</span>
            <div>
              <div class="zone-name">{{ zone.label }}</div>
              <div class="zone-desc">{{ zone.desc }}</div>
            </div>
            <el-tag size="small">{{ agentsByState[zone.state]?.length || 0 }}</el-tag>
          </div>

          <div v-if="(agentsByState[zone.state]?.length || 0) === 0" class="zone-empty">
            目前这个区域没人，安静如鸡 🐥
          </div>

          <div v-else class="agent-list">
            <article
              v-for="agent in agentsByState[zone.state]"
              :key="agent.id"
              class="agent-chip"
            >
              <div class="agent-bubble">{{ agent.bubble }}</div>
              <div class="agent-avatar" :class="`anim-${zone.state}`">{{ zone.emoji }}</div>
              <div class="agent-meta">
                <div class="agent-name">{{ agent.agentId }}</div>
                <div class="agent-time">{{ formatTime(agent.lastActivity) }}</div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </el-card>
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

type VisualState = 'idle' | 'writing' | 'researching' | 'executing' | 'syncing' | 'error'

type OfficeAgent = {
  id: string
  agentId: string
  visualState: VisualState
  lastActivity: string
  bubble: string
}

const loading = ref(false)
const errorMessage = ref('')
const sessions = ref<SessionItem[]>([])
const officeAgents = ref<OfficeAgent[]>([])
const agentFilter = ref('')
const lastRefreshAt = ref<Date | null>(null)
let timer: ReturnType<typeof setInterval> | null = null

const officeZones: Array<{ state: VisualState; label: string; emoji: string; desc: string }> = [
  { state: 'idle', label: '休息区', emoji: '🛋️', desc: '待命 / 暂无任务' },
  { state: 'writing', label: '文档区', emoji: '📝', desc: '写文档 / 产出内容' },
  { state: 'researching', label: '调研区', emoji: '🔎', desc: '查资料 / 搜证据' },
  { state: 'executing', label: '执行区', emoji: '⚙️', desc: '跑任务 / 执行中' },
  { state: 'syncing', label: '同步区', emoji: '☁️', desc: '同步状态 / 收尾' },
  { state: 'error', label: '排障区', emoji: '🚨', desc: '故障 / 需要关注' },
]

const bubbleTexts: Record<VisualState, string[]> = {
  idle: ['待命中，随叫随到～', '先喝口水，等任务', '我在，随时开工'],
  writing: ['文档写得飞起 ✍️', '正在整理关键结论', '边写边打磨表达'],
  researching: ['正在翻资料找证据', '我在检索关键线索', '查完这个就给你结论'],
  executing: ['任务执行中，别眨眼', '脚本正在疯狂搬砖', '我在跑流程，稳住'],
  syncing: ['状态同步中...', '收尾打包上传中', '进度正在对齐'],
  error: ['这里有异常，正在排查', '检测到故障，先止血', '警报拉响，马上修'],
}

const agentsByState = computed<Record<VisualState, OfficeAgent[]>>(() => {
  const grouped: Record<VisualState, OfficeAgent[]> = {
    idle: [],
    writing: [],
    researching: [],
    executing: [],
    syncing: [],
    error: [],
  }
  for (const item of officeAgents.value) grouped[item.visualState].push(item)
  return grouped
})

const onlineCount = computed(() => officeAgents.value.filter((item) => item.visualState !== 'idle').length)

const lastRefreshText = computed(() => {
  if (!lastRefreshAt.value) return '-'
  return lastRefreshAt.value.toLocaleTimeString()
})

function normalizeStatus(status: string) {
  const raw = String(status || '').toLowerCase().trim()
  if (raw === 'running') return 'running'
  if (raw === 'waiting') return 'waiting'
  if (raw === 'completed') return 'completed'
  if (raw === 'failed') return 'failed'
  return 'unknown'
}

function statusToVisualState(status: string): VisualState {
  const normalized = normalizeStatus(status)
  if (normalized === 'running') return 'executing'
  if (normalized === 'waiting') return 'idle'
  if (normalized === 'completed') return 'syncing'
  if (normalized === 'failed') return 'error'
  return 'researching'
}

function pickBubble(state: VisualState) {
  const words = bubbleTexts[state]
  return words[Math.floor(Math.random() * words.length)]
}

function formatTime(value: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function buildOfficeAgents(rows: SessionItem[]) {
  const map = new Map<string, SessionItem>()
  for (const row of rows) {
    const old = map.get(row.agentId)
    if (!old) {
      map.set(row.agentId, row)
      continue
    }
    const oldTs = new Date(old.lastActivity || old.createdAt || 0).getTime()
    const newTs = new Date(row.lastActivity || row.createdAt || 0).getTime()
    if (newTs >= oldTs) map.set(row.agentId, row)
  }

  officeAgents.value = Array.from(map.values()).map((row) => {
    const visualState = statusToVisualState(row.status)
    return {
      id: row.id,
      agentId: row.agentId || 'unknown',
      visualState,
      lastActivity: row.lastActivity || row.createdAt,
      bubble: pickBubble(visualState),
    }
  })
}

async function loadSessions() {
  loading.value = true
  errorMessage.value = ''
  try {
    const params = agentFilter.value.trim() ? { agentId: agentFilter.value.trim() } : undefined
    const { data } = await axios.get('/api/v1/agent-sessions', { params })
    sessions.value = Array.isArray(data?.sessions) ? data.sessions : []
    buildOfficeAgents(sessions.value)
    lastRefreshAt.value = new Date()
  } catch {
    sessions.value = []
    officeAgents.value = []
    errorMessage.value = '加载 Agent Session 失败，请检查 API 状态'
    ElMessage.error('加载 Agent Session 失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSessions()
  timer = setInterval(loadSessions, 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.sessions-page { display: grid; gap: 12px; }
.topbar { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.topbar h3 { margin: 0; }
.topbar-actions { display: flex; align-items: center; gap: 10px; }
.agent-filter { width: 320px; }
.stats-row { margin: 0; }
.office-title { font-weight: 700; }
.office-board { background: linear-gradient(180deg, #fcfdff 0%, #f5f8ff 100%); }

.office-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.zone {
  border: 1px solid #e4e8f2;
  border-radius: 12px;
  padding: 10px;
  min-height: 168px;
  background: #fff;
}

.zone-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.zone-emoji { font-size: 20px; }
.zone-name { font-weight: 700; color: #2b3344; }
.zone-desc { font-size: 12px; color: #69778c; }
.zone-empty { font-size: 12px; color: #8a96aa; padding: 12px 6px; }
.agent-list { display: grid; gap: 10px; }

.agent-chip {
  position: relative;
  border: 1px dashed #dbe2f2;
  border-radius: 10px;
  padding: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.agent-bubble {
  position: absolute;
  top: -18px;
  left: 8px;
  background: #fff;
  border: 1px solid #d8e0ef;
  border-radius: 999px;
  font-size: 11px;
  color: #50607a;
  padding: 2px 8px;
}

.agent-avatar {
  width: 30px;
  height: 30px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #eef3ff;
  font-size: 16px;
}

.agent-name { font-weight: 600; color: #2a344a; }
.agent-time { font-size: 12px; color: #76839a; }

.anim-idle { animation: idleFloat 2.2s ease-in-out infinite; }
.anim-writing { animation: writingTick 1.1s ease-in-out infinite; }
.anim-researching { animation: searchSwing 1.6s ease-in-out infinite; }
.anim-executing { animation: executeRun 0.9s ease-in-out infinite; }
.anim-syncing { animation: syncPulse 1.4s ease-in-out infinite; }
.anim-error { animation: errorShake 0.6s ease-in-out infinite; }

@keyframes idleFloat { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-3px)} }
@keyframes writingTick { 0%,100%{transform:rotate(0)} 50%{transform:rotate(-8deg)} }
@keyframes searchSwing { 0%,100%{transform:rotate(-8deg)} 50%{transform:rotate(8deg)} }
@keyframes executeRun { 0%,100%{transform:translateX(0)} 50%{transform:translateX(3px)} }
@keyframes syncPulse { 0%,100%{transform:scale(1)} 50%{transform:scale(1.14)} }
@keyframes errorShake { 0%,100%{transform:translateX(0)} 25%{transform:translateX(-2px)} 75%{transform:translateX(2px)} }

.zone-idle { background: linear-gradient(180deg, #fff 0%, #f6f9ff 100%); }
.zone-writing { background: linear-gradient(180deg, #fff 0%, #f7fbff 100%); }
.zone-researching { background: linear-gradient(180deg, #fff 0%, #f8fcff 100%); }
.zone-executing { background: linear-gradient(180deg, #fff 0%, #f7fff9 100%); }
.zone-syncing { background: linear-gradient(180deg, #fff 0%, #f6fbff 100%); }
.zone-error { background: linear-gradient(180deg, #fff 0%, #fff6f6 100%); }

@media (max-width: 1100px) {
  .office-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

@media (max-width: 900px) {
  .topbar { flex-direction: column; align-items: stretch; }
  .topbar-actions { width: 100%; }
  .agent-filter { width: 100%; }
  .office-grid { grid-template-columns: 1fr; }
}
</style>
