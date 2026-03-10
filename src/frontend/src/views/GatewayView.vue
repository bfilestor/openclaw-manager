<template>
  <div class="gateway-page">
    <div class="topbar">
      <h3>{{ t('gateway.title') }}</h3>
      <el-space>
        <el-button :loading="loading" @click="refresh">{{ t('common.actions.refresh') }}</el-button>
        <el-button :loading="diagnosing" @click="runDiagnose">{{ t('gateway.diagnose') }}</el-button>
      </el-space>
    </div>

    <el-alert
      v-if="nvmWarning"
      :title="t('gateway.nvmWarning')"
      type="warning"
      show-icon
      :closable="false"
    />

    <el-alert
      v-if="statusError"
      :title="t('gateway.connectFailed', { reason: statusError })"
      type="error"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12" class="cards">
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">{{ t('gateway.gatewayState', { state: gatewayStateText }) }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">{{ t('gateway.bindIp', { bind: bindText }) }}</el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <template #header>{{ t('gateway.operations') }}</template>
      <el-space>
        <el-button type="success" :loading="acting==='start'" :disabled="!canOperate || isCoolingDown || !!acting" @click="act('start')">
          {{ actionLabel('start') }}
        </el-button>
        <el-button type="warning" :loading="acting==='stop'" :disabled="!canOperate || isCoolingDown || !!acting" @click="act('stop')">
          {{ actionLabel('stop') }}
        </el-button>
        <el-button type="primary" :loading="acting==='restart'" :disabled="!canOperate || isCoolingDown || !!acting" @click="act('restart')">
          {{ actionLabel('restart') }}
        </el-button>
      </el-space>
    </el-card>

    <el-dialog v-model="diagnoseVisible" :title="t('gateway.diagnoseResultTitle')" width="860px">
      <el-alert :title="diagnoseSummary" type="info" show-icon :closable="false" />
      <el-scrollbar height="320px">
        <pre class="diagnose-box">{{ diagnoseLogs }}</pre>
      </el-scrollbar>
      <template #footer>
        <el-button @click="diagnoseVisible = false">{{ t('common.actions.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const { t } = useI18n()
const status = ref<any>({})
const nvmWarning = ref(false)
const loading = ref(false)
const statusError = ref('')
const diagnosing = ref(false)
const diagnoseVisible = ref(false)
const diagnoseLogs = ref('')
const diagnoseSummary = ref('')

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

function classifyGatewayError(err: any): string {
  const httpStatus = Number(err?.response?.status || 0)
  if (httpStatus === 401 || httpStatus === 403) return t('gateway.errorReasons.authDenied')
  if (httpStatus === 404) return t('gateway.errorReasons.apiNotFound')
  if (httpStatus >= 500) return t('gateway.errorReasons.managerInternal')

  const msg = String(err?.response?.data?.message || err?.response?.data?.error || err?.message || '').toLowerCase()
  if (msg.includes('timed out') || msg.includes('timeout')) return t('gateway.errorReasons.gatewayTimeout')
  if (msg.includes('connection refused') || msg.includes('connect: no such file') || msg.includes('dial tcp')) {
    return t('gateway.errorReasons.gatewayDown')
  }
  if (msg.includes('systemctl') || msg.includes('unit') || msg.includes('service')) {
    return t('gateway.errorReasons.systemdIssue')
  }
  if (msg.includes('network error') || msg.includes('failed to fetch')) return t('gateway.errorReasons.browserNetwork')
  return t('gateway.errorReasons.unknown')
}

async function refresh() {
  loading.value = true
  statusError.value = ''
  try {
    const { data } = await axios.get('/api/v1/gateway/status')
    status.value = {
      active_state: data?.service?.active_state,
      bind_addr: data?.bind_addr,
      port: data?.port,
    }
    nvmWarning.value = !!data?.nvm_warning
  } catch (err) {
    status.value = { active_state: 'unknown', bind_addr: '-', port: '-' }
    statusError.value = classifyGatewayError(err)
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

function actionLabel(op: 'start' | 'stop' | 'restart'): string {
  if (isCoolingDown.value) {
    return t(`gateway.actions.${op}WithCooldown`, { seconds: cooldownSec.value })
  }
  return t(`gateway.actions.${op}`)
}

async function act(op: 'start' | 'stop' | 'restart') {
  if (!canOperate.value || isCoolingDown.value || acting.value) return
  acting.value = op
  try {
    await axios.post(`/api/v1/gateway/${op}`)
    startCooldown(10)
    await refresh()
  } catch (err) {
    statusError.value = classifyGatewayError(err)
  } finally {
    acting.value = ''
  }
}

async function runDiagnose() {
  diagnosing.value = true
  diagnoseLogs.value = ''
  diagnoseSummary.value = ''
  try {
    const [doctorResp, logsResp] = await Promise.all([
      axios.post('/api/v1/gateway/doctor'),
      axios.get('/api/v1/gateway/logs', { params: { source: 'journald', lines: 120 } }),
    ])
    const nvmDetected = !!doctorResp?.data?.nvm_detected
    diagnoseSummary.value = nvmDetected ? t('gateway.diagnoseSummaryNvm') : t('gateway.diagnoseSummaryOk')
    const logs = Array.isArray(logsResp?.data?.logs) ? logsResp.data.logs : []
    diagnoseLogs.value = logs.join('\n') || t('gateway.noLogs')
    diagnoseVisible.value = true
  } catch (err) {
    ElMessage.error(t('gateway.diagnoseFailed', { reason: classifyGatewayError(err) }))
  } finally {
    diagnosing.value = false
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

.diagnose-box {
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.45;
  color: #374151;
  padding: 8px;
}
</style>
