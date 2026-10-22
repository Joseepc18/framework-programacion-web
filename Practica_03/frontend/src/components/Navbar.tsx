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
    <header className="h-16 shrink-0 bg-white border-b border-slate-200 flex items-center justify-between gap-2 px-4 md:px-8">
      <div className="min-w-0 flex items-center gap-2 md:gap-4">
        {/* Botón Hamburguesa siempre visible */}
        <button
          onClick={toggleSidebar}
          className="min-h-11 min-w-11 shrink-0 flex items-center justify-center rounded-md hover:bg-slate-100 transition text-slate-600 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          aria-label="Alternar menú"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <h2 className="truncate text-slate-600 font-medium text-base md:text-lg hidden sm:block">
          {user?.rol === "admin" ? "Panel de Administración" : "Tienda MultiCatálogo"}
        </h2>
      </div>

      <div className="shrink-0 flex items-center gap-2 md:gap-4">
        <InstalarApp ubicacion="navbar" />
        <Link
          to="/carrito"
          aria-label="Ver carrito"
          className="relative min-h-11 min-w-11 flex items-center justify-center hover:bg-slate-100 rounded-full transition"
        >
          <span className="text-xl">🛒</span>
          {totalItems > 0 && (
            <span className="absolute top-0 right-0 bg-indigo-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full transform translate-x-1 -translate-y-1">
              {totalItems}
            </span>
          )}
        </Link>

        <div className="flex items-center gap-2 md:gap-4">
          <span className="max-w-32 truncate text-xs md:text-sm text-slate-500 hidden sm:block">{user?.email}</span>
          <span
            className={`hidden md:inline-block text-xs font-semibold px-2 py-1 rounded-full uppercase ${
              user?.rol === "admin"
                ? "bg-amber-100 text-amber-700"
                : "bg-indigo-100 text-indigo-700"
            }`}
          >
            {user?.rol}
          </span>

          <div className="relative">
            {/* Avatar: al hacer click abre/cierra el menú (funciona en móvil y escritorio) */}
            <button
              onClick={() => setUserMenuOpen((prev) => !prev)}
              className="w-11 h-11 rounded-full bg-indigo-600 text-white font-bold uppercase flex items-center justify-center border border-slate-300"
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
                <div className="absolute right-0 top-full mt-1 w-36 bg-white border border-slate-200 rounded-lg shadow-lg z-50">
                  <button
                    onClick={handleLogout}
                    className="w-full min-h-11 text-left px-4 py-2 text-sm text-red-600 font-semibold hover:bg-red-50 rounded-md transition-colors"
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
