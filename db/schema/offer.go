package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Offer represents the offer entity in the database.
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
		field.Int("durationInNights"),
		field.Int("hotelid"),
	}
}

func (Offer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("hotel", Hotel.Type).
			Ref("offers").
			Field("hotelid").
			Unique().
			Required(),
	}
}

func (Offer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(
			"outbounddepartureairport",
			"durationInNights",
			"countadults",
			"countchildren",
			"price",
		),

		index.Fields("outbounddeparturedatetime", "inboundarrivaldatetime"),
	}
}
