import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createProject, getProject, listProjects } from "@/api/projects";
import { getTemplate, listTemplates } from "@/api/templates";
import { ProjectProvider } from "@/components/layout/project-provider";
import { TopNav } from "@/components/layout/top-nav";
import { ProjectCreationForm } from "@/components/projects/project-creation-form";

const push = vi.fn();

vi.mock("@/api/projects", () => ({
  createProject: vi.fn(),
  getProject: vi.fn(),
  listProjects: vi.fn(),
}));
vi.mock("@/api/templates", () => ({
  getTemplate: vi.fn(),
  listTemplates: vi.fn(),
}));
vi.mock("next/navigation", () => ({
  usePathname: () => "/projects/new",
  useRouter: () => ({ push }),
}));

const template = {
  id: "se",
  name: "Software Engineering",
  mode: "se",
  work_item_types: ["user_story", "task", "bug"],
  roles: ["product_owner", "scrum_master", "developer"],
  default_role: "developer",
  creator_role: "product_owner",
};
const createdProject = {
  id: "project-1",
  name: "New Project",
  description: "A project description",
  template_id: "se",
  mode: "se",
};
const createdProjectDetails = {
  ...createdProject,
  created_by: "u1",
  members: [],
};

function renderForm() {
  return render(
    <ProjectProvider>
      <TopNav minimal />
      <ProjectCreationForm />
    </ProjectProvider>,
  );
}

describe("ProjectCreationForm", () => {
  beforeEach(() => {
    localStorage.clear();
    push.mockReset();
    vi.mocked(listProjects).mockResolvedValue([]);
    vi.mocked(getProject).mockResolvedValue(createdProjectDetails);
    vi.mocked(listTemplates).mockResolvedValue([template]);
    vi.mocked(getTemplate).mockResolvedValue(template);
    vi.mocked(createProject).mockResolvedValue(createdProject);
  });

  it("requires explicit template selection, creates the Project, and navigates to its overview", async () => {
    renderForm();
    const name = await screen.findByLabelText("Project name");
    fireEvent.change(name, { target: { value: "New Project" } });
    const createButton = screen.getByRole("button", { name: "Create Project" });
    expect(createButton).toBeDisabled();

    fireEvent.change(await screen.findByLabelText("Project template"), {
      target: { value: "se" },
    });
    expect(await screen.findByText(/product_owner/)).toBeInTheDocument();
    expect(createButton).toBeEnabled();

    fireEvent.change(screen.getByLabelText(/Description/), {
      target: { value: "A project description" },
    });
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(createProject).toHaveBeenCalledWith({
        name: "New Project",
        description: "A project description",
        template_id: "se",
      });
      expect(push).toHaveBeenCalledWith("/projects/project-1");
    });
    expect(localStorage.getItem("chura.selectedProjectId")).toBe("project-1");
  });

  it("shows an empty-template state and prevents creation", async () => {
    vi.mocked(listTemplates).mockResolvedValue([]);
    renderForm();

    expect(
      await screen.findByText(/No templates are currently available/),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create Project" })).toBeDisabled();
  });

  it("surfaces template and project creation errors", async () => {
    vi.mocked(listTemplates)
      .mockRejectedValueOnce(new Error("Template service unavailable"))
      .mockResolvedValueOnce([template]);
    renderForm();
    expect(
      await screen.findByText("Template service unavailable"),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry loading templates" }));

    fireEvent.change(await screen.findByLabelText("Project name"), {
      target: { value: "New Project" },
    });
    fireEvent.change(await screen.findByLabelText("Project template"), {
      target: { value: "se" },
    });
    expect(await screen.findByText(/product_owner/)).toBeInTheDocument();
    vi.mocked(createProject).mockRejectedValue(new Error("name is required"));
    fireEvent.click(screen.getByRole("button", { name: "Create Project" }));
    expect(await screen.findByText("name is required")).toBeInTheDocument();
    expect(push).not.toHaveBeenCalled();
  });

  it("shows authorization failures without navigating away", async () => {
    vi.mocked(createProject).mockRejectedValue(
      new Error("Team Member role required (403)"),
    );
    renderForm();
    fireEvent.change(await screen.findByLabelText("Project name"), {
      target: { value: "New Project" },
    });
    fireEvent.change(await screen.findByLabelText("Project template"), {
      target: { value: "se" },
    });
    await screen.findByText(/product_owner/);
    fireEvent.click(screen.getByRole("button", { name: "Create Project" }));

    expect(
      await screen.findByText("Team Member role required (403)"),
    ).toBeInTheDocument();
    expect(push).not.toHaveBeenCalled();
  });

  it("allows retrying a selected template after a detail request fails", async () => {
    vi.mocked(getTemplate)
      .mockRejectedValueOnce(new Error("Template detail unavailable"))
      .mockResolvedValueOnce(template);
    renderForm();
    await screen.findByLabelText("Project name");
    fireEvent.change(await screen.findByLabelText("Project template"), {
      target: { value: "se" },
    });

    expect(await screen.findByText("Template detail unavailable")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry selected template" }));

    expect(await screen.findByText(/product_owner/)).toBeInTheDocument();
  });

  it("keeps the newly created selection if initial project discovery finishes late", async () => {
    let resolveProjects!: (
      projects: Awaited<ReturnType<typeof listProjects>>,
    ) => void;
    vi.mocked(listProjects).mockReturnValue(
      new Promise((resolve) => {
        resolveProjects = resolve;
      }),
    );
    renderForm();
    fireEvent.change(await screen.findByLabelText("Project name"), {
      target: { value: "New Project" },
    });
    fireEvent.change(await screen.findByLabelText("Project template"), {
      target: { value: "se" },
    });
    await screen.findByText(/product_owner/);
    fireEvent.click(screen.getByRole("button", { name: "Create Project" }));
    await waitFor(() =>
      expect(push).toHaveBeenCalledWith("/projects/project-1"),
    );

    resolveProjects([]);

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: "Project" })).toHaveValue(
        "project-1",
      );
    });
    expect(localStorage.getItem("chura.selectedProjectId")).toBe("project-1");
  });
});
