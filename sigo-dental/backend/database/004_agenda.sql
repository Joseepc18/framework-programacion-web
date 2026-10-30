-- Migración 004: agenda de citas (Fase 1, F1-3). Requiere las migraciones 002 y 003.
-- Las fechas se guardan en UTC (timestamptz); la zona horaria de la clínica es un parámetro.

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('zona_horaria', 'America/Guayaquil', 'Zona horaria de la clínica (nombre IANA); define el horario de la agenda.')
ON CONFLICT (clave) DO NOTHING;

-- Permite combinar igualdad (sillon_id) con solapamiento de rangos en la restricción de exclusión.
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS citas (
    id               INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id      INTEGER      NOT NULL REFERENCES pacientes (id),
    profesional_id   INTEGER      NOT NULL REFERENCES profesionales (id),
    sillon_id        INTEGER      NOT NULL REFERENCES sillones (id),
    procedimiento_id INTEGER      REFERENCES procedimientos (id),
    motivo           VARCHAR(200) NOT NULL DEFAULT '',
    notas            VARCHAR(1000) NOT NULL DEFAULT '',
    inicio           TIMESTAMPTZ  NOT NULL,
    fin              TIMESTAMPTZ  NOT NULL,
    estado           VARCHAR(20)  NOT NULL DEFAULT 'programado'
                     CHECK (estado IN ('programado', 'confirmado', 'en_sala', 'en_atencion', 'atendido', 'cancelado', 'no_asistio')),
    creado_por       INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (fin > inicio),
    -- RN-03: un sillón físico no puede tener dos citas a la vez, sean cuales sean los odontólogos.
    -- Las citas canceladas o con inasistencia liberan el sillón; las contiguas (fin = inicio) no chocan
    -- porque el rango [inicio, fin) excluye su extremo final.
    EXCLUDE USING gist (sillon_id WITH =, tstzrange(inicio, fin) WITH &&)
        WHERE (estado NOT IN ('cancelado', 'no_asistio'))
);

CREATE INDEX IF NOT EXISTS idx_citas_inicio ON citas (inicio);
CREATE INDEX IF NOT EXISTS idx_citas_paciente ON citas (paciente_id);
CREATE INDEX IF NOT EXISTS idx_citas_profesional ON citas (profesional_id, inicio);

-- Historial de cambios de estado: quién y cuándo. Nada se borra.
CREATE TABLE IF NOT EXISTS cita_historial (
    id              INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    cita_id         INTEGER      NOT NULL REFERENCES citas (id),
    estado_anterior VARCHAR(20),
    estado_nuevo    VARCHAR(20)  NOT NULL,
    motivo          VARCHAR(300) NOT NULL DEFAULT '',
    usuario_id      INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cita_historial_cita ON cita_historial (cita_id, id);
