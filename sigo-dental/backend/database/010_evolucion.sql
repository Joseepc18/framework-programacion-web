-- Migración 010: evolución SOEP, adendas, firma digital y procedimientos realizados (Fase 2, F2-4: RF-05 y RN-01).
-- Requiere las migraciones 002, 003, 004, 007 y 009. Nada se borra: los disparadores lo impiden.

-- Firma digital (nivel de firma electrónica simple): reautenticación con la contraseña, huella SHA-256 del contenido
-- canónico, identidad del firmante y fecha. La firma no se modifica ni se borra. Es el servicio común de notas,
-- adendas, consentimientos y recetas.
CREATE TABLE IF NOT EXISTS firmas (
    id                   INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    entidad              VARCHAR(20)  NOT NULL CHECK (entidad IN ('nota', 'adenda', 'receta', 'consentimiento')),
    entidad_id           INTEGER      NOT NULL,
    usuario_id           INTEGER      NOT NULL REFERENCES usuarios (id),
    profesional_id       INTEGER      NOT NULL REFERENCES profesionales (id),
    nombre_firmante      VARCHAR(150) NOT NULL,
    registro_profesional VARCHAR(50)  NOT NULL,
    huella_sha256        CHAR(64)     NOT NULL,
    contenido_canonico   TEXT         NOT NULL,
    firmado_en           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (entidad, entidad_id)
);

DROP TRIGGER IF EXISTS trg_firmas_inmutables ON firmas;
CREATE TRIGGER trg_firmas_inmutables BEFORE UPDATE OR DELETE ON firmas
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();

