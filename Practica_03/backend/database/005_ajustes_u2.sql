-- Migración 005: ajustes finales de U2 sobre un volumen existente.
--  - liquidaciones.periodo (AAAA-MM) para el estado de cuenta del distribuidor.
--  - Semilla alineada con el plan: la cuenta demo cliente@upse.edu.ec representa a
--    Luis Poveda (nodo 2), no a Diana Paz (nodo 7).
-- Es idempotente: se puede aplicar más de una vez.

BEGIN;

ALTER TABLE liquidaciones ADD COLUMN IF NOT EXISTS periodo VARCHAR(7);
UPDATE liquidaciones SET periodo = to_char(COALESCE(pagada_en, creado_en), 'YYYY-MM') WHERE periodo IS NULL;
ALTER TABLE liquidaciones ALTER COLUMN periodo SET DEFAULT to_char(NOW(), 'YYYY-MM');
ALTER TABLE liquidaciones ALTER COLUMN periodo SET NOT NULL;

-- Liberamos el correo en el nodo 7 antes de asignarlo al 2 (el correo es único por nodo).
UPDATE referidos SET email = NULL WHERE id = 7 AND email = 'cliente@upse.edu.ec';
UPDATE referidos SET email = 'cliente@upse.edu.ec' WHERE id = 2 AND email IS NULL;
UPDATE usuarios  SET referido_id = 2 WHERE email = 'cliente@upse.edu.ec' AND referido_id = 7;

COMMIT;
