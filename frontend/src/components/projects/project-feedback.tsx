"use client";

import Link from "next/link";
import { useEffect, useState, type FormEvent } from "react";
import {
  listProjectAuditors,
  listProjectFeedback,
  submitProjectFeedback,
  type ProjectAuditorAssignment,
  type ProjectFeedback,
} from "@/api/projects";
import { useProjects } from "@/components/layout/project-provider";
import { userFacingError } from "@/lib/user-facing-error";

type LoadState = "loading" | "ready" | "error";

export function ProjectFeedbackPage({ projectId }: { projectId: string }) {
  const {
    currentUser,
    projects,
    selectProject,
    selectedProjectId,
  } = useProjects();
  const [feedbackResult, setFeedbackResult] = useState<{
    projectId: string;
    items?: ProjectFeedback[];
    error?: string;
  } | null>(null);
  const [auditorResult, setAuditorResult] = useState<{
    projectId: string;
    items?: ProjectAuditorAssignment[];
    error?: string;
  } | null>(null);
  const [feedbackBody, setFeedbackBody] = useState("");
  const [pending, setPending] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const isAuditor = !!currentUser?.roles.includes("auditor");

  useEffect(() => {
    if (
      selectedProjectId !== projectId &&
      projects.some((project) => project.id === projectId)
    ) {
      selectProject(projectId);
    }
  }, [projectId, projects, selectProject, selectedProjectId]);

  useEffect(() => {
    let active = true;
    listProjectFeedback(projectId)
      .then((items) => {
        if (active) setFeedbackResult({ projectId, items });
      })
      .catch((reason: unknown) => {
        if (active) {
          setFeedbackResult({
            projectId,
            error: userFacingError(reason, "Unable to load Project feedback."),
          });
        }
      });
    return () => {
      active = false;
    };
  }, [projectId]);

  useEffect(() => {
    let active = true;
    listProjectAuditors(projectId)
      .then((items) => {
        if (active) setAuditorResult({ projectId, items });
      })
      .catch((reason: unknown) => {
        if (active) {
          setAuditorResult({
            projectId,
            error: userFacingError(
              reason,
              "Unable to load Auditor assignments.",
            ),
          });
        }
      });
    return () => {
      active = false;
    };
  }, [projectId]);

  const currentFeedback =
    feedbackResult?.projectId === projectId ? feedbackResult : null;
  const feedback = currentFeedback?.items ?? [];
  const feedbackState: LoadState = !currentFeedback
    ? "loading"
    : currentFeedback.error
      ? "error"
      : "ready";
  const currentAuditors =
    auditorResult?.projectId === projectId ? auditorResult : null;
  const auditorsState: LoadState = !currentAuditors
    ? "loading"
    : currentAuditors.error
      ? "error"
      : "ready";
  const assignedAuditor =
    !!currentUser &&
    (currentAuditors?.items ?? []).some(
      (assignment) => assignment.auditor_user_id === currentUser.user_id,
    );

  async function submitFeedback(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!feedbackBody.trim()) return;
    setPending(true);
    setSubmitError(null);
    try {
      const created = await submitProjectFeedback(projectId, feedbackBody);
      setFeedbackResult((current) => {
        if (current?.projectId !== projectId) {
          return { projectId, items: [created] };
        }
        return { projectId, items: [created, ...(current.items ?? [])] };
      });
      setFeedbackBody("");
    } catch (reason) {
      setSubmitError(
        userFacingError(reason, "Unable to submit Project feedback."),
      );
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="flex-1 overflow-y-auto p-6">
      <div className="mx-auto w-full max-w-3xl">
        <Link
          href={`/projects/${encodeURIComponent(projectId)}`}
          className="text-sm font-semibold text-brand-dark hover:underline"
        >
          Project settings
        </Link>
        <h1 className="mt-4 text-2xl font-bold text-text">Project feedback</h1>

        {feedbackState === "loading" && (
          <p role="status" className="mt-4 text-sm text-text-tertiary">
            Loading feedback…
          </p>
        )}
        {feedbackState === "error" && (
          <p role="alert" className="mt-4 text-sm text-red-600">
            {currentFeedback?.error}
          </p>
        )}
        {feedbackState === "ready" && feedback.length === 0 && (
          <p className="mt-4 text-sm text-text-tertiary">
            No Project feedback yet.
          </p>
        )}
        {feedbackState === "ready" && feedback.length > 0 && (
          <ul className="mt-4 space-y-3">
            {feedback.map((entry) => (
              <li
                key={entry.id}
                className="rounded-lg border border-border bg-surface p-4"
              >
                <p className="text-sm font-semibold text-text">
                  {entry.author_name || entry.author_email}
                </p>
                <p className="text-xs text-text-tertiary">
                  {new Date(entry.created_at).toLocaleString()}
                </p>
                <p className="mt-2 whitespace-pre-wrap text-sm text-text-secondary">
                  {entry.body}
                </p>
              </li>
            ))}
          </ul>
        )}

        {!isAuditor && (
          <p className="mt-4 text-xs text-text-tertiary">
            Project members can read feedback. Only assigned Auditors can
            submit it.
          </p>
        )}
        {isAuditor && auditorsState === "loading" && (
          <p role="status" className="mt-4 text-sm text-text-tertiary">
            Loading Auditor assignments…
          </p>
        )}
        {isAuditor && auditorsState === "error" && (
          <p role="alert" className="mt-4 text-sm text-red-600">
            {currentAuditors?.error}
          </p>
        )}
        {isAuditor && auditorsState === "ready" && !assignedAuditor && (
          <p className="mt-4 text-xs text-text-tertiary">
            Only Auditors assigned to this Project can submit feedback.
          </p>
        )}
        {isAuditor && assignedAuditor && (
          <form onSubmit={submitFeedback} className="mt-4 space-y-2">
            <label
              htmlFor="project-feedback-body"
              className="block text-xs font-semibold uppercase text-text-tertiary"
            >
              Add feedback
            </label>
            <textarea
              id="project-feedback-body"
              required
              maxLength={4000}
              value={feedbackBody}
              onChange={(event) => setFeedbackBody(event.target.value)}
              rows={4}
              className="w-full rounded-md border border-border bg-surface px-3 py-2 text-sm text-text"
            />
            {submitError && (
              <p role="alert" className="text-sm text-red-600">
                {submitError}
              </p>
            )}
            <button
              type="submit"
              disabled={pending || !feedbackBody.trim()}
              className="fl-cta px-3 py-2 text-sm"
            >
              Submit feedback
            </button>
            {pending && (
              <p role="status" className="sr-only">
                Submitting feedback…
              </p>
            )}
          </form>
        )}
      </div>
    </section>
  );
}
