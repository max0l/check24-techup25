import { Button, Card, CardContent, CardHeader, Divider, Typography } from "@mui/material";
import { Stack } from "@mui/system";
import Flight from "@/app/components/Flight/Flight";
import {Definitions} from "@/app/types/openapi"
import { Bed, RestaurantMenu, Water } from "@mui/icons-material";
type Offer = Definitions.DataOffer

export default function HotelOffer({offer}: {offer: Offer}) {
    function getTravelDurationString(departure: string | undefined, arrival: string | undefined) : string {
        if(!departure || !arrival) {
            return "";
        }

        const date1 = new Date(arrival);
        const date2 = new Date(departure);
        const difference = Math.abs(date1.getTime() - date2.getTime());
        return Math.ceil(difference / (1000 * 3600 * 24)).toString();
    }

    return (
        <Card>
            <CardHeader sx={{backgroundColor: "#ededed"}} title={<Typography fontWeight="bold">{offer.durationindays} Days - 
                {offer.outbounddepartureairport}, {offer.outboundarrivalairport}</Typography>}/>
            <CardContent>
                <Stack direction="row" justifyContent="space-between">
                    <Stack gap={2}>
                        <Flight 
                            inbound={true} 
                            departureDatetime={offer.outbounddeparturedatetime}
                            departureAirport={offer.outbounddepartureairport}
                            arrivalDatetime={offer.outboundarrivaldatetime}
                            arrivalAirport={offer.outboundarrivalairport}
                        />
                        <Flight 
                            inbound={false} 
                            departureDatetime={offer.inbounddeparturedatetime}
                            departureAirport={offer.inbounddepartureairport}
                            arrivalDatetime={offer.inboundarrivaldatetime}
                            arrivalAirport={offer.inboundarrivalairport}
                        />
                    </Stack>
                    <Stack gap={2}>
                        <Stack direction="row" alignItems="center">
                            <RestaurantMenu/>
                            <Typography ml={1} variant="body1">{offer.mealtype}</Typography>
                        </Stack>
                        <Stack direction="row" alignItems="center">
                            <Bed/>
                            <Typography ml={1} variant="body1">{offer.roomtype}</Typography>
                        </Stack>
                        {offer.oceanview && 
                            <Stack direction="row" alignItems="center">
                                <Water/>
                                <Typography ml={1} variant="body1">Oceanview</Typography>
                            </Stack>
                        }
                    </Stack>
                    <Stack justifyContent="end" gap={2}>
                        <Stack m={0} direction="row" divider={<Divider orientation="vertical" flexItem/>} spacing={1}>
                            <Typography variant="body1">{offer.durationindays} Nights</Typography>
                            <Typography variant="body1">{offer.countadults} Adults</Typography>
                            <Typography variant="body1">{offer.countchildren} Children</Typography>
                        </Stack>
                        <Typography variant="h5" textAlign="right">{offer.price} €</Typography>
                        <Button variant="contained">Book</Button>
                    </Stack>
                </Stack>
            </CardContent>
        </Card>
    )
}