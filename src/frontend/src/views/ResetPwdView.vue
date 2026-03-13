<template>
  <div class="login-page">
    <div class="login-panel">
      <h3>{{ t('resetpwd.title') }}</h3>
      <el-form label-position="top" class="login-form">
        <el-form-item :label="t('resetpwd.superToken')">
          <el-input v-model="superToken" type="password" show-password :placeholder="t('resetpwd.superTokenPlaceholder')" />
        </el-form-item>

        <el-form-item>
          <el-button @click="fetchAdmin">{{ t('resetpwd.fetchAdmin') }}</el-button>
        </el-form-item>

        <el-alert v-if="adminUsername" :title="t('resetpwd.adminUser', { username: adminUsername })" type="info" show-icon :closable="false" />

        <el-form-item :label="t('resetpwd.newPassword')" style="margin-top: 12px">
          <el-input v-model="newPassword" type="password" show-password :placeholder="t('resetpwd.newPasswordPlaceholder')" />
        </el-form-item>

        <el-form-item>
          <el-space>
            <el-button type="primary" @click="resetPwd">{{ t('resetpwd.submit') }}</el-button>
            <router-link to="/login">{{ t('resetpwd.backLogin') }}</router-link>
          </el-space>
        </el-form-item>

        <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
        <el-alert v-if="success" :title="success" type="success" show-icon :closable="false" />
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import axios from 'axios'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const superToken = ref('')
const adminUsername = ref('')
const newPassword = ref('')
const error = ref('')
const success = ref('')

async function fetchAdmin() {
  error.value = ''
  success.value = ''
  adminUsername.value = ''
  if (!superToken.value) {
    error.value = t('resetpwd.needSuperToken')
    return
  }
  try {
    const res = await axios.get('/api/v1/auth/resetpwd/admin', {
      params: { super_token: superToken.value },
    })
    adminUsername.value = String(res?.data?.username || '')
    if (!adminUsername.value) {
      error.value = t('resetpwd.fetchFailed')
    }
  } catch {
    error.value = t('resetpwd.fetchFailed')
  }
}

async function resetPwd() {
  error.value = ''
  success.value = ''
  if (!superToken.value) {
    error.value = t('resetpwd.needSuperToken')
    return
  }
  if (!newPassword.value) {
    error.value = t('resetpwd.needNewPassword')
    return
  }
  try {
    await axios.post('/api/v1/auth/resetpwd', {
      super_token: superToken.value,
      new_password: newPassword.value,
    })
    success.value = t('resetpwd.resetSuccess')
    newPassword.value = ''
  } catch {
    error.value = t('resetpwd.resetFailed')
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  box-sizing: border-box;
}
.login-panel {
  width: 100%;
  max-width: 460px;
  background: var(--oc-surface);
  border: 1px solid var(--oc-border);
  border-radius: 12px;
  padding: 18px;
  box-shadow: var(--oc-shadow);
}
</style>
