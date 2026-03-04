import { describe, it, expect } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'

describe('auth store', () => {
  it('set/clear session', () => {
    setActivePinia(createPinia())
    const s = useAuthStore()
    s.setSession('t', { user_id: 'u1', username: 'a', role: 'Admin' })
    expect(s.isAuthenticated).toBe(true)
    s.clear()
    expect(s.isAuthenticated).toBe(false)
  })
})
