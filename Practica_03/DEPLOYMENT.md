# Publicar MultiCatálogo

El frontend se publica en Vercel, la API y la caché en Render y PostgreSQL en Neon. El entorno local con Docker Compose se conserva.

## 1. PostgreSQL en Neon

1. Crear un proyecto y una base vacía. Copiar su URL de conexión con `sslmode=require` desde el panel; no subirla a Git.
2. Ejecutar `backend/database/init.sql` una sola vez en esa base mediante el editor SQL de Neon. No ejecutar las migraciones históricas sobre esta instalación nueva.
3. Antes de compartir el enlace, cambiar las contraseñas de las cuentas de demostración sembradas por el script. Conservar bcrypt; no guardar contraseñas en texto plano.

El script contiene datos semilla y no está diseñado para repetirse sobre una base con datos. No borra ni migra la base de Docker del equipo.

## 2. Caché y API en Render

1. Crear una instancia **Key Value**, plan Free, y copiar su URL interna de conexión.
2. Crear un **Web Service** desde este repositorio, rama `main`, en la misma región que Key Value.
3. Seleccionar runtime **Docker**, Root Directory `Practica_03/backend`, Dockerfile `./Dockerfile` y contexto `.`.
4. Configurar Health Check Path `/api/health` y estas variables:

| Variable | Valor |
|---|---|
| `DATABASE_URL` | URL de Neon con SSL. Tiene prioridad sobre `DB_*`. |
| `REDIS_URL` | URL interna de Key Value. Tiene prioridad sobre `REDIS_ADDR`. |
| `JWT_SECRET` | Secreto aleatorio privado de al menos 32 caracteres. |
| `APP_ENV` | `production` |
| `CORS_ALLOWED_ORIGINS` | Origen HTTPS exacto de Vercel, sin `/` final. Varios orígenes se separan por comas. |

Render asigna `PORT`; no fijarlo manualmente. La API mantiene Redis como dependencia de arranque. Esperar a que la caché esté disponible antes del despliegue.

## 3. Frontend en Vercel

1. Importar el repositorio, rama de producción `main`, Root Directory `Practica_03/frontend` y preset **Vite**.
2. Usar Build Command `npm run build` y Output Directory `dist`.
3. Definir `VITE_API_URL=https://<nombre-api>.onrender.com`, sin `/` final ni sufijo `/api`, y desplegar. Esta variable es pública y se incorpora durante el build.
4. Copiar el origen definitivo de Vercel a `CORS_ALLOWED_ORIGINS` en Render y volver a desplegar la API. Los dominios de previews necesitan autorización explícita si se van a utilizar.

`frontend/vercel.json` permite abrir y recargar rutas como `/login` o `/tienda`. Cambiar `VITE_API_URL` requiere un nuevo build del frontend.

## Verificación después de publicar

- Abrir `https://<nombre-api>.onrender.com/api/health`: debe indicar PostgreSQL y Redis operativos.
- Abrir `/login` y recargar `/tienda` directamente desde el dominio de Vercel.
- Iniciar sesión, consultar productos y comprobar permisos de cliente y administrador.
- Crear un pedido de prueba y confirmar stock, historial y comisiones.
- Desde el celular, comprobar navegación, tablas, instalación y apertura de la PWA.
- Probar el catálogo offline después de visitarlo con conexión. Login y compras requieren conexión.

El backend gratuito de Render se suspende tras 15 minutos sin tráfico y tarda aproximadamente un minuto en despertar. Su plan de cómputo de pago elimina esa suspensión. Neon y Key Value Free tienen cuotas; la caché puede perderse al reiniciar y se reconstruye desde PostgreSQL.

## Referencias

- [Render: Docker y monorepos](https://render.com/docs/monorepo-support)
- [Render Key Value](https://render.com/docs/key-value)
- [Límites gratuitos de Render](https://render.com/docs/free)
- [Vite en Vercel](https://vercel.com/docs/frameworks/frontend/vite)
- [Plan de Neon](https://neon.com/pricing)
