import { describe, expect, it, vi } from "vitest";
import { listProjects } from "@/api/projects";

describe("listProjects", () => {
  it("requests the first page from the same-origin authenticated proxy", async () => {
    const projects = {
      items: [
        { id: "p1", name: "Alpha", description: "", template_id: "se", mode: "se" },
      ],
      page: 1,
      page_size: 30,
      total: 1,
    };
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(projects), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(listProjects()).resolves.toEqual(projects.items);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects?page=1&page_size=30", {
      cache: "no-store",
    });
  });

  it("loads every page before returning projects", async () => {
    const firstPage = Array.from({ length: 30 }, (_, index) => ({
      id: `p${index + 1}`,
      name: `Project ${index + 1}`,
      description: "",
      template_id: "se",
      mode: "se",
    }));
    const lastPage = [
      { id: "p31", name: "Project 31", description: "", template_id: "se", mode: "se" },
      { id: "p32", name: "Project 32", description: "", template_id: "se", mode: "se" },
    ];
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({
        items: firstPage, page: 1, page_size: 30, total: 32,
      }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({
        items: lastPage, page: 2, page_size: 30, total: 32,
      }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(listProjects()).resolves.toHaveLength(32);
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/projects?page=2&page_size=30", {
      cache: "no-store",
    });
  });

  it("surfaces backend discovery errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "service unavailable" }), {
          status: 503,
        }),
      ),
    );

    await expect(listProjects()).rejects.toThrow("service unavailable");
  });

  it("does not return partial results when a later page fails", async () => {
    const firstPage = Array.from({ length: 30 }, (_, index) => ({
      id: `p${index + 1}`,
      name: `Project ${index + 1}`,
      description: "",
      template_id: "se",
      mode: "se",
    }));
    vi.stubGlobal(
      "fetch",
      vi.fn()
        .mockResolvedValueOnce(new Response(JSON.stringify({
          items: firstPage, page: 1, page_size: 30, total: 31,
        }), { status: 200 }))
        .mockResolvedValueOnce(new Response(JSON.stringify({ error: "service unavailable" }), {
          status: 503,
        })),
    );

    await expect(listProjects()).rejects.toThrow("service unavailable");
  });
});
