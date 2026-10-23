// Indicador numérico del panel: banda de color arriba, etiqueta pequeña y cifra grande.
type TonoKpi = "marino" | "turquesa" | "coral" | "arena";

interface KpiProps {
  etiqueta: string;
  valor: string;
  // Línea opcional bajo la cifra (por ejemplo, el periodo o una aclaración).
  nota?: string;
  tono?: TonoKpi;
}

const BANDA: Record<TonoKpi, string> = {
  marino: "border-t-marino-700",
  turquesa: "border-t-turquesa-500",
  coral: "border-t-coral-500",
  arena: "border-t-arena-400",
};

const Kpi = ({ etiqueta, valor, nota, tono = "marino" }: KpiProps) => (
  <div className={`border border-linea border-t-[6px] bg-papel p-5 ${BANDA[tono]}`}>
    <p className="text-xs font-semibold uppercase tracking-widest text-slate-600">{etiqueta}</p>
    <p className="mt-3 font-display text-5xl font-semibold leading-none text-marino-700 tabular-nums">
      {valor}
    </p>
    {nota && <p className="mt-3 text-sm text-slate-500">{nota}</p>}
  </div>
);

export default Kpi;
