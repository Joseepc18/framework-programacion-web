import { createContext, useContext } from 'react';
import type { Rol } from '../config/permisos';

export interface Usuario {
  email: string;
  rol: Rol;
}

// Motivo por el que se cerró la sesión sin que la persona lo pidiera; el login lo muestra.
export type AvisoSesion = 'inactividad' | 'expirada';

export interface AuthContextType {
  usuario: Usuario | null;
  aviso: AvisoSesion | null;
  iniciarSesion: (email: string, password: string) => Promise<void>;
  // Abre la sesión con un token que ya se obtuvo por otro camino (ingreso al portal con código OTP).
  establecerSesion: (sesion: { token: string; email: string; rol: Rol }) => void;
  cerrarSesion: () => void;
  descartarAviso: () => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const contexto = useContext(AuthContext);
  if (!contexto) {
    throw new Error('useAuth debe usarse dentro de un AuthProvider');
  }
  return contexto;
};
