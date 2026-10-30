import { Download } from 'lucide-react';
import { useInstalarApp } from '../pwa/instalacion';

// Ofrece instalar la aplicación cuando el navegador lo permite; si no, no ocupa lugar.
const BotonInstalar = ({ className = '' }: { className?: string }) => {
  const { disponible, instalar } = useInstalarApp();
  if (!disponible) return null;

  return (
    <button
      type="button"
      onClick={() => void instalar()}
      className={`inline-flex min-h-11 items-center justify-center gap-2 rounded-control border border-chip-acento-borde bg-chip-acento-fondo px-4 text-cuerpo font-medium text-chip-acento-texto transition hover:brightness-125 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento ${className}`}
    >
      <Download className="size-4" aria-hidden="true" />
      <span>Instalar app</span>
    </button>
  );
};

export default BotonInstalar;
