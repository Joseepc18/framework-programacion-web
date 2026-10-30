import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import './pwa/instalacion';
import App from './App';

// Aplicación instalable (portal del paciente): el service worker solo se registra en la versión de producción, para no
// complicar la depuración de la de desarrollo. Guarda el cascarón de la aplicación, nunca datos de la API.
if ('serviceWorker' in navigator && import.meta.env.PROD) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch(() => {
      // Sin service worker la aplicación funciona igual, solo que no abre sin conexión.
    });
  });
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
