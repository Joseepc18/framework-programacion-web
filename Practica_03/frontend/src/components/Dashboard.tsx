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
import Kpi from "./Kpi";
import Tarjeta from "./Tarjeta";

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

  const kpis: { etiqueta: string; valor: string; tono: "marino" | "turquesa" | "coral" | "arena" }[] = [
    { etiqueta: "Total Vendido", valor: dinero(resumen?.total_vendido ?? 0), tono: "marino" },
    { etiqueta: "Pedidos", valor: String(resumen?.num_pedidos ?? 0), tono: "turquesa" },
    { etiqueta: "Ticket Promedio", valor: dinero(resumen?.ticket_promedio ?? 0), tono: "arena" },
    { etiqueta: "Comisiones Devengadas", valor: dinero(resumen?.comisiones_devengadas ?? 0), tono: "coral" },
  ];

  const cambiarPeriodo = () => {
    setCargando(true);
  };

  return (
    <div>
      <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4 mb-6">
        <h1 className="text-3xl uppercase text-marino-700">Resumen General</h1>
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
              className="border border-linea bg-papel px-3 py-2 min-h-11 text-tinta"
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
              className="border border-linea bg-papel px-3 py-2 min-h-11 text-tinta"
            />
          </label>
          {(desde || hasta) && (
            <button
              onClick={() => {
                cambiarPeriodo();
                setDesde("");
                setHasta("");
              }}
              className="px-3 py-2 min-h-11 border border-marino-700 text-marino-700 hover:bg-marino-700 hover:text-arena-200"
            >
              Todo el historial
            </button>
          )}
        </div>
      </div>

      {error && (
        <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-4 border border-coral-200 border-l-4 border-l-coral-500">
          {error}
        </div>
      )}

      {/* ===== KPIs DEL NEGOCIO ===== */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-5 mb-8">
        {kpis.map((kpi) => (
          <Kpi key={kpi.etiqueta} etiqueta={kpi.etiqueta} valor={cargando ? "…" : kpi.valor} tono={kpi.tono} />
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        {/* ===== TOP PRODUCTOS ===== */}
        <Tarjeta titulo="Top Productos" sinRelleno>
          {cargando ? (
            <p className="p-5 text-slate-500">Cargando...</p>
          ) : productos.length === 0 ? (
            <p className="p-5 text-slate-500">Sin ventas en el periodo seleccionado.</p>
          ) : (
            <ol className="divide-y divide-linea">
              {productos.map((p, idx) => (
                <li key={p.producto_id} className="flex items-center justify-between gap-4 px-5 py-3 hover:bg-arena-50">
                  <div className="flex items-center gap-4 min-w-0">
                    <span className="w-8 shrink-0 font-display text-3xl leading-none text-coral-500">
                      {idx + 1}
                    </span>
                    <div className="min-w-0">
                      <p className="font-semibold text-tinta truncate">{p.nombre}</p>
                      <p className="text-xs text-slate-500">{p.unidades} unidades</p>
                    </div>
                  </div>
                  <p className="font-semibold text-marino-700 tabular-nums">{dinero(p.ingresos)}</p>
                </li>
              ))}
            </ol>
          )}
        </Tarjeta>

        {/* ===== TOP REFERIDOS ===== */}
        <Tarjeta titulo="Top Distribuidores" sinRelleno>
          {cargando ? (
            <p className="p-5 text-slate-500">Cargando...</p>
          ) : referidos.length === 0 ? (
            <p className="p-5 text-slate-500">Sin ventas de red en el periodo seleccionado.</p>
          ) : (
            <ol className="divide-y divide-linea">
              {referidos.map((r, idx) => (
                <li key={r.referido_id} className="flex items-center justify-between gap-4 px-5 py-3 hover:bg-arena-50">
                  <div className="flex items-center gap-4 min-w-0">
                    <span className="w-8 shrink-0 font-display text-3xl leading-none text-coral-500">
                      {idx + 1}
                    </span>
                    <p className="font-semibold text-tinta truncate">{r.nombre}</p>
                  </div>
                  <div className="text-right tabular-nums">
                    <p className="font-semibold text-marino-700">{dinero(r.ventas_red)}</p>
                    <p className="text-xs font-semibold text-turquesa-700">+{dinero(r.comisiones)}</p>
                  </div>
                </li>
              ))}
            </ol>
          )}
        </Tarjeta>

        {/* ===== COMISIONES PENDIENTES ===== */}
        <Tarjeta
          titulo="Comisiones Pendientes"
          derecha={dinero(deudaTotal)}
          sinRelleno
          className="lg:col-span-2"
        >
          {cargando ? (
            <p className="p-5 text-slate-500">Cargando...</p>
          ) : deudas.length === 0 ? (
            <p className="p-5 text-slate-500">No hay comisiones pendientes de pago.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead className="border-b border-marino-700 bg-arena-100 text-xs uppercase tracking-widest text-marino-700">
                  <tr>
                    <th className="text-left px-5 py-3">Distribuidor</th>
                    <th className="text-center px-5 py-3">Comisiones</th>
                    <th className="text-right px-5 py-3">Pendiente</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-linea">
                  {deudas.map((d) => (
                    <tr key={d.referido_id} className="hover:bg-arena-50">
                      <td className="px-5 py-3 font-medium text-tinta">{d.nombre}</td>
                      <td className="px-5 py-3 text-center tabular-nums">{d.cantidad}</td>
                      <td className="px-5 py-3 text-right font-semibold tabular-nums text-coral-600">
                        {dinero(d.pendiente)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Tarjeta>
      </div>
    </div>
  );
};

export default Dashboard;
