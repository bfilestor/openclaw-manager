<template>
  <div>
    <h3>Dashboard</h3>
    <div v-if="nvmWarning" class="banner">检测到 NVM Node 风险，建议修复</div>
    <div class="cards">
      <div>Gateway: {{ status.active_state || 'unknown' }}</div>
      <div>Bind: {{ status.bind_addr || '-' }}:{{ status.port || '-' }}</div>
    </div>
    <button :disabled="!canOperate" @click="act('start')">启动</button>
    <button :disabled="!canOperate" @click="act('stop')">停止</button>
    <button :disabled="!canOperate" @click="act('restart')">重启</button>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const status = ref<any>({})
const nvmWarning = ref(false)
const canOperate = computed(() => ['Operator','Admin'].includes(auth.user?.role || 'Viewer'))
let timer: any = null

async function refresh() {
  try {
    const { data } = await axios.get('/api/v1/gateway/status')
    status.value = { active_state: data?.service?.active_state, bind_addr: data?.bind_addr, port: data?.port }
    nvmWarning.value = !!data?.nvm_warning
  } catch {}
}
async function act(op: 'start'|'stop'|'restart') {
  await axios.post(`/api/v1/gateway/${op}`)
  await refresh()
}
onMounted(() => { refresh(); timer = setInterval(refresh, 30000) })
onUnmounted(() => clearInterval(timer))
</script>
<style scoped>.banner{background:#ffe8b0;padding:8px;margin-bottom:10px}</style>
