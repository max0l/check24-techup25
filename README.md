# Check24 Techup25

A life instance can be found here [check24.max0l.dev](https://check24.max0l.dev) for the frontend. And [check24-swagger.max0l.dev/swagger/index.html](https://check24-swagger.max0l.dev/swagger/index.html) for the swagger doc. More infos are further down.

<!-- TOC start (generated with https://github.com/derlin/bitdowntoc) -->

- [Check24 Techup25](#check24-techup25)
  - [Development](#development)
    - [Backend](#backend)
      - [ENV](#env)
      - [Seeding the data](#seeding-the-data)
      - [API defintion](#api-defintion)
      - [Swag code gen](#swag-code-gen)
      - [GOENT (ORM) Generation](#goent-orm-generation)
    - [Frontend](#frontend)
      - [(Re)Generate the Openapi Specs for the frontend](#regenerate-the-openapi-specs-for-the-frontend)
  - [Decisions I've Made](#decisions-ive-made)
    - [Postgresql](#postgresql)
    - [CSV to DB Migration](#csv-to-db-migration)
      - [Seeding the data only the first time](#seeding-the-data-only-the-first-time)
      - [Using the Service with the whole offers csv](#using-the-service-with-the-whole-offers-csv)
    - [Possible Improvements](#possible-improvements)
      - [Checking each Param individually](#checking-each-param-individually)
      - [Using a key-value store only](#using-a-key-value-store-only)
      - [Query Queue](#query-queue)
      - [Use something else then swaggo](#use-something-else-then-swaggo)
      - [Seeding the data](#seeding-the-data-1)
  - [Deployment](#deployment)

<!-- TOC end -->

<!-- TOC --><a name="development"></a>
## Development

<!-- TOC --><a name="backend"></a>
### Backend

Everything will be generated if you just run `docker compose up --build`.

> [!IMPORTANT]
> **Set the path of the offers.csv volume in the docker-compose.yml using the ENV OFFERS_CSV_PATH to your offers.csv or to the provided test CSV's in the testCSVs folder**

<!-- TOC --><a name="env"></a>
#### ENV

There are several ENV's that can be set. Most of them have a fallback required so `docker compose up` can just get fired up.

Here is the list of all ENV (you can copy the `.env.sample` file to `.env`):

- `LOG_LEVEL` sets to log level of the API (default is warn). Values it can be, are: trace, debug, info, warn, error, fatal, panic. **Logs could leak credentials if set to debug or lower**
- `REDIS_URL` sets the redis host:port url like "redis:6379"
- `REDIS_PASSWORD` sets the redis password
- `REDIS_DB` sets the default redis DB (default is 0)
- `SEED_DATA` can be set to false (default is true), if you don't want to seed the CSV into the DB
- `SKIP_INVALID_HOTELS` can be set to false (default is true), if you want to skip offers from the offers.csv during seeding if they reference a hotel that is **not** in the hotels.csv
- `OFFERS_CSV_PATH` sets the path of the offers.csv file
- `POSTGRES_HOST` sets the postgres host (the port is 5432)
- `POSTGRES_USER` sets the postgres user
- `POSTGRES_DB` sets the postgres DB
- `POSTGRES_PASSWORD` sets the postgres password

<!-- TOC --><a name="seeding-the-data"></a>
#### Seeding the data

The CSV's are read into a buffer and than processed in batches. These batches will then be sent to the DB.
This process can take up to **14 Hours** to fully migrate the data.

The data will be upserted into the DB.


<!-- TOC --><a name="api-defintion"></a>
#### API defintion

There is a generated API definition with a swagger UI site to test the API.
The swagger UI is reachable under `/swagger/index.html` (http://localhost:8080/swagger/index.html)

<!-- TOC --><a name="swag-code-gen"></a>
#### Swag code gen

If there are changes to the swaggo comments, please regenerate the generation of the swagger docs.

Install swaggo:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Run this command to (re-)generate the swag docs:

```bash
swag init -g api/api.go --parseInternal=true -o api/docs/
```

And run this command to reformat the comments:

```bash
swag fmt -g api/api.go
```

<!-- TOC --><a name="goent-orm-generation"></a>
#### GOENT (ORM) Generation

If you are doing local development without the docker compose, you'll need to run the following to generate the database definitions etc.

```bash
go install entgo.io/ent/cmd/ent@v0.14.5
```

```bash
go run -mod=mod entgo.io/ent/cmd/ent --target=./db/ent --feature sql/upsert generate ./db/schema
```

<!-- TOC --><a name="frontend"></a>
### Frontend

The frontend is configured in the docker-compose.yml to listen on port :8090 by default.

> [!IMPORTANT]
> **Navigate into the frontend folder and run the following commands...**

To install all dependencies:

```bash
npm i
```

To run the frontend just run:

```bash
npm run dev
```

<!-- TOC --><a name="regenerate-the-openapi-specs-for-the-frontend"></a>
#### (Re)Generate the Openapi Specs for the frontend

to regenerate the `openapi.d.ts`

```bash
npx  openapi-client-axios-typegen ../api/docs/swagger.json > app/types/openapi.d.ts
```

<!-- TOC --><a name="decisions-ive-made"></a>
## Decisions I've Made

<!-- TOC --><a name="postgresql"></a>
### Postgresql

I'm using postgresql because it is very scalable and i'm used to it.

I also considered using bbolt (or even redis) with buckets. This would allow for an in memory DB, that could be outsourced/loaded to the hard drive (SSD/HDD). I have not implemented this, because I was using postgres first and there is not enough time for the bbolt approach.

The data structure would be like this:

- i would create a bucket for each hotel with the hotel id as the key
- then in every hotel bucket i would create a bucket for each airport
- in each airport i would then create more buckets for every (available) number of adults
- in these buckets i would create buckets for every (available) number of children
- in these buckets I would create more buckets with either the duration or the departure date

The buckets would be created during the initial reading process.

With those buckets I would have a relative great key to search for offers then.

For example:

```text
/bbolt
├── hotel:123
│   ├── airport:BER
│   │   ├── countAdults:1
│   │   │   ├── countChildren:0
│   │   │   │   ├── duration:5
│   │   │   │   ├── duration:6
│   │   │   │   └── duration:7
│   │   │   ├── countChildren:1
│   │   │   │   ├── duration:7
│   │   │   │   └── duration:8
│   │   │   └── countChildren:2
│   │   │       └── duration:14
│   │   ├── countAdults:2
│   │   │   ├── countChildren:0
│   │   │   └── countChildren:2
│   │   └── countAdults:3
│   │       └── countChildren:0
│   ├── airport:MUC
│   └── airport:FRA
```

This approach would have some issues tho, because at some point you have to search linear (but this would also be the
case in the DB implementation) depending on the request. Also sorting is a bit more tricky, but doable.

<!-- TOC --><a name="csv-to-db-migration"></a>
### CSV to DB Migration

I want to search for offers more robust than searching the CSV file linear on an HDD or trying to get it into my computer memory.
That's why I'm migrating everything to a Postgresql DB, where I can search more easily.

The entries of the Offers match the lines of the CSV file (minus one because I did not migrate the header).

`SELECT count(*) FROM offers;` returns 92783430
and
`grep -c ^ offer.csv` returns 92783431

This will also migrate EVERY offer, even if the corresponding Hotel(id) is not present in the hotels.csv.

To tweak this I introduced the `SKIP_INVALID_HOTELS` env. If set to `true`, it will skip offers, that reverence a hotel
that is not in the hotels.csv. This will also reduce the offer amount by ~80%.

<!-- TOC --><a name="seeding-the-data-only-the-first-time"></a>
#### Seeding the data only the first time

I have implemented another env that can control whether you want to seed all the data when firing up the docker compose
or not.

By setting the env `SEED_DATA` to `true` or `false`, it will either read the CSV's and migrate/upsert the data or not.

<!-- TOC --><a name="using-the-service-with-the-whole-offers-csv"></a>
#### Using the Service with the whole offers csv

If you set the ENV to `SKIP_INVALID_HOTELS=true`, all entries of the offers.csv will be in the DB.
That means there will be over 92 Million offers. If set to false it will be "only" 18592896 offers.
Also, all the hotels, that are not in the CSV, will be inserted with a placeholder name.

This is affecting the query time intensively If there are 92 Million, a larger timeframe request could take up to 14 seconds.
If there are only 18 Million, it will be around 1.1 seconds and after caching (same for the 92 Million) ~15 ms.

<!-- TOC --><a name="possible-improvements"></a>
### Possible Improvements

<!-- TOC --><a name="checking-each-param-individually"></a>
#### Checking each Param individually

For the Hotel Offers request, where I made the decision that some of the params are optional, I should check them individually.
Currently I'm not doing this. I've made this decision because I looked at the frontend example after writing the backend.
After that I saw how the Minimal requirements Number 2 is meant (I thought these were individually requests).

But another problem arrising here is that the frontend will send the param with `0` if the param is not set (e.g. countchildren) so I can't distinguish between not set or 0.
This is not really a problem but you can't query the API for requests if you want offers where coundchildren could be anything (which might also be intended).

<!-- TOC --><a name="using-a-key-value-store-only"></a>
#### Using a key-value store only

Just like discriped prior, I was thinking about using bbolt as the DB.

<!-- TOC --><a name="query-queue"></a>
#### Query Queue

The implementation of the Query Queue feature could lead to more wait time under heavy load. But I did not check for this and I like the feature.
With the Query Queue, I use the cached object right away without querying the DB twice.

<!-- TOC --><a name="use-something-else-then-swaggo"></a>
#### Use something else then swaggo

With the provided API definition from the frontend, I could have generated some code for my API backend.
Also, swaggo only generates swagger 2.0 definitions.

In another project [max0l/nobody-is-perfect-go](https://github.com/max0l/nobody-is-perfect-go), I have used oapi-codegen. This introduced some other issues tho.

<!-- TOC --><a name="seeding-the-data-1"></a>
#### Seeding the data

Seeding the data takes really long. But I don't really know how to improve that.

<!-- TOC --><a name="deployment"></a>
## Deployment

I have deployed my service to my VPS. It's an ARM64, 6 Cores, 8GB Ram. But it was effective 4GB of RAM and at first I could not seed the data because I did run out of memory. After setting the batchSize to 10000 and the maxConcurrency to 2, it worked.

I have migrated only the offers that reference a hotel from the `hotels.csv`.

To do that I have modified the docker compose slightly:

```yaml
services:
  db:
    image: postgres:17-alpine3.22
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-password}
      POSTGRES_DB: ${POSTGRES_DB:-check24}
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data
    restart: always
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "${POSTGRES_USER:-user}", "-d", "${POSTGRES_DB:-check24}"]
      interval: 5s
      timeout: 4s
      retries: 5
    networks:
      - backend

  seed_data:
    build:
      context: .
      dockerfile: ./cmd/migrate/Dockerfile
    depends_on:
      db:
        condition: service_healthy
    environment:
      DATABASE_DSN: host=${POSTGRES_HOST:-db} port=5432 user=${POSTGRES_USER:-user} dbname=${POSTGRES_DB:-check24} password=${POSTGRES_PASSWORD:-password} sslmode=disable
      LOG_LEVEL: ${LOG_LEVEL:-warn}
      SEED_DATA: ${SEED_DATA:-true}
      SKIP_INVALID_HOTELS: ${SKIP_INVALID_HOTELS:-true}
    command: ["-hotel", "/hotels.csv", "-offer", "/offers.csv"]
    restart: "no"
    volumes:
      - ${OFFERS_CSV_PATH:?offers.csv is required}:/offers.csv
      - ./data/hotels.csv:/hotels.csv
    networks:
      - backend

  api:
    build:
      context: .
      dockerfile: ./cmd/api/Dockerfile
    depends_on:
      db:
        condition: service_healthy
      seed_data:
        condition: service_completed_successfully
    environment:
      DATABASE_DSN: host=${POSTGRES_HOST:-db} port=5432 user=${POSTGRES_USER:-user} dbname=${POSTGRES_DB:-check24} password=${POSTGRES_PASSWORD:-password} sslmode=disable
      LOG_LEVEL: ${LOG_LEVEL:-warn}
      REDIS_URL: ${REDIS_URL:-redis:6379}
      REDIS_PASSWORD: ${REDIS_PASSWORD:-password}
      REDIS_DB: ${REDIS_DB:-0}
    expose:
      - "8080:8080"
    restart: always
    healthcheck:
      test: curl -f http://localhost:8080/ping
      interval: 60s
      timeout: 5s
      retries: 5
      start_period: 3s
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.checkswag.tls.certresolver=netcup"
      - "traefik.http.routers.checkswag.rule=Host(`check24-swagger.max0l.dev`)"
      - "traefik.http.routers.checkswag.entrypoints=https"
      - "traefik.http.routers.checkswag.tls=true"
      - "traefik.http.services.checkswag.loadbalancer.server.port=8080"
    networks:
      - proxy
      - backend

  redis:
    image: redis:8-alpine
    ports:
      - "6379:6379"
    restart: always
    command: redis-server --save 20 1 --loglevel warning --requirepass ${REDIS_PASSWORD:-password}
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 4s
      retries: 5
    networks:
      - backend

  frontend:
    build:
      context: ./frontend
      dockerfile: ./Dockerfile
    restart: always
    depends_on:
      api:
        condition: service_started
    healthcheck:
      test: wget --spider -q http://127.0.0.1:3000 || exit 1
      interval: 60s
      retries: 5
      start_period: 20s
      timeout: 10s
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.check.tls.certresolver=netcup"
      - "traefik.http.routers.check.rule=Host(`check24.max0l.dev`)"
      - "traefik.http.routers.check.entrypoints=https"
      - "traefik.http.routers.check.tls=true"
      - "traefik.http.services.check.loadbalancer.server.port=3000"
    networks:
      - proxy
      - backend

volumes:
  db_data:

networks:
  proxy:
    external: true
  backend:
    external: true
```
