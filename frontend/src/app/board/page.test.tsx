import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { getSprintMetrics } from "@/api/analytics";
import { getActiveSprint } from "@/api/sprints";
import { listWorkItems, updateWorkItem } from "@/api/work-items";
import BoardPage from "@/app/board/page";
import type { SprintMetrics } from "@/types/analytics";
import type { Sprint, WorkItem } from "@/types/work-item";

const memberUser = {
  authenticated: true as const,
  user_id: "u1",
  email: "member@example.com",
  roles: ["member"],
};

const auditorUser = {
  authenticated: true as const,
  user_id: "a1",
  email: "auditor@example.com",
  roles: ["auditor"],
};

const useProjectsMock = vi.hoisted(() =>
  vi.fn((): {
    selectedProjectId: string | null;
    selectedProject?: {
      id: string;
      members: { id: string; name: string; initials: string; color: string }[];
    } | null;
    loadState: "loading" | "error" | "ready";
    currentUser: { authenticated: true; user_id: string; email: string; roles: string[] };
  } => ({
    selectedProjectId: "p1" as string | null,
    loadState: "ready" as "loading" | "error" | "ready",
    currentUser: {
      authenticated: true as const,
      user_id: "u1",
      email: "member@example.com",
      roles: ["member"],
    },
  })),
);

