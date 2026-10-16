// src/components/AdminPedidos.tsx
// Gestión del ciclo de vida de pedidos (Fase 3): filtro por estado,
// detalle expandible y cambio de estado con las transiciones válidas.
import { useEffect, useState } from "react";
import {
  getPedidosAdmin,
  getPedidoDetalle,
  cambiarEstadoPedido,
  type EstadoPedido,
  type Pedido,
  type PedidoDetalle,
} from "../services/pedidosService";
import { BadgeEstado } from "./MisPedidos";

// Transiciones permitidas por el backend (ciclo de vida del pedido).
const TRANSICIONES: Record<EstadoPedido, EstadoPedido[]> = {
  pendiente: ["preparando", "cancelado"],
  preparando: ["enviado", "cancelado"],
  enviado: ["entregado", "cancelado"],
  entregado: [],
  cancelado: [],
};

// Etiqueta legible de cada estado para el selector.
const ETIQUETAS: Record<EstadoPedido, string> = {
  pendiente: "Pendiente",
  preparando: "Preparando",
  enviado: "Enviado",
  entregado: "Entregado",
  cancelado: "Cancelado",
};

const FILTROS: Array<{ valor: string; texto: string }> = [
  { valor: "", texto: "Todos" },
  { valor: "pendiente", texto: "Pendientes" },
  { valor: "preparando", texto: "Preparando" },
  { valor: "enviado", texto: "Enviados" },
  { valor: "entregado", texto: "Entregados" },
  { valor: "cancelado", texto: "Cancelados" },
];

const AdminPedidos = () => {
  const [pedidos, setPedidos] = useState<Pedido[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [pagina, setPagina] = useState<number>(1);
  const [filtro, setFiltro] = useState<string>("");
  // La carga se deriva de la clave pedida frente a la última cargada: sin setState síncrono en el efecto.
  const [claveCargada, setClaveCargada] = useState<string>("");
  const cargando = claveCargada !== `${pagina}|${filtro}`;
  const [error, setError] = useState<string>("");

  const [abierto, setAbierto] = useState<number | null>(null);
  const [detalle, setDetalle] = useState<PedidoDetalle | null>(null);
  const [cargandoDetalle, setCargandoDetalle] = useState<boolean>(false);

  const LIMITE = 10;

  // Recarga al cambiar la página o el filtro de estado
  useEffect(() => {
    getPedidosAdmin(filtro, pagina, LIMITE)
      .then((res) => {
        setPedidos(res.data);
        setTotal(res.total);
        setError("");
      })
      .catch((err) => setError(err.message || "No se pudo cargar los pedidos"))
      .finally(() => setClaveCargada(`${pagina}|${filtro}`));
  }, [pagina, filtro]);

  const cambiarFiltro = (valor: string) => {
    setFiltro(valor);
    setPagina(1);
  };

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

  // Aplica una transición y refleja el cambio en la tabla sin recargar
  const transicionar = (p: Pedido, estado: EstadoPedido) => {
    const texto =
      estado === "cancelado"
        ? `¿Cancelar el pedido ${p.numero}? Se devolverá el stock y se revertirán las comisiones.`
        : `¿Cambiar el pedido ${p.numero} a "${ETIQUETAS[estado]}"?`;
    if (!window.confirm(texto)) return;

    cambiarEstadoPedido(p.id, estado)
      .then((actualizado) => {
        setPedidos((prev) =>
          prev.map((x) => (x.id === p.id ? { ...x, estado: actualizado.estado } : x))
        );
        setDetalle((prev) =>
          prev && prev.id === p.id ? { ...prev, estado: actualizado.estado } : prev
        );
      })
      .catch((err) => window.alert(err.message || "No se pudo cambiar el estado"));
  };

  const paginas = Math.max(1, Math.ceil(total / LIMITE));

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-slate-800">Administración de Pedidos</h1>
        <p className="text-slate-500 mt-1">
          Ciclo de vida: pendiente → preparando → enviado → entregado (o cancelado).
        </p>
      </div>

      {/* Filtro por estado */}
      <div className="flex flex-wrap gap-2 mb-4">
        {FILTROS.map((f) => (
          <button
            key={f.valor}
            onClick={() => cambiarFiltro(f.valor)}
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition ${
              filtro === f.valor
                ? "bg-indigo-600 text-white"
                : "bg-slate-100 text-slate-600 hover:bg-slate-200"
            }`}
          >
            {f.texto}
          </button>
        ))}
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
          <p className="p-6 text-slate-500">No hay pedidos con ese filtro.</p>
        ) : (
          <ul className="divide-y divide-slate-100">
            {pedidos.map((p) => (
              <li key={p.id}>
                <div className="p-4 flex flex-wrap items-center gap-3">
                  <div className="flex-1 min-w-[180px]">
                    <p className="font-mono font-semibold text-slate-800">{p.numero}</p>
                    <p className="text-sm text-slate-500">
                      {p.cliente_nombre} · {new Date(p.creado_en).toLocaleString("es-EC")}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-slate-800">${p.total.toFixed(2)}</p>
                    <p className="text-xs text-slate-500">{p.metodo_pago}</p>
                  </div>
                  <BadgeEstado estado={p.estado} />

                  {/* Selector de transición disponible */}
                  <div className="flex items-center gap-2">
                    {TRANSICIONES[p.estado].length > 0 ? (
                      <select
                        value=""
                        onChange={(e) => e.target.value && transicionar(p, e.target.value as EstadoPedido)}
                        className="px-3 py-1.5 rounded-lg border border-slate-300 text-sm bg-white focus:ring-2 focus:ring-indigo-600 outline-none"
                      >
                        <option value="">Cambiar estado...</option>
                        {TRANSICIONES[p.estado].map((destino) => (
                          <option key={destino} value={destino}>
                            → {ETIQUETAS[destino]}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <span className="text-xs text-slate-400">Estado final</span>
                    )}
                    <button
                      onClick={() => alternarDetalle(p)}
                      className="text-indigo-600 hover:underline text-sm font-medium"
                    >
                      {abierto === p.id ? "Ocultar" : "Detalle"}
                    </button>
                  </div>
                </div>

                {/* Detalle con líneas del pedido */}
                {abierto === p.id && (
                  <div className="bg-slate-50 p-4 border-t border-slate-100">
                    {cargandoDetalle ? (
                      <p className="text-sm text-slate-500">Cargando detalle...</p>
                    ) : detalle ? (
                      <div className="space-y-3">
                        <div className="grid sm:grid-cols-4 gap-3 text-sm">
                          <div>
                            <p className="text-slate-500">Cliente</p>
                            <p className="font-medium text-slate-800">{detalle.cliente_nombre}</p>
                          </div>
                          <div>
                            <p className="text-slate-500">Correo</p>
                            <p className="font-medium text-slate-800">{detalle.cliente_email}</p>
                          </div>
                          <div>
                            <p className="text-slate-500">Envío</p>
                            <p className="font-medium text-slate-800">
                              {detalle.cliente_direccion}, {detalle.cliente_ciudad}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-500">Teléfono</p>
                            <p className="font-medium text-slate-800">
                              {detalle.cliente_telefono || "—"}
                            </p>
                          </div>
                        </div>

                        <table className="w-full text-sm">
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

export default AdminPedidos;
