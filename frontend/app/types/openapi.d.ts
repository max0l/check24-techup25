import type {
  OpenAPIClient,
  Parameters,
  UnknownParamsObject,
  OperationResponse,
  AxiosRequestConfig,
} from 'openapi-client-axios'; 

declare namespace Definitions {
    export interface DataHotelResponse {
        hotelid?: number;
        hotelname?: string;
        hotelstars?: number; // float32
        offers?: DataOffer[];
    }
    export interface DataOffer {
        countadults?: number;
        countchildren?: number;
        durationInDays?: number;
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
        price?: number; // float32
        roomtype?: string;
    }
    export interface DataOfferResponse {
        /**
         * example:
         * 2
         */
        Countadults?: number;
        /**
         * example:
         * 1
         */
        Countchildren?: number;
        /**
         * example:
         * Hotel Sunshine
         */
        HotelName?: string;
        /**
         * example:
         * 4.5
         */
        HotelStars?: number;
        /**
         * example:
         * PMI
         */
        Inboundarrivalairport?: string;
        /**
         * example:
         * 2025-12-31T00:00:00+02:00
         */
        Inboundarrivaldatetime?: string;
        /**
         * example:
         * BER
         */
        Inbounddepartureairport?: string;
        /**
         * example:
         * 2025-12-31T00:00:00+02:00
         */
        Inbounddeparturedatetime?: string;
        /**
         * example:
         * allinclusive
         */
        Mealtye?: string;
        /**
         * example:
         * true
         */
        Ocenview?: boolean;
        /**
         * example:
         * BER
         */
        Outboundarrivalairport?: string;
        /**
         * example:
         * 2025-01-01T00:00:00+02:00
         */
        Outboundarrivaldatetime?: string;
        /**
         * example:
         * PMI
         */
        Outbounddepartureairport?: string;
        /**
         * example:
         * 2025-01-01T00:00:00+02:00
         */
        Outbounddeparturedatetime?: string;
        /**
         * example:
         * 299.99
         */
        Price?: number;
        /**
         * example:
         * double
         */
        Rommtype?: string;
    }
}
declare namespace Paths {
    namespace Hotel$Hotelid {
        namespace Get {
            namespace Parameters {
                /**
                 * ID of the hotel
                 */
                export type Hotelid = number;
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
            }
            namespace Responses {
                export type $200 = Definitions.DataHotelResponse;
            }
        }
    }
    namespace Offers {
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
                DepartureAirports: /* List of departure airports, e.g. FRA, MUC */ Parameters.DepartureAirports;
                EarliestDepartureDate: /* Earliest departure date in format RFC3339 */ Parameters.EarliestDepartureDate;
                LatestReturnDate: /* Latest return date in format RFC3339 */ Parameters.LatestReturnDate;
                CountAdults: /* Number of adults */ Parameters.CountAdults;
                CountChildren?: /* Number of children */ Parameters.CountChildren;
                Duration: /* Duration of the trip in days */ Parameters.Duration;
                Site: /* Site, e.g. 1 or 2 */ Parameters.Site;
            }
            namespace Responses {
                export type $200 = Definitions.DataOfferResponse[];
            }
        }
    }
}

export interface OperationMethods {
}

export interface PathsDictionary {
  ['/hotel/{hotelid}']: {
  }
  ['/offers']: {
  }
}

export type Client = OpenAPIClient<OperationMethods, PathsDictionary>
