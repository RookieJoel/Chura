import { describe, expect, it, vi } from "vitest";
import {
  addProjectMember,
  assignProjectRole,
  createProject,
  deleteProject,
  getProject,
  listProjectAuditors,
  listProjectFeedback,
  listProjects,
  removeProjectMember,
  selfAssignAuditor,
  submitProjectFeedback,
  updateProject,
} from "@/api/projects";

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
        created_by: "u1",
        members: [
          {
            user_id: "u1",
            name: "Owner User",
            email: "owner@example.com",
            role: "product_owner",
          },
        ],
      };
      const fetchMock = vi.fn().mockResolvedValue(
        new Response(JSON.stringify(project), { status: 200 }),
      );
      vi.stubGlobal("fetch", fetchMock);

      await expect(getProject("project/id")).resolves.toMatchObject({
        ...project,
        members: [
          {
            id: "u1",
            email: "owner@example.com",
            name: "Owner User",
            role: "product_owner",
            initials: "OU",
            color: "#64748B",
          },
        ],
      });
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/projects/project%2Fid",
        {
          cache: "no-store",
          headers: { "content-type": "application/json" },
        },
      );
    });

    it("rejects malformed Project detail data instead of treating members as empty", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(
          new Response(
            JSON.stringify({
              id: "p1",
              name: "Alpha",
              description: "",
              template_id: "se",
              mode: "se",
              created_by: "u1",
            }),
            { status: 200 },
          ),
        ),
      );

      await expect(getProject("p1")).rejects.toThrow(
        "Invalid Project detail response",
      );
    });

    it("uses explicit authenticated endpoints for member mutations and normalizes their responses", async () => {
      const response = {
        id: "p1",
        name: "Alpha",
        description: "",
        template_id: "se",
        mode: "se",
        created_by: "owner",
        members: [
          {
            user_id: "u/2",
            email: "member@example.com",
            name: "Project Member",
            role: "",
          },
        ],
      };
      const fetchMock = vi.fn().mockImplementation(
        () => new Response(JSON.stringify(response), { status: 200 }),
      );
      vi.stubGlobal("fetch", fetchMock);

      await expect(addProjectMember("p1", "member@example.com")).resolves.toMatchObject({
        members: [{ id: "u/2", role: "", initials: "PM" }],
      });
      await assignProjectRole("p1", "u/2", "developer");
      await removeProjectMember("p1", "u/2");

      expect(fetchMock).toHaveBeenNthCalledWith(
        1,
        "/api/v1/projects/p1/members",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ email: "member@example.com" }),
        }),
      );
      expect(fetchMock).toHaveBeenNthCalledWith(
        2,
        "/api/v1/projects/p1/members/u%2F2/role",
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify({ role: "developer" }),
        }),
      );
      expect(fetchMock).toHaveBeenNthCalledWith(
        3,
        "/api/v1/projects/p1/members/u%2F2",
        expect.objectContaining({ method: "DELETE" }),
      );
    });
  });

  it("calls auditor, feedback, update, and delete endpoints", async () => {
    const assignment = {
      project_id: "p1",
      auditor_user_id: "a1",
      auditor_name: "Auditor",
      auditor_email: "a1@example.com",
      assigned_at: "2026-01-02T03:04:05Z",
    };
    const feedback = {
      id: "f1",
      project_id: "p1",
      author_user_id: "a1",
      author_name: "Auditor",
      author_email: "a1@example.com",
      body: "note",
      created_at: "2026-01-02T03:05:05Z",
    };
    const project = {
      id: "p1",
      name: "Alpha",
      description: "",
      template_id: "se",
      mode: "se",
      created_by: "u1",
      members: [],
    };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify(assignment), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify([assignment]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify([feedback]), { status: 200 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(feedback), { status: 201 }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(project), { status: 200 }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(selfAssignAuditor("p1")).resolves.toEqual(assignment);
    await expect(listProjectAuditors("p1")).resolves.toEqual([assignment]);
    await expect(listProjectFeedback("p1")).resolves.toEqual([feedback]);
    await expect(submitProjectFeedback("p1", "note")).resolves.toEqual(feedback);
    await expect(updateProject("p1", { name: "Beta" })).resolves.toMatchObject({
      id: "p1",
    });
    await expect(deleteProject("p1")).resolves.toBeUndefined();
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
