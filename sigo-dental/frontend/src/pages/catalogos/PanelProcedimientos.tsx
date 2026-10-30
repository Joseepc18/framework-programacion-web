import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Casilla, Chip, Tabla, Td, Th, Tarjeta } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { actualizarProcedimiento, crearProcedimiento, listarProcedimientos, type Procedimiento } from '../../services/catalogoService';
import { aCentavos, deCentavos, formatearDinero } from '../../utils/dinero';

const VACIO = { codigo: '', nombre: '', especialidad: 'Odontología general', precio: '', duracion: '30', consentimiento: false };

const PanelProcedimientos = () => {
  const { datos, error, recargar } = useCarga(() => listarProcedimientos(true), 'procedimientos');
  const accion = useAccion();
  const [form, setForm] = useState(VACIO);
  const [editando, setEditando] = useState<Procedimiento | null>(null);
  const [errorForm, setErrorForm] = useState('');

  const cancelar = () => {
    setEditando(null);
    setForm(VACIO);
    setErrorForm('');
    accion.limpiar();
  };

  const cambiar = (campo: keyof typeof VACIO, valor: string | boolean) => setForm((f) => ({ ...f, [campo]: valor }));

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const precio = aCentavos(form.precio);
    const duracion = Number(form.duracion);
    if (precio === null) return setErrorForm('El precio debe ser un importe válido, por ejemplo 45.50.');
    if (!Number.isInteger(duracion) || duracion <= 0) return setErrorForm('La duración debe ser un número entero de minutos mayor que cero.');
    setErrorForm('');
    const datosEnvio = {
      codigo: form.codigo,
      nombre: form.nombre,
      especialidad: form.especialidad,
      precio_centavos: precio,
      duracion_min: duracion,
      requiere_consentimiento: form.consentimiento,
      activo: editando?.activo ?? true,
    };
    const bien = await accion.ejecutar(() => (editando ? actualizarProcedimiento(editando.id, datosEnvio) : crearProcedimiento(datosEnvio)));
    if (bien) {
      cancelar();
      recargar();
    }
  };

  const alternar = async (p: Procedimiento) => {
    const datosEnvio = {
      codigo: p.codigo,
      nombre: p.nombre,
      especialidad: p.especialidad,
      precio_centavos: p.precio_centavos,
      duracion_min: p.duracion_min,
      requiere_consentimiento: p.requiere_consentimiento,
      activo: !p.activo,
    };
    if (await accion.ejecutar(() => actualizarProcedimiento(p.id, datosEnvio))) recargar();
  };

  const editar = (p: Procedimiento) => {
    setEditando(p);
    setForm({ codigo: p.codigo, nombre: p.nombre, especialidad: p.especialidad, precio: deCentavos(p.precio_centavos), duracion: String(p.duracion_min), consentimiento: p.requiere_consentimiento });
  };

  return (
    <div className="space-y-6">
      <Tarjeta>
        <form onSubmit={guardar} className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <Campo id="codigo-proc" etiqueta="Código" value={form.codigo} maxLength={20} required onChange={(e) => cambiar('codigo', e.target.value)} />
          <Campo id="nombre-proc" etiqueta="Nombre" value={form.nombre} maxLength={150} required onChange={(e) => cambiar('nombre', e.target.value)} />
          <Campo id="especialidad-proc" etiqueta="Especialidad" value={form.especialidad} maxLength={100} required onChange={(e) => cambiar('especialidad', e.target.value)} />
          <Campo id="precio-proc" etiqueta="Precio" ayuda="En dólares, por ejemplo 45.50" inputMode="decimal" value={form.precio} required onChange={(e) => cambiar('precio', e.target.value)} />
          <Campo id="duracion-proc" etiqueta="Duración (minutos)" type="number" min={1} value={form.duracion} required onChange={(e) => cambiar('duracion', e.target.value)} />
          <div className="flex items-end"><Casilla id="consentimiento-proc" etiqueta="Requiere consentimiento informado" checked={form.consentimiento} onChange={(e) => cambiar('consentimiento', e.target.checked)} /></div>
          <div className="flex gap-3 md:col-span-2 xl:col-span-3">
            <Boton type="submit" disabled={accion.ocupado}>{editando ? 'Guardar cambios' : 'Agregar procedimiento'}</Boton>
            {editando && <Boton variante="secundario" onClick={cancelar}>Cancelar</Boton>}
          </div>
        </form>
        {(errorForm || accion.error) && <div className="mt-4"><Aviso>{errorForm || accion.error}</Aviso></div>}
      </Tarjeta>

      {error && <Aviso>{error}</Aviso>}
      {datos && (
        <Tabla etiqueta="Procedimientos">
          <thead><tr><Th>Código</Th><Th>Procedimiento</Th><Th>Precio</Th><Th>Duración</Th><Th>Consentimiento</Th><Th>Estado</Th><Th>Acciones</Th></tr></thead>
          <tbody>
            {datos.map((p) => (
              <tr key={p.id}>
                <Td className="font-mono">{p.codigo}</Td>
                <Td>{p.nombre}<br /><span className="text-etiqueta text-texto-suave">{p.especialidad}</span></Td>
                <Td className="font-mono">{formatearDinero(p.precio_centavos)}</Td>
                <Td className="font-mono">{p.duracion_min} min</Td>
                <Td>{p.requiere_consentimiento ? 'Sí' : 'No'}</Td>
                <Td><Chip tono={p.activo ? 'acento' : 'neutro'}>{p.activo ? 'Activo' : 'Desactivado'}</Chip></Td>
                <Td className="space-x-2">
                  <Boton variante="secundario" onClick={() => editar(p)}>Editar</Boton>
                  <Boton variante="secundario" disabled={accion.ocupado} onClick={() => alternar(p)}>{p.activo ? 'Desactivar' : 'Activar'}</Boton>
                </Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </div>
  );
};

export default PanelProcedimientos;
