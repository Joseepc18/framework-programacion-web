// src/App.tsx
import { type ReactNode } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './components/Dashboard';
import Catalogo from './components/Catalogo';
import Storefront from './components/Storefront';
import DetalleProducto from './components/DetalleProducto';
import MiRed from './components/MiRed';
import Carrito from './components/Carrito';
import Checkout from './components/Checkout';
import Confirmacion from './components/Confirmacion';
import Login from './components/Login';
import Registro from './components/Registro';
import AdminProductos from './components/AdminProductos';
import MisPedidos from './components/MisPedidos';
import AdminPedidos from './components/AdminPedidos';
import MisComisiones from './components/MisComisiones';
import AdminLiquidaciones from './components/AdminLiquidaciones';
import AdminUsuarios from './components/AdminUsuarios';
import { CartProvider } from './context/CartContext';
import { AuthProvider, useAuth } from './context/AuthContext';

// Componente para proteger las rutas privadas: exige estar autenticado
const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();

  // Si no está autenticado, lo enviamos al login
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Si está autenticado, renderiza las rutas hijas (Outlet)
  return <Outlet />;
};

// Componente para rutas exclusivas del administrador (Tema 5: roles)
const AdminRoute = () => {
  const { isAuthenticated, user } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Si el usuario no es admin, lo enviamos a la tienda (su vista por defecto)
  if (user?.rol !== 'admin') {
    return <Navigate to="/tienda" replace />;
  }

  return <Outlet />;
};

// Proveedor intermedio del carrito: al cambiar de usuario se remonta con key,
// de modo que cada cuenta lea y persista SU propio carrito en localStorage.
const CartBoundary = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  return (
    <CartProvider key={user?.email ?? "anonimo"}>
      {children}
    </CartProvider>
  );
};

function App() {
  return (
    <AuthProvider> {/* Proveedor de Autenticación */}
      <CartBoundary> {/* Proveedor del Carrito (uno por usuario) */}
        <BrowserRouter>
          <Routes>
            {/* Rutas públicas */}
            <Route path="/login" element={<Login />} />
            <Route path="/registro" element={<Registro />} />

            {/* Rutas protegidas (requieren login) */}
            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                {/* Solo administrador: Dashboard y Red multinivel */}
                <Route element={<AdminRoute />}>
                  <Route index element={<Dashboard />} />
                  <Route path="admin/productos" element={<AdminProductos />} />
                  <Route path="admin/pedidos" element={<AdminPedidos />} />
                  <Route path="admin/liquidaciones" element={<AdminLiquidaciones />} />
                  <Route path="admin/usuarios" element={<AdminUsuarios />} />
                </Route>

                {/* Ambos roles: flujo de compra en línea */}
                <Route path="tienda" element={<Storefront />} />
                <Route path="catalogo" element={<Catalogo />} />
                <Route path="producto/:id" element={<DetalleProducto />} />
                <Route path="carrito" element={<Carrito />} />
                <Route path="checkout" element={<Checkout />} />
                <Route path="confirmacion" element={<Confirmacion />} />
                {/* Historial propio: el backend filtra por el usuario del token */}
                <Route path="mis-pedidos" element={<MisPedidos />} />
                {/* Red y ledger propios: cada cuenta ve SU subárbol y SU ledger */}
                <Route path="mi-red" element={<MiRed />} />
                <Route path="comisiones" element={<MisComisiones />} />
              </Route>
            </Route>

            {/* Ruta comodín para capturar 404 y redirigir */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </CartBoundary>
    </AuthProvider>
  );
}

export default App;