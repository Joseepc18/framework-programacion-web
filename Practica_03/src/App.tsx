// src/App.tsx
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './components/Dashboard';
import Catalogo from './components/Catalogo';
import MiRed from './components/MiRed';
import Carrito from './components/Carrito';
import Login from './components/Login';
import { CartProvider } from './context/CartContext';
import { AuthProvider, useAuth } from './context/AuthContext';

// Componente para proteger las rutas privadas
const ProtectedRoute = () => {
  const { isAuthenticated } = useAuth();
  
  // Si no está autenticado, lo enviamos al login
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  // Si está autenticado, renderiza las rutas hijas (Outlet)
  return <Outlet />;
};

function App() {
  return (
    <AuthProvider> {/* Proveedor de Autenticación */}
      <CartProvider> {/* Proveedor del Carrito */}
        <BrowserRouter>
          <Routes>
            {/* Ruta pública */}
            <Route path="/login" element={<Login />} />

            {/* Rutas protegidas */}
            <Route element={<ProtectedRoute />}>
              <Route path="/" element={<Layout />}>
                <Route index element={<Dashboard />} />
                <Route path="catalogo" element={<Catalogo />} />
                <Route path="mi-red" element={<MiRed />} />
                <Route path="carrito" element={<Carrito />} />
              </Route>
            </Route>
            
            {/* Ruta comodín para capturar 404 y redirigir */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </CartProvider>
    </AuthProvider>
  );
}

export default App;