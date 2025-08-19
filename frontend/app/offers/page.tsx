"use client";

import { OpenAPIClientAxios, Parameters } from "openapi-client-axios";
import { Client, Definitions, Paths } from "@/app/types/openapi"
import HotelOffer from "@/app/components/HotelOffer/HotelOffer";
import { useRouter, useSearchParams } from 'next/navigation';
import { useEffect, useState } from "react";
import { Box, Rating, Typography } from "@mui/material";
import { Stack } from "@mui/system";
import { GetHotelOffersFromQuery, GetHotelIdFromQuery, GetHotelOffersToQuery } from "@/app/types/converter";
import qs from "qs";
import PaginationBar from "@/app/components/Pagination/PaginationBar";


type HotelOfferResp = Definitions.DataHotelResponse

//type GetHotelResp   = Paths.Hotel$Hotelid.Get.Responses.$200;

export default function Page() {
    const query = useSearchParams()
    const [hotelOffer, setHotelOffer] = useState<HotelOfferResp>();
    const [queryParameters, setQueryParameters] = useState<Paths.Hotel$Hotelid.Get.QueryParameters>()
    const router = useRouter()

    async function fetchData(parameters: Paths.Hotel$Hotelid.Get.QueryParameters) {
        setQueryParameters(parameters)
        const hotelid = GetHotelIdFromQuery(query)
        router.push("/offers?" + GetHotelOffersToQuery(parameters, hotelid));
        const api = new OpenAPIClientAxios({
            definition: 'http://localhost:8080/swagger/doc.json', withServer: 0,
            axiosConfigDefaults: {
                baseURL: 'http://localhost:8080',
                paramsSerializer: { serialize: p => qs.stringify(p, { arrayFormat: "comma" }) }
            }
        })
        const client = await api.init<Client>()
        const response = await client.get<HotelOfferResp>(`/hotel/${hotelid.hotelid}`, { params: parameters });
        setHotelOffer(response.data)
    }

    useEffect(() => {
        const params = GetHotelOffersFromQuery(query);
        fetchData(params).catch(console.error);
    }, []);

    if (hotelOffer == null) {
        return <p>Loading offers...</p>
    }

    const PAGE_SIZE = 50;

    const total = hotelOffer.count_available_offers ?? 0;
    const page = queryParameters?.site || 1;

    const handlePageChange = (p: number) => {
        if (!queryParameters) return;
        window.scrollTo({ top: 0, behavior: "smooth" });
        fetchData({ ...queryParameters, site: p });
    };


    return (
        <>
            <Stack direction="row" pt={5} alignItems="center">
                <Typography variant="h4" mr={2}>{hotelOffer.hotelname ?? "NaN"}: {hotelOffer.hotelid}</Typography>
                <Rating value={hotelOffer.hotelstars ?? 0} readOnly />
            </Stack>
            <Stack direction="row" height="250px" pt={2}>
                <Box sx={{ borderTopLeftRadius: "5px", borderBottomLeftRadius: "5px", backgroundImage: `url("/hotels/${((hotelOffer.hotelid ?? 0) % 40) + 1}.jpg")`, width: "50%", backgroundSize: "cover", backgroundPosition: "center" }} />
                <Box sx={{ borderTopRightRadius: "5px", borderBottomRightRadius: "5px", backgroundImage: `url("/rooms/${((hotelOffer.hotelid ?? 0) % 30) + 1}.jpg")`, width: "50%", backgroundSize: "cover", backgroundPosition: "center" }} />
            </Stack>
            <Typography pt={2} variant="h6">Offers:</Typography>
            {total > 0 && (
                <PaginationBar
                    page={page}
                    totalItems={total}
                    pageSize={PAGE_SIZE}
                    onChange={handlePageChange}
                />
            )}
            <Stack gap={2} mt={1}>
                {hotelOffer?.offers?.map(offer =>
                    <HotelOffer key={offer.offerid} offer={offer} />
                )}
            </Stack>
        </>
    )
}