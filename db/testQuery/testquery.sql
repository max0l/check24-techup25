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


select hotels.*, best_prices.price
FROM (
SELECT offers.hotelid as hotelid, MIN(offers.price) AS price
FROM offers
WHERE
    "offers"."outbounddepartureairport" IN ('BER')
    AND "offers"."outbounddeparturedatetime" >= '2025-01-01 00:00:00+02'
    AND "offers"."inboundarrivaldatetime" <= '2025-12-31 00:00:00+02'
    AND "offers"."countadults" = 2
    AND "offers"."countchildren" = 0
    AND "offers"."duration_in_days" = 7
GROUP BY offers.hotelid
) AS best_prices
JOIN hotels ON hotels.id = best_prices.hotelid
ORDER BY best_prices.price ASC
LIMIT 100;