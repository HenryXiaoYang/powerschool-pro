import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@/features/auth/session' // registers the API auth hooks before the first request
import { App } from '@/app/App'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
