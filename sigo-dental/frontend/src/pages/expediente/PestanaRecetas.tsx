import { useState, type FormEvent } from 'react';
import FormularioFirma from '../../components/FormularioFirma';
import { Aviso, Boton, Campo, Casilla, Chip, Tarjeta } from '../../components/ui';
import { TriangleAlert } from 'lucide-react';
import { useAuth } from '../../context/useAuth';
import { useAccion } from '../../hooks/useAccion';
import { mensajeDeApi, useCarga } from '../../hooks/useCarga';
import {
  actualizarReceta, anularReceta, crearReceta, firmarReceta, listarRecetas, MEDICAMENTO_VACIO, pdfReceta, verificarAlergias,
  type AvisoAlergia, type Medicamento, type Receta,
} from '../../services/recetaService';
import { abrirBlob } from '../../utils/archivos';
import { fechaHoraLocal } from '../../utils/fechas';

const ETIQUETA_ESTADO = { borrador: 'Borrador', firmada: 'Firmada', anulada: 'Anulada' } as const;

const AvisosDeAlergia = ({ avisos }: { avisos: AvisoAlergia[] }) => (
  avisos.length === 0 ? null : (
    <div role="alert" className="flex gap-3 rounded-control border border-chip-alerta-borde bg-chip-alerta-fondo p-3 text-chip-alerta-texto">
      <TriangleAlert className="mt-0.5 size-5 shrink-0" aria-hidden="true" />
      <div>
        <p className="font-semibold">Posible alergia: revisa antes de recetar</p>
        <ul className="list-disc pl-5 text-etiqueta">
          {avisos.map((a) => <li key={a.medicamento}><strong>{a.medicamento}</strong> coincide con la alergia registrada en la anamnesis («{a.alergia}»).</li>)}
        </ul>
      </div>
    </div>
  )
);

interface FormularioProps {
  pacienteId: number;
  inicial?: Receta;
  alTerminar: () => void;
  alCancelar?: () => void;
}

// Formulario de receta: medicamentos en filas y aviso de alergia en cuanto se escribe el nombre de un medicamento.
const FormularioReceta = ({ pacienteId, inicial, alTerminar, alCancelar }: FormularioProps) => {
  const accion = useAccion();
  const [medicamentos, setMedicamentos] = useState<Medicamento[]>(inicial?.medicamentos ?? [{ ...MEDICAMENTO_VACIO }]);
  const [indicaciones, setIndicaciones] = useState(inicial?.indicaciones ?? '');
  // Los nombres se confirman al salir del campo: de ahí sale la consulta de alergias.
  const [confirmados, setConfirmados] = useState(() => (inicial?.medicamentos ?? []).map((m) => m.nombre).join('|'));
  const alergias = useCarga(
    () => verificarAlergias(pacienteId, confirmados.split('|').filter((n) => n.trim() !== '')),
    `alergias-${pacienteId}-${confirmados}`,
  );

  const cambiar = (i: number, campo: keyof Medicamento, valor: string) =>
    setMedicamentos(medicamentos.map((m, j) => (j === i ? { ...m, [campo]: valor } : m)));

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const datos = { indicaciones, medicamentos };
    if (await accion.ejecutar(() => (inicial ? actualizarReceta(pacienteId, inicial.id, datos) : crearReceta(pacienteId, datos)))) alTerminar();
  };

  return (
    <form onSubmit={guardar} className="space-y-4">
      <AvisosDeAlergia avisos={alergias.datos?.avisos ?? []} />
      {medicamentos.map((m, i) => (
        <fieldset key={i} className="space-y-3 rounded-control border border-vidrio-borde p-4">
          <legend className="px-2 text-etiqueta text-texto-suave">Medicamento {i + 1}</legend>
          <div className="grid gap-3 md:grid-cols-4">
            <div className="md:col-span-2">
              <Campo id={`med-nombre-${i}`} etiqueta="Nombre y concentración" required maxLength={150} value={m.nombre}
                onChange={(e) => cambiar(i, 'nombre', e.target.value)}
                onBlur={() => setConfirmados(medicamentos.map((x) => x.nombre).join('|'))} />
            </div>
            <Campo id={`med-dosis-${i}`} etiqueta="Dosis" required maxLength={100} value={m.dosis} onChange={(e) => cambiar(i, 'dosis', e.target.value)} />
            <Campo id={`med-frecuencia-${i}`} etiqueta="Frecuencia" required maxLength={100} value={m.frecuencia} onChange={(e) => cambiar(i, 'frecuencia', e.target.value)} />
            <Campo id={`med-duracion-${i}`} etiqueta="Duración" required maxLength={100} value={m.duracion} onChange={(e) => cambiar(i, 'duracion', e.target.value)} />
            <div className="md:col-span-3">
              <Campo id={`med-indicaciones-${i}`} etiqueta="Indicaciones (opcional)" maxLength={300} value={m.indicaciones} onChange={(e) => cambiar(i, 'indicaciones', e.target.value)} />
            </div>
          </div>
          {medicamentos.length > 1 && (
            <Boton variante="secundario" onClick={() => { const resto = medicamentos.filter((_, j) => j !== i); setMedicamentos(resto); setConfirmados(resto.map((x) => x.nombre).join('|')); }}>Quitar</Boton>
          )}
        </fieldset>
      ))}
      <Boton variante="secundario" disabled={medicamentos.length >= 20} onClick={() => setMedicamentos([...medicamentos, { ...MEDICAMENTO_VACIO }])}>Agregar medicamento</Boton>
      <Campo id="receta-indicaciones" etiqueta="Indicaciones generales (opcional)" maxLength={500} value={indicaciones} onChange={(e) => setIndicaciones(e.target.value)} />
      {accion.error && <Aviso>{accion.error}</Aviso>}
      <div className="flex flex-wrap gap-3">
        <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Guardando…' : inicial ? 'Guardar cambios' : 'Guardar borrador'}</Boton>
        {alCancelar && <Boton variante="secundario" onClick={alCancelar}>Cancelar</Boton>}
      </div>
    </form>
  );
};

