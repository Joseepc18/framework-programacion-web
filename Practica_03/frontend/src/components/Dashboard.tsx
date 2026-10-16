// src/components/Dashboard.tsx
// Dashboard del admin con datos reales del negocio (Fase 6):
// KPIs de ventas, ranking de productos y de distribuidores, y deuda de comisiones.
import { useEffect, useState } from "react";
import {
  getComisionesPendientes,
  getResumenVentas,
  getTopProductos,
  getTopReferidos,
  type RangoReporte,
  type ResumenVentas,
  type TopProducto,
  type TopReferido,
} from "../services/reportesService";
import type { Deuda } from "../services/liquidacionesService";

const dinero = (valor: number): string =>
  `$${valor.toLocaleString("es-EC", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

const Dashboard = () => {
  const [desde, setDesde] = useState<string>("");
  const [hasta, setHasta] = useState<string>("");
  const [resumen, setResumen] = useState<ResumenVentas | null>(null);
  const [productos, setProductos] = useState<TopProducto[]>([]);
  const [referidos, setReferidos] = useState<TopReferido[]>([]);
  const [deudas, setDeudas] = useState<Deuda[]>([]);
  const [cargando, setCargando] = useState<boolean>(true);
  const [error, setError] = useState<string>("");

  // Recarga los cuatro reportes en paralelo cada vez que cambia el periodo.
  useEffect(() => {
    let vigente = true;
    const rango: RangoReporte = { desde: desde || undefined, hasta: hasta || undefined };

    Promise.all([
      getResumenVentas(rango),
      getTopProductos(rango),
      getTopReferidos(rango),
      getComisionesPendientes(),
    ])
      .then(([res, prod, ref, deu]) => {
        if (!vigente) return;
        setResumen(res);
        setProductos(prod);
        setReferidos(ref);
        setDeudas(deu);
        setError("");
      })
      .catch((err) => {
        if (vigente) setError(err.message || "No se pudieron cargar los reportes");
      })
      .finally(() => {
        if (vigente) setCargando(false);
      });

    return () => {
      vigente = false;
    };
  }, [desde, hasta]);

  const deudaTotal = deudas.reduce((suma, d) => suma + d.pendiente, 0);

  const kpis = [
    { etiqueta: "Total Vendido", valor: dinero(resumen?.total_vendido ?? 0), color: "text-indigo-600" },
    { etiqueta: "Pedidos", valor: String(resumen?.num_pedidos ?? 0), color: "text-indigo-600" },
    { etiqueta: "Ticket Promedio", valor: dinero(resumen?.ticket_promedio ?? 0), color: "text-amber-600" },
    { etiqueta: "Comisiones Devengadas", valor: dinero(resumen?.comisiones_devengadas ?? 0), color: "text-green-600" },
  ];

  const cambiarPeriodo = () => {
    setCargando(true);
  };

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4 mb-6">
        <h1 className="text-2xl font-bold text-slate-800">Resumen General</h1>
        <div className="flex flex-wrap items-end gap-3 text-sm">
          <label className="flex flex-col text-slate-500">
            Desde
            <input
              type="date"
              value={desde}
              max={hasta || undefined}
              onChange={(e) => {
                cambiarPeriodo();
                setDesde(e.target.value);
              }}
              className="border border-slate-300 rounded-lg px-3 py-2 min-h-11 text-slate-800"
            />
          </label>
          <label className="flex flex-col text-slate-500">
            Hasta
            <input
              type="date"
              value={hasta}
              min={desde || undefined}
              onChange={(e) => {
                cambiarPeriodo();
                setHasta(e.target.value);
              }}
              className="border border-slate-300 rounded-lg px-3 py-2 min-h-11 text-slate-800"
            />
          </label>
          {(desde || hasta) && (
            <button
              onClick={() => {
                cambiarPeriodo();
                setDesde("");
                setHasta("");
              }}
              className="px-3 py-2 min-h-11 rounded-lg border border-slate-300 text-slate-700 hover:bg-slate-50"
            >
              Todo el historial
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">
          {error}
        </div>
      )}

      {/* ===== KPIs DEL NEGOCIO ===== */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
        {kpis.map((kpi) => (
          <div key={kpi.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <p className="text-sm text-slate-500 uppercase font-semibold">{kpi.etiqueta}</p>
            <p className={`text-3xl font-bold mt-2 ${kpi.color}`}>
              {cargando ? "…" : kpi.valor}
            </p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* ===== TOP PRODUCTOS ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Top Productos</h2>
          {cargando ? (
            <p className="text-slate-500">Cargando...</p>
          ) : productos.length === 0 ? (
            <p className="text-slate-500">Sin ventas en el periodo seleccionado.</p>
          ) : (
            <div className="space-y-3">
              {productos.map((p, idx) => (
                <div
                  key={p.producto_id}
                  className="flex items-center justify-between p-3 rounded-lg border border-slate-100"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <span className="w-8 h-8 shrink-0 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <div className="min-w-0">
                      <p className="font-semibold text-slate-800 truncate">{p.nombre}</p>
                      <p className="text-xs text-slate-500">{p.unidades} unidades</p>
                    </div>
                  </div>
                  <p className="font-bold text-slate-800">{dinero(p.ingresos)}</p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* ===== TOP REFERIDOS ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Top Distribuidores</h2>
          {cargando ? (
            <p className="text-slate-500">Cargando...</p>
          ) : referidos.length === 0 ? (
            <p className="text-slate-500">Sin ventas de red en el periodo seleccionado.</p>
          ) : (
            <div className="space-y-3">
              {referidos.map((r, idx) => (
                <div
                  key={r.referido_id}
                  className="flex items-center justify-between p-3 rounded-lg border border-slate-100"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <span className="w-8 h-8 shrink-0 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center">
                      {idx + 1}
                    </span>
                    <p className="font-semibold text-slate-800 truncate">{r.nombre}</p>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-slate-800">{dinero(r.ventas_red)}</p>
                    <p className="text-xs text-indigo-600 font-semibold">+{dinero(r.comisiones)}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* ===== COMISIONES PENDIENTES ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100 lg:col-span-2">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-bold text-slate-800">Comisiones Pendientes</h2>
            <p className="text-amber-600 font-bold">{dinero(deudaTotal)}</p>
          </div>
          {cargando ? (
            <p className="text-slate-500">Cargando...</p>
          ) : deudas.length === 0 ? (
            <p className="text-slate-500">No hay comisiones pendientes de pago.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-slate-500 uppercase text-xs">
                  <tr>
                    <th className="text-left p-3">Distribuidor</th>
                    <th className="text-center p-3">Comisiones</th>
                    <th className="text-right p-3">Pendiente</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {deudas.map((d) => (
                    <tr key={d.referido_id}>
                      <td className="p-3 font-medium text-slate-800">{d.nombre}</td>
                      <td className="p-3 text-center">{d.cantidad}</td>
                      <td className="p-3 text-right font-semibold text-amber-700">{dinero(d.pendiente)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
