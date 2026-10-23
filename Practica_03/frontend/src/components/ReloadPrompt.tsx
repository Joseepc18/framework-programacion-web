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
      className="fixed bottom-4 right-4 left-4 sm:left-auto z-50 max-w-sm border border-white/10 bg-[#1a1033] p-4 text-white shadow-2xl"
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
            className="min-h-[44px] flex-1 bg-violet-600 px-4 text-sm font-medium hover:bg-violet-500"
          >
            Actualizar
          </button>
        )}
        <button
          id="reload-prompt-close"
          type="button"
          onClick={cerrar}
          className="min-h-[44px] flex-1 border border-white/20 px-4 text-sm hover:bg-papel/10"
        >
          Cerrar
        </button>
      </div>
    </div>
  );
}
