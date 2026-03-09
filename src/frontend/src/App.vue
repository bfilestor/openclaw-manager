<template>
  <el-config-provider :locale="elementLocale">
    <el-container v-if="auth.isAuthenticated" class="app-shell">
      <el-aside class="app-aside" width="236px">
        <div class="brand">
          <div class="brand-title">{{ t('common.appName') }}</div>
          <el-text class="brand-sub" type="info">{{ t('common.operationsConsole') }}</el-text>
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
                <span class="nav-label">{{ t(item.labelKey) }}</span>
              </span>
            </el-menu-item>
          </el-menu>
        </el-scrollbar>

        <div class="aside-footer">
          <el-tag size="small" :type="roleTagType">{{ currentRoleLabel }}</el-tag>
        </div>
      </el-aside>

      <el-container class="main-shell">
        <el-header class="app-header">
          <div class="header-tools">
            <div class="toolbar-group">
              <el-text type="info">{{ t('common.localeLabel') }}</el-text>
              <el-select v-model="localeModel" size="small" class="locale-select">
                <el-option value="zh-CN" :label="t('common.locales.zhCN')" />
                <el-option value="en-US" :label="t('common.locales.enUS')" />
              </el-select>
            </div>

            <div class="toolbar-group">
              <el-text type="info">{{ t('common.themeLabel') }}</el-text>
              <el-radio-group v-model="themeModel" size="small">
                <el-radio-button label="light">{{ t('common.themes.light') }}</el-radio-button>
                <el-radio-button label="dark">{{ t('common.themes.dark') }}</el-radio-button>
              </el-radio-group>
            </div>
          </div>

          <el-dropdown trigger="hover" @command="handleUserMenu">
            <span class="user-trigger">
              <el-avatar size="small" class="user-avatar">{{ usernameInitial }}</el-avatar>
              <span class="user-name">{{ auth.user?.username || t('common.user') }}</span>
              <el-tag size="small" :type="roleTagType">{{ currentRoleLabel }}</el-tag>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="change-password">{{ t('app.userMenu.changePassword') }}</el-dropdown-item>
                <el-dropdown-item command="logout" divided>{{ t('app.userMenu.logout') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-header>

        <el-main class="app-main">
          <router-view />
        </el-main>

        <el-dialog v-model="showPwdDialog" :title="t('app.passwordDialog.title')" width="420px">
          <el-form label-position="top">
            <el-form-item :label="t('app.passwordDialog.oldPassword')">
              <el-input
                v-model="passwordForm.old_password"
                type="password"
                show-password
                :placeholder="t('app.passwordDialog.oldPasswordPlaceholder')"
              />
            </el-form-item>
            <el-form-item :label="t('app.passwordDialog.newPassword')">
              <el-input
                v-model="passwordForm.new_password"
                type="password"
                show-password
                :placeholder="t('app.passwordDialog.newPasswordPlaceholder')"
              />
            </el-form-item>
            <el-form-item :label="t('app.passwordDialog.confirmPassword')">
              <el-input
                v-model="confirmPassword"
                type="password"
                show-password
                :placeholder="t('app.passwordDialog.confirmPasswordPlaceholder')"
              />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-space>
              <el-button @click="showPwdDialog = false">{{ t('common.actions.cancel') }}</el-button>
              <el-button type="primary" :loading="submittingPwd" @click="submitPasswordChange">
                {{ t('app.passwordDialog.submit') }}
              </el-button>
            </el-space>
          </template>
        </el-dialog>
      </el-container>
    </el-container>

    <div v-else class="guest-shell">
      <div class="guest-toolbar">
        <div class="toolbar-group">
          <el-text type="info">{{ t('common.localeLabel') }}</el-text>
          <el-select v-model="localeModel" size="small" class="locale-select">
            <el-option value="zh-CN" :label="t('common.locales.zhCN')" />
            <el-option value="en-US" :label="t('common.locales.enUS')" />
          </el-select>
        </div>

        <div class="toolbar-group">
          <el-text type="info">{{ t('common.themeLabel') }}</el-text>
          <el-radio-group v-model="themeModel" size="small">
            <el-radio-button label="light">{{ t('common.themes.light') }}</el-radio-button>
            <el-radio-button label="dark">{{ t('common.themes.dark') }}</el-radio-button>
          </el-radio-group>
        </div>
      </div>
      <router-view />
    </div>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import enUS from 'element-plus/es/locale/lang/en'
import zhCN from 'element-plus/es/locale/lang/zh-cn'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from './stores/auth'
import { type AppLocale, type AppTheme, usePreferencesStore } from './stores/preferences'

type NavItem = {
  path: string
  labelKey: string
  icon: string
  adminOnly?: boolean
}

const auth = useAuthStore()
const preferences = usePreferencesStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const navItems: NavItem[] = [
  { path: '/dashboard', labelKey: 'app.nav.dashboard', icon: '🏠' },
  { path: '/gateway', labelKey: 'app.nav.gateway', icon: '🌐' },
  { path: '/agents', labelKey: 'app.nav.agents', icon: '🤖' },
  { path: '/agent-sessions', labelKey: 'app.nav.sessions', icon: '🎬' },
  { path: '/bindings', labelKey: 'app.nav.bindings', icon: '🔗' },
  { path: '/skills', labelKey: 'app.nav.skills', icon: '🧩' },
  { path: '/config', labelKey: 'app.nav.config', icon: '⚙️' },
  { path: '/qqbot', labelKey: 'app.nav.qqbot', icon: '🐧' },
  { path: '/backups', labelKey: 'app.nav.backups', icon: '💾' },
  { path: '/tasks', labelKey: 'app.nav.tasks', icon: '✅' },
  { path: '/shell', labelKey: 'app.nav.shell', icon: '🖥️' },
  { path: '/admin/users', labelKey: 'app.nav.users', icon: '👥', adminOnly: true },
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
const currentRoleLabel = computed(() => t(`roles.${currentRole.value}`))
const showPwdDialog = ref(false)
const submittingPwd = ref(false)
const passwordForm = ref({ old_password: '', new_password: '' })
const confirmPassword = ref('')
const localeModel = computed<AppLocale>({
  get: () => preferences.locale,
  set: (value) => preferences.setLocale(value),
})
const themeModel = computed<AppTheme>({
  get: () => preferences.theme,
  set: (value) => preferences.setTheme(value),
})
const elementLocale = computed(() => (preferences.locale === 'en-US' ? enUS : zhCN))

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
    ElMessage.error(t('app.messages.fillPasswordInfo'))
    return
  }
  if (passwordForm.value.new_password !== confirmPassword.value) {
    ElMessage.error(t('app.messages.passwordMismatch'))
    return
  }
  submittingPwd.value = true
  try {
    await axios.put('/api/v1/users/me/password', passwordForm.value)
    ElMessage.success(t('app.messages.passwordChanged'))
    showPwdDialog.value = false
    resetPasswordForm()
    await logout()
  } catch {
    ElMessage.error(t('app.messages.passwordChangeFailed'))
  } finally {
    submittingPwd.value = false
  }
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  background: var(--oc-bg);
}

.app-aside {
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--oc-border);
  background: var(--oc-surface);
}

