import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { updateSprintStatus } from "@/api/sprints";
import { SprintActions } from "@/components/board/sprint-actions";
import type { Sprint } from "@/types/work-item";

vi.mock("@/api/sprints", () => ({
  updateSprintStatus: vi.fn(),
}));

const plannedSprint: Sprint = {
  id: "1",
  projectId: "p1",
  name: "Sprint 1",
  team: "Alpha",
  startDate: "2026-01-01",
  endDate: "2026-01-14",
  status: "planned",
};

describe("SprintActions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("disables Start when the Sprint has zero Work Items", () => {
    render(<SprintActions sprint={plannedSprint} workItemCount={0} />);

    const button = screen.getByRole("button", { name: "Start sprint" });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute(
      "title",
      "Assign work items to this Sprint before starting it",
    );
    expect(updateSprintStatus).not.toHaveBeenCalled();
  });

  it("starts the Sprint when Work Items are assigned", async () => {
    vi.mocked(updateSprintStatus).mockResolvedValue({
      ...plannedSprint,
      status: "active",
    });
    const onUpdated = vi.fn();

    render(
      <SprintActions
        sprint={plannedSprint}
        workItemCount={2}
        onUpdated={onUpdated}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Start sprint" }));

    await waitFor(() => {
      expect(updateSprintStatus).toHaveBeenCalledWith(plannedSprint, "active");
      expect(onUpdated).toHaveBeenCalledWith(
        expect.objectContaining({ status: "active" }),
      );
    });
  });

  it("surfaces backend empty-start validation errors", async () => {
    vi.mocked(updateSprintStatus).mockRejectedValue(
      new Error("sprint cannot start without work items (400)"),
    );

    render(<SprintActions sprint={plannedSprint} workItemCount={1} />);
    fireEvent.click(screen.getByRole("button", { name: "Start sprint" }));

    expect(
      await screen.findByText("sprint cannot start without work items (400)"),
    ).toBeInTheDocument();
  });

  it("confirms before ending and does not invent carry-over", async () => {
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);
    const active: Sprint = { ...plannedSprint, status: "active" };
    vi.mocked(updateSprintStatus).mockResolvedValue({
      ...active,
      status: "completed",
    });

    render(<SprintActions sprint={active} workItemCount={1} />);
    fireEvent.click(screen.getByRole("button", { name: "End sprint" }));

    await waitFor(() => {
      expect(confirm).toHaveBeenCalledWith(
        expect.stringContaining("no automatic carry-over"),
      );
      expect(updateSprintStatus).toHaveBeenCalledWith(active, "completed");
    });
    confirm.mockRestore();
  });
});
