import { beforeEach, describe, expect, it, vi } from "vitest";
import { proxyBackendGet, proxyBackendPost } from "@/lib/backend-proxy";
import { GET, POST } from "./route";

vi.mock("@/lib/backend-proxy", () => ({
  proxyBackendGet: vi.fn(),
  proxyBackendPost: vi.fn(),
}));

describe("Work Item comments routes", () => {
  beforeEach(() => {
    vi.mocked(proxyBackendGet).mockReset();
    vi.mocked(proxyBackendPost).mockReset();
  });

  it("forwards list requests with encoded Work Item IDs", async () => {
    const request = new Request(
      "http://localhost/api/v1/work-items/wi%2F1/comments",
    );
    const context = { params: Promise.resolve({ id: "wi/1" }) };
    await GET(request, context);

    expect(proxyBackendGet).toHaveBeenCalledWith(
      request,
      "/api/v1/work-items/wi%2F1/comments",
    );
  });

  it("forwards create requests to the authenticated proxy", async () => {
    const request = new Request(
      "http://localhost/api/v1/work-items/wi-1/comments",
      {
        method: "POST",
        body: JSON.stringify({ body: "Looks good" }),
      },
    );
    const context = { params: Promise.resolve({ id: "wi-1" }) };
    await POST(request, context);

    expect(proxyBackendPost).toHaveBeenCalledWith(
      request,
      "/api/v1/work-items/wi-1/comments",
    );
  });
});
