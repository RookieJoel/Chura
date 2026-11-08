import { afterEach, describe, expect, it, vi } from "vitest";
import {
  exportReviewSummary,
  getReviewSummary,
  getSprintMetrics,
  getSprintWorkSummary,
  listReflections,
  submitReflection,
} from "@/api/analytics";

describe("analytics api", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("requires a Sprint id before fetching work-summary", async () => {
    await expect(getSprintWorkSummary("  ")).rejects.toThrow("A Sprint is required");
  });

  it("maps work-summary buckets from snake_case", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          planned: { items: 4, points: 13 },
          completed: { items: 1, points: 3 },
          carry_over: { items: 2, points: 8 },
          blocked: { items: 1, points: 2 },
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(getSprintWorkSummary("s1")).resolves.toEqual({
      planned: { items: 4, points: 13 },
      completed: { items: 1, points: 3 },
      carryOver: { items: 2, points: 8 },
      blocked: { items: 1, points: 2 },
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/s1/work-summary",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("maps metrics including null progress without inventing calendar data", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          sprint: {
            id: "s1",
            project_id: "p1",
            name: "Sprint 1",
            team: "Alpha",
            status: "active",
          },
          work_summary: {
            planned: { items: 2, points: 5 },
            completed: { items: 0, points: 0 },
            carry_over: { items: 2, points: 5 },
            blocked: { items: 0, points: 0 },
          },
          completion_rate: 0,
          points_completion_rate: 0,
          status_breakdown: { todo: 2 },
          type_breakdown: { task: 2 },
          progress: null,
          work_items: [
            {
              id: "WI-1",
              title: "Task",
              type: "task",
              status: "todo",
              priority: "medium",
              story_points: 5,
              category: "carry_over",
            },
          ],
          calculated_at: "2026-02-01T12:00:00Z",
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const metrics = await getSprintMetrics("s%1");
    expect(metrics.progress).toBeNull();
    expect(metrics.workSummary.planned).toEqual({ items: 2, points: 5 });
    expect(metrics.completionRate).toBe(0);
    expect(metrics.workItems).toEqual([
      {
        id: "WI-1",
        title: "Task",
        type: "task",
        status: "todo",
        priority: "medium",
        assigneeId: undefined,
        storyPoints: 5,
        category: "carry_over",
      },
    ]);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/s%251/metrics",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("surfaces unauthorized metrics failures", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "forbidden" }), { status: 403 }),
      ),
    );

    await expect(getSprintMetrics("s1")).rejects.toThrow("forbidden (403)");
  });

  it("lists reflections mapped from snake_case", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify([
          {
            id: "r1",
            sprint_id: "s1",
            project_id: "p1",
            author_id: "u1",
            author_email: "u1@example.com",
            template_id: "se",
            answers: { went_well: "demos" },
            created_at: "2026-02-01T12:00:00Z",
            updated_at: "2026-02-01T12:00:00Z",
          },
        ]),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(listReflections("s1")).resolves.toEqual([
      {
        id: "r1",
        sprintId: "s1",
        projectId: "p1",
        authorId: "u1",
        authorEmail: "u1@example.com",
        templateId: "se",
        answers: { went_well: "demos" },
        createdAt: "2026-02-01T12:00:00Z",
        updatedAt: "2026-02-01T12:00:00Z",
      },
    ]);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/s1/reflections",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("submits the caller's reflection via PUT", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          id: "r1",
          sprint_id: "s1",
          project_id: "p1",
          author_id: "u1",
          author_email: "u1@example.com",
          template_id: "se",
          answers: { went_well: "pairing", stop: "late merges", start: "tests" },
          created_at: "2026-02-01T12:00:00Z",
          updated_at: "2026-02-01T13:00:00Z",
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const reflection = await submitReflection("s1", {
      went_well: "pairing",
      stop: "late merges",
      start: "tests",
    });
    expect(reflection.answers.went_well).toBe("pairing");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/s1/reflections/me",
      expect.objectContaining({
        method: "PUT",
        body: JSON.stringify({
          answers: {
            went_well: "pairing",
            stop: "late merges",
            start: "tests",
          },
        }),
      }),
    );
  });

  it("maps review summary quantitative and qualitative sections", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          title: "Sprint Review Summary",
          project: { id: "p1", name: "Demo" },
          sprint: {
            id: "s1",
            project_id: "p1",
            name: "Sprint 1",
            team: "Alpha",
            status: "completed",
          },
          template_id: "se",
          quantitative: {
            metrics: [
              {
                key: "planned_items",
                label: "Planned work items",
                unit: "items",
                value: 4,
              },
            ],
            work_summary: {
              planned: { items: 4, points: 13 },
              completed: { items: 2, points: 5 },
              carry_over: { items: 2, points: 8 },
              blocked: { items: 0, points: 0 },
            },
            progress: null,
          },
          qualitative: {
            topics: [
              {
                key: "went_well",
                question: "What went well?",
                responses: [
                  {
                    author_id: "u1",
                    author_email: "u1@example.com",
                    answer: "pairing",
                  },
                ],
              },
            ],
            respondents: 1,
            team_size: 3,
            response_rate: 33.3,
          },
          generated_at: "2026-02-14T12:00:00Z",
        }),
        { status: 200 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const summary = await getReviewSummary("s1", ["planned_items"]);
    expect(summary.quantitative.progress).toBeNull();
    expect(summary.quantitative.metrics[0]).toEqual({
      key: "planned_items",
      label: "Planned work items",
      unit: "items",
      value: 4,
    });
    expect(summary.qualitative.topics[0].responses[0].answer).toBe("pairing");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/s1/review-summary?metrics=planned_items",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("downloads a review summary export and returns the filename", async () => {
    const click = vi.fn();
    const remove = vi.fn();
    const appendChild = vi
      .spyOn(document.body, "appendChild")
      .mockImplementation((node) => node);
    vi.spyOn(document, "createElement").mockImplementation((tag) => {
      if (tag === "a") {
        return {
          href: "",
          download: "",
          rel: "",
          click,
          remove,
        } as unknown as HTMLAnchorElement;
      }
      return document.createElement(tag);
    });
    vi.stubGlobal("URL", {
      createObjectURL: vi.fn(() => "blob:review"),
      revokeObjectURL: vi.fn(),
    });

    const fetchMock = vi.fn().mockResolvedValue(
      new Response("title,value\n", {
        status: 200,
        headers: {
          "content-type": "text/csv",
          "content-disposition": 'attachment; filename="sprint-s1-review.csv"',
        },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      exportReviewSummary("s1", { format: "csv", metrics: ["completion_rate"] }),
    ).resolves.toBe("sprint-s1-review.csv");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/backend/sprints/s1/review-summary/export?metrics=completion_rate&format=csv",
      expect.objectContaining({ cache: "no-store" }),
    );
    expect(click).toHaveBeenCalled();
    expect(remove).toHaveBeenCalled();
    appendChild.mockRestore();
  });
});
