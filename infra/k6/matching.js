import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Counter, Rate } from "k6/metrics";

const matchLatency = new Trend("match_latency", true);
const bookDepthLatency = new Trend("book_depth_latency", true);
const matchesFired = new Counter("matches_fired");
const matchesFailed = new Rate("matches_failed");

export const options = {
  scenarios: {
    book_builder: {
      executor: "constant-vus",
      vus: 5,
      duration: "20s",
      exec: "buildBook",
    },
    market_taker: {
      executor: "ramping-vus",
      startVUs: 1,
      stages: [
        { duration: "10s", target: 5 },
        { duration: "20s", target: 15 },
        { duration: "10s", target: 1 },
      ],
      exec: "takeFromBook",
      startTime: "5s",
    },
  },
  thresholds: {
    match_latency: ["p(95)<200", "p(99)<500"],
    matches_failed: ["rate<0.05"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const REGISTRY_URL = __ENV.REGISTRY_URL || "http://localhost:8081";

const SELLER_API_KEY =
  __ENV.SELLER_API_KEY ||
  "d2f36fa3c1544c57e16c024fd48435b8bb627950f8727d06cb6ba168c0e50cd6";
const BUYER_API_KEY =
  __ENV.BUYER_API_KEY ||
  "9e214dc45c6db75645a1598bd60ec875db9192ed81f9da23c4093c4ea3af96bd";

const PRICE_LEVELS = [48000, 49000, 49500, 50000, 50500, 51000, 52000];

export function setup() {
  const depositRes = http.post(
    `${REGISTRY_URL}/participants/86303d8d-4429-41de-9a03-66c72d3fe06e/deposit`,
    JSON.stringify({ amount: 100000000 }),
    { headers: { "Content-Type": "application/json" } },
  );
  check(depositRes, { "buyer funded": (r) => r.status === 200 });
}

export function buildBook() {
  const price = PRICE_LEVELS[Math.floor(Math.random() * PRICE_LEVELS.length)];

  const res = http.post(
    `${BASE_URL}/orders`,
    JSON.stringify({
      symbol: "RELIANCE",
      side: "SELL",
      type: "LIMIT",
      quantity: 5,
      price: price,
    }),
    {
      headers: {
        "Content-Type": "application/json",
        "x-api-key": SELLER_API_KEY,
      },
    },
  );

  check(res, {
    "resting sell placed": (r) => r.status === 201,
  });

  sleep(0.2);
}

export function takeFromBook() {
  const price = 55000;

  const start = Date.now();
  const res = http.post(
    `${BASE_URL}/orders`,
    JSON.stringify({
      symbol: "RELIANCE",
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
  matchLatency.add(Date.now() - start);
  matchesFired.add(1);

  const ok = check(res, {
    "buy order accepted": (r) => r.status === 201,
  });

  if (!ok) {
    matchesFailed.add(1);
  }

  sleep(0.05);
}
