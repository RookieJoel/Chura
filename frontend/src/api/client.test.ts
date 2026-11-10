import { afterEach, describe, expect, it, vi } from "vitest";
import { SESSION_EXPIRED_MESSAGE, sameOriginRequest } from "@/api/client";

describe("sameOriginRequest", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("surfaces validation violation messages from the backend", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            error: "invalid input",
            violations: [
              { field: "status", message: "sprint cannot start without work items" },
            ],
          }),
          { status: 400 },
        ),
      ),
    );

    await expect(sameOriginRequest("/api/backend/sprints/1", { method: "PUT" })).rejects.toThrow(
      "sprint cannot start without work items (400)",
    );
  });

  it.each([
    ["the Next proxy", { error: "Unauthenticated" }],
    ["the backend", { error: "unauthenticated" }],
  ])("turns a 401 from %s into a sign-in message", async (_source, body) => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify(body), { status: 401 })),
    );

    await expect(
      sameOriginRequest("/api/v1/projects/p1/sprints", { method: "POST" }),
    ).rejects.toThrow(SESSION_EXPIRED_MESSAGE);
  });
});
