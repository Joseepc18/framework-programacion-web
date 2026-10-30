-- Migración 001: usuarios y roles del sistema.
-- Docker Compose ejecuta los .sql de esta carpeta en orden alfabético la primera vez que se
-- crea el volumen de PostgreSQL. En Neon se aplican a mano los archivos 0xx; los 9xx son datos
-- de prueba y no se cargan en producción.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Catálogo de roles. Los cinco salen del documento de requerimientos; codigo es el identificador
-- estable que usará el backend (JWT y RBAC), nombre es lo que ve la persona usuaria.
CREATE TABLE IF NOT EXISTS roles (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    codigo      VARCHAR(30)  NOT NULL UNIQUE,
    nombre      VARCHAR(60)  NOT NULL,
    descripcion TEXT         NOT NULL DEFAULT ''
);

INSERT INTO roles (codigo, nombre, descripcion) VALUES
    ('admin_ti',      'Administrador TI',      'Gestiona usuarios, roles y la configuración del sistema.'),
    ('recepcionista', 'Recepcionista o cajero', 'Atiende la agenda, registra pacientes y opera la caja.'),
    ('odontologo',    'Odontólogo',            'Atiende pacientes y gestiona el expediente clínico y los presupuestos.'),
    ('asistente',     'Asistente dental',      'Apoya al odontólogo en la atención y el inventario.'),
    ('auditor',       'Auditor o contador',    'Consulta información financiera y registros para revisión, sin modificarlos.')
ON CONFLICT (codigo) DO NOTHING;

-- Cuentas del sistema. La contraseña se guarda siempre hasheada con bcrypt (nunca en texto plano).
CREATE TABLE IF NOT EXISTS usuarios (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    email          VARCHAR(150) NOT NULL UNIQUE,
    password_hash  VARCHAR(100) NOT NULL,
    nombre         VARCHAR(150) NOT NULL,
    -- Cada cuenta tiene un solo rol; no se puede borrar un rol que tenga cuentas.
    rol_id         INTEGER      NOT NULL REFERENCES roles (id),
    activo         BOOLEAN      NOT NULL DEFAULT TRUE,
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Consultas por rol (listados de administración y comprobaciones de permisos).
CREATE INDEX IF NOT EXISTS idx_usuarios_rol ON usuarios (rol_id);
