// src/hooks/useInstallPrompt.ts
// Captura el evento `beforeinstallprompt` del navegador para poder ofrecer
// nuestro propio botón "Instalar app" en lugar de depender del icono de la barra.
import { useEffect, useState } from 'react';

// El evento no está en los tipos de TypeScript (es una API de Chromium).
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

export function useInstallPrompt() {
  const [evento, setEvento] = useState<BeforeInstallPromptEvent | null>(null);

  useEffect(() => {
    const alCapturar = (e: Event) => {
      e.preventDefault(); // evita el mini-banner automático; lo lanzamos nosotros
      setEvento(e as BeforeInstallPromptEvent);
    };
    const alInstalar = () => setEvento(null);

    window.addEventListener('beforeinstallprompt', alCapturar);
    window.addEventListener('appinstalled', alInstalar);
    return () => {
      window.removeEventListener('beforeinstallprompt', alCapturar);
      window.removeEventListener('appinstalled', alInstalar);
    };
  }, []);

  const instalar = async () => {
    if (!evento) return;
    await evento.prompt();
    await evento.userChoice;
    setEvento(null); // el evento solo se puede usar una vez
  };

  return { puedeInstalar: evento !== null, instalar };
}
