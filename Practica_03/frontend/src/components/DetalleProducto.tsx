// src/components/DetalleProducto.tsx
// Detalle de producto con ruta dinámica (Tema 5):
// - useParams para leer el :id de la URL (/producto/:id)
// - Galería de imágenes con lightbox a pantalla completa
// - Estados de carga, error y "no encontrado"
import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { getProductoById } from '../services/productosService';
import type { Producto } from '../data/productos';
import { useCart } from '../context/CartContext';
import ImagenProducto from './ImagenProducto';

const DetalleProducto = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addToCart } = useCart();

  const [producto, setProducto] = useState<Producto | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [imagenActiva, setImagenActiva] = useState<number>(0);
  const [lightboxAbierto, setLightboxAbierto] = useState<boolean>(false);
  const [agregado, setAgregado] = useState<boolean>(false);

  useEffect(() => {
    // La ruta dinámica nos da el id como string; lo convertimos a número.
    // La bandera "activo" evita actualizar estado si el componente se desmonta.
    let activo = true;
    getProductoById(Number(id))
      .then((data) => {
        if (!activo) return;
        setProducto(data ?? null);
        setLoading(false);
      })
      .catch(() => {
        if (activo) setLoading(false);
      });
    return () => {
      activo = false;
    };
  }, [id]);

  // Cerrar el lightbox con la tecla Escape
  useEffect(() => {
    if (!lightboxAbierto) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setLightboxAbierto(false);
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [lightboxAbierto]);

  if (loading) {
    return <div className="text-center p-8">Cargando producto...</div>;
  }

  if (!producto) {
    return (
      <div className="bg-papel p-10 border border-linea text-center">
        <p className="text-slate-600 text-lg mb-4">Producto no encontrado.</p>
        <Link to="/catalogo" className="text-coral-600 font-semibold underline underline-offset-4 hover:text-coral-700">
          ← Volver al catálogo
        </Link>
      </div>
    );
  }

  // Todas las imágenes de la galería: principal + adicionales
  const imagenes = [producto.img, ...producto.galeria];

  const handleAddToCart = () => {
    addToCart(producto);
    setAgregado(true);
    setTimeout(() => setAgregado(false), 1500);
  };

  const handlePrev = () =>
    setImagenActiva((prev) => (prev === 0 ? imagenes.length - 1 : prev - 1));
  const handleNext = () =>
    setImagenActiva((prev) => (prev === imagenes.length - 1 ? 0 : prev + 1));

  return (
    <div>
      <button
        onClick={() => navigate(-1)}
        className="mb-6 text-sm font-semibold uppercase tracking-widest text-marino-700 hover:text-coral-600"
      >
        ← Volver
      </button>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* ===== GALERÍA DE IMÁGENES ===== */}
        <div>
          <button
            type="button"
            onClick={() => setLightboxAbierto(true)}
            className="block w-full cursor-zoom-in border border-linea bg-papel p-2"
            aria-label="Ver imagen a pantalla completa"
          >
            <ImagenProducto
              src={imagenes[imagenActiva]}
              alt={producto.nombre}
              className="h-80 w-full object-cover md:h-96"
            />
          </button>
          <div className="flex gap-3 mt-4">
            {imagenes.map((img, idx) => (
              <button
                key={img}
                onClick={() => setImagenActiva(idx)}
                className={`overflow-hidden border-2 transition ${idx === imagenActiva
                    ? "border-coral-500"
                    : "border-transparent opacity-70 hover:opacity-100"
                  }`}
              >
                <ImagenProducto src={img} alt={`Vista ${idx + 1}`} className="w-20 h-16 object-cover" />
              </button>
            ))}
          </div>
        </div>

        {/* ===== INFORMACIÓN DEL PRODUCTO ===== */}
        <div>
          <p className="text-xs font-semibold uppercase tracking-widest text-coral-600">
            {producto.categoria}
          </p>
          <h1 className="mt-2 font-display text-4xl leading-tight text-tinta md:text-5xl">{producto.nombre}</h1>
          <div className="my-5 h-1 w-16 bg-coral-500" />
          <p className="mb-6 font-display text-5xl text-marino-700 tabular-nums">
            ${producto.precio.toFixed(2)}
          </p>
          <p className="text-slate-600 leading-relaxed mb-8">{producto.descripcion}</p>

          <button
            onClick={handleAddToCart}
            className={`w-full md:w-auto px-10 py-4 font-semibold uppercase tracking-wider transition ${agregado
                ? "bg-turquesa-700 text-white"
                : "bg-marino-700 text-arena-200 hover:bg-marino-800"
              }`}
          >
            {agregado ? "✓ Añadido al carrito" : "Añadir al Carrito"}
          </button>
        </div>
      </div>

      {/* ===== LIGHTBOX: imagen a pantalla completa ===== */}
      {lightboxAbierto && (
        <div
          className="fixed inset-0 z-50 bg-slate-950/95 flex items-center justify-center"
          onClick={() => setLightboxAbierto(false)}
        >
          <button
            className="absolute top-5 right-5 text-white text-4xl hover:text-slate-300 transition"
            onClick={() => setLightboxAbierto(false)}
            aria-label="Cerrar"
          >
            ×
          </button>

          <button
            onClick={(e) => {
              e.stopPropagation();
              handlePrev();
            }}
            className="absolute left-4 md:left-10 text-white text-4xl hover:text-indigo-400 transition"
            aria-label="Anterior"
          >
            ‹
          </button>

          <img
            src={imagenes[imagenActiva]}
            alt={producto.nombre}
            onClick={(e) => e.stopPropagation()}
            className="max-h-[85vh] max-w-[85vw] object-contain"
          />

          <button
            onClick={(e) => {
              e.stopPropagation();
              handleNext();
            }}
            className="absolute right-4 md:right-10 text-white text-4xl hover:text-indigo-400 transition"
            aria-label="Siguiente"
          >
            ›
          </button>

          <p className="absolute bottom-6 text-white/80 text-sm">
            {imagenActiva + 1} / {imagenes.length}
          </p>
        </div>
      )}
    </div>
  );
};

export default DetalleProducto;