import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import DashboardView from './views/DashboardView.vue'
import GatewayView from './views/GatewayView.vue'
import AdminUsersView from './views/AdminUsersView.vue'
import TasksView from './views/TasksView.vue'
import TaskShellView from './views/TaskShellView.vue'
import SkillsView from './views/SkillsView.vue'
import AgentsView from './views/AgentsView.vue'
/*import AgentSessionsView from './views/AgentSessionsView.vue'*/
import BindingsGraphView from './views/BindingsGraphView.vue'
import BackupsView from './views/BackupsView.vue'
import ConfigView from './views/ConfigView.vue'
import AgentWorkspaceMigrateView from './views/AgentWorkspaceMigrateView.vue'
import AgentWorkspaceFilesView from './views/AgentWorkspaceFilesView.vue'
import AgentWorkspaceFileEditorView from './views/AgentWorkspaceFileEditorView.vue'
import QQBotManageView from './views/QQBotManageView.vue'
import ApiProvidersView from './views/ApiProvidersView.vue'
import TokenUsageView from './views/TokenUsageView.vue'
import BotConversationDetailView from './views/BotConversationDetailView.vue'

const routes = [
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/dashboard', component: DashboardView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/gateway', component: GatewayView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/agents', component: AgentsView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  /*{ path: '/agent-sessions', component: AgentSessionsView, meta: { auth: true } },*/
  { path: '/agents/:id/workspace-migrate', component: AgentWorkspaceMigrateView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/agents/:id/workspace-files', component: AgentWorkspaceFilesView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/agents/:id/workspace-files/edit', component: AgentWorkspaceFileEditorView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/bindings', component: BindingsGraphView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/skills', component: SkillsView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/config', component: ConfigView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/qqbot', component: QQBotManageView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/api-providers', component: ApiProvidersView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/token-usage', component: TokenUsageView, meta: { auth: true, allowedRoles: ['User', 'Viewer', 'Operator', 'Admin'] } },
  { path: '/token-usage/:botId', component: BotConversationDetailView, meta: { auth: true, allowedRoles: ['User', 'Viewer', 'Operator', 'Admin'] } },
  { path: '/backups', component: BackupsView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/shell', component: TaskShellView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/tasks', component: TasksView, meta: { auth: true, allowedRoles: ['Viewer', 'Operator', 'Admin'] } },
  { path: '/admin/users', component: AdminUsersView, meta: { auth: true, admin: true, allowedRoles: ['Admin'] } },
  { path: '/', redirect: '/dashboard' }
]

const router = createRouter({ history: createWebHistory(), routes })
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isAuthenticated) return '/login'
  const role = auth.user?.role || 'Viewer'
  const allowedRoles = (to.meta as any)?.allowedRoles as string[] | undefined
  if (allowedRoles && !allowedRoles.includes(role)) {
    return role === 'User' ? '/token-usage' : '/dashboard'
  }
  if (to.meta.admin && auth.user?.role !== 'Admin') return '/dashboard'
})

export default router
