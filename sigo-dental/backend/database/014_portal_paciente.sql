-- Migración 014: portal del paciente (Fase 2, F2-8). Requiere las migraciones 001, 003 y 011 (códigos OTP).
-- El paciente es un rol nuevo (fuera de los cinco del documento de requerimientos, ver docs/plan/README.md, punto 5).
-- Ingresa con su documento de identidad y un código OTP; no tiene contraseña.

INSERT INTO roles (codigo, nombre, descripcion) VALUES
    ('paciente', 'Paciente', 'Accede al portal del paciente: sus citas, recetas y consentimientos firmados.')
ON CONFLICT (codigo) DO NOTHING;

-- Cuenta del portal de un paciente: se crea la primera vez que ingresa. Su contraseña es un valor aleatorio que nadie
-- conoce, así que no se puede entrar por /api/login; solo con el código OTP.
ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS paciente_id INTEGER UNIQUE REFERENCES pacientes (id);
