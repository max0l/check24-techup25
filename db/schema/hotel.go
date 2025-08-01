package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Hotel struct {
	ent.Schema
}

func (Hotel) Fields() []ent.Field {
	return []ent.Field{
		field.Int("hotelid").Unique(),
		field.String("hotelname"),
		field.Float32("hotelstars"),
	}
}
