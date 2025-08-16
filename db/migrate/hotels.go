package migrate

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"
	"sync"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
)

type KnownHotels struct {
	sync.RWMutex
	hotels map[int]ent.Hotel
}

func NewKnownHotels() *KnownHotels {
	return &KnownHotels{
		hotels: make(map[int]ent.Hotel),
	}
}

func (kh *KnownHotels) AddHotel(hotel *ent.Hotel) {
	kh.Lock()
	defer kh.Unlock()
	kh.hotels[hotel.ID] = *hotel
}

func (kh *KnownHotels) ContainsHotel(hotelID int) bool {
	kh.RLock()
	defer kh.RUnlock()
	_, exists := kh.hotels[hotelID]
	return exists
}

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
