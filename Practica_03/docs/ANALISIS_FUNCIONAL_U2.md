# ANÁLISIS FUNCIONAL — MultiCatálogo al cierre de la Unidad 2

## Estado local de Practica_03 — 2026-10-21

**Actualizado tras el cierre de U2: las premisas de seguridad y las brechas de negocio de este análisis ya están resueltas en Practica_03.**
La tabla describe el estado verificado en el código; la referencia del docente se conserva para comparación.

| Premisa | Evidencia local de la revisión estática |
|---|---|
| Autenticación segura | Resuelta: bcrypt, JWT real (2 h, `JWT_SECRET` obligatorio), `RequireAuth`/`RequireAdmin` y corte de cuentas inactivas. |
| Checkout | Resuelto: transacción con `FOR UPDATE`, precios y total recalculados en servidor, identidad tomada del token. |
| Identidad de la red | Resuelta: `usuarios.referido_id` (FK real) y nodo propio en `referidos` al registrarse. |
| Sesión del frontend | Resuelta: se persiste en `localStorage` (`multicatalogo_sesion`) y `apiFetch` envía el JWT. |
| Catálogo y caché | Resueltos: `productos:all` se invalida en el checkout y en cada escritura de productos. |
| Núcleo de negocio | Resuelto: ledger `comisiones`, liquidaciones, reportes, pedidos con estados y paginación. |

Evidencia: [authController.go](../backend/controllers/authController.go), [routes.go](../backend/routes/routes.go), [repository.go](../backend/repository/repository.go) y [AuthContext.tsx](../frontend/src/context/AuthContext.tsx).
Implementadas: registro, administración de usuarios y productos, pedidos, red personal, comisiones auditables, liquidaciones, reportes y paginación.
Verificado con `go build`, `go vet`, `go test` y el stack de Docker (health, login y rutas protegidas). Siguen sin pruebas automatizadas el repositorio SQL y el frontend.
El objetivo sigue siendo cubrir U1/U2; SaaS/multitenant con FastAPI y despliegue corresponden a las unidades posteriores.
El orden local propuesto está en [PLAN_IMPLEMENTACION_U2.md](PLAN_IMPLEMENTACION_U2.md), el avance en [REQUIREMENTS_CHECKLIST.md](REQUIREMENTS_CHECKLIST.md) y la navegación en [PROJECT_MAP.md](PROJECT_MAP.md).


---

## Referencia recibida del docente

> Todo el contenido siguiente se conserva como análisis recibido. Las afirmaciones sobre implementaciones y seguridad describen esa referencia, no un resultado verificado de Practica_03. No se ha inspeccionado el proyecto del docente.

> Informe de brecha funcional: qué le falta al sistema para operar un
> **negocio multinivel real** dentro del alcance ya visto en clase
> (Unidad 1: SPA React · Unidad 2: API Go/Fiber + PostgreSQL + Redis).
>
> El análisis proyecta la arquitectura completa del semestre (U3 SaaS,
> U4 despliegue) pero **solo reporta las funcionalidades implementables
> hasta U2**. Lo que corresponde a U3/U4 se indica como "fuera de alcance".

---

## 1. Resumen ejecutivo

El sistema hoy tiene un **núcleo transaccional sólido y seguro**: login con
bcrypt + JWT real, catálogo cacheado, checkout con transacción ACID y
comisiones asíncronas. Como demostración técnica es excelente.

**Pero como negocio, el sistema está "congelado":** nadie puede crear
productos, nadie puede registrarse, nadie puede consultar un pedido después
de comprarlo, y la red multinivel es una semilla fija que nunca crece.
Las únicas escrituras que existen son `login` (lectura) y `checkout`
(escritura). Falta todo el **back-office** — que es donde vive el 80 % de
una aplicación comercial real.

**Respuestas rápidas a las preguntas típicas:**

| ¿Ya se puede...? | Respuesta | Qué falta |
|---|---|---|
| Ingresar nuevos productos | ❌ No hay `POST /api/productos` ni pantalla admin | CRUD completo + vista admin |
| Registrar nuevos clientes | ❌ No existe endpoint de registro; solo 2 usuarios del seed | `POST /api/register` + formulario |
| Ver pedidos realizados | ❌ El pedido se guarda pero nadie puede consultarlo | `GET /api/pedidos` + historial |
| Dar de alta distribuidores | ❌ El árbol de referidos es seed estático sin vínculo a usuarios | Registro con código de referido |
| Ver reportes de ventas | ❌ Solo KPIs del seed en el Dashboard | Queries de agregación + pantallas |
| Liquidar comisiones | ❌ Las comisiones inflan un contador, no hay trazabilidad ni pago | Ledger de comisiones |

---

## 2. Lo que SÍ existe (base construida U1+U2)

- Catálogo público con búsqueda/filtros, detalle, caché Redis
- Carrito por usuario + checkout con transacción ACID real (stock `FOR UPDATE`, rollback)
- Comisiones asíncronas: goroutine sube hasta 3 niveles por `parent_id`
- Autenticación real: bcrypt + JWT (2h) + middleware `RequireAuth`/`RequireAdmin`
- Vistas: Storefront, Catálogo, Detalle, Carrito, Checkout, Confirmación, Dashboard KPIs, Mi Red
- Persistencia: `usuarios`, `productos`, `referidos`, `pedidos`, `pedido_items` + índices

