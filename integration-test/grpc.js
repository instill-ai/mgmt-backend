import http from "k6/http";
import { check } from "k6";
import * as constant from "./const.js"
import * as mgmtPrivate from "./grpc-private-user.js";
import * as mgmtPublic from "./grpc-public-user.js"
import * as mgmtPublicWithJwt from "./grpc-public-user-with-jwt.js"

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
    "metadata": {
      "Authorization": `Bearer ${loginResp.json().access_token}`
    }
  }
}

export default function (header) {
  /*
   * Management API - API CALLS
   */

  if (!constant.apiGatewayMode) {
    // ======== Private API
    mgmtPrivate.CheckPrivateListUsersAdmin();
    mgmtPrivate.CheckPrivateGetUserAdmin();
    mgmtPrivate.CheckPrivateLookUpUserAdmin();
    mgmtPrivate.CheckPrivateSubtractCredit();
  } else {
    // ======== Public API with instill-user-uid
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
  }

}

export function teardown(data) {

}
