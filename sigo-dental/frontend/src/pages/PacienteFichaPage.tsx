import { useState, type FormEvent, type ReactNode } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import AnamnesisCampos from '../components/AnamnesisCampos';
import CabeceraPaciente from '../components/CabeceraPaciente';
import { Aviso, Boton, BotonEnlace, Tarjeta } from '../components/ui';
import { useAuth } from '../context/useAuth';
import { useAccion } from '../hooks/useAccion';
import { useCarga } from '../hooks/useCarga';
import { listarSimples, type ItemCatalogo } from '../services/catalogoService';
import PestanaArchivos from './expediente/PestanaArchivos';
import PestanaOdontograma from './expediente/PestanaOdontograma';
import PestanaSignos from './expediente/PestanaSignos';
import { actualizarAnamnesis, ETIQUETA_CANAL, obtenerPaciente, type Anamnesis, type PacienteDetalle } from '../services/pacienteService';

const Dato = ({ etiqueta, children }: { etiqueta: string; children: ReactNode }) => (
  <div>
    <dt className="text-etiqueta text-texto-suave">{etiqueta}</dt>
    <dd className="mt-1">{children || <span className="text-texto-suave">—</span>}</dd>
  </div>
);

const PestanaDatos = ({ p }: { p: PacienteDetalle }) => (
  <div className="space-y-6">
    <Tarjeta>
      <h2 className="text-seccion font-semibold">Datos personales</h2>
      <dl className="mt-5 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        <Dato etiqueta="Fecha de nacimiento"><span className="font-mono">{p.fecha_nacimiento}</span></Dato>
        <Dato etiqueta="Sexo"><span className="capitalize">{p.sexo}</span></Dato>
        <Dato etiqueta="Teléfono"><span className="font-mono">{p.telefono}</span></Dato>
        <Dato etiqueta="Correo">{p.email}</Dato>
        <Dato etiqueta="Dirección">{p.direccion}</Dato>
        <Dato etiqueta="Canal preferido para avisos">{ETIQUETA_CANAL[p.canal_preferido]}</Dato>
      </dl>
    </Tarjeta>
    <Tarjeta>
      <h2 className="text-seccion font-semibold">Contacto de emergencia</h2>
      <dl className="mt-5 grid gap-5 md:grid-cols-3">
        <Dato etiqueta="Nombre">{p.emergencia_nombre}</Dato>
        <Dato etiqueta="Teléfono"><span className="font-mono">{p.emergencia_telefono}</span></Dato>
        <Dato etiqueta="Parentesco">{p.emergencia_parentesco}</Dato>
      </dl>
    </Tarjeta>
    {(p.tutor_nombre || p.edad < 18) && (
      <Tarjeta>
        <h2 className="text-seccion font-semibold">Tutor legal</h2>
        <dl className="mt-5 grid gap-5 md:grid-cols-3">
          <Dato etiqueta="Nombre">{p.tutor_nombre}</Dato>
          <Dato etiqueta="Documento"><span className="font-mono">{p.tutor_documento}</span></Dato>
          <Dato etiqueta="Teléfono"><span className="font-mono">{p.tutor_telefono}</span></Dato>
        </dl>
      </Tarjeta>
    )}
  </div>
);

interface PestanaAnamnesisProps {
  p: PacienteDetalle;
  condiciones: ItemCatalogo[];
  puedeEditar: boolean;
  alGuardar: () => void;
}

const PestanaAnamnesis = ({ p, condiciones, puedeEditar, alGuardar }: PestanaAnamnesisProps) => {
  const accion = useAccion();
  const [valor, setValor] = useState<Anamnesis>(p.anamnesis);
  const [guardado, setGuardado] = useState(false);

  // Las condiciones desactivadas del catálogo ya no se ofrecen, pero el paciente las conserva al guardar.
  const ocultas = p.anamnesis.condiciones_criticas.filter((id) => !condiciones.some((c) => c.id === id));

  const guardar = async (evento: FormEvent) => {
    evento.preventDefault();
    setGuardado(false);
    const bien = await accion.ejecutar(() => actualizarAnamnesis(p.id, { ...valor, condiciones_criticas: [...valor.condiciones_criticas, ...ocultas] }));
    if (bien) {
      setGuardado(true);
      alGuardar();
    }
  };

  if (!puedeEditar) {
    return (
      <Tarjeta>
        <dl className="grid gap-5 md:grid-cols-2">
          <Dato etiqueta="Enfermedades sistémicas">{p.anamnesis.enfermedades_sistemicas}</Dato>
          <Dato etiqueta="Alergias a fármacos">{p.anamnesis.alergias_farmacos}</Dato>
          <Dato etiqueta="Antecedentes quirúrgicos">{p.anamnesis.antecedentes_quirurgicos}</Dato>
          <Dato etiqueta="Coagulopatías">{p.anamnesis.coagulopatias}</Dato>
        </dl>
        <p className="mt-5 text-etiqueta text-texto-suave">La anamnesis la actualiza el odontólogo o el asistente dental.</p>
      </Tarjeta>
    );
  }

  return (
    <form onSubmit={guardar} className="space-y-5">
      <Tarjeta><AnamnesisCampos prefijo="ficha" valor={valor} onChange={setValor} condiciones={condiciones} /></Tarjeta>
      {accion.error && <Aviso>{accion.error}</Aviso>}
      {guardado && <Aviso tipo="info">Anamnesis guardada.</Aviso>}
      <Boton type="submit" disabled={accion.ocupado}>{accion.ocupado ? 'Guardando…' : 'Guardar anamnesis'}</Boton>
    </form>
  );
};

