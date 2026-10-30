-- Migración 009: periodontograma digital (Fase 2, F2-3: RF-04). Requiere las migraciones 003 y 007.
-- Un examen no se modifica: se compara con los anteriores y una corrección es un examen nuevo.
-- Convención del nivel de inserción clínica (NIC): NIC = profundidad al sondaje + margen gingival, con el margen medido
-- desde la unión cemento-esmalte: positivo si el margen está apical a ella (recesión) y negativo si está coronal.

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('bloqueo_nota_horas', '24', 'Horas durante las cuales una nota de evolución aún se puede editar; después se bloquea (RF-05).')
ON CONFLICT (clave) DO NOTHING;

CREATE TABLE IF NOT EXISTS periodontogramas (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id INTEGER      NOT NULL REFERENCES pacientes (id),
    fecha       DATE         NOT NULL,
    notas       VARCHAR(500) NOT NULL DEFAULT '',
    creado_por  INTEGER      NOT NULL REFERENCES usuarios (id),
    creado_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_periodontogramas_paciente ON periodontogramas (paciente_id, fecha DESC, id DESC);

-- Datos por diente. Solo dentición permanente (11-18, 21-28, 31-38, 41-48).
CREATE TABLE IF NOT EXISTS periodonto_dientes (
    examen_id INTEGER  NOT NULL REFERENCES periodontogramas (id),
    pieza     SMALLINT NOT NULL CHECK (pieza / 10 BETWEEN 1 AND 4 AND pieza % 10 BETWEEN 1 AND 8),
    -- Movilidad de 0 a 3 y furca de 0 (sin compromiso) a 3 (grado III).
    movilidad SMALLINT NOT NULL DEFAULT 0 CHECK (movilidad BETWEEN 0 AND 3),
    furca     SMALLINT NOT NULL DEFAULT 0 CHECK (furca BETWEEN 0 AND 3),
    PRIMARY KEY (examen_id, pieza)
);

-- Seis puntos de sondaje por diente: distal, medio y mesial por vestibular (DV, V, MV) y por lingual (DL, L, ML).
CREATE TABLE IF NOT EXISTS periodonto_sitios (
    examen_id        INTEGER  NOT NULL,
    pieza            SMALLINT NOT NULL,
    sitio            VARCHAR(2) NOT NULL CHECK (sitio IN ('DV', 'V', 'MV', 'DL', 'L', 'ML')),
    profundidad_mm   SMALLINT NOT NULL CHECK (profundidad_mm BETWEEN 0 AND 15),
    margen_mm        SMALLINT NOT NULL CHECK (margen_mm BETWEEN -10 AND 15),
    sangrado         BOOLEAN  NOT NULL DEFAULT FALSE,
    PRIMARY KEY (examen_id, pieza, sitio),
    FOREIGN KEY (examen_id, pieza) REFERENCES periodonto_dientes (examen_id, pieza),
    -- El NIC no puede ser negativo: el fondo del surco no está por encima de la unión cemento-esmalte.
    CHECK (profundidad_mm + margen_mm >= 0)
);

DROP TRIGGER IF EXISTS trg_periodontogramas_inmutables ON periodontogramas;
CREATE TRIGGER trg_periodontogramas_inmutables BEFORE UPDATE OR DELETE ON periodontogramas
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();
DROP TRIGGER IF EXISTS trg_periodonto_dientes_inmutables ON periodonto_dientes;
CREATE TRIGGER trg_periodonto_dientes_inmutables BEFORE UPDATE OR DELETE ON periodonto_dientes
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();
DROP TRIGGER IF EXISTS trg_periodonto_sitios_inmutables ON periodonto_sitios;
CREATE TRIGGER trg_periodonto_sitios_inmutables BEFORE UPDATE OR DELETE ON periodonto_sitios
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();
