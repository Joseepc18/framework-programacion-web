import { useEffect, useRef } from 'react';
import { Download, X } from 'lucide-react';
import { useInstallPrompt } from '../hooks/useInstallPrompt';

export default function InstalarApp({ ubicacion }: { ubicacion: 'login' | 'navbar' }) {
  const { instalada, puedeInstalar, instalando, instruccionesIOS, instalar, error } = useInstallPrompt();
  const dialogo = useRef<HTMLDialogElement>(null);
  const esNavbar = ubicacion === 'navbar';
  const tituloId = `instalar-titulo-${ubicacion}`;

  useEffect(() => {
    if (instalada) dialogo.current?.close();
  }, [instalada]);

  const solicitarInstalacion = () => {
    if (instruccionesIOS) dialogo.current?.showModal();
    else void instalar();
  };

  return (
    <>
      {!instalada && (puedeInstalar || instalando) && (
        <button
          id={`${ubicacion}-install-app`}
          type="button"
          onClick={solicitarInstalacion}
          disabled={instalando}
          aria-label="Instalar app"
          aria-haspopup={instruccionesIOS ? 'dialog' : undefined}
          title="Instala MultiCatálogo para acceder desde tu pantalla de inicio"
          className={`min-h-11 min-w-11 inline-flex items-center justify-center gap-2 rounded-lg border border-indigo-200 text-indigo-700 font-semibold hover:bg-indigo-50 transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 disabled:opacity-50 ${esNavbar ? 'px-3 text-sm' : 'mt-4 w-full px-4 py-3'}`}
        >
          <Download size={18} aria-hidden="true" />
          <span className={esNavbar ? 'hidden sm:inline' : ''}>
            {instalando ? 'Instalando…' : esNavbar ? 'Instalar' : 'Instalar app'}
          </span>
        </button>
      )}
      {error && (
        <span role="status" className={esNavbar ? 'sr-only' : 'mt-2 block text-sm text-slate-600'}>
          {error}
        </span>
      )}
      {instruccionesIOS && (
        <dialog
          ref={dialogo}
          aria-labelledby={tituloId}
          className="fixed inset-0 m-auto w-[calc(100%-2rem)] max-w-sm rounded-2xl border border-slate-200 bg-white p-6 text-slate-700 shadow-xl backdrop:bg-slate-900/40"
        >
          <div className="flex items-start justify-between gap-2">
            <h2 id={tituloId} className="pt-2 text-lg font-bold text-slate-900">Instalar MultiCatálogo</h2>
            <button
              type="button"
              aria-label="Cerrar instrucciones de instalación"
              onClick={() => dialogo.current?.close()}
              className="min-h-11 min-w-11 flex items-center justify-center rounded-lg hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-indigo-500"
            >
              <X size={20} aria-hidden="true" />
            </button>
          </div>
          <p className="mt-3 text-sm">En iPhone y iPad la instalación se realiza desde el navegador.</p>
          {instruccionesIOS === 'otro-navegador' && (
            <p className="mt-3 text-sm">Abre este mismo enlace en Safari y sigue estos pasos:</p>
          )}
          <ol className="mt-4 list-decimal space-y-3 pl-5 text-sm">
            <li>Pulsa <strong>Compartir</strong> en Safari.</li>
            <li>Selecciona <strong>Agregar a inicio</strong> (puede aparecer como <strong>Añadir a pantalla de inicio</strong>).</li>
            <li>Activa <strong>Abrir como app web</strong> si aparece y pulsa <strong>Agregar</strong>.</li>
          </ol>
          <p className="mt-4 text-sm text-slate-500">Si ya la agregaste, ábrela desde su icono en la pantalla de inicio.</p>
        </dialog>
      )}
    </>
  );
}
