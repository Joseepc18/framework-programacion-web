// src/components/MiRed.tsx
// Red multinivel del usuario autenticado (Fase 4):
// - Resumen: referidos activos, ventas de la red, comisiones del mes y nivel
// - Árbol jerárquico recursivo de SU subárbol (no el árbol global de la semilla)
// - Link de invitación /registro?ref=CODIGO con botón para copiarlo
import { useState, useEffect } from 'react';
import { contarRed, sumarVentasRed, sumarComisiones, nivelAlcanzado, comisionDeReferido, TASA_COMISION, type Referido } from '../data/red';
import { getRedMia } from '../services/redService';

import Kpi from "./Kpi";
// Colores y etiquetas por nivel para el árbol
const NIVEL_ESTILO: Record<number, { badge: string; border: string; tasa: string }> = {
  1: { badge: "bg-marino-100 text-marino-700", border: "border-marino-200", tasa: "10 %" },
  2: { badge: "bg-turquesa-100 text-turquesa-800", border: "border-turquesa-200", tasa: "5 %" },
  3: { badge: "bg-pink-100 text-pink-700", border: "border-pink-200", tasa: "2 %" },
};

// Componente recursivo: renderiza un referido y, debajo, a sus hijos
const NodoReferido = ({ nodo }: { nodo: Referido }) => {
  const estilo = NIVEL_ESTILO[nodo.nivel] ?? NIVEL_ESTILO[1];
  const comision = comisionDeReferido(nodo);

  return (
    <li>
      <div className={`bg-papel border ${estilo.border} p-4 flex flex-wrap items-center justify-between gap-3`}>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-slate-200 flex items-center justify-center font-bold text-slate-600">
            {nodo.nombre.charAt(0)}
          </div>
          <div>
            <p className="font-semibold text-slate-800">{nodo.nombre}</p>
            <span className={`inline-block mt-1 text-xs font-semibold uppercase tracking-wider px-2 py-0.5 ${estilo.badge}`}>
              Nivel {nodo.nivel} · {estilo.tasa} comisión
            </span>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-500">Ventas</p>
          <p className="font-bold text-slate-800">${nodo.ventas.toLocaleString()}</p>
          <p className="text-xs text-indigo-600 font-semibold mt-1">
            +${comision.toFixed(2)} para ti
          </p>
        </div>
      </div>

      {/* Hijos del referido (más profundidad en el árbol) */}
      {nodo.hijos && nodo.hijos.length > 0 && (
        <ul className="mt-3 ml-6 space-y-3 border-l-2 border-linea pl-4">
          {nodo.hijos.map((hijo) => (
            <NodoReferido key={hijo.id} nodo={hijo} />
          ))}
        </ul>
      )}
    </li>
  );
};

