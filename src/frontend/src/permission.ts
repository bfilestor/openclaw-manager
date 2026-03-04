import type { App, DirectiveBinding } from 'vue'
import { useAuthStore } from './stores/auth'

const level: Record<string, number> = { Viewer: 1, Operator: 2, Admin: 3 }

export function canAccess(required: 'Viewer'|'Operator'|'Admin') {
  const s = useAuthStore()
  const role = s.user?.role || 'Viewer'
  return level[role] >= level[required]
}

export default {
  install(app: App) {
    app.directive('permission', {
      mounted(el: HTMLElement, binding: DirectiveBinding) {
        const required = (binding.value || 'Viewer') as 'Viewer'|'Operator'|'Admin'
        if (!canAccess(required)) {
          el.setAttribute('disabled', 'true')
          el.setAttribute('title', `需要 ${required} 权限`)
        }
      }
    })
  }
}
