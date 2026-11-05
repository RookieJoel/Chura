import { beforeEach, describe, expect, it, vi } from "vitest";
import { getTemplate, listTemplates } from "@/api/templates";

const templates = [
  { id: "general", name: "General", mode: "general", creator_role: "owner" },
  { id: "se", name: "Software Engineering", mode: "se", creator_role: "product_owner" },
];

describe("template API", () => {
  beforeEach(() => {
    vi.unstubAllGlobals();
  });

  it("lists templates through the same-origin authenticated proxy", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(templates), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(listTemplates()).resolves.toEqual(templates);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/templates", {
      cache: "no-store",
      headers: { "content-type": "application/json" },
    });
  });

  it("retrieves a selected template with an encoded ID", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(templates[1]), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getTemplate("se/project")).resolves.toEqual(templates[1]);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/templates/se%2Fproject", {
      cache: "no-store",
      headers: { "content-type": "application/json" },
    });
  });

  it("surfaces API errors and rejects malformed template responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn()
        .mockResolvedValueOnce(
          new Response(JSON.stringify({ error: "Unauthenticated" }), {
            status: 401,
          }),
        )
        .mockResolvedValueOnce(
          new Response(JSON.stringify([{ id: "missing-fields" }]), {
            status: 200,
          }),
        ),
    );

    await expect(listTemplates()).rejects.toThrow("Unauthenticated (401)");
    await expect(listTemplates()).rejects.toThrow("Invalid template list response");
  });
});
