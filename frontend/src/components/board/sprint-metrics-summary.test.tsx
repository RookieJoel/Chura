import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SprintMetricsSummary } from "@/components/board/sprint-metrics-summary";
import type { SprintMetrics } from "@/types/analytics";

const baseMetrics: SprintMetrics = {
  sprint: {
    id: "s1",
    projectId: "p1",
    name: "Sprint 1",
    team: "Alpha",
    startDate: "2026-02-01",
    endDate: "2026-02-14",
    status: "active",
  },
  workSummary: {
    planned: { items: 4, points: 13 },
    completed: { items: 1, points: 3 },
    carryOver: { items: 2, points: 8 },
    blocked: { items: 1, points: 2 },
  },
  completionRate: 25,
  pointsCompletionRate: 23.1,
  statusBreakdown: {},
  typeBreakdown: {},
  progress: {
    totalDays: 14,
    elapsedDays: 7,
    remainingDays: 7,
    timeElapsedRate: 50,
  },
  workItems: [],
  calculatedAt: "2026-02-07T12:00:00Z",
};

describe("SprintMetricsSummary", () => {
  it("renders work-summary buckets and completion rates", () => {
    render(<SprintMetricsSummary metrics={baseMetrics} />);

    expect(screen.getByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("4")).toBeInTheDocument();
    expect(screen.getByText("(13 pts)")).toBeInTheDocument();
    expect(screen.getByText("Completed")).toBeInTheDocument();
    expect(screen.getByText("Carry-over")).toBeInTheDocument();
    expect(screen.getByText("Blocked")).toBeInTheDocument();
    expect(screen.getByText("25%")).toBeInTheDocument();
    expect(screen.getByText("23.1%")).toBeInTheDocument();
    expect(screen.getByLabelText("Sprint time elapsed")).toHaveAttribute(
      "aria-valuenow",
      "50",
    );
  });

  it("does not invent a progress chart when calendar progress is missing", () => {
    render(
      <SprintMetricsSummary
        metrics={{
          ...baseMetrics,
          progress: null,
        }}
      />,
    );

    expect(screen.queryByLabelText("Sprint time elapsed")).not.toBeInTheDocument();
    expect(
      screen.getByText(
        "Sprint calendar progress unavailable until start and end dates are set.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByText("Planned")).toBeInTheDocument();
  });

  it("shows loading and error states without fake metrics", () => {
    const { rerender } = render(
      <SprintMetricsSummary metrics={null} loading />,
    );
    expect(screen.getByText("Loading sprint metrics...")).toBeInTheDocument();

    rerender(
      <SprintMetricsSummary metrics={null} error="Unable to load metrics (403)" />,
    );
    expect(
      screen.getByText(/Sprint metrics unavailable\. Showing available Project/i),
    ).toBeInTheDocument();
    expect(screen.getByText("Unable to load metrics (403)")).toBeInTheDocument();
    expect(screen.queryByText("Planned")).not.toBeInTheDocument();
  });
});
