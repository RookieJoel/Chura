import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { assignProjectAuditor } from "@/api/projects";
import { updateWorkItem } from "@/api/work-items";
import { WorkItemDetailEditor } from "@/components/work-items/work-item-detail-editor";
import type { WorkItem } from "@/types/work-item";

vi.mock("next/navigation", () => ({ useRouter: () => ({ push: vi.fn() }) }));
vi.mock("@/api/projects", () => ({ getProject: vi.fn(), assignProjectAuditor: vi.fn() }));
vi.mock("@/api/work-items", () => ({
  deleteWorkItem: vi.fn(),
  updateWorkItem: vi.fn(),
  validateWorkItemAttributes: vi.fn(),
}));
vi.mock("@/components/work-items/work-item-comments", () => ({
  WorkItemComments: () => null,
}));
vi.mock("@/components/layout/project-provider", () => ({
  useProjects: () => ({
    projects: [{ id: "p1", name: "Website Refresh" }],
    selectedProject: null,
    selectedTemplate: {
      id: "general",
      name: "General",
      mode: "general",
      work_item_types: ["task"],
      roles: ["owner", "member"],
      default_role: "member",
      creator_role: "owner",
    },
    selectedProjectLoadState: "ready",
    currentUser: { authenticated: true, user_id: "u1", email: "u1@example.com", roles: ["member"] },
  }),
}));

const baseItem: WorkItem = {
  id: "wi-1",
  projectId: "p1",
  title: "Build the homepage",
  description: "",
  type: "task",
  status: "todo",
  blocked: false,
  priority: "medium",
  storyPoints: 0,
  epic: "",
  sprintId: null,
  labels: [],
  acceptanceCriteria: [],
  features: { environment: "Staging" },
  createdAt: "2026-11-02T09:00:00Z",
  updatedAt: "2026-11-02T09:00:00Z",
};

describe("WorkItemDetailEditor custom fields", () => {
  beforeEach(() => {
    vi.mocked(updateWorkItem).mockReset();
    vi.mocked(updateWorkItem).mockImplementation(async (input) => ({
      ...baseItem,
      features: input.features,
    }) as WorkItem);
  });

  it("shows saved custom fields as a field name and its details", () => {
    render(<WorkItemDetailEditor initialItem={baseItem} members={[]} sprints={[]} />);

    expect(screen.getByText("Custom fields")).toBeInTheDocument();
    expect(screen.getByLabelText("Custom field 1 name")).toHaveValue("environment");
    expect(screen.getByLabelText("Custom field 1 details")).toHaveValue("Staging");
  });

  it("keeps the same input while a whole field name is typed, then saves it", async () => {
    render(<WorkItemDetailEditor initialItem={baseItem} members={[]} sprints={[]} />);

    fireEvent.click(screen.getByRole("button", { name: "Add field" }));
    const nameInput = screen.getByLabelText("Custom field 2 name") as HTMLInputElement;
    nameInput.focus();
    // One character at a time: before the fix each keystroke rebuilt the row.
    let typed = "";
    for (const char of "browser") {
      typed += char;
      fireEvent.change(screen.getByLabelText("Custom field 2 name"), { target: { value: typed } });
      expect(screen.getByLabelText("Custom field 2 name")).toBe(nameInput);
    }
    expect(nameInput).toHaveFocus();
    expect(nameInput).toHaveValue("browser");

    fireEvent.change(screen.getByLabelText("Custom field 2 details"), {
      target: { value: "Safari 17" },
    });
    fireEvent.blur(screen.getByLabelText("Custom field 2 details"));
    await waitFor(() => expect(updateWorkItem).toHaveBeenCalledTimes(1));
    expect(vi.mocked(updateWorkItem).mock.calls[0][0].features).toEqual({
      environment: "Staging",
      browser: "Safari 17",
    });
  });

  it("asks for a field name instead of saving a nameless field", () => {
    render(<WorkItemDetailEditor initialItem={baseItem} members={[]} sprints={[]} />);

    fireEvent.click(screen.getByRole("button", { name: "Add field" }));
    fireEvent.change(screen.getByLabelText("Custom field 2 details"), { target: { value: "x" } });
    fireEvent.blur(screen.getByLabelText("Custom field 2 details"));

    expect(screen.getByText("Every custom field needs a field name")).toBeInTheDocument();
    expect(updateWorkItem).not.toHaveBeenCalled();
  });
});

describe("WorkItemDetailEditor Auditor", () => {
  const auditors = [{ id: "a1", name: "Ning Wongsa", initials: "NW", color: "#7C3AED" }];

  beforeEach(() => {
    vi.mocked(updateWorkItem).mockReset();
    vi.mocked(assignProjectAuditor).mockReset();
    vi.mocked(updateWorkItem).mockImplementation(async (input) => ({
      ...baseItem,
      reporterId: input.reporterId,
    }) as WorkItem);
  });

  it("assigns the picked Auditor to the Project, then saves them on the Work Item", async () => {
    const calls: string[] = [];
    vi.mocked(assignProjectAuditor).mockImplementation(async (projectId, auditorId) => {
      calls.push(`assign:${projectId}:${auditorId}`);
      return { project_id: projectId, auditor_user_id: auditorId, auditor_name: "", auditor_email: "", assigned_at: "" };
    });
    vi.mocked(updateWorkItem).mockImplementation(async (input) => {
      calls.push(`save:${input.reporterId}`);
      return { ...baseItem, reporterId: input.reporterId } as WorkItem;
    });
    render(<WorkItemDetailEditor initialItem={baseItem} members={[]} auditors={auditors} sprints={[]} />);

    expect(screen.getByLabelText("Auditor")).toHaveValue("");
    expect(screen.queryByLabelText("Reporter")).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Auditor"), { target: { value: "a1" } });

    await waitFor(() => expect(calls).toEqual(["assign:p1:a1", "save:a1"]));
  });

  it("does not save the Work Item when the Project assignment fails", async () => {
    vi.mocked(assignProjectAuditor).mockRejectedValue(new Error("choose a user with the Auditor role (400)"));
    render(<WorkItemDetailEditor initialItem={baseItem} members={[]} auditors={auditors} sprints={[]} />);

    fireEvent.change(screen.getByLabelText("Auditor"), { target: { value: "a1" } });

    expect(await screen.findByText(/choose a user with the Auditor role/)).toBeInTheDocument();
    expect(updateWorkItem).not.toHaveBeenCalled();
  });
});
