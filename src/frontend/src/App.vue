<template>
  <el-container v-if="auth.isAuthenticated" class="app-shell">
    <el-aside class="app-aside" width="236px">
      <div class="brand">
        <div class="brand-title">OpenClaw Manager</div>
        <el-text class="brand-sub" type="info">Operations Console</el-text>
      </div>

      <el-scrollbar class="nav-scroll">
        <el-menu :default-active="activePath" router class="nav-menu">
          <el-menu-item
            v-for="item in visibleNavItems"
            :key="item.path"
            :index="item.path"
          >
            <span class="nav-item-content">
              <span class="nav-icon" aria-hidden="true">{{ item.icon }}</span>
              <span class="nav-label">{{ item.label }}</span>
            </span>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>

      <div class="aside-footer">
        <el-tag size="small" :type="roleTagType">{{ currentRole }}</el-tag>
      </div>
    </el-aside>

    <el-container class="main-shell">
      <el-header class="app-header">
        <div class="header-title">
          <div class="title-main"></div>
        </div>

        <el-dropdown trigger="hover" @command="handleUserMenu">
          <span class="user-trigger">
            <el-avatar size="small" class="user-avatar">{{ usernameInitial }}</el-avatar>
            <span class="user-name">{{ auth.user?.username || 'User' }}</span>
            <el-tag size="small" :type="roleTagType">{{ currentRole }}</el-tag>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="change-password">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出系统</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>

      <el-main class="app-main">
        <router-view />
      </el-main>

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
    </el-container>
  </el-container>

  <router-view v-else />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

type NavItem = {
  path: string
  label: string
  icon: string
  adminOnly?: boolean
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const navItems: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard', icon: '🏠' },
  { path: '/gateway', label: 'Gateway', icon: '🌐' },
  { path: '/agents', label: 'Agents', icon: '🤖' },
  { path: '/agent-sessions', label: 'Sessions', icon: '🎬' },
  { path: '/bindings', label: 'Bindings', icon: '🔗' },
  { path: '/skills', label: 'Skills', icon: '🧩' },
  { path: '/config', label: 'Config', icon: '⚙️' },
  { path: '/qqbot', label: 'QQBot', icon: '🐧' },
  { path: '/backups', label: 'Backups', icon: '💾' },
  { path: '/tasks', label: 'Tasks', icon: '✅' },
  { path: '/shell', label: 'Shell', icon: '🖥️' },
  { path: '/admin/users', label: 'Users', icon: '👥', adminOnly: true },
]

const visibleNavItems = computed(() => {
  const isAdmin = auth.user?.role === 'Admin'
  return navItems.filter((item) => !item.adminOnly || isAdmin)
})

const activePath = computed(() => {
  const current = route.path || '/dashboard'
  const matched = [...visibleNavItems.value]
    .sort((a, b) => b.path.length - a.path.length)
    .find((item) => current === item.path || current.startsWith(`${item.path}/`))
  return matched?.path || '/dashboard'
})

const usernameInitial = computed(() => {
  const name = String(auth.user?.username || '').trim()
  return (name[0] || 'U').toUpperCase()
})

const currentRole = computed(() => auth.user?.role || 'Viewer')
const showPwdDialog = ref(false)
const submittingPwd = ref(false)
const passwordForm = ref({ old_password: '', new_password: '' })
const confirmPassword = ref('')

const roleTagType = computed<'info' | 'success' | 'warning'>(() => {
  if (currentRole.value === 'Admin') return 'warning'
  if (currentRole.value === 'Operator') return 'success'
  return 'info'
})

function resetPasswordForm() {
  passwordForm.value.old_password = ''
  passwordForm.value.new_password = ''
  confirmPassword.value = ''
}

async function logout() {
  try {
    await axios.post('/api/v1/auth/logout')
  } catch {
    // 无论后端是否成功，前端都应清理本地会话并回到登录页
  } finally {
    auth.clear()
    router.push('/login')
  }
}

async function handleUserMenu(command: string) {
  if (command === 'change-password') {
    showPwdDialog.value = true
    return
  }
  if (command === 'logout') {
    await logout()
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
    await logout()
  } catch {
    ElMessage.error('密码修改失败，请检查旧密码是否正确')
  } finally {
    submittingPwd.value = false
  }
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  background:
    radial-gradient(1000px 420px at -10% -5%, rgba(22, 119, 255, 0.12), transparent 55%),
    radial-gradient(900px 360px at 110% 102%, rgba(255, 120, 50, 0.12), transparent 58%),
    #f4f6fb;
}

.app-aside {
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e8f0;
  background: linear-gradient(180deg, #fff 0%, #fbfcff 100%);
}

.brand {
  padding: 18px 16px 12px;
}

.brand-title {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: 0.2px;
}

.brand-sub {
  font-size: 12px;
}

.nav-scroll {
  flex: 1;
  padding: 8px 10px;
}

.nav-menu {
  border-right: none;
  background: transparent;
}

.nav-menu :deep(.el-menu-item) {
  position: relative;
  overflow: hidden;
  border-radius: 10px;
  margin-bottom: 4px;
  height: 42px;
  transition: transform 0.18s ease, background-color 0.2s ease;
}

.nav-menu :deep(.el-menu-item:hover) {
  transform: translateX(2px);
}

.nav-menu :deep(.el-menu-item)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 999px;
  background: linear-gradient(180deg, #2f80ff 0%, #52c1ff 100%);
  transform: scaleY(0.25);
  opacity: 0;
  transform-origin: center;
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.nav-item-content {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  transition: transform 0.2s ease;
}

.nav-icon {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(145deg, #eef4ff, #e9fbff);
  font-size: 14px;
  box-shadow: inset 0 0 0 1px rgba(73, 110, 255, 0.12);
  transition: transform 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.nav-label {
  color: #2f3442;
  transition: color 0.2s ease;
}

.nav-menu :deep(.el-menu-item:hover .nav-icon) {
  transform: scale(1.06);
}

.nav-menu :deep(.is-active)::before {
  transform: scaleY(1);
  opacity: 1;
}

.nav-menu :deep(.is-active .nav-item-content) {
  transform: translateX(2px);
}

.nav-menu :deep(.is-active .nav-icon) {
  background: linear-gradient(145deg, #2f80ff, #4aa8ff);
  box-shadow: none;
  transform: scale(1.04);
}

.nav-menu :deep(.is-active .nav-label) {
  color: #1e4f9f;
  font-weight: 600;
}

.aside-footer {
  border-top: 1px solid #edf0f6;
  padding: 12px 16px 16px;
}

.main-shell {
  min-width: 0;
}

.app-header {
  height: 64px;
  border-bottom: 1px solid #e7ebf2;
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
}

.header-title {
  min-width: 0;
}

.title-main {
  font-size: 16px;
  font-weight: 700;
  line-height: 1.25;
}

.user-avatar {
  background: linear-gradient(135deg, #2a7fff, #44b0ff);
  color: #fff;
}

.user-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.user-name {
  color: #2e3440;
  font-size: 14px;
}

.app-main {
  padding: 14px;
}

@media (max-width: 900px) {
  .app-aside {
    width: 76px !important;
  }

  .brand-sub,
  .brand-title,
  .aside-footer {
    display: none;
  }

  .nav-scroll {
    padding: 10px 8px;
  }

  .nav-menu :deep(.el-menu-item) {
    justify-content: center;
    padding: 0 !important;
  }

  .nav-item-content {
    justify-content: center;
    width: 100%;
  }

  .nav-label {
    display: none;
  }

  .app-header {
    padding: 0 10px;
  }
}
</style>
