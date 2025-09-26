import exec from "k6/execution";
import http from "k6/http";
import { check, fail } from "k6";
import { Counter } from "k6/metrics";
import { createApiUrl, createLoginUrl } from "./utils.js";

const bench_uj_success_count = new Counter("bench_uj_success_count");

const bench_uj_login_fail_count = new Counter("bench_uj_login_fail_count");
const bench_uj_products_list_fail_count = new Counter(
  "bench_uj_products_list_fail_count"
);
const bench_uj_other_products_page_fail_count = new Counter(
  "bench_uj_other_products_page_fail_count"
);
const bench_uj_products_sorted_fail_count = new Counter(
  "bench_uj_products_sorted_fail_count"
);
const bench_uj_products_search_fail_count = new Counter(
  "bench_uj_products_search_fail_count"
);
const bench_uj_orders_submit_fail_count = new Counter(
  "bench_uj_orders_submit_fail_count"
);
const bench_uj_orders_history_fail_count = new Counter(
  "bench_uj_orders_history_fail_count"
);
const bench_uj_orders_sorted_fail_count = new Counter(
  "bench_uj_orders_sorted_fail_count"
);
const bench_uj_orders_search_fail_count = new Counter(
  "bench_uj_orders_search_fail_count"
);

const searchWordList = JSON.parse(open("./sampleData/search_words.json"));
const allOrders = JSON.parse(open("./sampleData/products_quantities.json"));

export function userJourneyScenario() {
  // Step 1: ログイン
  const user_name =
    "user" + String(Math.floor(Math.random() * 100) + 1).padStart(3, "0");
  const password = "password";

  const loginRes = http.post(
    createLoginUrl("login"),
    JSON.stringify({ user_name, password }),
    { headers: { "Content-Type": "application/json" } }
  );

  if (!check(loginRes, { "login succeeded": (r) => r.status === 200 })) {
    bench_uj_login_fail_count.add(1);
    fail(`Failed to login: status=${loginRes.status}`);
  }

  const cookies = loginRes.cookies || {};
  const arr =
    cookies.session_id || cookies["session_id"] || cookies["SESSION_ID"] || [];
  const sessionCookie =
    arr.length > 0 && arr[0] && arr[0].value ? arr[0].value : null;

  if (!sessionCookie) {
    bench_uj_login_fail_count.add(1);
    fail("No session_id cookie received");
  }

  const headers = {
    "Content-Type": "application/json",
    Cookie: `session_id=${sessionCookie}`,
  };

  // Step 2: 商品一覧表示
  const productsListRes = http.post(
    createApiUrl("/product"),
    JSON.stringify({
      search: "",
      type: "partial",
      page: 1,
      page_size: 20,
      sort_field: "name",
      sort_order: "asc",
    }),
    { headers }
  );

  const okProducts = check(productsListRes, {
    "products list loaded": (r) => r.status === 200,
  });
  if (!okProducts) {
    bench_uj_products_list_fail_count.add(1);
    fail("products list failed");
  }

  // Step 3: 別ページ表示
  const pageNum = Math.floor(Math.random() * 10) + 1;
  const nextPageRes = http.post(
    createApiUrl("/product"),
    JSON.stringify({
      search: "",
      type: "partial",
      page: pageNum,
      page_size: 20,
      sort_field: "name",
      sort_order: "asc",
    }),
    { headers }
  );

  const okNextPage = check(nextPageRes, {
    "other page loaded": (r) => r.status === 200,
  });
  if (!okNextPage) {
    bench_uj_other_products_page_fail_count.add(1);
    fail(`other page ${pageNum} failed`);
  }

  // Step 4: ソート変更
  const sortFields = ["name", "value", "weight"];
  const sortOrders = ["asc", "desc"];
  const sortField =
    sortFields[exec.scenario.iterationInInstance % sortFields.length];
  const sortOrder =
    sortOrders[exec.scenario.iterationInInstance % sortOrders.length];

  const sortedRes = http.post(
    createApiUrl("/product"),
    JSON.stringify({
      search: "",
      type: "partial",
      page: 1,
      page_size: 20,
      sort_field: sortField,
      sort_order: sortOrder,
    }),
    { headers }
  );

  const okSorted = check(sortedRes, {
    "sorted products loaded": (r) => r.status === 200,
  });
  if (!okSorted) {
    bench_uj_products_sorted_fail_count.add(1);
    fail(`sorting by ${sortField} ${sortOrder} failed`);
  }

  // Step 5: 商品検索
  const searchWord =
    searchWordList[exec.scenario.iterationInInstance % searchWordList.length]
      .keyword;
  const searchRes = http.post(
    createApiUrl("/product"),
    JSON.stringify({
      search: searchWord,
      type: "partial",
      page: 1,
      page_size: 20,
      sort_field: "name",
      sort_order: "asc",
    }),
    { headers }
  );

  const okSearch = check(searchRes, {
    "product search successful": (r) => r.status === 200,
  });
  if (!okSearch) {
    bench_uj_products_search_fail_count.add(1);
    fail("product search failed");
  }

  // Step 6: 注文作成
  const orderData =
    allOrders[exec.scenario.iterationInInstance % allOrders.length];
  const orderRes = http.post(
    createApiUrl("/product/post"),
    JSON.stringify(orderData),
    { headers }
  );

  const okOrder = check(orderRes, {
    "order created successfully": (r) => r.status === 201,
  });
  if (!okOrder) {
    bench_uj_orders_submit_fail_count.add(1);
    fail("order creation failed");
  }

  // 全ステップ成功なら overall success をインクリメント
  bench_uj_success_count.add(1);
}