type Panel = 'editar' | 'firmar' | 'anular' | null;

const TarjetaReceta = ({ receta, pacienteId, puedeEmitir, alCambiar }: { receta: Receta; pacienteId: number; puedeEmitir: boolean; alCambiar: () => void }) => {
  const [panel, setPanel] = useState<Panel>(null);
  const [motivo, setMotivo] = useState('');
  const [copiar, setCopiar] = useState(true);
  const accion = useAccion();
  const [errorPdf, setErrorPdf] = useState('');
  const cerrar = () => { setPanel(null); alCambiar(); };

  const abrirPdf = async () => {
    setErrorPdf('');
    try {
      abrirBlob(await pdfReceta(pacienteId, receta.id), `receta-${receta.numero}.pdf`);
    } catch (fallo) {
      setErrorPdf(mensajeDeApi(fallo));
    }
  };

  return (
    <Tarjeta className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <h3 className="font-mono text-seccion font-semibold">{receta.numero}</h3>
        <Chip tono={receta.estado === 'firmada' ? 'acento' : receta.estado === 'anulada' ? 'alerta' : 'neutro'}>{ETIQUETA_ESTADO[receta.estado]}</Chip>
        <span className="text-etiqueta text-texto-suave">{fechaHoraLocal(receta.creada_en)} · {receta.profesional_nombre}</span>
      </div>
      {receta.estado !== 'anulada' && <AvisosDeAlergia avisos={receta.avisos_alergia} />}
      {panel === 'editar' ? (
        <FormularioReceta pacienteId={pacienteId} inicial={receta} alTerminar={cerrar} alCancelar={() => setPanel(null)} />
      ) : (
        <>
          <ol className="list-decimal space-y-2 pl-5">
            {receta.medicamentos.map((m, i) => (
              <li key={i}><strong>{m.nombre}</strong><br /><span className="text-etiqueta text-texto-suave">{m.dosis} · {m.frecuencia} · {m.duracion}{m.indicaciones && ` · ${m.indicaciones}`}</span></li>
            ))}
          </ol>
          {receta.indicaciones && <p className="text-etiqueta">{receta.indicaciones}</p>}
        </>
      )}
      {receta.firma && (
        <p className="text-etiqueta">
          <Chip tono={receta.firma.valida ? 'acento' : 'alerta'}>{receta.firma.valida ? 'Firma válida' : 'CONTENIDO ALTERADO'}</Chip>{' '}
          Firmada por <strong>{receta.firma.nombre_firmante}</strong> (reg. {receta.firma.registro_profesional}) · {fechaHoraLocal(receta.firma.firmado_en)}
        </p>
      )}
      {receta.anulacion && (
        <p className="text-etiqueta text-chip-alerta-texto">Anulada por {receta.anulacion.usuario_nombre} el {fechaHoraLocal(receta.anulacion.creada_en)}. Motivo: {receta.anulacion.motivo}</p>
      )}
      {errorPdf && <Aviso>{errorPdf}</Aviso>}
      {panel === 'firmar' && <FormularioFirma descripcion={`la receta ${receta.numero}`} alCancelar={() => setPanel(null)} alFirmar={async (p) => { await firmarReceta(pacienteId, receta.id, p); cerrar(); }} />}
      {panel === 'anular' && (
        <form className="space-y-3 rounded-control border border-vidrio-borde p-4" onSubmit={async (e: FormEvent) => { e.preventDefault(); if (await accion.ejecutar(() => anularReceta(pacienteId, receta.id, motivo, copiar))) cerrar(); }}>
          <Campo id={`anular-${receta.id}`} etiqueta="Motivo de la anulación" required maxLength={300} value={motivo} onChange={(e) => setMotivo(e.target.value)} />
          <Casilla id={`copiar-${receta.id}`} etiqueta="Crear un borrador nuevo con los mismos medicamentos" checked={copiar} onChange={(e) => setCopiar(e.target.checked)} />
          {accion.error && <Aviso>{accion.error}</Aviso>}
          <div className="flex flex-wrap gap-3">
            <Boton type="submit" variante="peligro" disabled={accion.ocupado}>Anular receta</Boton>
            <Boton variante="secundario" onClick={() => setPanel(null)}>Cancelar</Boton>
          </div>
        </form>
      )}
      {!panel && (
        <div className="flex flex-wrap gap-3">
          <Boton variante="secundario" onClick={abrirPdf}>Ver PDF / imprimir</Boton>
          {puedeEmitir && receta.estado === 'borrador' && <Boton variante="secundario" onClick={() => setPanel('editar')}>Editar</Boton>}
          {puedeEmitir && receta.estado === 'borrador' && <Boton onClick={() => setPanel('firmar')}>Firmar</Boton>}
          {puedeEmitir && receta.estado !== 'anulada' && <Boton variante="secundario" onClick={() => setPanel('anular')}>Anular</Boton>}
        </div>
      )}
    </Tarjeta>
  );
};

