-- Datos de prueba: un usuario por cada rol. SOLO para desarrollo; no cargar en producción.
-- La contraseña de prueba es la misma para las cinco cuentas y vive únicamente en este archivo:
-- Sigo-Dental-2026
-- Se hashea con bcrypt (costo 10) al insertar, con la extensión pgcrypto creada en la migración 001.

INSERT INTO usuarios (email, password_hash, nombre, rol_id)
SELECT datos.email, crypt('Sigo-Dental-2026', gen_salt('bf', 10)), datos.nombre, roles.id
FROM (VALUES
    ('admin@sigo-dental.test',         'Administrador TI de prueba',   'admin_ti'),
    ('recepcion@sigo-dental.test',     'Recepcionista de prueba',      'recepcionista'),
    ('odontologo@sigo-dental.test',    'Odontólogo de prueba',         'odontologo'),
    ('asistente@sigo-dental.test',     'Asistente dental de prueba',   'asistente'),
    ('auditor@sigo-dental.test',       'Auditor de prueba',            'auditor')
) AS datos (email, nombre, rol_codigo)
JOIN roles ON roles.codigo = datos.rol_codigo
ON CONFLICT (email) DO NOTHING;
