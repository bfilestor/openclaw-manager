<template>
  <div class="system-settings-page">
    <div class="topbar">
      <h3>{{ t('systemSettings.title') }}</h3>
      <el-space>
        <el-button :loading="loading" @click="load">{{ t('common.actions.refresh') }}</el-button>
      </el-space>
    </div>

    <el-card shadow="never">
      <el-form label-position="top">
        <el-form-item :label="t('systemSettings.publicRegistration')">
          <el-switch v-model="publicRegistration" :active-text="t('systemSettings.enabled')" :inactive-text="t('systemSettings.disabled')" />
        </el-form-item>
        <el-alert type="info" :closable="false" show-icon :title="t('systemSettings.tip')" />
        <div class="actions">
          <el-button type="primary" :loading="saving" @click="save">{{ t('common.actions.saveConfig') }}</el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const publicRegistration = ref(true)

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function load() {
  loading.value = true
  try {
    const { data } = await axios.get('/api/v1/system/settings')
    publicRegistration.value = Boolean(data?.public_registration)
  } catch (err) {
    ElMessage.error(parseError(err, t('systemSettings.messages.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  try {
    await axios.put('/api/v1/system/settings', { public_registration: publicRegistration.value })
    ElMessage.success(t('systemSettings.messages.saveSuccess'))
  } catch (err) {
    ElMessage.error(parseError(err, t('systemSettings.messages.saveFailed')))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.system-settings-page {
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

.actions {
  margin-top: 12px;
}
</style>
