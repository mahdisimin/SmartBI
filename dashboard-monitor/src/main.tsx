import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@/index.css'
import App from '@/App'

// The old auth store persisted the user's profile here; the session now lives
// in an HttpOnly cookie, so remove the stale copy. Remove this after a release.
try {
  localStorage.removeItem('auth-storage')
} catch {
  // Storage can be blocked (private mode, site-data settings) — nothing to clean
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
)