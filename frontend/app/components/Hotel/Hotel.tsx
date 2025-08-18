import {Box, Button, Card, CardMedia, Divider, Rating, Stack, Typography} from "@mui/material";
import {Definitions} from "@/app/types/openapi";
import BestHotelOffer = Definitions.DataOfferWithHotel

interface Properties {
    offer: BestHotelOffer
}

export default function Hotel({offer}: Properties) {
    return (
        <Card variant="outlined" sx={{display: 'flex'}}>
            <Box sx={{backgroundImage: `url("/hotels/${((offer.hotelid ?? 0) % 40) + 1}.jpg")`, width: "355.5px",height: "200px", backgroundSize: "cover"}}/>
            <Box sx={{
                p: 1,
                pl: 2,
                display: 'flex',
                width: '100%',
                flexDirection: 'row',
                justifyContent: 'space-between'
            }}>
                <Stack direction="column" justifyContent="space-between">
                    <Typography sx={{mr: 2}} variant="h6">{offer.hotelname}</Typography>

                    <Stack>
                        <Typography variant="body1">{offer.rommtype}</Typography>
                        <Typography variant="caption">Hotel ID: {offer.hotelid}</Typography>
                    </Stack>

                    {/*TODO: implement {offer.countAvailableOffers} feat*/}
                    <Button variant="contained">View other offers</Button>
                    {/*<Button variant="contained">View {offer.countAvailableOffers} offers</Button>*/}
                </Stack>
                <Stack direction="column" justifyContent="space-between" alignItems="flex-end">
                    <Rating value={offer.hotelstars} readOnly/>

                    <Stack>
                        <Stack m={0} direction="row" divider={<Divider orientation="vertical" flexItem/>}
                               spacing={1}>
                            <Typography>{offer.durationindays} Days</Typography>
                            <Typography>{offer.countadults} Adults</Typography>
                            <Typography>{offer.countchildren} Children</Typography>
                        </Stack>
                        <Typography variant="h6" textAlign="right">from {offer.price} €</Typography>
                    </Stack>
                </Stack>
            </Box>
        </Card>
    );
}