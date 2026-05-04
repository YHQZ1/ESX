import http from "k6/http";
import ws from "k6/ws";
import { check, sleep } from "k6";
import { Trend, Counter, Rate } from "k6/metrics";

const orderSubmitLatency = new Trend("order_submit_latency", true);
const tradeFilledLatency = new Trend("trade_filled_latency", true);
const ordersSubmitted = new Counter("orders_submitted");
const ordersFilled = new Counter("orders_filled");
const ordersFailed = new Rate("orders_failed");

export const options = {
  scenarios: {
    warmup: {
      executor: "constant-vus",
      vus: 2,
      duration: "10s",
      gracefulStop: "5s",
    },
    sustained: {
      executor: "ramping-vus",
      startVUs: 2,
      stages: [
        { duration: "15s", target: 10 },
        { duration: "30s", target: 20 },
        { duration: "15s", target: 5 },
      ],
      gracefulRampDown: "5s",
      startTime: "15s",
    },
  },
  thresholds: {
    order_submit_latency: ["p(95)<500", "p(99)<1000"],
    orders_failed: ["rate<0.01"],
    http_req_failed: ["rate<0.01"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const REGISTRY_URL = __ENV.REGISTRY_URL || "http://localhost:8081";

const BUYER_API_KEY =
  __ENV.BUYER_API_KEY ||
  "9e214dc45c6db75645a1598bd60ec875db9192ed81f9da23c4093c4ea3af96bd";
const SELLER_API_KEY =
  __ENV.SELLER_API_KEY ||
  "d2f36fa3c1544c57e16c024fd48435b8bb627950f8727d06cb6ba168c0e50cd6";

const SYMBOLS = ["RELIANCE", "TCS", "INFY", "HDFC", "WIPRO"];

function randomSymbol() {
  return SYMBOLS[Math.floor(Math.random() * SYMBOLS.length)];
}

function randomPrice(base, spread) {
  return base + Math.floor(Math.random() * spread);
}

export function setup() {
  const depositRes = http.post(
    `${REGISTRY_URL}/participants/86303d8d-4429-41de-9a03-66c72d3fe06e/deposit`,
    JSON.stringify({ amount: 50000000 }),
    { headers: { "Content-Type": "application/json" } },
  );
  check(depositRes, { "buyer funded": (r) => r.status === 200 });
  return {};
}

export default function () {
  const symbol = randomSymbol();
  const price = randomPrice(49000, 2000);

  const sellStart = Date.now();
  const sellRes = http.post(
    `${BASE_URL}/orders`,
    JSON.stringify({
      symbol: symbol,
      side: "SELL",
      type: "LIMIT",
      quantity: 1,
      price: price,
    }),
    {
      headers: {
        "Content-Type": "application/json",
        "x-api-key": SELLER_API_KEY,
      },
    },
  );
  orderSubmitLatency.add(Date.now() - sellStart);
  ordersSubmitted.add(1);

  const sellOk = check(sellRes, {
    "sell order accepted": (r) => r.status === 201,
    "sell has order_id": (r) => {
      try {
        return JSON.parse(r.body).order_id !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (!sellOk) {
    ordersFailed.add(1);
    sleep(0.1);
    return;
  }

  sleep(0.05);

  const buyStart = Date.now();
  const buyRes = http.post(
    `${BASE_URL}/orders`,
    JSON.stringify({
      symbol: symbol,
      side: "BUY",
      type: "LIMIT",
      quantity: 1,
      price: price,
    }),
    {
      headers: {
        "Content-Type": "application/json",
        "x-api-key": BUYER_API_KEY,
      },
    },
  );
  orderSubmitLatency.add(Date.now() - buyStart);
  ordersSubmitted.add(1);

  const buyOk = check(buyRes, {
    "buy order accepted": (r) => r.status === 201,
    "buy order filled": (r) => {
      try {
        return JSON.parse(r.body).status === "filled";
      } catch {
        return false;
      }
    },
  });

  if (buyOk) {
    tradeFilledLatency.add(Date.now() - sellStart);
    ordersFilled.add(1);
  } else {
    ordersFailed.add(1);
  }

  sleep(0.1);
}

export function teardown() {
  console.log("Load test complete");
}
