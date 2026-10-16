-- Migración 003: seguridad e índices sobre un volumen existente.
--  - Contraseñas en bcrypt (pgcrypto): las que sigan en texto plano se hashean una sola vez.
--  - Índices de las consultas frecuentes (árbol, pedidos por comprador, detalle y filtro por categoría).
-- Es idempotente: las contraseñas que ya son bcrypt ($2...) no se tocan.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

UPDATE usuarios
   SET password = crypt(password, gen_salt('bf', 10))
 WHERE password NOT LIKE '$2%';

CREATE INDEX IF NOT EXISTS idx_referidos_parent       ON referidos (parent_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_cliente_email  ON pedidos (cliente_email);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido    ON pedido_items (pedido_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria    ON productos (categoria);

COMMIT;
