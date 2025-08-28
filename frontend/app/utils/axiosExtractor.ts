// add this import
import axios from "axios";

// helper to extract a readable message from axios errors
export function extractAxiosErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const { response, message } = err;
    const code = response?.status ? `${response.status} ${response.statusText ?? ""}`.trim() : "";
    const data = response?.data;

    // Prefer server-provided message if present
    if (data) {
      if (typeof data === "string") {
        return code ? `${code}: ${data}` : data;
      }
      // Common shapes: { message }, { error }, ProblemDetails, etc.
      const serverMsg =
        (data.message ?? data.error ?? data.title ?? data.detail) ??
        // last resort: stringify (safe)
        (() => {
          try { return JSON.stringify(data); } catch { return ""; }
        })();

      if (serverMsg) return code ? `${code}: ${serverMsg}` : serverMsg;
    }
    // Fallback to axios error message
    return code ? `${code}: ${message}` : message;
  }
  return err instanceof Error ? err.message : "Unknown error";
}
