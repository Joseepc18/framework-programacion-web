import { useState } from "react";
import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";
import Navbar from "./Navbar";

const Layout = () => {
  const [isCollapsed, setIsCollapsed] = useState(false);   // colapsar en escritorio
  const [isMobileOpen, setIsMobileOpen] = useState(false); // abrir el cajón en móvil

  // El botón hamburguesa: en móvil abre/cierra el cajón; en escritorio colapsa
  const toggleSidebar = () => {
    if (window.innerWidth < 768) {
      setIsMobileOpen((prev) => !prev);
    } else {
      setIsCollapsed((prev) => !prev);
    }
  };

  const closeMobileMenu = () => setIsMobileOpen(false);

  return (
    <div className="flex h-screen bg-slate-50">
      <Sidebar
        isCollapsed={isCollapsed}
        isMobileOpen={isMobileOpen}
        closeMobileMenu={closeMobileMenu}
      />

      {/* Fondo oscuro en móvil cuando el cajón está abierto */}
      {isMobileOpen && (
        <div
          onClick={closeMobileMenu}
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
        />
      )}

      {/* Área de contenido principal */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <Navbar toggleSidebar={toggleSidebar} />
        <main className="flex-1 overflow-x-hidden overflow-y-auto p-4 md:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

export default Layout;