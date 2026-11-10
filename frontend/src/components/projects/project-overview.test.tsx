import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  addProjectMember,
  assignProjectRole,
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
import { getCurrentIdentity } from "@/api/auth";
import { getTemplate, listTemplates } from "@/api/templates";
import { ProjectProvider } from "@/components/layout/project-provider";
import { ProjectOverview } from "@/components/projects/project-overview";

const { push } = vi.hoisted(() => ({ push: vi.fn() }));

vi.mock("@/api/projects", () => ({
  addProjectMember: vi.fn(),
  assignProjectRole: vi.fn(),
  deleteProject: vi.fn(),
  getProject: vi.fn(),
  listProjectAuditors: vi.fn(),
  listProjectFeedback: vi.fn(),
  listProjects: vi.fn(),
  removeProjectMember: vi.fn(),
  selfAssignAuditor: vi.fn(),
  submitProjectFeedback: vi.fn(),
  updateProject: vi.fn(),
}));
vi.mock("@/api/auth", () => ({ getCurrentIdentity: vi.fn() }));
vi.mock("@/api/templates", () => ({
  getTemplate: vi.fn(),
  listTemplates: vi.fn(),
}));
vi.mock("next/navigation", () => ({ useRouter: () => ({ push }) }));

const project = {
  id: "p1",
  name: "Alpha",
  description: "A project description",
  template_id: "se",
  mode: "se",
  created_by: "u1",
  members: [],
};
const template = {
  id: "se",
  name: "Software Engineering",
  mode: "se",
  work_item_types: ["user_story", "task", "bug"],
  roles: ["product_owner", "scrum_master", "developer"],
  default_role: "developer",
  creator_role: "product_owner",
};

