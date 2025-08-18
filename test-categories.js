import http from "k6/http";
import { check } from "k6";

export const options = {
  stages: [
    { duration: "10s", target: 50 }, // ramp up ke 50 VU
    { duration: "50s", target: 100 }, // tahan di 100 VU
    { duration: "10s", target: 0 }, // ramp down ke 0
  ],
  thresholds: {
    http_req_duration: ["p(95)<100"], // 95% response time < 100ms
    http_req_failed: ["rate==0"], // error rate harus 0%
  },
};

export default function () {
  const res = http.get("http://localhost:3000/api/v2/categories");

  // Validasi response
  check(res, {
    "status is 200": (r) => r.status === 200,
    "body is not empty": (r) => r.body && r.body.length > 0,
  });
}
