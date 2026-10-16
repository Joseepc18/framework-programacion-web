-- Script de inicialización de la base de datos multicatalogo.
-- PostgreSQL lo ejecuta automáticamente la primera vez que se crea el contenedor
-- (está montado en /docker-entrypoint-initdb.d desde docker-compose.yml).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Tabla de la red multinivel. parent_id apunta a la misma tabla para formar el árbol;
-- la raíz (el usuario autenticado) es la única fila con parent_id NULL.
-- Se crea antes que usuarios porque usuarios.referido_id apunta aquí.
CREATE TABLE IF NOT EXISTS referidos (
    id         SERIAL PRIMARY KEY,
    nombre     VARCHAR(150)   NOT NULL,
    nivel      INTEGER        NOT NULL CHECK (nivel >= 0),
    ventas     NUMERIC(10, 2) NOT NULL DEFAULT 0,
    comisiones NUMERIC(10, 2) NOT NULL DEFAULT 0 CHECK (comisiones >= 0),
    -- Correo de la cuenta real que representa este nodo; es el enlace con usuarios.
    email      VARCHAR(150)   UNIQUE,
    -- Código de invitación del nodo (el link /registro?ref=CODIGO).
    codigo     VARCHAR(30)    UNIQUE,
    parent_id  INTEGER        REFERENCES referidos (id) ON DELETE CASCADE
);

-- Tabla de cuentas del sistema. Contraseña hasheada con bcrypt.
CREATE TABLE IF NOT EXISTS usuarios (
    id          SERIAL PRIMARY KEY,
    email       VARCHAR(150) NOT NULL UNIQUE,
    password    VARCHAR(100) NOT NULL,
    rol         VARCHAR(20)  NOT NULL CHECK (rol IN ('admin', 'cliente')),
    nombre      VARCHAR(150) NOT NULL DEFAULT '',
    activo      BOOLEAN      NOT NULL DEFAULT TRUE,
    creado_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- Nodo de la red multinivel que representa a esta cuenta (FK real).
    referido_id INTEGER      REFERENCES referidos (id)
);

-- Tabla del catálogo. La galería se guarda como un arreglo de URLs (TEXT[]).
CREATE TABLE IF NOT EXISTS productos (
    id          SERIAL PRIMARY KEY,
    nombre      VARCHAR(150)   NOT NULL,
    descripcion TEXT           NOT NULL DEFAULT '',
    precio      NUMERIC(10, 2) NOT NULL CHECK (precio >= 0),
    categoria   VARCHAR(50)    NOT NULL,
    img         TEXT           NOT NULL,
    galeria     TEXT[]         NOT NULL DEFAULT '{}',
    stock       INTEGER        NOT NULL DEFAULT 50 CHECK (stock >= 0),
    -- Borrado lógico: el producto sale del catálogo pero sigue en pedido_items.
    activo      BOOLEAN        NOT NULL DEFAULT TRUE
);

-- Tabla de pedidos realizados
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
    -- Dueño real del pedido (claim del token JWT).
    usuario_id        INTEGER        REFERENCES usuarios (id),
    -- Ciclo de vida: pendiente → preparando → enviado → entregado | cancelado.
    estado            VARCHAR(20)    NOT NULL DEFAULT 'pendiente'
                      CHECK (estado IN ('pendiente', 'preparando', 'enviado', 'entregado', 'cancelado')),
    creado_en         TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Tabla de ítems por cada pedido
CREATE TABLE IF NOT EXISTS pedido_items (
    id              SERIAL PRIMARY KEY,
    pedido_id       INTEGER        NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    producto_id     INTEGER        NOT NULL REFERENCES productos(id),
    nombre          VARCHAR(150)   NOT NULL,
    cantidad        INTEGER        NOT NULL CHECK (cantidad > 0),
    precio_unitario NUMERIC(10, 2) NOT NULL CHECK (precio_unitario >= 0),
    subtotal        NUMERIC(10, 2) NOT NULL CHECK (subtotal >= 0)
);

-- Tabla de liquidaciones: pagos agrupados que el admin hace a un distribuidor.
CREATE TABLE IF NOT EXISTS liquidaciones (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    referido_id INTEGER        NOT NULL REFERENCES referidos (id),
    total       NUMERIC(10, 2) NOT NULL DEFAULT 0 CHECK (total >= 0),
    estado      VARCHAR(20)    NOT NULL DEFAULT 'pendiente'
                CHECK (estado IN ('pendiente', 'pagada')),
    creado_en   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    pagada_en   TIMESTAMPTZ,
    -- Mes que cubre el pago (AAAA-MM), útil para el estado de cuenta del distribuidor.
    periodo     VARCHAR(7)     NOT NULL DEFAULT to_char(NOW(), 'YYYY-MM')
);

-- Registro contable (ledger) de cada comisión generada por un pedido.
-- A diferencia de referidos.ventas (un contador), aquí queda quién cobró,
-- de qué pedido, en qué nivel y en qué estado: es auditable.
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

-- Índices de consultas frecuentes: árbol por parent_id, pedidos por comprador, detalle por pedido y filtro por categoría.
CREATE INDEX IF NOT EXISTS idx_referidos_parent       ON referidos (parent_id);
CREATE INDEX IF NOT EXISTS idx_pedidos_cliente_email  ON pedidos (cliente_email);
CREATE INDEX IF NOT EXISTS idx_pedido_items_pedido    ON pedido_items (pedido_id);
CREATE INDEX IF NOT EXISTS idx_productos_categoria    ON productos (categoria);

