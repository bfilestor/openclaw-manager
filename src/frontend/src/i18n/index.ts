import { createI18n } from 'vue-i18n'
import { DEFAULT_LOCALE } from '../stores/preferences'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

export const i18n = createI18n({
  legacy: false,
  locale: DEFAULT_LOCALE,
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})
