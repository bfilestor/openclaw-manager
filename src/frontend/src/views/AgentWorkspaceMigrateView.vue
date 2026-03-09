<template>
  <div class="workspace-migrate-page">
    <div class="topbar">
      <h3>{{ t('workspaceMigrate.title') }}</h3>
      <el-button @click="goBack">{{ t('workspaceMigrate.backToAgents') }}</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
    />

    <el-card shadow="never">
      <template #header>{{ t('workspaceMigrate.infoTitle') }}</template>
      <el-form label-position="top">
        <el-form-item :label="t('workspaceMigrate.agentId')">
          <el-input :model-value="agentID" disabled />
        </el-form-item>

        <el-form-item :label="t('workspaceMigrate.oldPath')">
          <el-input :model-value="oldWorkspacePath" disabled />
        </el-form-item>

        <el-form-item :label="t('workspaceMigrate.newPath')">
          <el-input
            v-model="newWorkspacePath"
            :placeholder="t('workspaceMigrate.newPathPlaceholder')"
            clearable
          />
        </el-form-item>

        <el-alert
          :title="t('workspaceMigrate.warning')"
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
          {{ t('workspaceMigrate.submit') }}
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
import { useI18n } from 'vue-i18n'

type AgentItem = {
  agent_id: string
  workspace_path: string
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
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
    errorMessage.value = t('workspaceMigrate.messages.missingAgentId')
    return
  }
  loading.value = true
  errorMessage.value = ''
  try {
    const { data } = await axios.get<AgentItem>(`/api/v1/agents/${encodeURIComponent(agentID.value)}`)
    oldWorkspacePath.value = String(data?.workspace_path || '').trim()
    if (!oldWorkspacePath.value) {
      errorMessage.value = t('workspaceMigrate.messages.emptyWorkspace')
      return
    }
    newWorkspacePath.value = oldWorkspacePath.value
  } catch (err) {
    errorMessage.value = parseError(err, t('workspaceMigrate.messages.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function submitMigrate() {
  if (!agentID.value) return
  if (!oldWorkspacePath.value) {
    ElMessage.error(t('workspaceMigrate.messages.oldPathEmpty'))
    return
  }
  const target = newWorkspacePath.value.trim()
  if (!target) {
    ElMessage.error(t('workspaceMigrate.messages.needNewPath'))
    return
  }
  if (target === oldWorkspacePath.value) {
    ElMessage.error(t('workspaceMigrate.messages.samePath'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('workspaceMigrate.messages.confirmContent', {
        agentId: agentID.value,
        oldPath: oldWorkspacePath.value,
        newPath: target,
      }),
      t('workspaceMigrate.messages.confirmTitle'),
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
    ElMessage.success(t('workspaceMigrate.messages.success'))
    await router.push('/agents')
  } catch (err) {
    ElMessage.error(parseError(err, t('workspaceMigrate.messages.failed')))
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
