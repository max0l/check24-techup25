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
			SetID(hotel.Hotelid).
			SetHotelstars(hotel.Hotelstars).
			OnConflict(
				sql.ConflictColumns("id"),
				sql.ResolveWithNewValues()).
			UpdateNewValues().
			ExecX(ctx)
	}
}

const batchSize = 1000

func MigrateOffers(ctx context.Context, client *ent.Client, offers chan *data.Offer) {
	var (
		offerBatch []*ent.OfferCreate
		err        error
	)
	for offer := range offers {
		hotel, err := client.Hotel.Get(ctx, offer.Hotelid)
		if err != nil {
			log.Warn().Err(err).Msgf("Could not find hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
			continue
		}
		offerDb := client.Offer.Create().
			SetOfferid(offer.Offerid).
			SetHotelid(hotel).
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
			SetDurationInDays(offer.DurationInDays)

		offerBatch = append(offerBatch, offerDb)
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
				log.Debug().Msgf("Sent %v offers to the DB", len(offerBatch))
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
