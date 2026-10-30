# SIGO-Dental

Sistema integral de gestión odontológica y expediente clínico digital: aplicación web responsive para clínicas odontológicas (pacientes, expediente y odontograma, presupuestos, agenda, caja e inventario).

Proyecto de la materia Framework de Programación Web (UPSE). Se construye a partir del Documento de Requerimientos de Software v1.0 (29/10/2026). Trabajo individual.

## Estado

En construcción: existe el backend (`backend/`) con `GET /api/health`, el modelo de datos de usuarios y roles, y el login con JWT y control de acceso por roles; el frontend (`frontend/`) tiene el layout responsive, el login conectado a la API con rutas y menú según el rol, y una página que muestra `GET /api/health`; aún no tiene módulos clínicos. El proyecto se construye por etapas; cada etapa queda desplegada y funcionando antes de empezar la siguiente.

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
- `002_catalogo.sql`: catálogos de la clínica (sillones, profesionales, procedimientos con precio en centavos, parámetros como moneda, zona horaria y horario de agenda).
- `003_pacientes.sql`: pacientes, número de historia correlativo y anamnesis con alertas médicas.
- `004_agenda.sql`: citas, su historial de estados y la regla RN-03 (un sillón no puede tener dos citas que se solapen, garantizada con una restricción `EXCLUDE` de PostgreSQL).
- `005_notificaciones.sql`: bitácora de avisos de cita (un solo mensaje vigente por tipo).
- `006_facturacion.sql`: cobros, líneas, pagos, anulaciones y el contador de comprobantes sin huecos.
- `900_datos_prueba.sql`: un usuario por rol con contraseña bcrypt, solo para desarrollo. Las credenciales de prueba están en ese archivo.
- `901_`, `902_` y `903_`: catálogo, pacientes y citas de ejemplo, solo para desarrollo.

Para aplicar un script nuevo en una base ya creada (sin borrarla), por ejemplo:

```bash
docker exec -i sigo-dental-db psql -v ON_ERROR_STOP=1 -q -U sigo_dental -d sigo_dental < backend/database/006_facturacion.sql
```

Para recrear la base desde cero (borra sus datos):

```bash
docker compose down -v
docker compose up -d postgres redis
```

En Neon se aplican a mano los scripts `0xx` (por ejemplo con `psql "$DATABASE_URL" -f database/001_usuarios_roles.sql`); los `9xx` son datos de prueba y no se cargan.

#### Módulos de la Fase 1

| Módulo | Quién lo usa | Endpoints principales |
| --- | --- | --- |
| Catálogos (sillones, profesionales, procedimientos, parámetros) | lectura: todos los roles autenticados; escritura: `admin_ti` | `/api/sillones`, `/api/profesionales`, `/api/procedimientos`, `/api/parametros`, `/api/catalogo/:recurso` |
| Pacientes y anamnesis | consulta: recepción, odontólogo y asistente; alta y edición: recepción; anamnesis: personal clínico | `/api/pacientes`, `/api/pacientes/:id/anamnesis` |
| Agenda | consulta: recepción, odontólogo y asistente; crear y reprogramar: recepción; estados: según el rol | `/api/citas`, `/api/citas/:id/estado`, `/api/agenda/horario` |
| Notificaciones | bitácora: recepción; confirmación por enlace: pública | `/api/citas/:id/notificaciones`, `POST /api/citas/confirmar` |
| Cobros | consulta: recepción y auditor; cobrar, registrar pagos y anular: recepción | `/api/cobros`, `/api/cobros/:id/pagos`, `/api/cobros/:id/anular` |

Reglas que conviene conocer: el dinero se guarda en centavos enteros y el servidor calcula siempre precios y totales; nada se borra (los pacientes y catálogos se desactivan y los cobros se anulan con motivo); las horas se guardan en UTC y se muestran en la zona horaria de la clínica (`America/Guayaquil`); el número de comprobante (`C-000001`…) no deja huecos aunque haya cobros simultáneos; un cobro no admite pagos que superen su total.

Los avisos de cita (confirmación inmediata y recordatorios 48 h y 24 h antes) salen por un `Notificador` simulado: se registran en la bitácora pero no se envía nada real hasta tener las credenciales de un proveedor de WhatsApp o SMS. Variables en `backend/.env.example`: `NOTIFICADOR`, `FRONTEND_URL` (arma el enlace de confirmación) y `NOTIFICACIONES_INTERVALO`. El enlace lleva un token firmado con `JWT_SECRET` y deja de valer si la cita se reprograma o ya empezó.

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

Para entrar hay que iniciar sesión en `/login` con una de las cuentas de `backend/database/900_datos_prueba.sql` (una por rol; la contraseña está solo en ese archivo). Sin sesión, cualquier otra ruta redirige al login. El menú muestra solo lo que el rol puede abrir: Inicio para todos, Estado del sistema (`/estado`, consulta `GET /api/health`) solo para `admin_ti`, y un área por rol (`/area/<rol>`) que consulta la ruta de prueba `GET /api/prueba/<rol>` del backend; abrir el área de otro rol muestra "Acceso denegado". La sesión se cierra a los 15 minutos sin actividad (RNF-05), al pulsar "Cerrar sesión" o si la API responde 401, y el token solo se guarda en memoria: recargar la página (F5) también cierra la sesión. La barra lateral es fija desde 768 px de ancho (se puede colapsar a solo iconos) y es un cajón con botón de menú en pantallas más angostas, para cumplir el RNF-07 (escritorio, laptop y tablets de 10,5 pulgadas o más). La interfaz sigue la dirección visual Aurora ([docs/diseno.md](docs/diseno.md)): tema oscuro con paneles de vidrio, un solo acento turquesa y las fuentes Sora y JetBrains Mono, que se sirven desde el propio proyecto (paquetes `@fontsource`). Los colores, radios y tipografías son tokens del bloque `@theme` de `frontend/src/index.css`. `vercel.json` redirige todas las rutas a `index.html` para poder recargar páginas del router en Vercel.

## Documentación

- Arquitectura: [docs/architecture.md](docs/architecture.md)
- Dirección visual: [docs/diseno.md](docs/diseno.md)
