package migrate

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"

	"github.com/rs/zerolog/log"
)

func MigrateOffers(ctx context.Context, client *ent.Client, offers chan *data.Offer) {
	var (
		offerBatch []*ent.OfferCreate
		err        error
	)
	for offer := range offers {

		// This will make the insertion a little bit faster
		if offer.Hotelid > 998 {
			log.Trace().Err(err).Msgf("Skipping hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
			continue
		}
		hotel, err := client.Hotel.Get(ctx, offer.Hotelid)
		if err != nil {
			log.Warn().Err(err).Msgf("Could not find hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
			continue
		}

		offerDb := createEntOffer(client, offer, hotel)

		offerBatch = append(offerBatch, offerDb)
		if len(offerBatch) == batchSize {
			log.Trace().Msgf("Sending %v offers to the DB", len(offerBatch))
			sendBulkOffersToDb(ctx, client, err, offerBatch)
			offerBatch = offerBatch[:0]
		}
	}

	if len(offerBatch) > 0 && len(offerBatch) < batchSize {
		log.Trace().Msgf("Sending last %v offers to the DB", len(offerBatch))
		sendBulkOffersToDb(ctx, client, err, offerBatch)
	}
}

func createEntOffer(client *ent.Client, offer *data.Offer, hotel *ent.Hotel) *ent.OfferCreate {
	offerDb := client.Offer.Create().
		SetID(offer.Offerid).
		SetHotel(hotel).
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
	return offerDb
}
