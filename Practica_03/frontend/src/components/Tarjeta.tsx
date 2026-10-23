// Bloque de contenido del panel: franja azul marino con el título y cuerpo sobre papel.
import type { ReactNode } from "react";

interface TarjetaProps {
  titulo: string;
  // Contenido alineado a la derecha de la franja (un total, un enlace…).
  derecha?: ReactNode;
  // Quita el relleno del cuerpo, útil cuando contiene una tabla a todo el ancho.
  sinRelleno?: boolean;
  className?: string;
  children: ReactNode;
}

const Tarjeta = ({ titulo, derecha, sinRelleno = false, className = "", children }: TarjetaProps) => (
  <section className={`border border-linea bg-papel ${className}`}>
    <header className="flex items-center justify-between gap-4 bg-marino-700 px-5 py-3 text-arena-200">
      <h2 className="font-display text-lg uppercase tracking-wider">{titulo}</h2>
      {derecha && <div className="font-display text-lg tabular-nums text-coral-400">{derecha}</div>}
    </header>
    <div className={sinRelleno ? "" : "p-5"}>{children}</div>
  </section>
);

export default Tarjeta;
