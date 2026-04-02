import { useState, useEffect, type ReactNode } from 'react'
import { ThemeContext as BaseThemeContext, type ThemeType, type ThemeContextType } from './ThemeContextBase'

// Re-export for convenience
export type { ThemeType, ThemeContextType }
export const ThemeContext = BaseThemeContext

const THEME_STORAGE_KEY = 'hms_theme'

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<ThemeType>(() => {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    return (stored as ThemeType) || 'light'
  })

  const setTheme = (newTheme: ThemeType) => {
    setThemeState(newTheme)
    localStorage.setItem(THEME_STORAGE_KEY, newTheme)
  }

  useEffect(() => {
    const root = document.documentElement

    // Remove all theme classes
    root.classList.remove('theme-light', 'theme-dark', 'theme-high-contrast')

    // Add current theme class
    root.classList.add(`theme-${theme}`)

    // Also set data attribute for CSS selectors
    root.setAttribute('data-theme', theme)
  }, [theme])

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}
