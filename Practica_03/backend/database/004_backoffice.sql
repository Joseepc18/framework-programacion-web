-- Migración 004: back-office de la Unidad 2 sobre una base ya poblada.
-- Añade lo que las fases siguientes necesitan en usuarios, red, catálogo,
-- pedidos y el registro contable de comisiones (ledger + liquidaciones).
-- Es idempotente (IF NOT EXISTS / comprobaciones previas): se puede aplicar
-- más de una vez sin perder datos ni retroceder secuencias.
--
-- Aplicación sobre un volumen existente:
--   docker cp database/004_backoffice.sql multicatalogo-db:/tmp/004_backoffice.sql
--   docker exec multicatalogo-db psql -U multicatalogo -d multicatalogo -v ON_ERROR_STOP=1 -f /tmp/004_backoffice.sql
--
-- Las instalaciones limpias NO la necesitan: database/init.sql ya crea el esquema completo.

BEGIN;

-- ---------------------------------------------------------------------------
-- usuarios: nombre propio, cuenta activa, fecha de alta y vínculo con la red.
-- ---------------------------------------------------------------------------
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS nombre     VARCHAR(150) NOT NULL DEFAULT '';
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS activo     BOOLEAN      NOT NULL DEFAULT TRUE;
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS creado_en  TIMESTAMPTZ  NOT NULL DEFAULT NOW();
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS referido_id INTEGER;

-- ---------------------------------------------------------------------------
-- referidos: correo de la cuenta que representa el nodo, contador histórico y
-- código de invitación (el link /registro?ref=CODIGO).
-- ---------------------------------------------------------------------------
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS email      VARCHAR(150);
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS comisiones NUMERIC(10, 2) NOT NULL DEFAULT 0;
ALTER TABLE referidos ADD COLUMN IF NOT EXISTS codigo     VARCHAR(30);

-- ---------------------------------------------------------------------------
-- productos: borrado lógico para no romper pedido_items históricos.
-- ---------------------------------------------------------------------------
ALTER TABLE productos ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;

-- ---------------------------------------------------------------------------
-- pedidos: dueño real (claim del token) y ciclo de vida del envío.
-- ---------------------------------------------------------------------------
ALTER TABLE pedidos ADD COLUMN IF NOT EXISTS usuario_id INTEGER;
ALTER TABLE pedidos ALTER COLUMN estado SET DEFAULT 'pendiente';

