import { AppShell } from "@/components/layout/app-shell";
import { ProjectFeedbackPage } from "@/components/projects/project-feedback";

export default async function ProjectFeedbackRoute({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return (
    <AppShell>
      <ProjectFeedbackPage projectId={id} />
    </AppShell>
  );
}
