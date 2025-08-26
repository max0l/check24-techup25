FROM golang:1.24.5 AS BUILDER

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .

RUN go get entgo.io/ent/cmd/ent@v0.14.5
RUN go run -mod=mod entgo.io/ent/cmd/ent --target=./db/ent --feature sql/upsert,sql/modifier generate ./db/schema

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate

FROM scratch
COPY --from=BUILDER /bin/migrate /migrate

ENTRYPOINT ["/migrate"]