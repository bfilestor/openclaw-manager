<template>
  <div class="token-usage-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>{{ t('tokenUsage.pageTitle') }}</span>
          <el-button :loading="loading" @click="loadData">{{ t('common.actions.refresh') }}</el-button>
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
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getTokenUsageSummary, type BotUsageRow } from '../services/tokenUsage'

const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const errorMessage = ref('')
const summary = ref({ inputTokens: 0, outputTokens: 0, totalTokens: 0, estimatedCost: 0 })
const bots = ref<BotUsageRow[]>([])

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getTokenUsageSummary()
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
  router.push({ path: `/token-usage/${encodeURIComponent(row.botId)}` })
}

onMounted(loadData)
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

.summary-row {
  margin-top: 8px;
}
</style>
