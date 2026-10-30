# Arquitectura

## Estructura

Una SPA en React consume una API REST en Go. La API guarda los datos en PostgreSQL y usa Redis como caché de lecturas.

```
React SPA ──HTTPS/JWT──► API Go (Fiber) ──► PostgreSQL
                              └──────────► Redis (caché)
```

Capas del backend (`backend/`), tomadas de la Práctica 03: `routes` → `controllers` → `repository`, con la inyección de dependencias armada en `main.go`. Los controladores dependen de interfaces (`controllers/dependencias.go`), no de implementaciones concretas. `config` abre el pool de PostgreSQL y `cache` el cliente de Redis.

`GET /api/health` informa el estado de las dependencias: `200 ok` si responden ambas, `200 degraded` si Redis no responde (la API puede seguir sirviendo) y `503` si PostgreSQL no responde.

## Decisiones

- **Mismo stack que la Práctica 03.** Es un stack ya conocido y con un despliegue probado, lo que deja el esfuerzo en el dominio odontológico y no en la tecnología.
- **Etapas verticales desplegables.** El documento de requerimientos plantea 12 sprints; se avanza por módulos completos (base, pacientes, agenda, odontograma, presupuestos y caja) para tener siempre algo funcionando en producción.
- **Integraciones externas detrás de interfaces (por confirmar).** Facturación fiscal, mensajería y pagos se simulan en esta fase; cambiar a un servicio real no debe tocar el resto del sistema.

## Pendiente de decidir

- TODO: modelo de datos y migraciones.
- TODO: estrategia de MFA y de auditoría inmutable (RNF-05 y RNF-06).
- TODO: país y normativa aplicables.
