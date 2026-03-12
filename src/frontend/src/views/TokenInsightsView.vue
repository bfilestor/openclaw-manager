<template>
  <div class="token-insights-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>{{ t('tokenInsights.pageTitle') }}</span>
          <el-space>
            <el-button @click="goBack">{{ t('tokenInsights.backToOverview') }}</el-button>
            <el-button :loading="loading" @click="loadData">{{ t('common.actions.refresh') }}</el-button>
          </el-space>
        </div>
      </template>

      <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

      <el-row :gutter="12">
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="metric-card">
            <div class="metric-title">{{ t('tokenInsights.outputShareTitle') }}</div>
            <div class="metric-main">{{ outputSharePct }}%</div>
            <div class="metric-sub">{{ t('tokenInsights.outputShareDesc') }}</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="metric-card">
            <div class="metric-title">{{ t('tokenInsights.totalBilledTitle') }}</div>
            <div class="metric-main">{{ formatTokenCompact(totalBilled) }}</div>
            <div class="metric-sub">{{ totalBilled.toLocaleString('en-US') }}</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never">
      <template #header>{{ t('tokenInsights.bucketTitle') }}</template>
      <el-table :data="bucketRows" row-key="name" style="width: 100%">
        <el-table-column prop="name" :label="t('tokenInsights.columns.bucket')" min-width="160" />
        <el-table-column :label="t('tokenInsights.columns.sessions')" width="110">
          <template #default="{ row }">{{ row.sessions }}</template>
        </el-table-column>
        <el-table-column :label="t('tokenInsights.columns.inputTokens')" min-width="140">
          <template #default="{ row }">{{ formatTokenCompact(row.inputTokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('tokenInsights.columns.outputTokens')" min-width="140">
          <template #default="{ row }">{{ formatTokenCompact(row.outputTokens) }}</template>
        </el-table-column>
        <el-table-column :label="t('tokenInsights.columns.outputShare')" min-width="140">
          <template #default="{ row }">{{ row.outputSharePct }}%</template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-alert type="info" :closable="false" show-icon :title="t('tokenInsights.billingTipTitle')" :description="t('tokenInsights.billingTipDesc')" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

type SessionRow = {
  inputTokens: number
  outputTokens: number
  totalTokens: number
}

const loading = ref(false)
const errorMessage = ref('')
const sessions = ref<SessionRow[]>([])
const router = useRouter()
const { t } = useI18n()

const totalInput = computed(() => sessions.value.reduce((sum, s) => sum + Number(s.inputTokens || 0), 0))
const totalOutput = computed(() => sessions.value.reduce((sum, s) => sum + Number(s.outputTokens || 0), 0))
const totalBilled = computed(() => totalInput.value + totalOutput.value)
const outputSharePct = computed(() => {
  if (totalBilled.value <= 0) return '0.0'
  return ((totalOutput.value / totalBilled.value) * 100).toFixed(1)
})

const bucketRows = computed(() => {
  const acc = {
    short: { sessions: 0, inputTokens: 0, outputTokens: 0 },
    long: { sessions: 0, inputTokens: 0, outputTokens: 0 },
  }
  sessions.value.forEach((s) => {
    const input = Number(s.inputTokens || 0)
    const output = Number(s.outputTokens || 0)
    const target = input <= 10000 ? acc.short : acc.long
    target.sessions += 1
    target.inputTokens += input
    target.outputTokens += output
  })

  const mapRow = (name: string, raw: { sessions: number; inputTokens: number; outputTokens: number }) => {
    const billed = raw.inputTokens + raw.outputTokens
    return {
      name,
      sessions: raw.sessions,
      inputTokens: raw.inputTokens,
      outputTokens: raw.outputTokens,
      outputSharePct: billed > 0 ? ((raw.outputTokens / billed) * 100).toFixed(1) : '0.0',
    }
  }

  return [
    mapRow(t('tokenInsights.bucketShort'), acc.short),
    mapRow(t('tokenInsights.bucketLong'), acc.long),
  ]
})

function formatTokenCompact(value: number): string {
  if (!Number.isFinite(value)) return '0'
  const abs = Math.abs(value)
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(abs >= 10_000_000 ? 1 : 2)}M`
  if (abs >= 1_000) return `${(value / 1_000).toFixed(abs >= 10_000 ? 1 : 2)}K`
  return String(Math.round(value))
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    const botsRes = await axios.get('/api/v1/token-usage/summary')
    const bots = Array.isArray(botsRes?.data?.bots) ? botsRes.data.bots : []
    const allSessions: SessionRow[] = []
    for (const bot of bots) {
      const botId = String(bot?.botId || '')
      if (!botId) continue
      const detailRes = await axios.get(`/api/v1/token-usage/bots/${encodeURIComponent(botId)}/conversations`, {
        params: { page: 1, page_size: 100 },
      })
      const items = Array.isArray(detailRes?.data?.items) ? detailRes.data.items : []
      items.forEach((it: any) => {
        allSessions.push({
          inputTokens: Number(it?.inputTokens || 0),
          outputTokens: Number(it?.outputTokens || 0),
          totalTokens: Number(it?.totalTokens || 0),
        })
      })
    }
    sessions.value = allSessions
  } catch (err) {
    errorMessage.value = parseError(err, t('tokenInsights.messages.loadFailed'))
    ElMessage.error(errorMessage.value)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/token-usage')
}

onMounted(loadData)
</script>

<style scoped>
.token-insights-page {
  display: grid;
  gap: 12px;
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.metric-card {
  min-height: 140px;
}

.metric-title {
  font-size: 13px;
  color: var(--oc-text-muted);
}

.metric-main {
  margin-top: 10px;
  font-size: 28px;
  font-weight: 700;
}

.metric-sub {
  margin-top: 8px;
  color: var(--oc-text-muted);
}
</style>
