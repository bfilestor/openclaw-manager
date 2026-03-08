<template>
  <div class="login-page">
    <div class="login-panel">
      <h3>登录</h3>
      <el-form label-position="top" class="login-form">
        <el-form-item label="用户名">
          <el-input v-model="username" placeholder="用户名" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="password" type="password" show-password placeholder="密码" />
        </el-form-item>
        <el-form-item>
          <el-space>
            <el-button type="primary" @click="login">登录</el-button>
            <router-link to="/register">前往注册</router-link>
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
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
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
    error.value = '登录失败'
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
}
</style>
