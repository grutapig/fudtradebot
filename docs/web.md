curl 'https://www.asterdex.com/bapi/futures/v1/private/future/order/place-order' \
--compressed \
-X POST \
-H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:143.0) Gecko/20100101 Firefox/143.0' \
-H 'Accept: application/json, text/plain, */*' \
-H 'Accept-Language: en-US,ru-RU;q=0.94,de-DE;q=0.88,ru;q=0.82,en;q=0.76,pl-PL;q=0.71,pl;q=0.65,en-GB;q=0.59,de;q=0.53,fr-FR;q=0.47,fr;q=0.41,ja-JP;q=0.35,ja;q=0.29,uk-UA;q=0.24,uk;q=0.18,es-ES;q=0.12,es;q=0.06' \
-H 'Accept-Encoding: gzip, deflate, br, zstd' \
-H 'Referer: https://www.asterdex.com/en/futures/v1/SOLUSDT' \
-H 'Content-Type: application/json' \
-H 'Clienttype: web' \
-H 'csrftoken: 6cefb8d569a829f1bf8d59bb93923125' \
-H 'Origin: https://www.asterdex.com' \
-H 'Connection: keep-alive' \
-H 'Cookie: lang=en; bnc-uuid=eb8c3f26-b095-45fc-8106-d1889bf7b58f; p20t=web.98000008294477.6FF07DA3E4AF460B0AD9C50472068FAB; cr00=7991191D83593A982FDA7DE39AD72A76; d1og=web.98000008294477.DEB5C748CD800527203F66CA16588692; r2o1=web.98000008294477.6CFC4874E0B88143399B87BC3E578005; f30l=web.98000008294477.72C8D2C5F92C062024BE14CF5340E64B; address=HckZCxWyuDzMJZCCWov7uZvtnpGcs5PdLyHevQ9hLUR3; ph_phc_cY9He3TLHieapMUi0QLab7OSWOjg8S85aby7C5kqsUJ_posthog=%7B%22distinct_id%22%3A%22019a1309-5a54-794e-9d22-48ba5e2fc5bb%22%2C%22%24sesid%22%3A%5B1761483420877%2C%22019a206b-4efc-7c42-885b-ea3737228a4a%22%2C1761480494844%5D%2C%22%24epp%22%3Atrue%2C%22%24initial_person_info%22%3A%7B%22r%22%3A%22%24direct%22%2C%22u%22%3A%22https%3A%2F%2Fwww.asterdex.com%2Fen%2Ffutures%2Fv1%2FBTCUSDT%3Fref%3D947dd0%22%7D%7D' \
-H 'Sec-Fetch-Dest: empty' \
-H 'Sec-Fetch-Mode: cors' \
-H 'Sec-Fetch-Site: same-origin' \
-H 'Priority: u=0' \
-H 'Pragma: no-cache' \
-H 'Cache-Control: no-cache' \
-H 'TE: trailers' \
--data-raw '{"symbol":"SOLUSDT","clientOrderId":"web_AD_3128wjnh5ycxzx5ml","placeType":"order-form","positionSide":"LONG","side":"BUY","quantity":"0.06","type":"MARKET"}'


OPEN POSITION LONG
{"symbol":"SOLUSDT","clientOrderId":"web_AD_3128wjnh5ycxzx5ml","placeType":"order-form","positionSide":"LONG","side":"BUY","quantity":"0.06","type":"MARKET"}
CLOSE
{"symbol":"SOLUSDT","clientOrderId":"web_AD_ycffovq221ogc097g","placeType":"order-form","positionSide":"LONG","side":"SELL","quantity":"0.06","type":"MARKET"}

OPEN POSITION SHORT
{"symbol":"SOLUSDT","clientOrderId":"web_AD_16ag8bk72zva9c1y5","placeType":"order-form","positionSide":"SHORT","side":"SELL","quantity":"0.06","type":"MARKET"}
CLOSE SHORT
{"symbol":"SOLUSDT","clientOrderId":"web_AD_u0yqe2bmnydhh5tnd","placeType":"order-form","positionSide":"SHORT","side":"BUY","quantity":"0.06","type":"MARKET"}
