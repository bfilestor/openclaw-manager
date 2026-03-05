import { defineStore } from 'pinia'
import axios from 'axios'

export type Role = 'Viewer' | 'Operator' | 'Admin'

type SessionUser = { user_id: string; username: string; role: Role }

type PersistedSession = {
  accessToken: string
  user: SessionUser | null
}

const STORAGE_KEY = 'openclaw_manager_auth'

function loadSession(): PersistedSession {
  if (typeof window === 'undefined' || !window.localStorage) {
    return { accessToken: '', user: null }
  }
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { accessToken: '', user: null }
    const parsed = JSON.parse(raw) as PersistedSession
    if (!parsed || typeof parsed.accessToken !== 'string') {
      return { accessToken: '', user: null }
    }
    return {
      accessToken: parsed.accessToken,
      user: parsed.user ?? null
    }
  } catch {
    return { accessToken: '', user: null }
  }
}

function saveSession(session: PersistedSession) {
  if (typeof window === 'undefined' || !window.localStorage) return
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  } catch {
    // 浏览器禁用存储时静默降级为内存态
  }
}

function clearSession() {
  if (typeof window === 'undefined' || !window.localStorage) return
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    // ignore
  }
}

const persisted = loadSession()

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: persisted.accessToken as string,
    user: persisted.user as SessionUser | null,
    refreshing: null as Promise<void> | null
  }),
  getters: {
    isAuthenticated: (s) => !!s.accessToken
  },
  actions: {
    setSession(token: string, user: SessionUser) {
      this.accessToken = token
      this.user = user
      saveSession({ accessToken: token, user })
    },
    clear() {
      this.accessToken = ''
      this.user = null
      clearSession()
    },
    async ensureRefreshed() {
      if (this.refreshing) return this.refreshing
      this.refreshing = axios.post('/api/v1/auth/refresh').then((res) => {
        this.accessToken = res.data.access_token
        saveSession({ accessToken: this.accessToken, user: this.user })
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
