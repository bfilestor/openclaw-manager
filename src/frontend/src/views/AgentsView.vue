<template>
  <div class="agents-page">
    <div class="topbar">
      <h3>{{ t('agents.title') }}</h3>
      <el-button :loading="loading" @click="loadAgents">{{ t('common.actions.refresh') }}</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-row :gutter="12" class="stats-row">
      <el-col :xs="24" :sm="12">
        <el-card shadow="never">{{ t('agents.totalAgents', { count: agents.length }) }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="never" class="clickable-card" @click="goBindings">
          <div class="binding-card-content">
            <span>{{ t('agents.totalBindings', { count: totalBindings }) }}</span>
            <el-text type="primary">{{ t('agents.viewTopology') }}</el-text>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="agents" row-key="agent_id" style="width: 100%">
        <el-table-column prop="agent_id" :label="t('agents.columns.agentId')" min-width="180" />
        <el-table-column :label="t('agents.columns.workspacePath')" min-width="420">
          <template #default="{ row }">
            <el-text truncated>{{ row.workspace_path }}</el-text>
          </template>
        </el-table-column>
        <el-table-column :label="t('agents.columns.bindings')" width="120">
          <template #default="{ row }">
            <el-tag :type="row.bindings_count > 0 ? 'success' : 'info'">
              {{ row.bindings_count }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('agents.columns.actions')" width="180">
          <template #default="{ row }">
            <el-space>
              <el-button type="primary" link @click="goDetails(row)">{{ t('agents.viewDetails') }}</el-button>
              <el-button type="warning" link @click="goMigrate(row)">{{ t('agents.migrate') }}</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && agents.length === 0"
        :description="t('agents.empty')"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

type AgentItem = {
  agent_id: string
  workspace_path: string
  bindings_count: number
}

const loading = ref(false)
const errorMessage = ref('')
const agents = ref<AgentItem[]>([])
const totalBindings = computed(() => agents.value.reduce((sum, it) => sum + (it.bindings_count || 0), 0))
const router = useRouter()
const { t } = useI18n()

function goBindings() {
  router.push('/bindings')
}

function goMigrate(row: AgentItem) {
  router.push(`/agents/${encodeURIComponent(row.agent_id)}/workspace-migrate`)
}

function goDetails(row: AgentItem) {
  router.push(`/agents/${encodeURIComponent(row.agent_id)}/workspace-files`)
}

async function loadAgents() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get('/api/v1/agents')
    agents.value = Array.isArray(data?.agents) ? data.agents : []
  } catch {
    agents.value = []
    errorMessage.value = t('agents.messages.loadFailedHint')
    ElMessage.error(t('agents.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(loadAgents)
</script>

<style scoped>
.agents-page {
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
.stats-row {
  margin: 0;
}
.clickable-card {
  cursor: pointer;
}
.clickable-card:hover {
  border-color: var(--el-color-primary-light-5);
}
.binding-card-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
</style>
