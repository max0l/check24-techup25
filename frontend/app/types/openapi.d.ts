import type {
  OpenAPIClient,
  Parameters,
  UnknownParamsObject,
  OperationResponse,
  AxiosRequestConfig,
} from 'openapi-client-axios'; 

declare namespace Definitions {
    export interface DataBestOfferPerHotel {
        /**
         * example:
         * 2
         */
        countadults?: number;
        /**
         * example:
         * 1
         */
        countchildren?: number;
        /**
         * example:
         * 7
         */
        durationindays?: number;
        /**
         * example:
         * 123
         */
        hotelid?: number;
        /**
         * example:
         * Hotel Sunshine
         */
        hotelname?: string;
        /**
         * example:
         * 4.5
         */
        hotelstars?: number;
        /**
         * example:
         * PMI
         */
        inboundarrivalairport?: string;
        /**
         * example:
         * 2025-12-31T00:00:00+02:00
         */
        inboundarrivaldatetime?: string;
        /**
         * example:
         * BER
         */
        inbounddepartureairport?: string;
        /**
         * example:
         * 2025-12-31T00:00:00+02:00
         */
        inbounddeparturedatetime?: string;
        /**
         * example:
         * allinclusive
         */
        mealtye?: string;
        /**
         * example:
         * true
         */
        ocenview?: boolean;
        /**
         * example:
         * 1
         */
        offeramount?: number;
        /**
         * example:
         * BER
         */
        outboundarrivalairport?: string;
        /**
         * example:
         * 2025-01-01T00:00:00+02:00
         */
        outboundarrivaldatetime?: string;
        /**
         * example:
         * PMI
         */
        outbounddepartureairport?: string;
        /**
         * example:
         * 2025-01-01T00:00:00+02:00
         */
        outbounddeparturedatetime?: string;
        /**
         * example:
         * 299.99
         */
        price?: number;
        /**
         * example:
         * double
         */
        rommtype?: string;
    }
    export interface DataBestOfferPerHotelResponse {
        /**
         * example:
         * 5
         */
        countOffers?: number;
        offers?: DataBestOfferPerHotel[];
    }
    export interface DataHotelOffersResponse {
        count_available_offers?: number;
        hotelid?: number;
        hotelname?: string;
        hotelstars?: number;
        offers?: DataOffer[];
    }
    export interface DataOffer {
        countadults?: number;
        countchildren?: number;
        durationindays?: number;
        hotelid?: number;
        inboundarrivalairport?: string;
        inboundarrivaldatetime?: string;
        inbounddepartureairport?: string;
        inbounddeparturedatetime?: string;
        mealtype?: string;
        oceanview?: boolean;
        offerid?: number;
        outboundarrivalairport?: string;
        outboundarrivaldatetime?: string;
        outbounddepartureairport?: string;
        outbounddeparturedatetime?: string;
        price?: number;
        roomtype?: string;
    }
}
declare namespace Paths {
    namespace ApiHotel$Hotelid {
        namespace Get {
            namespace Parameters {
                /**
                 * Number of adults
                 */
                export type CountAdults = number;
                /**
                 * Number of children
                 */
                export type CountChildren = number;
                /**
                 * List of departure airports, e.g. FRA, MUC
                 */
                export type DepartureAirports = string[];
                /**
                 * Duration of the trip in days
                 */
                export type Duration = number;
                /**
                 * Earliest departure date in format RFC3339
                 */
                export type EarliestDepartureDate = string;
                /**
                 * ID of the hotel
                 */
                export type Hotelid = number;
                /**
                 * Latest return date in format RFC3339
                 */
                export type LatestReturnDate = string;
                /**
                 * Site, e.g. 1 or 2
                 */
                export type Site = number;
            }
            export interface PathParameters {
                hotelid: /* ID of the hotel */ Parameters.Hotelid;
            }
            export interface QueryParameters {
                site: /* Site, e.g. 1 or 2 */ Parameters.Site;
                departureAirports?: /* List of departure airports, e.g. FRA, MUC */ Parameters.DepartureAirports;
                earliestDepartureDate?: /* Earliest departure date in format RFC3339 */ Parameters.EarliestDepartureDate;
                latestReturnDate?: /* Latest return date in format RFC3339 */ Parameters.LatestReturnDate;
                countAdults?: /* Number of adults */ Parameters.CountAdults;
                countChildren?: /* Number of children */ Parameters.CountChildren;
                duration?: /* Duration of the trip in days */ Parameters.Duration;
            }
            namespace Responses {
                export type $200 = Definitions.DataHotelOffersResponse;
            }
        }
    }
    namespace ApiOffers {
        namespace Get {
            namespace Parameters {
                /**
                 * Number of adults
                 */
                export type CountAdults = number;
                /**
                 * Number of children
                 */
                export type CountChildren = number;
                /**
                 * List of departure airports, e.g. FRA, MUC
                 */
                export type DepartureAirports = string[];
                /**
                 * Duration of the trip in days
                 */
                export type Duration = number;
                /**
                 * Earliest departure date in format RFC3339
                 */
                export type EarliestDepartureDate = string;
                /**
                 * Latest return date in format RFC3339
                 */
                export type LatestReturnDate = string;
                /**
                 * Site, e.g. 1 or 2
                 */
                export type Site = number;
            }
            export interface QueryParameters {
                departureAirports: /* List of departure airports, e.g. FRA, MUC */ Parameters.DepartureAirports;
                earliestDepartureDate: /* Earliest departure date in format RFC3339 */ Parameters.EarliestDepartureDate;
                latestReturnDate: /* Latest return date in format RFC3339 */ Parameters.LatestReturnDate;
                countAdults: /* Number of adults */ Parameters.CountAdults;
                countChildren?: /* Number of children */ Parameters.CountChildren;
                duration: /* Duration of the trip in days */ Parameters.Duration;
                site?: /* Site, e.g. 1 or 2 */ Parameters.Site;
            }
            namespace Responses {
                export type $200 = Definitions.DataBestOfferPerHotelResponse;
            }
        }
    }
}

export interface OperationMethods {
}

export interface PathsDictionary {
  ['/api/hotel/{hotelid}']: {
  }
  ['/api/offers']: {
  }
}

export type Client = OpenAPIClient<OperationMethods, PathsDictionary>
