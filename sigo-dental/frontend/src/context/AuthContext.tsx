import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { useInactividad } from '../hooks/useInactividad';
import { iniciarSesion as solicitarSesion } from '../services/authService';
import { alSesionExpirada, borrarToken, guardarToken } from '../services/sesion';
import { AuthContext, type AuthContextType, type AvisoSesion, type Usuario } from './useAuth';

// El estado de la sesión (quién es y por qué se cerró) vive aquí; el token vive en services/sesion.ts.
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [usuario, setUsuario] = useState<Usuario | null>(null);
  const [aviso, setAviso] = useState<AvisoSesion | null>(null);

  const cerrarSesion = useCallback(() => {
    borrarToken();
    setUsuario(null);
    setAviso(null);
  }, []);

  const cerrarPorMotivo = useCallback((motivo: AvisoSesion) => {
    borrarToken();
    setUsuario(null);
    setAviso(motivo);
  }, []);

  const iniciarSesion = useCallback(async (email: string, password: string) => {
    const sesion = await solicitarSesion(email, password);
    guardarToken(sesion.token);
    setAviso(null);
    setUsuario({ email: sesion.email, rol: sesion.rol });
  }, []);

  const establecerSesion = useCallback((sesion: { token: string; email: string; rol: Usuario['rol'] }) => {
    guardarToken(sesion.token);
    setAviso(null);
    setUsuario({ email: sesion.email, rol: sesion.rol });
  }, []);

  // Un 401 de cualquier petición con sesión (token vencido o inválido) la cierra.
  useEffect(() => alSesionExpirada(() => cerrarPorMotivo('expirada')), [cerrarPorMotivo]);

  useInactividad(usuario !== null, () => cerrarPorMotivo('inactividad'));

  const valor = useMemo<AuthContextType>(
    () => ({ usuario, aviso, iniciarSesion, establecerSesion, cerrarSesion, descartarAviso: () => setAviso(null) }),
    [usuario, aviso, iniciarSesion, establecerSesion, cerrarSesion],
  );

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>;
};
