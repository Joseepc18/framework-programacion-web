-- Migración 006: cobros y comprobantes internos (Fase 1, F1-5). Requiere las migraciones 002, 003 y 004.
-- El dinero se guarda siempre en centavos enteros (nunca decimales flotantes). Nada se borra: un cobro se anula.

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('moneda', 'USD', 'Código ISO 4217 de la moneda con que se muestran los importes (los datos se guardan en centavos).')
ON CONFLICT (clave) DO NOTHING;

-- Contadores de numeración. Se incrementan dentro de la misma transacción que crea el documento:
-- el bloqueo de fila serializa las solicitudes simultáneas y, si la transacción falla, el número no se consume.
CREATE TABLE IF NOT EXISTS contadores (
    clave VARCHAR(40) PRIMARY KEY,
    valor BIGINT NOT NULL DEFAULT 0 CHECK (valor >= 0)
);

INSERT INTO contadores (clave, valor) VALUES ('comprobante', 0) ON CONFLICT (clave) DO NOTHING;

CREATE TABLE IF NOT EXISTS cobros (
    id                 INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    numero             VARCHAR(20) NOT NULL UNIQUE,
    paciente_id        INTEGER     NOT NULL REFERENCES pacientes (id),
    cita_id            INTEGER     REFERENCES citas (id),
    subtotal_centavos  INTEGER     NOT NULL CHECK (subtotal_centavos >= 0),
    -- Descuento en puntos básicos (10000 = 100 %); el importe se calcula en el servidor y se guarda ya redondeado.
    descuento_bp       INTEGER     NOT NULL DEFAULT 0 CHECK (descuento_bp BETWEEN 0 AND 10000),
    descuento_centavos INTEGER     NOT NULL DEFAULT 0 CHECK (descuento_centavos >= 0),
    total_centavos     INTEGER     NOT NULL CHECK (total_centavos >= 0),
    creado_por         INTEGER     NOT NULL REFERENCES usuarios (id),
    creado_en          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (total_centavos = subtotal_centavos - descuento_centavos)
);

CREATE INDEX IF NOT EXISTS idx_cobros_paciente ON cobros (paciente_id);

CREATE TABLE IF NOT EXISTS cobro_lineas (
    id                    INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    cobro_id              INTEGER      NOT NULL REFERENCES cobros (id),
    procedimiento_id      INTEGER      NOT NULL REFERENCES procedimientos (id),
    descripcion           VARCHAR(150) NOT NULL,
    cantidad              INTEGER      NOT NULL CHECK (cantidad > 0),
    precio_unitario_centavos INTEGER   NOT NULL CHECK (precio_unitario_centavos >= 0),
    total_centavos        INTEGER      NOT NULL CHECK (total_centavos = cantidad * precio_unitario_centavos)
);

CREATE INDEX IF NOT EXISTS idx_cobro_lineas_cobro ON cobro_lineas (cobro_id);

CREATE TABLE IF NOT EXISTS pagos (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    cobro_id       INTEGER      NOT NULL REFERENCES cobros (id),
    metodo         VARCHAR(20)  NOT NULL CHECK (metodo IN ('efectivo', 'tarjeta', 'transferencia')),
    monto_centavos INTEGER      NOT NULL CHECK (monto_centavos > 0),
    referencia     VARCHAR(100) NOT NULL DEFAULT '',
    creado_por     INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- Tarjeta y transferencia exigen su número de referencia; el efectivo no.
    CHECK (metodo = 'efectivo' OR referencia <> '')
);

CREATE INDEX IF NOT EXISTS idx_pagos_cobro ON pagos (cobro_id);

-- La anulación es un registro nuevo: el cobro y sus pagos quedan intactos y se conserva quién anuló y por qué.
CREATE TABLE IF NOT EXISTS cobro_anulaciones (
    id         INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    cobro_id   INTEGER      NOT NULL UNIQUE REFERENCES cobros (id),
    motivo     VARCHAR(300) NOT NULL CHECK (motivo <> ''),
    usuario_id INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Una cita se cobra una sola vez: puede tener otro cobro solo si el anterior se anuló.
-- (Se valida en la transacción de creación; un índice único no puede mirar la tabla de anulaciones.)
