// src/hooks/useInstallPrompt.ts
// Estado compartido entre el login y la barra de navegación.
import { createContext, useContext } from 'react';

export interface InstallPromptState {
  instalada: boolean;
  puedeInstalar: boolean;
  instalando: boolean;
  instruccionesIOS: 'safari' | 'otro-navegador' | null;
  error: string;
  instalar: () => Promise<void>;
}

export const InstallPromptContext = createContext<InstallPromptState | null>(null);

export function useInstallPrompt() {
  const context = useContext(InstallPromptContext);
  if (!context) throw new Error('useInstallPrompt requiere InstallPromptProvider');
  return context;
}
