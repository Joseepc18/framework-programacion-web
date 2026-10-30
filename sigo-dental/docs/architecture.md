# Arquitectura

## Estructura

Una SPA en React consume una API REST en Go. La API guarda los datos en PostgreSQL y usa Redis como caché de lecturas.

```
React SPA ──HTTPS/JWT──► API Go (Fiber) ──► PostgreSQL
                              └──────────► Redis (caché)
```

Capas del backend (`backend/`), tomadas de la Práctica 03: `routes` → `controllers` → `repository`, con la inyección de dependencias armada en `main.go`. Los controladores dependen de interfaces (`controllers/dependencias.go`), no de implementaciones concretas. `config` abre el pool de PostgreSQL y `cache` el cliente de Redis.

`middleware` valida el JWT y los roles, y `models` agrupa los tipos compartidos y el contrato de errores.

`GET /api/health` informa el estado de las dependencias: `200 ok` si responden ambas, `200 degraded` si Redis no responde (la API puede seguir sirviendo) y `503` si PostgreSQL no responde.

## Decisiones

- **Mismo stack que la Práctica 03.** Es un stack ya conocido y con un despliegue probado, lo que deja el esfuerzo en el dominio odontológico y no en la tecnología.
- **Etapas verticales desplegables.** El documento de requerimientos plantea 12 sprints; se avanza por módulos completos (base, pacientes, agenda, odontograma, presupuestos y caja) para tener siempre algo funcionando en producción.
- **Integraciones externas detrás de interfaces (por confirmar).** Facturación fiscal, mensajería y pagos se simulan en esta fase; cambiar a un servicio real no debe tocar el resto del sistema.

- **Migraciones como SQL numerado y plano.** Los scripts de `backend/database/` se numeran (`001_...`) y PostgreSQL los ejecuta en orden al crear el volumen, igual que en la Práctica 03. No añade dependencias y los mismos archivos se aplican a mano en Neon. Los datos de prueba llevan prefijo `9xx`, así quedan siempre al final y se distinguen de las migraciones que sí van a producción. Limitación: no hay control automático de qué se aplicó; si duele, se pasa a una herramienta como goose o golang-migrate reutilizando los mismos archivos.
- **Roles en una tabla, un rol por usuario.** `roles` guarda los cinco roles del documento con un `codigo` estable (`admin_ti`, `recepcionista`, `odontologo`, `asistente`, `auditor`) y `usuarios.rol_id` apunta a ella. Las contraseñas se guardan solo como hash bcrypt en `usuarios.password_hash`. Los permisos por rol se definen con el RBAC (login y JWT).
- **JWT con el rol dentro y RBAC por middleware.** `POST /api/login` verifica el hash bcrypt y emite un JWT HS256 de 2 horas con el `codigo` del rol. `RequireAuth` valida firma, algoritmo y expiración (`401`) y `RequireRoles(...)` compara el rol con la lista permitida (`403`, y también si falta el rol). Los permisos de cada rol sobre los módulos reales se irán declarando en las rutas con `RequireRoles`, siguiendo el principio de menor privilegio; hoy hay una ruta de prueba por rol (`/api/prueba/...`) que solo acepta al suyo. Un correo inexistente, una contraseña errónea y una cuenta desactivada responden el mismo `401`, y para un correo inexistente se compara contra un hash falso para no delatarlo por el tiempo de respuesta. `JWT_SECRET` es obligatoria y de al menos 32 caracteres: sin ella la API no arranca y no existe una clave de desarrollo por defecto. Limitación conocida: el middleware no consulta la base, así que desactivar una cuenta no invalida los tokens ya emitidos hasta que expiren (2 horas); si hace falta cortar la sesión al instante, se agrega una comprobación de cuenta activa en `RequireAuth`.

## Pendiente de decidir

- TODO: estrategia de MFA y de auditoría inmutable (RNF-05 y RNF-06).
- TODO: país y normativa aplicables.
