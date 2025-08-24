# Check24 Techup25

## Development

### Backend

Everything will be generated if you just run `docker compose up --build`.

> [!IMPORTANT]
> **Set the path of the offers.csv volume in the docker-compose.yml to your offers.csv or to the provided test csv's in the testCSVs folder**

#### ENV

There are several env's that can be set. All of them have a fallback required so it docker compose up can just get fired up.

Here is the list of all ENV (you can copy the `.env.sample` file to `.env`):

- `LOG_LEVEL` sets to log level of the API (default is warn). Values it can be, are: trace, debug, info, warn, error, fatal, panic
- `REDIS_URL` sets the redis host:port url like "redis:6379"
- `REDIS_PASSWORD` sets the redis password
- `REDIS_DB` sets the default redis DB (default is 0)
- `SEED_DATA` can be set to false (default is true), if you don't want to seed the CSV into the DB
- `SKIP_INVALID_HOTELS` can be set to false (default is true), if you want to skip offers from the offers.csv during seeding if they reference a hotel that is **not** in the hotels.csv
- `POSTGRES_HOST` sets the postgres host (the port is 5432)
- `POSTGRES_USER` sets the postgres user
- `POSTGRES_DB` sets the postgres DB
- `POSTGRES_PASSWORD` sets the postgres password

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

#### GOENT (ORM) Generation

If you are doing local development without the docker compose, you'll need to run the following to generate the database definitions etc.

```bash
go install entgo.io/ent/cmd/ent@v0.14.5
```

```bash
go run -mod=mod entgo.io/ent/cmd/ent --target=./db/ent --feature sql/upsert generate ./db/schema
```

### Frontend

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

#### (Re)Generate the Openapi Specs for the frontend

to regenerate the `openapi.d.ts`

```bash
npx  openapi-client-axios-typegen ../api/docs/swagger.json > app/types/openapi.d.ts
```

## Decisions I've Made

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

#### Seeding the data only the first time

I have implemented another env that can control whether you want to seed all the data when firing up the docker compose
or not.

By setting the env `SEED_DATA` to `true` or `false`, it will either read the CSV's and migrate/upsert the data or not.

#### Using the Service with the whole offers csv

If you set the ENV to `SKIP_INVALID_HOTELS=true`, all entries of the offers.csv will be in the DB.
That means there will be over 92 Million offers. If set to false it will be "only" 18592896 offers.
Also, all the hotels, that are not in the CSV, will be inserted with a placeholder name.

This is affecting the query time intensively If there are 92 Million, a larger timeframe request could take up to 14 seconds.
If there are only 18 Million, it will be arround 1.1 seconds and after caching (same for the 92 Million) ~15 ms.