.brand {
  padding: 18px 16px 12px;
}

.brand-title {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: 0.2px;
  color: var(--oc-text);
}

.brand-sub {
  font-size: 12px;
  color: var(--oc-text-muted);
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
  border-radius: 8px;
  margin-bottom: 4px;
  height: 42px;
  color: var(--oc-text);
  transition: transform 0.18s ease, background-color 0.2s ease;
}

.nav-menu :deep(.el-menu-item:hover) {
  background: var(--oc-surface-muted);
  transform: translateX(2px);
}

.nav-menu :deep(.el-menu-item)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 7px;
  bottom: 7px;
  width: 3px;
  border-radius: 4px;
  background: var(--oc-accent);
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
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--oc-surface-muted);
  font-size: 14px;
  box-shadow: inset 0 0 0 1px var(--oc-border);
  transition: transform 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
}

.nav-label {
  color: var(--oc-text);
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
  background: var(--oc-accent);
  box-shadow: inset 0 0 0 1px var(--oc-accent);
  transform: scale(1.04);
}

.nav-menu :deep(.is-active .nav-label) {
  color: var(--oc-accent);
  font-weight: 600;
}

.aside-footer {
  border-top: 1px solid var(--oc-border);
  padding: 12px 16px 16px;
}

.main-shell {
  min-width: 0;
}

.app-header {
  height: 64px;
  border-bottom: 1px solid var(--oc-border);
  background: var(--oc-surface);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
}

.header-tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toolbar-group {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.locale-select {
  width: 120px;
}

.user-avatar {
  background: var(--oc-accent);
  color: var(--oc-accent-contrast);
}

.user-trigger {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.user-name {
  color: var(--oc-text);
  font-size: 14px;
}

.app-main {
  padding: 14px;
}

.guest-shell {
  min-height: 100vh;
}

.guest-toolbar {
  position: fixed;
  right: 14px;
  top: 12px;
  z-index: 100;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  background: var(--oc-surface);
  border: 1px solid var(--oc-border);
  border-radius: 10px;
  padding: 8px 10px;
  box-shadow: var(--oc-shadow);
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
    gap: 8px;
  }

  .header-tools {
    gap: 6px;
  }

  .toolbar-group :deep(.el-text) {
    display: none;
  }

  .locale-select {
    width: 96px;
  }

  .guest-toolbar {
    left: 8px;
    right: 8px;
    justify-content: space-between;
  }

  .user-name {
    display: none;
  }
}
</style>
