import { AppShell } from "@/components/layout/app-shell";
import { ProjectOverview } from "@/components/projects/project-overview";

export default async function ProjectPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return (
    <AppShell minimal>
      <ProjectOverview projectId={id} />
    </AppShell>
  );
}
