import { beforeEach, describe, expect, it, vi } from "vitest";
import { getTemplate, listTemplates } from "@/api/templates";

const templates = [
  {
    id: "general",
    name: "General",
    mode: "general",
    work_item_types: ["task"],
    roles: ["owner", "member"],
    default_role: "member",
    creator_role: "owner",
  },
  {
    id: "se",
    name: "Software Engineering",
    mode: "se",
    work_item_types: ["user_story", "task", "bug"],
    roles: ["product_owner", "scrum_master", "developer"],
    default_role: "developer",
    creator_role: "product_owner",
    validation_rules: {
      required_fields: ["title"],
      story_points: { greater_than: 0 },
      type_rules: [
        {
          type: "user_story",
          description_format: "user_story",
          require_acceptance_criteria: true,
        },
      ],
    },
  },
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

    await expect(listTemplates()).rejects.toThrow(
      "Your session has expired. Please sign in again to continue.",
    );
    await expect(listTemplates()).rejects.toThrow("Invalid template list response");
  });
});
