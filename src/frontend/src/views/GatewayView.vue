<template>
  <div class="gateway-page">
    <div class="topbar">
      <h3>Gateway 管理</h3>
      <el-space>
        <el-tag :type="gatewayTagType">状态: {{ gatewayStateText }}</el-tag>
        <el-button :loading="loading" @click="refresh">刷新</el-button>
      </el-space>
    </div>

    <el-alert
      v-if="nvmWarning"
      title="检测到 NVM Node 风险，建议修复"
      type="warning"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12" class="cards">
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">Gateway 状态: {{ gatewayStateText }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">Bind IP: {{ bindText }}</el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <template #header>操作</template>
      <el-space>
        <el-button type="success" :loading="acting==='start'" :disabled="!canOperate || isCoolingDown || !!acting" @click="act('start')">
          {{ actionLabel('启动') }}
        </el-button>
        <el-button type="warning" :loading="acting==='stop'" :disabled="!canOperate || isCoolingDown || !!acting" @click="act('stop')">
          {{ actionLabel('停止') }}
        </el-button>
        <el-button type="primary" :loading="acting==='restart'" :disabled="!canOperate || isCoolingDown || !!acting" @click="act('restart')">
          {{ actionLabel('重启') }}
        </el-button>
      </el-space>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const status = ref<any>({})
const nvmWarning = ref(false)
const loading = ref(false)

const canOperate = computed(() => ['Operator', 'Admin'].includes(auth.user?.role || 'Viewer'))
const acting = ref<'' | 'start' | 'stop' | 'restart'>('')
const cooldownSec = ref(0)
const isCoolingDown = computed(() => cooldownSec.value > 0)

const gatewayStateText = computed(() => String(status.value.active_state || 'unknown'))
const bindText = computed(() => `${status.value.bind_addr || '-'}:${status.value.port || '-'}`)
const gatewayTagType = computed<'success' | 'warning' | 'info'>(() => {
  const s = gatewayStateText.value.toLowerCase()
  if (s === 'active' || s === 'running') return 'success'
  if (s === 'activating' || s === 'reloading') return 'warning'
  return 'info'
})

let timer: any = null
let cooldownTimer: any = null

async function refresh() {
  loading.value = true
  try {
    const { data } = await axios.get('/api/v1/gateway/status')
    status.value = {
      active_state: data?.service?.active_state,
      bind_addr: data?.bind_addr,
      port: data?.port,
    }
    nvmWarning.value = !!data?.nvm_warning
  } finally {
    loading.value = false
  }
}

function startCooldown(seconds = 10) {
  cooldownSec.value = seconds
  if (cooldownTimer) clearInterval(cooldownTimer)
  cooldownTimer = setInterval(() => {
    cooldownSec.value = Math.max(0, cooldownSec.value - 1)
    if (cooldownSec.value <= 0 && cooldownTimer) {
      clearInterval(cooldownTimer)
      cooldownTimer = null
    }
  }, 1000)
}

function actionLabel(base: string): string {
  return isCoolingDown.value ? `${base} (${cooldownSec.value}s)` : base
}

async function act(op: 'start' | 'stop' | 'restart') {
  if (!canOperate.value || isCoolingDown.value || acting.value) return
  acting.value = op
  try {
    await axios.post(`/api/v1/gateway/${op}`)
    startCooldown(10)
    await refresh()
  } finally {
    acting.value = ''
  }
}

onMounted(() => {
  refresh()
  timer = setInterval(refresh, 30000)
})

onUnmounted(() => {
  clearInterval(timer)
  if (cooldownTimer) clearInterval(cooldownTimer)
})
</script>

<style scoped>
.gateway-page {
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
.cards {
  margin: 0;
}
</style>
