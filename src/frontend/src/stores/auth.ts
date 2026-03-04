import { defineStore } from 'pinia'
import axios from 'axios'

export type Role = 'Viewer' | 'Operator' | 'Admin'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: '' as string,
    user: null as null | { user_id: string; username: string; role: Role },
    refreshing: null as Promise<void> | null
  }),
  getters: {
    isAuthenticated: (s) => !!s.accessToken
  },
  actions: {
    setSession(token: string, user: { user_id: string; username: string; role: Role }) {
      this.accessToken = token
      this.user = user
    },
    clear() {
      this.accessToken = ''
      this.user = null
    },
    async ensureRefreshed() {
      if (this.refreshing) return this.refreshing
      this.refreshing = axios.post('/api/v1/auth/refresh').then((res) => {
        this.accessToken = res.data.access_token
      }).catch(() => {
        this.clear()
        throw new Error('refresh failed')
      }).finally(() => {
        this.refreshing = null
      })
      return this.refreshing
    }
  }
})

axios.interceptors.request.use((cfg) => {
  const s = useAuthStore()
  if (s.accessToken) cfg.headers.Authorization = `Bearer ${s.accessToken}`
  return cfg
})

axios.interceptors.response.use((r) => r, async (err) => {
  const s = useAuthStore()
  if (err?.response?.status === 401 && s.isAuthenticated) {
    await s.ensureRefreshed()
    return axios(err.config)
  }
  return Promise.reject(err)
})
