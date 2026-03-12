<template>
  <div class="users-page">
    <div class="topbar">
      <h3>{{ t('adminUsers.title') }}</h3>
      <el-space>
        <el-tag type="info">{{ t('adminUsers.totalUsers', { count: users.length }) }}</el-tag>
        <el-tag type="warning">{{ t('adminUsers.adminCount', { count: adminCount }) }}</el-tag>
        <el-button type="primary" @click="openCreateDialog">{{ t('adminUsers.createUser') }}</el-button>
        <el-button :loading="loading" @click="load">{{ t('common.actions.refresh') }}</el-button>
      </el-space>
    </div>

    <el-table v-loading="loading" :data="users" border row-key="user_id" style="width: 100%">
      <el-table-column prop="username" :label="t('adminUsers.columns.username')" min-width="140" />
      <el-table-column :label="t('adminUsers.columns.role')" min-width="180">
        <template #default="{ row }">
          <el-select v-model="row.role" :disabled="row.user_id === meId" @change="changeRole(row)">
            <el-option :label="t('roles.User')" value="User" />
            <el-option :label="t('roles.Viewer')" value="Viewer" />
            <el-option :label="t('roles.Operator')" value="Operator" />
            <el-option :label="t('roles.Admin')" value="Admin" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column :label="t('adminUsers.columns.accountBinding')" min-width="240">
        <template #default="{ row }">
          <el-space>
            <el-input v-model="row.account_id" :placeholder="t('adminUsers.accountIdPlaceholder')" clearable />
            <el-input-number v-model="row.token_limit" :min="0" :step="1000" :placeholder="t('adminUsers.tokenLimitPlaceholder')" />
            <el-button size="small" @click="saveAccountBinding(row)">{{ t('common.actions.confirm') }}</el-button>
          </el-space>
        </template>
      </el-table-column>
      <el-table-column :label="t('adminUsers.columns.status')" min-width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'disabled' ? 'danger' : 'success'">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('adminUsers.columns.actions')" min-width="300">
        <template #default="{ row }">
          <el-space>
            <el-button size="small" @click="openResetPasswordDialog(row)">{{ t('adminUsers.resetPassword') }}</el-button>
            <el-button size="small" :disabled="row.user_id === meId" @click="toggleDisable(row)">
              {{ row.status === 'disabled' ? t('adminUsers.enable') : t('adminUsers.disable') }}
            </el-button>
            <el-button size="small" type="danger" :disabled="deleteDisabled(row)" @click="del(row)">
              {{ t('common.actions.delete') }}
            </el-button>
          </el-space>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="createDialogVisible" :title="t('adminUsers.createUser')" width="460px">
      <el-form label-position="top">
        <el-form-item :label="t('adminUsers.columns.username')">
          <el-input v-model="createForm.username" maxlength="32" :placeholder="t('adminUsers.usernamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('adminUsers.initialPassword')">
          <el-input v-model="createForm.password" type="password" show-password :placeholder="t('adminUsers.passwordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('adminUsers.columns.role')">
          <el-select v-model="createForm.role" style="width: 100%">
            <el-option :label="t('roles.User')" value="User" />
            <el-option :label="t('roles.Viewer')" value="Viewer" />
            <el-option :label="t('roles.Operator')" value="Operator" />
            <el-option :label="t('roles.Admin')" value="Admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="createDialogVisible = false">{{ t('common.actions.cancel') }}</el-button>
          <el-button type="primary" :loading="creating" @click="createUser">{{ t('adminUsers.create') }}</el-button>
        </el-space>
      </template>
    </el-dialog>

    <el-dialog v-model="resetDialogVisible" :title="t('adminUsers.resetDialogTitle', { username: resetTarget?.username || '' })" width="460px">
      <el-form label-position="top">
        <el-form-item :label="t('adminUsers.newPassword')">
          <el-input v-model="resetForm.new_password" type="password" show-password :placeholder="t('adminUsers.passwordPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="resetDialogVisible = false">{{ t('common.actions.cancel') }}</el-button>
          <el-button type="primary" :loading="resetting" @click="resetPassword">{{ t('adminUsers.confirmReset') }}</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'

