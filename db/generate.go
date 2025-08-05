package db

//go:generate go run -mod=mod entgo.io/ent/cmd/ent --feature sql/modifier --target=./ent --feature sql/upsert generate ./schema
