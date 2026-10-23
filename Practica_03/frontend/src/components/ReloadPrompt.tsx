// src/components/ReloadPrompt.tsx
// Banner que avisa cuando hay una versión nueva de la app (service worker en espera)
// y cuando la app ya quedó lista para funcionar sin conexión.
import { useRegisterSW } from 'virtual:pwa-register/react';

export default function ReloadPrompt() {
  const {
    offlineReady: [offlineReady, setOfflineReady],
    needRefresh: [needRefresh, setNeedRefresh],
    updateServiceWorker,
  } = useRegisterSW();

  const cerrar = () => {
    setOfflineReady(false);
    setNeedRefresh(false);
  };

  if (!offlineReady && !needRefresh) return null;

  return (
    <div
      role="status"
      id="reload-prompt"
      className="fixed bottom-4 right-4 left-4 sm:left-auto z-50 max-w-sm border border-marino-600 border-t-4 border-t-coral-500 bg-marino-700 p-4 text-arena-200"
    >
      <p className="text-sm">
        {needRefresh
          ? 'Hay una versión nueva disponible.'
          : 'La app ya está lista para usarse sin conexión.'}
      </p>
      <div className="mt-3 flex gap-2">
        {needRefresh && (
          <button
            id="reload-prompt-update"
            type="button"
            onClick={() => updateServiceWorker(true)}
            className="min-h-[44px] flex-1 bg-coral-500 px-4 text-sm font-semibold uppercase tracking-wider text-marino-900 hover:bg-coral-400"
          >
            Actualizar
          </button>
        )}
        <button
          id="reload-prompt-close"
          type="button"
          onClick={cerrar}
          className="min-h-[44px] flex-1 border border-arena-200/40 px-4 text-sm uppercase tracking-wider hover:bg-white/10"
        >
          Cerrar
        </button>
      </div>
    </div>
  );
}
