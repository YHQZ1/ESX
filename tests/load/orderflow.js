import http from "k6/http";
import { check } from "k6";

export const options = {
  scenarios: {
    ramp_up: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 20 },
        { duration: "30s", target: 20 },
        { duration: "10s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<100"],
    http_req_failed: ["rate<0.01"],
  },
};

const GATEWAY_URL = "http://localhost:8080";
const REGISTRY_URL = "http://localhost:8081";

// The setup function runs ONCE before the load test begins.
export function setup() {
  const ts = new Date().getTime();
  const email = `k6_trader_${ts}@esx.com`;

  // 1. Register User
  const regRes = http.post(
    `${REGISTRY_URL}/participants/register`,
    JSON.stringify({
      name: "k6 Load Tester",
      email: email,
    }),
    { headers: { "Content-Type": "application/json" } },
  );

  if (regRes.status !== 201) {
    throw new Error(`Failed to register participant: ${regRes.body}`);
  }

  const apiKey = regRes.json("api_key");
  const participantId = regRes.json("participant_id");

  // 2. Fund User: Increased to 100,000,000,000 paise (100 Crore paise)
  http.post(
    `${REGISTRY_URL}/participants/${participantId}/deposit`,
    JSON.stringify({
      amount: 100000000000,
    }),
    { headers: { "Content-Type": "application/json" } },
  );

  return { apiKey: apiKey };
}

// The default function runs repeatedly for every Virtual User (VU)
export default function (data) {
  const price = 49500 + Math.floor(Math.random() * 1000);

  const payload = JSON.stringify({
    symbol: "RELIANCE",
    side: "BUY",
    type: "LIMIT",
    quantity: Math.floor(Math.random() * 5) + 1,
    price: price,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
      "x-api-key": data.apiKey,
    },
  };

  const res = http.post(`${GATEWAY_URL}/orders`, payload, params);

  check(res, {
    "status is 201": (r) => r.status === 201,
    "order accepted": (r) => r.json("order_id") !== undefined,
  });
}
