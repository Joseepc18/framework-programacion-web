import { Link } from 'react-router-dom';
import { AREAS, ETIQUETA_ROL, rutaDeArea } from '../config/permisos';
import { useAuth } from '../context/useAuth';

const HomePage = () => {
  const { usuario } = useAuth();
  const area = AREAS.find((a) => a.rol === usuario?.rol);

  return (
    <section aria-labelledby="titulo-inicio">
      <h1 id="titulo-inicio" className="text-pagina font-semibold">SIGO-Dental</h1>
      <div className="vidrio mt-6 max-w-2xl rounded-tarjeta p-6">
        <p className="text-texto-suave">
          Sistema de gestión odontológica y expediente clínico digital.
          {usuario && ` Iniciaste sesión como ${ETIQUETA_ROL[usuario.rol]}; el menú muestra solo lo que tu rol puede abrir.`}{' '}
          Los módulos clínicos se agregarán en las siguientes etapas.
        </p>
        {area && (
          <Link
            to={rutaDeArea(area)}
            className="mt-5 inline-flex min-h-11 items-center justify-center gap-2 rounded-control bg-acento px-5 text-cuerpo font-semibold text-sobre-acento transition hover:brightness-110 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
          >
            Ir a mi área
          </Link>
        )}
      </div>
    </section>
  );
};

export default HomePage;
