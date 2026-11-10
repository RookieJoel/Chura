import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextResponse } from "next/server";
import { proxyBackendGet } from "@/lib/backend-proxy";
import { GET } from "./route";

vi.mock("@/lib/backend-proxy", () => ({ proxyBackendGet: vi.fn() }));

describe("templates route", () => {
  beforeEach(() => vi.mocked(proxyBackendGet).mockReset());

  it("forwards template discovery to the authenticated backend proxy", async () => {
    const request = new Request("http://localhost/api/v1/templates");
    vi.mocked(proxyBackendGet).mockResolvedValue(NextResponse.json([]));

    await GET(request);

    expect(proxyBackendGet).toHaveBeenCalledWith(
      request,
      "/api/v1/templates",
    );
  });
});
