// src/components/AdminLiquidaciones.tsx
// Cierre del ciclo del dinero (Fase 5): el admin ve cuánto debe cada
// distribuidor, liquida su saldo de una vez y consulta el histórico de pagos.
import { useEffect, useState } from "react";
import Kpi from "./Kpi";
import {
  getDeudas,
  getLiquidaciones,
  liquidarDistribuidor,
  type Deuda,
  type Liquidacion,
} from "../services/liquidacionesService";

const AdminLiquidaciones = () => {
  const [deudas, setDeudas] = useState<Deuda[]>([]);
  const [historico, setHistorico] = useState<Liquidacion[]>([]);
  const [pagina, setPagina] = useState<number>(1);
  const [totalPagos, setTotalPagos] = useState<number>(0);
  // `version` fuerza una recarga tras liquidar; la carga se deriva de la clave pedida
  // frente a la última cargada, sin setState síncrono en el efecto.
  const [version, setVersion] = useState<number>(0);
  const [claveCargada, setClaveCargada] = useState<string>("");
  const clave = `${pagina}|${version}`;
  const cargando = claveCargada !== clave;
  const [error, setError] = useState<string>("");
  const [liquiando, setLiquiando] = useState<number | null>(null);

  const LIMITE = 10;

  // Carga la deuda y el histórico en paralelo: son lecturas independientes
  useEffect(() => {
    Promise.all([getDeudas(), getLiquidaciones(pagina, LIMITE)])
      .then(([deudasRes, pagosRes]) => {
        setDeudas(deudasRes);
        setHistorico(pagosRes.data);
        setTotalPagos(pagosRes.total);
        setError("");
      })
      .catch((err) => setError(err.message || "No se pudo cargar el módulo"))
      .finally(() => setClaveCargada(`${pagina}|${version}`));
  }, [pagina, version]);

  // Liquida el saldo completo de un distribuidor con confirmación previa
  const liquidar = (d: Deuda) => {
    const texto = `¿Liquidar $${d.pendiente.toFixed(2)} a ${d.nombre} (${d.cantidad} comisiones)?`;
    if (!window.confirm(texto)) return;

    setLiquiando(d.referido_id);
    liquidarDistribuidor(d.referido_id)
      .then(() => {
        setPagina(1);
        setVersion((v) => v + 1);
      })
      .catch((err) => {
        window.alert(err.message || "No se pudo realizar la liquidación");
      })
      .finally(() => setLiquiando(null));
  };

  const paginas = Math.max(1, Math.ceil(totalPagos / LIMITE));
  const deudaTotal = deudas.reduce((suma, d) => suma + d.pendiente, 0);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl uppercase text-marino-700">Liquidaciones</h1>
        <p className="text-slate-500 mt-1">
          Pago de comisiones pendientes a los distribuidores de la red.
        </p>
      </div>

      {/* Resumen de la deuda total */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-6">
        <Kpi etiqueta="Deuda total" valor={`$${deudaTotal.toFixed(2)}`} tono="coral" />
        <Kpi etiqueta="Distribuidores con saldo" valor={String(deudas.length)} tono="marino" />
        <Kpi etiqueta="Pagos realizados" valor={String(totalPagos)} tono="turquesa" />
      </div>

      {error && (
        <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-4 border border-coral-200 border-l-4 border-l-coral-500">
          {error}
        </div>
      )}

      {/* ===== DEUDA POR DISTRIBUIDOR ===== */}
      <div className="bg-papel border border-linea overflow-hidden mb-8">
        <h2 className="text-lg font-bold text-slate-800 p-4 border-b border-linea">
          Deuda por distribuidor
        </h2>
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando deudas...</p>
        ) : deudas.length === 0 ? (
          <p className="p-6 text-slate-500">
            No hay comisiones pendientes: todo el ledger está al día.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full tabular-nums text-sm">
              <thead className="bg-arena-100 text-marino-700 uppercase tracking-widest text-xs border-b border-marino-700">
                <tr>
                  <th className="text-left p-3">Distribuidor</th>
                  <th className="text-left p-3">Correo</th>
                  <th className="text-center p-3">Comisiones</th>
                  <th className="text-right p-3">Pendiente</th>
                  <th className="text-right p-3">Acción</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-linea">
                {deudas.map((d) => (
                  <tr key={d.referido_id} className="hover:bg-arena-50">
                    <td className="p-3 font-medium text-slate-800">{d.nombre}</td>
                    <td className="p-3 text-slate-600">{d.email || "—"}</td>
                    <td className="p-3 text-center">{d.cantidad}</td>
                    <td className="p-3 text-right font-semibold text-coral-600">
                      ${d.pendiente.toFixed(2)}
                    </td>
                    <td className="p-3 text-right">
                      <button
                        onClick={() => liquidar(d)}
                        disabled={liquiando === d.referido_id}
                        className="bg-turquesa-700 text-white text-xs font-semibold px-3 py-1.5 hover:bg-turquesa-800 transition disabled:opacity-50"
                      >
                        {liquiando === d.referido_id
                          ? "Liquidando..."
                          : "Liquidar"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* ===== HISTÓRICO DE PAGOS ===== */}
      <div className="bg-papel border border-linea overflow-hidden">
        <h2 className="text-lg font-bold text-slate-800 p-4 border-b border-linea">
          Histórico de pagos
        </h2>
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando histórico...</p>
        ) : historico.length === 0 ? (
          <p className="p-6 text-slate-500">Todavía no se ha realizado ningún pago.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full tabular-nums text-sm">
              <thead className="bg-arena-100 text-marino-700 uppercase tracking-widest text-xs border-b border-marino-700">
                <tr>
                  <th className="text-left p-3">#</th>
                  <th className="text-left p-3">Distribuidor</th>
                  <th className="text-center p-3">Periodo</th>
                  <th className="text-center p-3">Comisiones</th>
                  <th className="text-right p-3">Total</th>
                  <th className="text-center p-3">Estado</th>
                  <th className="text-right p-3">Fecha</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-linea">
                {historico.map((l) => (
                  <tr key={l.id} className="hover:bg-arena-50">
                    <td className="p-3 text-slate-500">{l.id}</td>
                    <td className="p-3 font-medium text-slate-800">{l.nombre}</td>
                    <td className="p-3 text-center text-slate-600">{l.periodo}</td>
                    <td className="p-3 text-center">{l.cantidad_comisiones}</td>
                    <td className="p-3 text-right font-semibold text-slate-800">
                      ${l.total.toFixed(2)}
                    </td>
                    <td className="p-3 text-center">
                      <span
                        className={`text-xs font-semibold uppercase tracking-wider px-2 py-1 ${
                          l.estado === "pagada"
                            ? "bg-turquesa-100 text-turquesa-800"
                            : "bg-arena-200 text-marino-800"
                        }`}
                      >
                        {l.estado === "pagada" ? "Pagada" : "Pendiente"}
                      </span>
                    </td>
                    <td className="p-3 text-right text-slate-500">
                      {new Date(l.creado_en).toLocaleString("es-EC")}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Paginación del histórico */}
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
            Página {pagina} de {paginas} · {totalPagos} pagos
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

export default AdminLiquidaciones;
