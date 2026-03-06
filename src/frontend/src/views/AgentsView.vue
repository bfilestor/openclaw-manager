<template>
  <div class="agents-page">
    <div class="topbar">
      <h3>Agents</h3>
      <el-button :loading="loading" @click="loadAgents">刷新</el-button>
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
        <el-card shadow="never">Agent 数量: {{ agents.length }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="never" class="clickable-card" @click="goBindings">
          <div class="binding-card-content">
            <span>Bindings 总数: {{ totalBindings }}</span>
            <el-text type="primary">点击查看拓扑</el-text>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <el-table v-loading="loading" :data="agents" row-key="agent_id" style="width: 100%">
        <el-table-column prop="agent_id" label="Agent ID" min-width="180" />
        <el-table-column label="Workspace 位置" min-width="420">
          <template #default="{ row }">
            <el-text truncated>{{ row.workspace_path }}</el-text>
          </template>
        </el-table-column>
        <el-table-column label="Bindings" width="120">
          <template #default="{ row }">
            <el-tag :type="row.bindings_count > 0 ? 'success' : 'info'">
              {{ row.bindings_count }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link @click="goMigrate(row)">迁移</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty
        v-if="!loading && agents.length === 0"
        description="当前系统没有可用 Agent"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

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

function goBindings() {
  router.push('/bindings')
}

function goMigrate(row: AgentItem) {
  router.push(`/agents/${encodeURIComponent(row.agent_id)}/workspace-migrate`)
}

async function loadAgents() {
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get('/api/v1/agents')
    agents.value = Array.isArray(data?.agents) ? data.agents : []
  } catch {
    agents.value = []
    errorMessage.value = '加载 Agent 列表失败，请检查服务状态后重试'
    ElMessage.error('加载 Agent 列表失败')
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
