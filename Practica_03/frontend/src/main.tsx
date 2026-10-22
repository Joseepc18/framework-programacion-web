import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { InstallPromptProvider } from './context/InstallPromptProvider'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <InstallPromptProvider>
      <App />
    </InstallPromptProvider>
  </StrictMode>,
)
