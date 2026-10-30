-- Citas de prueba alrededor del día en que se ejecuta el script. SOLO para desarrollo; no cargar en producción.
-- Requiere 004_agenda.sql y los datos de prueba 900, 901 y 902. Las horas son de la zona America/Guayaquil.

WITH base AS (
    SELECT (date_trunc('day', NOW() AT TIME ZONE 'America/Guayaquil')) AS hoy
), datos (documento, dia, hora, minutos, sillon, estado, motivo) AS (
    VALUES
        ('0900000001', 0,  9, 45, 'Consultorio 1/Sillón A', 'confirmado', 'Limpieza dental'),
        ('0900000002', 0, 10, 60, 'Consultorio 1/Sillón A', 'programado', 'Restauración con resina'),
        ('0900000004', 0,  9, 30, 'Consultorio 1/Sillón B', 'en_sala',    'Consulta y valoración'),
        ('0900000003', 0, 11, 45, 'Consultorio 2/Sillón A', 'programado', 'Control de niño sano'),
        ('0900000001', 1,  8, 90, 'Consultorio 1/Sillón A', 'programado', 'Endodoncia'),
        ('0900000004', 2, 15, 30, 'Consultorio 1/Sillón B', 'programado', 'Control de diabetes y encías')
), nuevas AS (
    INSERT INTO citas (paciente_id, profesional_id, sillon_id, motivo, inicio, fin, estado, creado_por)
    SELECT p.id, (SELECT MIN(id) FROM profesionales), s.id, d.motivo,
           (b.hoy + (d.dia || ' days')::interval + (d.hora || ' hours')::interval) AT TIME ZONE 'America/Guayaquil',
           (b.hoy + (d.dia || ' days')::interval + (d.hora || ' hours')::interval + (d.minutos || ' minutes')::interval) AT TIME ZONE 'America/Guayaquil',
           d.estado, (SELECT MIN(id) FROM usuarios WHERE rol_id = (SELECT id FROM roles WHERE codigo = 'recepcionista'))
      FROM datos d
      CROSS JOIN base b
      JOIN pacientes p ON p.documento = d.documento
      JOIN sillones s ON s.nombre = split_part(d.sillon, '/', 2)
       AND s.consultorio_id = (SELECT id FROM consultorios WHERE nombre = split_part(d.sillon, '/', 1))
    RETURNING id, estado
)
INSERT INTO cita_historial (cita_id, estado_nuevo, motivo, usuario_id)
SELECT id, 'programado', 'Cita agendada', (SELECT MIN(id) FROM usuarios WHERE rol_id = (SELECT id FROM roles WHERE codigo = 'recepcionista'))
  FROM nuevas;
