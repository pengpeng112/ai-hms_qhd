import { createContext } from 'react'

export type ThemeType = 'light' | 'dark' | 'high-contrast'

export interface ThemeContextType {
  theme: ThemeType
  setTheme: (theme: ThemeType) => void
}

export const ThemeContext = createContext<ThemeContextType | undefined>(undefined)
