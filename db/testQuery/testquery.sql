SELECT *
FROM "offers"
WHERE
    "offers"."outbounddepartureairport" IN ('BER')
    AND "offers"."outbounddeparturedatetime" >= '2025-01-01 00:00:00+02'
    AND "offers"."inboundarrivaldatetime" <= '2025-12-31 00:00:00+02'
    AND "offers"."countadults" = 2
    AND "offers"."countchildren" = 0
    AND "offers"."duration_in_days" = 7
ORDER BY "offers"."price" ASC;

SELECT *
FROM (
  SELECT DISTINCT ON (o.hotelid) o.*
  FROM offers o
  WHERE
    o.outbounddepartureairport IN ('BER','DUS','FRA','MUC')
    AND o.outbounddeparturedatetime >= TIMESTAMPTZ '2025-01-01 00:00:00+02'
    AND o.inboundarrivaldatetime  <  TIMESTAMPTZ '2026-01-01 00:00:00+02'
    AND o.countadults = 2
    AND o.countchildren = 0
    AND o.duration_in_days = 7
  ORDER BY o.hotelid, o.price ASC, o.id
) t
ORDER BY t.price ASC;
