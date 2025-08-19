import { Suspense } from "react";
import { Box, LinearProgress, Typography } from "@mui/material";
import ClientPage from "./ClientPage";

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
      <ClientPage />
    </Suspense>
  );
}
