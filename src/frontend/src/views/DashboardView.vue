<template>
  <div class="dashboard-page">
    <h3>Dashboard</h3>
    <el-alert v-if="nvmWarning" title="检测到 NVM Node 风险，建议修复" type="warning" show-icon :closable="false" />
    <el-row :gutter="12" class="cards">
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">Gateway: {{ status.active_state || 'unknown' }}</el-card>
      </el-col>
      <el-col :xs="24" :sm="12">
        <el-card shadow="hover">Bind: {{ status.bind_addr || '-' }}:{{ status.port || '-' }}</el-card>
      </el-col>
    </el-row>
    <el-space>
      <el-button type="success" :disabled="!canOperate" @click="act('start')">启动</el-button>
      <el-button type="warning" :disabled="!canOperate" @click="act('stop')">停止</el-button>
      <el-button type="primary" :disabled="!canOperate" @click="act('restart')">重启</el-button>
    </el-space>
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
<style scoped>
.dashboard-page { display: grid; gap: 12px; }
.cards { margin: 0; }
</style>
