-- Migración 002: catálogo médico (Fase 1, F1-1) y parámetros del sistema.
-- Nada del catálogo se borra: se desactiva (activo = FALSE), porque otros registros lo referencian.

-- Parámetros configurables por el administrador TI. valor siempre es texto; cada módulo lo interpreta.
CREATE TABLE IF NOT EXISTS parametros (
    clave       VARCHAR(60)  PRIMARY KEY,
    valor       VARCHAR(200) NOT NULL,
    descripcion TEXT         NOT NULL DEFAULT ''
);

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('agenda_hora_inicio', '07:00', 'Hora local en que abre la agenda (HH:MM).'),
    ('agenda_hora_fin',    '21:00', 'Hora local en que cierra la agenda (HH:MM).')
ON CONFLICT (clave) DO NOTHING;

CREATE TABLE IF NOT EXISTS consultorios (
    id     INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    activo BOOLEAN      NOT NULL DEFAULT TRUE
);

-- Sillón dental físico. Su nombre solo debe ser único dentro del consultorio.
CREATE TABLE IF NOT EXISTS sillones (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    consultorio_id INTEGER      NOT NULL REFERENCES consultorios (id),
    nombre         VARCHAR(100) NOT NULL,
    activo         BOOLEAN      NOT NULL DEFAULT TRUE,
    UNIQUE (consultorio_id, nombre)
);

CREATE INDEX IF NOT EXISTS idx_sillones_consultorio ON sillones (consultorio_id);

-- Profesional de la clínica: la cuenta de un usuario (rol odontólogo) más sus datos de ejercicio.
CREATE TABLE IF NOT EXISTS profesionales (
    id                   INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    usuario_id           INTEGER      NOT NULL UNIQUE REFERENCES usuarios (id),
    especialidad         VARCHAR(100) NOT NULL DEFAULT 'Odontología general',
    registro_profesional VARCHAR(50)  NOT NULL UNIQUE,
    activo               BOOLEAN      NOT NULL DEFAULT TRUE
);

-- Procedimientos con precio en centavos (nunca decimales flotantes) y duración estimada para la agenda.
CREATE TABLE IF NOT EXISTS procedimientos (
    id                      INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    codigo                  VARCHAR(20)  NOT NULL UNIQUE,
    nombre                  VARCHAR(150) NOT NULL,
    especialidad            VARCHAR(100) NOT NULL DEFAULT 'Odontología general',
    precio_centavos         INTEGER      NOT NULL CHECK (precio_centavos >= 0),
    duracion_min            INTEGER      NOT NULL CHECK (duracion_min > 0),
    requiere_consentimiento BOOLEAN      NOT NULL DEFAULT FALSE,
    activo                  BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS convenios (
    id     INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    activo BOOLEAN      NOT NULL DEFAULT TRUE
);

-- Condiciones médicas que disparan el badge de advertencia en el expediente (RF-02). Es configurable.
CREATE TABLE IF NOT EXISTS condiciones_criticas (
    id     INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    activo BOOLEAN      NOT NULL DEFAULT TRUE
);

INSERT INTO condiciones_criticas (nombre) VALUES
    ('Alergia a penicilina'),
    ('Anticoagulado'),
    ('Coagulopatía'),
    ('Cardiopatía'),
    ('Diabetes'),
    ('Embarazo')
ON CONFLICT (nombre) DO NOTHING;
