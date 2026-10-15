// src/components/Dashboard.tsx
// Resumen general con KPIs calculados desde los datos reales de la red:
// - La red se consulta al backend (GET /api/red) y se resume con funciones puras de data/red.ts
import { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { getRed } from '../services/redService';
import {
  contarRed,
  sumarVentasRed,
  sumarComisiones,
  nivelAlcanzado,
  comisionDeReferido,
  type Referido,
} from '../data/red';

const Dashboard = () => {
  const { user } = useAuth();
  const [red, setRed] = useState<Referido | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Pedimos la red al backend: cada usuario tiene su propia clave de caché (red:tree:{email}).
  useEffect(() => {
    let activo = true;

    getRed(user?.email)
      .then((datos) => {
        if (activo) setRed(datos);
      })
      .catch((e: Error) => {
        if (activo) setError(e.message);
      });

    return () => {
      activo = false;
    };
  }, [user?.email]);

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 p-6 rounded-lg">
        {error}
      </div>
    );
  }

  if (!red) {
    return <p className="text-slate-500">Cargando red multinivel…</p>;
  }

  const referidosDirectos = red.hijos?.length ?? 0;
  const referidosActivos = contarRed(red);
  const ventasRed = sumarVentasRed(red);
  const comisionesMes = sumarComisiones(red);
  const nivel = nivelAlcanzado(referidosDirectos);

  const kpis = [
    { etiqueta: "Ventas de la Red", valor: `$${ventasRed.toLocaleString()}`, color: "text-indigo-600" },
    { etiqueta: "Referidos Activos", valor: referidosActivos.toString(), color: "text-indigo-600" },
    { etiqueta: "Comisiones del Mes", valor: `$${comisionesMes.toFixed(2)}`, color: "text-green-600" },
    { etiqueta: "Nivel Actual", valor: nivel, color: "text-amber-600" },
  ];

  // Top 3 referidos por ventas (sin recursión: solo nivel 1)
  const topReferidos = [...(red.hijos ?? [])]
    .sort((a, b) => b.ventas - a.ventas)
    .slice(0, 3);

  return (
    <div>
      <h1 className="text-2xl font-bold text-slate-800 mb-6">Resumen General</h1>

      {/* ===== KPIS DERIVADOS DEL ESTADO ===== */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
        {kpis.map((kpi) => (
          <div key={kpi.etiqueta} className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
            <p className="text-sm text-slate-500 uppercase font-semibold">{kpi.etiqueta}</p>
            <p className={`text-3xl font-bold mt-2 ${kpi.color}`}>{kpi.valor}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* ===== TOP REFERIDOS ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Top Referidos del Mes</h2>
          <div className="space-y-3">
            {topReferidos.map((ref, idx) => (
              <div
                key={ref.id}
                className="flex items-center justify-between p-3 rounded-lg border border-slate-100 hover:bg-slate-50 transition"
              >
                <div className="flex items-center gap-3">
                  <span className="w-8 h-8 rounded-full bg-indigo-100 text-indigo-700 font-bold flex items-center justify-center">
                    {idx + 1}
                  </span>
                  <div>
                    <p className="font-semibold text-slate-800">{ref.nombre}</p>
                    <p className="text-xs text-slate-500">Nivel {ref.nivel}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="font-bold text-slate-800">${ref.ventas.toLocaleString()}</p>
                  <p className="text-xs text-indigo-600 font-semibold">
                    +${comisionDeReferido(ref).toFixed(2)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* ===== PROGRESO DE NIVEL ===== */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-slate-100">
          <h2 className="text-lg font-bold text-slate-800 mb-4">Progreso al Siguiente Nivel</h2>
          <p className="text-slate-500 text-sm mb-4">
            {referidosDirectos} referidos directos. El nivel se calcula con:
          </p>
          <ul className="space-y-2 text-sm text-slate-600 mb-6">
            <li className="flex justify-between"><span>Bronce</span><span>0 – 1 directos</span></li>
            <li className="flex justify-between"><span>Plata</span><span>2 – 3 directos</span></li>
            <li className="flex justify-between"><span>Oro</span><span>4 – 5 directos</span></li>
            <li className="flex justify-between"><span>Diamante</span><span>6+ directos</span></li>
          </ul>

          <div className="h-3 rounded-full bg-slate-100 overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 rounded-full transition-all"
              style={{ width: `${Math.min((referidosDirectos / 6) * 100, 100)}%` }}
            />
          </div>
          <p className="text-center text-sm text-slate-500 mt-2 font-semibold">
            Nivel actual: {nivel} ({referidosDirectos} / 6 directos para Diamante)
          </p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