const MiRed = () => {
  const [red, setRed] = useState<Referido | null>(null);
  const [codigo, setCodigo] = useState<string>("");
  const [cargando, setCargando] = useState<boolean>(true);
  const [error, setError] = useState<string>("");
  const [copiado, setCopiado] = useState<boolean>(false);

  useEffect(() => {
    getRedMia()
      .then((data) => {
        setRed(data.red);
        setCodigo(data.codigo);
        setError("");
      })
      .catch((err) => {
        setError(err.message || "No se pudo cargar tu red de referidos");
      })
      .finally(() => {
        setCargando(false);
      });
  }, []);

  // Link de invitación absoluto: se puede compartir por WhatsApp o correo
  const linkInvitacion = codigo
    ? `${window.location.origin}/registro?ref=${codigo}`
    : "";

  // Copia el link al portapapeles y confirma visualmente la acción
  const copiarLink = async () => {
    try {
      await navigator.clipboard.writeText(linkInvitacion);
      setCopiado(true);
      window.setTimeout(() => setCopiado(false), 2000);
    } catch {
      window.prompt("Copia el link de invitación:", linkInvitacion);
    }
  };

  // Los KPIs y el árbol se calculan SOLO con la red real del token:
  // nunca mezclamos con los datos de ejemplo de la semilla.
  const referidosActivos = red ? contarRed(red) : 0;
  const ventasRed = red ? sumarVentasRed(red) : 0;
  const comisionesMes = red ? sumarComisiones(red) : 0;
  const nivel = red ? nivelAlcanzado(red.hijos?.length ?? 0) : "—";

  const resumen = [
    { etiqueta: "Referidos Activos", valor: referidosActivos.toString(), tono: "marino" as const },
    { etiqueta: "Ventas de la Red", valor: `$${ventasRed.toLocaleString()}`, tono: "turquesa" as const },
    { etiqueta: "Comisiones del Mes", valor: `$${comisionesMes.toFixed(2)}`, tono: "arena" as const },
    { etiqueta: "Nivel Alcanzado", valor: nivel, tono: "coral" as const },
  ];

  return (
    <div>
      <h1 className="text-3xl uppercase text-marino-700 mb-2">Mi Red de Referidos</h1>
      <p className="text-slate-500 mb-6">
        Estructura multinivel: gana el {TASA_COMISION[1] * 100} % de nivel 1, {TASA_COMISION[2] * 100} % de nivel 2 y {TASA_COMISION[3] * 100} % de nivel 3.
      </p>

      {error && (
        <div className="bg-coral-50 text-coral-700 p-3 text-sm mb-6 border border-coral-200 border-l-4 border-l-coral-500">
          {error}
        </div>
      )}

      {/* ===== LINK DE INVITACIÓN ===== */}
      <div className="bg-papel border border-linea p-6 mb-8">
        <h2 className="text-lg font-bold text-slate-800 mb-1">Invita a tu red</h2>
        <p className="text-sm text-slate-500 mb-4">
          Comparte tu enlace: quien se registre con él quedará vinculado a tu nodo.
        </p>
        {linkInvitacion ? (
          <div className="flex flex-col sm:flex-row gap-3">
            <input
              type="text"
              readOnly
              value={linkInvitacion}
              onFocus={(e) => e.currentTarget.select()}
              className="flex-1 px-3 py-2 border border-slate-300 bg-slate-50 text-sm text-slate-700 font-mono"
              aria-label="Link de invitación"
            />
            <button
              onClick={copiarLink}
              className={`px-4 py-2 font-semibold text-white transition ${
                copiado ? "bg-turquesa-700" : "bg-marino-700 hover:bg-marino-800"
              }`}
            >
              {copiado ? "¡Copiado!" : "Copiar link"}
            </button>
          </div>
        ) : (
          <p className="text-sm text-slate-500">
            Tu cuenta no tiene código de invitación asignado.
          </p>
        )}
      </div>

      {/* ===== TARJETAS DE RESUMEN ===== */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {resumen.map((item) => (
          <Kpi key={item.etiqueta} etiqueta={item.etiqueta} valor={item.valor} tono={item.tono} />
        ))}
      </div>

      {/* ===== ÁRBOL DE REFERIDOS ===== */}
      <div className="bg-papel border border-linea p-6">
        <h2 className="text-lg font-bold text-slate-800 mb-4">
          Jerarquía de tu Red {cargando && <span className="text-xs text-slate-400 font-normal">(actualizando...)</span>}
        </h2>

        {/* Raíz: el usuario (solo cuando su red real ya cargó) */}
        {red ? (
          <>
            <div className="bg-marino-700 text-arena-200 p-4 mb-4 flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-arena-200 text-marino-800 flex items-center justify-center font-bold">
                  {red.nombre.charAt(0)}
                </div>
                <div>
                  <p className="font-semibold">{red.nombre}</p>
                  <p className="text-xs text-slate-300">Nivel alcanzado: {nivel}</p>
                </div>
              </div>
              <p className="font-bold text-turquesa-700">
                Comisiones del mes: ${comisionesMes.toFixed(2)}
              </p>
            </div>

            {(red.hijos?.length ?? 0) === 0 ? (
              <p className="text-sm text-slate-500">
                Todavía no tienes referidos. Comparte tu link de invitación para empezar.
              </p>
            ) : (
              <ul className="space-y-3">
                {red.hijos?.map((hijo) => (
                  <NodoReferido key={hijo.id} nodo={hijo} />
                ))}
              </ul>
            )}
          </>
        ) : cargando ? (
          <p className="text-sm text-slate-500">Cargando tu red...</p>
        ) : (
          <p className="text-sm text-slate-500">
            Tu cuenta aún no tiene un nodo de red asociado.
          </p>
        )}
      </div>
    </div>
  );
};

export default MiRed;