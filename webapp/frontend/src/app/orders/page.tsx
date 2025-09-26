"use client";
import React, { useEffect, useState, useCallback, useRef } from "react";
import Chip from "@mui/material/Chip";
import CustomDataGrid from "../../components/CustomDataGrid";
import {
  GridColDef,
  GridRenderCellParams,
  GridSortModel,
} from "@mui/x-data-grid";
import axios from "axios";
import {
  Box,
  Container,
  Typography,
  TextField,
  Button,
  Paper,
  FormControl,
  FormLabel,
  RadioGroup,
  FormControlLabel,
  Radio,
} from "@mui/material";
import { useRouter } from "next/navigation";

type ShippedStatus = "completed" | "delivering" | "shipping";

type OrdersRow = {
  id: number;
  product_name: string;
  shipped_status: ShippedStatus;
  created_at: Date | string;
  arrived_at: {
    Time: string;
    Valid: boolean;
  };
};

type SearchType = "partial" | "prefix";

export default function OrdersPage() {
  const [ordersRow, setOrdersRow] = useState<OrdersRow[]>([]);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [searchType, setSearchType] = useState<SearchType>("partial");
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(20);
  const [sortModel, setSortModel] = useState<GridSortModel>([
    { field: "order_id", sort: "desc" },
  ]);

  // 入力値を保持するref
  const searchQueryInputRef = useRef<HTMLInputElement>(null);
  const [pendingSearchQuery, setPendingSearchQuery] = useState<string>("");
  const [pendingSearchType, setPendingSearchType] =
    useState<SearchType>("partial");

  const router = useRouter();

  // fetchOrdersはstateの値のみ使う
  const fetchOrders = useCallback(
    (
      pageNum: number = page,
      size: number = pageSize,
      model: GridSortModel = sortModel
    ) => {
      const sortF = model[0]?.field ?? "order_id";
      const sortO = model[0]?.sort ?? "desc";
      setIsLoading(true);
      axios
        .post("/api/v1/orders", {
          search: searchQuery,
          type: searchType,
          page: pageNum,
          page_size: size,
          sort_field: sortF,
          sort_order: sortO,
        })
        .then((response) => {
          setOrdersRow(response.data.data);
          setTotalCount(response.data.total);
        })
        .catch((error) => {
          console.error("注文データの取得に失敗しました:", error);
          setOrdersRow([]);
        })
        .finally(() => setIsLoading(false));
    },
    [page, pageSize, searchQuery, searchType, sortModel]
  );

  useEffect(() => {
    fetchOrders(page, pageSize, sortModel);
  }, [page, pageSize, searchQuery, searchType, sortModel]);

  // 検索ボタン押下時のみstateを更新
  const handleSearch = (event: React.FormEvent) => {
    event.preventDefault();
    setSearchQuery(pendingSearchQuery);
    setSearchType(pendingSearchType);
    setPage(1); // 検索時は1ページ目に戻す
  };

  // 入力値はstateで保持
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPendingSearchQuery(e.target.value);
  };
  const handleTypeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPendingSearchType(e.target.value as SearchType);
  };

  const handleReset = () => {
    setPendingSearchQuery("");
    setPendingSearchType("partial");
    setSearchQuery("");
    setSearchType("partial");
    setPage(1);
  };

  // ページ変更時のハンドラ
  const handlePageChange = (newPage: number) => {
    setPage(newPage + 1);
  };

  // ページサイズ変更時のハンドラ
  const handlePageSizeChange = (newSize: number) => {
    setPageSize(newSize);
    setPage(1);
  };

  // ソート変更時のハンドラ
  const handleSortModelChange = (model: GridSortModel) => {
    if (model.length === 0) return; // 未ソートは無視
    setSortModel(model);
  };

  const columns: GridColDef[] = [
    {
      field: "order_id",
      headerName: "注文ID",
      flex: 0.5,
      minWidth: 80,
      headerAlign: "left",
      align: "left",
    },
    {
      field: "product_name",
      headerName: "商品名",
      flex: 1.5,
      minWidth: 200,
    },
    {
      field: "shipped_status",
      headerName: "配送ステータス",
      flex: 1,
      minWidth: 80,
      renderCell: (params) => renderStatus(params.value as ShippedStatus),
    },
    {
      field: "created_at",
      headerName: "注文日時",
      headerAlign: "right",
      align: "right",
      flex: 1,
      minWidth: 80,
    },
    {
      field: "arrived_at",
      headerName: "配送完了日時",
      flex: 1,
      minWidth: 150,
      renderCell: (
        params: GridRenderCellParams<
          OrdersRow,
          { Time: string; Valid: boolean } | null
        >
      ) => {
        if (params.value && params.value.Valid) {
          return new Date(params.value.Time).toLocaleString("ja-JP");
        }
        return "未定";
      },
    },
  ];

  function renderStatus(status: ShippedStatus) {
    switch (status) {
      case "completed":
        return <Chip label="配送完了" color="success" size="small" />;
      case "delivering":
        return <Chip label="配送中" color="primary" size="small" />;
      case "shipping":
        return <Chip label="出荷準備" color="default" size="small" />;
      default:
        return <Chip label="不明" color="default" size="small" />;
    }
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        注文一覧
      </Typography>

      {/* productsページへ遷移するボタン */}
      <Box sx={{ mb: 2, display: "flex", justifyContent: "flex-end" }}>
        <Button
          variant="outlined"
          color="primary"
          onClick={() => router.push("/product")}
        >
          商品一覧ページへ
        </Button>
      </Box>

      {/* 総レコード数とページ表示 */}
      <Box sx={{ mb: 2 }}>
        <Typography variant="body2">
          総件数: {totalCount}件 / ページ: {page} / 全
          {Math.ceil(totalCount / pageSize)}ページ
        </Typography>
      </Box>

      {/* 検索フォーム */}
      <Paper elevation={1} sx={{ p: 2, mb: 3 }}>
        <form onSubmit={handleSearch}>
          <Box
            sx={{ display: "flex", gap: 2, alignItems: "flex-start", mb: 2 }}
          >
            <TextField
              placeholder="商品名またはステータスで検索..."
              value={pendingSearchQuery}
              onChange={handleInputChange}
              size="small"
              sx={{ minWidth: 300, flex: 1 }}
              disabled={isLoading}
              inputRef={searchQueryInputRef}
            />
            <FormControl component="fieldset">
              <FormLabel component="legend" sx={{ fontSize: "0.875rem" }}>
                検索タイプ
              </FormLabel>
              <RadioGroup
                row
                value={pendingSearchType}
                onChange={handleTypeChange}
              >
                <FormControlLabel
                  value="partial"
                  control={<Radio size="small" />}
                  label="部分一致"
                  disabled={isLoading}
                />
                <FormControlLabel
                  value="prefix"
                  control={<Radio size="small" />}
                  label="前方一致"
                  disabled={isLoading}
                />
              </RadioGroup>
            </FormControl>
          </Box>
          <Box sx={{ display: "flex", gap: 2, alignItems: "center" }}>
            <Button type="submit" variant="contained" disabled={isLoading}>
              検索
            </Button>
            <Button
              variant="outlined"
              onClick={handleReset}
              disabled={isLoading}
            >
              リセット
            </Button>
          </Box>
        </form>
      </Paper>

      <Box sx={{ height: 650, width: "100%" }}>
        <CustomDataGrid
          rows={ordersRow}
          columns={columns}
          getRowId={(row) => row.order_id}
          loading={isLoading}
          rowCount={totalCount}
          paginationModel={{ page: page - 1, pageSize: pageSize }}
          paginationMode="server"
          onPaginationModelChange={({ page: newPage, pageSize: newSize }) => {
            if (newSize !== pageSize) {
              handlePageSizeChange(newSize);
            } else if (newPage !== page - 1) {
              handlePageChange(newPage);
            }
          }}
          sortingMode="server"
          sortModel={sortModel}
          onSortModelChange={handleSortModelChange}
          sortingOrder={["asc", "desc"]} // ← 2段階のみ
        />
      </Box>
    </Container>
  );
}
