# SIGO-Dental

Sistema integral de gestión odontológica y expediente clínico digital: aplicación web responsive para clínicas odontológicas (pacientes, expediente y odontograma, presupuestos, agenda, caja e inventario).

Proyecto de la materia Framework de Programación Web (UPSE). Se construye a partir del Documento de Requerimientos de Software v1.0 (29/10/2026). Trabajo individual.

## Estado

En construcción: existe el esqueleto del backend (`backend/`) con el endpoint `GET /api/health`; el frontend aún no. El proyecto se construye por etapas; cada etapa queda desplegada y funcionando antes de empezar la siguiente.

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
cp .env.example .env          # ajustar valores si es necesario
docker compose up -d postgres redis
go run .                      # API en http://localhost:3001
go test ./...
```

Comprobar el estado de PostgreSQL y Redis:

```bash
curl http://localhost:3001/api/health
```

Responde `200` con `status: ok`, `200` con `status: degraded` si Redis no responde, o `503` si PostgreSQL no responde.

Para levantar también la API en contenedor: `docker compose up --build`.

### Frontend

TODO: se documenta cuando exista el código.

## Documentación

- Arquitectura: [docs/architecture.md](docs/architecture.md)
