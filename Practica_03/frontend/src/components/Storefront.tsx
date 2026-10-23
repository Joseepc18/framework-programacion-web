// src/components/Storefront.tsx
// Portada de la tienda:
// - Presentación con llamada a la acción y mosaico de productos reales
// - Categorías tomadas del catálogo (no una lista fija)
// - Productos destacados
// - Explicación del modelo multinivel
import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { getProductos } from '../services/productosService';
import type { Producto } from '../data/productos';
import ImagenProducto from './ImagenProducto';

const PASOS = [
  {
    titulo: "Explora y compra",
    texto: "Elige los productos del catálogo y genera tu primera venta.",
  },
  {
    titulo: "Invita a tu red",
    texto: "Comparte tu enlace de referido. Cada invitado suma un nivel a tu red.",
  },
  {
    titulo: "Gana comisiones",
    texto: "Recibe 10 %, 5 % y 2 % de las ventas de tus niveles 1, 2 y 3.",
  },
];

const Storefront = () => {
  const { user } = useAuth();
  const [productos, setProductos] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    getProductos()
      .then((data) => {
        setProductos(data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  const destacados = productos.slice(0, 4);

  // Categorías reales del catálogo, con la cantidad de productos de cada una.
  const categorias = useMemo(() => {
    const conteo = new Map<string, number>();
    for (const p of productos) conteo.set(p.categoria, (conteo.get(p.categoria) ?? 0) + 1);
    return [...conteo.entries()].sort((a, b) => a[0].localeCompare(b[0], "es"));
  }, [productos]);

  return (
    <div className="space-y-20 pb-8">
      {/* ===== PRESENTACIÓN ===== */}
      <section className="grid items-center gap-10 border-b border-linea pb-16 pt-4 lg:grid-cols-[1.1fr_1fr]">
        <div>
          <p className="text-xs font-semibold uppercase tracking-widest text-coral-600">
            Tienda con red de ventas
          </p>
          <h1 className="mt-4 font-display text-4xl leading-[1.1] text-tinta sm:text-5xl lg:text-6xl">
            El comercio de siempre, con una red que trabaja para ti.
          </h1>
          <p className="mt-6 max-w-xl text-lg text-slate-600">
            Compra productos de distintas categorías y, cuando invites a otras personas,
            gana una comisión por cada venta de tu red.
          </p>
          <div className="mt-8 flex flex-wrap items-center gap-x-8 gap-y-4">
            <Link
              to="/catalogo"
              className="bg-marino-700 px-8 py-4 font-semibold text-arena-200 transition hover:bg-marino-800"
            >
              Ver el catálogo
            </Link>
            {user?.rol === "admin" && (
              <Link
                to="/mi-red"
                className="font-semibold text-coral-600 underline underline-offset-4 hover:text-coral-700"
              >
                Conocer el plan de comisiones
              </Link>
            )}
          </div>
        </div>

        {/* Mosaico asimétrico con los primeros productos del catálogo */}
        <div className="grid grid-cols-2 gap-3">
          {loading
            ? [0, 1, 2].map((n) => <div key={n} className="aspect-[4/5] animate-pulse bg-slate-100" />)
            : destacados.slice(0, 3).map((prod, i) => (
                <Link
                  key={prod.id}
                  to={`/producto/${prod.id}`}
                  className={i === 0 ? "row-span-2" : ""}
                >
                  <ImagenProducto
                    src={prod.img}
                    alt={prod.nombre}
                    className={`w-full object-cover ${i === 0 ? "h-full min-h-72" : "aspect-square"}`}
                  />
                </Link>
              ))}
        </div>
      </section>

      {/* ===== CATEGORÍAS ===== */}
      <section className="grid gap-8 lg:grid-cols-[1fr_2fr]">
        <div>
          <h2 className="text-3xl text-tinta">Categorías</h2>
          <p className="mt-3 text-slate-600">
            {loading ? "Cargando el catálogo..." : `${productos.length} productos en ${categorias.length} categorías.`}
          </p>
        </div>
        <ul className="border-t border-tinta">
          {categorias.map(([cat, cantidad]) => (
            <li key={cat} className="border-b border-linea">
              <Link
                to={`/catalogo?categoria=${encodeURIComponent(cat)}`}
                className="group flex items-baseline justify-between py-4 transition hover:bg-papel"
              >
                <span className="font-display text-xl text-tinta group-hover:text-marino-600">{cat}</span>
                <span className="text-sm text-slate-500">
                  {cantidad} {cantidad === 1 ? "producto" : "productos"}
                </span>
              </Link>
            </li>
          ))}
        </ul>
      </section>

      {/* ===== PRODUCTOS DESTACADOS ===== */}
      <section>
        <div className="mb-8 flex items-end justify-between border-b border-tinta pb-3">
          <h2 className="text-3xl text-tinta">Destacados</h2>
          <Link to="/catalogo" className="text-sm font-semibold text-coral-600 underline underline-offset-4 hover:text-coral-700">
            Ver todo
          </Link>
        </div>

        {loading ? (
          <p className="text-slate-500">Cargando destacados...</p>
        ) : (
          <div className="grid grid-cols-2 gap-x-5 gap-y-10 lg:grid-cols-4">
            {destacados.map((prod) => (
              <Link key={prod.id} to={`/producto/${prod.id}`} className="group">
                <div className="overflow-hidden bg-slate-100">
                  <ImagenProducto
                    src={prod.img}
                    alt={prod.nombre}
                    className="aspect-square w-full object-cover transition-transform duration-500 group-hover:scale-105"
                  />
                </div>
                <p className="mt-3 text-xs font-semibold uppercase tracking-widest text-coral-600">
                  {prod.categoria}
                </p>
                <h3 className="mt-1 font-display text-lg text-tinta">{prod.nombre}</h3>
                <p className="mt-1 font-semibold text-marino-700">${prod.precio.toFixed(2)}</p>
              </Link>
            ))}
          </div>
        )}
      </section>

      {/* ===== CÓMO FUNCIONA ===== */}
      <section>
        <h2 className="mb-10 border-b border-tinta pb-3 text-3xl text-tinta">Cómo funciona</h2>
        <ol className="grid gap-10 md:grid-cols-3 md:gap-0 md:divide-x md:divide-linea">
          {PASOS.map((paso, i) => (
            <li key={paso.titulo} className="md:px-8 md:first:pl-0 md:last:pr-0">
              <span className="font-display text-5xl text-coral-400">{String(i + 1).padStart(2, "0")}</span>
              <h3 className="mt-3 font-display text-xl text-tinta">{paso.titulo}</h3>
              <p className="mt-2 text-slate-600">{paso.texto}</p>
            </li>
          ))}
        </ol>
      </section>

      {/* ===== LLAMADA FINAL ===== */}
      <section className="bg-marino-700 px-6 py-16 text-center text-arena-200 md:px-12">
        <h2 className="font-display text-3xl md:text-4xl">Empieza tu propia red hoy</h2>
        <p className="mx-auto mt-4 max-w-xl text-marino-200">
          Únete a JP Mercantil y convierte tu red de contactos en una fuente de ingresos.
        </p>
        <Link
          to="/catalogo"
          className="mt-8 inline-block bg-arena-200 px-10 py-4 font-semibold text-marino-800 transition hover:bg-white"
        >
          Explorar el catálogo
        </Link>
      </section>
    </div>
  );
};

export default Storefront;
