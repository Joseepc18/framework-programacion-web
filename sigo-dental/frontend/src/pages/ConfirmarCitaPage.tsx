import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Aviso, Boton } from '../components/ui';
import { useAccion } from '../hooks/useAccion';
import { confirmarPorEnlace } from '../services/citaService';

const formatoCita = new Intl.DateTimeFormat('es', { weekday: 'long', day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' });

// Página pública a la que lleva el enlace del mensaje: no pide sesión ni muestra datos del paciente.
// La confirmación se hace con un botón (no al abrir el enlace) para que una vista previa del mensaje no confirme sola.
const ConfirmarCitaPage = () => {
  const [parametros] = useSearchParams();
  const token = parametros.get('token') ?? '';
  const accion = useAccion();
  const [inicio, setInicio] = useState<string | null>(null);

  const confirmar = () =>
    accion.ejecutar(async () => {
      setInicio((await confirmarPorEnlace(token)).inicio);
    });

  return (
    <main className="grid min-h-dvh place-items-center p-4">
      <div className="vidrio w-full max-w-md rounded-tarjeta p-6 md:p-8">
        <p className="text-seccion font-semibold">
          SIGO<span className="text-acento">-Dental</span>
        </p>
        <h1 className="mt-4 text-pagina font-semibold">Confirma tu cita</h1>

        <div className="mt-5 space-y-5" aria-live="polite">
          {!token && <Aviso>El enlace está incompleto. Abre el enlace completo del mensaje que recibiste.</Aviso>}
          {token && inicio === null && (
            <>
              <p className="text-texto-suave">Confirma que asistirás a tu cita en la clínica.</p>
              {accion.error && <Aviso>{accion.error}</Aviso>}
              <Boton onClick={() => void confirmar()} disabled={accion.ocupado} className="w-full">
                {accion.ocupado ? 'Confirmando…' : 'Confirmar mi cita'}
              </Boton>
            </>
          )}
          {inicio !== null && (
            <Aviso tipo="info">
              ¡Listo! Tu cita quedó confirmada para el <span className="font-medium first-letter:uppercase">{formatoCita.format(new Date(inicio))}</span>. Te esperamos.
            </Aviso>
          )}
        </div>
      </div>
    </main>
  );
};

export default ConfirmarCitaPage;
