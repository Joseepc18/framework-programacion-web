import { Link } from 'react-router-dom';

const HomePage = () => (
  <section aria-labelledby="titulo-inicio">
    <h1 id="titulo-inicio" className="text-pagina font-semibold">SIGO-Dental</h1>
    <div className="vidrio mt-6 max-w-2xl rounded-tarjeta p-6">
      <p className="text-texto-suave">
        Sistema de gestión odontológica y expediente clínico digital. Esta es la base de la interfaz: los módulos
        clínicos y el inicio de sesión se agregarán en las siguientes etapas.
      </p>
      <Link
        to="/estado"
        className="inline-flex min-h-11 items-center justify-center gap-2 rounded-control bg-acento px-5 text-cuerpo font-semibold text-sobre-acento transition hover:brightness-110 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento mt-5"
      >
        Ver el estado del sistema
      </Link>
    </div>
  </section>
);

export default HomePage;
