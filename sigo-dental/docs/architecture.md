# Arquitectura

## Estructura

Una SPA en React consume una API REST en Go. La API guarda los datos en PostgreSQL y usa Redis como caché de lecturas.

```
React SPA ──HTTPS/JWT──► API Go (Fiber) ──► PostgreSQL
                              └──────────► Redis (caché)
```

Capas del backend (`backend/`), tomadas de la Práctica 03: `routes` → `controllers` → `repository`, con la inyección de dependencias armada en `main.go`. Los controladores dependen de interfaces (`controllers/dependencias.go`), no de implementaciones concretas. `config` abre el pool de PostgreSQL y `cache` el cliente de Redis.

`middleware` valida el JWT y los roles, y `models` agrupa los tipos compartidos y el contrato de errores.

Frontend (`frontend/`): React con TypeScript, Vite, Tailwind y React Router. `src/components` tiene el layout (`Layout`, `Sidebar` y `Topbar`) y `RutaProtegida`, `src/pages` las pantallas, `src/context` la sesión, `src/hooks` el cierre por inactividad y `src/config/permisos.ts` los roles con su menú y sus rutas. `src/services` es el acceso a la API: `http.ts` es el cliente HTTP único, lee la URL base de `VITE_API_URL` y agrega el Bearer; los demás servicios dependen de él. El estilo sale de los tokens del bloque `@theme` de `src/index.css` (dirección Aurora, ver `docs/diseno.md`); los componentes no escriben colores sueltos. Las fuentes Sora y JetBrains Mono se autoalojan con `@fontsource` para no depender de un tercero en tiempo de ejecución.

`GET /api/health` informa el estado de las dependencias: `200 ok` si responden ambas, `200 degraded` si Redis no responde (la API puede seguir sirviendo) y `503` si PostgreSQL no responde.

## Decisiones

- **Mismo stack que la Práctica 03.** Es un stack ya conocido y con un despliegue probado, lo que deja el esfuerzo en el dominio odontológico y no en la tecnología.
- **Etapas verticales desplegables.** El documento de requerimientos plantea 12 sprints; se avanza por módulos completos (base, pacientes, agenda, odontograma, presupuestos y caja) para tener siempre algo funcionando en producción.
- **Integraciones externas detrás de interfaces (por confirmar).** Facturación fiscal, mensajería y pagos se simulan en esta fase; cambiar a un servicio real no debe tocar el resto del sistema.

- **Migraciones como SQL numerado y plano.** Los scripts de `backend/database/` se numeran (`001_...`) y PostgreSQL los ejecuta en orden al crear el volumen, igual que en la Práctica 03. No añade dependencias y los mismos archivos se aplican a mano en Neon. Los datos de prueba llevan prefijo `9xx`, así quedan siempre al final y se distinguen de las migraciones que sí van a producción. Limitación: no hay control automático de qué se aplicó; si duele, se pasa a una herramienta como goose o golang-migrate reutilizando los mismos archivos.
- **Roles en una tabla, un rol por usuario.** `roles` guarda los cinco roles del documento con un `codigo` estable (`admin_ti`, `recepcionista`, `odontologo`, `asistente`, `auditor`) y `usuarios.rol_id` apunta a ella. Las contraseñas se guardan solo como hash bcrypt en `usuarios.password_hash`. Los permisos por rol se definen con el RBAC (login y JWT).
- **JWT con el rol dentro y RBAC por middleware.** `POST /api/login` verifica el hash bcrypt y emite un JWT HS256 de 2 horas con el `codigo` del rol. `RequireAuth` valida firma, algoritmo y expiración (`401`) y `RequireRoles(...)` compara el rol con la lista permitida (`403`, y también si falta el rol). Los permisos de cada rol sobre los módulos reales se irán declarando en las rutas con `RequireRoles`, siguiendo el principio de menor privilegio; hoy hay una ruta de prueba por rol (`/api/prueba/...`) que solo acepta al suyo. Un correo inexistente, una contraseña errónea y una cuenta desactivada responden el mismo `401`, y para un correo inexistente se compara contra un hash falso para no delatarlo por el tiempo de respuesta. `JWT_SECRET` es obligatoria y de al menos 32 caracteres: sin ella la API no arranca y no existe una clave de desarrollo por defecto. Limitación conocida: el middleware no consulta la base, así que desactivar una cuenta no invalida los tokens ya emitidos hasta que expiren (2 horas); si hace falta cortar la sesión al instante, se agrega una comprobación de cuenta activa en `RequireAuth`.
- **Token JWT solo en memoria en el frontend.** `localStorage` lo expone a cualquier script (XSS) y sobrevive al cierre del navegador, lo que en una estación compartida de una clínica deja una sesión abierta con datos clínicos; `sessionStorage` sigue siendo legible por un XSS durante toda la pestaña. En memoria no queda copia que robar ni olvidar. El costo es que recargar la página (F5) cierra la sesión, aceptable con un cierre automático a los 15 minutos. Una cookie `httpOnly` sería más robusta, pero exige cambios en el backend (cookies, CSRF y CORS con credenciales) y queda como mejora posible.
- **Cierre de sesión por inactividad y por 401.** RNF-05 pide 15 minutos: un hook compara cada 15 segundos la hora de la última actividad (puntero, teclado, scroll, táctil) y también al volver a la pestaña, para que funcione aunque el equipo se suspenda. Un 401 de cualquier petición enviada con token (vencido o inválido) cierra la sesión y lleva al login con un aviso. Cerrar sesión solo descarta el token en el cliente: no se revoca en el servidor, y sigue valiendo hasta que expire (2 horas).
- **Permisos del frontend solo para mostrar.** `config/permisos.ts` declara qué rol ve cada ruta y de ahí salen el menú, las rutas y la pantalla "Acceso denegado". No es una barrera de seguridad: el backend vuelve a validar el rol en cada petición con `RequireRoles`. Hoy cada rol tiene un área de prueba que consulta su ruta `/api/prueba/...`, y "Estado del sistema" queda solo para `admin_ti`; se reemplazan por los módulos reales.
- **Layout responsive con la barra lateral como cajón por debajo de 768 px.** El RNF-07 pide escritorio (1920x1080), laptop (1366x768) y tablets de 10,5 pulgadas o más, que quedan todas por encima del punto de corte `md` de Tailwind: ahí la barra es fija y se colapsa a iconos. Por debajo se vuelve un cajón, de modo que la aplicación no se rompe en pantallas menores aunque no sean un objetivo del requerimiento.
- **Puerto 5173 fijo en desarrollo.** El CORS del backend solo admite por defecto `localhost:5173` y `127.0.0.1:5173`; con `strictPort` Vite falla si el puerto está ocupado en vez de moverse a uno que el backend bloquearía sin dar un error claro.
- **Sin PWA en el frontend.** La Práctica 03 sí la usaba; aquí no hace falta en esta etapa y se evita el service worker, que complica depurar versiones mientras se construye la base.

## Pendiente de decidir

- TODO: estrategia de MFA y de auditoría inmutable (RNF-05 y RNF-06).
- TODO: país y normativa aplicables.
