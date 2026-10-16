// src/components/Sidebar.tsx
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

// NUEVO: Agregamos la función a la interfaz de props
interface SidebarProps {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  closeMobileMenu: () => void;
}

// Definimos las opciones de navegación según el rol (Tema 5)
interface NavItem {
  to: string;
  label: string;
  title: string;
  icon: string; // ruta del icono SVG (stroke)
  soloAdmin?: boolean;
}

const NAV_ITEMS: NavItem[] = [
  {
    to: "/",
    label: "Dashboard",
    title: "Dashboard",
    soloAdmin: true,
    icon: "M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6",
  },
  {
    to: "/tienda",
    label: "Tienda",
    title: "Tienda",
    icon: "M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z",
  },
  {
    to: "/catalogo",
    label: "Catálogo",
    title: "Catálogo",
    icon: "M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z",
  },
  {
    to: "/mi-red",
    label: "Mi Red",
    title: "Mi Red",
    icon: "M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z",
  },
  {
    to: "/comisiones",
    label: "Comisiones",
    title: "Mis comisiones",
    icon: "M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 9v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
  },
  {
    to: "/mis-pedidos",
    label: "Mis Pedidos",
    title: "Historial de mis pedidos",
    icon: "M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4",
  },
  {
    to: "/admin/pedidos",
    label: "Pedidos",
    title: "Gestión de pedidos",
    soloAdmin: true,
    icon: "M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z",
  },
  {
    to: "/admin/productos",
    label: "Productos",
    title: "Administración de productos",
    soloAdmin: true,
    icon: "M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4",
  },
  {
    to: "/admin/usuarios",
    label: "Usuarios",
    title: "Administración de usuarios",
    soloAdmin: true,
    icon: "M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z",
  },
  {
    to: "/admin/liquidaciones",
    label: "Liquidaciones",
    title: "Liquidación de comisiones",
    soloAdmin: true,
    icon: "M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z",
  },
];

const Sidebar = ({ isCollapsed, isMobileOpen, closeMobileMenu }: SidebarProps) => {
  const { user } = useAuth();
  const { pathname } = useLocation();

  // Filtramos las opciones según el rol del usuario
  const items = NAV_ITEMS.filter((item) => !item.soloAdmin || user?.rol === "admin");

  return (
    <aside
      className={`
        fixed inset-y-0 left-0 z-50 transform bg-green-500 text-white flex flex-col transition-all duration-300 ease-in-out
        ${isMobileOpen ? "translate-x-0" : "-translate-x-full"}
        w-64
        md:relative md:translate-x-0
        ${isCollapsed ? "md:w-20" : "md:w-64"}
      `}
    >
      <div className={`p-4 md:p-6 text-xl font-bold border-b border-slate-700 flex items-center ${isCollapsed ? 'md:justify-center' : 'justify-start'} whitespace-nowrap`}>
        <span className="md:hidden">MultiCatálogo</span>
        <span className="hidden md:inline">{isCollapsed ? "MC" : "MultiCatálogo"}</span>
      </div>

      <nav className="flex-1 p-3 space-y-2 overflow-y-auto">
        {items.map((item) => {
          // Resaltamos la opción activa según la ruta actual
          const esActivo =
            item.to === "/"
              ? pathname === "/"
              : pathname.startsWith(item.to);

          return (
            <Link
              key={item.to}
              to={item.to}
              onClick={closeMobileMenu}
              className={`flex items-center gap-3 p-3 rounded transition ${
                esActivo
                  ? "bg-indigo-600 text-white"
                  : "hover:bg-slate-800"
              } ${isCollapsed ? 'md:justify-center' : ''}`}
              title={item.title}
            >
              <svg className="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={item.icon} />
              </svg>
              <span className={`whitespace-nowrap ${isCollapsed ? 'md:hidden' : ''}`}>
                {item.label}
              </span>
            </Link>
          );
        })}
      </nav>

      <div className="p-3 border-t border-slate-700 text-xs text-slate-300">
        {isCollapsed ? (
          <p className="text-center uppercase">{user?.rol}</p>
        ) : (
          <p>
            Conectado como <span className="font-semibold uppercase">{user?.rol}</span>
          </p>
        )}
      </div>
    </aside>
  );
};

export default Sidebar;