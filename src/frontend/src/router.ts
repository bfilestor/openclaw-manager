import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import DashboardView from './views/DashboardView.vue'
import AdminUsersView from './views/AdminUsersView.vue'
import TasksView from './views/TasksView.vue'
import BusinessViews from './views/BusinessViews.vue'

const Simple = (name: string) => ({ template: `<div>${name}</div>` })

const routes = [
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/dashboard', component: DashboardView, meta: { auth: true } },
  { path: '/gateway', component: DashboardView, meta: { auth: true } },
  { path: '/agents', component: BusinessViews, meta: { auth: true } },
  { path: '/skills', component: BusinessViews, meta: { auth: true } },
  { path: '/config', component: BusinessViews, meta: { auth: true } },
  { path: '/backups', component: BusinessViews, meta: { auth: true } },
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
