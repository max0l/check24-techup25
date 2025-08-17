"use client";

import SearchForm from "@/app/components/SearchForm/SearchForm";
import {Stack, Typography} from "@mui/material";
import Hotel from "@/app/components/Hotel/Hotel";
import {OpenAPIClientAxios} from "openapi-client-axios";
import {Client, Definitions, Paths} from "@/app/types/openapi";
import {useEffect, useState} from "react";
import Link from "next/link";
import {useRouter, useSearchParams} from "next/navigation";
import BestHotelOffer = Definitions.DataOfferResponse
import {GetBestOffersByHotelFromQuery, GetBestOffersByHotelToQuery} from "@/app/types/converter";
import qs from "qs";

export default function HomePage() {
    const [offers, setOffers] = useState<BestHotelOffer>();
    const [queryParameters, setQueryParameters] = useState<Paths.Offers.Get.QueryParameters>();
    const router = useRouter();
    const query = useSearchParams();

    // update the search form and automatically load offers if a search is existing
    useEffect(() => {
        const parameters = GetBestOffersByHotelFromQuery(query);
        // parameters should be validated here, but as this is a just a very simple implementation we skip this for now
        if(parameters.earliestDepartureDate == null) {
            return;
        }

        load(parameters).catch(console.error);
    }, []);


    async function onSubmitSearchForm(departureAirports: string[], countAdults: number, countChildren: number, duration: number, earliestDeparture: string, latestReturn: string) {
        const parameters: Paths.Offers.Get.QueryParameters = {
            earliestDepartureDate: earliestDeparture,
            latestReturnDate: latestReturn,
            countAdults: countAdults,
            countChildren: countChildren,
            departureAirports: departureAirports,
            duration: duration,
        };

        await load(parameters);
    }

    async function load(parameters: Paths.Offers.Get.QueryParameters) {
        setQueryParameters(parameters);
        router.push("/?" + GetBestOffersByHotelToQuery(parameters));
        const api = new OpenAPIClientAxios({definition: 'http://localhost:8080/swagger/doc.json', withServer: 0,
            axiosConfigDefaults: {baseURL: 'http://localhost:8080', 
                paramsSerializer: {serialize: p => qs.stringify(p, {arrayFormat: "comma"})}}
        })
        const client = await api.init<Client>()
        const response = await client.get<BestHotelOffer>(`/offers`, {params: parameters });
        setOffers(response.data);
    }

    return (
        <>
            <Typography sx={{mb: "50px", mt: "40px"}} variant="h3">CHECK24 Holiday Challenge</Typography>
            <SearchForm submitCallback={onSubmitSearchForm}/>
            <Typography variant="h4" sx={{mt: "60px", mb: "30px"}}>Hotels for your Mallorca-Trip:</Typography>
            <Stack gap={3}>
                {offers?.offers?.map(offer =>
                    <Link key={offer.hotelid}
                          href={{pathname: '/offers', query: {...queryParameters, hotelId: offer.hotelid}}}
                          style={{textDecoration: "none"}}>
                        <Hotel offer={offer}/>
                    </Link>
                )}
            </Stack>
        </>
    )
}