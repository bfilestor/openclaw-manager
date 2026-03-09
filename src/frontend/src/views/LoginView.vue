<template>
  <div class="login-page">
    <div class="login-panel">
      <h3>{{ t('login.title') }}</h3>
      <el-form label-position="top" class="login-form">
        <el-form-item :label="t('login.username')">
          <el-input v-model="username" :placeholder="t('login.username')" />
        </el-form-item>
        <el-form-item :label="t('login.password')">
          <el-input v-model="password" type="password" show-password :placeholder="t('login.password')" />
        </el-form-item>
        <el-form-item>
          <el-space>
            <el-button type="primary" @click="login">{{ t('login.login') }}</el-button>
            <router-link to="/register">{{ t('login.gotoRegister') }}</router-link>
          </el-space>
        </el-form-item>
        <el-alert v-if="error" :title="error" type="error" show-icon :closable="false" />
      </el-form>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const { t } = useI18n()
const username = ref('')
const password = ref('')
const error = ref('')

async function login() {
  error.value = ''
  try {
    const res = await axios.post('/api/v1/auth/login', { username: username.value, password: password.value })
    auth.setSession(res.data.access_token, res.data.user)
    router.push('/dashboard')
  } catch {
    error.value = t('login.loginFailed')
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
  max-width: 420px;
  background: var(--oc-surface);
  border: 1px solid var(--oc-border);
  border-radius: 12px;
  padding: 18px;
  box-shadow: var(--oc-shadow);
}
</style>
