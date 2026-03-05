<template>
  <div class="register-page">
    <h3>注册</h3>
    <el-form label-position="top" class="register-form">
      <el-form-item label="用户名">
        <el-input v-model="username" placeholder="用户名" />
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="password" type="password" show-password placeholder="密码" />
      </el-form-item>
      <el-form-item label="确认密码">
        <el-input v-model="confirm" type="password" show-password placeholder="确认密码" />
      </el-form-item>
      <el-form-item label="密码强度">
        <el-tag :type="strength === '强' ? 'success' : strength === '中' ? 'warning' : 'danger'">{{ strength }}</el-tag>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :disabled="!canSubmit" @click="register">注册</el-button>
      </el-form-item>
      <el-alert v-if="msg" :title="msg" :type="msg.includes('成功') ? 'success' : 'error'" show-icon :closable="false" />
    </el-form>
  </div>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const password = ref('')
const confirm = ref('')
const msg = ref('')

const strength = computed(() => {
  const p = password.value
  if (p.length < 8) return '弱'
  const hasLetter = /[A-Za-z]/.test(p)
  const hasNum = /\d/.test(p)
  const hasSpecial = /[^A-Za-z\d]/.test(p)
  if (hasLetter && hasNum && hasSpecial) return '强'
  if (hasLetter && hasNum) return '中'
  return '弱'
})
const canSubmit = computed(() => password.value === confirm.value && strength.value !== '弱' && username.value.length >= 3)

async function register() {
  try {
    await axios.post('/api/v1/auth/register', { username: username.value, password: password.value })
    msg.value = '注册成功，请登录'
    setTimeout(() => router.push('/login'), 300)
  } catch {
    msg.value = '注册失败'
  }
}
</script>
<style scoped>
.register-page { max-width: 420px; }
</style>
