package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Hotel represents the hotel entity in the database.
type Hotel struct {
	ent.Schema
}

func (Hotel) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique(),
		field.String("hotelname"),
		field.Float32("hotelstars"),
	}
}

func (Hotel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("offers", Offer.Type).StorageKey(edge.Column("hotelid")),
	}
}

func (Hotel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").Unique(),
	}
}
