# PLAN DE IMPLEMENTACIÓN — Back-office MultiCatálogo (alcance U1+U2)

## Estado local de Practica_03 — 2026-10-16

> **Actualización:** fases 0–7 y administración de usuarios implementadas y verificadas contra PostgreSQL/Redis reales (API en `:3001`). Cierre U2: pruebas con fakes de controladores y middleware, lint 0/0 y auditoría móvil a 375 px. Pendiente: pruebas del repositorio SQL (requieren una base de pruebas).

**Estado local: todas las fases del plan están implementadas en Practica_03** (ver la tabla «Estado de fases» al final).

| Fase local | Estado |
|---|---|
| Previa: seguridad y checkout | Hecha: bcrypt, JWT, autorización, identidad por token, precios en servidor, caché coherente. |
| 0–7 | Hechas: `004_backoffice.sql`/`005_ajustes_u2.sql`, registro, productos, pedidos, ledger y red propia, liquidaciones, reportes, paginación y `?q=`. |

La **administración de usuarios** pedida por [ANALISIS_FUNCIONAL_U2.md](ANALISIS_FUNCIONAL_U2.md) también está hecha (`/api/usuarios`, `/admin/usuarios`).
Orden recomendado: seguridad y checkout → datos → registro → administración de usuarios/productos → pedidos → red y registro contable → liquidaciones → reportes y paginación.
Cada fase se verificó con `curl` y el stack de Docker; la cobertura automatizada cubre controladores, middleware y dominio.
Consultar [REQUIREMENTS_CHECKLIST.md](REQUIREMENTS_CHECKLIST.md) para el estado local y [PROJECT_MAP.md](PROJECT_MAP.md) para las rutas.

---

## Referencia recibida del docente

> Todo el contenido siguiente, incluidas las fases y casillas marcadas, se conserva como material recibido. Sus declaraciones de finalización no certifican la práctica local. No se ha inspeccionado el proyecto del docente.

> Plan checkeable para convertir el sistema actual (U1+U2 completado) en un
> negocio MLM mínimo funcional. Derivado de `ANALISIS_FUNCIONAL_U2.md`.
>
> **Regla:** una fase se marca `[x]` solo cuando su criterio de verificación
> pasa (curl/SQL para backend, flujo en navegador para frontend).
> Al cerrar cada fase se actualizan `REQUIREMENTS_CHECKLIST.md` y `PROJECT_MAP.md`.
>
> **Principios (los mismos del checklist):** seguridad por defecto (toda
> escritura nace con `RequireAuth`/`RequireAdmin`), DI por interfaces,
> mobile-first en vistas nuevas.

---

## FASE 0 — Migración de datos previa

> Antes de escribir una línea de Go: las tablas nuevas y columnas que las
> fases siguientes necesitan. Todo en una migración `004_backoffice.sql`
> idempotente (`IF NOT EXISTS`) aplicable al volumen existente.

- [x] `ALTER usuarios`: `+ nombre VARCHAR`, `+ activo BOOLEAN DEFAULT true`, `+ referido_id INTEGER REFERENCES referidos(id)`, `+ creado_en TIMESTAMP DEFAULT NOW()`
- [x] `ALTER referidos`: `+ codigo VARCHAR UNIQUE` (link de invitación) — generar códigos para el seed (`'TU'`, `'ANA'`, `'LUIS'`...)
- [x] `ALTER productos`: `+ activo BOOLEAN DEFAULT true` (borrado lógico)
- [x] `ALTER pedidos`: `+ estado VARCHAR DEFAULT 'pendiente'`, `+ usuario_id INTEGER REFERENCES usuarios(id)`
- [x] Tabla `comisiones(id, pedido_id, referido_id, nivel, porcentaje, monto, estado DEFAULT 'pendiente', creado_en)` + índices `(referido_id, pedido_id)`
- [x] Tabla `liquidaciones(id, referido_id, total, estado, creado_en, pagada_en)` + `comisiones.liquidacion_id`
- [x] Enlazar el seed: `usuarios.referido_id` → admin→0 ("Tú"), cliente→2 ("Luis Poveda")
- [x] **Verificación:** `\d usuarios` y `\d comisiones` en psql muestran las columnas; `SELECT referido_id FROM usuarios` devuelve 0 y 2

**Archivos:** `database/004_backoffice.sql` (+ `init.sql` actualizado para instalaciones limpias)

---

## FASE 1 — Registro de usuarios + vínculo referido real

