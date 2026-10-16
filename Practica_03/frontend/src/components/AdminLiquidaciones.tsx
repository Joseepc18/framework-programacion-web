// src/components/AdminLiquidaciones.tsx
// Cierre del ciclo del dinero (Fase 5): el admin ve cuánto debe cada
// distribuidor, liquida su saldo de una vez y consulta el histórico de pagos.
import { useEffect, useState } from "react";
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
  const [cargando, setCargando] = useState<boolean>(true);
  const [error, setError] = useState<string>("");
  const [liquiando, setLiquiando] = useState<number | null>(null);

  const LIMITE = 10;

  // Carga la deuda y el histórico en paralelo: son lecturas independientes
  const cargar = (pag: number) => {
    setCargando(true);
    Promise.all([getDeudas(), getLiquidaciones(pag, LIMITE)])
      .then(([deudasRes, pagosRes]) => {
        setDeudas(deudasRes);
        setHistorico(pagosRes.data);
        setTotalPagos(pagosRes.total);
        setError("");
      })
      .catch((err) => setError(err.message || "No se pudo cargar el módulo"))
      .finally(() => setCargando(false));
  };

  useEffect(() => {
    cargar(pagina);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pagina]);

  // Liquida el saldo completo de un distribuidor con confirmación previa
  const liquidar = (d: Deuda) => {
    const texto = `¿Liquidar $${d.pendiente.toFixed(2)} a ${d.nombre} (${d.cantidad} comisiones)?`;
    if (!window.confirm(texto)) return;

    setLiquiando(d.referido_id);
    liquidarDistribuidor(d.referido_id)
      .then(() => {
        setPagina(1);
        cargar(1);
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
        <h1 className="text-2xl font-bold text-slate-800">Liquidaciones</h1>
        <p className="text-slate-500 mt-1">
          Pago de comisiones pendientes a los distribuidores de la red.
        </p>
      </div>

      {/* Resumen de la deuda total */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-6">
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <p className="text-sm text-slate-500 uppercase font-semibold">
            Deuda total
          </p>
          <p className="text-3xl font-bold mt-2 text-amber-600">
            ${deudaTotal.toFixed(2)}
          </p>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <p className="text-sm text-slate-500 uppercase font-semibold">
            Distribuidores con saldo
          </p>
          <p className="text-3xl font-bold mt-2 text-indigo-600">{deudas.length}</p>
        </div>
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <p className="text-sm text-slate-500 uppercase font-semibold">
            Pagos realizados
          </p>
          <p className="text-3xl font-bold mt-2 text-green-600">{totalPagos}</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm mb-4 border border-red-200">
          {error}
        </div>
      )}

      {/* ===== DEUDA POR DISTRIBUIDOR ===== */}
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden mb-8">
        <h2 className="text-lg font-bold text-slate-800 p-4 border-b border-slate-100">
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
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-slate-500 uppercase text-xs">
                <tr>
                  <th className="text-left p-3">Distribuidor</th>
                  <th className="text-left p-3">Correo</th>
                  <th className="text-center p-3">Comisiones</th>
                  <th className="text-right p-3">Pendiente</th>
                  <th className="text-right p-3">Acción</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {deudas.map((d) => (
                  <tr key={d.referido_id}>
                    <td className="p-3 font-medium text-slate-800">{d.nombre}</td>
                    <td className="p-3 text-slate-600">{d.email || "—"}</td>
                    <td className="p-3 text-center">{d.cantidad}</td>
                    <td className="p-3 text-right font-semibold text-amber-700">
                      ${d.pendiente.toFixed(2)}
                    </td>
                    <td className="p-3 text-right">
                      <button
                        onClick={() => liquidar(d)}
                        disabled={liquiando === d.referido_id}
                        className="bg-green-600 text-white text-xs font-semibold px-3 py-1.5 rounded-lg hover:bg-green-700 transition disabled:opacity-50"
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
      <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
        <h2 className="text-lg font-bold text-slate-800 p-4 border-b border-slate-100">
          Histórico de pagos
        </h2>
        {cargando ? (
          <p className="p-6 text-slate-500">Cargando histórico...</p>
        ) : historico.length === 0 ? (
          <p className="p-6 text-slate-500">Todavía no se ha realizado ningún pago.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-slate-500 uppercase text-xs">
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
              <tbody className="divide-y divide-slate-100">
                {historico.map((l) => (
                  <tr key={l.id}>
                    <td className="p-3 text-slate-500">{l.id}</td>
                    <td className="p-3 font-medium text-slate-800">{l.nombre}</td>
                    <td className="p-3 text-center text-slate-600">{l.periodo}</td>
                    <td className="p-3 text-center">{l.cantidad_comisiones}</td>
                    <td className="p-3 text-right font-semibold text-slate-800">
                      ${l.total.toFixed(2)}
                    </td>
                    <td className="p-3 text-center">
                      <span
                        className={`text-xs font-semibold px-2 py-1 rounded-full ${
                          l.estado === "pagada"
                            ? "bg-green-100 text-green-700"
                            : "bg-amber-100 text-amber-700"
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
            className="px-3 py-1.5 rounded border border-slate-300 text-slate-700 disabled:opacity-40 hover:bg-slate-50"
          >
            Anterior
          </button>
          <span className="text-slate-500">
            Página {pagina} de {paginas} · {totalPagos} pagos
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

export default AdminLiquidaciones;
