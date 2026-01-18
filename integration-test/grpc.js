import http from "k6/http";
import { check } from "k6";
import * as constant from "./const.js"

/*
 * gRPC Integration Tests
 *
 * Note: gRPC tests are disabled because we always use the API Gateway,
 * which exposes HTTP endpoints via grpc-gateway transcoding, not native gRPC.
 *
 * All integration testing is done via REST tests (rest.js).
 * This file only performs a basic login check to verify the test setup works.
 */

export let options = {
  setupTimeout: "300s",
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export function setup() {
  var loginResp = http.request("POST", `${constant.mgmtPublicHost}/auth/login`, JSON.stringify({
    "username": constant.defaultUsername,
    "password": constant.defaultPassword,
  }))

  check(loginResp, {
    [`POST /${constant.mgmtVersion}/auth/login response status is 200`]: (
      r
    ) => r.status === 200,
  });

  return {
    metadata: {
      "authorization": `Bearer ${loginResp.json().accessToken}`
    }
  }
}

export default function (header) {
  // gRPC tests are skipped - all testing done via REST (rest.js)
}

export function teardown(data) {
}
