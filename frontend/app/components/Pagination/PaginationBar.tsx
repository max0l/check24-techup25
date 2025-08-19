"use client";

import { Box, Pagination, CircularProgress } from "@mui/material";

type Props = {
  page: number;
  totalItems: number;
  pageSize?: number;
  onChange: (page: number) => void;
  loading?: boolean;
};

export default function PaginationBar({
  page,
  totalItems,
  pageSize = 50,
  onChange,
  loading = false,
}: Props) {
  const pageCount = Math.max(1, Math.ceil(totalItems / pageSize));

  return (
    <Box display="flex" justifyContent="center" my={2} mx="auto" position="relative">
      <Box
        sx={{
          pointerEvents: loading ? "none" : "auto",
          opacity: loading ? 0.6 : 1,
        }}
      >
        <Pagination
          count={pageCount}
          page={page}
          size="large"
          color="standard"
          onChange={(_, p) => onChange(p)}
          sx={{
            "& .MuiPaginationItem-root": { color: "#334155" },
            "& .MuiPaginationItem-root:hover": { backgroundColor: "rgba(14,165,233,0.2)" },
            "& .MuiPaginationItem-ellipsis": { color: "#64748b" },
            "& .MuiPaginationItem-root.Mui-selected": {
              backgroundColor: "#053773",
              color: "#fff",
            },
          }}
        />
      </Box>

      {loading && (
        <Box
          position="absolute"
          right={-40}
          top="50%"
          sx={{ transform: "translateY(-50%)" }}
          aria-hidden
        >
        </Box>
      )}
    </Box>
  );
}
