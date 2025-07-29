package data

type Hotel struct {
	Hotelid    int     `csv:"hotelid"`
	Hotelname  string  `csv:"hotelname"`
	Hotelstars float32 `csv:"hotelstars"`
}
