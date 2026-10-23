// src/components/Confirmacion.tsx
// Pantalla de confirmación de pedido (Tema 5):
// - Lee los datos del pedido enviados desde Checkout (location.state)
// - Muestra el número de pedido y el resumen
import { Link, useLocation } from 'react-router-dom';

interface PedidoConfirmado {
  numero: string;
  items: { nombre: string; cantidad: number; precio: number }[];
  total: number;
  cliente: { nombre: string; email: string; ciudad: string };
  fecha: string;
}

const Confirmacion = () => {
  const location = useLocation();
  const pedido = (location.state as { pedido?: PedidoConfirmado } | null)?.pedido;

  // Si no hay datos de pedido (acceso directo a la URL), mostramos un mensaje genérico
  if (!pedido) {
    return (
      <div className="bg-papel p-10 border border-linea text-center">
        <h1 className="text-3xl uppercase text-marino-700 mb-2">¡Gracias por tu compra!</h1>
        <p className="text-slate-500 mb-6">Tu pedido ha sido registrado correctamente.</p>
        <div className="flex flex-wrap justify-center gap-3">
          <Link
            to="/mis-pedidos"
            className="inline-block bg-turquesa-700 text-white font-semibold px-6 py-3 hover:bg-turquesa-800 transition"
          >
            Ver mis pedidos
          </Link>
          <Link
            to="/tienda"
            className="inline-block bg-marino-700 text-arena-200 font-semibold px-6 py-3 hover:bg-marino-800 transition"
          >
            Volver a la Tienda
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-papel border border-linea overflow-hidden">
        <div className="border-b-4 border-coral-500 bg-marino-700 p-8 text-center text-arena-200">
          <h1 className="font-display text-4xl uppercase">¡Pedido confirmado!</h1>
          <p className="mt-2 text-marino-200">
            Número de pedido: <span className="font-mono font-bold text-coral-400">{pedido.numero}</span>
          </p>
        </div>

        <div className="p-8">
          <div className="grid grid-cols-2 gap-4 mb-6 text-sm">
            <div>
              <p className="text-slate-500">Cliente</p>
              <p className="font-semibold text-slate-800">{pedido.cliente.nombre}</p>
            </div>
            <div>
              <p className="text-slate-500">Fecha</p>
              <p className="font-semibold text-slate-800">{pedido.fecha}</p>
            </div>
            <div>
              <p className="text-slate-500">Correo</p>
              <p className="font-semibold text-slate-800">{pedido.cliente.email}</p>
            </div>
            <div>
              <p className="text-slate-500">Ciudad</p>
              <p className="font-semibold text-slate-800">{pedido.cliente.ciudad}</p>
            </div>
          </div>

          <h2 className="font-display text-xl uppercase text-marino-700 mb-3">Detalle del pedido</h2>
          <div className="space-y-2 border-b border-linea pb-4 mb-4">
            {pedido.items.map((item, idx) => (
              <div key={idx} className="flex justify-between text-sm">
                <span className="text-slate-600">
                  {item.nombre} × {item.cantidad}
                </span>
                <span className="font-semibold">
                  ${(item.precio * item.cantidad).toFixed(2)}
                </span>
              </div>
            ))}
          </div>
          <div className="flex justify-between items-center mb-8">
            <span className="text-xs font-semibold uppercase tracking-widest text-slate-600">Total pagado</span>
            <span className="font-display text-4xl text-marino-700">
              ${pedido.total.toFixed(2)}
            </span>
          </div>

          <div className="flex gap-4">
            <Link
              to="/mis-pedidos"
              className="flex-1 text-center bg-turquesa-700 text-white font-semibold py-3 hover:bg-turquesa-800 transition"
            >
              Ver mis pedidos
            </Link>
            <Link
              to="/tienda"
              className="flex-1 text-center bg-marino-700 text-arena-200 font-semibold py-3 hover:bg-marino-800 transition"
            >
              Seguir Comprando
            </Link>
          </div>
          <div className="mt-4 text-center">
            <Link to="/catalogo" className="text-sm text-indigo-600 hover:underline">
              Ver Catálogo
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Confirmacion;