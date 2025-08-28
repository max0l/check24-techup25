"use client";

import SearchForm from "@/app/components/SearchForm/SearchForm";
import { Stack, Typography, Box, CircularProgress, LinearProgress } from "@mui/material";
import Hotel from "@/app/components/Hotel/Hotel";
import { OpenAPIClientAxios } from "openapi-client-axios";
import { Client, Definitions, Paths } from "@/app/types/openapi";
import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import BestHotelOffer = Definitions.DataBestOfferPerHotelResponse;
import {
    GetBestOffersByHotelFromQuery,
    GetBestOffersByHotelToQuery,
} from "@/app/types/converter";
import qs from "qs";
import { coerceAndValidate } from "@/app/utils/validators";
import {extractAxiosErrorMessage} from "@/app/utils/axiosExtractor"
import PaginationBar from "@/app/components/Pagination/PaginationBar";

const apiBase =
    typeof window === "undefined"
        ? process.env.INTERNAL_API_BASE_URL ?? "http://api:8080"
        : process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";

export default function HomePage() {
    const [offers, setOffers] = useState<BestHotelOffer>();
    const [queryParameters, setQueryParameters] =
        useState<Paths.Offers.Get.QueryParameters>();
    const router = useRouter();
    const query = useSearchParams();

    const [displayPage, setDisplayPage] = useState<number>(1);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [requestError, setRequestError] = useState<string>();
    const PAGE_SIZE = 50;
    const totalOfferAmount = offers?.countOffers ?? 0;

    const [lastRequestTime, setLastRequestTime] = useState<number>();

    const apiClient = useMemo(() => {
        const api = new OpenAPIClientAxios({
            definition: `/swagger/doc.json`,
            withServer: { url: apiBase },
            axiosConfigDefaults: {
                baseURL: apiBase,
                paramsSerializer: {
                    serialize: (p) => qs.stringify(p, { arrayFormat: "comma" }),
                },
            },
        });
        return api;
    }, [])

    useEffect(() => {
        const raw = GetBestOffersByHotelFromQuery(query);
        const { valid, params } = coerceAndValidate(raw);
        if (!valid) {
            console.log("Query Parameter Error", raw);
            return;
        }
        setDisplayPage(params.site || 1);
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
        if (isLoading) return;
        setIsLoading(true);
        setRequestError(undefined)
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
            setIsLoading(false)
            return;
        }
        await load(params);
    }

    async function load(parameters: Paths.Offers.Get.QueryParameters) {
        setRequestError(undefined)
        try {
            const start = performance.now();

            const client = await apiClient.init<Client>();
            const response = await client.get<BestHotelOffer>(`/offers`, {
                params: parameters,
            });

            const end = performance.now();
            const durationMs = end - start;
            setLastRequestTime(durationMs);

            setOffers(response.data);
            setQueryParameters(parameters);
            setDisplayPage(parameters.site ?? 1);

            router.push("/?" + GetBestOffersByHotelToQuery(parameters));
        } catch (err) {
            setOffers(undefined);
            setLastRequestTime(undefined);
            setRequestError(extractAxiosErrorMessage(err));
        } finally {
            setIsLoading(false);
        }
    }

    const href = (qp: Paths.Offers.Get.QueryParameters, hotelid?: number) =>
        `/offers?${qs.stringify(
            { ...qp, site: 1, hotelId: hotelid },
            { arrayFormat: "comma" }
        )}`;

    const handlePageChange = (p: number) => {
        if (!queryParameters || isLoading || p == displayPage) return;
        window.scrollTo({ top: 0, behavior: "smooth" });
        load({ ...queryParameters, site: p });
    };
    return (
        <>
            {isLoading && <LinearProgress />}

            <Typography sx={{ mb: "50px", mt: "40px" }} variant="h3">
                CHECK24 Techup 25
            </Typography>

            <SearchForm submitCallback={onSubmitSearchForm} />

            <Typography variant="h4" sx={{ mt: "40px", mb: "10px" }}>
                Hotels for your Mallorca-Trip:
            </Typography>

            {!isLoading && requestError && (
                <Typography color="error" sx={{ mt: 2 }}>
                    Something went wrong: {requestError}
                </Typography>
            )}

            {isLoading ? (
                <>
                    {totalOfferAmount > 0 && (
                        <PaginationBar
                            page={displayPage}
                            totalItems={totalOfferAmount}
                            pageSize={PAGE_SIZE}
                            onChange={handlePageChange}
                            loading={true}
                        />
                    )}
                    <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
                        <CircularProgress size={40} />
                    </Box>
                </>
            ) : (
                <>
                    {lastRequestTime && (
                        <Typography variant="caption">
                            Request took {lastRequestTime.toFixed(0)} ms
                        </Typography>
                    )}

                    {totalOfferAmount > 0 && (
                        <PaginationBar
                            page={displayPage}
                            totalItems={totalOfferAmount}
                            pageSize={PAGE_SIZE}
                            onChange={handlePageChange}
                            loading={false}
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

                    {totalOfferAmount > 0 && (
                        <PaginationBar
                            page={displayPage}
                            totalItems={totalOfferAmount}
                            pageSize={PAGE_SIZE}
                            onChange={handlePageChange}
                            loading={false}
                        />
                    )}
                </>
            )}
        </>
    );

}
