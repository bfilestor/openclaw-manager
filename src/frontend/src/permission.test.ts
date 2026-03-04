import { describe, it, expect } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './stores/auth'
import { canAccess } from './permission'

describe('permission', () => {
  it('role compare', () => {
    setActivePinia(createPinia())
    const s = useAuthStore()
    s.setSession('t', { user_id: 'u1', username: 'a', role: 'Operator' })
    expect(canAccess('Viewer')).toBe(true)
    expect(canAccess('Admin')).toBe(false)
  })
})
