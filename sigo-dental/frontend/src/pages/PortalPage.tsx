import { useState } from 'react';
import { Aviso, Boton, Chip, Tarjeta, Titulo } from '../components/ui';
import ChipEstado from '../components/ChipEstado';
import { useAccion } from '../hooks/useAccion';
import { mensajeDeApi, useCarga } from '../hooks/useCarga';
import {
  cancelarMiCita, confirmarMiCita, listarMisCitas, listarMisConsentimientos, listarMisRecetas, obtenerPerfil, pdfMiConsentimiento, pdfMiReceta,
  type CitaPortal,
} from '../services/portalService';
import { abrirBlob } from '../utils/archivos';
import { fechaHoraLocal } from '../utils/fechas';

const TarjetaCita = ({ cita, alCambiar }: { cita: CitaPortal; alCambiar: () => void }) => {
  const accion = useAccion();
  const [cancelando, setCancelando] = useState(false);
  const [motivo, setMotivo] = useState('');
  // La hora se toma al abrir la tarjeta: basta para decidir si la cita aún no empezó.
  const [ahora] = useState(() => Date.now());
  const futura = new Date(cita.inicio).getTime() > ahora;
  const cambiar = (f: () => Promise<unknown>) => accion.ejecutar(async () => { await f(); setCancelando(false); alCambiar(); });

  return (
    <li className="space-y-3 rounded-control border border-vidrio-borde p-4">
      <div className="flex flex-wrap items-center gap-3">
        <span className="font-semibold">{fechaHoraLocal(cita.inicio)}</span>
        <ChipEstado estado={cita.estado} />
      </div>
      <p className="text-etiqueta text-texto-suave">
        {cita.procedimiento_nombre || cita.motivo || 'Consulta'} · {cita.profesional_nombre} · {cita.consultorio_nombre}
      </p>
      {accion.error && <Aviso>{accion.error}</Aviso>}
      {futura && cita.estado === 'programado' && <Boton disabled={accion.ocupado} onClick={() => cambiar(() => confirmarMiCita(cita.id))}>Confirmar mi asistencia</Boton>}
      {futura && (cita.estado === 'programado' || cita.estado === 'confirmado') && !cancelando && (
        <Boton variante="secundario" onClick={() => setCancelando(true)}>Cancelar la cita</Boton>
      )}
      {cancelando && (
        <div className="space-y-3 rounded-control border border-vidrio-borde p-3">
          <label htmlFor={`motivo-${cita.id}`} className="text-etiqueta text-texto-suave">Motivo (opcional)</label>
          <input id={`motivo-${cita.id}`} maxLength={200} value={motivo} onChange={(e) => setMotivo(e.target.value)}
            className="mt-1 min-h-11 w-full rounded-control border border-chip-neutro-borde bg-chip-neutro-fondo px-4 text-cuerpo text-texto" />
          <div className="flex flex-wrap gap-3">
            <Boton variante="peligro" disabled={accion.ocupado} onClick={() => cambiar(() => cancelarMiCita(cita.id, motivo))}>Sí, cancelar</Boton>
            <Boton variante="secundario" onClick={() => setCancelando(false)}>No</Boton>
          </div>
        </div>
      )}
    </li>
  );
};

