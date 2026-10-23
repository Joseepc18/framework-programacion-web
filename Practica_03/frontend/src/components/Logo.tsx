// Marca de JP Mercantil: monograma JP sobre fondo verde y, opcionalmente, el nombre en serif.
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
  const colorNombre = tono === "claro" ? "text-crema" : "text-tinta";

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
        <rect width="512" height="512" fill="#1f3a2e" />
        <g
          transform="translate(256 256) scale(.74) translate(-256 -266)"
          fill="none"
          stroke="#f6f1e7"
          strokeWidth="40"
          strokeLinecap="square"
        >
          <path d="M136 112h80" />
          <path d="M176 112v192a72 72 0 0 1-72 72" />
          <path d="M288 112v268" />
          <path d="M288 112h48a72 72 0 0 1 0 144h-48" />
          <rect x="84" y="424" width="344" height="16" fill="#b4532a" stroke="none" />
        </g>
      </svg>
      {conNombre && (
        <span className={`font-serif text-xl font-semibold leading-none tracking-tight ${colorNombre}`}>
          JP <span className="font-normal">Mercantil</span>
        </span>
      )}
    </span>
  );
};

export default Logo;
