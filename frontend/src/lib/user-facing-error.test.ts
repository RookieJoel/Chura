import { describe, expect, it } from "vitest";
import { userFacingError } from "@/lib/user-facing-error";

describe("userFacingError", () => {
  it("maps known API error codes to readable copy", () => {
    expect(userFacingError(new Error("forbidden (403)"), "fallback")).toBe(
      "You don't have permission to do that.",
    );
    expect(userFacingError("not found", "fallback")).toBe(
      "This Project was not found or is no longer available.",
    );
    expect(
      userFacingError(
        new Error("user is already a member of this project (409)"),
        "fallback",
      ),
    ).toBe("That person is already a member of this Project.");
  });

  it("keeps readable backend sentences and uses the fallback for opaque codes", () => {
    expect(
      userFacingError(new Error("Project members unavailable"), "fallback"),
    ).toBe("Project members unavailable");
    expect(userFacingError(new Error("xyzzy"), "Unable to save")).toBe(
      "Unable to save",
    );
    expect(userFacingError(undefined, "Unable to save")).toBe("Unable to save");
  });
});