> La puerta de entrada del negocio: cualquiera se crea cuenta y, si trae
> código de referido, su nodo cuelga del referente — la red empieza a crecer.

**Backend**

- [x] `POST /api/register` `{nombre, email, password, ref?}`: valida campos, hashea con `bcrypt.GenerateFromPassword`, crea `usuarios` (rol `cliente`) + su nodo en `referidos` **en una transacción**; con `ref` válido → `parent_id` del referente
- [x] `409` si el email ya existe; `400` si el código `ref` no existe
- [x] Login rechaza usuarios `activo=false` (403 "cuenta desactivada")
- [x] `CalularComisiones` arranca desde `usuarios.referido_id` (FK real, ya no por coincidencia de IDs)

**Frontend**

- [x] Vista `/registro` (pública): formulario nombre/email/password/código referido (prefill desde `?ref=` del query string)
- [x] Enlace "Crear cuenta" desde `Login.tsx`; tras registrar → login automático o redirect a `/login`

**Verificación**

- [x] `curl POST /api/register` → `201`; login con esa cuenta → `200` JWT
- [x] `SELECT parent_id FROM referidos WHERE codigo='NUEVO'` → apunta al referente correcto
- [x] Registro con email duplicado → `409`; con `ref` inválido → `400`

---

## FASE 2 — CRUD de productos (admin)

> El catálogo se administra desde la app, no desde el seed. Borrado lógico
> para no romper `pedido_items` históricos.

**Backend**

- [x] `POST /api/productos` (admin): validar nombre/precio/categoría > 0 → `201`
- [x] `PUT /api/productos/:id` (admin): edición parcial de campos
- [x] `DELETE /api/productos/:id` (admin): `activo=false` (no borra la fila)
- [x] `GET /api/productos` público filtra `activo=true`; `GET /api/admin/productos` los incluye inactivos
- [x] `DEL productos:all` tras cada escritura (mismo patrón que checkout)

**Frontend**

- [x] Ruta `/admin/productos` (AdminRoute): tabla (img, nombre, categoría, precio, stock, activo)
- [x] Formulario crear/editar (modal o página) con validación móvil-first
- [x] Toggle activar/desactivar con confirmación

**Verificación**

- [x] `POST` como cliente → `403`; como admin → `201` y aparece en catálogo
- [x] `DELETE` → el producto desaparece del catálogo pero sigue en `pedido_items`
- [x] Clave `productos:all` eliminada de Redis tras el `POST`

---

## FASE 3 — Consulta y gestión de pedidos

> El pedido deja de ser un agujero negro: el cliente ve su historial y el
> admin gestiona el ciclo de vida. Cancelar devuelve stock (espejo ACID).

**Backend**

- [x] `GET /api/pedidos/mios`: pedidos del `usuario_id` del token (paginado `?page&limit`)
- [x] `GET /api/pedidos` (admin): todos, paginado + filtro `?estado=`
- [x] `GET /api/pedidos/:id`: cabecera + `pedido_items` con nombre de producto (JOIN)
- [x] `PATCH /api/pedidos/:id/estado` (admin): transiciones válidas `pendiente→preparando→enviado→entregado` y `cancelado`
- [x] `checkout` guarda `usuario_id` del claim (además del email)
- [x] Cancelación en **transacción**: `estado='cancelado'` + `UPDATE productos SET stock = stock + cantidad` por ítem + `UPDATE comisiones SET estado='revertida'`

**Frontend**

- [x] `/mis-pedidos` (cliente): lista con número, fecha, total, estado + detalle expandible
- [x] `/admin/pedidos` (admin): tabla con filtro por estado + cambio de estado (select/acciones)
- [x] Enlace a `/mis-pedidos` desde navbar/confirmación

**Verificación**

- [x] Cliente A no ve pedidos de cliente B (`/api/pedidos/mios` filtrado por token)
- [x] `PATCH estado=cancelado` → stock restaurado en `productos` + comisiones `revertida`
- [x] Transición inválida (`entregado→pendiente`) → `400`

---

## FASE 4 — Ledger de comisiones + red propia

> El dinero del MLM se vuelve auditable: cada venta escribe quién cobró qué
> porcentaje, y cada usuario ve SU red, no el árbol global del seed.

**Backend**

