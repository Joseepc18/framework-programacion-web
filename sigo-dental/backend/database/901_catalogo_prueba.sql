-- Datos de prueba del catálogo. SOLO para desarrollo; no cargar en producción.
-- Requiere 900_datos_prueba.sql (la cuenta del odontólogo) y 002_catalogo.sql.

INSERT INTO consultorios (nombre) VALUES ('Consultorio 1'), ('Consultorio 2')
ON CONFLICT (nombre) DO NOTHING;

INSERT INTO sillones (consultorio_id, nombre)
SELECT c.id, s.nombre
FROM (VALUES ('Consultorio 1', 'Sillón A'), ('Consultorio 1', 'Sillón B'), ('Consultorio 2', 'Sillón A'))
     AS s (consultorio, nombre)
JOIN consultorios c ON c.nombre = s.consultorio
ON CONFLICT (consultorio_id, nombre) DO NOTHING;

INSERT INTO profesionales (usuario_id, especialidad, registro_profesional)
SELECT u.id, 'Odontología general', 'REG-0001'
FROM usuarios u
WHERE u.email = 'odontologo@sigo-dental.test'
ON CONFLICT (usuario_id) DO NOTHING;

INSERT INTO procedimientos (codigo, nombre, especialidad, precio_centavos, duracion_min, requiere_consentimiento) VALUES
    ('CONS-01', 'Consulta y valoración',      'Odontología general', 2500,  30, FALSE),
    ('LIMP-01', 'Limpieza dental',            'Odontología general', 4000,  45, FALSE),
    ('REST-01', 'Restauración con resina',    'Odontología general', 6000,  60, FALSE),
    ('EXOD-01', 'Extracción simple',          'Cirugía oral',        5500,  45, TRUE),
    ('ENDO-01', 'Endodoncia unirradicular',   'Endodoncia',         18000, 90, TRUE)
ON CONFLICT (codigo) DO NOTHING;

INSERT INTO convenios (nombre) VALUES ('Seguro de prueba A'), ('Seguro de prueba B')
ON CONFLICT (nombre) DO NOTHING;
