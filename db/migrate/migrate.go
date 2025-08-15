package migrate

import (
	"check24-techup25/db/ent"
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
)

const batchSize = 3800

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
