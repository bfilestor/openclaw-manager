import { defineStore } from 'pinia'

export type AppLocale = 'zh-CN' | 'en-US'
export type AppTheme = 'light' | 'dark'

type PersistedPreferences = {
  locale: AppLocale
  theme: AppTheme
}

const STORAGE_KEY = 'openclaw_manager_preferences_v1'
const SUPPORTED_LOCALES: AppLocale[] = ['zh-CN', 'en-US']
const SUPPORTED_THEMES: AppTheme[] = ['light', 'dark']

export const DEFAULT_LOCALE: AppLocale = 'zh-CN'
export const DEFAULT_THEME: AppTheme = 'light'

function isLocale(value: unknown): value is AppLocale {
  return typeof value === 'string' && SUPPORTED_LOCALES.includes(value as AppLocale)
}

function isTheme(value: unknown): value is AppTheme {
  return typeof value === 'string' && SUPPORTED_THEMES.includes(value as AppTheme)
}

function loadPreferences(): PersistedPreferences {
  if (typeof window === 'undefined' || !window.localStorage) {
    return { locale: DEFAULT_LOCALE, theme: DEFAULT_THEME }
  }

  let fallbackTheme = DEFAULT_THEME
  try {
    if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
      fallbackTheme = 'dark'
    }
  } catch {
    // ignore
  }

  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { locale: DEFAULT_LOCALE, theme: fallbackTheme }
    const parsed = JSON.parse(raw) as Partial<PersistedPreferences>
    return {
      locale: isLocale(parsed.locale) ? parsed.locale : DEFAULT_LOCALE,
      theme: isTheme(parsed.theme) ? parsed.theme : fallbackTheme,
    }
  } catch {
    return { locale: DEFAULT_LOCALE, theme: fallbackTheme }
  }
}

function savePreferences(data: PersistedPreferences) {
  if (typeof window === 'undefined' || !window.localStorage) return
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  } catch {
    // ignore
  }
}

function applyTheme(theme: AppTheme) {
  if (typeof document === 'undefined') return
  document.documentElement.setAttribute('data-theme', theme)
}

const persisted = loadPreferences()

export const usePreferencesStore = defineStore('preferences', {
  state: () => ({
    locale: persisted.locale as AppLocale,
    theme: persisted.theme as AppTheme,
  }),
  actions: {
    bootstrap() {
      applyTheme(this.theme)
    },
    setLocale(locale: AppLocale) {
      if (!isLocale(locale)) return
      this.locale = locale
      savePreferences({ locale: this.locale, theme: this.theme })
    },
    setTheme(theme: AppTheme) {
      if (!isTheme(theme)) return
      this.theme = theme
      applyTheme(theme)
      savePreferences({ locale: this.locale, theme: this.theme })
    },
    toggleTheme() {
      this.setTheme(this.theme === 'light' ? 'dark' : 'light')
    },
  },
})
