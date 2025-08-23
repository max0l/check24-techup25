package data

type Hotel struct {
	Hotelid    int
	Hotelname  string
	Hotelstars float32
}

type HotelOffersResponse struct {
	Hotelid              int     `json:"hotelid"`
	Hotelname            string  `json:"hotelname"`
	Hotelstars           float32 `json:"hotelstars"`
	Offers               []Offer `json:"offers"`
	CountAvailableOffers int     `json:"count_available_offers"`
}
