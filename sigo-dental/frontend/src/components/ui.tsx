// Piezas de interfaz compartidas por los módulos. Solo usan los tokens de Aurora (docs/diseno.md).
import { Link, type LinkProps } from 'react-router-dom';
import type { ButtonHTMLAttributes, InputHTMLAttributes, ReactNode, SelectHTMLAttributes, TextareaHTMLAttributes } from 'react';

const ENFOQUE = 'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento';

export const CLASE_CAMPO = `mt-2 min-h-11 w-full rounded-control border border-chip-neutro-borde bg-chip-neutro-fondo px-4 text-cuerpo text-texto ${ENFOQUE}`;

interface CampoBase {
  id: string;
  etiqueta: string;
  ayuda?: string;
}

const Etiqueta = ({ id, etiqueta, ayuda }: CampoBase) => (
  <>
    <label htmlFor={id} className="text-etiqueta font-medium text-texto-suave">{etiqueta}</label>
    {ayuda && <p className="text-etiqueta text-texto-suave">{ayuda}</p>}
  </>
);

export const Campo = ({ id, etiqueta, ayuda, ...resto }: CampoBase & InputHTMLAttributes<HTMLInputElement>) => (
  <div>
    <Etiqueta id={id} etiqueta={etiqueta} ayuda={ayuda} />
    <input id={id} className={CLASE_CAMPO} {...resto} />
  </div>
);

export const CampoTexto = ({ id, etiqueta, ayuda, ...resto }: CampoBase & TextareaHTMLAttributes<HTMLTextAreaElement>) => (
  <div>
    <Etiqueta id={id} etiqueta={etiqueta} ayuda={ayuda} />
    <textarea id={id} rows={3} className={`${CLASE_CAMPO} py-3`} {...resto} />
  </div>
);

export const Selector = ({ id, etiqueta, ayuda, children, ...resto }: CampoBase & SelectHTMLAttributes<HTMLSelectElement>) => (
  <div>
    <Etiqueta id={id} etiqueta={etiqueta} ayuda={ayuda} />
    <select id={id} className={`${CLASE_CAMPO} bg-panel-solido`} {...resto}>
      {children}
    </select>
  </div>
);

export const Casilla = ({ id, etiqueta, ...resto }: { id: string; etiqueta: string } & InputHTMLAttributes<HTMLInputElement>) => (
  <label htmlFor={id} className="flex min-h-11 items-center gap-3 text-cuerpo">
    <input id={id} type="checkbox" className="size-5 accent-acento" {...resto} />
    {etiqueta}
  </label>
);

type Variante = 'primario' | 'secundario' | 'peligro';

const VARIANTES: Record<Variante, string> = {
  primario: 'border-transparent bg-acento font-semibold text-sobre-acento hover:brightness-110',
  secundario: 'border-chip-neutro-borde bg-chip-neutro-fondo text-texto hover:bg-item-activo',
  peligro: 'border-chip-alerta-borde bg-chip-alerta-fondo text-chip-alerta-texto hover:brightness-125',
};

export const Boton = ({
  variante = 'primario',
  className = '',
  type = 'button',
  ...resto
}: { variante?: Variante } & ButtonHTMLAttributes<HTMLButtonElement>) => (
  <button
    type={type}
    className={`inline-flex min-h-11 items-center justify-center gap-2 rounded-control border px-5 text-cuerpo transition disabled:opacity-50 ${ENFOQUE} ${VARIANTES[variante]} ${className}`}
    {...resto}
  />
);

export const BotonEnlace = ({
  variante = 'primario',
  className = '',
  ...resto
}: { variante?: Variante } & LinkProps) => (
  <Link
    className={`inline-flex min-h-11 items-center justify-center gap-2 rounded-control border px-5 text-cuerpo transition ${ENFOQUE} ${VARIANTES[variante]} ${className}`}
    {...resto}
  />
);

type TonoChip = 'neutro' | 'acento' | 'alerta';

const TONOS_CHIP: Record<TonoChip, string> = {
  neutro: 'border-chip-neutro-borde bg-chip-neutro-fondo text-chip-neutro-texto',
  acento: 'border-chip-acento-borde bg-chip-acento-fondo text-chip-acento-texto',
  alerta: 'border-chip-alerta-borde bg-chip-alerta-fondo text-chip-alerta-texto',
};

// El estado siempre lleva texto: el color solo lo refuerza.
export const Chip = ({ tono = 'neutro', children }: { tono?: TonoChip; children: ReactNode }) => (
  <span className={`inline-block whitespace-nowrap rounded-full border px-3 py-1 text-etiqueta font-medium ${TONOS_CHIP[tono]}`}>
    {children}
  </span>
);

export const Aviso = ({ tipo = 'error', children }: { tipo?: 'error' | 'info'; children: ReactNode }) => (
  <p
    role={tipo === 'error' ? 'alert' : 'status'}
    className={`rounded-control border p-3 text-etiqueta ${
      tipo === 'error' ? TONOS_CHIP.alerta : TONOS_CHIP.neutro
    }`}
  >
    {children}
  </p>
);

export const Titulo = ({ children, id }: { children: ReactNode; id?: string }) => (
  <h1 id={id} className="text-pagina font-semibold">{children}</h1>
);

export const Tarjeta = ({ children, denso = false, className = '' }: { children: ReactNode; denso?: boolean; className?: string }) => (
  <div className={`${denso ? 'vidrio-denso' : 'vidrio'} rounded-tarjeta p-5 md:p-6 ${className}`}>{children}</div>
);

export const Tabla = ({ children, etiqueta }: { children: ReactNode; etiqueta: string }) => (
  <div className="vidrio-denso overflow-x-auto rounded-tarjeta">
    <table className="w-full min-w-max text-left text-cuerpo">
      <caption className="sr-only">{etiqueta}</caption>
      {children}
    </table>
  </div>
);

export const Th = ({ children }: { children?: ReactNode }) => (
  <th scope="col" className="border-b border-vidrio-borde px-4 py-3 text-etiqueta font-medium text-texto-suave">{children}</th>
);

export const Td = ({ children, className = '' }: { children?: ReactNode; className?: string }) => (
  <td className={`border-b border-vidrio-borde px-4 py-3 align-middle ${className}`}>{children}</td>
);
