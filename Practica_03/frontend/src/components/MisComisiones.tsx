// src/components/MisComisiones.tsx
// Panel del ledger de comisiones del usuario autenticado (Fase 4).
// Muestra cada fila que le ha acreditado la red y los totales por estado.
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import {
  getMisComisiones,
  type Comision,
  type EstadoComision,
} from "../services/comisionesService";

// Colores del badge según el estado de la fila del ledger.
const COLOR_ESTADO: Record<EstadoComision, string> = {
  pendiente: "bg-amber-100 text-amber-700",
  aprobada: "bg-blue-100 text-blue-700",
  pagada: "bg-green-100 text-green-700",
  revertida: "bg-red-100 text-red-700",
};

// Texto legible de cada estado.
const TEXTO_ESTADO: Record<EstadoComision, string> = {
  pendiente: "Pendiente",
  aprobada: "Aprobada",
  pagada: "Pagada",
  revertida: "Revertida",
};

const MisComisiones = () => {
  const [comisiones, setComisiones] = useState<Comision[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [pendiente, setPendiente] = useState<number>(0);
  const [pagado, setPagado] = useState<number>(0);
  const [pagina, setPagina] = useState<number>(1);
  // La carga se deriva de la clave pedida frente a la última cargada: sin setState síncrono en el efecto.
  const [claveCargada, setClaveCargada] = useState<string>("");
  const cargando = claveCargada !== String(pagina);
  const [error, setError] = useState<string>("");

  const LIMITE = 10;

  useEffect(() => {
    getMisComisiones(pagina, LIMITE)
      .then((res) => {
        setComisiones(res.data);
        setTotal(res.total);
        setPendiente(res.pendiente);
        setPagado(res.pagado);
        setError("");
      })
      .catch((err) => setError(err.message || "No se pudo cargar tu ledger"))
      .finally(() => setClaveCargada(String(pagina)));
  }, [pagina]);

  const paginas = Math.max(1, Math.ceil(total / LIMITE));

  const resumen = [
    { etiqueta: "Pendiente de cobro", valor: pendiente, color: "text-amber-600" },
    { etiqueta: "Ya pagado", valor: pagado, color: "text-green-600" },
    {
      etiqueta: "Total devengado",
      valor: pendiente + pagado,
      color: "text-indigo-600",
    },
  ];

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Mis Comisiones</h1>
          <p className="text-slate-500 mt-1">
            Ledger de lo que tu red te ha acreditado por cada pedido.
          </p>
        </div>
        <Link
          to="/mi-red"
          className="bg-indigo-600 text-white font-semibold px-4 py-2 rounded-lg hover:bg-indigo-700 transition"
        >
          Ver mi red
        </Link>
      </div>

      {/* Totales por estado */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-6">
        {resumen.map((item) => (
          <div
            key={item.etiqueta}
            className="bg-white p-6 rounded-lg shadow-sm border border-slate-100"
          >
            <p className="text-sm text-slate-500 uppercase font-semibold">
              {item.etiqueta}
            </p>
            <p className={`text-3xl font-bold mt-2 ${item.color}`}>
              ${item.valor.toFixed(2)}
            </p>
          </div>
        ))}
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando comisiones...</p>
        ) : comisiones.length === 0 ? (
          <div className="p-8 text-center">
            <p className="text-slate-500 mb-4">
              Aún no has generado comisiones. Invita gente a tu red para que
              aparezcan aquí.
            </p>
            <Link
              to="/mi-red"
              className="inline-block bg-indigo-600 text-white font-semibold px-5 py-2 rounded-lg hover:bg-indigo-700 transition"
            >
              Ir a Mi Red
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-slate-500 uppercase text-xs">
                <tr>
                  <th className="text-left p-3">Pedido</th>
                  <th className="text-center p-3">Nivel</th>
                  <th className="text-center p-3">Porcentaje</th>
                  <th className="text-right p-3">Monto</th>
                  <th className="text-center p-3">Estado</th>
                  <th className="text-right p-3">Fecha</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {comisiones.map((cm) => (
                  <tr key={cm.id}>
                    <td className="p-3 font-mono text-slate-800">
                      {cm.pedido_numero}
                    </td>
                    <td className="p-3 text-center">
                      <span className="text-xs font-semibold px-2 py-1 rounded-full bg-indigo-100 text-indigo-700">
                        Nivel {cm.nivel}
                      </span>
                    </td>
                    <td className="p-3 text-center text-slate-600">
                      {(cm.porcentaje * 100).toFixed(0)} %
                    </td>
                    <td className="p-3 text-right font-semibold text-slate-800">
                      ${cm.monto.toFixed(2)}
                    </td>
                    <td className="p-3 text-center">
                      <span
                        className={`text-xs font-semibold px-2 py-1 rounded-full ${COLOR_ESTADO[cm.estado]}`}
                      >
                        {TEXTO_ESTADO[cm.estado]}
                      </span>
                    </td>
                    <td className="p-3 text-right text-slate-500">
                      {new Date(cm.creado_en).toLocaleDateString("es-EC")}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
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
            Página {pagina} de {paginas} · {total} comisiones
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

export default MisComisiones;
