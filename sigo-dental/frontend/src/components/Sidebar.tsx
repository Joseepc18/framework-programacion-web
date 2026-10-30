import { NavLink } from 'react-router-dom';
import { Activity, House, type LucideIcon } from 'lucide-react';

interface SidebarProps {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  closeMobileMenu: () => void;
}

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
}

const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Inicio', icon: House },
  { to: '/estado', label: 'Estado del sistema', icon: Activity },
];

// Desde 768px (md) la barra es fija y puede colapsarse a solo iconos; por debajo es un cajón deslizable.
const Sidebar = ({ isCollapsed, isMobileOpen, closeMobileMenu }: SidebarProps) => (
  <aside
    id="barra-lateral"
    className={`fixed inset-y-0 left-0 z-50 flex w-64 shrink-0 flex-col bg-clinica-900 text-clinica-100 transition-all duration-300 ease-in-out md:relative md:translate-x-0 ${
      isMobileOpen ? 'translate-x-0' : '-translate-x-full'
    } ${isCollapsed ? 'md:w-20' : 'md:w-64'}`}
  >
    <div className="flex h-14 items-center border-b border-white/10 px-5 whitespace-nowrap">
      <span className="text-lg font-bold text-white">
        SIGO<span className={isCollapsed ? 'md:hidden' : ''}>-Dental</span>
      </span>
    </div>

    <nav aria-label="Principal" className="flex-1 space-y-0.5 overflow-y-auto py-3">
      {NAV_ITEMS.map(({ to, label, icon: Icono }) => (
        <NavLink
          key={to}
          to={to}
          end={to === '/'}
          onClick={closeMobileMenu}
          title={label}
          className={({ isActive }) =>
            `flex items-center gap-3 border-l-4 px-5 py-3 transition focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-clinica-200 ${
              isActive
                ? 'border-clinica-500 bg-white/10 font-semibold text-white'
                : 'border-transparent hover:bg-white/5 hover:text-white'
            } ${isCollapsed ? 'md:justify-center md:px-0' : ''}`
          }
        >
          <Icono className="size-5 shrink-0" aria-hidden="true" />
          <span className={`whitespace-nowrap ${isCollapsed ? 'md:hidden' : ''}`}>{label}</span>
        </NavLink>
      ))}
    </nav>
  </aside>
);

export default Sidebar;
