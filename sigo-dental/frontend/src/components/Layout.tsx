import { useState } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import Topbar from './Topbar';

// Punto de corte de Tailwind `md`: por debajo la barra lateral es un cajón; desde ahí, una columna fija.
const ESCRITORIO = '(min-width: 768px)';

const Layout = () => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  const toggleSidebar = () => {
    if (window.matchMedia(ESCRITORIO).matches) {
      setIsCollapsed((prev) => !prev);
    } else {
      setIsMobileOpen((prev) => !prev);
    }
  };

  const closeMobileMenu = () => setIsMobileOpen(false);

  return (
    <div className="app-layout flex">
      <Sidebar isCollapsed={isCollapsed} isMobileOpen={isMobileOpen} closeMobileMenu={closeMobileMenu} />

      {isMobileOpen && (
        <div onClick={closeMobileMenu} aria-hidden="true" className="fixed inset-0 z-40 bg-slate-900/60 md:hidden" />
      )}

      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <Topbar toggleSidebar={toggleSidebar} isMenuOpen={isMobileOpen || !isCollapsed} />
        <main className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden p-4 md:p-8">
          <div className="mx-auto w-full max-w-[1600px]">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;
