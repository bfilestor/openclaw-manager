<template>
  <div class="tasks-page">
    <h3>Tasks</h3>
    <el-row :gutter="16">
      <el-col :xs="24" :md="10" :lg="8">
        <el-card shadow="never">
          <template #header>任务列表</template>
          <el-table :data="tasks" style="width: 100%" highlight-current-row @row-click="select">
            <el-table-column prop="task_type" label="类型" min-width="110" />
            <el-table-column label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'FAILED' ? 'danger' : row.status === 'SUCCEEDED' ? 'success' : 'info'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="14" :lg="16">
        <el-card shadow="never">
          <template #header>
            日志 {{ selected?.task_id ? `(Task: ${selected.task_id})` : '' }}
          </template>
          <el-space class="toolbar">
            <el-checkbox v-model="autoScroll">自动滚动</el-checkbox>
            <el-input v-model="keyword" placeholder="搜索日志" clearable />
          </el-space>
          <pre ref="logBox" class="log-box">{{ filteredLog }}</pre>
        </el-card>
      </el-col>
    </el-row>
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
<style scoped>
.tasks-page { display: grid; gap: 12px; }
.toolbar { margin-bottom: 8px; }
.log-box {
  height: 320px;
  overflow: auto;
  background: #111;
  color: #ddd;
  padding: 8px;
  border-radius: 6px;
}
</style>
