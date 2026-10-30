# SIGO-Dental

Sistema integral de gestión odontológica y expediente clínico digital: aplicación web responsive para clínicas odontológicas (pacientes, expediente y odontograma, presupuestos, agenda, caja e inventario).

Proyecto de la materia Framework de Programación Web (UPSE). Se construye a partir del Documento de Requerimientos de Software v1.0 (29/10/2026). Trabajo individual.

## Estado

En construcción: existe el backend (`backend/`) con `GET /api/health`, el modelo de datos de usuarios y roles, y el login con JWT y control de acceso por roles; el frontend (`frontend/`) tiene el esqueleto con el layout responsive y una página que muestra `GET /api/health`, y aún no tiene login ni módulos clínicos. El proyecto se construye por etapas; cada etapa queda desplegada y funcionando antes de empezar la siguiente.

## Alcance

- **Dentro:** portal web responsive para personal clínico y administrativo, con roles (administrador TI, recepcionista o cajero, odontólogo, asistente dental, auditor o contador).
- **Fuera de la Fase 1:** app móvil nativa, procesamiento de imágenes 3D y visor DICOM/PACS.
- **Integraciones externas** (facturación fiscal, WhatsApp/SMS, pasarela de pagos): TODO confirmar. Propuesta: simularlas detrás de interfaces para no depender de servicios de pago.

## Stack

El mismo de la Práctica 03:

- **Frontend:** React, TypeScript, Vite, Tailwind CSS.
- **Backend:** Go con Fiber.
- **Datos:** PostgreSQL y Redis (caché).
- **Despliegue:** Vercel (frontend), Render (API y Redis) y Neon (PostgreSQL).

## Instalación, ejecución y pruebas

### Backend

Requiere Go 1.26 y Docker. Desde `backend/`:

```bash
cp .env.example .env          # ajustar valores y definir JWT_SECRET (ver abajo)
docker compose up -d postgres redis
go run .                      # API en http://localhost:3001
go test ./...
```

Comprobar el estado de PostgreSQL y Redis:

```bash
curl http://localhost:3001/api/health
```

Responde `200` con `status: ok`, `200` con `status: degraded` si Redis no responde, o `503` si PostgreSQL no responde.

#### Autenticación y roles

`JWT_SECRET` es obligatoria y debe tener al menos 32 caracteres; la API no arranca sin ella. Se genera una propia, por ejemplo con `openssl rand -base64 48`, y se escribe en `.env` (nunca en `.env.example`).

`POST /api/login` recibe `{"email": "...", "password": "..."}` y responde `200` con `{"token", "email", "rol"}`. El token es un JWT (HS256, 2 horas) con el rol en el claim `rol` y se envía en las demás peticiones como `Authorization: Bearer <token>`. Un correo inexistente, una contraseña errónea y una cuenta desactivada responden lo mismo: `401 Credenciales incorrectas`.

Un middleware protege las rutas: `401` si falta el token o es inválido o expirado, y `403` si el rol no tiene permiso. Hay una ruta de prueba por rol que solo acepta a ese rol: `GET /api/prueba/admin-ti`, `/recepcion`, `/odontologo`, `/asistente` y `/auditor`. Con las cuentas de `900_datos_prueba.sql`:

```bash
TOKEN=$(curl -s -X POST http://localhost:3001/api/login -H 'Content-Type: application/json' \n  -d '{"email":"odontologo@sigo-dental.test","password":"<contraseña de 900_datos_prueba.sql>"}' | sed -E 's/.*"token":"([^"]+)".*//')
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/prueba/odontologo   # 200
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/prueba/auditor      # 403
```

Para levantar también la API en contenedor: `docker compose up --build`.

#### Base de datos

PostgreSQL ejecuta los scripts de `backend/database/` en orden alfabético, **solo la primera vez que se crea el volumen**:

- `001_usuarios_roles.sql`: tablas `roles` (con los cinco roles) y `usuarios`.
- `900_datos_prueba.sql`: un usuario por rol con contraseña bcrypt, solo para desarrollo. Las credenciales de prueba están en ese archivo.

Para aplicar un script nuevo en local, o recrear la base desde cero (borra sus datos):

```bash
docker compose down -v
docker compose up -d postgres redis
```

En Neon se aplican a mano los scripts `0xx` (por ejemplo con `psql "$DATABASE_URL" -f database/001_usuarios_roles.sql`); los `9xx` son datos de prueba y no se cargan.

### Frontend

Requiere Node.js 20.19 o superior (o 22.12+), que es lo que pide Vite 8. Desde `frontend/`, con la API del backend en ejecución:

```bash
cp .env.example .env          # VITE_API_URL apunta a la API (http://localhost:3001 por defecto)
npm install
npm run dev                   # http://localhost:5173
npm run build                 # comprobación de tipos y build de producción
npm run lint
```

El servidor de desarrollo usa siempre el puerto 5173 (si está ocupado, falla en lugar de cambiar de puerto), porque es el origen que admite el CORS del backend por defecto. Para probar desde otro equipo de la red, como una tablet, hay que agregar ese origen a `CORS_ALLOWED_ORIGINS` en el `.env` del backend.

El menú tiene dos páginas: Inicio y Estado del sistema (`/estado`), que consulta `GET /api/health` y muestra el estado de PostgreSQL y Redis. La barra lateral es fija desde 768 px de ancho (se puede colapsar a solo iconos) y es un cajón con botón de menú en pantallas más angostas, para cumplir el RNF-07 (escritorio, laptop y tablets de 10,5 pulgadas o más). `vercel.json` redirige todas las rutas a `index.html` para poder recargar páginas del router en Vercel.

## Documentación

- Arquitectura: [docs/architecture.md](docs/architecture.md)
- Dirección visual: [docs/diseno.md](docs/diseno.md)
