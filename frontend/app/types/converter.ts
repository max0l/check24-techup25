import {Paths} from "@/app/types/openapi";
import {ReadonlyURLSearchParams} from "next/navigation";

export function GetBestOffersByHotelToQuery(input: Paths.Offers.Get.QueryParameters): string {
    const params = {
        EarliestDepartureDate: input.earliestDepartureDate,
        LatestReturnDate: input.latestReturnDate,
        Duration: input.duration.toString(),
        CountAdults: input.countAdults.toString(),
        CountChildren: (input.countChildren ?? 0).toString(),
        site: (input.site ?? 1).toString(),
    };

    const searchParams = new URLSearchParams(params);
    
    // cant pass an object with an array to URLSearchParams --> append individually
    input.departureAirports.forEach((airport) => {
        searchParams.append("DepartureAirports", airport);
    })

    return searchParams.toString();
}

export function GetBestOffersByHotelFromQuery(query: ReadonlyURLSearchParams): Paths.Offers.Get.QueryParameters {
    return {
        earliestDepartureDate: query.get("earliestDepartureDate") as string,
        latestReturnDate: query.get("latestReturnDate") as string,
        countAdults: parseInt(query.get("countAdults") as string),
        countChildren: parseInt(query.get("countChildren") as string),
        departureAirports: query.getAll("departureAirports"),
        duration: parseInt(query.get("duration") as string),
        site: parseInt(query.get("site") as string),
    };
}

export function GetHotelOffersFromQuery(query: ReadonlyURLSearchParams): Paths.Hotel$Hotelid.Get.QueryParameters {
    return {
        earliestDepartureDate: query.get("earliestDepartureDate") as string,
        latestReturnDate: query.get("latestReturnDate") as string,
        countAdults: parseInt(query.get("countAdults") as string),
        countChildren: parseInt(query.get("countChildren") as string),
        departureAirports: query.getAll("departureAirports"),
        duration: parseInt(query.get("duration") as string),
        site: parseInt(query.get("site") as string),
    }
}

export function GetHotelIdFromQuery(query: ReadonlyURLSearchParams): Paths.Hotel$Hotelid.Get.PathParameters {
    return {
        hotelid: parseInt(query.get("hotelid") as string)
    }
}