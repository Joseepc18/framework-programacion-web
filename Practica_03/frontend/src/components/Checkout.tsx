// src/components/Checkout.tsx
// Flujo de checkout conectado al backend (Fase 4):
// - Formulario de envío con validación básica
// - Resumen del pedido en tiempo real
// - Envía el pedido a POST /api/checkout (transacción ACID) y redirige a /confirmacion
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { useAuth } from '../context/AuthContext';
import { procesarCheckout, type PedidoRequest } from '../services/checkoutService';

import Tarjeta from "./Tarjeta";
interface DatosEnvio {
  nombre: string;
  email: string;
  direccion: string;
  ciudad: string;
  telefono: string;
  metodoPago: string;
}

const METODOS_PAGO = ["Tarjeta de crédito", "Transferencia bancaria", "Efectivo contra entrega"];

const Checkout = () => {
  const { cart, totalPrice, clearCart } = useCart();
  const { user } = useAuth();
  const navigate = useNavigate();

  const [datos, setDatos] = useState<DatosEnvio>({
    nombre: "",
    email: "",
    direccion: "",
    ciudad: "",
    telefono: "",
    metodoPago: METODOS_PAGO[0],
  });
  const [procesando, setProcesando] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  // Si el carrito está vacío, mostramos un aviso en lugar del formulario
  if (cart.length === 0) {
    return (
      <div className="bg-papel p-10 border border-linea text-center">
        <p className="text-slate-600 text-lg mb-4">No tienes productos en el carrito.</p>
        <Link to="/catalogo" className="text-indigo-600 font-semibold hover:underline">
          ← Ir al catálogo
        </Link>
      </div>
    );
  }

  const handleChange = (campo: keyof DatosEnvio, valor: string) => {
    setDatos((prev) => ({ ...prev, [campo]: valor }));
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setProcesando(true);
    setError(null);

    // Armamos el pedido que espera el backend: datos del cliente, método de pago e items del carrito.
    const pedido: PedidoRequest = {
      cliente: {
        nombre: datos.nombre,
        email: datos.email,
        direccion: datos.direccion,
        ciudad: datos.ciudad,
        telefono: datos.telefono,
      },
      // La cuenta autenticada identifica el nodo del comprador en la red multinivel.
      usuarioEmail: user?.email ?? datos.email,
      metodoPago: datos.metodoPago,
      items: cart.map((item) => ({
        id: item.id,
        nombre: item.nombre,
        cantidad: item.cantidad,
        precio: item.precio,
      })),
    };

    try {
      // El backend guarda el pedido en una transacción ACID y responde con el pedido ya confirmado.
      const confirmado = await procesarCheckout(pedido);

      // Vaciamos el carrito y pasamos los datos del pedido a la confirmación.
      clearCart();
      navigate("/confirmacion", { state: { pedido: confirmado } });
    } catch (err) {
      // Si la transacción falló, la pantalla no avanza y el carrito se conserva.
      setError(err instanceof Error ? err.message : "No se pudo procesar el pedido");
      setProcesando(false);
    }
  };

  return (
    <div>
      <h1 className="text-3xl uppercase text-marino-700 mb-6">Finalizar Compra</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* ===== FORMULARIO DE ENVÍO ===== */}
        <form onSubmit={handleSubmit} className="lg:col-span-2 space-y-5">
          <Tarjeta titulo="Datos de Envío">

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold uppercase tracking-widest text-slate-600 mb-1">
                  Nombre completo *
                </label>
                <input
                  type="text"
                  required
                  value={datos.nombre}
                  onChange={(e) => handleChange("nombre", e.target.value)}
                  className="w-full px-4 py-2.5 border border-linea bg-papel focus:ring-2 focus:ring-marino-700 focus:border-marino-700 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold uppercase tracking-widest text-slate-600 mb-1">
                  Correo electrónico *
                </label>
                <input
                  type="email"
                  required
                  value={datos.email}
                  onChange={(e) => handleChange("email", e.target.value)}
                  className="w-full px-4 py-2.5 border border-linea bg-papel focus:ring-2 focus:ring-marino-700 focus:border-marino-700 outline-none transition"
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-xs font-semibold uppercase tracking-widest text-slate-600 mb-1">
                  Dirección *
                </label>
                <input
                  type="text"
                  required
                  value={datos.direccion}
                  onChange={(e) => handleChange("direccion", e.target.value)}
                  className="w-full px-4 py-2.5 border border-linea bg-papel focus:ring-2 focus:ring-marino-700 focus:border-marino-700 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold uppercase tracking-widest text-slate-600 mb-1">
                  Ciudad *
                </label>
                <input
                  type="text"
                  required
                  value={datos.ciudad}
                  onChange={(e) => handleChange("ciudad", e.target.value)}
                  className="w-full px-4 py-2.5 border border-linea bg-papel focus:ring-2 focus:ring-marino-700 focus:border-marino-700 outline-none transition"
                />
              </div>
              <div>
                <label className="block text-xs font-semibold uppercase tracking-widest text-slate-600 mb-1">
                  Teléfono *
                </label>
                <input
                  type="tel"
                  required
                  value={datos.telefono}
                  onChange={(e) => handleChange("telefono", e.target.value)}
                  className="w-full px-4 py-2.5 border border-linea bg-papel focus:ring-2 focus:ring-marino-700 focus:border-marino-700 outline-none transition"
                />
              </div>
            </div>
          </Tarjeta>

          <Tarjeta titulo="Método de Pago">
            <div className="space-y-3">
              {METODOS_PAGO.map((metodo) => (
                <label
                  key={metodo}
                  className="flex items-center gap-3 p-3 border border-linea cursor-pointer hover:border-marino-700 has-[:checked]:border-marino-700 has-[:checked]:bg-arena-100 transition"
                >
                  <input
                    type="radio"
                    name="metodoPago"
                    value={metodo}
                    checked={datos.metodoPago === metodo}
                    onChange={(e) => handleChange("metodoPago", e.target.value)}
                    className="accent-marino-700"
                  />
                  <span className="text-sm text-slate-700">{metodo}</span>
                </label>
              ))}
            </div>
          </Tarjeta>

          {/* Alerta de error: la compra no avanzó y el pedido no se registró */}
          {error && (
            <div className="bg-coral-50 border border-coral-200 border-l-4 border-l-coral-500 text-coral-700 px-4 py-3 text-sm">
              <p className="font-semibold">No se pudo completar el pedido</p>
              <p>{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={procesando}
            className="w-full bg-coral-500 text-marino-900 py-3 font-semibold uppercase tracking-wider hover:bg-coral-400 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {procesando ? "Procesando pago..." : "Confirmar Pedido"}
          </button>
        </form>

        {/* ===== RESUMEN DEL PEDIDO ===== */}
        <Tarjeta titulo="Resumen del Pedido" className="h-fit">
          <div className="space-y-3 border-b border-linea pb-4 mb-4">
            {cart.map((item) => (
              <div key={item.id} className="flex justify-between gap-4 text-sm">
                <span className="text-slate-600">
                  {item.nombre} × {item.cantidad}
                </span>
                <span className="font-semibold text-tinta tabular-nums">
                  ${(item.precio * item.cantidad).toFixed(2)}
                </span>
              </div>
            ))}
          </div>
          <div className="flex justify-between items-baseline">
            <span className="text-xs font-semibold uppercase tracking-widest text-slate-600">Total</span>
            <span className="font-display text-4xl text-marino-700 tabular-nums">
              ${totalPrice.toFixed(2)}
            </span>
          </div>
        </Tarjeta>
      </div>
    </div>
  );
};

export default Checkout;