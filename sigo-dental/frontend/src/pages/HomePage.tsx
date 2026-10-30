import { Link, Navigate } from 'react-router-dom';
import { ETIQUETA_ROL, menuPara } from '../config/permisos';
import { useAuth } from '../context/useAuth';

const HomePage = () => {
  const { usuario } = useAuth();
  // El paciente no tiene panel de inicio del personal: va directo a su portal.
  if (usuario?.rol === 'paciente') return <Navigate to="/portal" replace />;

  // Accesos directos: los mismos módulos que el menú, solo los que el rol puede abrir.
  const accesos = usuario ? menuPara(usuario.rol).filter((elemento) => elemento.descripcion) : [];

  return (
    <section aria-labelledby="titulo-inicio" className="space-y-6">
      <div>
        <h1 id="titulo-inicio" className="text-pagina font-semibold">SIGO-Dental</h1>
        {usuario && <p className="mt-2 text-texto-suave">Sesión iniciada como {ETIQUETA_ROL[usuario.rol]}.</p>}
      </div>
      <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        {accesos.map(({ to, label, icon: Icono, descripcion }) => (
          <li key={to}>
            <Link
              to={to}
              className="vidrio flex h-full min-h-11 items-start gap-4 rounded-tarjeta p-5 transition hover:bg-item-activo focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
            >
              <Icono className="mt-0.5 size-6 shrink-0 text-acento" aria-hidden="true" />
              <span>
                <span className="block text-seccion font-semibold">{label}</span>
                <span className="mt-1 block text-texto-suave">{descripcion}</span>
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
};

export default HomePage;
