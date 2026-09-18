// src/components/Carrito.tsx
import { useCart } from '../context/CartContext';

const Carrito = () => {
  const { cart, removeFromCart, totalPrice } = useCart();

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Tu Carrito de Compras</h1>
      
      {cart.length === 0 ? (
        <div className="bg-white p-8 rounded-lg border border-slate-200 shadow-sm text-center">
          <p className="text-slate-500">Tu carrito está vacío actualmente.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Lista de productos */}
          <div className="lg:col-span-2 space-y-4">
            {cart.map((item) => (
              <div key={item.id} className="flex items-center justify-between bg-white p-4 rounded-lg border border-slate-200 shadow-sm">
                <div className="flex items-center gap-4">
                  <img src={item.img} alt={item.nombre} className="w-16 h-16 object-cover rounded-md" />
                  <div>
                    <h3 className="font-semibold text-slate-800">{item.nombre}</h3>
                    <p className="text-sm text-slate-500">Cantidad: {item.cantidad}</p>
                  </div>
                </div>
                <div className="flex items-center gap-6">
                  <p className="font-bold text-indigo-600">${(item.precio * item.cantidad).toFixed(2)}</p>
                  <button 
                    onClick={() => removeFromCart(item.id)}
                    className="text-red-500 hover:text-red-700 transition"
                    title="Eliminar producto"
                  >
                    🗑️
                  </button>
                </div>
              </div>
            ))}
          </div>

          {/* Resumen de pago */}
          <div className="bg-white p-6 rounded-lg border border-slate-200 shadow-sm h-fit">
            <h2 className="text-lg font-bold text-slate-800 mb-4">Resumen del Pedido</h2>
            <div className="flex justify-between border-b border-slate-100 pb-4 mb-4">
              <span className="text-slate-600">Subtotal</span>
              <span className="font-semibold">${totalPrice.toFixed(2)}</span>
            </div>
            <div className="flex justify-between items-center mb-6">
              <span className="text-slate-800 font-bold">Total a Pagar</span>
              <span className="text-2xl font-bold text-indigo-600">${totalPrice.toFixed(2)}</span>
            </div>
            <button className="w-full bg-indigo-600 text-white py-3 rounded-lg font-semibold hover:bg-indigo-700 transition">
              Proceder al Pago
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Carrito;