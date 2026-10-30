// Service worker de SIGO-Dental (F2-8): deja instalable la aplicación y hace que el cascarón (HTML, JS, CSS, fuentes e
// iconos) abra sin conexión. NUNCA guarda datos de la API ni respuestas con información clínica: solo recursos
// estáticos del propio sitio. Sin conexión, la pantalla abre pero las consultas muestran que no hay conexión.
const VERSION = 'sigo-dental-v1';
const CASCARON = ['/', '/index.html', '/manifest.webmanifest', '/favicon.svg', '/icons/icono-192.png', '/icons/icono-512.png'];

self.addEventListener('install', (evento) => {
  evento.waitUntil(caches.open(VERSION).then((cache) => cache.addAll(CASCARON)).then(() => self.skipWaiting()));
});

self.addEventListener('activate', (evento) => {
  evento.waitUntil(
    caches.keys()
      .then((claves) => Promise.all(claves.filter((clave) => clave !== VERSION).map((clave) => caches.delete(clave))))
      .then(() => self.clients.claim()),
  );
});

self.addEventListener('fetch', (evento) => {
  const peticion = evento.request;
  const url = new URL(peticion.url);
  // Solo GET del mismo origen: la API vive en otro origen y no se toca.
  if (peticion.method !== 'GET' || url.origin !== self.location.origin) return;

  // Navegación: primero la red (para tener siempre la versión nueva); sin conexión, el cascarón guardado.
  if (peticion.mode === 'navigate') {
    evento.respondWith(fetch(peticion).catch(() => caches.match('/index.html')));
    return;
  }

  // Recursos con nombre con huella (assets/) o iconos: del caché si ya están; si no, de la red y se guardan.
  evento.respondWith(
    caches.match(peticion).then((guardado) => {
      if (guardado) return guardado;
      return fetch(peticion).then((respuesta) => {
        if (respuesta.ok && (url.pathname.startsWith('/assets/') || url.pathname.startsWith('/icons/'))) {
          const copia = respuesta.clone();
          caches.open(VERSION).then((cache) => cache.put(peticion, copia));
        }
        return respuesta;
      });
    }),
  );
});
