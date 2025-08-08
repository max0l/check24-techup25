package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Offer struct {
	ent.Schema
}

func (Offer) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Unique(),
		field.Time("outbounddeparturedatetime"),
		field.Time("inbounddeparturedatetime"),
		field.Int("countadults"),
		field.Int("countchildren"),
		field.Float32("price"),
		field.String("inbounddepartureairport"),
		field.String("inboundarrivalairport"),
		field.Time("inboundarrivaldatetime"),
		field.String("outbounddepartureairport"),
		field.String("outboundarrivalairport"),
		field.Time("outboundarrivaldatetime"),
		field.String("mealtype"),
		field.Bool("oceanview"),
		field.String("roomtype"),
		field.Int("durationInDays"),
	}
}

func (Offer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("hotelid", Hotel.Type).
			Ref("offers").
			Unique(),
	}
}

func (Offer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").Unique(),
		index.Fields("price", "outbounddepartureairport", "durationInDays", "countadults", "countchildren"),
	}
}
