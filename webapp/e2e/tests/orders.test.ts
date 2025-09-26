import { test, expect } from "@playwright/test";
import knapsackResults from "../tests/sampleData/ordersKnapSackResults.json";

test.describe("GET api/robot/delivery-plan?capacity=100", () => {
  test("配送計画を取得できること (200: OK)", async ({ request }) => {
    const dataIndex = Number(process.env.USE_DATAINDEX ?? 0);

    const response = await request.get("/api/robot/delivery-plan?capacity=50", {
      headers: {
        "X-API-KEY": "test-robot-key",
      },
    });
    const responseBody = await response.json();
    const sortOrdersById = (orders) =>
      [...orders].sort((a, b) => a.order_id - b.order_id);

    const sortedResponseOrders = sortOrdersById(responseBody.orders);
    const sortedExpectedOrders = sortOrdersById(
      knapsackResults[dataIndex].orders
    );

    // 元のオブジェクトのordersを、ソート済みの配列で上書きする// これにより、オブジェクト全体をtoEqualで一度に比較できる
    const actual = { ...responseBody, orders: sortedResponseOrders };
    const expected = {
      ...knapsackResults[dataIndex],
      orders: sortedExpectedOrders,
    };
    expect(actual).toEqual(expected);
    expect(response.status()).toBe(200);
  });
});
