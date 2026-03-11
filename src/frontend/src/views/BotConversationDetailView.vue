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
        <el-table-column prop="preview" :label="t('tokenUsage.columns.preview')" min-width="280" show-overflow-tooltip />
        <el-table-column :label="t('tokenUsage.columns.actions')" width="120">
          <template #default="{ row }">
            <el-button type="primary" link @click="openMessages(row.sessionId)">{{ t('tokenUsage.viewMessages') }}</el-button>
          </template>
        </el-table-column>
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

    <el-drawer v-model="messagesDrawer" :title="t('tokenUsage.messagesTitle', { sessionId: currentSessionId })" size="50%">
      <el-empty v-if="sessionMessages.length === 0 && !messagesLoading" :description="t('tokenUsage.emptyMessages')" />
      <el-timeline v-loading="messagesLoading">
        <el-timeline-item v-for="(msg, idx) in sessionMessages" :key="`${idx}-${msg.timestamp}`" :timestamp="msg.timestamp" placement="top">
          <el-tag size="small">{{ msg.role }}</el-tag>
          <div class="msg-text">{{ msg.text }}</div>
        </el-timeline-item>
      </el-timeline>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getBotConversations, getSessionMessages, type ConversationItem, type SessionMessage } from '../services/tokenUsage'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const items = ref<ConversationItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const messagesDrawer = ref(false)
const messagesLoading = ref(false)
const currentSessionId = ref('')
const sessionMessages = ref<SessionMessage[]>([])

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

async function openMessages(sessionId: string) {
  currentSessionId.value = sessionId
  messagesDrawer.value = true
  messagesLoading.value = true
  sessionMessages.value = []
  try {
    sessionMessages.value = await getSessionMessages(sessionId, 120)
  } catch (err) {
    ElMessage.error(parseError(err, t('tokenUsage.messages.loadMessagesFailed')))
  } finally {
    messagesLoading.value = false
  }
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

.msg-text {
  margin-top: 6px;
  white-space: pre-wrap;
  line-height: 1.5;
}
</style>
