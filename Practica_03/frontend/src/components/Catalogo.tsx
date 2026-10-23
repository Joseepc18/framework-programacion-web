// src/components/Catalogo.tsx
// Catálogo interactivo con búsqueda y filtros (Tema 5):
// - useMemo para filtrar la lista sin recalcular en cada render
// - Estado de carga y estado vacío
// - Tarjetas enlazadas al detalle /producto/:id
import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { getProductos } from '../services/productosService';
import type { Producto } from '../data/productos';
import ImagenProducto from './ImagenProducto';

const POR_PAGINA = 8;

const Catalogo = () => {
  const { addToCart } = useCart();
  const [productos, setProductos] = useState<Producto[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  // Lectura del query string (?categoria=...) enviado desde el Storefront
  const [searchParams, setSearchParams] = useSearchParams();

  // Estado de los filtros
  const [busqueda, setBusqueda] = useState<string>("");
  const [categoria, setCategoria] = useState<string>(searchParams.get("categoria") ?? "Todas");

  useEffect(() => {
    // Consumimos el servicio de productos (mock que simula la API)
    getProductos()
      .then((data) => {
        setProductos(data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  // Filtrado con useMemo: solo se recalcula cuando cambian busqueda/categoria/productos
  const productosFiltrados = useMemo(() => {
    const texto = busqueda.trim().toLowerCase();
    return productos.filter((prod) => {
      const coincideCategoria = categoria === "Todas" || prod.categoria === categoria;
      const coincideTexto =
        texto === "" ||
        prod.nombre.toLowerCase().includes(texto) ||
        prod.descripcion.toLowerCase().includes(texto);
      return coincideCategoria && coincideTexto;
    });
  }, [productos, busqueda, categoria]);

  // Categorías reales del catálogo con su cantidad de productos (no una lista fija).
  const categorias = useMemo(() => {
    const conteo = new Map<string, number>();
    for (const p of productos) conteo.set(p.categoria, (conteo.get(p.categoria) ?? 0) + 1);
    return [...conteo.entries()].sort((a, b) => a[0].localeCompare(b[0], "es"));
  }, [productos]);

  // Paginación sobre la lista ya filtrada (el catálogo completo llega cacheado desde la API)
  const [pagina, setPagina] = useState<number>(1);
  const totalPaginas = Math.max(1, Math.ceil(productosFiltrados.length / POR_PAGINA));
  const paginaActual = Math.min(pagina, totalPaginas);
  const productosPagina = productosFiltrados.slice(
    (paginaActual - 1) * POR_PAGINA,
    paginaActual * POR_PAGINA
  );

  const handleCategoriaChange = (value: string) => {
    setCategoria(value);
    setPagina(1);
    // Sincronizamos el query string para que la URL sea compartible
    if (value === "Todas") {
      setSearchParams({});
    } else {
      setSearchParams({ categoria: value });
    }
  };

  if (loading) {
    return <div className="text-center p-8">Cargando catálogo...</div>;
  }

  return (
    <div>
      <h1 className="text-3xl uppercase text-marino-700 mb-2">Catálogo de Productos</h1>
      <p className="text-slate-500 mb-6">
        {productosFiltrados.length} producto(s) encontrado(s)
      </p>

      {/* ===== BÚSQUEDA ===== */}
      <input
        type="text"
        value={busqueda}
        onChange={(e) => {
          setBusqueda(e.target.value);
          setPagina(1);
        }}
        placeholder="Buscar por nombre o descripción..."
        className="mb-4 w-full px-4 py-3 border border-linea bg-papel focus:ring-2 focus:ring-marino-700 focus:border-marino-700 outline-none transition"
      />

      {/* ===== FILTRO POR CATEGORÍA ===== */}
      <div className="mb-8 flex flex-wrap gap-2" role="group" aria-label="Filtrar por categoría">
        {[["Todas", productos.length] as const, ...categorias].map(([cat, cantidad]) => {
          const activa = categoria === cat;
          return (
            <button
              key={cat}
              onClick={() => handleCategoriaChange(cat)}
              aria-pressed={activa}
              className={`min-h-11 border px-4 text-sm font-semibold uppercase tracking-wider transition ${
                activa
                  ? "border-marino-700 bg-marino-700 text-arena-200"
                  : "border-linea bg-papel text-marino-700 hover:border-marino-700"
              }`}
            >
              {cat} <span className={activa ? "text-coral-400" : "text-slate-500"}>{cantidad}</span>
            </button>
          );
        })}
      </div>

      {/* ===== GRILLA DE PRODUCTOS ===== */}
      {productosFiltrados.length === 0 ? (
        <div className="bg-papel p-10 border border-linea text-center">
          <p className="text-slate-500">
            No hay productos que coincidan con tu búsqueda.
          </p>
          <button
            onClick={() => {
              setBusqueda("");
              setCategoria("Todas");
              setSearchParams({});
            }}
            className="mt-4 text-coral-600 font-semibold underline underline-offset-4 hover:text-coral-700"
          >
            Limpiar filtros
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {productosPagina.map((prod) => (
            <div
              key={prod.id}
              className="bg-papel overflow-hidden border border-linea flex flex-col"
            >
              <Link to={`/producto/${prod.id}`}>
                <ImagenProducto
                  src={prod.img}
                  alt={prod.nombre}
                  className="w-full h-44 object-cover hover:opacity-90 transition"
                />
              </Link>
              <div className="p-4 flex flex-col flex-1">
                <p className="text-xs uppercase tracking-widest text-coral-600 font-semibold">
                  {prod.categoria}
                </p>
                <Link to={`/producto/${prod.id}`}>
                  <h3 className="font-display text-lg text-tinta hover:text-marino-600 transition">
                    {prod.nombre}
                  </h3>
                </Link>
                <p className="text-marino-700 font-semibold mt-2 mb-4">
                  ${prod.precio.toFixed(2)}
                </p>

                <button
                  onClick={() => addToCart(prod)}
                  className="mt-auto w-full min-h-11 bg-marino-700 text-arena-200 py-2 text-sm hover:bg-marino-800 transition"
                >
                  Añadir al Carrito
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ===== PAGINACIÓN ===== */}
      {totalPaginas > 1 && (
        <div className="flex items-center justify-between mt-8 text-sm">
          <button
            onClick={() => setPagina(Math.max(1, paginaActual - 1))}
            disabled={paginaActual === 1}
            className="px-4 py-2 min-h-11 border border-marino-700 text-marino-700 font-medium disabled:opacity-40 hover:bg-marino-700 hover:text-arena-200"
          >
            Anterior
          </button>
          <span className="text-slate-500">
            Página {paginaActual} de {totalPaginas}
          </span>
          <button
            onClick={() => setPagina(Math.min(totalPaginas, paginaActual + 1))}
            disabled={paginaActual === totalPaginas}
            className="px-4 py-2 min-h-11 border border-marino-700 text-marino-700 font-medium disabled:opacity-40 hover:bg-marino-700 hover:text-arena-200"
          >
            Siguiente
          </button>
        </div>
      )}
    </div>
  );
};

export default Catalogo;