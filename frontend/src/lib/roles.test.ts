import { describe, expect, it } from "vitest";
import { canMutateExecution, hasAuditorRole } from "@/lib/roles";

describe("roles", () => {
  it("allows Team Members to mutate execution surfaces", () => {
    expect(canMutateExecution(["member"])).toBe(true);
    expect(canMutateExecution(["member", "auditor"])).toBe(true);
  });

  it("keeps pure Auditors observe-only for execution mutations", () => {
    expect(canMutateExecution(["auditor"])).toBe(false);
    expect(canMutateExecution([])).toBe(false);
    expect(canMutateExecution(null)).toBe(false);
  });

  it("detects Auditor role", () => {
    expect(hasAuditorRole(["auditor"])).toBe(true);
    expect(hasAuditorRole(["member"])).toBe(false);
  });
});
