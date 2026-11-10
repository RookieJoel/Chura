import { afterEach, describe, expect, it, vi } from "vitest";
import {
  createSprint,
  getActiveSprint,
  listSprints,
  updateSprintStatus,
} from "@/api/sprints";

describe("sprints api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("lists sprints for the selected Project id", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: "1",
            project_id: "p1",
            name: "Sprint 1",
            team: "Alpha",
            status: "planned",
          },
        ]),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(listSprints("p1")).resolves.toEqual([
      {
        id: "1",
        projectId: "p1",
        name: "Sprint 1",
        team: "Alpha",
        startDate: "",
        endDate: "",
        status: "planned",
      },
    ]);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints?project_id=p1",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("requires a Project id before creating a Sprint", async () => {
    await expect(
      createSprint({ projectId: "  ", name: "Sprint 1", team: "Alpha" }),
    ).rejects.toThrow("A Project must be selected");
  });

  it("returns only the active Sprint for the selected Project", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: "1",
            project_id: "p1",
            name: "Sprint 1",
            team: "Alpha",
            status: "planned",
          },
          {
            id: "2",
            project_id: "p1",
            name: "Sprint 2",
            team: "Alpha",
            status: "active",
          },
          {
            id: "3",
            project_id: "p1",
            name: "Sprint 0",
            team: "Alpha",
            status: "completed",
          },
        ]),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getActiveSprint("p1")).resolves.toMatchObject({
      id: "2",
      status: "active",
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints?project_id=p1",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("returns null when the selected Project has no active Sprint", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: "1",
            project_id: "p1",
            name: "Sprint 1",
            team: "Alpha",
            status: "planned",
          },
        ]),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getActiveSprint("p1")).resolves.toBeNull();
  });

  it("updates sprint status through the authenticated proxy", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "1",
          project_id: "p1",
          name: "Sprint 1",
          team: "Alpha",
          status: "active",
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      updateSprintStatus(
        {
          id: "1",
          projectId: "p1",
          name: "Sprint 1",
          team: "Alpha",
          startDate: "",
          endDate: "",
          status: "planned",
        },
        "active",
      ),
    ).resolves.toMatchObject({ status: "active" });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/1",
      expect.objectContaining({ method: "PUT" }),
    );
  });
});