export function userJourneyScenario2() {
  // Step 1: ログイン
  const user_name =
    "user" + String(Math.floor(Math.random() * 100) + 1).padStart(3, "0");
  const password = "password";

  const loginRes = http.post(
    createLoginUrl("login"),
    JSON.stringify({ user_name, password }),
    { headers: { "Content-Type": "application/json" } }
  );

  if (!check(loginRes, { "login succeeded": (r) => r.status === 200 })) {
    bench_uj_login_fail_count.add(1);
    fail(`Failed to login: status=${loginRes.status}`);
  }

  const cookies = loginRes.cookies || {};
  const arr =
    cookies.session_id || cookies["session_id"] || cookies["SESSION_ID"] || [];
  const sessionCookie =
    arr.length > 0 && arr[0] && arr[0].value ? arr[0].value : null;

  if (!sessionCookie) {
    bench_uj_login_fail_count.add(1);
    fail("No session_id cookie received");
  }

  const headers = {
    "Content-Type": "application/json",
    Cookie: `session_id=${sessionCookie}`,
  };

  // Step 7: 注文履歴確認
  const ordersHistoryRes = http.post(
    createApiUrl("/orders"),
    JSON.stringify({
      search: "",
      type: "partial",
      page: 1,
      page_size: 20,
      sort_field: "created_at",
      sort_order: "desc",
    }),
    { headers }
  );

  const okHistory = check(ordersHistoryRes, {
    "order history loaded": (r) => r.status === 200,
  });
  if (!okHistory) {
    bench_uj_orders_history_fail_count.add(1);
    fail("order history failed");
  }

  // step 8: 注文履歴のソート変更
  const sortFieldOrders = [
    "order_id",
    "name",
    "shipped_status",
    "created_at",
    "arrived_at",
  ];
  const sortOrderOrders = ["asc", "desc"];
  const sortFieldOrder =
    sortFieldOrders[exec.scenario.iterationInInstance % sortFieldOrders.length];
  const sortOrderOrder =
    sortOrderOrders[exec.scenario.iterationInInstance % sortOrderOrders.length];

  const ordersHistorySortedRes = http.post(
    createApiUrl("/orders"),
    JSON.stringify({
      search: "",
      type: "partial",
      page: 1,
      page_size: 20,
      sort_field: sortFieldOrder,
      sort_order: sortOrderOrder,
    }),
    { headers }
  );

  const okOrdersSorted = check(ordersHistorySortedRes, {
    "order history sorted loaded": (r) => r.status === 200,
  });
  if (!okOrdersSorted) {
    bench_uj_orders_sorted_fail_count.add(1);
    fail(`order history sorting by ${sortFieldOrder} ${sortOrderOrder} failed`);
  }

  const searchWord =
    searchWordList[exec.scenario.iterationInInstance % searchWordList.length]
      .keyword;
  // step 9: 注文検索
  const searchType = ["partial", "prefix"][
    exec.scenario.iterationInInstance % 2
  ];
  const searchOrderRes = http.post(
    createApiUrl("/orders"),
    JSON.stringify({
      search: searchWord,
      type: searchType,
      page: 1,
      page_size: 20,
      sort_field: "created_at",
      sort_order: "desc",
    }),
    { headers }
  );

  const okSearchOrder = check(searchOrderRes, {
    "order search successful": (r) => r.status === 200,
  });
  if (!okSearchOrder) {
    bench_uj_orders_search_fail_count.add(1);
    fail("order search failed");
  }

  // 全ステップ成功なら overall success をインクリメント
  bench_uj_success_count.add(1);
}
