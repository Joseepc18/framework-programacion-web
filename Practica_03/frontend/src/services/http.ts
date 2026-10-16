// src/services/http.ts
// Cliente HTTP centralizado para el consumo de la API RESTful de MultiCatálogo.
// Inyecta automáticamente el token JWT en la cabecera Authorization: Bearer <token>
// y estandariza el manejo de errores (ApiError).

export interface ApiErrorPayload {
  status: number;
  message: string;
  details?: unknown;
}

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

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3001';

export const getStoredToken = (): string | null => {
  try {
    const raw = localStorage.getItem('multicatalogo_sesion');
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    return parsed?.token || null;
  } catch {
    return null;
  }
};

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = endpoint.startsWith('http') ? endpoint : `${API_BASE_URL}${endpoint}`;
  const headers = new Headers(options.headers || {});

  // Añadimos Content-Type si no está definido y enviamos cuerpo
  if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
    headers.set('Content-Type', 'application/json');
  }

  // Inyectamos Bearer token si existe sesión activa
  const token = getStoredToken();
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  // Manejo de respuestas no exitosas (4xx, 5xx)
  if (!response.ok) {
    let errorData: Partial<ApiErrorPayload>;
    try {
      errorData = await response.json();
    } catch {
      // Si la respuesta no es JSON válido
      errorData = { message: response.statusText || 'Error en la petición' };
    }

    throw new ApiError(
      response.status,
      errorData.message || `Error del servidor (${response.status})`,
      errorData.details
    );
  }

  // Si la respuesta es 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json() as Promise<T>;
}
