<template>
  <div class="workspace-migrate-page">
    <div class="topbar">
      <h3>Workspace 迁移</h3>
      <el-button @click="goBack">返回 Agent 列表</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-card shadow="never">
      <template #header>迁移信息</template>
      <el-form label-position="top">
        <el-form-item label="Agent ID">
          <el-input :model-value="agentID" disabled />
        </el-form-item>

        <el-form-item label="旧目录地址">
          <el-input :model-value="oldWorkspacePath" disabled />
        </el-form-item>

        <el-form-item label="新目录地址">
          <el-input
            v-model="newWorkspacePath"
            placeholder="例如：/home/mixi/.openclaw/workspace-xcoder-v2"
            clearable
          />
        </el-form-item>

        <el-alert
          title="保存后会移动旧目录下的全部内容，并更新 openclaw.json 对应 Agent 的 workspace，最后自动重启 openclaw gateway。"
          type="warning"
          show-icon
          :closable="false"
        />

        <el-button
          type="primary"
          class="submit-btn"
          :loading="submitting"
          :disabled="!newWorkspacePath.trim()"
          @click="submitMigrate"
        >
          保存并迁移
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'

type AgentItem = {
  agent_id: string
  workspace_path: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const oldWorkspacePath = ref('')
const newWorkspacePath = ref('')

const agentID = computed(() => String(route.params.id || '').trim())

function goBack() {
  router.push('/agents')
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function loadAgent() {
  if (!agentID.value) {
    errorMessage.value = '缺少 agent_id 参数'
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get<AgentItem>(`/api/v1/agents/${encodeURIComponent(agentID.value)}`)
    oldWorkspacePath.value = String(data?.workspace_path || '').trim()
    if (!oldWorkspacePath.value) {
      errorMessage.value = '该 Agent 未返回 workspace 路径，暂时无法迁移'
      return
    }
    newWorkspacePath.value = oldWorkspacePath.value
  } catch (err) {
    errorMessage.value = parseError(err, '加载 Agent 信息失败')
  } finally {
    loading.value = false
  }
}

async function submitMigrate() {
  if (!agentID.value) return
  if (!oldWorkspacePath.value) {
    ElMessage.error('旧目录地址为空，无法迁移')
    return
  }
  const target = newWorkspacePath.value.trim()
  if (!target) {
    ElMessage.error('请填写新目录地址')
    return
  }
  if (target === oldWorkspacePath.value) {
    ElMessage.error('新目录地址不能与旧目录相同')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确认将 ${agentID.value} 的 workspace 迁移到新目录？\n\n旧目录：${oldWorkspacePath.value}\n新目录：${target}`,
      '迁移确认',
      { type: 'warning' }
    )
  } catch {
    return
  }

  submitting.value = true
  try {
    await axios.post(`/api/v1/agents/${encodeURIComponent(agentID.value)}/workspace/migrate`, {
      new_workspace_path: target,
    })
    ElMessage.success('迁移完成，Gateway 已重启')
    await router.push('/agents')
  } catch (err) {
    ElMessage.error(parseError(err, '迁移失败'))
  } finally {
    submitting.value = false
  }
}

onMounted(loadAgent)
</script>

<style scoped>
.workspace-migrate-page {
  display: grid;
  gap: 12px;
}
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.topbar h3 {
  margin: 0;
}
.submit-btn {
  margin-top: 12px;
}
</style>