// Portal del paciente (F2-8): optimizado para el teléfono. Solo muestra lo del propio paciente.
const PortalPage = () => {
  const perfil = useCarga(obtenerPerfil, 'portal-perfil');
  const citas = useCarga(listarMisCitas, 'portal-citas');
  const recetas = useCarga(listarMisRecetas, 'portal-recetas');
  const consentimientos = useCarga(listarMisConsentimientos, 'portal-consentimientos');
  const [errorPdf, setErrorPdf] = useState('');

  const abrir = async (obtener: () => Promise<Blob>, nombre: string) => {
    setErrorPdf('');
    try {
      abrirBlob(await obtener(), nombre);
    } catch (fallo) {
      setErrorPdf(mensajeDeApi(fallo));
    }
  };

  return (
    <section className="mx-auto max-w-2xl space-y-6" aria-labelledby="portal-titulo">
      <div>
        <Titulo id="portal-titulo">Hola{perfil.datos ? `, ${perfil.datos.nombres}` : ''}</Titulo>
        {perfil.datos && <p className="mt-1 text-texto-suave">Historia clínica <span className="font-mono">{perfil.datos.numero_historia}</span></p>}
      </div>
      {errorPdf && <Aviso>{errorPdf}</Aviso>}

      <Tarjeta className="space-y-4">
        <h2 className="text-seccion font-semibold">Mis citas</h2>
        {citas.error && <Aviso>{citas.error}</Aviso>}
        {citas.datos?.length === 0 && <p className="text-texto-suave">No tienes citas recientes ni próximas.</p>}
        <ul className="space-y-3">{citas.datos?.map((c) => <TarjetaCita key={`${c.id}-${c.estado}`} cita={c} alCambiar={citas.recargar} />)}</ul>
      </Tarjeta>

      <Tarjeta className="space-y-4">
        <h2 className="text-seccion font-semibold">Mis recetas</h2>
        {recetas.error && <Aviso>{recetas.error}</Aviso>}
        {recetas.datos?.length === 0 && <p className="text-texto-suave">Aún no tienes recetas.</p>}
        <ul className="space-y-3">
          {recetas.datos?.map((r) => (
            <li key={r.id} className="space-y-2 rounded-control border border-vidrio-borde p-4">
              <div className="flex flex-wrap items-center gap-3">
                <span className="font-mono font-semibold">{r.numero}</span>
                <Chip tono={r.estado === 'anulada' ? 'alerta' : 'acento'}>{r.estado === 'anulada' ? 'Anulada' : 'Firmada'}</Chip>
                <span className="text-etiqueta text-texto-suave">{fechaHoraLocal(r.creada_en)} · {r.profesional_nombre}</span>
              </div>
              <ul className="list-disc pl-5 text-etiqueta">{r.medicamentos.map((m, i) => <li key={i}>{m.nombre}: {m.dosis}, {m.frecuencia}, {m.duracion}</li>)}</ul>
              <Boton variante="secundario" onClick={() => abrir(() => pdfMiReceta(r.id), `receta-${r.numero}.pdf`)}>Ver PDF</Boton>
            </li>
          ))}
        </ul>
      </Tarjeta>

      <Tarjeta className="space-y-4">
        <h2 className="text-seccion font-semibold">Mis consentimientos firmados</h2>
        {consentimientos.error && <Aviso>{consentimientos.error}</Aviso>}
        {consentimientos.datos?.length === 0 && <p className="text-texto-suave">Aún no tienes consentimientos firmados.</p>}
        <ul className="space-y-3">
          {consentimientos.datos?.map((c) => (
            <li key={c.id} className="space-y-2 rounded-control border border-vidrio-borde p-4">
              <div className="flex flex-wrap items-center gap-3">
                <span className="font-medium">{c.tipo === 'revocacion' ? 'Revocación · ' : ''}{c.procedimiento_nombre}</span>
                <Chip tono={c.revocado || c.tipo === 'revocacion' ? 'alerta' : 'acento'}>{c.tipo === 'revocacion' ? 'Revocación' : c.revocado ? 'Revocado' : c.vigente ? 'Vigente' : 'Firmado'}</Chip>
                {c.firmado_en && <span className="text-etiqueta text-texto-suave">{fechaHoraLocal(c.firmado_en)}</span>}
              </div>
              <Boton variante="secundario" onClick={() => abrir(() => pdfMiConsentimiento(c.id), `consentimiento-${c.id}.pdf`)}>Ver PDF</Boton>
            </li>
          ))}
        </ul>
      </Tarjeta>
    </section>
  );
};

export default PortalPage;
