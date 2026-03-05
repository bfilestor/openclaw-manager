<template>
  <div class="dashboard-page">
    <div class="topbar">
      <h3>Dashboard</h3>
      <el-dropdown trigger="hover" @command="handleUserMenu">
        <span class="user-trigger">
          <el-avatar size="small">{{ userInitial }}</el-avatar>
          <span class="username">{{ auth.user?.username || 'User' }}</span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="change-password">修改密码</el-dropdown-item>
            <el-dropdown-item command="logout" divided>退出系统</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
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

    <el-dialog v-model="showPwdDialog" title="修改密码" width="420px">
      <el-form label-position="top">
        <el-form-item label="旧密码">
          <el-input v-model="passwordForm.old_password" type="password" show-password placeholder="请输入旧密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="passwordForm.new_password" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input v-model="confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-space>
          <el-button @click="showPwdDialog = false">取消</el-button>
          <el-button type="primary" :loading="submittingPwd" @click="submitPasswordChange">确认修改</el-button>
        </el-space>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const status = ref<any>({})
const nvmWarning = ref(false)
const canOperate = computed(() => ['Operator','Admin'].includes(auth.user?.role || 'Viewer'))
const userInitial = computed(() => (auth.user?.username?.[0] || 'U').toUpperCase())
const showPwdDialog = ref(false)
const submittingPwd = ref(false)
const passwordForm = ref({ old_password: '', new_password: '' })
const confirmPassword = ref('')
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
function resetPasswordForm() {
  passwordForm.value.old_password = ''
  passwordForm.value.new_password = ''
  confirmPassword.value = ''
}
async function handleUserMenu(command: string) {
  if (command === 'change-password') {
    showPwdDialog.value = true
    return
  }
  if (command === 'logout') {
    try {
      await axios.post('/api/v1/auth/logout')
    } catch {
      // 无论后端是否成功，前端都应清理本地会话并回到登录页
    } finally {
      auth.clear()
      router.push('/login')
    }
  }
}
async function submitPasswordChange() {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    ElMessage.error('请完整填写密码信息')
    return
  }
  if (passwordForm.value.new_password !== confirmPassword.value) {
    ElMessage.error('两次输入的新密码不一致')
    return
  }
  submittingPwd.value = true
  try {
    await axios.put('/api/v1/users/me/password', passwordForm.value)
    ElMessage.success('密码修改成功，请重新登录')
    showPwdDialog.value = false
    resetPasswordForm()
    try {
      await axios.post('/api/v1/auth/logout')
    } catch {
      // ignore
    } finally {
      auth.clear()
      router.push('/login')
    }
  } catch {
    ElMessage.error('密码修改失败，请检查旧密码是否正确')
  } finally {
    submittingPwd.value = false
  }
}
onMounted(() => { refresh(); timer = setInterval(refresh, 30000) })
onUnmounted(() => clearInterval(timer))
</script>
<style scoped>
.dashboard-page { display: grid; gap: 12px; }
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.topbar h3 { margin: 0; }
.user-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}
.username { font-size: 14px; }
.cards { margin: 0; }
</style>