-- Datos semilla de la red. Se insertan de arriba hacia abajo para respetar la clave foránea.
-- El admin representa la raíz (nodo 0) y el cliente demo a Luis Poveda (nodo 2), vinculados por email.
INSERT INTO referidos (id, nombre, nivel, ventas, codigo, email, parent_id) VALUES
    (0, 'Tú',            0, 2400, 'TU',     'admin@upse.edu.ec',    NULL),
    (1, 'Ana García',    1, 1200, 'ANA',    NULL,                   0),
    (2, 'Luis Poveda',   1,  850, 'LUIS',   'cliente@upse.edu.ec',  0),
    (3, 'Marta Sánchez', 1,  430, 'MARTA',  NULL,                   0),
    (4, 'Carlos Ruiz',   2,  500, 'CARLOS', NULL,                   1),
    (5, 'Sofía León',    2,  430, 'SOFIA',  NULL,                   1),
    (6, 'Marco Díaz',    2,  380, 'MARCO',  NULL,                   2),
    (7, 'Diana Paz',     3,  300, 'DIANA',  NULL,                   4)
ON CONFLICT (id) DO NOTHING;

-- Datos semilla de cuentas con contraseñas encriptadas mediante bcrypt.
-- referido_id enlaza cada cuenta con su nodo de la red por el email.
INSERT INTO usuarios (id, email, password, rol, nombre, referido_id) VALUES
    (1, 'admin@upse.edu.ec',   crypt('123456', gen_salt('bf', 10)), 'admin',    'Administrador', 0),
    (2, 'cliente@upse.edu.ec', crypt('123456', gen_salt('bf', 10)), 'cliente',  'Cliente Demo',  2)
ON CONFLICT (email) DO NOTHING;

-- Datos semilla del catálogo: los 8 productos que usa el frontend.
INSERT INTO productos (id, nombre, descripcion, precio, categoria, img, galeria) VALUES
    (1, 'Serum Revitalizante',
     'Serum concentrado con vitamina C y ácido hialurónico. Ilumina la piel, reduce manchas y aporta hidratación profunda desde la primera aplicación.',
     45.00, 'Serum', 'https://picsum.photos/seed/serum/600',
     ARRAY['https://picsum.photos/seed/serum1/600', 'https://picsum.photos/seed/serum2/600', 'https://picsum.photos/seed/serum3/600']),
    (2, 'Crema Hidratante Pro',
     'Crema facial de textura ligera con niacinamida y manteca de karité. Hidratación de 24 horas y barrera cutánea fortalecida para todo tipo de piel.',
     32.50, 'Crema', 'https://picsum.photos/seed/crema/600',
     ARRAY['https://picsum.photos/seed/crema1/600', 'https://picsum.photos/seed/crema2/600', 'https://picsum.photos/seed/crema3/600']),
    (3, 'Tónico Purificante',
     'Tónico sin alcohol con hamamelis y agua de rosas. Limpia poros, equilibra el pH y prepara la piel para el resto de la rutina diaria.',
     28.00, 'Tónico', 'https://picsum.photos/seed/tonico/600',
     ARRAY['https://picsum.photos/seed/tonico1/600', 'https://picsum.photos/seed/tonico2/600', 'https://picsum.photos/seed/tonico3/600']),
    (4, 'Mascarilla Nocturna',
     'Mascarilla de noche con colágeno y vitamina E. Repara la piel mientras duermes y recupera la luminosidad al despertar.',
     50.00, 'Mascarilla', 'https://picsum.photos/seed/mascarilla/600',
     ARRAY['https://picsum.photos/seed/mascarilla1/600', 'https://picsum.photos/seed/mascarilla2/600', 'https://picsum.photos/seed/mascarilla3/600']),
    (5, 'Serum Anti-Edad Retinol',
     'Serum de retinol encapsulado que suaviza líneas de expresión y mejora la firmeza. Uso nocturno recomendado.',
     58.00, 'Serum', 'https://picsum.photos/seed/retinol/600',
     ARRAY['https://picsum.photos/seed/retinol1/600', 'https://picsum.photos/seed/retinol2/600', 'https://picsum.photos/seed/retinol3/600']),
    (6, 'Crema Contorno de Ojos',
     'Contorno de ojos con cafeína y péptidos. Reduce bolsas y ojeras, hidrata la zona más delicada del rostro.',
     26.00, 'Crema', 'https://picsum.photos/seed/ojos/600',
     ARRAY['https://picsum.photos/seed/ojos1/600', 'https://picsum.photos/seed/ojos2/600', 'https://picsum.photos/seed/ojos3/600']),
    (7, 'Tónico Exfoliante AHA-BHA',
     'Exfoliación química suave con ácidos AHA y BHA. Renueva la textura de la piel y desobstruye los poros.',
     34.00, 'Tónico', 'https://picsum.photos/seed/exfoliante/600',
     ARRAY['https://picsum.photos/seed/exfoliante1/600', 'https://picsum.photos/seed/exfoliante2/600', 'https://picsum.photos/seed/exfoliante3/600']),
    (8, 'Kit Rutina Completa',
     'Kit integral con serum, crema, tónico y mascarilla. La rutina perfecta para comenzar: limpieza, tratamiento e hidratación.',
     129.00, 'Kit', 'https://picsum.photos/seed/kit/600',
     ARRAY['https://picsum.photos/seed/kit1/600', 'https://picsum.photos/seed/kit2/600', 'https://picsum.photos/seed/kit3/600']);

-- Como los IDs se insertaron a mano, ajustamos las secuencias para que los próximos INSERT no choquen.
SELECT setval('referidos_id_seq', (SELECT MAX(id) FROM referidos));
SELECT setval('usuarios_id_seq',  (SELECT MAX(id) FROM usuarios));
SELECT setval('productos_id_seq', (SELECT MAX(id) FROM productos));
