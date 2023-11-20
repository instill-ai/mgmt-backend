import http from "k6/http";
import { check } from "k6";
import * as constant from "./const.js"
import * as mgmtPrivate from "./rest-private-user.js";
import * as mgmtPublic from "./rest-public-user.js"
import * as mgmtPublicWithJwt from "./rest-public-user-with-jwt.js"

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
    "headers": {
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

  } else {
    // ======== Public API with jwt-sub
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
    mgmtPublic.CheckPublicMetrics(header);
  }

}

export function teardown(data) {

}
