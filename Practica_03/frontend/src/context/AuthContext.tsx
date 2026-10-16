// src/context/AuthContext.tsx
import { createContext, useContext, useState, type ReactNode } from 'react';

// 1. Tipos de rol que maneja la aplicación
export type Rol = 'admin' | 'cliente';

// 2. Usuario autenticado: correo + rol (lo entrega la API en /api/login)
export interface Usuario {
  email: string;
  rol: Rol;
}

interface SesionGuardada {
  user: Usuario;
  token: string;
}

interface AuthContextType {
  isAuthenticated: boolean;
  user: Usuario | null;
  token: string | null;
  login: (usuario: Usuario, token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth debe ser usado dentro de un AuthProvider");
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

const STORAGE_KEY = 'multicatalogo_sesion';

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [user, setUser] = useState<Usuario | null>(() => {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      if (!data) return null;
      const parsed: SesionGuardada = JSON.parse(data);
      return parsed.user || null;
    } catch {
      return null;
    }
  });

  const [token, setToken] = useState<string | null>(() => {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      if (!data) return null;
      const parsed: SesionGuardada = JSON.parse(data);
      return parsed.token || null;
    } catch {
      return null;
    }
  });

  const isAuthenticated = Boolean(user && token);

  const login = (usuario: Usuario, jwtToken: string) => {
    setUser(usuario);
    setToken(jwtToken);
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ user: usuario, token: jwtToken }));
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    localStorage.removeItem(STORAGE_KEY);
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};