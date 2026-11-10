import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextResponse } from "next/server";
import { proxyBackendGet } from "@/lib/backend-proxy";
import { GET } from "./route";

vi.mock("@/lib/backend-proxy", () => ({ proxyBackendGet: vi.fn() }));

describe("template detail route", () => {
  beforeEach(() => vi.mocked(proxyBackendGet).mockReset());

  it("forwards the encoded template ID to the authenticated backend proxy", async () => {
    const request = new Request("http://localhost/api/v1/templates/se%2Ftest");
    vi.mocked(proxyBackendGet).mockResolvedValue(NextResponse.json({}));

    await GET(request, { params: Promise.resolve({ id: "se/test" }) });

    expect(proxyBackendGet).toHaveBeenCalledWith(
      request,
      "/api/v1/templates/se%2Ftest",
    );
  });
});
