import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Chip, Selector, Tabla, Td, Th, Tarjeta } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { actualizarSillon, crearSillon, listarSillones, listarSimples, type Sillon } from '../../services/catalogoService';

const PanelSillones = () => {
  const sillones = useCarga(() => listarSillones(true), 'sillones');
  const consultorios = useCarga(() => listarSimples('consultorios'), 'consultorios-activos');
  const accion = useAccion();
  const [consultorioId, setConsultorioId] = useState('');
  const [nombre, setNombre] = useState('');
  const [editando, setEditando] = useState<Sillon | null>(null);

  const cancelar = () => {
    setEditando(null);
    setNombre('');
    accion.limpiar();
  };

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const bien = await accion.ejecutar(() =>
      editando ? actualizarSillon(editando.id, nombre, editando.activo) : crearSillon(Number(consultorioId), nombre),
    );
    if (bien) {
      cancelar();
      sillones.recargar();
    }
  };

  const alternar = async (sillon: Sillon) => {
    if (await accion.ejecutar(() => actualizarSillon(sillon.id, sillon.nombre, !sillon.activo))) sillones.recargar();
  };

  return (
    <div className="space-y-6">
      <Tarjeta>
        <form onSubmit={guardar} className="grid gap-4 md:grid-cols-[1fr_1fr_auto] md:items-end">
          {editando ? (
            <p className="text-texto-suave">Consultorio: <strong className="text-texto">{editando.consultorio_nombre}</strong></p>
          ) : (
            <Selector id="consultorio-sillon" etiqueta="Consultorio" required value={consultorioId} onChange={(e) => setConsultorioId(e.target.value)}>
              <option value="">Selecciona un consultorio</option>
              {consultorios.datos?.map((c) => <option key={c.id} value={c.id}>{c.nombre}</option>)}
            </Selector>
          )}
          <Campo id="nombre-sillon" etiqueta={editando ? 'Editar sillón' : 'Nuevo sillón'} value={nombre} maxLength={100} required onChange={(e) => setNombre(e.target.value)} />
          <div className="flex gap-3">
            <Boton type="submit" disabled={accion.ocupado}>{editando ? 'Guardar cambios' : 'Agregar'}</Boton>
            {editando && <Boton variante="secundario" onClick={cancelar}>Cancelar</Boton>}
          </div>
        </form>
        {accion.error && <div className="mt-4"><Aviso>{accion.error}</Aviso></div>}
      </Tarjeta>

      {sillones.error && <Aviso>{sillones.error}</Aviso>}
      {sillones.datos && (
        <Tabla etiqueta="Sillones">
          <thead><tr><Th>Consultorio</Th><Th>Sillón</Th><Th>Estado</Th><Th>Acciones</Th></tr></thead>
          <tbody>
            {sillones.datos.map((s) => (
              <tr key={s.id}>
                <Td>{s.consultorio_nombre}</Td>
                <Td>{s.nombre}</Td>
                <Td><Chip tono={s.activo ? 'acento' : 'neutro'}>{s.activo ? 'Activo' : 'Desactivado'}</Chip></Td>
                <Td className="space-x-2">
                  <Boton variante="secundario" onClick={() => { setEditando(s); setNombre(s.nombre); }}>Editar</Boton>
                  <Boton variante="secundario" disabled={accion.ocupado} onClick={() => alternar(s)}>{s.activo ? 'Desactivar' : 'Activar'}</Boton>
                </Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </div>
  );
};

export default PanelSillones;