- [x] `CrearPedido` inserta en `comisiones` (nivel 1=10%, 2=5%, 3=2% sobre el `precio_unitario×cantidad` del pedido) **dentro de la transacción** — el ledger nace con el pedido, la goroutine solo actualiza `ventas` como contador
- [x] `GET /api/red/mia`: subárbol desde el `referido_id` del token (no el global)
- [x] `GET /api/comisiones/mias`: lista del ledger del usuario (pedido, nivel, monto, estado) + total pendiente/pagado
- [x] `domain/comisiones.go`: función pura `RepartirComisiones(total, tasas)` → `[]Comision` sin I/O (la usa el repo; testeable)

**Frontend**

- [x] `MiRed` muestra la red del usuario del token + su link de referido (`/registro?ref=CODIGO`) con botón copiar
- [x] Panel "Mis comisiones": tabla del ledger + totales por estado

**Verificación**

- [x] Compra como cliente → `SELECT * FROM comisiones WHERE pedido_id=?` devuelve fila(s) con monto=nivel×total
- [x] Dos usuarios distintos en `/api/red/mia` ven árboles distintos
- [x] El ledger suma lo mismo que `referidos.ventas` incrementó

---

## FASE 5 — Liquidaciones a distribuidores

> Cerrar el ciclo del dinero: el admin paga lo que el ledger debe.

- [x] `POST /api/liquidaciones` (admin) `{referido_id}`: transacción — suma comisiones `pendiente`, crea `liquidaciones` con `total`, marca comisiones `pagada` + `liquidacion_id`
- [x] `GET /api/liquidaciones` (admin): historial de pagos
- [x] `GET /api/liquidaciones/mias`: el distribuidor ve sus cobros
- [x] Vista admin `/admin/liquidaciones` (deuda por distribuidor + botón liquidar)

**Verificación**

- [x] Liquidar → comisiones pasan a `pagada`, no pueden liquidarse dos veces
- [x] `SUM(comisiones pendiente)` del referido queda en 0 tras liquidar

---

## FASE 6 — Reportes y dashboard real

> El Dashboard deja de mostrar el seed y pasa a reflejar el negocio.

- [x] `GET /api/reportes/resumen?desde&hasta`: total vendido, nº pedidos, ticket promedio, comisiones devengadas
- [x] `GET /api/reportes/top-productos` (SUM cantidad×precio GROUP BY producto)
- [x] `GET /api/reportes/top-referidos` (ventas de la red por referido)
- [x] `GET /api/reportes/comisiones-pendientes` (deuda por distribuidor)
- [x] Dashboard admin consume los endpoints reales (loading + estado vacío)

**Verificación**

- [x] Los KPIs coinciden con `SELECT SUM(total) FROM pedidos` del periodo

---

## FASE 7 — Operaciones de API (dureza final)

> Detalles que separan una demo de un servicio usable con volumen.

- [x] Paginación estándar `{data, page, limit, total}` en productos, pedidos, usuarios
- [x] `GET /api/productos?q=` búsqueda server-side (`ILIKE`)
- [x] Paginación en las vistas de listado del front
- [x] Revisión final: TODAS las rutas de escritura tienen `RequireAuth`; las admin `RequireAdmin`

---

## Estado de fases

| Fase | Entregable | Estado |
|---|---|---|
| 0 | Migración `004_backoffice.sql` aplicada (+ secuencia `referidos_id_seq`) | ✅ 2026-10-16 |
| 1 | Registro + referido real (verificado: `?ref=LUIS` → parent_id=2, 409/400) | ✅ 2026-10-16 |
| 2 | CRUD productos admin (POST/PUT/DELETE-lógico, `?q=` ILIKE, invalida caché) | ✅ 2026-10-16 |
| 3 | Pedidos: `/mios`, admin, detalle dueño-only, estados, cancelar→stock+clawback | ✅ 2026-10-16 |
| 4 | Ledger `comisiones` en la tx + `/red/mia` + `domain.RepartirComisiones` pura | ✅ 2026-10-16 |
| 5 | Liquidaciones `pendiente→pagada` (idempotente, bloqueo FOR UPDATE) | ✅ 2026-10-16 |
| 6 | Reportes SQL reales + Dashboard consumiendo `/api/reportes/*` | ✅ 2026-10-16 |
| 7 | `Pagina{data,page,limit,total}` + `?q=` admin + auditoría de rutas | ✅ 2026-10-16 |

> **Nota F7:** la búsqueda `?q=` se implementó en `/api/admin/productos`
> (donde aporta valor operativo); el catálogo público sigue cacheado completo.

> Al completar 0-4 el sistema ya **vende, referencia y traza comisiones**:
> MLM mínimo funcional. 5-7 lo llevan a operación diaria real.
