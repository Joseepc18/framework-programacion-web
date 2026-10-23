-- Reparte en el tiempo los pedidos que crea scripts/demo-data.mjs.
-- La API siempre fecha los pedidos "ahora"; sin esto la demostración mostraría
-- cinco semanas de ventas con una sola fecha. Solo se tocan los pedidos de las
-- cuentas de demostración y únicamente los creados en las últimas 24 horas,
-- así que volver a ejecutarlo no mueve nada.
--
-- Local:  docker exec -i multicatalogo-db psql -U multicatalogo -d multicatalogo -v ON_ERROR_STOP=1 < scripts/demo-fechas.sql

BEGIN;

WITH demo AS (
    SELECT id,
           ROW_NUMBER() OVER (ORDER BY id) AS n,
           COUNT(*)     OVER ()            AS total
      FROM pedidos
     WHERE cliente_email IN (
             'cliente@upse.edu.ec', 'vale.cedeno@hotmail.com', 'mateoalvarado.97@gmail.com',
             'camizambrano@outlook.es', 'avillon@gmail.com', 'dani.reyes@upse.edu.ec',
             'jtomala@yahoo.com', 'kpanchana@gmail.com', 'lissquimi@hotmail.com')
       AND creado_en > NOW() - INTERVAL '24 hours'
)
UPDATE pedidos p
   SET creado_en = NOW() - INTERVAL '1 hour' * ((d.total - d.n) * 58 + (d.n * 7) % 13)
  FROM demo d
 WHERE p.id = d.id;

-- El número de pedido lleva la fecha: se regenera con la fecha nueva.
UPDATE pedidos
   SET numero = 'MC-' || TO_CHAR(creado_en, 'YYMMDD') || '-' || id
 WHERE numero LIKE 'MC-%-%'
   AND numero <> 'MC-' || TO_CHAR(creado_en, 'YYMMDD') || '-' || id
   AND cliente_email IN (
         'cliente@upse.edu.ec', 'vale.cedeno@hotmail.com', 'mateoalvarado.97@gmail.com',
         'camizambrano@outlook.es', 'avillon@gmail.com', 'dani.reyes@upse.edu.ec',
         'jtomala@yahoo.com', 'kpanchana@gmail.com', 'lissquimi@hotmail.com');

-- Cada comisión nace con su pedido.
UPDATE comisiones c
   SET creado_en = p.creado_en
  FROM pedidos p
 WHERE c.pedido_id = p.id
   AND c.creado_en <> p.creado_en
   AND p.cliente_email IN (
         'cliente@upse.edu.ec', 'vale.cedeno@hotmail.com', 'mateoalvarado.97@gmail.com',
         'camizambrano@outlook.es', 'avillon@gmail.com', 'dani.reyes@upse.edu.ec',
         'jtomala@yahoo.com', 'kpanchana@gmail.com', 'lissquimi@hotmail.com');

-- La liquidación de Valentina se pagó hace unos días.
UPDATE liquidaciones
   SET creado_en = NOW() - INTERVAL '6 days',
       pagada_en = NOW() - INTERVAL '6 days',
       periodo   = TO_CHAR(NOW() - INTERVAL '6 days', 'YYYY-MM')
 WHERE creado_en > NOW() - INTERVAL '24 hours'
   AND referido_id = (SELECT referido_id FROM usuarios WHERE email = 'vale.cedeno@hotmail.com');

COMMIT;
