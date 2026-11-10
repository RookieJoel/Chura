"use client";

import { useEffect, useMemo, useState } from "react";
import {
  exportReviewSummary,
  getReviewSummary,
  listReflections,
  submitReflection,
} from "@/api/analytics";
import { listSprints } from "@/api/sprints";
import { AppShell } from "@/components/layout/app-shell";
import { NoProjectSelected } from "@/components/layout/no-project-selected";
import { useProjects } from "@/components/layout/project-provider";
import { ReflectionForm } from "@/components/reports/reflection-form";
import { ReflectionList } from "@/components/reports/reflection-list";
import { ReviewSummaryPanel } from "@/components/reports/review-summary-panel";
import { reflectionTopicsFromTemplate } from "@/lib/reflection-utils";
import type { SprintReflection, SprintReviewSummary } from "@/types/analytics";
import type { Sprint } from "@/types/work-item";

function pickDefaultSprint(sprints: Sprint[]): string {
  const active = sprints.find((sprint) => sprint.status === "active");
  if (active) return active.id;
  const completed = sprints.find((sprint) => sprint.status === "completed");
  if (completed) return completed.id;
  return sprints[0]?.id ?? "";
}

export default function ReportsPage() {
  const {
    selectedProjectId,
    selectedTemplate,
    currentUser,
    loadState: projectsLoadState,
  } = useProjects();

  const [sprints, setSprints] = useState<Sprint[]>([]);
  const [selectedSprintId, setSelectedSprintId] = useState("");
  const [sprintsLoading, setSprintsLoading] = useState(false);
  const [sprintsError, setSprintsError] = useState<string>();

  const [reflections, setReflections] = useState<SprintReflection[]>([]);
  const [reflectionsLoading, setReflectionsLoading] = useState(false);
  const [reflectionsError, setReflectionsError] = useState<string | null>(null);

  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const [summary, setSummary] = useState<SprintReviewSummary | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(false);
  const [summaryError, setSummaryError] = useState<string | null>(null);
  const [selectedMetrics, setSelectedMetrics] = useState<string[]>([]);
  const [exporting, setExporting] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  const templateTopics = selectedTemplate?.reflection?.topics;
  const topics = useMemo(
    () => reflectionTopicsFromTemplate(templateTopics),
    [templateTopics],
  );
  const availableMetrics = selectedTemplate?.sprint_review?.metrics ?? [];
  const canSubmitReflection = !!currentUser?.roles.includes("member");
  const myReflection =
    reflections.find(
      (reflection) =>
        reflection.authorId &&
        currentUser?.user_id &&
        reflection.authorId === currentUser.user_id,
    ) ?? null;

  /* eslint-disable react-hooks/set-state-in-effect -- project-scoped Sprint list reset before fetch */
  useEffect(() => {
    let active = true;
    setSprints([]);
    setSelectedSprintId("");
    setSprintsError(undefined);
    setReflections([]);
    setReflectionsError(null);
    setSummary(null);
    setSummaryError(null);
    setExportError(null);
    setSelectedMetrics([]);

    if (!selectedProjectId) {
      setSprintsLoading(false);
      return () => {
        active = false;
      };
    }

    setSprintsLoading(true);
    listSprints(selectedProjectId)
      .then((nextSprints) => {
        if (!active) return;
        setSprints(nextSprints);
        setSelectedSprintId(pickDefaultSprint(nextSprints));
      })
      .catch((cause) => {
        if (active) {
          setSprintsError(
            cause instanceof Error ? cause.message : "Unable to load sprints",
          );
        }
      })
      .finally(() => {
        if (active) setSprintsLoading(false);
      });

    return () => {
      active = false;
    };
  }, [selectedProjectId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  /* eslint-disable react-hooks/set-state-in-effect -- reload reflections when Sprint changes */
  useEffect(() => {
    let active = true;
    setReflections([]);
    setReflectionsError(null);
    setSummary(null);
    setSummaryError(null);
    setExportError(null);
    setSubmitError(null);

    if (!selectedSprintId) {
      setReflectionsLoading(false);
      return () => {
        active = false;
      };
    }

    setReflectionsLoading(true);
    listReflections(selectedSprintId)
      .then((items) => {
        if (!active) return;
        setReflections(items);
      })
      .catch((cause) => {
        if (active) {
          setReflectionsError(
            cause instanceof Error
              ? cause.message
              : "Unable to load reflections",
          );
        }
      })
      .finally(() => {
        if (active) setReflectionsLoading(false);
      });

    return () => {
      active = false;
    };
  }, [selectedSprintId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleSubmitReflection = async (answers: Record<string, string>) => {
    if (!selectedSprintId || submitting) return;
    setSubmitting(true);
    setSubmitError(null);
    try {
      const saved = await submitReflection(selectedSprintId, answers);
      setReflections((current) => {
        const withoutMine = current.filter(
          (item) => item.authorId !== saved.authorId,
        );
        return [...withoutMine, saved];
      });
    } catch (cause) {
      setSubmitError(
        cause instanceof Error ? cause.message : "Unable to save reflection",
      );
    } finally {
      setSubmitting(false);
    }
  };

  const handleGenerateSummary = async () => {
    if (!selectedSprintId || summaryLoading) return;
    setSummaryLoading(true);
    setSummaryError(null);
    try {
      const next = await getReviewSummary(
        selectedSprintId,
        selectedMetrics.length > 0 ? selectedMetrics : undefined,
      );
      setSummary(next);
    } catch (cause) {
      setSummary(null);
      setSummaryError(
        cause instanceof Error
          ? cause.message
          : "Unable to generate review summary",
      );
    } finally {
      setSummaryLoading(false);
    }
  };

  const handleExport = async (format: "csv" | "md") => {
    if (!selectedSprintId || exporting) return;
    setExporting(true);
    setExportError(null);
    try {
      await exportReviewSummary(selectedSprintId, {
        format,
        metrics: selectedMetrics.length > 0 ? selectedMetrics : undefined,
      });
    } catch (cause) {
      setExportError(
        cause instanceof Error
          ? cause.message
          : "Unable to export review summary",
      );
    } finally {
      setExporting(false);
    }
  };

  const toggleMetric = (key: string) => {
    setSelectedMetrics((current) =>
      current.includes(key)
        ? current.filter((item) => item !== key)
        : [...current, key],
    );
  };

  return (
    <AppShell>
      {!selectedProjectId ? (
        <NoProjectSelected
          title="No Project selected"
          description="Select a Project to submit reflections and generate Sprint Review reports."
        />
      ) : (
        <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
          <header className="border-b border-border bg-surface px-8 py-5">
            <h1 className="m-0 text-lg font-bold text-text">Reports</h1>
            <p className="m-0 mt-1 text-sm text-text-secondary">
              Sprint reflections and review summaries for the selected Project.
            </p>
          </header>

          <div className="space-y-8 px-8 py-6">
            {projectsLoadState === "loading" || sprintsLoading ? (
              <p role="status" className="m-0 text-sm text-text-tertiary">
                Loading sprints…
              </p>
            ) : null}

            {sprintsError ? (
              <p role="alert" className="m-0 text-sm text-red-600">
                {sprintsError}
              </p>
            ) : null}

            {!sprintsLoading && !sprintsError && sprints.length === 0 ? (
              <p className="m-0 text-sm text-text-tertiary">
                This Project has no Sprints yet. Create one under Sprints to
                capture reflections and review summaries.
              </p>
            ) : null}

            {sprints.length > 0 ? (
              <>
                <label className="block max-w-md">
                  <span className="text-[11px] font-bold uppercase tracking-wide text-text-tertiary">
                    Sprint
                  </span>
                  <select
                    value={selectedSprintId}
                    onChange={(event) => setSelectedSprintId(event.target.value)}
                    className="mt-1.5 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text outline-none focus:border-brand"
                  >
                    {sprints.map((sprint) => (
                      <option key={sprint.id} value={sprint.id}>
                        {sprint.name} ({sprint.status})
                      </option>
                    ))}
                  </select>
                </label>

                <section
                  aria-labelledby="my-reflection-heading"
                  className="rounded-lg border border-border bg-surface p-5"
                >
                  <h2
                    id="my-reflection-heading"
                    className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary"
                  >
                    My reflection
                  </h2>
                  {canSubmitReflection ? (
                    <>
                      <p className="m-0 mt-1 mb-4 text-sm text-text-secondary">
                        Answer the retrospective questions for this Sprint.
                        Submitting again replaces your previous answers.
                      </p>
                      <ReflectionForm
                        topics={topics}
                        existing={myReflection}
                        submitting={submitting}
                        error={submitError}
                        onSubmit={handleSubmitReflection}
                      />
                    </>
                  ) : (
                    <p className="m-0 mt-2 text-sm text-text-tertiary">
                      Only Team Members can submit reflections. You can still
                      read the team list and generate a review summary below.
                    </p>
                  )}
                </section>

                <section
                  aria-labelledby="team-reflections-heading"
                  className="rounded-lg border border-border bg-surface p-5"
                >
                  <h2
                    id="team-reflections-heading"
                    className="m-0 mb-4 text-xs font-bold uppercase tracking-wide text-text-tertiary"
                  >
                    Team reflections
                  </h2>
                  <ReflectionList
                    reflections={reflections}
                    topics={topics}
                    loading={reflectionsLoading}
                    error={reflectionsError}
                  />
                </section>

                <div className="rounded-lg border border-border bg-surface p-5">
                  <ReviewSummaryPanel
                    availableMetrics={availableMetrics}
                    selectedMetrics={selectedMetrics}
                    onToggleMetric={toggleMetric}
                    summary={summary}
                    loading={summaryLoading}
                    error={summaryError}
                    exporting={exporting}
                    exportError={exportError}
                    onGenerate={handleGenerateSummary}
                    onExport={handleExport}
                  />
                </div>
              </>
            ) : null}
          </div>
        </div>
      )}
    </AppShell>
  );
}
