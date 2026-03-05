<template>
  <div class="users-page">
    <h3>用户管理</h3>
    <el-table :data="users" border style="width: 100%">
      <el-table-column prop="username" label="用户名" min-width="140" />
      <el-table-column label="角色" min-width="180">
        <template #default="{ row }">
          <el-select v-model="row.role" :disabled="row.user_id===meId" @change="changeRole(row)">
            <el-option label="Viewer" value="Viewer" />
            <el-option label="Operator" value="Operator" />
            <el-option label="Admin" value="Admin" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="状态" min-width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'disabled' ? 'danger' : 'success'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="180">
        <template #default="{ row }">
          <el-space>
            <el-button size="small" :disabled="row.user_id===meId" @click="toggleDisable(row)">
              {{ row.status==='disabled'?'启用':'禁用' }}
            </el-button>
            <el-button size="small" type="danger" :disabled="row.user_id===meId" @click="del(row)">删除</el-button>
          </el-space>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import axios from 'axios'
import { useAuthStore } from '../stores/auth'
const auth = useAuthStore()
const meId = auth.user?.user_id || ''
const users = ref<any[]>([])
async function load(){ const {data}=await axios.get('/api/v1/users'); users.value=data.users||[] }
async function changeRole(u:any){ await axios.put(`/api/v1/users/${u.user_id}/role`,{role:u.role}); await load() }
async function toggleDisable(u:any){ await axios.post(`/api/v1/users/${u.user_id}/disable`,{disabled:u.status!=='disabled'}); await load() }
async function del(u:any){ if(confirm('确认删除?')){ await axios.delete(`/api/v1/users/${u.user_id}`); await load() } }
onMounted(load)
</script>
<style scoped>
.users-page { display: grid; gap: 12px; }
</style>
