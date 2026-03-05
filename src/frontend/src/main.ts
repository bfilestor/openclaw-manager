import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import 'element-plus/dist/index.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

const auth = useAuthStore(pinia)
auth.bootstrapSession().finally(() => {
  app.use(router).use(ElementPlus).mount('#app')
})
