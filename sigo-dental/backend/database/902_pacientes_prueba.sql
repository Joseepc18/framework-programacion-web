-- Pacientes de prueba con su anamnesis y alertas médicas. SOLO para desarrollo; no cargar en producción.
-- Requiere 002_catalogo.sql, 003_pacientes.sql y 901_catalogo_prueba.sql (convenios).

WITH nuevos AS (
    INSERT INTO pacientes (documento, nombres, apellidos, fecha_nacimiento, sexo, telefono, email, direccion,
                           emergencia_nombre, emergencia_telefono, emergencia_parentesco,
                           tutor_nombre, tutor_documento, tutor_telefono, convenio_id, canal_preferido)
    VALUES
        ('0900000001', 'Ana',    'Prueba',   DATE '1990-05-20', 'femenino',  '0991112233', 'ana@correo.test',    'Calle 1 y Av. Principal', 'Luis Prueba',   '0994445566', 'Hermano',
         '', '', '', (SELECT id FROM convenios WHERE nombre = 'Seguro de prueba A'), 'whatsapp'),
        ('0900000002', 'Carlos', 'Mendoza',  DATE '1984-11-02', 'masculino', '0982223344', '',                   'Barrio Centro',           'Rosa Mendoza',  '0985556677', 'Esposa',
         '', '', '', NULL, 'sms'),
        ('0900000003', 'Sofía',  'Lara',     DATE '2016-03-12', 'femenino',  '0973334455', '',                   'Ciudadela Norte',         'Marta Lara',    '0976667788', 'Madre',
         'Marta Lara', '0907777777', '0976667788', NULL, 'whatsapp'),
        ('0900000004', 'Jorge',  'Ríos',     DATE '1972-08-30', 'masculino', '0964445566', 'jorge@correo.test',  'Av. del Mar',             'Elena Ríos',    '0967778899', 'Hija',
         '', '', '', (SELECT id FROM convenios WHERE nombre = 'Seguro de prueba B'), 'email')
    ON CONFLICT (documento) DO NOTHING
    RETURNING id, documento
), anamnesis_nueva AS (
    INSERT INTO anamnesis (paciente_id, enfermedades_sistemicas, alergias_farmacos, antecedentes_quirurgicos, coagulopatias)
    SELECT n.id, a.enfermedades, a.alergias, a.cirugias, a.coagulopatias
    FROM nuevos n
    JOIN (VALUES
        ('0900000001', 'Hipertensión controlada', 'Penicilina', 'Apendicectomía (2010)', 'Anticoagulado con warfarina'),
        ('0900000002', 'Ninguna',                 'Ninguna',    'Ninguna',               'Ninguna'),
        ('0900000003', 'Ninguna',                 'Ninguna',    'Ninguna',               'Ninguna'),
        ('0900000004', 'Diabetes tipo 2',         'Ninguna',    'Ninguna',               'Ninguna')
    ) AS a (documento, enfermedades, alergias, cirugias, coagulopatias) ON a.documento = n.documento
    RETURNING paciente_id
)
INSERT INTO paciente_condiciones_criticas (paciente_id, condicion_id)
SELECT n.id, c.id
FROM nuevos n
JOIN (VALUES
    ('0900000001', 'Alergia a penicilina'),
    ('0900000001', 'Anticoagulado'),
    ('0900000004', 'Diabetes')
) AS pc (documento, condicion) ON pc.documento = n.documento
JOIN condiciones_criticas c ON c.nombre = pc.condicion
ON CONFLICT DO NOTHING;