vi.mock("@/components/layout/app-shell", () => ({
  AppShell: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
vi.mock("@/components/layout/project-provider", () => ({
  useProjects: () => useProjectsMock(),
}));
vi.mock("@/api/sprints", () => ({ getActiveSprint: vi.fn() }));
vi.mock("@/api/analytics", () => ({ getSprintMetrics: vi.fn() }));
vi.mock("@/api/work-items", () => ({
  listWorkItems: vi.fn(),
  updateWorkItem: vi.fn(),
}));

const activeSprint: Sprint = {
  id: "s-active",
  projectId: "p1",
  name: "Sprint 2",
  team: "Alpha",
  startDate: "2026-02-01",
  endDate: "2026-02-14",
  status: "active",
};

const activeMetrics: SprintMetrics = {
  sprint: activeSprint,
  workSummary: {
    planned: { items: 2, points: 8 },
    completed: { items: 0, points: 0 },
    carryOver: { items: 2, points: 8 },
    blocked: { items: 0, points: 0 },
  },
  completionRate: 0,
  pointsCompletionRate: 0,
  statusBreakdown: { todo: 1, in_progress: 1 },
  typeBreakdown: { task: 2 },
  progress: {
    totalDays: 14,
    elapsedDays: 3,
    remainingDays: 11,
    timeElapsedRate: 21.4,
  },
  workItems: [],
  calculatedAt: "2026-02-03T12:00:00Z",
};

function workItem(overrides: Partial<WorkItem> & Pick<WorkItem, "id" | "status" | "storyPoints">): WorkItem {
  return {
    projectId: "p1",
    title: overrides.id,
    type: "task",
    blocked: false,
    priority: "medium",
    epic: "",
    sprintId: "s-active",
    labels: [],
    acceptanceCriteria: [],
    createdAt: "2026-02-01T00:00:00Z",
    updatedAt: "2026-02-01T00:00:00Z",
    ...overrides,
  };
}

describe("BoardPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useProjectsMock.mockReturnValue({
      selectedProjectId: "p1",
      loadState: "ready",
      currentUser: memberUser,
    });
    vi.mocked(getSprintMetrics).mockResolvedValue(activeMetrics);
  });

  it("prompts for a Project when none is selected", async () => {
    useProjectsMock.mockReturnValue({
      selectedProjectId: null,
      loadState: "ready",
      currentUser: memberUser,
    });

    render(<BoardPage />);

    expect(
      await screen.findByRole("heading", {
        name: "Select a Project to open the Board",
      }),
    ).toBeInTheDocument();
    expect(getActiveSprint).not.toHaveBeenCalled();
    expect(listWorkItems).not.toHaveBeenCalled();
  });

  it("links to Backlog and Sprints when the selected Project has no active Sprint", async () => {
    vi.mocked(getActiveSprint).mockResolvedValue(null);
    // Members come from the Project already loaded by the provider.
    useProjectsMock.mockReturnValue({
      selectedProjectId: "p1",
      selectedProject: {
        id: "p1",
        members: [{ id: "m1", name: "Ada", initials: "A", color: "#111" }],
      },
      loadState: "ready",
      currentUser: memberUser,
    });

    render(<BoardPage />);

    expect(await screen.findByText("No active sprint")).toBeInTheDocument();
    expect(
      screen.getByText(/Sprint metrics are unavailable until a Sprint is active/i),
    ).toBeInTheDocument();
    expect(screen.getByText("Ada")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open Backlog" })).toHaveAttribute(
      "href",
      "/backlog",
    );
    expect(screen.getByRole("link", { name: "Manage sprints" })).toHaveAttribute(
      "href",
      "/sprints",
    );
    expect(getActiveSprint).toHaveBeenCalledWith("p1");
    expect(listWorkItems).not.toHaveBeenCalled();
  });

  it("loads only the selected Project's active Sprint work items and column points", async () => {
    vi.mocked(getActiveSprint).mockResolvedValue(activeSprint);
    vi.mocked(listWorkItems).mockResolvedValue([
      workItem({ id: "WI-1", title: "Todo item", status: "todo", storyPoints: 3 }),
      workItem({
        id: "WI-2",
        title: "Doing item",
        status: "in_progress",
        storyPoints: 5,
      }),
    ]);

    render(<BoardPage />);

    expect(await screen.findByText("Sprint 2")).toBeInTheDocument();
    expect(screen.getByText("8 story points")).toBeInTheDocument();
    expect(screen.getByText("Todo item")).toBeInTheDocument();
    expect(screen.getByText("Doing item")).toBeInTheDocument();
    expect(listWorkItems).toHaveBeenCalledWith({
      sprintId: "s-active",
      projectId: "p1",
    });
    expect(getSprintMetrics).toHaveBeenCalledWith("s-active");
    expect(await screen.findByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("Carry-over")).toBeInTheDocument();
    expect(screen.getByLabelText("Sprint time elapsed")).toBeInTheDocument();

    expect(screen.getByText("3 pts")).toBeInTheDocument();
    expect(screen.getByText("5 pts")).toBeInTheDocument();
    expect(screen.getByLabelText("Status for Todo item")).toHaveValue("todo");
    expect(screen.getByLabelText("Status for Doing item")).toHaveValue(
      "in_progress",
    );
  });

  it("keeps the board usable when metrics fail and does not invent progress", async () => {
    vi.mocked(getActiveSprint).mockResolvedValue(activeSprint);
    vi.mocked(listWorkItems).mockResolvedValue([
      workItem({ id: "WI-1", title: "Still visible", status: "todo", storyPoints: 2 }),
    ]);
    vi.mocked(getSprintMetrics).mockRejectedValue(new Error("forbidden (403)"));

    render(<BoardPage />);

    expect(await screen.findByText("Still visible")).toBeInTheDocument();
    expect(
      await screen.findByText(/Sprint metrics unavailable/i),
    ).toBeInTheDocument();
    expect(screen.getByText("forbidden (403)")).toBeInTheDocument();
    expect(screen.queryByLabelText("Sprint time elapsed")).not.toBeInTheDocument();
    expect(screen.queryByText("Planned")).not.toBeInTheDocument();
  });

  it("lets Auditors inspect the board without status mutation controls", async () => {
    useProjectsMock.mockReturnValue({
      selectedProjectId: "p1",
      loadState: "ready",
      currentUser: auditorUser,
    });
    vi.mocked(getActiveSprint).mockResolvedValue(activeSprint);
    vi.mocked(listWorkItems).mockResolvedValue([
      workItem({ id: "WI-1", title: "Observe me", status: "todo", storyPoints: 2 }),
    ]);

    render(<BoardPage />);

    expect(await screen.findByText("Observe me")).toBeInTheDocument();
    expect(screen.queryByLabelText("Status for Observe me")).not.toBeInTheDocument();
    expect(listWorkItems).toHaveBeenCalledWith({
      sprintId: "s-active",
      projectId: "p1",
    });
  });

  it("persists status changes and refreshes board point totals", async () => {
    const todoItem = workItem({
      id: "WI-1",
      title: "Move me",
      status: "todo",
      storyPoints: 3,
    });
    vi.mocked(getActiveSprint).mockResolvedValue(activeSprint);
    vi.mocked(listWorkItems).mockResolvedValue([todoItem]);
    vi.mocked(updateWorkItem).mockResolvedValue({
      ...todoItem,
      status: "in_progress",
    });

    render(<BoardPage />);

    expect(await screen.findByText("Move me")).toBeInTheDocument();
    expect(screen.getByText("3 story points")).toBeInTheDocument();
    expect(screen.getByLabelText("Status for Move me")).toHaveValue("todo");

    fireEvent.change(screen.getByLabelText("Status for Move me"), {
      target: { value: "in_progress" },
    });

    await waitFor(() => {
      expect(updateWorkItem).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "WI-1",
          status: "in_progress",
          projectId: "p1",
        }),
      );
    });

    await waitFor(() => {
      expect(screen.getByLabelText("Status for Move me")).toHaveValue(
        "in_progress",
      );
    });
    expect(screen.getByText("3 pts")).toBeInTheDocument();
    expect(screen.getByText("3 story points")).toBeInTheDocument();
    await waitFor(() => {
      expect(getSprintMetrics).toHaveBeenCalledTimes(2);
    });
  });

  it("clears previous board data when the selected Project changes", async () => {
    vi.mocked(getActiveSprint).mockResolvedValue(activeSprint);
    vi.mocked(listWorkItems).mockResolvedValue([
      workItem({ id: "WI-1", title: "Project A item", status: "todo", storyPoints: 2 }),
    ]);

    const { rerender } = render(<BoardPage />);
    expect(await screen.findByText("Project A item")).toBeInTheDocument();

    useProjectsMock.mockReturnValue({
      selectedProjectId: "p2",
      loadState: "ready",
      currentUser: memberUser,
    });
    vi.mocked(getActiveSprint).mockResolvedValue(null);
    vi.mocked(listWorkItems).mockClear();

    rerender(<BoardPage />);

    expect(await screen.findByText("No active sprint")).toBeInTheDocument();
    expect(screen.queryByText("Project A item")).not.toBeInTheDocument();
    expect(getActiveSprint).toHaveBeenLastCalledWith("p2");
    expect(listWorkItems).not.toHaveBeenCalled();
  });
});
