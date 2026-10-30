import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { Aviso, Boton, Campo, Chip, Selector, Tabla, Td, Th, Titulo } from '../components/ui';
import { useCarga } from '../hooks/useCarga';
import {
  consultarAuditoria, ETIQUETA_ACCION, LIMITE_AUDITORIA, type AccionAuditoria, type FiltroAuditoria,
} from '../services/expedienteService';
import { fechaHoraLocal } from '../utils/fechas';

const FILTRO_VACIO: FiltroAuditoria = { usuario: '', paciente: '', accion: '', desde: '', hasta: '', pagina: 1 };

const ETIQUETA_ENTIDAD: Record<string, string> = {
  paciente: 'Ficha del paciente', pacientes: 'Búsqueda de pacientes', anamnesis: 'Anamnesis', signos_vitales: 'Signos vitales',
  archivo: 'Archivo', archivos: 'Lista de archivos', auditoria: 'Auditoría',
};

// Consulta de la bitácora de accesos (RNF-06): solo lectura, para el administrador TI y el auditor.
const AuditoriaPage = () => {
  const [formulario, setFormulario] = useState(FILTRO_VACIO);
  const [aplicado, setAplicado] = useState(FILTRO_VACIO);
  const pagina = useCarga(() => consultarAuditoria(aplicado), JSON.stringify(aplicado));
  const poner = (campo: keyof FiltroAuditoria) => (e: { target: { value: string } }) => setFormulario({ ...formulario, [campo]: e.target.value });

  const buscar = (evento: FormEvent) => {
    evento.preventDefault();
    setAplicado({ ...formulario, pagina: 1 });
  };
  const limpiar = () => {
    setFormulario(FILTRO_VACIO);
    setAplicado(FILTRO_VACIO);
  };

  const total = pagina.datos?.total ?? 0;
  const paginas = Math.max(1, Math.ceil(total / LIMITE_AUDITORIA));
  const irA = (n: number) => setAplicado({ ...aplicado, pagina: n });

  return (
    <section className="space-y-6" aria-labelledby="auditoria-titulo">
      <div>
        <Titulo id="auditoria-titulo">Auditoría de accesos</Titulo>
        <p className="mt-1 text-texto-suave">Registro inmutable de quién consultó, creó, modificó o exportó datos clínicos. Solo lectura.</p>
      </div>

      <form onSubmit={buscar} className="vidrio space-y-5 rounded-tarjeta p-5 md:p-6">
        <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-5">
          <Campo id="aud-usuario" etiqueta="Usuario (correo)" value={formulario.usuario} onChange={poner('usuario')} maxLength={100} />
          <Campo id="aud-paciente" etiqueta="Paciente (nombre, documento o historia)" value={formulario.paciente} onChange={poner('paciente')} maxLength={100} />
          <Selector id="aud-accion" etiqueta="Acción" value={formulario.accion} onChange={poner('accion')}>
            <option value="">Todas</option>
            {(Object.keys(ETIQUETA_ACCION) as AccionAuditoria[]).map((a) => <option key={a} value={a}>{ETIQUETA_ACCION[a]}</option>)}
          </Selector>
          <Campo id="aud-desde" etiqueta="Desde" type="date" value={formulario.desde} onChange={poner('desde')} />
          <Campo id="aud-hasta" etiqueta="Hasta" type="date" value={formulario.hasta} onChange={poner('hasta')} />
        </div>
        <div className="flex flex-wrap gap-3">
          <Boton type="submit">Filtrar</Boton>
          <Boton variante="secundario" onClick={limpiar}>Limpiar filtros</Boton>
        </div>
      </form>

      {pagina.error && <Aviso>{pagina.error}</Aviso>}
      {pagina.cargando && !pagina.datos && <p>Cargando…</p>}
      {pagina.datos && (
        <>
          <p className="text-etiqueta text-texto-suave" aria-live="polite">{total} {total === 1 ? 'registro' : 'registros'} · página {aplicado.pagina} de {paginas}</p>
          <Tabla etiqueta="Bitácora de accesos">
            <thead>
              <tr><Th>Fecha</Th><Th>Usuario</Th><Th>Acción</Th><Th>Registro</Th><Th>Paciente</Th><Th>IP</Th></tr>
            </thead>
            <tbody>
              {pagina.datos.registros.map((r) => (
                <tr key={r.id}>
                  <Td className="font-mono text-etiqueta">{fechaHoraLocal(r.ocurrido_en)}</Td>
                  <Td>{r.usuario_email || '—'}<span className="block text-etiqueta text-texto-suave">{r.rol}</span></Td>
                  <Td><Chip tono={r.accion === 'consulta' ? 'neutro' : 'acento'}>{ETIQUETA_ACCION[r.accion]}</Chip></Td>
                  <Td>{ETIQUETA_ENTIDAD[r.entidad] ?? r.entidad}{r.entidad_id !== null && <span className="font-mono text-etiqueta text-texto-suave"> #{r.entidad_id}</span>}</Td>
                  <Td>{r.paciente_id !== null ? <span>{r.paciente_nombre} <Link className="font-mono text-etiqueta text-acento underline" to={`/pacientes/${r.paciente_id}`}>#{r.paciente_id}</Link></span> : '—'}</Td>
                  <Td className="font-mono text-etiqueta">{r.ip || '—'}</Td>
                </tr>
              ))}
              {pagina.datos.registros.length === 0 && <tr><Td className="text-texto-suave">Sin registros para este filtro.</Td></tr>}
            </tbody>
          </Tabla>
          <div className="flex items-center gap-3">
            <Boton variante="secundario" disabled={aplicado.pagina <= 1} onClick={() => irA(aplicado.pagina - 1)}>Anterior</Boton>
            <Boton variante="secundario" disabled={aplicado.pagina >= paginas} onClick={() => irA(aplicado.pagina + 1)}>Siguiente</Boton>
          </div>
        </>
      )}
    </section>
  );
};

export default AuditoriaPage;