---

## 3. Funcionalidades faltantes por dominio (alcance U1+U2)

Cada bloque indica: **Backend** (endpoint U2) + **Frontend** (vista U1) +
**Prioridad** de negocio.

### 3.1 Gestión de catálogo (CRUD de productos) — 🔴 CRÍTICO

Un negocio no puede depender de `init.sql` para subir mercadería.

| Capa | Falta implementar |
|---|---|
| Backend | `POST /api/productos` (crear), `PUT /api/productos/:id` (editar nombre/precio/descripción/categoría/stock), `DELETE /api/productos/:id` (o `PATCH .../desactivar` — borrado lógico para no romper `pedido_items` históricos). Todas con `RequireAuth` + **`RequireAdmin`**. Invalidar `productos:all` en cada escritura. |
| Frontend | Ruta admin `/admin/productos`: tabla con paginación + formulario de alta/edición + confirmación de borrado. Validación con los mismos tipos que el modelo. |
| Datos | Columna `activo BOOLEAN DEFAULT true` (borrado lógico). Upload de imágenes queda fuera de alcance (storage/CDN → post-sílabo); hoy basta campo `img` con URL. |

> **Lección de diseño:** escribir en `productos` exige decidir invalidación
> de caché (`DEL productos:all`) — mismo patrón que ya usa el checkout.

### 3.2 Gestión de clientes y usuarios — 🔴 CRÍTICO

Hoy existen exactamente 2 cuentas hardcodeadas. Sin registro no hay negocio.

| Capa | Falta implementar |
|---|---|
| Backend | `POST /api/register` (email + password + nombre → `bcrypt.GenerateFromPassword`, rol `cliente` por defecto). Admin: `GET /api/usuarios` (listado), `PATCH /api/usuarios/:id` (cambiar rol, activar/desactivar). Protegidas con `RequireAdmin`. |
| Frontend | Vista `/registro` pública conectada al endpoint + enlace desde Login. Vista admin `/admin/usuarios` con tabla de cuentas. |
| Datos | `usuarios.activo BOOLEAN`, `usuarios.nombre`, `usuarios.creado_en`. Login debe rechazar cuentas inactivas (`401`/`403`). |

> **Regla de negocio:** el cliente que se registra **debería** poder hacerlo
> con código de referido (ver 3.4) — es la puerta de crecimiento del MLM.

### 3.3 Gestión y consulta de pedidos — 🔴 CRÍTICO

Los pedidos se guardan pero desaparecen: ni el cliente ni el admin pueden verlos.

| Capa | Falta implementar |
|---|---|
| Backend | `GET /api/pedidos/mios` (cliente ve SUS pedidos — filtrar por `cliente_email` del claim JWT). `GET /api/pedidos` (admin, paginado `?page=&limit=`). `GET /api/pedidos/:id` con sus `pedido_items`. `PATCH /api/pedidos/:id/estado` (admin: `pendiente→preparando→enviado→entregado`, `cancelado`). |
| Frontend | `/mis-pedidos` para el cliente (lista + detalle con ítems). `/admin/pedidos` para el admin con cambio de estado. |
| Datos | Columna `pedidos.estado VARCHAR DEFAULT 'pendiente'`, `pedidos.usuario_id` (FK real al comprador — hoy solo guarda email). |

> **Regla de negocio MLM:** cancelar un pedido debe **devolver el stock**
> y revertir las comisiones — la transacción de cancelación es el espejo
> ACID del checkout.

### 3.4 Red multinivel real — 🔴 CRÍTICO (es el corazón del negocio)

La red actual es decorativa: un seed fijo donde el vínculo usuario↔referido
funciona por coincidencia de IDs. Un MLM real necesita que la red **crezca**.

| Capa | Falta implementar |
|---|---|
| Backend | Al registrarse, crear nodo en `referidos` enlazado al `parent_id` del código de referido (`POST /api/register?ref=CODIGO`). `GET /api/red/mia` → árbol **del usuario del token** (subárbol desde su nodo), no el árbol global. Comisiones deben escribir en un **ledger**: tabla `comisiones(pedido_id, referido_id, nivel, porcentaje, monto, estado)` insertada en la misma transacción del pedido — auditable, en vez de solo `UPDATE ventas`. |
| Frontend | Mi link de referido visible en Mi Red (`/registro?ref=...`). Vista de MI red (no la global) con mis comisiones por nivel y por pedido. |
| Datos | `usuarios.referido_id INTEGER REFERENCES referidos(id)` — la FK real que hoy no existe. `referidos.codigo VARCHAR UNIQUE` para el link de invitación. |

> **Advertencia de experto MLM:** sin ledger, las comisiones no son
> defendibles ante auditoría ni ante el distribuidor que reclama su pago.
> `ventas` acumulada es un contador, no un estado de cuenta.

