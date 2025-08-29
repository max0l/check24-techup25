package migrate

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"
	"errors"
	"fmt"
	"os"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
)

// MigrateOffers migrates offers from the given channel to the database in batches.
func MigrateOffers(ctx context.Context, client *ent.Client, knownHotels *KnownHotels,
	offers chan *data.Offer) {
	var (
		offerBatch []*ent.OfferCreate
		err        error
	)
	for offer := range offers {

		// This will make the insertion a little bit faster
		hotel, err := handleOfferHotel(ctx, client, offer, knownHotels)
		if err != nil {
			if !errors.Is(err, data.ErrHotelIsNotInCSV) {
				log.Error().Err(err).Msgf("Could not handle hotel for offer %v", offer.Offerid)
			}
			continue
		}

		offerDb := createEntOffer(client, offer, hotel)

		offerBatch = append(offerBatch, offerDb)
		if len(offerBatch) == batchSize {
			log.Debug().Msgf("Sending %v offers to the DB", len(offerBatch))
			sendBulkOffersToDb(ctx, client, err, offerBatch)
			offerBatch = offerBatch[:0]
		}
	}

	if len(offerBatch) > 0 && len(offerBatch) < batchSize {
		log.Debug().Msgf("Sending last %v offers to the DB", len(offerBatch))
		sendBulkOffersToDb(ctx, client, err, offerBatch)
	}
}

// handleOfferHotel checks if the hotel exists for the given offer and creates it if it does not exist.
// If the hotel ID is greater than 998, it will skip the creation if the environment variable SKIP_INVALID_HOTELS is set to true.
// The number 998 is used because querying the DB for every hotel will slow down the migration process significantly.
func handleOfferHotel(ctx context.Context, client *ent.Client, offer *data.Offer,
	knownHotels *KnownHotels) (*ent.Hotel, error) {
	if doSkipInvalidHotels() && offer.Hotelid > 998 {
		log.Trace().Msgf("Skipping hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
		return nil, data.ErrHotelIsNotInCSV
	}

	if offer.Hotelid > 998 && !seenHotel(offer.Hotelid, knownHotels) {
		log.Trace().Msgf("Creating hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
		err := client.Hotel.Create().
			SetID(offer.Hotelid).
			SetHotelname(fmt.Sprintf("Placeholder for %v", offer.Hotelid)).
			SetHotelstars(0).
			OnConflict(sql.ConflictColumns("id"), sql.ResolveWithNewValues()).
			UpdateNewValues().
			Exec(ctx)
		if err != nil {
			log.Error().Err(err).Msgf("Could not create hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
			return nil, err
		}
	}

	hotel, err := client.Hotel.Get(ctx, offer.Hotelid)
	if err != nil {
		log.Warn().Err(err).Msgf("Could not find hotel with ID %d for offer %d", offer.Hotelid, offer.Offerid)
		return nil, err
	}
	knownHotels.AddHotel(hotel)
	return hotel, nil
}

func seenHotel(hotelid int, knownHotels *KnownHotels) bool {
	if knownHotels.ContainsHotel(hotelid) {
		log.Trace().Msgf("Hotel with ID %d already exists in known hotels", hotelid)
		return true
	}
	log.Trace().Msgf("Hotel with ID %d does not exist in known hotels, adding it", hotelid)
	return false
}

// doSkipInvalidHotels checks if the environment variable SKIP_INVALID_HOTELS is set to true.
// If not set, it defaults to false.
func doSkipInvalidHotels() bool {
	skipInvalidHotels := os.Getenv("SKIP_INVALID_HOTELS")
	if skipInvalidHotels == "" {
		return false
	}
	if skipInvalidHotels == "true" {
		return true
	}
	return false
}

// createEntOffer creates an ent.OfferCreate object from the given data.Offer and ent.Hotel.
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
		SetDurationInNights(offer.DurationInDays)
	return offerDb
}

// sendBulkOffersToDb sends a batch of offers to the database.
func sendBulkOffersToDb(ctx context.Context, client *ent.Client, err error, offerBatch []*ent.OfferCreate) {
	err = client.Offer.CreateBulk(offerBatch...).
		OnConflict(
			sql.ConflictColumns("id"),
			sql.ResolveWithNewValues(),
		).
		UpdateNewValues().
		Exec(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Could not send offers")
	} else {
		log.Debug().Msgf("Sent %v offers to the DB", len(offerBatch))
	}
}
