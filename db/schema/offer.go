package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Offer struct {
	ent.Schema
}

func (Offer) Fields() []ent.Field {
	return []ent.Field{
		field.Int("hotelid"),
		field.Time("outbounddeparturedatetime"),
		field.Time("inbounddeparturedatetime"),
		field.Int("countadults"),
		field.Int("countchildren"),
		field.Float("price"),
		field.String("inbounddepartureairport"),
		field.String("inboundarrivalairport"),
		field.Time("inboundarrivaldatetime"),
		field.String("outbounddepartureairport"),
		field.String("outboundarrivalairport"),
		field.Time("outboundarrivaldatetime"),
		field.String("mealtype"),
		field.Bool("oceanview"),
		field.String("roomtype"),
	}
}
