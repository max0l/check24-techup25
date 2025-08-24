package migrate

import (
	"check24-techup25/data"
	"check24-techup25/db/ent"
	"context"
	"sync"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"
)

// KnownHotels stores known hotels by their ID in a map.
type KnownHotels struct {
	sync.RWMutex
	hotels map[int]ent.Hotel
}

// NewKnownHotels creates a new KnownHotels instance.
func NewKnownHotels() *KnownHotels {
	return &KnownHotels{
		hotels: make(map[int]ent.Hotel),
	}
}

// AddHotel adds a hotel to the KnownHotels map.
func (kh *KnownHotels) AddHotel(hotel *ent.Hotel) {
	kh.Lock()
	defer kh.Unlock()
	kh.hotels[hotel.ID] = *hotel
}

// ContainsHotel checks if a hotel with the given ID is in the KnownHotels map.
func (kh *KnownHotels) ContainsHotel(hotelID int) bool {
	kh.RLock()
	defer kh.RUnlock()
	_, exists := kh.hotels[hotelID]
	return exists
}

// MigrateHotels reads from the hotels channel and sends the hotels to the database in batches.
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
