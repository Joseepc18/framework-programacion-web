import { Link } from 'react-router-dom';

const NotFoundPage = () => (
  <section aria-labelledby="titulo-404">
    <h1 id="titulo-404" className="text-2xl font-semibold">Página no encontrada</h1>
    <p className="mt-2 text-slate-600">La dirección que buscas no existe.</p>
    <Link to="/" className="mt-4 inline-block font-medium text-clinica-700 underline">
      Volver al inicio
    </Link>
  </section>
);

export default NotFoundPage;
