<template>
  <div>
    <h3>登录</h3>
    <input v-model="username" placeholder="用户名" />
    <input v-model="password" placeholder="密码" type="password" />
    <button @click="login">登录</button>
    <router-link to="/register">前往注册</router-link>
    <p v-if="error" style="color:red">{{ error }}</p>
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
