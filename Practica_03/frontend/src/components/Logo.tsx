// Marca de JP Mercantil: ícono de tienda sobre fondo azul marino y, opcionalmente, el nombre.
interface LogoProps {
  // Lado del monograma en píxeles.
  size?: number;
  // Muestra el nombre junto al monograma.
  conNombre?: boolean;
  // "claro" para fondos oscuros, "oscuro" para fondos claros.
  tono?: "claro" | "oscuro";
  className?: string;
}

const Logo = ({ size = 36, conNombre = true, tono = "oscuro", className = "" }: LogoProps) => {
  const colorNombre = tono === "claro" ? "text-arena-200" : "text-tinta";

  return (
    <span className={`inline-flex items-center gap-3 ${className}`}>
      <svg
        width={size}
        height={size}
        viewBox="0 0 512 512"
        role="img"
        aria-label="JP Mercantil"
        className="shrink-0"
      >
        <rect width="512" height="512" fill="#0e2a47" />
        {/* Ícono "store" de Lucide (licencia ISC) */}
        <g
          transform="translate(96 96) scale(13.33)"
          fill="none"
          stroke="#f2e3c8"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="m2 7 4.41-4.41A2 2 0 0 1 7.83 2h8.34a2 2 0 0 1 1.42.59L22 7" />
          <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
          <path d="M15 22v-4a2 2 0 0 0-2-2h-2a2 2 0 0 0-2 2v4" />
          <path d="M2 7h20" stroke="#ff6b4a" />
          <path
            d="M22 7v3a2 2 0 0 1-2 2a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 16 12a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 12 12a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 8 12a2.7 2.7 0 0 1-1.59-.63.7.7 0 0 0-.82 0A2.7 2.7 0 0 1 4 12a2 2 0 0 1-2-2V7"
            stroke="#ff6b4a"
          />
        </g>
      </svg>
      {conNombre && (
        <span className={`font-display text-xl font-semibold leading-none tracking-tight ${colorNombre}`}>
          JP <span className="font-normal">Mercantil</span>
        </span>
      )}
    </span>
  );
};

export default Logo;