-- Nota de evolución SOEP (Subjetivo, Objetivo, Evaluación, Plan). Editable hasta que vence el plazo (parámetro
-- bloqueo_nota_horas, lo hace cumplir la API con su reloj) o hasta que se firma (lo hace cumplir también la base).
CREATE TABLE IF NOT EXISTS notas_clinicas (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id    INTEGER     NOT NULL REFERENCES pacientes (id),
    cita_id        INTEGER     REFERENCES citas (id),
    -- Profesional responsable: el único que puede firmarla (el asistente redacta, no firma).
    profesional_id INTEGER     NOT NULL REFERENCES profesionales (id),
    autor_id       INTEGER     NOT NULL REFERENCES usuarios (id),
    subjetivo      TEXT        NOT NULL CHECK (btrim(subjetivo) <> ''),
    objetivo       TEXT        NOT NULL CHECK (btrim(objetivo) <> ''),
    evaluacion     TEXT        NOT NULL CHECK (btrim(evaluacion) <> ''),
    plan           TEXT        NOT NULL CHECK (btrim(plan) <> ''),
    creada_en      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actualizada_en TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notas_paciente ON notas_clinicas (paciente_id, creada_en DESC);

CREATE OR REPLACE FUNCTION proteger_nota() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'notas_clinicas es un registro inmutable: no admite DELETE' USING ERRCODE = 'SG001';
    END IF;
    IF EXISTS (SELECT 1 FROM firmas WHERE entidad = 'nota' AND entidad_id = OLD.id) THEN
        RAISE EXCEPTION 'una nota firmada no se modifica: corrígela con una adenda (RN-01)' USING ERRCODE = 'SG001';
    END IF;
    IF NEW.paciente_id <> OLD.paciente_id OR NEW.profesional_id <> OLD.profesional_id OR NEW.autor_id <> OLD.autor_id
       OR NEW.creada_en <> OLD.creada_en THEN
        RAISE EXCEPTION 'la nota no cambia de paciente, profesional ni autor' USING ERRCODE = 'SG001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_nota_protegida ON notas_clinicas;
CREATE TRIGGER trg_nota_protegida BEFORE UPDATE OR DELETE ON notas_clinicas
    FOR EACH ROW EXECUTE FUNCTION proteger_nota();

-- Procedimientos realizados al paciente. Se inicia "en curso" (para lo que exige consentimiento se valida antes) y se
-- completa como "realizado": desde entonces no se borra ni se cambia, y la base lo impide (RN-01). Solo uno en curso
-- se puede cancelar.
CREATE TABLE IF NOT EXISTS procedimientos_realizados (
    id               INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id      INTEGER      NOT NULL REFERENCES pacientes (id),
    procedimiento_id INTEGER      NOT NULL REFERENCES procedimientos (id),
    pieza            SMALLINT     CHECK (pieza / 10 BETWEEN 1 AND 8 AND pieza % 10 BETWEEN 1 AND 8),
    superficie       VARCHAR(10)  CHECK (superficie IN ('mesial', 'distal', 'oclusal', 'vestibular', 'lingual')),
    profesional_id   INTEGER      NOT NULL REFERENCES profesionales (id),
    cita_id          INTEGER      REFERENCES citas (id),
    estado           VARCHAR(10)  NOT NULL DEFAULT 'en_curso' CHECK (estado IN ('en_curso', 'realizado', 'cancelado')),
    notas            VARCHAR(300) NOT NULL DEFAULT '',
    registrado_por   INTEGER      NOT NULL REFERENCES usuarios (id),
    iniciado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    realizado_en     TIMESTAMPTZ,
    CHECK ((estado = 'realizado') = (realizado_en IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_proc_realizados_paciente ON procedimientos_realizados (paciente_id, iniciado_en DESC);

CREATE OR REPLACE FUNCTION proteger_procedimiento_realizado() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'procedimientos_realizados es un registro inmutable: no admite DELETE (RN-01)' USING ERRCODE = 'SG001';
    END IF;
    IF OLD.estado <> 'en_curso' THEN
        RAISE EXCEPTION 'un procedimiento realizado o cancelado no se modifica: corrígelo con una adenda (RN-01)' USING ERRCODE = 'SG001';
    END IF;
    IF NEW.paciente_id <> OLD.paciente_id OR NEW.procedimiento_id <> OLD.procedimiento_id OR NEW.pieza IS DISTINCT FROM OLD.pieza
       OR NEW.superficie IS DISTINCT FROM OLD.superficie OR NEW.profesional_id <> OLD.profesional_id
       OR NEW.registrado_por <> OLD.registrado_por OR NEW.iniciado_en <> OLD.iniciado_en THEN
        RAISE EXCEPTION 'el procedimiento solo cambia de estado' USING ERRCODE = 'SG001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_procedimiento_protegido ON procedimientos_realizados;
CREATE TRIGGER trg_procedimiento_protegido BEFORE UPDATE OR DELETE ON procedimientos_realizados
    FOR EACH ROW EXECUTE FUNCTION proteger_procedimiento_realizado();

-- Adenda: la única forma de corregir una nota firmada o un procedimiento realizado. Es una nota nueva, enlazada a la
-- original y firmada por el profesional responsable al crearla.
CREATE TABLE IF NOT EXISTS adendas (
    id                        INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id               INTEGER     NOT NULL REFERENCES pacientes (id),
    nota_id                   INTEGER     REFERENCES notas_clinicas (id),
    procedimiento_realizado_id INTEGER    REFERENCES procedimientos_realizados (id),
    profesional_id            INTEGER     NOT NULL REFERENCES profesionales (id),
    autor_id                  INTEGER     NOT NULL REFERENCES usuarios (id),
    texto                     TEXT        NOT NULL CHECK (btrim(texto) <> ''),
    creada_en                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((nota_id IS NOT NULL) <> (procedimiento_realizado_id IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS idx_adendas_nota ON adendas (nota_id);
CREATE INDEX IF NOT EXISTS idx_adendas_paciente ON adendas (paciente_id, creada_en DESC);

DROP TRIGGER IF EXISTS trg_adendas_inmutables ON adendas;
CREATE TRIGGER trg_adendas_inmutables BEFORE UPDATE OR DELETE ON adendas
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();
