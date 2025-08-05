package data

type Hotel struct {
	Hotelid    int
	Hotelname  string
	Hotelstars float32
}

type HotelResponse struct {
	Hotelid    int
	Hotelname  string
	Hotelstars float32
	Offers     []Offer
}
