# Check24 Techup25

## Development

### Swag code gen

Install swaggo:
```
go install github.com/swaggo/swag/cmd/swag@latest
```

Run this command to (re-)generate the swag docs:
```
swag init -g api/api.go --parseInternal=true -o api/docs/
```

And run this command to reformat the comments:
```
swag fmt -g api/api.go
```

## Decisions I've Made

### CSV to DB Migration

I want to search for offers more robust then searching the CSV file linear on an HDD or trying to get it into my computer memory.
That's why I'm migrating everything to a Postgresql DB, were I can search more easily.

The entries of the Offers match the lines of the CSV file (minus one because I did not migrate the header).

`SELECT count(*) FROM offers;` returns 92783430
and
`grep -c ^ offer.csv` returns 92783431