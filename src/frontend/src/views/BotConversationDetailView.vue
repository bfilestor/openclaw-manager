<template>
  <div class="bot-detail-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-title-row">
          <span>{{ t('tokenUsage.detailTitle', { botId }) }}</span>
          <el-button @click="goBack">{{ t('tokenUsage.backToList') }}</el-button>
        </div>
      </template>

      <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon :closable="false" />

      <el-table v-loading="loading" :data="items" row-key="sessionKey" style="width: 100%">
        <el-table-column prop="updatedAt" :label="t('tokenUsage.columns.updatedAt')" min-width="180" />
        <el-table-column prop="agentId" :label="t('tokenUsage.columns.agentId')" width="120" />
        <el-table-column prop="modelProvider" :label="t('tokenUsage.columns.provider')" min-width="140" />
        <el-table-column prop="totalTokens" :label="t('tokenUsage.columns.totalTokens')" width="130" />
        <el-table-column prop="estimatedCost" :label="t('tokenUsage.columns.estimatedCost')" width="130">
          <template #default="{ row }">${{ Number(row.estimatedCost || 0).toFixed(4) }}</template>
        </el-table-column>
        <el-table-column prop="preview" :label="t('tokenUsage.columns.preview')" min-width="320" show-overflow-tooltip />
      </el-table>

      <div class="pager">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="total"
          :current-page="page"
          :page-size="pageSize"
          @current-change="onPageChange"
        />
      </div>

      <el-text type="info">{{ t('tokenUsage.maxHint') }}</el-text>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getBotConversations, type ConversationItem } from '../services/tokenUsage'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const items = ref<ConversationItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

const botId = computed(() => String(route.params.botId || ''))

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function loadData() {
  if (!botId.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    const data = await getBotConversations(botId.value, page.value, pageSize)
    items.value = data.items
    total.value = data.total
  } catch (err) {
    errorMessage.value = parseError(err, t('tokenUsage.messages.loadDetailFailed'))
    ElMessage.error(errorMessage.value)
  } finally {
    loading.value = false
  }
}

function onPageChange(next: number) {
  page.value = next
  loadData()
}

function goBack() {
  router.push('/token-usage')
}

onMounted(loadData)
</script>

<style scoped>
.bot-detail-page {
  display: grid;
}

.card-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pager {
  margin: 12px 0;
  display: flex;
  justify-content: flex-end;
}
</style>
