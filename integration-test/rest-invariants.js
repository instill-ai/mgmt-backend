import {
    check,
    group
} from "k6";
import http from "k6/http";
import {
    randomString
} from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

import * as constant from "./const.js";

/**
 * TEST SUITE: AIP Resource Refactoring Invariants
 *
 * PURPOSE:
 * Tests the critical invariants defined in the AIP Resource Refactoring plan.
 * These invariants ensure the system maintains data integrity and follows AIP standards.
 *
 * NOTE: Users and Organizations have special handling in AIP refactoring:
 * - They are namespace resources where slug = id for URL purposes
 * - The id is mutable (users can change their username)
 * - uid is the immutable internal identifier
 *
 * RED FLAGS (RF) - Hard Invariants:
 * - RF-2: name is the only canonical identifier
 *
 * YELLOW FLAGS (YF) - Strict Guardrails:
 * - YF-2: Slug resolution must not leak into services
 */

export function checkInvariants(header) {

    // ===============================================================
    // RF-2: name is the Only Canonical Identifier
    // For users/orgs: name format is "users/{id}" or "organizations/{id}"
    // ===============================================================
    group("RF-2: name is the canonical identifier (Users)", () => {
        // Test: Get authenticated user and verify name format
        group("Verify user name format", () => {
            const getResp = http.get(
                `${constant.mgmtPublicHost}/user`,
                header
            );

            check(getResp, {
                "[RF-2] GET /user returns 200": (r) => r.status === 200,
                "[RF-2] user has name field": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && body.user.name;
                },
                "[RF-2] user name starts with users/": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && body.user.name && body.user.name.startsWith("users/");
                },
                "[RF-2] user name format matches pattern": (r) => {
                    const body = JSON.parse(r.body);
                    if (!body.user || !body.user.name) return false;
                    // Pattern: users/{id}
                    const pattern = new RegExp(`^users/[^/]+$`);
                    return pattern.test(body.user.name);
                },
                "[RF-2] user id equals last segment of name": (r) => {
                    const body = JSON.parse(r.body);
                    if (!body.user || !body.user.name || !body.user.id) return false;
                    const segments = body.user.name.split("/");
                    return segments[segments.length - 1] === body.user.id;
                }
            });
        });

        // Test: Get user by ID via users endpoint
        group("Verify user accessible by ID", () => {
            const userId = constant.defaultUsername;
            const getResp = http.get(
                `${constant.mgmtPublicHost}/users/${userId}`,
                header
            );

            check(getResp, {
                "[RF-2] GET /users/{id} returns 200": (r) => r.status === 200,
                "[RF-2] Response contains correct user": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && body.user.id === userId;
                }
            });
        });
    });

    // ===============================================================
    // User/Org Exception: AIP refactoring - uid is internal only
    // Per AIP refactoring: uid is NOT exposed in API (internal use only)
    // Only id (username) is exposed and used in URLs
    // ===============================================================
    group("User/Org Exception: id is the public identifier", () => {
        // Test: User has id field (uid is internal only, not exposed)
        group("Verify user has id (uid is internal)", () => {
            const getResp = http.get(
                `${constant.mgmtPublicHost}/user`,
                header
            );

            check(getResp, {
                "[UserOrg] user has id field": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && body.user.id;
                },
                "[UserOrg] id is the username (not UUID)": (r) => {
                    const body = JSON.parse(r.body);
                    if (!body.user || !body.user.id) return false;
                    // ID should be the username, not a UUID
                    const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
                    return !uuidPattern.test(body.user.id);
                },
                "[UserOrg] user has slug field": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && body.user.slug !== undefined;
                },
                "[UserOrg] user has aliases array": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && Array.isArray(body.user.aliases);
                }
            });
        });
    });

    // ===============================================================
    // YF-2: Backend only accepts canonical IDs
    // For users: id is the canonical identifier used in URLs
    // ===============================================================
    group("YF-2: Backend only accepts canonical IDs", () => {
        // Test: GET user by valid ID should work
        group("GET user by valid ID succeeds", () => {
            const userId = constant.defaultUsername;
            const getResp = http.get(
                `${constant.mgmtPublicHost}/users/${userId}`,
                header
            );

            check(getResp, {
                "[YF-2a] GET by valid ID returns 200": (r) => r.status === 200,
                "[YF-2a] GET by valid ID returns correct user": (r) => {
                    const body = JSON.parse(r.body);
                    return body.user && body.user.id === userId;
                }
            });
        });

        // Test: GET by invalid/non-existent ID should fail
        group("GET by invalid ID fails", () => {
            const getResp = http.get(
                `${constant.mgmtPublicHost}/users/non-existent-user-id-12345`,
                header
            );

            check(getResp, {
                "[YF-2b] GET by invalid ID returns 404": (r) => r.status === 404
            });
        });
    });

    // ===============================================================
    // Access Token Tests: Hierarchical name format
    // ===============================================================
    group("Access Tokens: name format validation", () => {
        const randomSuffix = randomString(8);
        let tokenID;
        let tokenName;

        // Create test access token
        group("Setup: Create test access token", () => {
            const createPayload = {
                id: `test-token-${randomSuffix}`,
                ttl: 3600 // 1 hour TTL
            };

            const createResp = http.post(
                `${constant.mgmtPublicHost}/tokens`,
                JSON.stringify(createPayload),
                header
            );

            if (createResp.status === 200 || createResp.status === 201) {
                const body = JSON.parse(createResp.body);
                if (body.token) {
                    tokenID = body.token.id;
                    tokenName = body.token.name;
                }
            }
        });

        if (tokenID && tokenName) {
            // Test: Access token name format
            // Current format is: tokens/{id} (flat pattern)
            group("Verify access token name format", () => {
                check({ tokenName, tokenID }, {
                    "[Token] name starts with tokens/": (d) => d.tokenName.startsWith("tokens/"),
                    "[Token] name ends with token id": (d) => d.tokenName.endsWith(d.tokenID),
                    "[Token] name format matches pattern": (d) => {
                        // Pattern: tokens/{id}
                        const pattern = new RegExp(`^tokens/[^/]+$`);
                        return pattern.test(d.tokenName);
                    }
                });
            });

            // Test: id is derived from name (last segment)
            group("Verify token id is derived from name", () => {
                check({ tokenName, tokenID }, {
                    "[Token] id equals last segment of name": (d) => {
                        const segments = d.tokenName.split("/");
                        return segments[segments.length - 1] === d.tokenID;
                    }
                });
            });

            // Cleanup
            http.request(
                "DELETE",
                `${constant.mgmtPublicHost}/tokens/${tokenID}`,
                null,
                header
            );
        } else {
            console.log("Access token creation failed, skipping token tests");
        }
    });
}
