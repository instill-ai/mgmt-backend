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

export function setup() {}

export default function (data) {
  /*
   * Management API - API CALLS
   */

  if (__ENV.MODE != "api-gateway" && __ENV.MODE != "localhost") {

    // ======== Private API
    mgmtPrivate.CheckPrivateListUsersAdmin();
    mgmtPrivate.CheckPrivateCreateUserAdmin();
    mgmtPrivate.CheckPrivateGetUserAdmin();
    mgmtPrivate.CheckPrivateLookUpUserAdmin();
    mgmtPrivate.CheckPrivateUpdateUserAdmin();
    mgmtPrivate.CheckPrivateDeleteUserAdmin();
    mgmtPrivate.CheckPrivateValidateToken();

    // // ======== Public API with jwt-sub
    mgmtPublicWithJwt.CheckPublicQueryAuthenticatedUser();
    mgmtPublicWithJwt.CheckPublicPatchAuthenticatedUser();

  }

  // ======== Public API
  mgmtPublic.CheckHealth();
  mgmtPublic.CheckPublicQueryAuthenticatedUser();
  mgmtPublic.CheckPublicPatchAuthenticatedUser();
  mgmtPublic.CheckPublicCreateToken();
  mgmtPublic.CheckPublicListTokens();
  mgmtPublic.CheckPublicGetToken();
  mgmtPublic.CheckPublicDeleteToken();
}

export function teardown(data) {

}
