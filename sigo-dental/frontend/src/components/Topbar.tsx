import { Menu } from 'lucide-react';

interface TopbarProps {
  toggleSidebar: () => void;
  isMenuOpen: boolean;
}

const Topbar = ({ toggleSidebar, isMenuOpen }: TopbarProps) => (
  <header className="flex h-14 shrink-0 items-center gap-3 border-b border-slate-200 bg-white px-4 md:px-6">
    <button
      type="button"
      onClick={toggleSidebar}
      aria-label="Mostrar u ocultar el menú"
      aria-controls="barra-lateral"
      aria-expanded={isMenuOpen}
      className="rounded-md p-2 text-slate-700 hover:bg-slate-100 focus-visible:outline-2 focus-visible:outline-clinica-600"
    >
      <Menu className="size-5" aria-hidden="true" />
    </button>
    <span className="font-semibold text-slate-800">Sistema de gestión odontológica</span>
  </header>
);

export default Topbar;
