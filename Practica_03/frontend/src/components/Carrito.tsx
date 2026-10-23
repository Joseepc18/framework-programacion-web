// src/components/Carrito.tsx
// Carrito de compras con controles de cantidad (Tema 5):
// - incrementQuantity / decrementQuantity desde el contexto
// - Enlace al flujo de checkout
import { Link } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import ImagenProducto from './ImagenProducto';
import Tarjeta from './Tarjeta';

const Carrito = () => {
  const {
    cart,
    removeFromCart,
    incrementQuantity,
    decrementQuantity,
    totalPrice,
  } = useCart();

  return (
    <div>
      <h1 className="text-3xl uppercase text-marino-700 mb-6">Tu Carrito de Compras</h1>

      {cart.length === 0 ? (
        <div className="bg-papel p-8 border border-linea text-center">
          <p className="text-slate-500">Tu carrito está vacío actualmente.</p>
          <Link
            to="/catalogo"
            className="inline-block mt-4 bg-marino-700 text-arena-200 font-semibold px-6 py-3 hover:bg-marino-800 transition"
          >
            Explorar Catálogo
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Lista de productos */}
          <div className="lg:col-span-2 space-y-4">
            {cart.map((item) => (
              <div
                key={item.id}
                className="flex flex-col gap-4 bg-papel p-4 border border-linea sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="flex min-w-0 items-center gap-4">
                  <ImagenProducto
                    src={item.img}
                    alt={item.nombre}
                    className="w-16 h-16 object-cover"
                  />
                  <div>
                    <Link to={`/producto/${item.id}`}>
                      <h3 className="font-display text-lg text-tinta hover:text-marino-600 transition">
                        {item.nombre}
                      </h3>
                    </Link>
                    <p className="text-sm text-slate-500">${item.precio.toFixed(2)} c/u</p>
                  </div>
                </div>

                <div className="flex items-center justify-between gap-4">
                  {/* Control de cantidad */}
                  <div className="flex items-center border border-linea overflow-hidden">
                    <button
                      onClick={() => decrementQuantity(item.id)}
                      className="min-h-11 min-w-11 px-3 py-2 text-marino-700 hover:bg-arena-100 transition"
                      aria-label="Disminuir cantidad"
                    >
                      −
                    </button>
                    <span className="flex min-h-11 min-w-11 items-center justify-center px-3 py-2 text-sm font-semibold text-slate-800 border-x border-linea">
                      {item.cantidad}
                    </span>
                    <button
                      onClick={() => incrementQuantity(item.id)}
                      className="min-h-11 min-w-11 px-3 py-2 text-marino-700 hover:bg-arena-100 transition"
                      aria-label="Aumentar cantidad"
                    >
                      +
                    </button>
                  </div>

                  <p className="font-display text-xl text-marino-700 w-24 text-right tabular-nums">
                    ${(item.precio * item.cantidad).toFixed(2)}
                  </p>

                  <button
                    onClick={() => removeFromCart(item.id)}
                    className="text-xs font-semibold uppercase tracking-wider text-coral-600 underline underline-offset-4 hover:text-coral-700 transition"
                    title="Eliminar producto"
                  >
                    Quitar
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Resumen de pago */}
          <Tarjeta titulo="Resumen del Pedido" className="h-fit">
            <div className="flex justify-between border-b border-linea pb-4 mb-4">
              <span className="text-slate-600">Subtotal</span>
              <span className="font-semibold tabular-nums">${totalPrice.toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-baseline mb-6">
              <span className="text-xs font-semibold uppercase tracking-widest text-slate-600">Total a pagar</span>
              <span className="font-display text-4xl text-marino-700 tabular-nums">
                ${totalPrice.toFixed(2)}
              </span>
            </div>
            <Link
              to="/checkout"
              className="block w-full text-center bg-coral-500 text-marino-900 py-3 font-semibold uppercase tracking-wider hover:bg-coral-400 transition"
            >
              Proceder al Pago
            </Link>
          </Tarjeta>
        </div>
      )}
    </div>
  );
};

export default Carrito;