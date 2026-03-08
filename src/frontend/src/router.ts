import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import DashboardView from './views/DashboardView.vue'
import AdminUsersView from './views/AdminUsersView.vue'
import TasksView from './views/TasksView.vue'
import SkillsView from './views/SkillsView.vue'
import AgentsView from './views/AgentsView.vue'
import AgentSessionsView from './views/AgentSessionsView.vue'
import BindingsGraphView from './views/BindingsGraphView.vue'
import BackupsView from './views/BackupsView.vue'
import ConfigView from './views/ConfigView.vue'
import AgentWorkspaceMigrateView from './views/AgentWorkspaceMigrateView.vue'
import AgentWorkspaceFilesView from './views/AgentWorkspaceFilesView.vue'
import AgentWorkspaceFileEditorView from './views/AgentWorkspaceFileEditorView.vue'

const routes = [
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/dashboard', component: DashboardView, meta: { auth: true } },
  { path: '/gateway', component: DashboardView, meta: { auth: true } },
  { path: '/agents', component: AgentsView, meta: { auth: true } },
  { path: '/agent-sessions', component: AgentSessionsView, meta: { auth: true } },
  { path: '/agents/:id/workspace-migrate', component: AgentWorkspaceMigrateView, meta: { auth: true } },
  { path: '/agents/:id/workspace-files', component: AgentWorkspaceFilesView, meta: { auth: true } },
  { path: '/agents/:id/workspace-files/edit', component: AgentWorkspaceFileEditorView, meta: { auth: true } },
  { path: '/bindings', component: BindingsGraphView, meta: { auth: true } },
  { path: '/skills', component: SkillsView, meta: { auth: true } },
  { path: '/config', component: ConfigView, meta: { auth: true } },
  { path: '/backups', component: BackupsView, meta: { auth: true } },
  { path: '/tasks', component: TasksView, meta: { auth: true } },
  { path: '/admin/users', component: AdminUsersView, meta: { auth: true, admin: true } },
  { path: '/', redirect: '/dashboard' }
]

const router = createRouter({ history: createWebHistory(), routes })
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isAuthenticated) return '/login'
  if (to.meta.admin && auth.user?.role !== 'Admin') return '/dashboard'
})

export default router
