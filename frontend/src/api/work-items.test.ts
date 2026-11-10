import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  createComment,
  fallbackWebsocketURL,
  listComments,
  listWorkItems,
  validateWorkItemAttributes,
} from "@/api/work-items";

describe("validateWorkItemAttributes", () => {
  beforeEach(() => {
    vi.unstubAllGlobals();
  });

  it("posts attributes to the same-origin validate proxy", async () => {
    const body = {
      valid: false,
      violations: [
        { field: "story_points", message: "story points must be greater than 0" },
      ],
    };
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(body), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      validateWorkItemAttributes("project/1", {
        type: "user_story",
        title: "As a student I want AC",
        description: "As a student I want AC so that planning works",
        story_points: 0,
        acceptance_criteria: ["Given a backlog When I create Then AC saves"],
      }),
    ).resolves.toEqual(body);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/projects/project%2F1/work-items/validate",
      {
        method: "POST",
        body: JSON.stringify({
          type: "user_story",
          title: "As a student I want AC",
          description: "As a student I want AC so that planning works",
          story_points: 0,
          acceptance_criteria: ["Given a backlog When I create Then AC saves"],
        }),
        cache: "no-store",
        headers: { "content-type": "application/json" },
      },
    );
  });

  it("rejects malformed validation responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ ok: true }), { status: 200 }),
      ),
    );

    await expect(
      validateWorkItemAttributes("p1", {
        type: "task",
        title: "Ship slice 12",
      }),
    ).rejects.toThrow("Invalid work item validation response");
  });
});

describe("listWorkItems", () => {
  it("requires an explicit Project id (no seed env fallback)", async () => {
    await expect(listWorkItems({})).rejects.toThrow("A Project must be selected");
    await expect(listWorkItems({ sprintId: "s1" })).rejects.toThrow(
      "A Project must be selected",
    );
  });
});

describe("Work Item comments client", () => {
  beforeEach(() => {
    vi.unstubAllGlobals();
  });

  const apiComment = {
    id: "c1",
    work_item_id: "wi/1",
    author_user_id: "u1",
    author_name: "Ada",
    author_email: "ada@example.com",
    body: "Ship it",
    created_at: "2026-03-01T12:00:00Z",
  };

  it("lists comments through the same-origin proxy and maps fields", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify([apiComment]), { status: 200 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(listComments("wi/1")).resolves.toEqual([
      {
        id: "c1",
        workItemId: "wi/1",
        authorUserId: "u1",
        authorName: "Ada",
        authorEmail: "ada@example.com",
        body: "Ship it",
        createdAt: "2026-03-01T12:00:00Z",
      },
    ]);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/work-items/wi%2F1/comments",
      expect.objectContaining({ cache: "no-store" }),
    );
  });

  it("creates a comment through the same-origin proxy", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(apiComment), { status: 201 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(createComment("wi-1", "Ship it")).resolves.toMatchObject({
      id: "c1",
      body: "Ship it",
      authorName: "Ada",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/work-items/wi-1/comments",
      {
        method: "POST",
        body: JSON.stringify({ body: "Ship it" }),
        cache: "no-store",
        headers: { "content-type": "application/json" },
      },
    );
  });

  it("surfaces unauthorized create errors clearly", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ error: "forbidden" }), { status: 403 }),
      ),
    );

    await expect(createComment("wi-1", "Nope")).rejects.toThrow(
      "forbidden (403)",
    );
  });

  it("rejects malformed comment list responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify([{ id: "bad" }]), { status: 200 }),
      ),
    );

    await expect(listComments("wi-1")).rejects.toThrow(
      "Invalid Work Item comments response",
    );
  });
});


describe("fallbackWebsocketURL", () => {
  it("uses Kong on 127.0.0.1:8083 for a local dev page", () => {
    expect(
      fallbackWebsocketURL({ protocol: "http:", hostname: "localhost", host: "localhost:3000" }),
    ).toBe("ws://127.0.0.1:8083/ws/work-items");
  });

  it("stays same-origin on any deployed host", () => {
    expect(
      fallbackWebsocketURL({ protocol: "https:", hostname: "chura.example.com", host: "chura.example.com" }),
    ).toBe("wss://chura.example.com/ws/work-items");
  });
});
