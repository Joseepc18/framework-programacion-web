import { Menu } from 'lucide-react';

interface TopbarProps {
  toggleSidebar: () => void;
  isMenuOpen: boolean;
}

const Topbar = ({ toggleSidebar, isMenuOpen }: TopbarProps) => (
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
    <span className="text-etiqueta font-medium text-texto-suave sm:text-cuerpo">Sistema de gestión odontológica</span>
  </header>
);

export default Topbar;
