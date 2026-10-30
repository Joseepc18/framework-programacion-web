-- Migración 015: contenido de los archivos del paciente dentro de la base de datos (ALMACEN_ARCHIVOS=postgres).
-- Alternativa al disco y a S3 para entornos sin disco persistente, como el plan gratuito de Render. Solo se inserta y se
-- lee: igual que el resto del expediente, nada se modifica ni se borra.
CREATE TABLE IF NOT EXISTS archivos_contenido (
    clave      text PRIMARY KEY,
    tipo_mime  text NOT NULL DEFAULT '',
    contenido  bytea NOT NULL,
    creado_en  timestamptz NOT NULL DEFAULT now()
);
