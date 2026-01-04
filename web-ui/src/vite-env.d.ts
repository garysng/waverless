/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ADMIN_USERNAME: string
  readonly VITE_ADMIN_PASSWORD: string
  readonly VITE_API_BACKEND_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
