import { LogOut, Menu } from 'lucide-react';
import { ETIQUETA_ROL } from '../config/permisos';
import { useAuth } from '../context/useAuth';

interface TopbarProps {
  toggleSidebar: () => void;
  isMenuOpen: boolean;
}

const Topbar = ({ toggleSidebar, isMenuOpen }: TopbarProps) => {
  const { usuario, cerrarSesion } = useAuth();

  return (
    <header className="vidrio flex h-16 shrink-0 items-center gap-3 border-x-0 border-t-0 px-4 md:px-6">
      <button
        type="button"
        onClick={toggleSidebar}
        aria-label="Mostrar u ocultar el menú"
        aria-controls="barra-lateral"
        aria-expanded={isMenuOpen}
        className="grid size-11 place-items-center rounded-control text-texto hover:bg-chip-neutro-fondo focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
      >
        <Menu className="size-5" aria-hidden="true" />
      </button>
      <span className="hidden text-etiqueta font-medium text-texto-suave sm:inline sm:text-cuerpo">
        Sistema de gestión odontológica
      </span>

      {usuario && (
        <div className="ml-auto flex min-w-0 items-center gap-3">
          <div className="min-w-0 text-right leading-tight">
            <p className="hidden truncate text-etiqueta text-texto-suave md:block">{usuario.email}</p>
            <p className="truncate text-etiqueta font-medium text-texto">{ETIQUETA_ROL[usuario.rol]}</p>
          </div>
          <button
            type="button"
            onClick={cerrarSesion}
            className="inline-flex min-h-11 items-center gap-2 rounded-control border border-chip-neutro-borde bg-chip-neutro-fondo px-4 text-cuerpo font-medium text-texto transition hover:bg-item-activo focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-acento"
          >
            <LogOut className="size-4" aria-hidden="true" />
            <span className="hidden sm:inline">Cerrar sesión</span>
            <span className="sr-only sm:hidden">Cerrar sesión</span>
          </button>
        </div>
      )}
    </header>
  );
};

export default Topbar;
