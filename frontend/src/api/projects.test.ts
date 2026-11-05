import { describe, expect, it, vi } from "vitest";
import { createProject, getProject, listProjects } from "@/api/projects";

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
      headers: { "content-type": "application/json" },
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
      headers: { "content-type": "application/json" },
    });
  });

  describe("Project API", () => {
    it("creates a Project through the authenticated same-origin route", async () => {
      const project = {
        id: "p1",
        name: "Alpha",
        description: "",
        template_id: "se",
        mode: "se",
      };
      const fetchMock = vi.fn().mockResolvedValue(
        new Response(JSON.stringify(project), { status: 201 }),
      );
      vi.stubGlobal("fetch", fetchMock);

      await expect(
        createProject({
          name: "Alpha",
          description: "",
          template_id: "se",
        }),
      ).resolves.toEqual(project);
      expect(fetchMock).toHaveBeenCalledWith("/api/v1/projects", {
        method: "POST",
        body: JSON.stringify({
          name: "Alpha",
          description: "",
          template_id: "se",
        }),
        cache: "no-store",
        headers: { "content-type": "application/json" },
      });
    });

    it("surfaces creation authorization and validation errors", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(
          new Response(JSON.stringify({ error: "Team Member role required" }), {
            status: 403,
          }),
        ),
      );

      await expect(
        createProject({ name: "Alpha", description: "", template_id: "se" }),
      ).rejects.toThrow("Team Member role required (403)");
    });

    it("retrieves a Project through the authenticated backend proxy", async () => {
      const project = {
        id: "project/id",
        name: "Alpha",
        description: "",
        template_id: "se",
        mode: "se",
      };
      const fetchMock = vi.fn().mockResolvedValue(
        new Response(JSON.stringify(project), { status: 200 }),
      );
      vi.stubGlobal("fetch", fetchMock);

      await expect(getProject("project/id")).resolves.toEqual(project);
      expect(fetchMock).toHaveBeenCalledWith("/api/backend/projects/project%2Fid", {
        cache: "no-store",
        headers: { "content-type": "application/json" },
      });
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
