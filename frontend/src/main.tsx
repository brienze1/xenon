import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@/app/App'
import { installZoom } from '@/app/zoom'
import '@/shared/styles/fonts.css'
import '@/shared/styles/tokens.css'
import '@/shared/styles/global.css'

const root = document.getElementById('root')

if (root) {
  installZoom()
  createRoot(root).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}
