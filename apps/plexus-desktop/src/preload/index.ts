import { contextBridge, ipcRenderer } from 'electron'

// Expose a safe API surface to the renderer process
contextBridge.exposeInMainWorld('plexusAPI', {
  theme: {
    get: (): Promise<string> => ipcRenderer.invoke('theme:get'),
    set: (theme: 'light' | 'dark' | 'system'): Promise<void> =>
      ipcRenderer.invoke('theme:set', theme),
  },
  platform: process.platform,
  versions: {
    node: process.versions.node,
    chrome: process.versions.chrome,
    electron: process.versions.electron,
  },
})

// Type declaration for renderer (used in tsconfig via global augmentation)
declare global {
  interface Window {
    plexusAPI: {
      theme: {
        get(): Promise<string>
        set(theme: 'light' | 'dark' | 'system'): Promise<void>
      }
      platform: string
      versions: { node: string; chrome: string; electron: string }
    }
  }
}
