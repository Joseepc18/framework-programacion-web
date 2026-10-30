import { TriangleAlert } from 'lucide-react';

// Badge de advertencia de alta visibilidad (RF-02): siempre con texto, nunca solo color.
// Con `compacto` se usa dentro de listas y citas; sin él, en la cabecera del expediente.
const AlertasMedicas = ({ alertas, compacto = false }: { alertas: string[]; compacto?: boolean }) => {
  if (alertas.length === 0) return null;

  if (compacto) {
    return (
      <ul className="flex flex-wrap gap-1" aria-label="Alertas médicas">
        {alertas.map((alerta) => (
          <li key={alerta} className="inline-flex items-center gap-1 rounded-full border border-chip-alerta-borde bg-chip-alerta-fondo px-2 py-0.5 text-etiqueta font-medium text-chip-alerta-texto">
            <TriangleAlert className="size-3.5 shrink-0" aria-hidden="true" />
            {alerta}
          </li>
        ))}
      </ul>
    );
  }

  return (
    <div role="alert" className="flex flex-wrap items-center gap-x-3 gap-y-2 rounded-control border-2 border-chip-alerta-borde bg-chip-alerta-fondo px-4 py-3 text-chip-alerta-texto">
      <span className="inline-flex items-center gap-2 text-seccion font-semibold">
        <TriangleAlert className="size-6 shrink-0" aria-hidden="true" />
        Alerta médica
      </span>
      <ul className="flex flex-wrap gap-2">
        {alertas.map((alerta) => (
          <li key={alerta} className="rounded-full border border-chip-alerta-borde px-3 py-1 text-cuerpo font-semibold">{alerta}</li>
        ))}
      </ul>
    </div>
  );
};

export default AlertasMedicas;
