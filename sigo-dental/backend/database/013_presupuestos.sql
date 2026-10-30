-- Migración 013: presupuestos multinivel y aprobación a citas (Fase 2, F2-7: RF-07 y RF-08).
-- Requiere 002, 004, 006 (contadores), 008 (hallazgos) y 010 (procedimientos realizados). Dinero en centavos enteros.
-- Un presupuesto presentado o resuelto no se edita: un cambio genera una versión nueva. Nada se borra una vez presentado.

INSERT INTO contadores (clave, valor) VALUES ('presupuesto', 0) ON CONFLICT (clave) DO NOTHING;

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('descuento_roles', 'odontologo,admin_ti', 'Roles (códigos separados por comas) que pueden autorizar un descuento en un presupuesto.')
ON CONFLICT (clave) DO NOTHING;

CREATE TABLE IF NOT EXISTS presupuestos (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    numero         VARCHAR(20)  NOT NULL,
    version        INTEGER      NOT NULL DEFAULT 1 CHECK (version >= 1),
    origen_id      INTEGER      REFERENCES presupuestos (id),
    paciente_id    INTEGER      NOT NULL REFERENCES pacientes (id),
    profesional_id INTEGER      NOT NULL REFERENCES profesionales (id),
    estado         VARCHAR(16)  NOT NULL DEFAULT 'borrador'
                   CHECK (estado IN ('borrador', 'presentado', 'aprobado_total', 'aprobado_parcial', 'rechazado')),
    notas          VARCHAR(500) NOT NULL DEFAULT '',
    creado_por     INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    presentado_por INTEGER      REFERENCES usuarios (id),
    presentado_en  TIMESTAMPTZ,
    resuelto_por   INTEGER      REFERENCES usuarios (id),
    resuelto_en    TIMESTAMPTZ,
    UNIQUE (numero, version),
    -- Una versión nueva se crea una sola vez a partir de cada presupuesto.
    CHECK ((origen_id IS NULL) = (version = 1))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_presupuesto_origen ON presupuestos (origen_id) WHERE origen_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_presupuestos_paciente ON presupuestos (paciente_id, id DESC);

CREATE TABLE IF NOT EXISTS presupuesto_items (
    id                      INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    presupuesto_id          INTEGER      NOT NULL REFERENCES presupuestos (id),
    fase                    SMALLINT     NOT NULL CHECK (fase BETWEEN 1 AND 20),
    sesion                  SMALLINT     NOT NULL CHECK (sesion BETWEEN 1 AND 50),
    procedimiento_id        INTEGER      NOT NULL REFERENCES procedimientos (id),
    descripcion             VARCHAR(150) NOT NULL,
    pieza                   SMALLINT     CHECK (pieza / 10 BETWEEN 1 AND 8 AND pieza % 10 BETWEEN 1 AND 8),
    superficie              VARCHAR(10)  CHECK (superficie IN ('mesial', 'distal', 'oclusal', 'vestibular', 'lingual')),
    hallazgo_id             INTEGER      REFERENCES odontograma_hallazgos (id),
    honorario_centavos      INTEGER      NOT NULL CHECK (honorario_centavos >= 0),
    laboratorio_centavos    INTEGER      NOT NULL DEFAULT 0 CHECK (laboratorio_centavos >= 0),
    descuento_centavos      INTEGER      NOT NULL DEFAULT 0 CHECK (descuento_centavos >= 0),
    descuento_autorizado_por INTEGER     REFERENCES usuarios (id),
    total_centavos          INTEGER      NOT NULL,
    -- NULL = sin decidir; verdadero = aprobado (pasa a las tareas agendables); falso = rechazado.
    aprobado                BOOLEAN,
    orden                   SMALLINT     NOT NULL DEFAULT 0,
    -- Todo descuento queda con quién lo autorizó, y el total es la suma del ítem: honorario + laboratorio - descuento.
    CHECK (descuento_centavos = 0 OR descuento_autorizado_por IS NOT NULL),
    CHECK (descuento_centavos <= honorario_centavos + laboratorio_centavos),
    CHECK (total_centavos = honorario_centavos + laboratorio_centavos - descuento_centavos)
);

CREATE INDEX IF NOT EXISTS idx_presupuesto_items ON presupuesto_items (presupuesto_id, fase, sesion, orden);

-- Vínculo de las tareas aprobadas con la agenda (F1-3) y con lo realizado (F2-4).
ALTER TABLE citas ADD COLUMN IF NOT EXISTS presupuesto_item_id INTEGER REFERENCES presupuesto_items (id);
ALTER TABLE procedimientos_realizados ADD COLUMN IF NOT EXISTS presupuesto_item_id INTEGER REFERENCES presupuesto_items (id);
CREATE INDEX IF NOT EXISTS idx_citas_item ON citas (presupuesto_item_id) WHERE presupuesto_item_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_proc_realizados_item ON procedimientos_realizados (presupuesto_item_id) WHERE presupuesto_item_id IS NOT NULL;

-- Presupuesto: nunca se borra. Del borrador al presentado y de ahí a un estado resuelto; después ya no cambia.
CREATE OR REPLACE FUNCTION proteger_presupuesto() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'presupuestos es un registro inmutable: no admite DELETE' USING ERRCODE = 'SG001';
    END IF;
    IF NEW.numero <> OLD.numero OR NEW.version <> OLD.version OR NEW.paciente_id <> OLD.paciente_id
       OR NEW.profesional_id <> OLD.profesional_id OR NEW.origen_id IS DISTINCT FROM OLD.origen_id THEN
        RAISE EXCEPTION 'el presupuesto no cambia de número, versión, paciente ni profesional' USING ERRCODE = 'SG001';
    END IF;
    IF OLD.estado = 'borrador' THEN
        RETURN NEW;
    ELSIF OLD.estado = 'presentado' AND NEW.estado IN ('aprobado_total', 'aprobado_parcial', 'rechazado') AND NEW.notas = OLD.notas THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'un presupuesto presentado o resuelto no se edita: crea una versión nueva' USING ERRCODE = 'SG001';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_presupuesto_protegido ON presupuestos;
CREATE TRIGGER trg_presupuesto_protegido BEFORE UPDATE OR DELETE ON presupuestos
    FOR EACH ROW EXECUTE FUNCTION proteger_presupuesto();

-- Ítems: libres mientras el presupuesto es borrador; al estar presentado solo cambia la decisión (aprobado).
CREATE OR REPLACE FUNCTION proteger_presupuesto_item() RETURNS trigger AS $$
DECLARE
    estado_padre VARCHAR(16);
BEGIN
    SELECT estado INTO estado_padre FROM presupuestos WHERE id = COALESCE(OLD.presupuesto_id, NEW.presupuesto_id);
    IF estado_padre = 'borrador' THEN
        IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
        RETURN NEW;
    END IF;
    IF TG_OP = 'UPDATE' AND estado_padre = 'presentado' AND NEW.presupuesto_id = OLD.presupuesto_id AND NEW.procedimiento_id = OLD.procedimiento_id
       AND NEW.honorario_centavos = OLD.honorario_centavos AND NEW.laboratorio_centavos = OLD.laboratorio_centavos
       AND NEW.descuento_centavos = OLD.descuento_centavos AND NEW.total_centavos = OLD.total_centavos
       AND NEW.fase = OLD.fase AND NEW.sesion = OLD.sesion AND OLD.aprobado IS NULL THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION 'los ítems de un presupuesto presentado o resuelto no se editan: crea una versión nueva' USING ERRCODE = 'SG001';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_presupuesto_item_protegido ON presupuesto_items;
CREATE TRIGGER trg_presupuesto_item_protegido BEFORE UPDATE OR DELETE ON presupuesto_items
    FOR EACH ROW EXECUTE FUNCTION proteger_presupuesto_item();
