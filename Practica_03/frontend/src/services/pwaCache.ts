// src/services/pwaCache.ts
// Nombre de la caché de runtime de la API pública (debe coincidir con `cacheName`
// en vite.config.ts). Se borra al cerrar sesión para no dejar datos entre usuarios.
export const API_CACHE_NAME = 'api-productos';

export const limpiarCacheApi = async (): Promise<void> => {
  if (!('caches' in window)) return;
  try {
    await caches.delete(API_CACHE_NAME);
  } catch {
    // Sin Cache Storage disponible no hay nada que limpiar.
  }
};
