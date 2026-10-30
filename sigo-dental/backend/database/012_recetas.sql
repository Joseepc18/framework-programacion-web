-- Migración 012: recetas (Fase 2, F2-6). Requiere 003, 004, 006 (contadores), 007 y 010 (firmas).
-- Una receta nace como borrador, se firma con el servicio de firma de F2-4 y desde entonces no se edita: se anula,
-- con registro, y si hace falta se emite otra. Nada se borra.

INSERT INTO contadores (clave, valor) VALUES ('receta', 0) ON CONFLICT (clave) DO NOTHING;

CREATE TABLE IF NOT EXISTS recetas (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    numero         VARCHAR(20)  NOT NULL UNIQUE,
    paciente_id    INTEGER      NOT NULL REFERENCES pacientes (id),
    profesional_id INTEGER      NOT NULL REFERENCES profesionales (id),
    cita_id        INTEGER      REFERENCES citas (id),
    indicaciones   VARCHAR(500) NOT NULL DEFAULT '',
    creada_por     INTEGER      NOT NULL REFERENCES usuarios (id),
    creada_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recetas_paciente ON recetas (paciente_id, creada_en DESC);

CREATE TABLE IF NOT EXISTS receta_medicamentos (
    id           INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    receta_id    INTEGER      NOT NULL REFERENCES recetas (id),
    orden        SMALLINT     NOT NULL,
    nombre       VARCHAR(150) NOT NULL CHECK (btrim(nombre) <> ''),
    dosis        VARCHAR(100) NOT NULL CHECK (btrim(dosis) <> ''),
    frecuencia   VARCHAR(100) NOT NULL CHECK (btrim(frecuencia) <> ''),
    duracion     VARCHAR(100) NOT NULL CHECK (btrim(duracion) <> ''),
    indicaciones VARCHAR(300) NOT NULL DEFAULT '',
    UNIQUE (receta_id, orden)
);

-- La anulación es un registro nuevo: la receta queda intacta y se conserva quién anuló, por qué y qué receta la reemplaza.
CREATE TABLE IF NOT EXISTS receta_anulaciones (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    receta_id   INTEGER      NOT NULL UNIQUE REFERENCES recetas (id),
    motivo      VARCHAR(300) NOT NULL CHECK (btrim(motivo) <> ''),
    usuario_id  INTEGER      NOT NULL REFERENCES usuarios (id),
    reemplazo_id INTEGER     REFERENCES recetas (id),
    creada_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trg_receta_anulaciones_inmutables ON receta_anulaciones;
CREATE TRIGGER trg_receta_anulaciones_inmutables BEFORE UPDATE OR DELETE ON receta_anulaciones
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();

-- Una receta firmada no se modifica (ni sus medicamentos); un borrador sí. Ninguna receta se borra.
CREATE OR REPLACE FUNCTION proteger_receta() RETURNS trigger AS $$
DECLARE
    receta INTEGER;
BEGIN
    IF TG_TABLE_NAME = 'recetas' THEN
        IF TG_OP = 'DELETE' THEN
            RAISE EXCEPTION 'recetas es un registro inmutable: no admite DELETE' USING ERRCODE = 'SG001';
        END IF;
        receta := OLD.id;
        IF NEW.numero <> OLD.numero OR NEW.paciente_id <> OLD.paciente_id OR NEW.profesional_id <> OLD.profesional_id THEN
            RAISE EXCEPTION 'la receta no cambia de número, paciente ni profesional' USING ERRCODE = 'SG001';
        END IF;
    ELSE
        receta := COALESCE(OLD.receta_id, NEW.receta_id);
    END IF;
    IF EXISTS (SELECT 1 FROM firmas WHERE entidad = 'receta' AND entidad_id = receta) THEN
        RAISE EXCEPTION 'una receta firmada no se edita: anúlala y emite otra' USING ERRCODE = 'SG001';
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_receta_protegida ON recetas;
CREATE TRIGGER trg_receta_protegida BEFORE UPDATE OR DELETE ON recetas
    FOR EACH ROW EXECUTE FUNCTION proteger_receta();
DROP TRIGGER IF EXISTS trg_receta_medicamentos_protegidos ON receta_medicamentos;
CREATE TRIGGER trg_receta_medicamentos_protegidos BEFORE UPDATE OR DELETE ON receta_medicamentos
    FOR EACH ROW EXECUTE FUNCTION proteger_receta();
