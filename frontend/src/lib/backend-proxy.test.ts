import { beforeEach, describe, expect, it, vi } from "vitest";
import { getToken } from "next-auth/jwt";
import { proxyBackendPost } from "@/lib/backend-proxy";

vi.mock("next-auth/jwt", () => ({ getToken: vi.fn() }));
vi.mock("../../auth", () => ({ sessionCookieName: "authjs.session-token" }));

describe("proxyBackendPost", () => {
  beforeEach(() => {
    process.env.CHURA_API_URL = "http://backend.test";
    process.env.AUTH_SECRET = "test-secret";
    vi.mocked(getToken).mockReset();
    vi.unstubAllGlobals();
  });

  it("requires a session access token before forwarding a request", async () => {
    vi.mocked(getToken).mockResolvedValue(null);
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const request = new Request("http://localhost/api/v1/projects", {
      method: "POST",
      body: "{}",
    });

    const response = await proxyBackendPost(request, "/api/v1/projects");

    expect(response.status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("forwards the authenticated body and backend error status unchanged", async () => {
    vi.mocked(getToken).mockResolvedValue({ accessToken: "test-access-token" });
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: "name is required" }), {
        status: 400,
        headers: { "content-type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const request = new Request("http://localhost/api/v1/projects", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ name: "" }),
    });

    const response = await proxyBackendPost(request, "/api/v1/projects");

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("http://backend.test/api/v1/projects");
    expect(init.method).toBe("POST");
    const headers = new Headers(init.headers);
    expect(headers.get("Authorization")).toBe("Bearer test-access-token");
    expect(headers.get("content-type")).toBe("application/json");
    expect(init.body).toBeInstanceOf(ArrayBuffer);
    expect(new TextDecoder().decode(init.body as ArrayBuffer)).toBe(
      JSON.stringify({ name: "" }),
    );
    expect(init.cache).toBe("no-store");
    expect(response.status).toBe(400);
    await expect(response.json()).resolves.toEqual({ error: "name is required" });
  });
});
