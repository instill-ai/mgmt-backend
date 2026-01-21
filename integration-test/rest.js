import http from "k6/http";
import { check } from "k6";
import encoding from "k6/encoding";
import * as constant from "./const.js"
import * as mgmtPublic from "./rest-public-user.js"
import * as mgmtPublicWithJwt from "./rest-public-user-with-jwt.js"
import * as restInvariants from "./rest-invariants.js"

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

  return {
    "headers": {
      "Authorization": `Basic ${basicAuth}`
    }
  }
}

export default function (header) {
  /*
   * Management API - REST Integration Tests
   *
   * All tests run via API Gateway (grpc-gateway transcoding).
   */

  // ======== Public API with instill-user-uid (auth validation)
  mgmtPublicWithJwt.CheckPublicGetUser();
  mgmtPublicWithJwt.CheckPublicPatchAuthenticatedUser();

  // ======== Public API
  mgmtPublic.CheckHealth();
  mgmtPublic.CheckPublicGetUser(header);
  mgmtPublic.CheckPublicPatchAuthenticatedUser(header);
  mgmtPublic.CheckPublicCreateToken(header);
  mgmtPublic.CheckPublicListTokens(header);
  mgmtPublic.CheckPublicGetToken(header);
  mgmtPublic.CheckPublicDeleteToken(header);
  mgmtPublic.CheckPublicGetRemainingCredit(header);
  mgmtPublic.CheckPublicMetrics(header);

  // AIP Resource Refactoring Invariants
  restInvariants.checkInvariants(header);
}

export function teardown(data) {

}