// Pestañas del expediente. Las clínicas solo las ven el odontólogo y el asistente (la recepción ve datos y anamnesis).
interface Pestana {
  id: string;
  etiqueta: string;
  clinica?: boolean;
}

const PESTANAS: readonly Pestana[] = [
  { id: 'datos', etiqueta: 'Datos del paciente' },
  { id: 'anamnesis', etiqueta: 'Anamnesis' },
  { id: 'odontograma', etiqueta: 'Odontograma', clinica: true },
  { id: 'signos', etiqueta: 'Signos vitales', clinica: true },
  { id: 'archivos', etiqueta: 'Archivos', clinica: true },
];

// Expediente del paciente: cabecera con la alerta médica y pestañas por módulo. La pestaña activa viaja en la URL
// (?tab=) para poder enlazarla desde otras pantallas.
const PacienteFichaPage = () => {
  const { id: idRuta } = useParams();
  const id = Number(idRuta);
  const { usuario } = useAuth();
  const [parametros, setParametros] = useSearchParams();
  const paciente = useCarga(() => obtenerPaciente(id), `ficha-${id}`);
  const condiciones = useCarga(() => listarSimples('condiciones-criticas'), 'condiciones-activas');
  const esClinico = usuario?.rol === 'odontologo' || usuario?.rol === 'asistente';
  const visibles = PESTANAS.filter((t) => !t.clinica || esClinico);
  const activa = visibles.find((t) => t.id === parametros.get('tab'))?.id ?? 'datos';

  if (paciente.error) return <Aviso>{paciente.error}</Aviso>;
  if (!paciente.datos || !condiciones.datos) return <p>Cargando…</p>;
  const p = paciente.datos;

  const contenido = () => {
    switch (activa) {
      case 'anamnesis':
        return <PestanaAnamnesis key={p.id} p={p} condiciones={condiciones.datos ?? []} puedeEditar={esClinico} alGuardar={paciente.recargar} />;
      case 'odontograma':
        return <PestanaOdontograma pacienteId={p.id} edad={p.edad} puedeEditar={usuario?.rol === 'odontologo'} />;
      case 'signos':
        return <PestanaSignos pacienteId={p.id} />;
      case 'archivos':
        return <PestanaArchivos pacienteId={p.id} />;
      default:
        return <PestanaDatos p={p} />;
    }
  };

  return (
    <section className="space-y-6" aria-label={`Expediente de ${p.nombres} ${p.apellidos}`}>
      <CabeceraPaciente paciente={p} />
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div role="tablist" aria-label="Secciones del expediente" className="flex flex-wrap gap-2">
          {visibles.map((t) => (
            <button
              key={t.id}
              id={`tab-${t.id}`}
              role="tab"
              type="button"
              aria-selected={t.id === activa}
              aria-controls="panel-ficha"
              onClick={() => setParametros(t.id === 'datos' ? {} : { tab: t.id }, { replace: true })}
              className={`min-h-11 rounded-control border px-4 text-cuerpo transition focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${
                t.id === activa ? 'border-item-activo-borde bg-item-activo font-semibold text-texto' : 'border-transparent text-texto-suave hover:bg-chip-neutro-fondo hover:text-texto'
              }`}
            >
              {t.etiqueta}
            </button>
          ))}
        </div>
        {usuario?.rol === 'recepcionista' && <BotonEnlace variante="secundario" to={`/pacientes/${p.id}/editar`}>Editar datos</BotonEnlace>}
      </div>
      <div id="panel-ficha" role="tabpanel" aria-labelledby={`tab-${activa}`}>
        {contenido()}
      </div>
    </section>
  );
};

export default PacienteFichaPage;
