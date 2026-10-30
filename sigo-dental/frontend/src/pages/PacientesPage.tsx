import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import AlertasMedicas from '../components/AlertasMedicas';
import { Aviso, BotonEnlace, Campo, Casilla, Chip, Tabla, Td, Th, Titulo } from '../components/ui';
import { useAuth } from '../context/useAuth';
import { useCarga } from '../hooks/useCarga';
import { listarPacientes } from '../services/pacienteService';

const PacientesPage = () => {
  const { usuario } = useAuth();
  const esRecepcion = usuario?.rol === 'recepcionista';
  const [texto, setTexto] = useState('');
  const [busqueda, setBusqueda] = useState('');
  const [todos, setTodos] = useState(false);

  // La búsqueda se envía 300 ms después de dejar de escribir.
  useEffect(() => {
    const espera = setTimeout(() => setBusqueda(texto), 300);
    return () => clearTimeout(espera);
  }, [texto]);

  const { datos, error, cargando } = useCarga(() => listarPacientes(busqueda, todos && esRecepcion), `${busqueda}|${todos}`);

  return (
    <section aria-labelledby="titulo-pacientes" className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <Titulo id="titulo-pacientes">Pacientes</Titulo>
        {esRecepcion && <BotonEnlace to="/pacientes/nuevo">Nuevo paciente</BotonEnlace>}
      </div>

      <div className="vidrio flex flex-wrap items-end gap-x-6 gap-y-3 rounded-tarjeta p-5">
        <div className="min-w-64 flex-1">
          <Campo id="buscar-paciente" etiqueta="Buscar por nombre, documento o número de historia" type="search" value={texto} onChange={(e) => setTexto(e.target.value)} />
        </div>
        {esRecepcion && <Casilla id="ver-desactivados" etiqueta="Mostrar desactivados" checked={todos} onChange={(e) => setTodos(e.target.checked)} />}
      </div>

      {error && <Aviso>{error}</Aviso>}
      <div aria-live="polite">{cargando && <p>Buscando…</p>}</div>
      {datos && datos.length === 0 && !cargando && <p className="text-texto-suave">No hay pacientes que coincidan con la búsqueda.</p>}
      {datos && datos.length > 0 && (
        <Tabla etiqueta="Pacientes">
          <thead>
            <tr><Th>Historia</Th><Th>Paciente</Th><Th>Documento</Th><Th>Edad</Th><Th>Convenio</Th><Th>Alertas</Th><Th><span className="sr-only">Abrir</span></Th></tr>
          </thead>
          <tbody>
            {datos.map((p) => (
              <tr key={p.id}>
                <Td className="font-mono">{p.numero_historia}</Td>
                <Td>
                  <Link to={`/pacientes/${p.id}`} className="font-semibold text-texto underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-acento">
                    {p.apellidos}, {p.nombres}
                  </Link>
                  {!p.activo && <> <Chip>Desactivado</Chip></>}
                </Td>
                <Td className="font-mono">{p.documento}</Td>
                <Td className="font-mono">{p.edad}</Td>
                <Td>{p.convenio_nombre || 'Sin convenio'}</Td>
                <Td><AlertasMedicas alertas={p.alertas} compacto /></Td>
                <Td><Link to={`/pacientes/${p.id}`} className="text-acento underline-offset-4 hover:underline">Abrir expediente</Link></Td>
              </tr>
            ))}
          </tbody>
        </Tabla>
      )}
    </section>
  );
};

export default PacientesPage;
