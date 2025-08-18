"use client";

import SearchForm from "@/app/components/SearchForm/SearchForm";
import { Pagination, Stack, Typography } from "@mui/material";
import Hotel from "@/app/components/Hotel/Hotel";
import { OpenAPIClientAxios } from "openapi-client-axios";
import { Client, Definitions, Paths } from "@/app/types/openapi";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import BestHotelOffer = Definitions.DataOfferResponse;
import {
    GetBestOffersByHotelFromQuery,
    GetBestOffersByHotelToQuery,
} from "@/app/types/converter";
import qs from "qs";
import { coerceAndValidate } from "@/app/utils/validators";

export default function HomePage() {
    const [offers, setOffers] = useState<BestHotelOffer>();
    const [queryParameters, setQueryParameters] =
        useState<Paths.Offers.Get.QueryParameters>();
    const router = useRouter();
    const query = useSearchParams();

    useEffect(() => {
        const raw = GetBestOffersByHotelFromQuery(query);
        const { valid, params } = coerceAndValidate(raw);
        if (!valid) {
            console.log("Query Parameter Error", raw);
            return;
        }
        load(params).catch(console.error);
    }, []);

    async function onSubmitSearchForm(
        departureAirports: string[],
        countAdults: number,
        countChildren: number,
        duration: number,
        earliestDeparture: string,
        latestReturn: string,
        site: number
    ) {
        const { valid, params } = coerceAndValidate({
            earliestDepartureDate: earliestDeparture,
            latestReturnDate: latestReturn,
            countAdults,
            countChildren,
            departureAirports,
            duration,
            site: 1,
        });
        if (!valid) {
            console.log("Form Parameter Error");
            return;
        }
        await load(params);
    }

    async function load(parameters: Paths.Offers.Get.QueryParameters) {
        setQueryParameters(parameters);
        router.push("/?" + GetBestOffersByHotelToQuery(parameters));
        const api = new OpenAPIClientAxios({
            definition: "http://localhost:8080/swagger/doc.json",
            withServer: 0,
            axiosConfigDefaults: {
                baseURL: "http://localhost:8080",
                paramsSerializer: {
                    serialize: (p) => qs.stringify(p, { arrayFormat: "comma" }),
                },
            },
        });
        const client = await api.init<Client>();
        const response = await client.get<BestHotelOffer>(`/offers`, {
            params: parameters,
        });
        setOffers(response.data);
    }

    const href = (qp: Paths.Offers.Get.QueryParameters, hotelid?: number) =>
        `/offers?${qs.stringify(
            { ...qp, hotelId: hotelid },
            { arrayFormat: "comma" }
        )}`;

    return (
        <>
            <Typography sx={{ mb: "50px", mt: "40px" }} variant="h3">
                CHECK24 Techup 25
            </Typography>
            <SearchForm submitCallback={onSubmitSearchForm} />
            <Typography variant="h4" sx={{ mt: "60px", mb: "30px" }}>
                Hotels for your Mallorca-Trip:
            </Typography>
            {(offers?.countOffers ?? 0) > 0 && (
                <Pagination
                    count={Math.ceil((offers?.countOffers ?? 1) / 50)}
                    page={queryParameters?.site ?? 1}   // controlled component
                    onChange={(_, page) => {
                        if (!queryParameters) return;
                        load({ ...queryParameters, site: page });
                    }}
                />
            )}
            <Stack gap={3}>
                {offers?.offers?.map((offer) => (
                    <Link
                        key={offer.hotelid}
                        href={href(queryParameters!, offer.hotelid)}
                        style={{ textDecoration: "none" }}
                    >
                        <Hotel offer={offer} />
                    </Link>
                ))}
            </Stack>
        </>
    );
}
