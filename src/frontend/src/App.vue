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
            <span>{{ item.label }}</span>
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
          <div class="title-main">{{ currentNavLabel }}</div>
          <el-text type="info">{{ activePath }}</el-text>
        </div>

        <el-space>
          <el-avatar size="small" class="user-avatar">{{ usernameInitial }}</el-avatar>
          <el-text>{{ auth.user?.username || '-' }}</el-text>
          <el-button text type="primary" @click="logout">退出登录</el-button>
        </el-space>
      </el-header>

      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>

  <router-view v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

type NavItem = {
  path: string
  label: string
  adminOnly?: boolean
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const navItems: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard' },
  { path: '/gateway', label: 'Gateway' },
  { path: '/agents', label: 'Agents' },
  { path: '/bindings', label: 'Bindings' },
  { path: '/skills', label: 'Skills' },
  { path: '/config', label: 'Config' },
  { path: '/backups', label: 'Backups' },
  { path: '/tasks', label: 'Tasks' },
  { path: '/admin/users', label: 'Users', adminOnly: true },
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

const currentNavLabel = computed(() => {
  return visibleNavItems.value.find((item) => item.path === activePath.value)?.label || 'Dashboard'
})

const usernameInitial = computed(() => {
  const name = String(auth.user?.username || '').trim()
  return (name[0] || 'U').toUpperCase()
})

const currentRole = computed(() => auth.user?.role || 'Viewer')

const roleTagType = computed<'info' | 'success' | 'warning'>(() => {
  if (currentRole.value === 'Admin') return 'warning'
  if (currentRole.value === 'Operator') return 'success'
  return 'info'
})

function logout() {
  auth.clear()
  router.push('/login')
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
  border-radius: 10px;
  margin-bottom: 4px;
  height: 42px;
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

  .nav-menu :deep(.el-menu-item span) {
    font-size: 0;
  }

  .app-header {
    padding: 0 10px;
  }
}
</style>
