import { Navigate, Outlet } from 'react-router-dom';
import type { Rol } from '../config/permisos';
import { useAuth } from '../context/useAuth';
import AccesoDenegadoPage from '../pages/AccesoDenegadoPage';

// Sin sesión lleva al login; con sesión pero sin el rol requerido muestra "Acceso denegado".
// Solo es la parte visible del control: el backend vuelve a validar el rol en cada petición.
const RutaProtegida = ({ roles }: { roles?: readonly Rol[] }) => {
  const { usuario } = useAuth();

  if (!usuario) return <Navigate to="/login" replace />;
  if (roles && !roles.includes(usuario.rol)) return <AccesoDenegadoPage />;
  return <Outlet />;
};

export default RutaProtegida;
