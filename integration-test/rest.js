import http from "k6/http";
import {check, group} from "k6";

import * as constant from "./const.js";
import * as user from "./rest-user.js";

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

  // Health check
  group("Management API: Health check", () => {
    check(http.request("GET", `${constant.mgmtHost}/health/mgmt`), {
      [`GET /${constant.mgmtVersion}/health/mgmt response status is 200`]: (
        r
      ) => r.status === 200,
    });
  });

  // User
  user.CheckList();
  user.CheckCreate();
  user.CheckGet();
  user.CheckLookUp();
  user.CheckUpdate();
  user.CheckDelete();
}

export function teardown(data) {}
