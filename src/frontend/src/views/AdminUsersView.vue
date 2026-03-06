<template>
  <div class="users-page">
    <div class="topbar">
      <h3>用户管理</h3>
      <el-space>
        <el-tag type="info">用户总数: {{ users.length }}</el-tag>
        <el-tag type="warning">Admin: {{ adminCount }}</el-tag>
        <el-button type="primary" @click="openCreateDialog">新增用户</el-button>
        <el-button :loading="loading" @click="load">刷新</el-button>
      </el-space>
    </div>

    <el-table v-loading="loading" :data="users" border row-key="user_id" style="width: 100%">
      <el-table-column prop="username" label="用户名" min-width="140" />
      <el-table-column label="角色" min-width="180">
        <template #default="{ row }">
          <el-select v-model="row.role" :disabled="row.user_id === meId" @change="changeRole(row)">
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
      <el-table-column label="操作" min-width="300">
        <template #default="{ row }">
          <el-space>
            <el-button size="small" @click="openResetPasswordDialog(row)">修改密码</el-button>
            <el-button size="small" :disabled="row.user_id === meId" @click="toggleDisable(row)">
              {{ row.status === 'disabled' ? '启用' : '禁用' }}
            </el-button>
            <el-button
              size="small"
              type="danger"
              :disabled="deleteDisabled(row)"
              @click="del(row)"
            >
              删除
            </el-button>
          </el-space>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="createDialogVisible" title="新增用户" width="460px">
      <el-form label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="createForm.username" maxlength="32" placeholder="3-32位字母/数字/下划线" />
        </el-form-item>
        <el-form-item label="初始密码">
          <el-input v-model="createForm.password" type="password" show-password placeholder="至少8位，包含字母和数字" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createForm.role" style="width: 100%">
            <el-option label="Viewer" value="Viewer" />
            <el-option label="Operator" value="Operator" />
            <el-option label="Admin" value="Admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="creating" @click="createUser">创建</el-button>
        </el-space>
      </template>
    </el-dialog>

    <el-dialog v-model="resetDialogVisible" :title="`修改密码 - ${resetTarget?.username || ''}`" width="460px">
      <el-form label-position="top">
        <el-form-item label="新密码">
          <el-input v-model="resetForm.new_password" type="password" show-password placeholder="至少8位，包含字母和数字" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="resetDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="resetting" @click="resetPassword">确认修改</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '../stores/auth'

type UserItem = {
  user_id: string
  username: string
  role: 'Viewer' | 'Operator' | 'Admin'
  status: 'active' | 'disabled'
}

const auth = useAuthStore()
const meId = computed(() => auth.user?.user_id || '')

const loading = ref(false)
const creating = ref(false)
const resetting = ref(false)
const users = ref<UserItem[]>([])

const createDialogVisible = ref(false)
const createForm = ref({
  username: '',
  password: '',
  role: 'Viewer' as 'Viewer' | 'Operator' | 'Admin',
})

const resetDialogVisible = ref(false)
const resetTarget = ref<UserItem | null>(null)
const resetForm = ref({ new_password: '' })

const adminCount = computed(() => users.value.filter((u) => u.role === 'Admin').length)

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function load() {
  loading.value = true
  try {
    const { data } = await axios.get('/api/v1/users')
    users.value = Array.isArray(data?.users) ? data.users : []
  } catch (err) {
    ElMessage.error(parseError(err, '加载用户列表失败'))
    users.value = []
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  createForm.value = { username: '', password: '', role: 'Viewer' }
  createDialogVisible.value = true
}

async function createUser() {
  if (!createForm.value.username.trim() || !createForm.value.password) {
    ElMessage.warning('请完整填写新增用户信息')
    return
  }
  creating.value = true
  try {
    await axios.post('/api/v1/users', createForm.value)
    ElMessage.success('用户创建成功')
    createDialogVisible.value = false
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, '创建用户失败'))
  } finally {
    creating.value = false
  }
}

async function changeRole(u: UserItem) {
  try {
    await axios.put(`/api/v1/users/${u.user_id}/role`, { role: u.role })
    ElMessage.success('角色更新成功')
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, '角色更新失败'))
    await load()
  }
}

async function toggleDisable(u: UserItem) {
  try {
    await axios.post(`/api/v1/users/${u.user_id}/disable`, { disabled: u.status !== 'disabled' })
    ElMessage.success('状态更新成功')
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, '状态更新失败'))
    await load()
  }
}

function deleteDisabled(u: UserItem): boolean {
  if (u.user_id === meId.value) return true
  if (u.role === 'Admin' && adminCount.value <= 1) return true
  return false
}

async function del(u: UserItem) {
  if (deleteDisabled(u)) {
    ElMessage.warning(u.role === 'Admin' ? '系统仅剩一个 Admin，不能删除' : '当前用户不能删除')
    return
  }
  try {
    await ElMessageBox.confirm(`确认删除用户 ${u.username}？`, '删除确认', { type: 'warning' })
  } catch {
    return
  }
  try {
    await axios.delete(`/api/v1/users/${u.user_id}`)
    ElMessage.success('删除成功')
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, '删除失败'))
  }
}

function openResetPasswordDialog(u: UserItem) {
  resetTarget.value = u
  resetForm.value = { new_password: '' }
  resetDialogVisible.value = true
}

async function resetPassword() {
  const target = resetTarget.value
  if (!target) return
  if (!resetForm.value.new_password) {
    ElMessage.warning('请输入新密码')
    return
  }
  resetting.value = true
  try {
    await axios.put(`/api/v1/users/${target.user_id}/password`, {
      new_password: resetForm.value.new_password,
    })
    ElMessage.success('密码修改成功')
    resetDialogVisible.value = false
    resetForm.value = { new_password: '' }
  } catch (err) {
    ElMessage.error(parseError(err, '密码修改失败'))
  } finally {
    resetting.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.users-page {
  display: grid;
  gap: 12px;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.topbar h3 {
  margin: 0;
}
</style>
