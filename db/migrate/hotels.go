package migrate

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
)

func MigrateHotels(ctx context.Context, client *ent.Client, hotels <-chan *data.Hotel) {
	for hotel := range hotels {
		log.Trace().Msg("Sending a hotel to the db")
		client.Hotel.Create().
			SetHotelname(hotel.Hotelname).
			SetID(hotel.Hotelid).
			SetHotelstars(hotel.Hotelstars).
			OnConflict(
				sql.ConflictColumns("id"),
				sql.ResolveWithNewValues()).
			UpdateNewValues().
			ExecX(ctx)
	}
}
