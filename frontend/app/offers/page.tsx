import { Suspense } from "react";
import { Box, LinearProgress, Typography } from "@mui/material";
import ClientOffersPage from "./ClientOffersPage";

export default function Page() {
  return (
    <Suspense
      fallback={
        <Box mt={6}>
          <LinearProgress />
          <Typography mt={2}>Loading offers...</Typography>
        </Box>
      }
    >
      <ClientOffersPage />
    </Suspense>
  );
}
