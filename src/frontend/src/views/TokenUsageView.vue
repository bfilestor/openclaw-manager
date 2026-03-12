<template>
  <div class="token-usage-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>{{ t('tokenUsage.pageTitle') }}</span>
          <div class="header-tools">
            <el-select v-model="days" class="days-select" @change="loadData">
              <el-option :label="t('tokenUsage.range.all')" :value="0" />
              <el-option :label="t('tokenUsage.range.today')" :value="1" />
              <el-option :label="t('tokenUsage.range.days7')" :value="7" />
              <el-option :label="t('tokenUsage.range.days30')" :value="30" />
            </el-select>
            <el-button @click="exportCsv">{{ t('tokenUsage.exportCsv') }}</el-button>
            <el-button :loading="loading" @click="loadData">{{ t('common.actions.refresh') }}</el-button>
          </div>
        </div>
      </template>

      <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

      <el-row :gutter="12" class="summary-row">
        <el-col :xs="24" :md="8">
          <el-statistic :title="t('tokenUsage.summary.totalTokens')" :value="summary.totalTokens" />
        </el-col>
        <el-col :xs="24" :md="8">
          <el-statistic :title="t('tokenUsage.summary.inputTokens')" :value="summary.inputTokens" />
        </el-col>
        <el-col :xs="24" :md="8">
          <el-statistic :title="t('tokenUsage.summary.estimatedCost')" :value="summary.estimatedCost" :precision="4">
            <template #prefix>$</template>
          </el-statistic>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="never">
      <template #header>{{ t('tokenUsage.botListTitle') }}</template>
      <el-table v-loading="loading" :data="bots" row-key="botId" style="width: 100%" @row-click="goDetail">
        <el-table-column prop="botId" :label="t('tokenUsage.columns.botId')" min-width="160" />
        <el-table-column prop="sessions" :label="t('tokenUsage.columns.sessions')" width="120" />
        <el-table-column prop="totalTokens" :label="t('tokenUsage.columns.totalTokens')" min-width="160" />
        <el-table-column prop="estimatedCost" :label="t('tokenUsage.columns.estimatedCost')" min-width="160">
          <template #default="{ row }">${{ Number(row.estimatedCost || 0).toFixed(4) }}</template>
        </el-table-column>
        <el-table-column :label="t('tokenUsage.columns.actions')" width="120">
          <template #default="{ row }">
            <el-button type="primary" link @click.stop="goDetail(row)">{{ t('tokenUsage.viewDetail') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getTokenUsageSummary, type BotUsageRow } from '../services/tokenUsage'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const errorMessage = ref('')
const summary = ref({ inputTokens: 0, outputTokens: 0, totalTokens: 0, estimatedCost: 0 })
const bots = ref<BotUsageRow[]>([])
const days = ref(0)

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getTokenUsageSummary(days.value)
    summary.value = data.total
    bots.value = data.bots
  } catch (err) {
    errorMessage.value = parseError(err, t('tokenUsage.messages.loadFailed'))
    ElMessage.error(errorMessage.value)
  } finally {
    loading.value = false
  }
}

function goDetail(row: BotUsageRow) {
  router.push({
    path: `/token-usage/${encodeURIComponent(row.botId)}`,
    query: days.value > 0 ? { days: String(days.value) } : undefined,
  })
}

function exportCsv() {
  const header = ['botId', 'sessions', 'inputTokens', 'outputTokens', 'totalTokens', 'estimatedCost']
  const rows = bots.value.map((row) => [
    row.botId,
    String(row.sessions),
    String(row.inputTokens),
    String(row.outputTokens),
    String(row.totalTokens),
    Number(row.estimatedCost || 0).toFixed(6),
  ])
  const csv = [header, ...rows].map((line) => line.map((x) => `"${String(x).replaceAll('"', '""')}"`).join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `token-usage-${days.value || 'all'}d.csv`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

onMounted(() => {
  const queryDays = Number(route.query.days || 0)
  if (Number.isFinite(queryDays) && [0, 1, 7, 30].includes(queryDays)) {
    days.value = queryDays
  }
  loadData()
})
</script>

<style scoped>
.token-usage-page {
  display: grid;
  gap: 12px;
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-tools {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.days-select {
  width: 140px;
}

.summary-row {
  margin-top: 8px;
}
</style>
