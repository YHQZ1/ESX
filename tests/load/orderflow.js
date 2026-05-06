import http from "k6/http";
import { check } from "k6";

export const options = {
  scenarios: {
    ramp_up: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 100 }, // Fast ramp to 100 VUs
        { duration: "30s", target: 100 }, // Hold steady
        { duration: "10s", target: 200 }, // Spike to 200 VUs
        { duration: "20s", target: 200 }, // Hold spike
        { duration: "10s", target: 0 }, // Cool down
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<100"],
    http_req_failed: ["rate<0.01"],
  },
};

const GATEWAY_URL = "http://localhost:8080";

export function setup() {
  const buyerKeys = [];
  const sellerKeys = [];

  for (let i = 0; i < 20; i++) {
    const bKey = __ENV[`BUYER_KEY_${i}`];
    if (!bKey) throw new Error(`Missing BUYER_KEY_${i} env var`);
    buyerKeys.push(bKey);

    const sKey = __ENV[`SELLER_KEY_${i}`];
    if (!sKey) throw new Error(`Missing SELLER_KEY_${i} env var`);
    sellerKeys.push(sKey);
  }

  return { buyerKeys, sellerKeys };
}

export default function (data) {
  const vuIndex = __VU - 1;
  const buyerKey = data.buyerKeys[vuIndex % data.buyerKeys.length];

  // Select a random seller for each iteration to avoid row-lock contention
  const sellerKey =
    data.sellerKeys[Math.floor(Math.random() * data.sellerKeys.length)];

  const price = 49500 + Math.floor(Math.random() * 1000);
  const qty = Math.floor(Math.random() * 5) + 1;
  const isSell = __ITER % 2 === 0;

  if (isSell) {
    const sellRes = http.post(
      `${GATEWAY_URL}/orders`,
      JSON.stringify({
        symbol: "RELIANCE",
        side: "SELL",
        type: "LIMIT",
        quantity: qty,
        price: price,
      }),
      {
        headers: {
          "Content-Type": "application/json",
          "x-api-key": sellerKey, // <-- Use the randomized seller key here
        },
      },
    );
    check(sellRes, {
      "sell order accepted": (r) => r.status === 201,
      "sell order_id present": (r) => r.json("order_id") !== undefined,
    });
  } else {
    const buyRes = http.post(
      `${GATEWAY_URL}/orders`,
      JSON.stringify({
        symbol: "RELIANCE",
        side: "BUY",
        type: "LIMIT",
        quantity: qty,
        price: price,
      }),
      {
        headers: {
          "Content-Type": "application/json",
          "x-api-key": buyerKey,
        },
      },
    );
    check(buyRes, {
      "buy order accepted": (r) => r.status === 201,
      "buy order_id present": (r) => r.json("order_id") !== undefined,
    });
  }
}
