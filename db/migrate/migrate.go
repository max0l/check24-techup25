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
		log.Debug().Msg("Sending a hotel to the db")
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

const batchSize = 10000

func MigrateOffers(ctx context.Context, client *ent.Client, offers chan *data.Offer) {
	var (
		offerBatch []*ent.OfferCreate
		err        error
	)
	for offer := range offers {
		offerBatch = append(offerBatch, client.Offer.Create().
			SetOfferid(offer.Offerid).
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
			SetRoomtype(offer.Roomtype))
		if len(offerBatch) >= batchSize {
			err = client.Offer.CreateBulk(offerBatch...).
				OnConflict(
					sql.ConflictColumns("offerid"),
					sql.ResolveWithNewValues(),
				).
				UpdateNewValues().
				Exec(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Could not send offers")
			} else {
				log.Debug().Msg("Sent 500 offers to the DB")
			}
			offerBatch = offerBatch[:0]
		}
	}

	if len(offerBatch) > 0 {
		err = client.Offer.CreateBulk(offerBatch...).
			OnConflict(
				sql.ConflictColumns("offerid"),
				sql.ResolveWithNewValues(),
			).
			UpdateNewValues().
			Exec(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Could not send last offers")
		}
	}
}
