"use client";

import { useEffect, useMemo, useState } from "react";
import { getWorkItem } from "@/api/work-items";
import { listMembers, membersFromAuditors, membersOf } from "@/api/members";
import { listAuditorDirectory } from "@/api/projects";
import { listSprints } from "@/api/sprints";
import { AppShell } from "@/components/layout/app-shell";
import { NoProjectSelected } from "@/components/layout/no-project-selected";
import { useProjects } from "@/components/layout/project-provider";
import { WorkItemDetailEditor } from "@/components/work-items/work-item-detail-editor";

export function WorkItemDetailPage({ id }: { id: string }) {
  const { selectedProjectId, selectedProject, loadState: projectsLoadState } = useProjects();
  const [item, setItem] = useState<Awaited<ReturnType<typeof getWorkItem>>>();
  const [otherProjectMembers, setOtherProjectMembers] = useState<Awaited<ReturnType<typeof listMembers>>>([]);
  const [sprints, setSprints] = useState<Awaited<ReturnType<typeof listSprints>>>([]);
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(true);
  const [auditorEntries, setAuditorEntries] = useState<
    Awaited<ReturnType<typeof listAuditorDirectory>>
  >([]);

  useEffect(() => {
    let active = true;
    // Optional: without the list the Auditor field just shows "No Auditor".
    listAuditorDirectory()
      .then((items) => {
        if (active) setAuditorEntries(items);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, []);
  const auditors = useMemo(() => membersFromAuditors(auditorEntries), [auditorEntries]);

  /* eslint-disable react-hooks/set-state-in-effect -- reset detail state when id/Project changes */
  useEffect(() => {
    let active = true;
    setItem(undefined);
    setOtherProjectMembers([]);
    setSprints([]);
    setError(undefined);

    if (!selectedProjectId) {
      setLoading(false);
      return () => {
        active = false;
      };
    }

    setLoading(true);
    getWorkItem(id)
      .then(async (nextItem) => {
        if (!nextItem) throw new Error("Work Item not found");
        const projectId = nextItem.projectId ?? selectedProjectId;
        // The selected Project's members are already loaded by the provider
        // (derived below); only fetch members of some other Project.
        const [nextMembers, nextSprints] = await Promise.all([
          projectId === selectedProjectId ? [] : listMembers(projectId),
          listSprints(projectId),
        ]);
        if (!active) return;
        setItem(nextItem);
        setOtherProjectMembers(nextMembers);
        setSprints(nextSprints);
      })
      .catch((cause) => {
        if (active) {
          setError(cause instanceof Error ? cause.message : "Unable to load Work Item");
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [id, selectedProjectId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const itemProjectId = item?.projectId ?? selectedProjectId;
  const members = useMemo(
    () =>
      selectedProject && itemProjectId === selectedProject.id
        ? membersOf(selectedProject)
        : otherProjectMembers,
    [itemProjectId, otherProjectMembers, selectedProject],
  );

  return (
    <AppShell>
      {projectsLoadState === "loading" || (selectedProjectId && loading) ? (
        <p className="p-8 text-sm text-text-tertiary">Loading Work Item...</p>
      ) : !selectedProjectId ? (
        <NoProjectSelected title="Select a Project to open Work Items" />
      ) : error ? (
        <p className="p-8 text-sm text-red-600">{error}</p>
      ) : !item ? (
        <p className="p-8 text-sm text-text-tertiary">Loading Work Item...</p>
      ) : (
        <WorkItemDetailEditor
          initialItem={item}
          members={members}
          auditors={auditors}
          sprints={sprints}
        />
      )}
    </AppShell>
  );
}
