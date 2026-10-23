// src/components/MisComisiones.tsx
// Panel del ledger de comisiones del usuario autenticado (Fase 4).
// Muestra cada fila que le ha acreditado la red y los totales por estado.
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import Kpi from "./Kpi";
import {
  getMisComisiones,
  type Comision,
  type EstadoComision,
} from "../services/comisionesService";

// Colores del badge según el estado de la fila del ledger.
const COLOR_ESTADO: Record<EstadoComision, string> = {
  pendiente: "bg-arena-200 text-marino-800",
  aprobada: "bg-marino-100 text-marino-700",
  pagada: "bg-turquesa-100 text-turquesa-800",
  revertida: "bg-coral-100 text-coral-700",
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
    { etiqueta: "Pendiente de cobro", valor: pendiente, tono: "coral" as const },
    { etiqueta: "Ya pagado", valor: pagado, tono: "turquesa" as const },
    {
      etiqueta: "Total devengado",
      valor: pendiente + pagado,
      tono: "marino" as const,
    },
  ];

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div>
          <h1 className="text-3xl uppercase text-marino-700">Mis Comisiones</h1>
          <p className="text-slate-500 mt-1">
            Ledger de lo que tu red te ha acreditado por cada pedido.
          </p>
        </div>
        <Link
          to="/mi-red"
          className="bg-marino-700 text-arena-200 font-semibold px-4 py-2 hover:bg-marino-800 transition"
        >
          Ver mi red
        </Link>
      </div>

      {/* Totales por estado */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-6">
        {resumen.map((item) => (
          <Kpi
            key={item.etiqueta}
            etiqueta={item.etiqueta}
            valor={`$${item.valor.toFixed(2)}`}
            tono={item.tono}
          />
        ))}
      </div>

      {error && (
        <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-4 border border-coral-200 border-l-4 border-l-coral-500">
          {error}
        </div>
      )}

      <div className="bg-papel border border-linea overflow-hidden">
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
              className="inline-block bg-marino-700 text-arena-200 font-semibold px-5 py-2 hover:bg-marino-800 transition"
            >
              Ir a Mi Red
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-arena-100 text-marino-700 uppercase tracking-widest text-xs border-b border-marino-700">
                <tr>
                  <th className="text-left p-3">Pedido</th>
                  <th className="text-center p-3">Nivel</th>
                  <th className="text-center p-3">Porcentaje</th>
                  <th className="text-right p-3">Monto</th>
                  <th className="text-center p-3">Estado</th>
                  <th className="text-right p-3">Fecha</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-linea">
                {comisiones.map((cm) => (
                  <tr key={cm.id}>
                    <td className="p-3 font-mono text-slate-800">
                      {cm.pedido_numero}
                    </td>
                    <td className="p-3 text-center">
                      <span className="text-xs font-semibold uppercase tracking-wider px-2 py-1 bg-marino-100 text-marino-700">
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
                        className={`text-xs font-semibold uppercase tracking-wider px-2 py-1 ${COLOR_ESTADO[cm.estado]}`}
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
            className="px-3 py-1.5 border border-marino-700 text-marino-700 font-medium disabled:opacity-40 hover:bg-marino-700 hover:text-arena-200"
          >
            Anterior
          </button>
          <span className="text-slate-500">
            Página {pagina} de {paginas} · {total} comisiones
          </span>
          <button
            onClick={() => setPagina((p) => Math.min(paginas, p + 1))}
            disabled={pagina === paginas}
            className="px-3 py-1.5 border border-marino-700 text-marino-700 font-medium disabled:opacity-40 hover:bg-marino-700 hover:text-arena-200"
          >
            Siguiente
          </button>
        </div>
      )}
    </div>
  );
};

export default MisComisiones;
