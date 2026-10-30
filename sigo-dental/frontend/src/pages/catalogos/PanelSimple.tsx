import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Chip, Tabla, Td, Th, Tarjeta } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { actualizarSimple, crearSimple, listarSimples, type ItemCatalogo, type RecursoSimple } from '../../services/catalogoService';

// Catálogo de solo nombre (consultorios, convenios y condiciones críticas): lista, alta, edición y desactivación.
const PanelSimple = ({ recurso, singular }: { recurso: RecursoSimple; singular: string }) => {
  const { datos, error, cargando, recargar } = useCarga(() => listarSimples(recurso, true), recurso);
  const accion = useAccion();
  const [nombre, setNombre] = useState('');
  const [editando, setEditando] = useState<ItemCatalogo | null>(null);

  const cancelar = () => {
    setEditando(null);
    setNombre('');
    accion.limpiar();
  };

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const bien = await accion.ejecutar(() =>
      editando ? actualizarSimple(recurso, editando.id, nombre, editando.activo) : crearSimple(recurso, nombre),
    );
    if (bien) {
      cancelar();
      recargar();
    }
  };

  const alternar = async (item: ItemCatalogo) => {
    if (await accion.ejecutar(() => actualizarSimple(recurso, item.id, item.nombre, !item.activo))) recargar();
  };

  return (
    <div className="space-y-6">
      <Tarjeta>
        <form onSubmit={guardar} className="grid gap-4 md:grid-cols-[1fr_auto] md:items-end">
          <Campo id={`nombre-${recurso}`} etiqueta={editando ? `Editar ${singular}` : `Nuevo ${singular}`} value={nombre} maxLength={100} required onChange={(e) => setNombre(e.target.value)} />
          <div className="flex gap-3">
            <Boton type="submit" disabled={accion.ocupado}>{editando ? 'Guardar cambios' : 'Agregar'}</Boton>
            {editando && <Boton variante="secundario" onClick={cancelar}>Cancelar</Boton>}
          </div>
        </form>
        {accion.error && <div className="mt-4"><Aviso>{accion.error}</Aviso></div>}
      </Tarjeta>

      {error && <Aviso>{error}</Aviso>}
      {cargando && !datos && <p>Cargando…</p>}
      {datos && (
        <Tabla etiqueta={recurso}>
          <thead><tr><Th>Nombre</Th><Th>Estado</Th><Th>Acciones</Th></tr></thead>
          <tbody>
            {datos.map((item) => (
              <tr key={item.id}>
                <Td>{item.nombre}</Td>
                <Td><Chip tono={item.activo ? 'acento' : 'neutro'}>{item.activo ? 'Activo' : 'Desactivado'}</Chip></Td>
                <Td className="space-x-2">
                  <Boton variante="secundario" onClick={() => { setEditando(item); setNombre(item.nombre); }}>Editar</Boton>
                  <Boton variante="secundario" disabled={accion.ocupado} onClick={() => alternar(item)}>{item.activo ? 'Desactivar' : 'Activar'}</Boton>
                </Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </div>
  );
};

export default PanelSimple;
