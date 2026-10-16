-- Migración 002: checkout transaccional sobre un volumen que solo tenía el catálogo y la red.
-- Añade el inventario y las tablas de pedidos. Es idempotente (IF NOT EXISTS).
-- Las instalaciones limpias no la necesitan: database/init.sql ya crea todo el esquema.

BEGIN;

-- Inventario simulado que descuenta cada compra; el CHECK impide stock negativo.
ALTER TABLE productos ADD COLUMN IF NOT EXISTS stock INTEGER NOT NULL DEFAULT 50 CHECK (stock >= 0);

-- Correo de la cuenta que representa cada nodo de la red y saldo acumulado de comisiones.
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS email      VARCHAR(150);
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS comisiones NUMERIC(10, 2) NOT NULL DEFAULT 0;

-- Cabecera del pedido guardada dentro de la transacción del checkout.
CREATE TABLE IF NOT EXISTS pedidos (
    id                SERIAL PRIMARY KEY,
    numero            VARCHAR(20)    NOT NULL DEFAULT '',
    cliente_nombre    VARCHAR(150)   NOT NULL,
    cliente_email     VARCHAR(150)   NOT NULL,
    cliente_direccion VARCHAR(200)   NOT NULL DEFAULT '',
    cliente_ciudad    VARCHAR(100)   NOT NULL DEFAULT '',
    cliente_telefono  VARCHAR(30)    NOT NULL DEFAULT '',
    metodo_pago       VARCHAR(50)    NOT NULL,
    total             NUMERIC(10, 2) NOT NULL CHECK (total >= 0),
    estado            VARCHAR(20)    NOT NULL DEFAULT 'pendiente',
    creado_en         TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Detalle del pedido: la clave foránea a productos aborta la transacción si el producto no existe.
CREATE TABLE IF NOT EXISTS pedido_items (
    id              SERIAL PRIMARY KEY,
    pedido_id       INTEGER        NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    producto_id     INTEGER        NOT NULL REFERENCES productos (id),
    nombre          VARCHAR(150)   NOT NULL,
    cantidad        INTEGER        NOT NULL CHECK (cantidad > 0),
    precio_unitario NUMERIC(10, 2) NOT NULL CHECK (precio_unitario >= 0),
    subtotal        NUMERIC(10, 2) NOT NULL CHECK (subtotal >= 0)
);

COMMIT;