type Role = 'User' | 'Viewer' | 'Operator' | 'Admin'

type UserItem = {
  user_id: string
  username: string
  role: Role
  status: 'active' | 'disabled'
  account_id?: string
  token_limit?: number
}

const auth = useAuthStore()
const { t } = useI18n()
const meId = computed(() => auth.user?.user_id || '')

const loading = ref(false)
const creating = ref(false)
const resetting = ref(false)
const users = ref<UserItem[]>([])

const createDialogVisible = ref(false)
const createForm = ref({
  username: '',
  password: '',
  role: 'Viewer' as Role,
})

const resetDialogVisible = ref(false)
const resetTarget = ref<UserItem | null>(null)
const resetForm = ref({ new_password: '' })

const adminCount = computed(() => users.value.filter((u) => u.role === 'Admin').length)

function statusLabel(status: 'active' | 'disabled') {
  return status === 'disabled' ? t('adminUsers.statusDisabled') : t('adminUsers.statusActive')
}

function parseError(err: any, fallback: string): string {
  const msg = err?.response?.data?.message || err?.response?.data?.error || err?.message
  return typeof msg === 'string' && msg ? msg : fallback
}

async function load() {
  loading.value = true
  try {
    const { data } = await axios.get('/api/v1/users')
    const list = Array.isArray(data?.users) ? data.users : []
    users.value = list.map((u: any) => ({ ...u, account_id: '', token_limit: 0 }))
    await Promise.all(users.value.map(async (row) => {
      try {
        const { data: bind } = await axios.get(`/api/v1/users/${row.user_id}/account-binding`)
        row.account_id = String(bind?.account_id || '')
        row.token_limit = Number(bind?.token_limit || 0)
      } catch {
        row.account_id = ''
        row.token_limit = 0
      }
    }))
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.loadFailed')))
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
    ElMessage.warning(t('adminUsers.messages.fillCreateForm'))
    return
  }
  creating.value = true
  try {
    await axios.post('/api/v1/users', createForm.value)
    ElMessage.success(t('adminUsers.messages.createSuccess'))
    createDialogVisible.value = false
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.createFailed')))
  } finally {
    creating.value = false
  }
}

async function changeRole(u: UserItem) {
  try {
    await axios.put(`/api/v1/users/${u.user_id}/role`, { role: u.role })
    ElMessage.success(t('adminUsers.messages.roleUpdated'))
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.roleUpdateFailed')))
    await load()
  }
}

async function saveAccountBinding(u: UserItem) {
  try {
    await axios.put(`/api/v1/users/${u.user_id}/account-binding`, {
      account_id: String(u.account_id || ''),
      token_limit: Number(u.token_limit || 0),
    })
    ElMessage.success(t('adminUsers.messages.bindingUpdated'))
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.bindingUpdateFailed')))
  }
}

async function toggleDisable(u: UserItem) {
  try {
    await axios.post(`/api/v1/users/${u.user_id}/disable`, { disabled: u.status !== 'disabled' })
    ElMessage.success(t('adminUsers.messages.statusUpdated'))
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.statusUpdateFailed')))
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
    ElMessage.warning(u.role === 'Admin' ? t('adminUsers.messages.lastAdminProtected') : t('adminUsers.messages.cannotDeleteCurrent'))
    return
  }
  try {
    await ElMessageBox.confirm(
      t('adminUsers.messages.confirmDelete', { username: u.username }),
      t('adminUsers.messages.deleteConfirmTitle'),
      { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    await axios.delete(`/api/v1/users/${u.user_id}`)
    ElMessage.success(t('adminUsers.messages.deleteSuccess'))
    await load()
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.deleteFailed')))
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
    ElMessage.warning(t('adminUsers.messages.needNewPassword'))
    return
  }
  resetting.value = true
  try {
    await axios.put(`/api/v1/users/${target.user_id}/password`, {
      new_password: resetForm.value.new_password,
    })
    ElMessage.success(t('adminUsers.messages.passwordResetSuccess'))
    resetDialogVisible.value = false
    resetForm.value = { new_password: '' }
  } catch (err) {
    ElMessage.error(parseError(err, t('adminUsers.messages.passwordResetFailed')))
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
