import { useEffect, useRef, useState, type ReactNode } from 'react';
import { InstallPromptContext } from '../hooks/useInstallPrompt';

interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

function estaEnModoApp() {
  return window.matchMedia('(display-mode: standalone)').matches ||
    (navigator as Navigator & { standalone?: boolean }).standalone === true;
}

function obtenerInstruccionesIOS() {
  const esIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) ||
    (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1);
  if (!esIOS) return null;
  return /Safari/.test(navigator.userAgent) &&
    !/CriOS|FxiOS|EdgiOS|OPiOS/.test(navigator.userAgent)
    ? 'safari' as const : 'otro-navegador' as const;
}

export function InstallPromptProvider({ children }: { children: ReactNode }) {
  const [evento, setEvento] = useState<BeforeInstallPromptEvent | null>(null);
  const [instalada, setInstalada] = useState(estaEnModoApp);
  const [instalando, setInstalando] = useState(false);
  const [error, setError] = useState('');
  const [instruccionesIOS] = useState(obtenerInstruccionesIOS);
  const ocupado = useRef(false);

  useEffect(() => {
    const modoApp = window.matchMedia('(display-mode: standalone)');
    const alCapturar = (e: Event) => {
      e.preventDefault();
      setEvento(e as BeforeInstallPromptEvent);
      setError('');
    };
    const alInstalar = () => {
      setInstalada(true);
      setEvento(null);
      setError('');
    };
    const alCambiarModo = () => {
      if (estaEnModoApp()) alInstalar();
    };
    window.addEventListener('beforeinstallprompt', alCapturar);
    window.addEventListener('appinstalled', alInstalar);
    modoApp.addEventListener('change', alCambiarModo);
    return () => {
      window.removeEventListener('beforeinstallprompt', alCapturar);
      window.removeEventListener('appinstalled', alInstalar);
      modoApp.removeEventListener('change', alCambiarModo);
    };
  }, []);

  const instalar = async () => {
    if (!evento || instalada || ocupado.current) return;
    ocupado.current = true;
    setInstalando(true);
    setError('');
    // Cada evento permite una sola solicitud, incluso si se cancela o falla.
    setEvento(null);
    try {
      await evento.prompt();
      await evento.userChoice;
    } catch {
      setError('No se pudo abrir la instalación. Intenta desde el menú del navegador.');
    } finally {
      ocupado.current = false;
      setInstalando(false);
    }
  };

  return (
    <InstallPromptContext.Provider value={{
      instalada,
      puedeInstalar: !instalada && (evento !== null || instruccionesIOS !== null),
      instalando,
      instruccionesIOS,
      error,
      instalar,
    }}>
      {children}
    </InstallPromptContext.Provider>
  );
}
