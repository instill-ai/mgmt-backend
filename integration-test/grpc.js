import http from "k6/http";
import { check } from "k6";
import encoding from "k6/encoding";
import * as constant from "./const.js"

/*
 * gRPC Integration Tests
 *
 * Note: gRPC tests are disabled because we always use the API Gateway,
 * which exposes HTTP endpoints via grpc-gateway transcoding, not native gRPC.
 *
 * All integration testing is done via REST tests (rest.js).
 * This file only performs a basic health check to verify the test setup works.
 */

export let options = {
  setupTimeout: "300s",
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export function setup() {
  // CE edition uses Basic Auth for all authenticated requests
  const basicAuth = encoding.b64encode(`${constant.defaultUsername}:${constant.defaultPassword}`);

  // Verify health endpoint is accessible
  var healthResp = http.request("GET", `${constant.mgmtPublicHost}/health/mgmt`);
  check(healthResp, {
    [`GET /${constant.mgmtVersion}/health/mgmt response status is 200`]: (r) => r.status === 200,
  });

  return {
    metadata: {
      "authorization": `Basic ${basicAuth}`
    }
  }
}

export default function (header) {
  // gRPC tests are skipped - all testing done via REST (rest.js)
}

export function teardown(data) {
}
