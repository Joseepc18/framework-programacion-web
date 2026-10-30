-- Migración 011: consentimiento informado digital y códigos OTP (Fase 2, F2-5: RF-06). Requiere 002, 003, 005, 007 y 010.
-- Un consentimiento firmado es inmutable; una revocación es un documento nuevo. Nada se borra.

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('consentimiento_vigencia_dias', '180', 'Días de vigencia de un consentimiento firmado para iniciar un procedimiento que lo exige.')
ON CONFLICT (clave) DO NOTHING;

-- Códigos de un solo uso (OTP) para firmar consentimientos y para el ingreso al portal del paciente. Solo se guarda su
-- hash (HMAC con la clave del servidor). Vencen a los 10 minutos y admiten 3 intentos.
CREATE TABLE IF NOT EXISTS codigos_otp (
    id            INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    proposito     VARCHAR(20) NOT NULL CHECK (proposito IN ('consentimiento', 'portal')),
    referencia_id INTEGER     NOT NULL,
    codigo_hash   CHAR(64)    NOT NULL,
    expira_en     TIMESTAMPTZ NOT NULL,
    intentos      SMALLINT    NOT NULL DEFAULT 0 CHECK (intentos BETWEEN 0 AND 3),
    usado_en      TIMESTAMPTZ,
    creado_en     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_codigos_otp_ref ON codigos_otp (proposito, referencia_id, id DESC);

-- La bitácora de notificaciones también registra los códigos enviados al paciente, que no pertenecen a una cita.
ALTER TABLE notificaciones ALTER COLUMN cita_id DROP NOT NULL;
ALTER TABLE notificaciones ADD COLUMN IF NOT EXISTS paciente_id INTEGER REFERENCES pacientes (id);
ALTER TABLE notificaciones DROP CONSTRAINT IF EXISTS notificaciones_tipo_check;
ALTER TABLE notificaciones ADD CONSTRAINT notificaciones_tipo_check
    CHECK (tipo IN ('confirmacion', 'recordatorio_48h', 'recordatorio_24h', 'otp_consentimiento', 'otp_portal'));
ALTER TABLE notificaciones DROP CONSTRAINT IF EXISTS notificaciones_destino_check;
ALTER TABLE notificaciones ADD CONSTRAINT notificaciones_destino_check CHECK (cita_id IS NOT NULL OR paciente_id IS NOT NULL);
CREATE INDEX IF NOT EXISTS idx_notificaciones_paciente ON notificaciones (paciente_id, id DESC);

-- Formatos legales por procedimiento. Los campos {{paciente}}, {{documento}}, {{procedimiento}}, {{pieza}},
-- {{profesional}} y {{fecha}} se completan solos. Sin procedimiento (NULL) es el formato general de respaldo.
CREATE TABLE IF NOT EXISTS consentimiento_plantillas (
    id               INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    procedimiento_id INTEGER      REFERENCES procedimientos (id),
    titulo           VARCHAR(150) NOT NULL CHECK (btrim(titulo) <> ''),
    cuerpo           TEXT         NOT NULL CHECK (btrim(cuerpo) <> ''),
    activo           BOOLEAN      NOT NULL DEFAULT TRUE,
    creado_en        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Una sola plantilla activa por procedimiento (y una general).
CREATE UNIQUE INDEX IF NOT EXISTS uq_plantilla_activa ON consentimiento_plantillas (COALESCE(procedimiento_id, 0)) WHERE activo;

INSERT INTO consentimiento_plantillas (procedimiento_id, titulo, cuerpo)
SELECT NULL, 'Consentimiento informado para procedimiento odontológico',
       E'Yo, {{paciente}}, con documento de identidad {{documento}}, declaro que el/la Dr./Dra. {{profesional}} me ha explicado, en un lenguaje que comprendo, el procedimiento «{{procedimiento}}» {{pieza}}, su finalidad, los beneficios esperados, los riesgos y molestias posibles, y las alternativas de tratamiento, incluida la opción de no realizarlo.\n\nHe podido hacer preguntas y fueron respondidas a mi satisfacción. Entiendo que puedo revocar este consentimiento en cualquier momento antes de que se realice el procedimiento.\n\nAutorizo de manera libre y voluntaria que se realice el procedimiento indicado.\n\nFecha: {{fecha}}.'
WHERE NOT EXISTS (SELECT 1 FROM consentimiento_plantillas WHERE procedimiento_id IS NULL AND activo);

-- Consentimiento (o su revocación) de un paciente para un procedimiento. Nace "pendiente" con el texto ya completado
-- (se guarda tal cual se mostró, con su huella) y pasa a "firmado" una sola vez, con la firma manuscrita o el código OTP.
CREATE TABLE IF NOT EXISTS consentimientos (
    id                 INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id        INTEGER      NOT NULL REFERENCES pacientes (id),
    procedimiento_id   INTEGER      NOT NULL REFERENCES procedimientos (id),
    profesional_id     INTEGER      NOT NULL REFERENCES profesionales (id),
    plantilla_id       INTEGER      REFERENCES consentimiento_plantillas (id),
    tipo               VARCHAR(14)  NOT NULL DEFAULT 'consentimiento' CHECK (tipo IN ('consentimiento', 'revocacion')),
    revoca_a           INTEGER      REFERENCES consentimientos (id),
    motivo             VARCHAR(300) NOT NULL DEFAULT '',
    titulo             VARCHAR(150) NOT NULL,
    texto              TEXT         NOT NULL,
    huella_texto       CHAR(64)     NOT NULL,
    estado             VARCHAR(10)  NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'firmado')),
    firmante_tipo      VARCHAR(10)  NOT NULL CHECK (firmante_tipo IN ('paciente', 'tutor')),
    firmante_nombre    VARCHAR(150) NOT NULL,
    firmante_documento VARCHAR(30)  NOT NULL,
    modalidad          VARCHAR(10)  CHECK (modalidad IN ('manuscrita', 'otp')),
    trazo              JSONB,
    dispositivo        VARCHAR(400) NOT NULL DEFAULT '',
    firmado_en         TIMESTAMPTZ,
    archivo_id         INTEGER      REFERENCES archivos_paciente (id),
    creado_por         INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK ((estado = 'firmado') = (firmado_en IS NOT NULL AND modalidad IS NOT NULL AND archivo_id IS NOT NULL)),
    CHECK ((tipo = 'revocacion') = (revoca_a IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_consentimientos_paciente ON consentimientos (paciente_id, creado_en DESC);
-- Una revocación firmada por consentimiento basta; los pendientes pueden repetirse.
CREATE INDEX IF NOT EXISTS idx_consentimientos_revoca ON consentimientos (revoca_a) WHERE revoca_a IS NOT NULL;

CREATE OR REPLACE FUNCTION proteger_consentimiento() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'consentimientos es un registro inmutable: no admite DELETE' USING ERRCODE = 'SG001';
    END IF;
    IF OLD.estado = 'firmado' THEN
        RAISE EXCEPTION 'un consentimiento firmado no se modifica' USING ERRCODE = 'SG001';
    END IF;
    IF NEW.paciente_id <> OLD.paciente_id OR NEW.procedimiento_id <> OLD.procedimiento_id OR NEW.texto <> OLD.texto
       OR NEW.huella_texto <> OLD.huella_texto OR NEW.tipo <> OLD.tipo OR NEW.revoca_a IS DISTINCT FROM OLD.revoca_a
       OR NEW.firmante_nombre <> OLD.firmante_nombre OR NEW.firmante_documento <> OLD.firmante_documento THEN
        RAISE EXCEPTION 'el texto y los datos de un consentimiento no cambian: solo se firma' USING ERRCODE = 'SG001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_consentimiento_protegido ON consentimientos;
CREATE TRIGGER trg_consentimiento_protegido BEFORE UPDATE OR DELETE ON consentimientos
    FOR EACH ROW EXECUTE FUNCTION proteger_consentimiento();