describe("ProjectOverview", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    push.mockReset();
    vi.mocked(listProjects).mockResolvedValue([project]);
    vi.mocked(getProject).mockResolvedValue(project);
    vi.mocked(getTemplate).mockResolvedValue(template);
    vi.mocked(listTemplates).mockResolvedValue([template]);
    vi.mocked(listProjectAuditors).mockResolvedValue([]);
    vi.mocked(listProjectFeedback).mockResolvedValue([]);
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "u1",
      email: "owner@example.com",
      roles: ["member"],
    });
  });

  it("loads the requested Project details without Sprint or Work Item content", async () => {
    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    expect(await screen.findByRole("heading", { name: "Alpha" })).toBeInTheDocument();
    expect(screen.getByText("A project description")).toBeInTheDocument();
    expect(screen.getByText("No Project members.")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: /Sprint|Work Item/i })).not.toBeInTheDocument();
  });

  it("shows a clear access error when the requested Project cannot be loaded", async () => {
    vi.mocked(getProject).mockRejectedValue(new Error("forbidden (403)"));
    render(
      <ProjectProvider>
        <ProjectOverview projectId="missing" />
      </ProjectProvider>,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "You don't have permission to do that.",
    );
  });

  it("stages member changes and sends them only when Save settings is pressed", async () => {
    const withMembers = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "member@example.com",
          name: "Project Member",
          role: "developer",
          initials: "PM",
          color: "#64748B",
        },
      ],
    };
    const newcomer = {
      id: "u3",
      email: "new@example.com",
      name: "New Person",
      role: "",
      initials: "NP",
      color: "#64748B",
    };
    const calls: string[] = [];
    let server = withMembers;
    vi.mocked(getProject).mockResolvedValue(withMembers);
    vi.mocked(addProjectMember).mockImplementation(async (_p, email) => {
      calls.push(`add:${email}`);
      server = { ...server, members: [...server.members, newcomer] };
      return server;
    });
    vi.mocked(assignProjectRole).mockImplementation(async (_p, userId, role) => {
      calls.push(`role:${userId}:${role}`);
      server = {
        ...server,
        members: server.members.map((m) => (m.id === userId ? { ...m, role } : m)),
      };
      return server;
    });

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    const save = await screen.findByRole("button", { name: "Save settings" });
    expect(save).toBeDisabled();

    fireEvent.change(screen.getByLabelText("Member email"), {
      target: { value: "new@example.com" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add member" }));
    expect(screen.getByText(/Will be added when you save/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Project Role for new@example.com"), {
      target: { value: "developer" },
    });
    fireEvent.change(screen.getByLabelText("Project Role for Project Member"), {
      target: { value: "scrum_master" },
    });

    // Nothing has been sent yet.
    expect(addProjectMember).not.toHaveBeenCalled();
    expect(assignProjectRole).not.toHaveBeenCalled();
    expect(save).toBeEnabled();

    fireEvent.click(save);

    await waitFor(() => expect(push).toHaveBeenCalledWith("/board"));
    expect(calls).toEqual([
      "add:new@example.com",
      "role:u2:scrum_master",
      "role:u3:developer",
    ]);
    expect(updateProject).not.toHaveBeenCalled();
  });

  it("does not render member management controls for a non-owner", async () => {
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "u2",
      email: "member@example.com",
      roles: ["member"],
    });
    vi.mocked(getProject).mockResolvedValue({
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "member@example.com",
          name: "Project Member",
          role: "developer",
          initials: "PM",
          color: "#64748B",
        },
      ],
    });

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    expect(await screen.findByText("Project Member")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Add member" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Remove" })).not.toBeInTheDocument();
    expect(screen.getByText("Only a Project owner can manage members.")).toBeInTheDocument();
  });

  it("refuses to save while a member has no Project Role and highlights them", async () => {
    vi.mocked(getProject).mockResolvedValue({
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "member@example.com",
          name: "Project Member",
          role: "",
          initials: "PM",
          color: "#64748B",
        },
      ],
    });

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.change(await screen.findByLabelText("Name"), {
      target: { value: "Renamed" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save settings" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      `"No Project Role" isn't allowed. Choose a Project Role for Project Member before saving.`,
    );
    const roleSelect = screen.getByLabelText("Project Role for Project Member");
    expect(roleSelect.closest("li")).toHaveAttribute("data-missing-role", "true");
    expect(updateProject).not.toHaveBeenCalled();

    // Picking a role clears the highlight.
    fireEvent.change(roleSelect, { target: { value: "developer" } });
    expect(roleSelect.closest("li")).not.toHaveAttribute("data-missing-role");
  });

  it.each([
    ["an older backend's 404", "not found (404)"],
    ["the backend's violation", "this user does not exist (400)"],
  ])("says the user doesn't exist when adding an unknown email (%s)", async (_label, message) => {
    vi.mocked(getProject).mockResolvedValue({
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
      ],
    });
    vi.mocked(addProjectMember).mockRejectedValue(new Error(message));

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.change(await screen.findByLabelText("Member email"), {
      target: { value: "ghost@example.com" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add member" }));
    fireEvent.change(screen.getByLabelText("Project Role for ghost@example.com"), {
      target: { value: "developer" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save settings" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(
      "Unable to save: This user doesn't exist. Check the email address and try again.",
    );
    expect(alert).not.toHaveTextContent("This Project was not found");
  });

  it("keeps a removal staged and explains a backend refusal under Save", async () => {
    vi.mocked(getProject).mockResolvedValue({
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "member@example.com",
          name: "Project Member",
          role: "developer",
          initials: "PM",
          color: "#64748B",
        },
      ],
    });
    vi.mocked(removeProjectMember).mockRejectedValue(new Error("Member is locked (409)"));

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.click((await screen.findAllByRole("button", { name: "Remove" }))[1]);
    expect(screen.getByText(/Will be removed when you save/)).toBeInTheDocument();
    expect(removeProjectMember).not.toHaveBeenCalled();

    const save = screen.getByRole("button", { name: "Save settings" });
    fireEvent.click(save);

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Unable to save: Member is locked");
    expect(save.closest("form")).toContainElement(alert);
    expect(screen.getByText("Project Member")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Undo" })).toBeInTheDocument();
  });

  it("removes your own membership last and navigates away after saving", async () => {
    const projectWithTwoOwners = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "second-owner@example.com",
          name: "Second Owner",
          role: "product_owner",
          initials: "SO",
          color: "#64748B",
        },
      ],
    };
    vi.mocked(getProject).mockResolvedValue(projectWithTwoOwners);
    vi.mocked(removeProjectMember).mockResolvedValue({
      ...projectWithTwoOwners,
      members: [projectWithTwoOwners.members[1]],
    });

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.click((await screen.findAllByRole("button", { name: "Remove" }))[0]);
    expect(removeProjectMember).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Save settings" }));

    await waitFor(() => expect(push).toHaveBeenCalledWith("/backlog"));
    expect(removeProjectMember).toHaveBeenCalledWith("p1", "u1");
  });

  it("lets an Auditor self-assign without rendering feedback on Project settings", async () => {
    vi.mocked(getCurrentIdentity).mockResolvedValue({
      authenticated: true,
      user_id: "a1",
      email: "auditor@example.com",
      roles: ["auditor"],
    });
    vi.mocked(selfAssignAuditor).mockResolvedValue({
      project_id: "p1",
      auditor_user_id: "a1",
      auditor_name: "Auditor",
      auditor_email: "auditor@example.com",
      assigned_at: "2026-01-02T03:04:05Z",
    });
    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.click(
      await screen.findByRole("button", { name: "Assign myself as Auditor" }),
    );
    await waitFor(() => expect(selfAssignAuditor).toHaveBeenCalledWith("p1"));

    expect(screen.queryByLabelText("Add feedback")).not.toBeInTheDocument();
    expect(listProjectFeedback).not.toHaveBeenCalled();
    expect(submitProjectFeedback).not.toHaveBeenCalled();
  });

  it("does not load or render Project feedback", async () => {
    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    expect(await screen.findByRole("heading", { name: "Alpha" })).toBeInTheDocument();
    expect(screen.queryByText("Project feedback")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Add feedback")).not.toBeInTheDocument();
    expect(listProjectFeedback).not.toHaveBeenCalled();
  });

  it("allows an owner to save settings and confirm permanent deletion", async () => {
    const owned = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
      ],
    };
    vi.mocked(getProject).mockResolvedValue(owned);
    vi.mocked(updateProject).mockResolvedValue({
      ...owned,
      name: "Beta",
      description: "Updated",
    });
    vi.mocked(deleteProject).mockResolvedValue(undefined);

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.change(await screen.findByLabelText("Name"), {
      target: { value: "Beta" },
    });
    fireEvent.change(screen.getByLabelText("Description"), {
      target: { value: "Updated" },
    });
    // Leaving with unsaved changes asks first; "Keep editing" keeps them.
    fireEvent.click(screen.getByRole("button", { name: "Cancel and back to Board" }));
    expect(
      screen.getByText(/changes to this Project that haven't been saved/),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Keep editing" }));
    expect(screen.getByLabelText("Name")).toHaveValue("Beta");
    expect(push).not.toHaveBeenCalled();

    // "Discard and leave" drops them and goes back to the Board.
    fireEvent.click(screen.getByRole("button", { name: "Cancel and back to Board" }));
    fireEvent.click(screen.getByRole("button", { name: "Discard and leave" }));
    expect(push).toHaveBeenCalledWith("/board");
    expect(screen.getByLabelText("Name")).toHaveValue("Alpha");
    expect(screen.getByLabelText("Description")).toHaveValue(
      "A project description",
    );
    expect(updateProject).not.toHaveBeenCalled();
    push.mockClear();

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Beta" },
    });
    fireEvent.change(screen.getByLabelText("Description"), {
      target: { value: "Updated" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save settings" }));
    await waitFor(() =>
      expect(updateProject).toHaveBeenCalledWith("p1", {
        name: "Beta",
        description: "Updated",
      }),
    );
    expect(push).toHaveBeenCalledWith("/board");

    fireEvent.click(screen.getByRole("button", { name: "Delete Project" }));
    fireEvent.click(
      await screen.findByRole("button", { name: "Delete permanently" }),
    );
    await waitFor(() => expect(deleteProject).toHaveBeenCalledWith("p1"));
    expect(push).toHaveBeenCalledWith("/backlog");
  });

  it("updates the Template summary when the settings dropdown changes", async () => {
    const owned = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
      ],
    };
    const generalTemplate = {
      ...template,
      id: "general",
      name: "General",
      mode: "general",
    };
    vi.mocked(getProject).mockResolvedValue(owned);
    vi.mocked(listTemplates).mockResolvedValue([template, generalTemplate]);

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    const templateSelect = await screen.findByLabelText("Template");
    expect(templateSelect).toHaveValue("se");
    expect(templateSelect).toHaveDisplayValue("Software Engineering");

    fireEvent.change(templateSelect, {
      target: { value: "general" },
    });

    expect(templateSelect).toHaveValue("general");
    expect(templateSelect).toHaveDisplayValue("General");
  });

  it("lets the owner change the template even when members have no Project Role", async () => {
    const owned = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "member@example.com",
          name: "Project Member",
          role: "",
          initials: "PM",
          color: "#64748B",
        },
      ],
    };
    const generalTemplate = {
      ...template,
      id: "general",
      name: "General",
      mode: "general",
    };
    vi.mocked(getProject).mockResolvedValue(owned);
    vi.mocked(listTemplates).mockResolvedValue([template, generalTemplate]);

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.change(await screen.findByLabelText("Template"), {
      target: { value: "general" },
    });

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save settings" })).toBeEnabled();
    // The members list previews the reset: owner keeps the owner role, others are cleared.
    expect(screen.getByLabelText("Project Role for Project Owner")).toHaveValue(
      generalTemplate.creator_role,
    );
    expect(screen.getByLabelText("Project Role for Project Member")).toHaveValue("");
    // Roles stay editable during the preview; nothing is sent until Save.
    expect(screen.getByLabelText("Project Role for Project Member")).toBeEnabled();
  });

  it("applies roles picked during a template change after the template is saved", async () => {
    const owned = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
        {
          id: "u2",
          email: "member@example.com",
          name: "Project Member",
          role: "developer",
          initials: "PM",
          color: "#64748B",
        },
      ],
    };
    const generalTemplate = {
      ...template,
      id: "general",
      name: "General",
      mode: "general",
      roles: ["owner", "member"],
      default_role: "member",
      creator_role: "owner",
    };
    // What the server holds right after the template change: owner kept, others cleared.
    const switched = {
      ...owned,
      template_id: "general",
      mode: "general",
      members: [
        { ...owned.members[0], role: "owner" },
        { ...owned.members[1], role: "" },
      ],
    };
    const calls: string[] = [];
    vi.mocked(getProject).mockResolvedValue(owned);
    vi.mocked(listTemplates).mockResolvedValue([template, generalTemplate]);
    vi.mocked(getTemplate).mockImplementation(async (id) =>
      id === "general" ? generalTemplate : template,
    );
    vi.mocked(updateProject).mockImplementation(async () => {
      calls.push("updateProject");
      return switched;
    });
    vi.mocked(assignProjectRole).mockImplementation(async (_p, userId, role) => {
      calls.push(`assign:${userId}:${role}`);
      return {
        ...switched,
        members: switched.members.map((m) => (m.id === userId ? { ...m, role } : m)),
      };
    });

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.change(await screen.findByLabelText("Template"), {
      target: { value: "general" },
    });
    fireEvent.change(screen.getByLabelText("Project Role for Project Member"), {
      target: { value: "member" },
    });
    expect(screen.getByLabelText("Project Role for Project Member")).toHaveValue("member");
    expect(assignProjectRole).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Save settings" }));

    await waitFor(() => expect(push).toHaveBeenCalledWith("/board"));
    expect(calls).toEqual(["updateProject", "assign:u2:member"]);
  });


  it("shows why a settings save was refused next to the Save button", async () => {
    const owned = {
      ...project,
      members: [
        {
          id: "u1",
          email: "owner@example.com",
          name: "Project Owner",
          role: "product_owner",
          initials: "PO",
          color: "#64748B",
        },
      ],
    };
    const generalTemplate = { ...template, id: "general", name: "General", mode: "general" };
    vi.mocked(getProject).mockResolvedValue(owned);
    vi.mocked(listTemplates).mockResolvedValue([template, generalTemplate]);
    vi.mocked(updateProject).mockRejectedValue(
      new Error("This Project has 2 Sprints. Delete them before changing the template. (400)"),
    );

    render(
      <ProjectProvider>
        <ProjectOverview projectId="p1" />
      </ProjectProvider>,
    );

    fireEvent.change(await screen.findByLabelText("Template"), {
      target: { value: "general" },
    });
    const save = screen.getByRole("button", { name: "Save settings" });
    fireEvent.click(save);

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(
      "This Project has 2 Sprints. Delete them before changing the template.",
    );
    expect(save.closest("form")).toContainElement(alert);
    // The owner's choice is kept so they can retry after removing the Sprints.
    expect(screen.getByLabelText("Template")).toHaveValue("general");
    expect(push).not.toHaveBeenCalled();
  });
});
