import { esRol, type Rol } from '../config/permisos';
import { apiFetch } from './http';

export interface Sesion {
  token: string;
  email: string;
  rol: Rol;
}

interface RespuestaLogin {
  token?: string;
  email?: string;
  rol?: string;
}

// POST /api/login responde 200 con { token, email, rol }, o 401 "Credenciales incorrectas" sea cual sea el motivo.
export async function iniciarSesion(email: string, password: string): Promise<Sesion> {
  const datos = await apiFetch<RespuestaLogin>('/api/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  if (!datos.token || !datos.email || !esRol(datos.rol)) {
    throw new Error('Respuesta de inicio de sesión inválida');
  }
  return { token: datos.token, email: datos.email, rol: datos.rol };
}
