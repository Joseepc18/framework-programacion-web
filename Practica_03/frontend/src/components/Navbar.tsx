// src/components/Navbar.tsx
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { useAuth } from "../context/AuthContext";
import InstalarApp from './InstalarApp';

// Definimos la interfaz para las props
interface NavbarProps {
  toggleSidebar: () => void;
}

const Navbar = ({ toggleSidebar }: NavbarProps) => {
  const { totalItems } = useCart();
  const { logout, user } = useAuth();
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const navigate = useNavigate();

  const handleLogout = () => {
    setUserMenuOpen(false);
    logout();
    navigate("/login");
  };

  return (
    <header className="h-16 shrink-0 bg-papel border-b border-linea flex items-center justify-between gap-2 px-4 md:px-8">
      <div className="min-w-0 flex items-center gap-2 md:gap-4">
        {/* Botón Hamburguesa siempre visible */}
        <button
          onClick={toggleSidebar}
          className="min-h-11 min-w-11 shrink-0 flex items-center justify-center hover:bg-crema transition text-slate-600 focus:outline-none focus:ring-2 focus:ring-verde-500"
          aria-label="Alternar menú"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <h2 className="truncate text-tinta text-lg md:text-xl hidden sm:block">
          {user?.rol === "admin" ? "Panel de administración" : "JP Mercantil"}
        </h2>
      </div>

      <div className="shrink-0 flex items-center gap-2 md:gap-4">
        <InstalarApp ubicacion="navbar" />
        <Link
          to="/carrito"
          aria-label="Ver carrito"
          className="relative min-h-11 min-w-11 flex items-center justify-center text-tinta hover:bg-crema transition"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" strokeWidth={1.75} viewBox="0 0 24 24" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 8h14l-1 12H6L5 8zM9 8V6a3 3 0 016 0v2" />
          </svg>
          {totalItems > 0 && (
            <span className="absolute top-0.5 right-0 bg-arcilla-500 text-white text-[11px] font-bold min-w-5 h-5 px-1 flex items-center justify-center">
              {totalItems}
            </span>
          )}
        </Link>

        <div className="flex items-center gap-2 md:gap-4">
          <span className="max-w-32 truncate text-xs md:text-sm text-slate-500 hidden sm:block">{user?.email}</span>
          <span
            className={`hidden md:inline-block border text-[11px] font-semibold px-2 py-1 uppercase tracking-widest ${
              user?.rol === "admin"
                ? "border-arcilla-300 bg-arcilla-50 text-arcilla-700"
                : "border-verde-200 bg-verde-50 text-verde-700"
            }`}
          >
            {user?.rol}
          </span>

          <div className="relative">
            {/* Avatar: al hacer click abre/cierra el menú (funciona en móvil y escritorio) */}
            <button
              onClick={() => setUserMenuOpen((prev) => !prev)}
              className="w-11 h-11 bg-verde-700 text-crema font-serif text-lg uppercase flex items-center justify-center focus:outline-none focus:ring-2 focus:ring-verde-500 focus:ring-offset-2 focus:ring-offset-papel"
              aria-label="Menú de usuario"
              aria-expanded={userMenuOpen}
            >
              {user?.email?.charAt(0) ?? "?"}
            </button>

            {userMenuOpen && (
              <>
                {/* Capa invisible: al tocar fuera del menú, se cierra */}
                <div
                  className="fixed inset-0 z-40"
                  onClick={() => setUserMenuOpen(false)}
                />
                <div className="absolute right-0 top-full mt-1 w-40 bg-papel border border-linea z-50">
                  <button
                    onClick={handleLogout}
                    className="w-full min-h-11 text-left px-4 py-2 text-sm text-arcilla-700 font-semibold hover:bg-arcilla-50 transition-colors"
                  >
                    Cerrar Sesión
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Navbar;
