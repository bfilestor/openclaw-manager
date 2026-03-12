<template>
  <div class="register-page">
    <h3>{{ t('register.title') }}</h3>
    <el-alert
      v-if="registrationDisabled"
      :title="t('register.messages.disabled')"
      type="warning"
      show-icon
      :closable="false"
    />
    <el-form v-else label-position="top" class="register-form">
      <el-form-item :label="t('register.username')">
        <el-input v-model="username" :placeholder="t('register.username')" />
      </el-form-item>
      <el-form-item :label="t('register.password')">
        <el-input v-model="password" type="password" show-password :placeholder="t('register.password')" />
      </el-form-item>
      <el-form-item :label="t('register.confirmPassword')">
        <el-input v-model="confirm" type="password" show-password :placeholder="t('register.confirmPassword')" />
      </el-form-item>
      <el-form-item :label="t('register.passwordStrength')">
        <el-tag :type="strength === 'strong' ? 'success' : strength === 'medium' ? 'warning' : 'danger'">
          {{ t(`register.strength.${strength}`) }}
        </el-tag>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :disabled="!canSubmit" @click="register">{{ t('register.register') }}</el-button>
      </el-form-item>
      <el-alert v-if="msg" :title="msg" :type="msgType" show-icon :closable="false" />
    </el-form>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()
const username = ref('')
const password = ref('')
const confirm = ref('')
const msg = ref('')
const msgType = ref<'success' | 'error'>('success')
const registrationDisabled = ref(false)

const strength = computed(() => {
  const p = password.value
  if (p.length < 8) return 'weak'
  const hasLetter = /[A-Za-z]/.test(p)
  const hasNum = /\d/.test(p)
  const hasSpecial = /[^A-Za-z\d]/.test(p)
  if (hasLetter && hasNum && hasSpecial) return 'strong'
  if (hasLetter && hasNum) return 'medium'
  return 'weak'
})
const canSubmit = computed(() => password.value === confirm.value && strength.value !== 'weak' && username.value.length >= 3)

async function register() {
  try {
    await axios.post('/api/v1/auth/register', { username: username.value, password: password.value })
    msg.value = t('register.messages.success')
    msgType.value = 'success'
    setTimeout(() => router.push('/login'), 300)
  } catch {
    msg.value = t('register.messages.failed')
    msgType.value = 'error'
  }
}

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/v1/auth/public-registration')
    registrationDisabled.value = !Boolean(data?.public_registration)
  } catch {
    registrationDisabled.value = true
  }
})
</script>
<style scoped>
.register-page {
  max-width: 420px;
  margin: 72px auto 0;
  background: var(--oc-surface);
  border: 1px solid var(--oc-border);
  border-radius: 12px;
  padding: 18px;
  box-shadow: var(--oc-shadow);
}
</style>
