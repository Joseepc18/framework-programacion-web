// src/context/AuthContext.tsx
import { createContext, useContext, useState, type ReactNode } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  userEmail: string | null; // <-- 1. Nuevo estado para el correo
  login: (email: string) => void; // <-- 2. La función ahora recibe el correo
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

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [userEmail, setUserEmail] = useState<string | null>(null); // <-- 3. Estado local

  // 4. Actualizamos las funciones
  const login = (email: string) => {
    setIsAuthenticated(true);
    setUserEmail(email); // Guardamos el correo
  };
  
  const logout = () => {
    setIsAuthenticated(false);
    setUserEmail(null); // Limpiamos el correo al salir
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, userEmail, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};