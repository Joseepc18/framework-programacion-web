-- Migración 006: realinea referidos.comisiones con el ledger vigente.
-- Antes de esta corrección, cancelar un pedido revertía las filas de comisiones
-- pero dejaba el contador del nodo sin descontar. El contador pasa a ser la suma
-- del ledger sin las comisiones revertidas.
-- Es idempotente: se puede aplicar más de una vez.
--
-- Aplicación sobre un volumen existente:
--   docker cp database/006_contador_comisiones.sql multicatalogo-db:/tmp/006_contador_comisiones.sql
--   docker exec multicatalogo-db psql -U multicatalogo -d multicatalogo -v ON_ERROR_STOP=1 -f /tmp/006_contador_comisiones.sql

BEGIN;

UPDATE referidos r
   SET comisiones = COALESCE(c.vigente, 0)
  FROM referidos x
  LEFT JOIN (
        SELECT referido_id, SUM(monto) AS vigente
          FROM comisiones
         WHERE estado <> 'revertida'
         GROUP BY referido_id
       ) c ON c.referido_id = x.id
 WHERE r.id = x.id
   AND r.comisiones IS DISTINCT FROM COALESCE(c.vigente, 0);

COMMIT;
