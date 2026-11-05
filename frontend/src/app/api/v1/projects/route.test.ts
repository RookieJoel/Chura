import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextResponse } from "next/server";
import { proxyBackendGet, proxyBackendPost } from "@/lib/backend-proxy";
import { GET, POST } from "./route";

vi.mock("@/lib/backend-proxy", () => ({
  proxyBackendGet: vi.fn(),
  proxyBackendPost: vi.fn(),
}));

describe("projects route", () => {
  beforeEach(() => {
    vi.mocked(proxyBackendGet).mockReset();
    vi.mocked(proxyBackendPost).mockReset();
  });

  it("forwards the pagination query to the authenticated backend proxy", async () => {
    const request = new Request("http://localhost/api/v1/projects?page=2&page_size=30");
    vi.mocked(proxyBackendGet).mockResolvedValue(NextResponse.json({}));

    await GET(request);

    expect(proxyBackendGet).toHaveBeenCalledWith(
      request,
      "/api/v1/projects?page=2&page_size=30",
    );
  });

  it("forwards Project creation requests and bodies to the authenticated backend proxy", async () => {
    const request = new Request("http://localhost/api/v1/projects", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        name: "Project",
        description: "Description",
        template_id: "se",
      }),
    });
    vi.mocked(proxyBackendPost).mockResolvedValue(NextResponse.json({}, { status: 201 }));

    await POST(request);

    expect(proxyBackendPost).toHaveBeenCalledWith(
      request,
      "/api/v1/projects",
    );
  });
});
