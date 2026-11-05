import { AppShell } from "@/components/layout/app-shell";
import { ProjectCreationForm } from "@/components/projects/project-creation-form";

export default function NewProjectPage() {
  return (
    <AppShell minimal>
      <ProjectCreationForm />
    </AppShell>
  );
}
