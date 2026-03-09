import { createApp } from 'vue'
import { watch } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import { usePreferencesStore } from './stores/preferences'
import { i18n } from './i18n'
import { setupMockApi } from './mocks/mockApi'
import 'element-plus/dist/index.css'
import './styles/theme.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(i18n)

const preferences = usePreferencesStore(pinia)
preferences.bootstrap()
watch(
  () => preferences.locale,
  (locale) => {
    i18n.global.locale.value = locale
  },
  { immediate: true },
)

// 开发环境默认启用 mock（可通过 VITE_ENABLE_MOCK=0 关闭）
if (import.meta.env.DEV && import.meta.env.VITE_ENABLE_MOCK !== '0') {
  setupMockApi()
}

const auth = useAuthStore(pinia)
auth.bootstrapSession().finally(() => {
  app.use(router).use(ElementPlus).mount('#app')
})
