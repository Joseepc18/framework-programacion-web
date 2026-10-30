-- Migración 008: odontograma (Fase 2, F2-2: RF-03). Requiere las migraciones 003, 004 y 007.
-- Un hallazgo no se sobrescribe: corregirlo retira el anterior y crea otro, así el historial por visita se conserva.
-- Un hallazgo "realizado" (tratamiento hecho en la clínica) no se corrige ni se retira (RN-01).

CREATE TABLE IF NOT EXISTS odontograma_hallazgos (
    id              INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id     INTEGER      NOT NULL REFERENCES pacientes (id),
    -- Pieza en numeración FDI: permanentes 11-18, 21-28, 31-38, 41-48; temporales 51-55, 61-65, 71-75, 81-85.
    pieza           SMALLINT     NOT NULL CHECK (
        (pieza / 10 BETWEEN 1 AND 4 AND pieza % 10 BETWEEN 1 AND 8) OR (pieza / 10 BETWEEN 5 AND 8 AND pieza % 10 BETWEEN 1 AND 5)),
    -- NULL = pieza completa (endodoncia, implante, prótesis, exodoncia, ausente).
    superficie      VARCHAR(10)  CHECK (superficie IN ('mesial', 'distal', 'oclusal', 'vestibular', 'lingual')),
    tipo            VARCHAR(20)  NOT NULL CHECK (tipo IN ('caries', 'restauracion', 'endodoncia', 'implante', 'protesis_fija',
                                                           'protesis_removible', 'exodoncia', 'sellante', 'ausente')),
    estado          VARCHAR(10)  NOT NULL CHECK (estado IN ('requerido', 'existente', 'realizado')),
    observacion     VARCHAR(200) NOT NULL DEFAULT '',
    -- Visita en la que se registró: la cita En atención del paciente en ese momento, si había una.
    cita_id         INTEGER      REFERENCES citas (id),
    registrado_por  INTEGER      NOT NULL REFERENCES usuarios (id),
    registrado_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    retirado_por    INTEGER      REFERENCES usuarios (id),
    retirado_en     TIMESTAMPTZ,
    -- Si este hallazgo corrige a otro, apunta al que retiró.
    reemplaza_a     INTEGER      REFERENCES odontograma_hallazgos (id),
    CHECK ((retirado_por IS NULL) = (retirado_en IS NULL))
);

CREATE INDEX IF NOT EXISTS idx_hallazgos_paciente ON odontograma_hallazgos (paciente_id, registrado_en);
-- Una misma marca vigente no se repite en el mismo diente y superficie.
CREATE UNIQUE INDEX IF NOT EXISTS uq_hallazgo_vigente
    ON odontograma_hallazgos (paciente_id, pieza, COALESCE(superficie, ''), tipo) WHERE retirado_en IS NULL;

-- Solo se permite marcar como retirado un hallazgo vigente y no realizado; nada más de la fila cambia ni se borra.
CREATE OR REPLACE FUNCTION proteger_hallazgo() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'odontograma_hallazgos es un registro inmutable: no admite DELETE' USING ERRCODE = 'SG001';
    END IF;
    IF OLD.estado = 'realizado' OR OLD.retirado_en IS NOT NULL THEN
        RAISE EXCEPTION 'un hallazgo realizado o ya retirado no se modifica (RN-01)' USING ERRCODE = 'SG001';
    END IF;
    IF NEW.paciente_id <> OLD.paciente_id OR NEW.pieza <> OLD.pieza OR NEW.superficie IS DISTINCT FROM OLD.superficie
       OR NEW.tipo <> OLD.tipo OR NEW.estado <> OLD.estado OR NEW.observacion <> OLD.observacion
       OR NEW.registrado_por <> OLD.registrado_por OR NEW.registrado_en <> OLD.registrado_en THEN
        RAISE EXCEPTION 'solo se puede retirar un hallazgo, no editarlo: registra uno nuevo' USING ERRCODE = 'SG001';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_hallazgo_protegido ON odontograma_hallazgos;
CREATE TRIGGER trg_hallazgo_protegido BEFORE UPDATE OR DELETE ON odontograma_hallazgos
    FOR EACH ROW EXECUTE FUNCTION proteger_hallazgo();
