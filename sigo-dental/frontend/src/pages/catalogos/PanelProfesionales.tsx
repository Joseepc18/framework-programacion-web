import { useState, type FormEvent } from 'react';
import { Aviso, Boton, Campo, Chip, Selector, Tabla, Td, Th, Tarjeta } from '../../components/ui';
import { useAccion } from '../../hooks/useAccion';
import { useCarga } from '../../hooks/useCarga';
import { actualizarProfesional, crearProfesional, listarCandidatos, listarProfesionales, type Profesional } from '../../services/catalogoService';

const PanelProfesionales = () => {
  const profesionales = useCarga(() => listarProfesionales(true), 'profesionales');
  const candidatos = useCarga(listarCandidatos, 'candidatos');
  const accion = useAccion();
  const [usuarioId, setUsuarioId] = useState('');
  const [especialidad, setEspecialidad] = useState('Odontología general');
  const [registro, setRegistro] = useState('');
  const [editando, setEditando] = useState<Profesional | null>(null);

  const cancelar = () => {
    setEditando(null);
    setUsuarioId('');
    setRegistro('');
    setEspecialidad('Odontología general');
    accion.limpiar();
  };

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    const bien = await accion.ejecutar(() =>
      editando
        ? actualizarProfesional(editando.id, especialidad, registro, editando.activo)
        : crearProfesional(Number(usuarioId), especialidad, registro),
    );
    if (bien) {
      cancelar();
      profesionales.recargar();
      candidatos.recargar();
    }
  };

  const alternar = async (p: Profesional) => {
    if (await accion.ejecutar(() => actualizarProfesional(p.id, p.especialidad, p.registro_profesional, !p.activo))) {
      profesionales.recargar();
    }
  };

  return (
    <div className="space-y-6">
      <Tarjeta>
        <form onSubmit={guardar} className="grid gap-4 md:grid-cols-2 xl:grid-cols-[1fr_1fr_1fr_auto] xl:items-end">
          {editando ? (
            <p className="text-texto-suave">Cuenta: <strong className="text-texto">{editando.nombre}</strong></p>
          ) : (
            <Selector
              id="cuenta-profesional"
              etiqueta="Cuenta de odontólogo"
              ayuda={candidatos.datos?.length === 0 ? 'No hay cuentas de odontólogo sin vincular.' : undefined}
              required
              value={usuarioId}
              onChange={(e) => setUsuarioId(e.target.value)}
            >
              <option value="">Selecciona una cuenta</option>
              {candidatos.datos?.map((u) => <option key={u.id} value={u.id}>{u.nombre} ({u.email})</option>)}
            </Selector>
          )}
          <Campo id="especialidad-profesional" etiqueta="Especialidad" value={especialidad} maxLength={100} required onChange={(e) => setEspecialidad(e.target.value)} />
          <Campo id="registro-profesional" etiqueta="Registro profesional" value={registro} maxLength={50} required onChange={(e) => setRegistro(e.target.value)} />
          <div className="flex gap-3">
            <Boton type="submit" disabled={accion.ocupado}>{editando ? 'Guardar cambios' : 'Agregar'}</Boton>
            {editando && <Boton variante="secundario" onClick={cancelar}>Cancelar</Boton>}
          </div>
        </form>
        {accion.error && <div className="mt-4"><Aviso>{accion.error}</Aviso></div>}
      </Tarjeta>

      {profesionales.error && <Aviso>{profesionales.error}</Aviso>}
      {profesionales.datos && (
        <Tabla etiqueta="Profesionales">
          <thead><tr><Th>Profesional</Th><Th>Especialidad</Th><Th>Registro</Th><Th>Estado</Th><Th>Acciones</Th></tr></thead>
          <tbody>
            {profesionales.datos.map((p) => (
              <tr key={p.id}>
                <Td>{p.nombre}<br /><span className="text-etiqueta text-texto-suave">{p.email}</span></Td>
                <Td>{p.especialidad}</Td>
                <Td className="font-mono">{p.registro_profesional}</Td>
                <Td><Chip tono={p.activo ? 'acento' : 'neutro'}>{p.activo ? 'Activo' : 'Desactivado'}</Chip></Td>
                <Td className="space-x-2">
                  <Boton variante="secundario" onClick={() => { setEditando(p); setEspecialidad(p.especialidad); setRegistro(p.registro_profesional); }}>Editar</Boton>
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

export default PanelProfesionales;
