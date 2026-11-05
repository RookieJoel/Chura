"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getProject, type ProjectSummary } from "@/api/projects";
import { useProjects } from "@/components/layout/project-provider";

export function ProjectOverview({ projectId }: { projectId: string }) {
  const { projects, loadState, selectProject } = useProjects();
  const [result, setResult] = useState<{
    projectId: string;
    project?: ProjectSummary;
    error?: string;
  } | null>(null);

  useEffect(() => {
    let active = true;
    getProject(projectId)
      .then((result) => {
        if (active) setResult({ projectId, project: result });
      })
      .catch((reason: unknown) => {
        if (!active) return;
        setResult({
          projectId,
          error:
            reason instanceof Error ? reason.message : "Unable to load Project",
        });
      });
    return () => {
      active = false;
    };
  }, [projectId]);

  const currentResult = result?.projectId === projectId ? result : null;
  const project = currentResult?.project ?? null;
  const error = currentResult?.error ?? null;
  const loading = !currentResult;

  useEffect(() => {
    if (project && projects.some((item) => item.id === project.id)) {
      selectProject(project.id);
    }
  }, [project, projects, selectProject]);

  if (loading || loadState === "loading") {
    return (
      <section className="flex-1 p-6">
        <p role="status">Loading Project…</p>
      </section>
    );
  }

  if (error || !project) {
    return (
      <section className="flex-1 p-6">
        <div role="alert" className="text-sm text-red-600">
          {error ?? "Project not found or you do not have access to it."}
        </div>
        <Link
          href="/backlog"
          className="mt-4 inline-block text-sm font-semibold text-brand hover:underline"
        >
          Return to Projects
        </Link>
      </section>
    );
  }

  return (
    <section className="flex-1 p-6">
      <div className="mx-auto w-full max-w-3xl">
        <p className="text-xs font-bold uppercase tracking-wide text-text-tertiary">
          Project overview
        </p>
        <h1 className="mt-2 text-2xl font-bold text-text">{project.name}</h1>
        {project.description ? (
          <p className="mt-3 whitespace-pre-wrap text-sm text-text-secondary">
            {project.description}
          </p>
        ) : (
          <p className="mt-3 text-sm text-text-tertiary">
            No description has been added.
          </p>
        )}
        <dl className="mt-6 grid gap-4 rounded-lg border border-border bg-surface p-5 sm:grid-cols-2">
          <div>
            <dt className="text-xs font-semibold uppercase text-text-tertiary">
              Template
            </dt>
            <dd className="mt-1 text-sm font-semibold text-text">
              {project.template_id}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase text-text-tertiary">
              Mode
            </dt>
            <dd className="mt-1 text-sm font-semibold text-text">
              {project.mode}
            </dd>
          </div>
        </dl>
      </div>
    </section>
  );
}
