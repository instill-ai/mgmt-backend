import * as constant from "./const.js";
import * as privateAPI from "./rest-private-user.js";
import * as publicAPI from "./rest-public-user.js"

export let options = {
  setupTimeout: "300s",
  insecureSkipTLSVerify: true,
  thresholds: {
    checks: ["rate == 1.0"],
  },
};

export function setup() {
  constant;
}

export default function (data) {
  /*
   * Management API - API CALLS
   */

  if (__ENV.MODE != "api-gateway" && __ENV.MODE != "localhost") {

    // ======== Private API
    privateAPI.CheckPrivateListUsersAdmin();
    privateAPI.CheckPrivateCreateUserAdmin();
    privateAPI.CheckPrivateGetUserAdmin();
    privateAPI.CheckPrivateLookUpUserAdmin();
    privateAPI.CheckPrivateUpdateUserAdmin();
    privateAPI.CheckPrivateDeleteUserAdmin();
  }

  // ======== Public API
  publicAPI.CheckHealth();
  publicAPI.CheckPublicQueryAuthenticatedUser();
  publicAPI.CheckPublicPatchAuthenticatedUser();
}

export function teardown(data) {

}
