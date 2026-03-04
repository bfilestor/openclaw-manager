<template>
  <div>
    <h3>用户管理</h3>
    <table>
      <thead><tr><th>用户名</th><th>角色</th><th>状态</th><th>操作</th></tr></thead>
      <tbody>
        <tr v-for="u in users" :key="u.user_id">
          <td>{{ u.username }}</td>
          <td>
            <select v-model="u.role" :disabled="u.user_id===meId" @change="changeRole(u)">
              <option>Viewer</option><option>Operator</option><option>Admin</option>
            </select>
          </td>
          <td>{{ u.status }}</td>
          <td>
            <button :disabled="u.user_id===meId" @click="toggleDisable(u)">{{ u.status==='disabled'?'启用':'禁用' }}</button>
            <button :disabled="u.user_id===meId" @click="del(u)">删除</button>
          </td>
        </tr>
      </tbody>
    </table>
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
