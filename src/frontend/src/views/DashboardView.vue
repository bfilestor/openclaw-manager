<template>
  <div class="dashboard-page" v-loading="loading" :element-loading-text="t('dashboard.loading')">
    <div class="hero">
      <div>
        <h2>{{ t('dashboard.title') }}</h2>
        <p>{{ t('dashboard.subtitle') }}</p>
      </div>
      <el-tag :type="gatewayTagType" size="large">{{ t('dashboard.gatewayState', { state: gatewayStateText }) }}</el-tag>
    </div>

    <el-alert
      v-if="nvmWarning"
      :title="t('dashboard.nvmWarning')"
      type="warning"
      show-icon
      :closable="false"
    />

    <el-alert
      v-if="quotaAlert"
      :title="quotaAlert"
      :type="quotaAlertType"
      show-icon
      :closable="false"
    />

    <el-card v-if="quotaLimit > 0" shadow="never" class="quota-card">
      <div class="quota-head">
        <span>{{ t('dashboard.quotaProgress') }}</span>
        <el-tag :type="quotaAlertType">{{ Math.round(quotaRatio * 100) }}%</el-tag>
      </div>
      <el-progress :percentage="Math.min(100, Math.round(quotaRatio * 100))" :status="quotaStatus === 'exceeded' ? 'exception' : undefined" />
      <div class="quota-meta">
        <span>{{ t('dashboard.quotaUsed', { used: formatTokenCompact(quotaUsed) }) }}</span>
        <span>{{ t('dashboard.quotaRemaining', { remaining: formatTokenCompact(quotaRemaining) }) }}</span>
        <span>{{ t('dashboard.quotaLimit', { limit: formatTokenCompact(quotaLimit) }) }}</span>
      </div>
    </el-card>

    <el-card shadow="never" class="trend-card">
      <div class="trend-head">
        <span class="trend-icon">📈</span>
        <span class="trend-title">{{ t('dashboard.trendTitle') }}</span>
      </div>
      <el-row :gutter="10">
        <el-col :xs="24" :md="8">
          <div class="trend-item">
            <div class="trend-label">{{ t('dashboard.lastRefreshLabel') }}</div>
            <div class="trend-value">{{ lastRefreshText }}</div>
          </div>
        </el-col>
        <el-col :xs="24" :md="16">
          <div class="trend-item">
            <div class="trend-label">{{ t('dashboard.statusChangeLabel') }}</div>
            <div class="trend-value">{{ statusHint }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="14" class="cards stat-grid">
      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card gateway">
          <div class="stat-head">
            <span class="stat-icon">🚦</span>
            <span class="stat-title">{{ t('dashboard.cards.gateway.title') }}</span>
          </div>
          <div class="stat-main">{{ gatewayStateText }}</div>
          <div class="stat-sub">{{ t('dashboard.cards.gateway.sub') }}</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card bind">
          <div class="stat-head">
            <span class="stat-icon">🌐</span>
            <span class="stat-title">{{ t('dashboard.cards.bind.title') }}</span>
          </div>
          <div class="stat-main monospace">{{ bindText }}</div>
          <div class="stat-sub">{{ t('dashboard.cards.bind.sub') }}</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card skills">
          <div class="stat-head">
            <span class="stat-icon">🧩</span>
            <span class="stat-title">{{ t('dashboard.cards.skills.title') }}</span>
          </div>
          <div class="stat-main">{{ skillCount }}</div>
          <div class="stat-sub">{{ t('dashboard.cards.skills.sub') }}</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card agents">
          <div class="stat-head">
            <span class="stat-icon">🤖</span>
            <span class="stat-title">{{ t('dashboard.cards.agents.title') }}</span>
          </div>
          <div class="stat-main">{{ agentCount }}</div>
          <div class="stat-sub">{{ t('dashboard.cards.agents.sub') }}</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card bots">
          <div class="stat-head">
            <span class="stat-icon">🐧</span>
            <span class="stat-title">{{ t('dashboard.cards.bots.title') }}</span>
          </div>
          <div class="stat-main">{{ botCount }}</div>
          <div class="stat-sub">{{ t('dashboard.cards.bots.sub') }}</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card users">
          <div class="stat-head">
            <span class="stat-icon">👥</span>
            <span class="stat-title">{{ t('dashboard.cards.users.title') }}</span>
          </div>
          <div class="stat-main">{{ userCount }}</div>
          <div class="stat-sub">{{ t('dashboard.cards.users.sub') }}</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'
import { useI18n } from 'vue-i18n'

const loading = ref(false)
const status = ref<any>({})
const CACHE_KEY = 'openclaw_manager_dashboard_cache_v1'
const nvmWarning = ref(false)
const skillCount = ref(0)
const agentCount = ref(0)
const botCount = ref(0)
const userCount = ref(0)
const lastRefreshAt = ref('')
const statusHint = ref('')
const previousGatewayState = ref('')
const { t } = useI18n()
const quotaStatus = ref<'normal' | 'near' | 'exceeded' | ''>('')
const quotaUsed = ref(0)
const quotaLimit = ref(0)

let timer: any = null

const gatewayStateText = computed(() => String(status.value.active_state || 'unknown'))
const bindText = computed(() => {
  const host = status.value.bind_addr || '-'
  const port = status.value.port || '-'
  return `${host}:${port}`
})
const lastRefreshText = computed(() => {
  if (!lastRefreshAt.value) return t('common.emptyValue')
  const d = new Date(lastRefreshAt.value)
  if (Number.isNaN(d.getTime())) return lastRefreshAt.value
  return d.toLocaleString()
})

const gatewayTagType = computed<'success' | 'warning' | 'info'>(() => {
  const s = gatewayStateText.value.toLowerCase()
  if (s === 'active' || s === 'running') return 'success'
  if (s === 'activating' || s === 'reloading') return 'warning'
  return 'info'
})

const quotaAlertType = computed<'warning' | 'error'>(() => (quotaStatus.value === 'exceeded' ? 'error' : 'warning'))
const quotaRatio = computed(() => {
  if (!quotaLimit.value) return 0
  return quotaUsed.value / quotaLimit.value
})
const quotaRemaining = computed(() => Math.max(0, quotaLimit.value - quotaUsed.value))
const quotaAlert = computed(() => {
  if (!quotaLimit.value || quotaStatus.value === 'normal' || quotaStatus.value === '') return ''
  if (quotaStatus.value === 'exceeded') {
    return t('dashboard.quotaExceeded', { used: formatTokenCompact(quotaUsed.value), limit: formatTokenCompact(quotaLimit.value) })
  }
  return t('dashboard.quotaNear', { used: formatTokenCompact(quotaUsed.value), limit: formatTokenCompact(quotaLimit.value) })
})

function formatTokenCompact(value: number): string {
  if (!Number.isFinite(value)) return '0'
  const abs = Math.abs(value)
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(abs >= 10_000_000 ? 0 : 1)}M`
  if (abs >= 1_000) return `${(value / 1_000).toFixed(abs >= 10_000 ? 0 : 1)}K`
  return String(Math.round(value))
}

function countBotsFromConfig(cfg: any): number {
  const channels = cfg?.channels
  if (!channels || typeof channels !== 'object') return 0

  let total = 0
  for (const key of Object.keys(channels)) {
    const ch = channels[key]
    if (!ch || typeof ch !== 'object') continue

    if (ch.enabled !== false) total += 1

    const accounts = ch.accounts
    if (accounts && typeof accounts === 'object') {
      total += Object.keys(accounts).length
    }
  }
  return total
}

function loadCache() {
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return
    const cached = JSON.parse(raw)
    status.value = cached.status || {}
    nvmWarning.value = !!cached.nvmWarning
    skillCount.value = Number(cached.skillCount || 0)
    agentCount.value = Number(cached.agentCount || 0)
    botCount.value = Number(cached.botCount || 0)
    userCount.value = Number(cached.userCount || 0)
    lastRefreshAt.value = String(cached.lastRefreshAt || '')
    statusHint.value = String(cached.statusHint || '')
    previousGatewayState.value = String(cached.previousGatewayState || '')
    quotaStatus.value = String(cached.quotaStatus || '') as 'normal' | 'near' | 'exceeded' | ''
    quotaUsed.value = Number(cached.quotaUsed || 0)
    quotaLimit.value = Number(cached.quotaLimit || 0)
  } catch {
    // ignore cache parse errors
  }
}

function saveCache() {
  try {
    localStorage.setItem(CACHE_KEY, JSON.stringify({
      status: status.value,
      nvmWarning: nvmWarning.value,
      skillCount: skillCount.value,
      agentCount: agentCount.value,
      botCount: botCount.value,
      userCount: userCount.value,
      lastRefreshAt: lastRefreshAt.value,
      statusHint: statusHint.value,
      previousGatewayState: previousGatewayState.value,
      quotaStatus: quotaStatus.value,
      quotaUsed: quotaUsed.value,
      quotaLimit: quotaLimit.value,
    }))
  } catch {
    // ignore cache write errors
  }
}

async function refresh() {
  const firstLoad = !lastRefreshAt.value
  if (firstLoad) loading.value = true
  try {
    const [gatewayRes, skillsRes, agentsRes, configRes, usersRes, tokenUsageRes] = await Promise.all([
      axios.get('/api/v1/gateway/status'),
      axios.get('/api/v1/skills', { params: { scope: 'global' } }),
      axios.get('/api/v1/agents'),
      axios.get('/api/v1/config/openclaw'),
      axios.get('/api/v1/users').catch(() => ({ data: { users: [] } })),
      axios.get('/api/v1/token-usage/summary').catch(() => ({ data: {} })),
    ])

    const gd = gatewayRes.data
    const nextState = String(gd?.service?.active_state || 'unknown')
    status.value = {
      active_state: nextState,
      bind_addr: gd?.bind_addr,
      port: gd?.port,
    }
    nvmWarning.value = !!gd?.nvm_warning

    if (!previousGatewayState.value) {
      statusHint.value = t('dashboard.status.current', { state: nextState })
    } else if (previousGatewayState.value !== nextState) {
      statusHint.value = t('dashboard.status.changed', { from: previousGatewayState.value, to: nextState })
    } else {
      statusHint.value = t('dashboard.status.stable', { state: nextState })
    }
    previousGatewayState.value = nextState

    const skills = skillsRes.data?.skills
    skillCount.value = Array.isArray(skills) ? skills.length : 0

    const agents = agentsRes.data?.agents
    agentCount.value = Array.isArray(agents) ? agents.length : 0

    let cfg: any = {}
    try {
      cfg = JSON.parse(String(configRes.data?.content || '{}'))
    } catch {
      cfg = {}
    }
    botCount.value = countBotsFromConfig(cfg)

    const users = usersRes.data?.users
    userCount.value = Array.isArray(users) ? users.length : 0

    const quota = tokenUsageRes.data?.quota
    quotaStatus.value = quota?.status === 'near' || quota?.status === 'exceeded' ? quota.status : 'normal'
    quotaUsed.value = Number(quota?.usedTokens || 0)
    quotaLimit.value = Number(quota?.tokenLimit || 0)

    lastRefreshAt.value = new Date().toISOString()
    saveCache()
  } catch {
    // Keep silent to avoid interrupting dashboard rendering
  } finally {
    if (firstLoad) loading.value = false
  }
}

onMounted(() => {
  statusHint.value = t('dashboard.status.waiting')
  loadCache()
  if (!statusHint.value) statusHint.value = t('dashboard.status.waiting')
  refresh()
  timer = setInterval(refresh, 30000)
})
onUnmounted(() => clearInterval(timer))
</script>

<style scoped>
.dashboard-page {
  display: grid;
  gap: 12px;
}

.hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-radius: 14px;
  padding: 16px 18px;
  background: linear-gradient(135deg, #1d4ed8 0%, #3b82f6 40%, #22c1ff 100%);
  color: #fff;
}
.hero h2 {
  margin: 0;
  font-size: 22px;
  line-height: 1.2;
}
.hero p {
  margin: 6px 0 0;
  opacity: 0.92;
}

.quota-card,
.trend-card {
  border-radius: 12px;
  background: linear-gradient(145deg, #f8fafc, #eef2ff);
}

.quota-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.quota-meta {
  margin-top: 8px;
  display: flex;
  gap: 14px;
  color: #4b5563;
  font-size: 12px;
}
.trend-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.trend-icon {
  width: 30px;
  height: 30px;
  border-radius: 9px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #e0e7ff;
}
.trend-title {
  font-weight: 600;
}
.trend-item {
  background: rgba(255, 255, 255, 0.7);
  border-radius: 10px;
  padding: 10px;
}
.trend-label {
  font-size: 12px;
  color: #6b7280;
}
.trend-value {
  margin-top: 4px;
  font-weight: 600;
  color: #1f2937;
}

.cards {
  margin: 0;
}

.stat-grid {
  row-gap: 14px;
}

.stat-card {
  border-radius: 12px;
  min-height: 148px;
  height: 100%;
}
.stat-card :deep(.el-card__body) {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.stat-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.stat-icon {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.5);
}
.stat-title {
  font-size: 14px;
  color: #2e3440;
}
.stat-main {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
  color: #111827;
}
.stat-sub {
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

.monospace {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 22px;
}

.gateway {
  background: linear-gradient(145deg, #ecfdf5, #d1fae5);
}
.bind {
  background: linear-gradient(145deg, #eff6ff, #dbeafe);
}
.skills {
  background: linear-gradient(145deg, #f5f3ff, #ede9fe);
}
.agents {
  background: linear-gradient(145deg, #fff7ed, #ffedd5);
}
.bots {
  background: linear-gradient(145deg, #fdf2f8, #fce7f3);
}
.users {
  background: linear-gradient(145deg, #f0fdf4, #dcfce7);
}
</style>
