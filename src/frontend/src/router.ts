import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'

const Simple = (name: string) => ({ template: `<div>${name}</div>` })

const routes = [
  { path: '/login', component: LoginView },
  { path: '/register', component: RegisterView },
  { path: '/dashboard', component: Simple('Dashboard'), meta: { auth: true } },
  { path: '/gateway', component: Simple('Gateway'), meta: { auth: true } },
  { path: '/agents', component: Simple('Agents'), meta: { auth: true } },
  { path: '/skills', component: Simple('Skills'), meta: { auth: true } },
  { path: '/config', component: Simple('Config'), meta: { auth: true } },
  { path: '/backups', component: Simple('Backups'), meta: { auth: true } },
  { path: '/tasks', component: Simple('Tasks'), meta: { auth: true } },
  { path: '/admin/users', component: Simple('Users'), meta: { auth: true, admin: true } },
  { path: '/', redirect: '/dashboard' }
]

const router = createRouter({ history: createWebHistory(), routes })
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.auth && !auth.isAuthenticated) return '/login'
  if (to.meta.admin && auth.user?.role !== 'Admin') return '/dashboard'
})

export default router
