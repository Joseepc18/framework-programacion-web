-- Migración 007: auditoría inmutable, signos vitales y archivos del paciente (Fase 2, F2-1: RNF-06).
-- Requiere las migraciones 001 a 006. Nada se borra y los registros clínicos no se modifican.

INSERT INTO parametros (clave, valor, descripcion) VALUES
    ('archivos_max_mb',   '15', 'Tamaño máximo, en megabytes, de cada archivo que se sube al expediente (radiografías, fotos y PDF).'),
    ('clinica_nombre',    'Clínica Odontológica SIGO-Dental', 'Nombre de la clínica en el encabezado de recetas, consentimientos y presupuestos.'),
    ('clinica_direccion', '', 'Dirección de la clínica para el encabezado de los documentos impresos.'),
    ('clinica_telefono',  '', 'Teléfono de la clínica para el encabezado de los documentos impresos.')
ON CONFLICT (clave) DO NOTHING;

-- Función común de los disparadores de inmutabilidad. El código SG001 lo reconoce la API como "registro inmutable".
CREATE OR REPLACE FUNCTION rechazar_cambio() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION '% es un registro inmutable: no admite %', TG_TABLE_NAME, TG_OP USING ERRCODE = 'SG001';
END;
$$ LANGUAGE plpgsql;

-- Bitácora de accesos a datos clínicos (RNF-06): quién consultó, creó, modificó o exportó qué registro, cuándo y desde dónde.
-- Es de solo inserción: el disparador impide UPDATE, DELETE y TRUNCATE incluso a quien sea dueño de la tabla, y se
-- retiran esos permisos a PUBLIC. En producción la API debe conectarse con un rol que solo tenga INSERT y SELECT (docs/architecture.md).
CREATE TABLE IF NOT EXISTS auditoria_accesos (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ocurrido_en   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    usuario_id    INTEGER      REFERENCES usuarios (id),
    usuario_email VARCHAR(150) NOT NULL DEFAULT '',
    rol           VARCHAR(30)  NOT NULL DEFAULT '',
    accion        VARCHAR(20)  NOT NULL CHECK (accion IN ('consulta', 'creacion', 'modificacion', 'exportacion', 'firma', 'anulacion')),
    entidad       VARCHAR(40)  NOT NULL,
    entidad_id    INTEGER,
    paciente_id   INTEGER      REFERENCES pacientes (id),
    ip            VARCHAR(64)  NOT NULL DEFAULT '',
    metodo        VARCHAR(10)  NOT NULL DEFAULT '',
    ruta          VARCHAR(200) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_auditoria_fecha ON auditoria_accesos (ocurrido_en DESC);
CREATE INDEX IF NOT EXISTS idx_auditoria_paciente ON auditoria_accesos (paciente_id, ocurrido_en DESC);
CREATE INDEX IF NOT EXISTS idx_auditoria_usuario ON auditoria_accesos (usuario_id, ocurrido_en DESC);

DROP TRIGGER IF EXISTS trg_auditoria_sin_cambios ON auditoria_accesos;
CREATE TRIGGER trg_auditoria_sin_cambios BEFORE UPDATE OR DELETE ON auditoria_accesos
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();
DROP TRIGGER IF EXISTS trg_auditoria_sin_truncate ON auditoria_accesos;
CREATE TRIGGER trg_auditoria_sin_truncate BEFORE TRUNCATE ON auditoria_accesos
    FOR EACH STATEMENT EXECUTE FUNCTION rechazar_cambio();
REVOKE UPDATE, DELETE, TRUNCATE ON auditoria_accesos FROM PUBLIC;

-- Signos vitales por visita. Cada toma es un registro nuevo; una corrección es otra toma.
CREATE TABLE IF NOT EXISTS signos_vitales (
    id                      INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id             INTEGER      NOT NULL REFERENCES pacientes (id),
    cita_id                 INTEGER      REFERENCES citas (id),
    sistolica               SMALLINT     CHECK (sistolica BETWEEN 50 AND 260),
    diastolica              SMALLINT     CHECK (diastolica BETWEEN 20 AND 160),
    frecuencia_cardiaca     SMALLINT     CHECK (frecuencia_cardiaca BETWEEN 20 AND 250),
    frecuencia_respiratoria SMALLINT     CHECK (frecuencia_respiratoria BETWEEN 4 AND 80),
    temperatura             NUMERIC(3,1) CHECK (temperatura BETWEEN 30 AND 45),
    saturacion              SMALLINT     CHECK (saturacion BETWEEN 50 AND 100),
    peso_kg                 NUMERIC(5,1) CHECK (peso_kg BETWEEN 0.5 AND 600),
    talla_cm                NUMERIC(4,1) CHECK (talla_cm BETWEEN 30 AND 250),
    observaciones           VARCHAR(300) NOT NULL DEFAULT '',
    registrado_por          INTEGER      NOT NULL REFERENCES usuarios (id),
    registrado_en           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    -- La presión se registra completa o no se registra.
    CHECK ((sistolica IS NULL) = (diastolica IS NULL)),
    CHECK (diastolica IS NULL OR diastolica < sistolica),
    CHECK (sistolica IS NOT NULL OR frecuencia_cardiaca IS NOT NULL OR frecuencia_respiratoria IS NOT NULL
           OR temperatura IS NOT NULL OR saturacion IS NOT NULL OR peso_kg IS NOT NULL OR talla_cm IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_signos_paciente ON signos_vitales (paciente_id, registrado_en DESC);

DROP TRIGGER IF EXISTS trg_signos_inmutables ON signos_vitales;
CREATE TRIGGER trg_signos_inmutables BEFORE UPDATE OR DELETE ON signos_vitales
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();

-- Archivos del paciente (radiografías, fotos, PDF). El contenido vive en el almacén (disco o S3) bajo clave_almacen;
-- aquí solo quedan los metadatos y la huella SHA-256 para detectar alteraciones.
CREATE TABLE IF NOT EXISTS archivos_paciente (
    id              INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    paciente_id     INTEGER      NOT NULL REFERENCES pacientes (id),
    categoria       VARCHAR(20)  NOT NULL CHECK (categoria IN ('radiografia', 'fotografia', 'documento', 'consentimiento', 'receta', 'presupuesto')),
    descripcion     VARCHAR(200) NOT NULL DEFAULT '',
    nombre_original VARCHAR(200) NOT NULL,
    tipo_mime       VARCHAR(100) NOT NULL,
    tamano_bytes    INTEGER      NOT NULL CHECK (tamano_bytes > 0),
    sha256          CHAR(64)     NOT NULL,
    clave_almacen   VARCHAR(200) NOT NULL UNIQUE,
    subido_por      INTEGER      NOT NULL REFERENCES usuarios (id),
    subido_en       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_archivos_paciente ON archivos_paciente (paciente_id, subido_en DESC);

DROP TRIGGER IF EXISTS trg_archivos_inmutables ON archivos_paciente;
CREATE TRIGGER trg_archivos_inmutables BEFORE UPDATE OR DELETE ON archivos_paciente
    FOR EACH ROW EXECUTE FUNCTION rechazar_cambio();
