"use client";

import { Box, Pagination } from "@mui/material";

type Props = {
    page: number;
    totalItems: number;
    pageSize?: number;
    onChange: (page: number) => void;
};

export default function PaginationBar({
    page,
    totalItems,
    pageSize = 50,
    onChange,
}: Props) {
    const pageCount = Math.max(1, Math.ceil(totalItems / pageSize));

    return (
        <Box
            display="flex"
            justifyContent="center"
            my={2}
            mx="auto"
        >
            <Pagination
                count={pageCount}
                page={page}
                size="large"
                color="standard"
                onChange={(_, p) => onChange(p)}
                sx={{
                    "& .MuiPaginationItem-root": {
                        color: "#334155",
                    },
                    "& .MuiPaginationItem-root:hover": {
                        backgroundColor: "rgba(14,165,233,0.2)",
                    },
                    "& .MuiPaginationItem-ellipsis": {
                        color: "#64748b",
                    },
                    "& .MuiPaginationItem-root.Mui-selected": {
                        backgroundColor: "#053773",
                        color: "#fff",
                    },
                }}
            />
        </Box>
    );
}
