"use client";

import { OpenAPIClientAxios } from "openapi-client-axios";
import { Client, Definitions, Paths } from "@/app/types/openapi";
import HotelOffer from "@/app/components/HotelOffer/HotelOffer";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState, useMemo } from "react";
import { Box, CircularProgress, Rating, Typography, LinearProgress } from "@mui/material";
import { Stack } from "@mui/system";
import { GetHotelOffersFromQuery, GetHotelIdFromQuery, GetHotelOffersToQuery } from "@/app/types/converter";
import qs from "qs";
import PaginationBar from "@/app/components/Pagination/PaginationBar";
import {extractAxiosErrorMessage} from "@/app/utils/axiosExtractor"

const apiBase =
    typeof window === "undefined"
        ? process.env.INTERNAL_API_BASE_URL ?? "http://api:8080"
        : process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";

type HotelOfferResp = Definitions.DataHotelOffersResponse;
export default function Page() {
    const query = useSearchParams();
    const router = useRouter();

    const [hotelOffer, setHotelOffer] = useState<HotelOfferResp>();
    const [queryParameters, setQueryParameters] =
        useState<Paths.Hotel$Hotelid.Get.QueryParameters>();
    const [displayPage, setDisplayPage] = useState<number>(1);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [requestError, setRequestError] = useState<string>();

    const PAGE_SIZE = 50;
    const totalOfferAmount = hotelOffer?.count_available_offers ?? 0;

    const [lastRequestTime, setLastRequestTime] = useState<number>();

    const apiClient = useMemo(() => {
        const api = new OpenAPIClientAxios({
            definition: `/swagger/doc.json`,
            withServer: 0,
            axiosConfigDefaults: {
                baseURL: apiBase,
                paramsSerializer: { serialize: (p) => qs.stringify(p, { arrayFormat: "comma" }) },
            },
        });
        return api;
    }, []);

    async function fetchData(parameters: Paths.Hotel$Hotelid.Get.QueryParameters) {
        setRequestError(undefined)
        try {
            setIsLoading(true);
            const start = performance.now();

            const hotelid = GetHotelIdFromQuery(query);
            const client = await apiClient.init<Client>();
            const response = await client.get<HotelOfferResp>(`/hotel/${hotelid.hotelid}`, {
                params: parameters,
            });

            const end = performance.now();
            const durationMs = end - start;
            setLastRequestTime(durationMs);

            setHotelOffer(response.data);
            setQueryParameters(parameters);
            setDisplayPage(parameters.site ?? 1);

            router.push("/offers?" + GetHotelOffersToQuery(parameters, hotelid));
        } catch (err) {
            setHotelOffer(undefined)
            setLastRequestTime(undefined);
            setRequestError(extractAxiosErrorMessage(err));
        } finally {
            setIsLoading(false);
        }
    }

    useEffect(() => {
        const params = GetHotelOffersFromQuery(query);
        // initialize "committed" page from initial params
        setDisplayPage(params.site || 1);
        fetchData(params).catch(console.error);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const handlePageChange = (p: number) => {
        if (!queryParameters || isLoading || p === displayPage) return;
        window.scrollTo({ top: 0, behavior: "smooth" });
        fetchData({ ...queryParameters, site: p });
    };

    if (!hotelOffer) {
        return (
            <Box mt={6}>
                <LinearProgress />
                <Typography mt={2}>Loading offers...</Typography>
            </Box>
        );
    }

    return (
        <>
            {isLoading && <LinearProgress />}

            {!isLoading && requestError && (
                <Typography color="error" sx={{ mt: 2 }}>
                    Something went wrong: {requestError}
                </Typography>
            )}

            <Stack direction="row" pt={5} alignItems="center">
                <Typography variant="h4" mr={2}>
                    {hotelOffer.hotelname ?? "NaN"}: {hotelOffer.hotelid}
                </Typography>
                <Rating value={hotelOffer.hotelstars ?? 0} readOnly />
            </Stack>

            {lastRequestTime && (
                <Typography variant="caption">
                    Request took {lastRequestTime.toFixed(0)} ms
                </Typography>
            )}

            <Stack direction="row" height="250px" pt={2}>
                <Box
                    sx={{
                        borderTopLeftRadius: "5px",
                        borderBottomLeftRadius: "5px",
                        backgroundImage: `url("/hotels/${((hotelOffer.hotelid ?? 0) % 40) + 1}.jpg")`,
                        width: "50%",
                        backgroundSize: "cover",
                        backgroundPosition: "center",
                    }}
                />
                <Box
                    sx={{
                        borderTopRightRadius: "5px",
                        borderBottomRightRadius: "5px",
                        backgroundImage: `url("/rooms/${((hotelOffer.hotelid ?? 0) % 30) + 1}.jpg")`,
                        width: "50%",
                        backgroundSize: "cover",
                        backgroundPosition: "center",
                    }}
                />
            </Stack>

            <Typography pt={2} variant="h6">
                Found {hotelOffer.count_available_offers ?? 0} Offers:
            </Typography>

            {totalOfferAmount > 0 && (
                <PaginationBar
                    page={displayPage}
                    totalItems={totalOfferAmount}
                    pageSize={PAGE_SIZE}
                    onChange={handlePageChange}
                    loading={isLoading}
                />
            )}
            <Stack gap={2} mt={1} position="relative">
                {isLoading ? (
                    <Box
                        sx={{ display: "flex", justifyContent: "center", py: 4 }}
                    >
                        <CircularProgress size={40} />
                    </Box>
                ) : (
                    hotelOffer?.offers?.map((offer) => (
                        <HotelOffer key={offer.offerid} offer={offer} />
                    ))
                )}
            </Stack>
        </>
    );
}
