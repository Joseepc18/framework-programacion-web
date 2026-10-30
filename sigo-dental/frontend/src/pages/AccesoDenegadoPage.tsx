import { Link } from 'react-router-dom';
import { ETIQUETA_ROL } from '../config/permisos';
import { useAuth } from '../context/useAuth';

const AccesoDenegadoPage = () => {
  const { usuario } = useAuth();

  return (
    <section aria-labelledby="titulo-denegado">
      <h1 id="titulo-denegado" className="text-pagina font-semibold">Acceso denegado</h1>
      <div className="vidrio mt-6 max-w-2xl rounded-tarjeta p-6">
        <p className="text-texto-suave">
          {usuario ? `Tu rol (${ETIQUETA_ROL[usuario.rol]}) no tiene permiso` : 'No tienes permiso'} para ver esta sección.
        </p>
        <Link
          to="/"
          className="mt-5 inline-flex min-h-11 items-center justify-center gap-2 rounded-control bg-acento px-5 text-cuerpo font-semibold text-sobre-acento transition hover:brightness-110 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
        >
          Volver al inicio
        </Link>
      </div>
    </section>
  );
};

export default AccesoDenegadoPage;