### 3.5 Liquidaciones y pagos a distribuidores — 🟠 ALTO

| Capa | Falta implementar |
|---|---|
| Backend | `GET /api/comisiones/mias` (cliente/distribuidor), `POST /api/liquidaciones` (admin marca comisiones `pendiente→pagada`). Estados: `pendiente`, `aprobada`, `pagada`, `revertida`. |
| Datos | Tabla `liquidaciones(id, referido_id, total, periodo, estado, pagada_en)` + `comisiones.liquidacion_id` nullable. |

### 3.6 Reportes y KPIs reales — 🟠 ALTO

El Dashboard muestra datos del seed, no del negocio.

| Capa | Falta implementar |
|---|---|
| Backend | `GET /api/reportes/ventas?desde&hasta` (total vendido, nº pedidos, ticket promedio), `GET /api/reportes/top-productos`, `GET /api/reportes/top-referidos`, `GET /api/reportes/comisiones` (deuda de comisiones por distribuidor). SQL de agregación (`SUM/GROUP BY`) — buen ejercicio U2-T3. |
| Frontend | Dashboard admin real: tarjetas KPI + tabla de top productos/vendedores. Gráficos opcionales (fuera de alcance estricto). |

### 3.7 Operaciones básicas de API — 🟡 MEDIO

- **Paginación** en `GET /api/productos` y todos los listados (`?page&limit` + `total`) — indispensable cuando el catálogo crece.
- **Validación central** de entrada (hoy cada controlador valida a mano).
- **Búsqueda server-side** (`?q=`) — la búsqueda actual es solo en cliente sobre la lista cacheada.
- `RequireAdmin` aplicado a TODAS las rutas de escritura nuevas (está creado, falta usarlo).
- **Emails transaccionales** (confirmación de pedido, bienvenida): fuera de alcance estricto — requiere SMTP/mailer, se puede simular con un `Mailer` inyectado que hoy haga `log.Println`.

### 3.8 Estado del pedido y cancelación — 🟠 ALTO

Hoy un pedido nace "pagado" de facto. Un flujo real mínimo:

```
pendiente → preparando → enviado → entregado
    └──────► cancelado (devuelve stock + revierte comisiones)
```

---

## 4. Matriz: funcionalidad → dónde encaja en el sílabo

| Funcionalidad | Backend (U2) | Frontend (U1) | Prioridad |
|---|---|---|---|
| CRUD productos | U2-T2/T4 (endpoints + RequireAdmin) | U1-T5 (vista admin) | 🔴 |
| Registro clientes + CRUD usuarios | U2-T4 (bcrypt ya existe) | U1-T5 (vistas) | 🔴 |
| Consulta/gestión pedidos | U2-T2 (endpoints + paginación) | U1-T5 (`/mis-pedidos`, `/admin/pedidos`) | 🔴 |
| Red real: registro con `?ref=` + árbol propio | U2-T4 (identidad por token ya lista) | U1-T5 (Mi Red personal + link) | 🔴 |
| Ledger de comisiones | U2-T1/T4 (misma transacción del pedido) | — | 🔴 |
| Liquidaciones | U2-T4 | U1-T5 | 🟠 |
| Reportes/KPIs | U2-T3 (agregaciones SQL) | U1-T5 (Dashboard) | 🟠 |
| Paginación server-side | U2-T2 | U1-T4 | 🟡 |
| Pasarela de pago real | Post-sílabo | — | ⚪ |
| Multitenant/SaaS | **U3** | — | ⚪ fuera de alcance |
| Docker/proxy/OWASP | **U4** | — | ⚪ fuera de alcance |
| Outbox/worker, refresh tokens, emails reales | Roadmap post-sílabo | — | ⚪ |

---

## 5. Orden de implementación sugerido (dentro de U2)

1. **Registro + CRUD usuarios** — sin clientes nuevos no hay nada que probar.
2. **Vínculo `usuarios.referido_id` + registro con `?ref=`** — la red empieza a crecer de verdad; es lo que convierte esto en MLM.
3. **CRUD productos (admin)** — catálogo administrable.
4. **`GET /api/pedidos` + estados** — el pedido deja de ser un agujero negro; cancelación devuelve stock.
5. **Ledger `comisiones`** — trazabilidad del dinero; luego liquidaciones.
6. **Reportes** — el Dashboard pasa de demo a herramienta de gestión.

Con los pasos 1-5, el sistema ya **vende, referencia y paga comisiones
trazables**: es un MLM mínimo funcional real. Los pasos restantes son
escala (reportes, paginación) y el salto a producción (U3/U4 + roadmap).

---

## 6. Deuda técnica conocida (no bloquea, pero registrar)

- `ventas` en `referidos` seguirá existiendo como contador de lectura rápida;
  el ledger `comisiones` debe ser la **fuente de verdad** y `ventas` un cache.
- Goroutine fire-and-forget sigue siendo el mecanismo de comisiones —
  aceptable en el curso; el roadmap ya prevé Outbox persistente.
- La FK usuario↔referido (3.4) debe hacerse ANTES de poblar más usuarios:
  cambiar el vínculo con datos en producción es más costoso.
