FROM golang:1.24.5 AS BUILDER

WORKDIR /app

COPY go.mod .

RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go install entgo.io/ent/cmd/ent@v0.14.5

COPY . .

RUN swag init -g api/api.go --parseInternal=true -o api/docs/
RUN go run -mod=mod entgo.io/ent/cmd/ent --target=./db/ent --feature sql/upsert,sql/modifier generate ./db/schema

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine

RUN apk add --no-cache curl

COPY --from=BUILDER /api /api

EXPOSE 8080

ENTRYPOINT ["/api"]