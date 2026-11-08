import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  exportReviewSummary,
  getReviewSummary,
  listReflections,
  submitReflection,
} from "@/api/analytics";
import { listSprints } from "@/api/sprints";
import ReportsPage from "@/app/reports/page";
import type { SprintReflection, SprintReviewSummary } from "@/types/analytics";
import type { Sprint } from "@/types/work-item";

const { useProjectsMock, selectedTemplate, memberUser } = vi.hoisted(() => {
  const selectedTemplate = {
    id: "se",
    name: "Software Engineering",
    mode: "se",
    work_item_types: ["user_story", "task", "bug"],
    roles: ["product_owner", "developer"],
    default_role: "developer",
    creator_role: "product_owner",
    reflection: {
      topics: [
        {
          key: "went_well",
          question: "What went well?",
          required: true,
          max_runes: 2000,
        },
        {
          key: "stop",
          question: "What should we stop doing?",
          required: true,
          max_runes: 2000,
        },
        {
          key: "start",
          question: "What should we start doing?",
          required: true,
          max_runes: 2000,
        },
      ],
    },
    sprint_review: {
      title: "Sprint Review Summary",
      metrics: [
        {
          key: "planned_items",
          label: "Planned work items",
          unit: "items",
        },
        {
          key: "completion_rate",
          label: "Completion rate",
          unit: "percent",
        },
      ],
      default_metrics: ["planned_items", "completion_rate"],
    },
  };

  const memberUser = {
    authenticated: true as const,
    user_id: "u1",
    email: "member@example.com",
    roles: ["member"],
  };

  return {
    selectedTemplate,
    memberUser,
    useProjectsMock: vi.fn(() => ({
      selectedProjectId: "p1" as string | null,
      selectedTemplate,
      currentUser: memberUser,
      loadState: "ready" as "loading" | "error" | "ready",
    })),
  };
});

vi.mock("@/components/layout/app-shell", () => ({
  AppShell: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
vi.mock("@/components/layout/project-provider", () => ({
  useProjects: () => useProjectsMock(),
}));
vi.mock("@/api/sprints", () => ({ listSprints: vi.fn() }));
vi.mock("@/api/analytics", () => ({
  listReflections: vi.fn(),
  submitReflection: vi.fn(),
  getReviewSummary: vi.fn(),
  exportReviewSummary: vi.fn(),
}));

const sprint: Sprint = {
  id: "s1",
  projectId: "p1",
  name: "Sprint 1",
  team: "Alpha",
  startDate: "2026-02-01",
  endDate: "2026-02-14",
  status: "active",
};

const reflection: SprintReflection = {
  id: "r1",
  sprintId: "s1",
  projectId: "p1",
  authorId: "u1",
  authorEmail: "member@example.com",
  templateId: "se",
  answers: {
    went_well: "pairing",
    stop: "late merges",
    start: "tests first",
  },
  createdAt: "2026-02-10T12:00:00Z",
  updatedAt: "2026-02-10T12:00:00Z",
};

const summary: SprintReviewSummary = {
  title: "Sprint Review Summary",
  project: { id: "p1", name: "Demo" },
  sprint,
  templateId: "se",
  quantitative: {
    metrics: [
      {
        key: "planned_items",
        label: "Planned work items",
        unit: "items",
        value: 4,
      },
    ],
    progress: null,
  },
  qualitative: {
    topics: [
      {
        key: "went_well",
        question: "What went well?",
        responses: [
          {
            authorId: "u1",
            authorEmail: "member@example.com",
            answer: "pairing",
          },
        ],
      },
    ],
    respondents: 1,
    teamSize: 3,
    responseRate: 33.3,
  },
  generatedAt: "2026-02-14T12:00:00Z",
};

describe("ReportsPage", () => {
  beforeEach(() => {
    useProjectsMock.mockReset();
    useProjectsMock.mockReturnValue({
      selectedProjectId: "p1",
      selectedTemplate,
      currentUser: memberUser,
      loadState: "ready",
    });
    vi.mocked(listSprints).mockReset();
    vi.mocked(listReflections).mockReset();
    vi.mocked(submitReflection).mockReset();
    vi.mocked(getReviewSummary).mockReset();
    vi.mocked(exportReviewSummary).mockReset();

    vi.mocked(listSprints).mockResolvedValue([sprint]);
    vi.mocked(listReflections).mockResolvedValue([]);
    vi.mocked(submitReflection).mockResolvedValue(reflection);
    vi.mocked(getReviewSummary).mockResolvedValue(summary);
    vi.mocked(exportReviewSummary).mockResolvedValue("sprint-s1-review.csv");
  });

  it("loads reflections for the selected Sprint and submits a new one", async () => {
    render(<ReportsPage />);

    await waitFor(() => {
      expect(listSprints).toHaveBeenCalledWith("p1");
      expect(listReflections).toHaveBeenCalledWith("s1");
    });

    fireEvent.change(screen.getByRole("textbox", { name: /What went well/i }), {
      target: { value: "pairing" },
    });
    fireEvent.change(
      screen.getByRole("textbox", { name: /What should we stop doing/i }),
      { target: { value: "late merges" } },
    );
    fireEvent.change(
      screen.getByRole("textbox", { name: /What should we start doing/i }),
      { target: { value: "tests first" } },
    );
    fireEvent.click(screen.getByRole("button", { name: /Submit my reflection/i }));

    await waitFor(() => {
      expect(submitReflection).toHaveBeenCalledWith("s1", {
        went_well: "pairing",
        stop: "late merges",
        start: "tests first",
      });
    });

    const teamList = await screen.findByRole("list");
    expect(teamList).toHaveTextContent("member@example.com");
    expect(teamList).toHaveTextContent("pairing");
    expect(
      screen.getByRole("button", { name: /Update my reflection/i }),
    ).toBeInTheDocument();
  });

  it("generates and exports a review summary", async () => {
    render(<ReportsPage />);

    await waitFor(() => expect(listReflections).toHaveBeenCalledWith("s1"));

    fireEvent.click(screen.getByRole("checkbox", { name: "Planned work items" }));
    fireEvent.click(screen.getByRole("button", { name: /Generate summary/i }));

    await waitFor(() => {
      expect(getReviewSummary).toHaveBeenCalledWith("s1", ["planned_items"]);
    });

    expect(await screen.findByText("Sprint Review Summary")).toBeInTheDocument();
    expect(screen.getByText("4")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Export CSV/i }));
    await waitFor(() => {
      expect(exportReviewSummary).toHaveBeenCalledWith("s1", {
        format: "csv",
        metrics: ["planned_items"],
      });
    });
  });

  it("hides the reflection form for auditors", async () => {
    useProjectsMock.mockReturnValue({
      selectedProjectId: "p1",
      selectedTemplate,
      currentUser: {
        authenticated: true as const,
        user_id: "a1",
        email: "auditor@example.com",
        roles: ["auditor"],
      },
      loadState: "ready",
    });

    render(<ReportsPage />);

    await waitFor(() => expect(listReflections).toHaveBeenCalled());
    expect(
      screen.getByText(/Only Team Members can submit reflections/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /Submit my reflection/i }),
    ).not.toBeInTheDocument();
  });
});
