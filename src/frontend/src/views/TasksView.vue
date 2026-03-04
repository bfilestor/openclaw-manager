<template>
  <div>
    <h3>Tasks</h3>
    <div style="display:flex;gap:16px">
      <div style="min-width:320px">
        <div v-for="t in tasks" :key="t.task_id" @click="select(t)" :style="{padding:'6px',cursor:'pointer',background:t.status==='FAILED'?'#ffe0e0':'transparent'}">
          {{ t.task_type }} - {{ t.status }}
        </div>
      </div>
      <div style="flex:1">
        <div>
          <label><input type="checkbox" v-model="autoScroll" />自动滚动</label>
          <input v-model="keyword" placeholder="搜索日志" />
        </div>
        <pre ref="logBox" style="height:320px;overflow:auto;background:#111;color:#ddd;padding:8px">{{ filteredLog }}</pre>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
const tasks = ref<any[]>([])
const selected = ref<any>(null)
const logText = ref('')
const autoScroll = ref(true)
const keyword = ref('')
const logBox = ref<HTMLElement | null>(null)
const filteredLog = computed(()=> keyword.value ? logText.value.split('\n').filter(l=>l.includes(keyword.value)).join('\n') : logText.value)
async function load(){ const {data}=await axios.get('/api/v1/tasks'); tasks.value=data.tasks||[] }
async function select(t:any){ selected.value=t; logText.value=''; const token='x'; const es=new EventSource(`/api/v1/tasks/${t.task_id}/events?token=${token}`); es.onmessage=(ev)=>{ logText.value += ev.data+'\n'; if(autoScroll.value && logBox.value){ logBox.value.scrollTop = logBox.value.scrollHeight } } }
onMounted(load)
</script>
