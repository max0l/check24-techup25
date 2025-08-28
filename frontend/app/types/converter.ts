import { Paths } from "@/app/types/openapi";
import { ReadonlyURLSearchParams } from "next/navigation";

export function GetBestOffersByHotelToQuery(input: Paths.Offers.Get.QueryParameters): string {
    const params = {
        earliestDepartureDate: input.earliestDepartureDate,
        latestReturnDate: input.latestReturnDate,
        duration: input.duration.toString(),
        countAdults: input.countAdults.toString(),
        countChildren: (input.countChildren ?? 0).toString(),
        site: (input.site ?? 1).toString(),
        departureAirports: input.departureAirports.join(",")
    };

    const searchParams = new URLSearchParams(params);
    return searchParams.toString();
}

export function GetBestOffersByHotelFromQuery(query: ReadonlyURLSearchParams): Paths.Offers.Get.QueryParameters {
    return {
        earliestDepartureDate: query.get("earliestDepartureDate") as string,
        latestReturnDate: query.get("latestReturnDate") as string,
        countAdults: parseInt(query.get("countAdults") as string),
        countChildren: parseInt(query.get("countChildren") as string),
        departureAirports: query.get("departureAirports")?.split(",") ?? [],
        duration: parseInt(query.get("duration") as string),
        site: parseInt(query.get("site") as string),
    };
}

export function GetHotelOffersFromQuery(query: ReadonlyURLSearchParams): Paths.Hotel$Hotelid.Get.QueryParameters {
    return {
        earliestDepartureDate: query.get("earliestDepartureDate") as string ?? "",
        latestReturnDate: query.get("latestReturnDate") as string ?? "",
        countAdults: parseInt(query.get("countAdults") as string) || 0,
        countChildren: parseInt(query.get("countChildren") as string) || 0,
        departureAirports: query.get("departureAirports")?.split(",") ?? [],
        duration: parseInt(query.get("duration") as string) || 0,
        site: parseInt(query.get("site") as string) || 1,
    }
}
export function GetHotelIdFromQuery(query: ReadonlyURLSearchParams): Paths.Hotel$Hotelid.Get.PathParameters {
    return {
        hotelid: parseInt(query.get("hotelId") as string)
    }
}

export function GetHotelOffersToQuery(parameters: Paths.Hotel$Hotelid.Get.QueryParameters,
    hotelid: Paths.Hotel$Hotelid.Get.PathParameters
): string {
    const params = {
        earliestDepartureDate: parameters.earliestDepartureDate ?? "",
        latestReturnDate: parameters.latestReturnDate ?? "",
        countAdults: (parameters.countAdults ?? 0).toString(),
        countChildren: (parameters.countChildren ?? 0).toString(),
        duration: (parameters.duration ?? 0).toString(),
        departureAirports: parameters.departureAirports?.join(",") ?? "",
        site: (parameters.site ?? 1).toString(),
        hotelId: hotelid.hotelid.toString()
    };
    const searchParams = new URLSearchParams(params);
    return searchParams.toString();
}