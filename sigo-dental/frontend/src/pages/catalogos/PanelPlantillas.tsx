import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, CampoTexto, Casilla, Chip, Selector, Tarjeta } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { listarProcedimientos } from '../../services/catalogoService';
import { guardarPlantilla, listarPlantillas, type Plantilla } from '../../services/consentimientoService';

const NUEVA = { id: null as number | null, procedimiento_id: '' as number | '', titulo: '', cuerpo: '', activo: true };

// Formatos legales de consentimiento por procedimiento. Los campos {{paciente}}, {{documento}}, {{procedimiento}},
// {{pieza}}, {{profesional}} y {{fecha}} se completan solos. Los consentimientos ya preparados conservan su texto.
const PanelPlantillas = () => {
  const plantillas = useCarga(listarPlantillas, 'plantillas');
  const procedimientos = useCarga(() => listarProcedimientos(true), 'procedimientos-todos');
  const accion = useAccion();
  const [edicion, setEdicion] = useState<typeof NUEVA | null>(null);

  const editar = (p: Plantilla) => setEdicion({ id: p.id, procedimiento_id: p.procedimiento_id ?? '', titulo: p.titulo, cuerpo: p.cuerpo, activo: p.activo });

  const guardar = async (e: FormEvent) => {
    e.preventDefault();
    if (!edicion) return;
    const bien = await accion.ejecutar(() => guardarPlantilla(edicion.id, {
      procedimiento_id: edicion.procedimiento_id === '' ? null : edicion.procedimiento_id, titulo: edicion.titulo, cuerpo: edicion.cuerpo, activo: edicion.activo,
    }));
    if (bien) { setEdicion(null); plantillas.recargar(); }
  };

  return (
    <div className="space-y-6">
      {plantillas.error && <Aviso>{plantillas.error}</Aviso>}
      {!edicion && <Boton onClick={() => setEdicion(NUEVA)}>Nueva plantilla</Boton>}
      {edicion && (
        <Tarjeta>
          <form onSubmit={guardar} className="space-y-4">
            <h2 className="text-seccion font-semibold">{edicion.id ? 'Editar plantilla' : 'Nueva plantilla'}</h2>
            <div className="grid gap-4 md:grid-cols-2">
              <Selector id="pl-proc" etiqueta="Procedimiento" ayuda="Sin procedimiento es el formato general de respaldo." value={edicion.procedimiento_id}
                onChange={(e) => setEdicion({ ...edicion, procedimiento_id: e.target.value ? Number(e.target.value) : '' })}>
                <option value="">General (todos los procedimientos)</option>
                {procedimientos.datos?.map((p) => <option key={p.id} value={p.id}>{p.nombre}{p.requiere_consentimiento ? ' · exige consentimiento' : ''}</option>)}
              </Selector>
              <Campo id="pl-titulo" etiqueta="Título" required maxLength={150} value={edicion.titulo} onChange={(e) => setEdicion({ ...edicion, titulo: e.target.value })} />
            </div>
            <CampoTexto id="pl-cuerpo" etiqueta="Texto del consentimiento" required maxLength={8000} rows={10} value={edicion.cuerpo}
              onChange={(e) => setEdicion({ ...edicion, cuerpo: e.target.value })} ayuda="Campos automáticos: {{paciente}} {{documento}} {{procedimiento}} {{pieza}} {{profesional}} {{fecha}}" />
            <Casilla id="pl-activo" etiqueta="Activa" checked={edicion.activo} onChange={(e) => setEdicion({ ...edicion, activo: e.target.checked })} />
            {accion.error && <Aviso>{accion.error}</Aviso>}
            <div className="flex flex-wrap gap-3">
              <Boton type="submit" disabled={accion.ocupado}>Guardar</Boton>
              <Boton variante="secundario" onClick={() => setEdicion(null)}>Cancelar</Boton>
            </div>
          </form>
        </Tarjeta>
      )}
      <ul className="space-y-3">
        {plantillas.datos?.map((p) => (
          <li key={p.id} className="vidrio-denso space-y-2 rounded-tarjeta p-4">
            <div className="flex flex-wrap items-center gap-3">
              <span className="font-medium">{p.titulo}</span>
              <Chip>{p.procedimiento_nombre}</Chip>
              <Chip tono={p.activo ? 'acento' : 'neutro'}>{p.activo ? 'Activa' : 'Desactivada'}</Chip>
              <Boton variante="secundario" onClick={() => editar(p)}>Editar</Boton>
            </div>
            <p className="line-clamp-3 whitespace-pre-wrap text-etiqueta text-texto-suave">{p.cuerpo}</p>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default PanelPlantillas;