-- ---------------------------------------------------------------------------
-- Tabla de liquidaciones: pagos agrupados que el admin hace a un distribuidor.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS liquidaciones (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    referido_id INTEGER        NOT NULL REFERENCES referidos (id),
    total       NUMERIC(10, 2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    estado      VARCHAR(20)    NOT NULL DEFAULT 'pendiente'
                CHECK (estado IN ('pendiente', 'pagada')),
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    pagada_en   TIMESTAMPTZ
);

-- ---------------------------------------------------------------------------
-- Tabla comisiones: registro contable (ledger) de cada comisión por pedido.
-- A diferencia de referidos.ventas (un contador), aquí queda quién cobró,
-- de qué pedido, en qué nivel y en qué estado: es auditable.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS comisiones (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    pedido_id      INTEGER        NOT NULL REFERENCES pedidos (id) ON DELETE CASCADE,
    referido_id    INTEGER        NOT NULL REFERENCES referidos (id),
    nivel          INTEGER        NOT NULL CHECK (nivel BETWEEN 1 AND 3),
    porcentaje     NUMERIC(5, 4)  NOT NULL,
    monto          NUMERIC(10, 2) NOT NULL CHECK (monto >= 0),
    estado         VARCHAR(20)    NOT NULL DEFAULT 'pendiente'
                   CHECK (estado IN ('pendiente', 'aprobada', 'pagada', 'revertida')),
    liquidacion_id INTEGER        REFERENCES liquidaciones (id),
    creado_en      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Índices de las consultas habituales: deuda por distribuidor y detalle por pedido.
CREATE INDEX IF NOT EXISTS idx_comisiones_referido_pedido ON comisiones (referido_id, pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_pedido          ON comisiones (pedido_id);
CREATE INDEX IF NOT EXISTS idx_comisiones_liquidacion     ON comisiones (liquidacion_id);

-- ---------------------------------------------------------------------------
-- Claves foráneas nuevas: se agregan solo si todavía no existen.
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'usuarios'::regclass
          AND contype  = 'f'
          AND conname  = 'usuarios_referido_id_fkey'
    ) THEN
        ALTER TABLE usuarios
            ADD CONSTRAINT usuarios_referido_id_fkey
            FOREIGN KEY (referido_id) REFERENCES referidos (id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'pedidos'::regclass
          AND contype  = 'f'
          AND conname  = 'pedidos_usuario_id_fkey'
    ) THEN
        ALTER TABLE pedidos
            ADD CONSTRAINT pedidos_usuario_id_fkey
            FOREIGN KEY (usuario_id) REFERENCES usuarios (id);
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Unicidad de código de invitación y de correo por nodo de la red.
-- Un índice único admite varios NULL, así que los nodos sin código siguen
-- siendo válidos; solo chocan dos códigos concretos iguales.
-- ---------------------------------------------------------------------------
CREATE UNIQUE INDEX IF NOT EXISTS referidos_codigo_key ON referidos (codigo);
CREATE UNIQUE INDEX IF NOT EXISTS referidos_email_key  ON referidos (email);

-- ---------------------------------------------------------------------------
-- Ciclo de vida del pedido: reemplaza al estado 'confirmado' con el que
-- nacían todos los pedidos. Se migran las filas existentes antes de imponer
-- la restricción para no dejar estados huérfanos.
-- ---------------------------------------------------------------------------
UPDATE pedidos SET estado = 'pendiente' WHERE estado = 'confirmado';

ALTER TABLE pedidos DROP CONSTRAINT IF EXISTS pedidos_estado_check;
ALTER TABLE pedidos ADD CONSTRAINT pedidos_estado_check
    CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado'));

-- ---------------------------------------------------------------------------
-- Semilla de la red: código de invitación de cada nodo.
-- Solo se completa donde todavía no tiene código (sigue siendo idempotente).
-- ---------------------------------------------------------------------------
UPDATE referidos SET codigo = 'TU'     WHERE codigo IS NULL AND id = 0;
UPDATE referidos SET codigo = 'ANA'    WHERE codigo IS NULL AND id = 1;
UPDATE referidos SET codigo = 'LUIS'   WHERE codigo IS NULL AND id = 2;
UPDATE referidos SET codigo = 'MARTA'  WHERE codigo IS NULL AND id = 3;
UPDATE referidos SET codigo = 'CARLOS' WHERE codigo IS NULL AND id = 4;
UPDATE referidos SET codigo = 'SOFIA'  WHERE codigo IS NULL AND id = 5;
UPDATE referidos SET codigo = 'MARCO'  WHERE codigo IS NULL AND id = 6;
UPDATE referidos SET codigo = 'DIANA'  WHERE codigo IS NULL AND id = 7;

-- ---------------------------------------------------------------------------
-- Enlace de las cuentas con su nodo de la red usando el correo como fuente de
-- verdad (relación 1:1 garantizada por los dos índices únicos): así no se
-- asume correspondencia numérica entre usuarios.id y referidos.id.
-- COALESCE conserva un vínculo ya personalizado si se vuelve a ejecutar.
-- ---------------------------------------------------------------------------
UPDATE usuarios u
   SET referido_id = COALESCE(u.referido_id, r.id)
  FROM referidos r
 WHERE r.email = u.email
   AND u.referido_id IS NULL;

-- Nombre visible para las cuentas semilla que todavía no tienen uno.
UPDATE usuarios SET nombre = 'Administrador'
 WHERE email = 'admin@upse.edu.ec' AND nombre = '';

UPDATE usuarios SET nombre = 'Cliente Demo'
 WHERE email = 'cliente@upse.edu.ec' AND nombre = '';

-- Nodos que todavía no tienen correo toman el de la cuenta que representan
-- (solo si ese correo no está ya en uso en otro nodo).
UPDATE referidos r
   SET email = u.email
  FROM usuarios u
 WHERE u.referido_id = r.id
   AND r.email IS NULL
   AND NOT EXISTS (SELECT 1 FROM referidos x WHERE x.email = u.email);

-- ---------------------------------------------------------------------------
-- Secuencias: los IDs de la semilla se insertaron a mano, así que hay que
-- adelantar los contadores. GREATEST evita retroceder la secuencia si la
-- migración se ejecuta de nuevo cuando ya hay filas nuevas.
-- ---------------------------------------------------------------------------
SELECT setval('referidos_id_seq',
      GREATEST((SELECT COALESCE(MAX(id), 0) FROM referidos),
               (SELECT last_value FROM referidos_id_seq)));

SELECT setval('usuarios_id_seq',
      GREATEST((SELECT COALESCE(MAX(id), 0) FROM usuarios),
               (SELECT last_value FROM usuarios_id_seq)));

SELECT setval('productos_id_seq',
      GREATEST((SELECT COALESCE(MAX(id), 0) FROM productos),
               (SELECT last_value FROM productos_id_seq)));

COMMIT;