// Recetas del paciente (F2-6): historial numerado, firma del odontólogo y anulación con registro.
const PestanaRecetas = ({ pacienteId }: { pacienteId: number }) => {
  const { usuario } = useAuth();
  const puedeEmitir = usuario?.rol === 'odontologo';
  const recetas = useCarga(() => listarRecetas(pacienteId), `recetas-${pacienteId}`);
  const [nueva, setNueva] = useState(false);

  return (
    <div className="space-y-6">
      {puedeEmitir && (
        nueva ? (
          <Tarjeta className="space-y-4">
            <h2 className="text-seccion font-semibold">Nueva receta</h2>
            <FormularioReceta pacienteId={pacienteId} alTerminar={() => { setNueva(false); recetas.recargar(); }} alCancelar={() => setNueva(false)} />
          </Tarjeta>
        ) : <Boton onClick={() => setNueva(true)}>Nueva receta</Boton>
      )}
      {recetas.error && <Aviso>{recetas.error}</Aviso>}
      {recetas.datos?.length === 0 && <p className="text-texto-suave">El paciente aún no tiene recetas.</p>}
      {recetas.datos?.map((r) => <TarjetaReceta key={`${r.id}-${r.estado}`} receta={r} pacienteId={pacienteId} puedeEmitir={puedeEmitir} alCambiar={recetas.recargar} />)}
    </div>
  );
};

export default PestanaRecetas;
