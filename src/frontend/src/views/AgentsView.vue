<template>
  <div class="agents-page">
    <div class="topbar">
      <h3>{{ t('agents.title') }}</h3>
      <el-space>
        <el-button type="primary" @click="openCreateDialog">{{ t('agents.create.button') }}</el-button>
        <el-button :loading="loading" @click="loadAgents">{{ t('common.actions.refresh') }}</el-button>
      </el-space>
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

    <el-dialog v-model="createVisible" :title="t('agents.create.title')" width="520px">
      <el-form :model="createForm" label-position="top">
        <el-form-item :label="t('agents.create.agentIdLabel')" :error="agentIdError" required>
          <el-input v-model="createForm.agent_id" :placeholder="t('agents.create.agentIdPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('agents.create.templateLabel')">
          <el-select v-model="createForm.template_agent_id" style="width: 100%" :placeholder="t('agents.create.templatePlaceholder')">
            <el-option
              v-for="item in agents"
              :key="item.agent_id"
              :label="item.agent_id"
              :value="item.agent_id"
            />
          </el-select>
        </el-form-item>
        <el-alert :title="workspaceHintText" type="info" :closable="false" />
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="createVisible = false">{{ t('common.actions.cancel') }}</el-button>
          <el-button type="primary" :loading="createLoading" @click="submitCreateAgent">{{ t('agents.create.submit') }}</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios, { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

type AgentItem = {
  agent_id: string
  workspace_path: string
  bindings_count: number
}

const loading = ref(false)
const createLoading = ref(false)
const createVisible = ref(false)
const errorMessage = ref('')
const agentIdError = ref('')
const agents = ref<AgentItem[]>([])
const createForm = ref({
  agent_id: '',
  template_agent_id: ''
})
const totalBindings = computed(() => agents.value.reduce((sum, it) => sum + (it.bindings_count || 0), 0))
const router = useRouter()
const { t } = useI18n()
const workspaceHintText = computed(() => {
  const agent = createForm.value.agent_id.trim() || '{agent_name}'
  return t('agents.create.workspaceHint', { agent })
})

function goBindings() {
  router.push('/bindings')
}

function goMigrate(row: AgentItem) {
  router.push(`/agents/${encodeURIComponent(row.agent_id)}/workspace-migrate`)
}

function goDetails(row: AgentItem) {
  router.push(`/agents/${encodeURIComponent(row.agent_id)}/workspace-files`)
}

function openCreateDialog() {
  createForm.value = {
    agent_id: '',
    template_agent_id: agents.value[0]?.agent_id || ''
  }
  agentIdError.value = ''
  createVisible.value = true
}

function isValidCreateAgentID(agentID: string) {
  return /^[A-Za-z0-9_]{1,64}$/.test(agentID)
}

function extractCreateError(err: unknown) {
  const fallback = t('agents.create.createFailed')
  const axiosErr = err as AxiosError<any>
  const data = axiosErr?.response?.data
  if (data?.fields && typeof data.fields === 'object') {
    const fields = Object.values(data.fields).filter((it) => typeof it === 'string') as string[]
    if (fields.length > 0) {
      return `${fallback}: ${fields.join(', ')}`
    }
  }
  if (typeof data?.error === 'string' && data.error.trim()) {
    return `${fallback}: ${data.error}`
  }
  if (typeof data?.detail === 'string' && data.detail.trim()) {
    return `${fallback}: ${data.detail}`
  }
  return fallback
}

async function submitCreateAgent() {
  const agentID = createForm.value.agent_id.trim()
  const templateID = createForm.value.template_agent_id.trim()
  if (!agentID) {
    agentIdError.value = t('agents.create.agentIdRequired')
    return
  }
  if (!isValidCreateAgentID(agentID)) {
    agentIdError.value = t('agents.create.agentIdError')
    return
  }
  if (!templateID) {
    ElMessage.error(t('agents.create.templateRequired'))
    return
  }
  agentIdError.value = ''
  createLoading.value = true
  try {
    await axios.post('/api/v1/agents', {
      agent_id: agentID,
      template_agent_id: templateID
    })
    createVisible.value = false
    await loadAgents()
    ElMessage.success(t('agents.create.createdAndBindingHint'))
  } catch (err) {
    ElMessage.error(extractCreateError(err))
  } finally {
    createLoading.value = false
  }
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
