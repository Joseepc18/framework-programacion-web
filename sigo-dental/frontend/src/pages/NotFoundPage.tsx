import { Link } from 'react-router-dom';

const NotFoundPage = () => (
  <section aria-labelledby="titulo-404">
    <h1 id="titulo-404" className="text-pagina font-semibold">Página no encontrada</h1>
    <div className="vidrio mt-6 max-w-2xl rounded-tarjeta p-6">
      <p className="text-texto-suave">La dirección que buscas no existe.</p>
      <Link
        to="/"
        className="inline-flex min-h-11 items-center justify-center gap-2 rounded-control bg-acento px-5 text-cuerpo font-semibold text-sobre-acento transition hover:brightness-110 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento mt-5"
      >
        Volver al inicio
      </Link>
    </div>
  </section>
);

export default NotFoundPage;
