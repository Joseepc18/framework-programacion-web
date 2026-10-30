// Cliente HTTP centralizado para consumir la API de SIGO-Dental.
// La URL base sale de VITE_API_URL; si falta, se usa la API local.
// La cabecera Authorization con el JWT se agrega junto con el login.

export class ApiError extends Error {
  status: number;
  details?: unknown;

  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

const API_BASE_URL = (import.meta.env.VITE_API_URL || 'http://localhost:3001').replace(/\/+$/, '');

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  if (options.body && typeof options.body === 'string' && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, { ...options, headers });

  if (!response.ok) {
    let cuerpo: { message?: string; details?: unknown } = {};
    try {
      cuerpo = await response.json();
    } catch {
      // La respuesta de error no es JSON: se usa el texto de estado HTTP.
    }
    throw new ApiError(
      response.status,
      cuerpo.message || response.statusText || `Error del servidor (${response.status})`,
      cuerpo.details ?? cuerpo,
    );
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json() as Promise<T>;
}
