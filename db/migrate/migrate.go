package migrate

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"
	"entgo.io/ent/dialect/sql"
)

func MigrateHotels(ctx context.Context, client *ent.Client, hotels <-chan *data.Hotel) {
	for hotel := range hotels {
		client.Hotel.Create().
			SetHotelname(hotel.Hotelname).
			SetHotelid(hotel.Hotelid).
			SetHotelstars(hotel.Hotelstars).
			OnConflict(
				sql.ConflictColumns("hotelid"),
				sql.ResolveWithNewValues()).
			UpdateNewValues().
			ExecX(ctx)
	}
}

func MigrateOffers(ctx context.Context, client *ent.Client, offers chan *data.Offer) {
	for offer := range offers {
		client.Offer.Create().
			SetHotelid(offer.Hotelid).
			SetOutbounddeparturedatetime(offer.Outbounddeparturedatetime).
			SetInbounddeparturedatetime(offer.Inbounddeparturedatetime).
			SetCountadults(offer.Countadults).
			SetCountchildren(offer.Countchildren).
			SetPrice(offer.Price).
			SetInbounddepartureairport(offer.Inbounddepartureairport).
			SetInboundarrivalairport(offer.Inboundarrivalairport).
			SetInboundarrivaldatetime(offer.Inboundarrivaldatetime).
			SetOutbounddepartureairport(offer.Outbounddepartureairport).
			SetOutboundarrivalairport(offer.Outboundarrivalairport).
			SetOutboundarrivaldatetime(offer.Outboundarrivaldatetime).
			SetMealtype(offer.Mealtype).
			SetOceanview(offer.Oceanview).
			SetRoomtype(offer.Roomtype).
			OnConflict(
				sql.ConflictColumns("hotelid", "outbounddeparturedatetime", "inbounddeparturedatetime", "roomtype"),
				sql.ResolveWithNewValues()).
			UpdateNewValues().
			ExecX(ctx)
	}
}
