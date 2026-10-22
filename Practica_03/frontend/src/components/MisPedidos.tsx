// src/components/MisPedidos.tsx
// Historial de pedidos del cliente autenticado (Fase 3).
// Cada fila se expande para mostrar las líneas del pedido (GET /api/pedidos/:id).
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  getMisPedidos,
  getPedidoDetalle,
  type EstadoPedido,
  type Pedido,
  type PedidoDetalle,
} from "../services/pedidosService";

// Colores asociados a cada estado del ciclo de vida.
const COLOR_ESTADO: Record<EstadoPedido, string> = {
  pendiente: "bg-amber-100 text-amber-700",
  preparando: "bg-blue-100 text-blue-700",
  enviado: "bg-indigo-100 text-indigo-700",
  entregado: "bg-green-100 text-green-700",
  cancelado: "bg-red-100 text-red-700",
};

// Texto legible de cada estado.
const TEXTO_ESTADO: Record<EstadoPedido, string> = {
  pendiente: "Pendiente",
  preparando: "Preparando",
  enviado: "Enviado",
  entregado: "Entregado",
  cancelado: "Cancelado",
};

// Etiqueta de estado reutilizable (también la usa la vista administrativa).
export const BadgeEstado = ({ estado }: { estado: EstadoPedido }) => (
  <span
    className={`text-xs font-semibold px-2 py-1 rounded-full ${COLOR_ESTADO[estado]}`}
  >
    {TEXTO_ESTADO[estado]}
  </span>
);

const MisPedidos = () => {
  const [pedidos, setPedidos] = useState<Pedido[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [pagina, setPagina] = useState<number>(1);
  // La carga se deriva de la clave pedida frente a la última cargada: sin setState síncrono en el efecto.
  const [claveCargada, setClaveCargada] = useState<string>("");
  const cargando = claveCargada !== String(pagina);
  const [error, setError] = useState<string>("");

  // Detalle expandible: id del pedido abierto y su información cargada
  const [abierto, setAbierto] = useState<number | null>(null);
  const [detalle, setDetalle] = useState<PedidoDetalle | null>(null);
  const [cargandoDetalle, setCargandoDetalle] = useState<boolean>(false);

  const LIMITE = 10;

  useEffect(() => {
    getMisPedidos(pagina, LIMITE)
      .then((res) => {
        setPedidos(res.data);
        setTotal(res.total);
        setError("");
      })
      .catch((err) => setError(err.message || "No se pudo cargar tus pedidos"))
      .finally(() => setClaveCargada(String(pagina)));
  }, [pagina]);

  // Alterna el detalle expandible de un pedido (lo carga solo la primera vez)
  const alternarDetalle = (p: Pedido) => {
    if (abierto === p.id) {
      setAbierto(null);
      setDetalle(null);
      return;
    }
    setAbierto(p.id);
    setDetalle(null);
    setCargandoDetalle(true);
    getPedidoDetalle(p.id)
      .then(setDetalle)
      .catch((err) => setError(err.message || "No se pudo cargar el detalle"))
      .finally(() => setCargandoDetalle(false));
  };

  const paginas = Math.max(1, Math.ceil(total / LIMITE));

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Mis Pedidos</h1>
          <p className="text-slate-500 mt-1">
            Historial de tus compras con su estado actual.
          </p>
        </div>
        <Link
          to="/tienda"
          className="bg-indigo-600 text-white font-semibold px-4 py-2 rounded-lg hover:bg-indigo-700 transition"
        >
          Seguir comprando
        </Link>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando pedidos...</p>
        ) : pedidos.length === 0 ? (
          <div className="p-8 text-center">
            <p className="text-slate-500 mb-4">Todavía no has realizado pedidos.</p>
            <Link
              to="/tienda"
              className="inline-block bg-indigo-600 text-white font-semibold px-5 py-2 rounded-lg hover:bg-indigo-700 transition"
            >
              Ir a la Tienda
            </Link>
          </div>
        ) : (
          <ul className="divide-y divide-slate-100">
            {pedidos.map((p) => (
              <li key={p.id}>
                <div className="p-4 flex flex-wrap items-center gap-3">
                  <div className="flex-1 min-w-[180px]">
                    <p className="font-mono font-semibold text-slate-800">{p.numero}</p>
                    <p className="text-sm text-slate-500">
                      {new Date(p.creado_en).toLocaleString("es-EC")}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-slate-800">${p.total.toFixed(2)}</p>
                    <p className="text-xs text-slate-500">{p.cliente_ciudad}</p>
                  </div>
                  <BadgeEstado estado={p.estado} />
                  <div className="flex gap-3 text-sm font-medium">
                    <button
                      onClick={() => alternarDetalle(p)}
                      className="text-indigo-600 hover:underline"
                    >
                      {abierto === p.id ? "Ocultar detalle" : "Ver detalle"}
                    </button>
                  </div>
                </div>

                {/* Detalle expandible con las líneas del pedido */}
                {abierto === p.id && (
                  <div className="bg-slate-50 p-4 border-t border-slate-100">
                    {cargandoDetalle ? (
                      <p className="text-sm text-slate-500">Cargando detalle...</p>
                    ) : detalle ? (
                      <div className="space-y-3">
                        <div className="grid sm:grid-cols-3 gap-3 text-sm">
                          <div>
                            <p className="text-slate-500">Cliente</p>
                            <p className="font-medium text-slate-800">{detalle.cliente_nombre}</p>
                          </div>
                          <div>
                            <p className="text-slate-500">Dirección</p>
                            <p className="font-medium text-slate-800">
                              {detalle.cliente_direccion}, {detalle.cliente_ciudad}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-500">Método de pago</p>
                            <p className="font-medium text-slate-800">{detalle.metodo_pago}</p>
                          </div>
                        </div>

                        <div className="overflow-x-auto">
                          <table className="w-full min-w-[420px] text-sm">
                            <thead className="text-slate-500 text-xs uppercase">
                              <tr>
                                <th className="text-left p-2">Producto</th>
                                <th className="text-right p-2">Cant.</th>
                                <th className="text-right p-2">Precio</th>
                                <th className="text-right p-2">Subtotal</th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-200">
                              {detalle.items.map((it) => (
                                <tr key={it.producto_id}>
                                  <td className="p-2 text-slate-700">{it.nombre}</td>
                                  <td className="p-2 text-right">{it.cantidad}</td>
                                  <td className="p-2 text-right">${it.precio_unitario.toFixed(2)}</td>
                                  <td className="p-2 text-right font-medium">
                                    ${it.subtotal.toFixed(2)}
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    ) : (
                      <p className="text-sm text-red-600">No se pudo cargar el detalle.</p>
                    )}
                  </div>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Paginación */}
      {paginas > 1 && (
        <div className="flex items-center justify-between mt-4 text-sm">
          <button
            onClick={() => setPagina((p) => Math.max(1, p - 1))}
            disabled={pagina === 1}
            className="px-3 py-1.5 rounded border border-slate-300 text-slate-700 disabled:opacity-40 hover:bg-slate-50"
          >
            Anterior
          </button>
          <span className="text-slate-500">
            Página {pagina} de {paginas} · {total} pedidos
          </span>
          <button
            onClick={() => setPagina((p) => Math.min(paginas, p + 1))}
            disabled={pagina === paginas}
            className="px-3 py-1.5 rounded border border-slate-300 text-slate-700 disabled:opacity-40 hover:bg-slate-50"
          >
            Siguiente
          </button>
        </div>
      )}
    </div>
  );
};

export default MisPedidos;
