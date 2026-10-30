import { ApiError, apiFetch } from './http';

export interface HealthStatus {
  status: 'ok' | 'degraded' | 'error';
  database: 'up' | 'down';
  redis?: 'up' | 'down';
  message: string;
}

// GET /api/health responde 503 con un cuerpo válido cuando PostgreSQL está caído:
// ese caso es un resultado que se muestra, no un fallo de red.
export async function obtenerEstadoSalud(): Promise<HealthStatus> {
  try {
    return await apiFetch<HealthStatus>('/api/health');
  } catch (error) {
    if (error instanceof ApiError && error.status === 503) {
      const detalle = error.details as Partial<HealthStatus> | undefined;
      if (detalle?.status === 'error') {
        return { status: 'error', database: 'down', message: error.message, ...detalle };
      }
    }
    throw error;
  }
}
