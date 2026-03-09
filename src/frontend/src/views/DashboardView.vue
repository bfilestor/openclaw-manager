<template>
  <div class="dashboard-page">
    <div class="hero">
      <div>
        <h2>OpenClaw Dashboard</h2>
        <p>系统总览入口 · 一眼看到关键状态</p>
      </div>
      <el-tag :type="gatewayTagType" size="large">Gateway: {{ gatewayStateText }}</el-tag>
    </div>

    <el-alert
      v-if="nvmWarning"
      title="检测到 NVM Node 风险，建议修复"
      type="warning"
      show-icon
      :closable="false"
    />

    <el-row :gutter="14" class="cards">
      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card gateway">
          <div class="stat-head">
            <span class="stat-icon">🚦</span>
            <span class="stat-title">Gateway 状态</span>
          </div>
          <div class="stat-main">{{ gatewayStateText }}</div>
          <div class="stat-sub">服务运行状态实时刷新</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card bind">
          <div class="stat-head">
            <span class="stat-icon">🌐</span>
            <span class="stat-title">Bind 信息</span>
          </div>
          <div class="stat-main monospace">{{ bindText }}</div>
          <div class="stat-sub">Gateway 监听地址</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="8">
        <el-card shadow="hover" class="stat-card skills">
          <div class="stat-head">
            <span class="stat-icon">🧩</span>
            <span class="stat-title">Skills 数量</span>
          </div>
          <div class="stat-main">{{ skillCount }}</div>
          <div class="stat-sub">当前已安装技能总数</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="12">
        <el-card shadow="hover" class="stat-card agents">
          <div class="stat-head">
            <span class="stat-icon">🤖</span>
            <span class="stat-title">Agent 数量</span>
          </div>
          <div class="stat-main">{{ agentCount }}</div>
          <div class="stat-sub">当前系统可用 Agent</div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :lg="12">
        <el-card shadow="hover" class="stat-card bots">
          <div class="stat-head">
            <span class="stat-icon">🐧</span>
            <span class="stat-title">Bot 数量</span>
          </div>
          <div class="stat-main">{{ botCount }}</div>
          <div class="stat-sub">按 channels + accounts 聚合</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'

const status = ref<any>({})
const nvmWarning = ref(false)
const skillCount = ref(0)
const agentCount = ref(0)
const botCount = ref(0)

let timer: any = null

const gatewayStateText = computed(() => String(status.value.active_state || 'unknown'))
const bindText = computed(() => {
  const host = status.value.bind_addr || '-'
  const port = status.value.port || '-'
  return `${host}:${port}`
})
const gatewayTagType = computed<'success' | 'warning' | 'info'>(() => {
  const s = gatewayStateText.value.toLowerCase()
  if (s === 'active' || s === 'running') return 'success'
  if (s === 'activating' || s === 'reloading') return 'warning'
  return 'info'
})

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

async function refresh() {
  try {
    const [gatewayRes, skillsRes, agentsRes, configRes] = await Promise.all([
      axios.get('/api/v1/gateway/status'),
      axios.get('/api/v1/skills', { params: { scope: 'global' } }),
      axios.get('/api/v1/agents'),
      axios.get('/api/v1/config/openclaw'),
    ])

    const gd = gatewayRes.data
    status.value = {
      active_state: gd?.service?.active_state,
      bind_addr: gd?.bind_addr,
      port: gd?.port,
    }
    nvmWarning.value = !!gd?.nvm_warning

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
  } catch {
    // 静默失败，避免打断 Dashboard 展示
  }
}

onMounted(() => {
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

.cards {
  margin: 0;
}

.stat-card {
  border-radius: 12px;
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
</style>
