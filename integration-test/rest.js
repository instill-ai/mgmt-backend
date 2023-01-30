import http from "k6/http";
import {check, group} from "k6";

import * as constant from "./const.js";
import * as adminAPI from "./rest-admin-user.js";
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

  // ======== Admin API
  adminAPI.CheckAdminList();
  adminAPI.CheckAdminCreate();
  adminAPI.CheckAdminGet();
  adminAPI.CheckAdminLookUp();
  adminAPI.CheckAdminUpdate();
  adminAPI.CheckAdminDelete();

  // ======== Public API
  publicAPI.CheckHealth();
  publicAPI.CheckPublicGet();
  publicAPI.CheckPublicUpdate();
}

export function teardown(data) {}
