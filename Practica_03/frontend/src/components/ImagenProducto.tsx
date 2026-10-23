// Imagen de producto con reemplazo: si la URL no carga (enlace roto o que no es una
// imagen), se muestra un recuadro con el ícono de tienda en lugar del ícono roto del navegador.
import { useState } from "react";

interface ImagenProductoProps {
  src: string;
  alt: string;
  className?: string;
}

const ImagenProducto = ({ src, alt, className = "" }: ImagenProductoProps) => {
  // Guardamos la URL que falló, de modo que cambiar de producto reintenta con la nueva.
  const [fallida, setFallida] = useState<string | null>(null);
  const roto = !src || fallida === src;

  if (roto) {
    return (
      <div
        role="img"
        aria-label={alt}
        className={`flex items-center justify-center bg-slate-100 text-slate-400 ${className}`}
      >
        <svg
          viewBox="0 0 24 24"
          className="h-1/3 max-h-16 w-1/3 max-w-16"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <path d="m2 7 4.41-4.41A2 2 0 0 1 7.83 2h8.34a2 2 0 0 1 1.42.59L22 7" />
          <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
          <path d="M15 22v-4a2 2 0 0 0-2-2h-2a2 2 0 0 0-2 2v4" />
          <path d="M2 7h20" />
        </svg>
      </div>
    );
  }

  return (
    <img
      src={src}
      alt={alt}
      loading="lazy"
      onError={() => setFallida(src)}
      className={className}
    />
  );
};

export default ImagenProducto;
