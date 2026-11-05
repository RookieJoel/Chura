import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { listProjects } from "@/api/projects";
import { AppShell } from "@/components/layout/app-shell";
import { ProjectProvider } from "@/components/layout/project-provider";
import { TopNav } from "@/components/layout/top-nav";

vi.mock("@/api/projects", () => ({ listProjects: vi.fn() }));
vi.mock("next/navigation", () => ({
  usePathname: () => "/backlog",
  useRouter: (() => {
    const router = { push: vi.fn(), refresh: vi.fn(), replace: vi.fn() };
    return () => router;
  })(),
}));

const projects = [
  { id: "p1", name: "Alpha", description: "", template_id: "se", mode: "se" },
  { id: "p2", name: "Beta", description: "", template_id: "se", mode: "se" },
];

function renderTopNav() {
  return render(
    <ProjectProvider>
      <TopNav />
    </ProjectProvider>,
  );
}

describe("TopNav project selector", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(listProjects).mockReset();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            authenticated: true,
            user_id: "u1",
            email: "user@example.com",
            roles: ["member"],
          }),
        ),
      ),
    );
  });

  it("loads accessible Projects, replaces an inaccessible stored ID, and persists selection", async () => {
    localStorage.setItem("chura.selectedProjectId", "stale");
    vi.mocked(listProjects).mockResolvedValue(projects);
    renderTopNav();

    const selector = await screen.findByRole("combobox", { name: "Project" });
    await waitFor(() => expect(selector).toHaveValue("p1"));
    expect(localStorage.getItem("chura.selectedProjectId")).toBe("p1");

    fireEvent.change(selector, { target: { value: "p2" } });
    expect(selector).toHaveValue("p2");
    expect(localStorage.getItem("chura.selectedProjectId")).toBe("p2");
  });

  it("uses the project provider in the actual application shell", async () => {
    vi.mocked(listProjects).mockResolvedValue(projects);
    render(<AppShell><main>Page content</main></AppShell>);

    expect(await screen.findByRole("combobox", { name: "Project" })).toHaveValue("p1");
    expect(screen.getByText("Alpha")).toBeInTheDocument();
    expect(screen.getByText("Page content")).toBeInTheDocument();
  });

  it("shows loading and empty Project states", async () => {
    vi.mocked(listProjects).mockResolvedValue([]);
    renderTopNav();
    expect(screen.getByRole("status", { name: "Loading Projects" })).toBeInTheDocument();
    expect(await screen.findByText("No accessible Projects")).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Project" })).toBeDisabled();
    expect(screen.getByRole("link", { name: "Create Project" })).toHaveAttribute(
      "href",
      "/projects/new",
    );
    expect(localStorage.getItem("chura.selectedProjectId")).toBeNull();
  });

  it("keeps the Project overview shell free of Sprint and Work Item controls", async () => {
    vi.mocked(listProjects).mockResolvedValue(projects);
    render(<AppShell minimal><main>Project overview content</main></AppShell>);

    expect(await screen.findByText("Project overview content")).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "Project" })).toBeInTheDocument();
    expect(screen.queryByText("Search work items")).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Sprints" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Board" })).not.toBeInTheDocument();
  });

  it("shows discovery errors and allows retry", async () => {
    vi.mocked(listProjects)
      .mockRejectedValueOnce(new Error("service unavailable"))
      .mockResolvedValueOnce(projects);
    renderTopNav();
    expect(await screen.findByRole("alert")).toHaveTextContent("service unavailable");
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await waitFor(() =>
      expect(screen.getByRole("combobox", { name: "Project" })).toHaveValue("p1"),
    );
    expect(listProjects).toHaveBeenCalledTimes(2);
  });
});
