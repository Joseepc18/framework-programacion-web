// Instalación de la aplicación (PWA). El navegador avisa con `beforeinstallprompt` cuando la app se puede instalar; ese
// evento llega una sola vez y casi siempre antes de que React monte los componentes, así que se guarda aquí desde que
// se importa este módulo (main.tsx lo importa primero) y los componentes lo leen con `useInstalarApp`.
import { useSyncExternalStore } from 'react';

interface EventoInstalacion extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>;
}

let evento: EventoInstalacion | null = null;
const oyentes = new Set<() => void>();
const avisar = () => oyentes.forEach((oyente) => oyente());

// Con la app ya instalada y abierta como aplicación, el botón sobra.
const yaInstalada = () =>
  typeof window !== 'undefined' && window.matchMedia('(display-mode: standalone)').matches;

if (typeof window !== 'undefined') {
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    evento = e as EventoInstalacion;
    avisar();
  });
  window.addEventListener('appinstalled', () => {
    evento = null;
    avisar();
  });
}

const suscribirse = (oyente: () => void) => {
  oyentes.add(oyente);
  return () => oyentes.delete(oyente);
};

export function useInstalarApp() {
  const disponible = useSyncExternalStore(suscribirse, () => evento !== null && !yaInstalada());

  const instalar = async () => {
    if (!evento) return;
    await evento.prompt();
    // El aviso solo se puede usar una vez, la acepte o no la persona.
    evento = null;
    avisar();
  };

  return { disponible, instalar };
}
