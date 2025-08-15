import {Paths} from "@/app/types/openapi";
import {ReadonlyURLSearchParams} from "next/navigation";

export function GetBestOffersByHotelToQuery(input: Paths.Offers.Get.QueryParameters): string {
    const params = {
        EarliestDepartureDate: input.EarliestDepartureDate,
        LatestReturnDate: input.LatestReturnDate,
        Duration: input.Duration.toString(),
        CountAdults: input.CountAdults.toString(),
        CountChildren: (input.CountChildren ?? 0).toString(),
        site: input.Site.toString()
    };

    const searchParams = new URLSearchParams(params);
    
    // cant pass an object with an array to URLSearchParams --> append individually
    input.DepartureAirports.forEach((airport) => {
        searchParams.append("DepartureAirports", airport);
    })

    return searchParams.toString();
}

export function GetBestOffersByHotelFromQuery(query: ReadonlyURLSearchParams): Paths.Offers.Get.QueryParameters {
    return {
        EarliestDepartureDate: query.get("EarliestDepartureDate") as string,
        LatestReturnDate: query.get("LatestReturnDate") as string,
        CountAdults: parseInt(query.get("CountAdults") as string),
        CountChildren: parseInt(query.get("CountChildren") as string),
        DepartureAirports: query.getAll("DepartureAirports"),
        Duration: parseInt(query.get("Duration") as string),
        Site: parseInt(query.get("Site") as string),
    };
}

export function GetHotelOffersFromQuery(query: ReadonlyURLSearchParams): Paths.Hotel$Hotelid.Get.PathParameters & Paths.Hotel$Hotelid.Get.QueryParameters {
    return {
        hotelid: parseInt(query.get("hotelid") as string),
        site: parseInt(query.get("site") as string),
    }
}