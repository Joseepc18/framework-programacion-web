-- Migración 003: pacientes, anamnesis y condiciones críticas (Fase 1, F1-2: RF-01 y RF-02).
-- Requiere 002_catalogo.sql (convenios y condiciones críticas). Los pacientes no se borran: se desactivan.

-- El número de historia clínica sale de una secuencia; puede haber huecos si se revierte un alta, y no importa.
CREATE SEQUENCE IF NOT EXISTS historia_clinica_seq;

CREATE TABLE IF NOT EXISTS pacientes (
    id                    INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    numero_historia       VARCHAR(20)  NOT NULL UNIQUE
                          DEFAULT ('HC-' || lpad(nextval('historia_clinica_seq')::text, 6, '0')),
    documento             VARCHAR(30)  NOT NULL UNIQUE,
    nombres               VARCHAR(100) NOT NULL,
    apellidos             VARCHAR(100) NOT NULL,
    fecha_nacimiento      DATE         NOT NULL CHECK (fecha_nacimiento >= DATE '1900-01-01'),
    sexo                  VARCHAR(10)  NOT NULL CHECK (sexo IN ('femenino', 'masculino', 'otro')),
    telefono              VARCHAR(30)  NOT NULL,
    email                 VARCHAR(150) NOT NULL DEFAULT '',
    direccion             VARCHAR(200) NOT NULL DEFAULT '',
    emergencia_nombre     VARCHAR(150) NOT NULL,
    emergencia_telefono   VARCHAR(30)  NOT NULL,
    emergencia_parentesco VARCHAR(50)  NOT NULL DEFAULT '',
    -- Tutor legal: el backend lo exige cuando la edad es menor de 18 años.
    tutor_nombre          VARCHAR(150) NOT NULL DEFAULT '',
    tutor_documento       VARCHAR(30)  NOT NULL DEFAULT '',
    tutor_telefono        VARCHAR(30)  NOT NULL DEFAULT '',
    convenio_id           INTEGER REFERENCES convenios (id),
    -- Canal por el que se le envían las notificaciones (F1-4).
    canal_preferido       VARCHAR(10)  NOT NULL DEFAULT 'whatsapp' CHECK (canal_preferido IN ('whatsapp', 'sms', 'email')),
    activo                BOOLEAN      NOT NULL DEFAULT TRUE,
    creado_en             TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pacientes_apellidos ON pacientes (lower(apellidos), lower(nombres));

-- Historial médico general. Cada apartado es obligatorio: si no hay nada que declarar se escribe "Ninguna".
CREATE TABLE IF NOT EXISTS anamnesis (
    paciente_id             INTEGER PRIMARY KEY REFERENCES pacientes (id),
    enfermedades_sistemicas TEXT NOT NULL CHECK (btrim(enfermedades_sistemicas) <> ''),
    alergias_farmacos       TEXT NOT NULL CHECK (btrim(alergias_farmacos) <> ''),
    antecedentes_quirurgicos TEXT NOT NULL CHECK (btrim(antecedentes_quirurgicos) <> ''),
    coagulopatias           TEXT NOT NULL CHECK (btrim(coagulopatias) <> ''),
    actualizado_por         INTEGER REFERENCES usuarios (id),
    actualizado_en          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Condiciones críticas que tiene cada paciente; disparan el badge de advertencia del expediente.
CREATE TABLE IF NOT EXISTS paciente_condiciones_criticas (
    paciente_id  INTEGER NOT NULL REFERENCES pacientes (id),
    condicion_id INTEGER NOT NULL REFERENCES condiciones_criticas (id),
    PRIMARY KEY (paciente_id, condicion_id)
);
