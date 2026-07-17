import { useEffect, useState } from 'react'
import { Moon, Sun } from 'lucide-react'

export function ThemeToggle() {
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    const stored = localStorage.getItem('plexus-theme')
    const dark = stored === 'dark' || (!stored && window.matchMedia('(prefers-color-scheme: dark)').matches)
    setIsDark(dark)
    document.documentElement.classList.toggle('dark', dark)
    window.plexusAPI?.setTheme(dark ? 'dark' : 'light')
  }, [])

  function toggle() {
    const next = !isDark
    setIsDark(next)
    document.documentElement.classList.toggle('dark', next)
    localStorage.setItem('plexus-theme', next ? 'dark' : 'light')
    window.plexusAPI?.setTheme(next ? 'dark' : 'light')
  }

  return (
    <button
      type="button"
      onClick={toggle}
      className="flex h-8 w-8 items-center justify-center rounded text-plexus-text-subtle hover:bg-black/5 hover:text-plexus-text dark:hover:bg-white/5"
      title={isDark ? 'Light mode' : 'Dark mode'}
    >
      {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </button>
  )
}

declare global {
  interface Window {
    plexusAPI?: { setTheme: (theme: string) => void }
  }
}
