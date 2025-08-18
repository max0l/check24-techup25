// validators.ts
import {Paths} from "@/app/types/openapi";

export function coerceAndValidate(q: Partial<Paths.Offers.Get.QueryParameters>) {
  const earliestDepartureDate = q.earliestDepartureDate ?? "";
  const latestReturnDate = q.latestReturnDate ?? "";
  const countAdults = Number(q.countAdults ?? 0);
  const countChildren = Number(q.countChildren ?? 0);
  const duration = Number(q.duration ?? 0);
  const departureAirports = Array.isArray(q.departureAirports)
    ? q.departureAirports
    : (q.departureAirports ? [q.departureAirports] : []);
  const site = Number(q.site ?? 1)

  const valid =
    earliestDepartureDate !== "" &&
    latestReturnDate !== "" &&
    Number.isFinite(countAdults) &&
    countAdults > 0 &&
    departureAirports.length > 0 &&
    Number.isFinite(duration) &&
    duration > 0 &&
    site > 0;

  return { valid, params: {
    earliestDepartureDate, latestReturnDate, countAdults, countChildren,
    duration, departureAirports, site
  }};
}
