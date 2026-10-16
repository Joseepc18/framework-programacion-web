// src/components/Catalogo.tsx
// Catálogo interactivo con búsqueda y filtros (Tema 5):
// - useMemo para filtrar la lista sin recalcular en cada render
// - Estado de carga y estado vacío
// - Tarjetas enlazadas al detalle /producto/:id
import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { getProductos } from '../services/productosService';
import { categorias, type Producto } from '../data/productos';

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
      <h1 className="text-2xl font-bold text-slate-800 mb-2">Catálogo de Productos</h1>
      <p className="text-slate-500 mb-6">
        {productosFiltrados.length} producto(s) encontrado(s)
      </p>

      {/* ===== BARRA DE BÚSQUEDA Y FILTROS ===== */}
      <div className="flex flex-col md:flex-row gap-4 mb-8">
        <input
          type="text"
          value={busqueda}
          onChange={(e) => {
            setBusqueda(e.target.value);
            setPagina(1);
          }}
          placeholder="Buscar por nombre o descripción..."
          className="flex-1 px-4 py-3 rounded-lg border border-slate-300 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
        />
        <select
          value={categoria}
          onChange={(e) => handleCategoriaChange(e.target.value)}
          className="px-4 py-3 rounded-lg border border-slate-300 bg-white focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 outline-none transition"
        >
          <option value="Todas">Todas las categorías</option>
          {categorias.map((cat) => (
            <option key={cat} value={cat}>
              {cat}
            </option>
          ))}
        </select>
      </div>

      {/* ===== GRILLA DE PRODUCTOS ===== */}
      {productosFiltrados.length === 0 ? (
        <div className="bg-white p-10 rounded-lg border border-slate-200 text-center">
          <p className="text-slate-500">
            No hay productos que coincidan con tu búsqueda.
          </p>
          <button
            onClick={() => {
              setBusqueda("");
              setCategoria("Todas");
              setSearchParams({});
            }}
            className="mt-4 text-indigo-600 font-semibold hover:underline"
          >
            Limpiar filtros
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {productosPagina.map((prod) => (
            <div
              key={prod.id}
              className="bg-white rounded-lg overflow-hidden border border-slate-200 shadow-sm flex flex-col"
            >
              <Link to={`/producto/${prod.id}`}>
                <img
                  src={prod.img}
                  alt={prod.nombre}
                  className="w-full h-40 object-cover hover:opacity-90 transition"
                />
              </Link>
              <div className="p-4 flex flex-col flex-1">
                <p className="text-xs uppercase tracking-wide text-indigo-500 font-semibold">
                  {prod.categoria}
                </p>
                <Link to={`/producto/${prod.id}`}>
                  <h3 className="font-semibold text-slate-700 hover:text-indigo-600 transition">
                    {prod.nombre}
                  </h3>
                </Link>
                <p className="text-indigo-600 font-bold mt-2 mb-4">
                  ${prod.precio.toFixed(2)}
                </p>

                <button
                  onClick={() => addToCart(prod)}
                  className="mt-auto w-full min-h-11 bg-slate-900 text-white py-2 rounded text-sm hover:bg-indigo-600 transition"
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
            className="px-4 py-2 min-h-11 rounded border border-slate-300 text-slate-700 disabled:opacity-40 hover:bg-slate-50"
          >
            Anterior
          </button>
          <span className="text-slate-500">
            Página {paginaActual} de {totalPaginas}
          </span>
          <button
            onClick={() => setPagina(Math.min(totalPaginas, paginaActual + 1))}
            disabled={paginaActual === totalPaginas}
            className="px-4 py-2 min-h-11 rounded border border-slate-300 text-slate-700 disabled:opacity-40 hover:bg-slate-50"
          >
            Siguiente
          </button>
        </div>
      )}
    </div>
  );
};

export default Catalogo;