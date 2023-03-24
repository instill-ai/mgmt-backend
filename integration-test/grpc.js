import * as constant from "./const.js";
import * as privateAPI from "./grpc-private-user.js";
import * as publicAPI from "./grpc-public-user.js"

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
    privateAPI.CheckAdminList();
    privateAPI.CheckAdminCreate();
    privateAPI.CheckAdminGet();
    privateAPI.CheckAdminLookUp();
    privateAPI.CheckAdminUpdate();
    privateAPI.CheckAdminDelete();
  }

  // ======== Public API
  publicAPI.CheckHealth();
  publicAPI.CheckPublicGet();
  publicAPI.CheckPublicUpdate();
}

export function teardown(data) {

}
