"use client";

import { useEffect, useState } from "react";
import { createComment, listComments } from "@/api/work-items";
import type { WorkItemComment } from "@/types/work-item";

type LoadState = "loading" | "ready" | "error";

const MAX_COMMENT_BODY = 4000;

export function WorkItemComments({ workItemId }: { workItemId: string }) {
  const [comments, setComments] = useState<WorkItemComment[]>([]);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [loadError, setLoadError] = useState<string | null>(null);
  const [body, setBody] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  /* eslint-disable react-hooks/set-state-in-effect -- reload comments when Work Item changes */
  useEffect(() => {
    let active = true;
    setLoadState("loading");
    setLoadError(null);
    setSubmitError(null);
    setBody("");
    setComments([]);

    listComments(workItemId)
      .then((items) => {
        if (!active) return;
        setComments(items);
        setLoadState("ready");
      })
      .catch((cause) => {
        if (!active) return;
        setLoadError(
          cause instanceof Error ? cause.message : "Unable to load comments",
        );
        setLoadState("error");
      });

    return () => {
      active = false;
    };
  }, [workItemId]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    const trimmed = body.trim();
    if (!trimmed || submitting) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const created = await createComment(workItemId, trimmed);
      setComments((current) => [created, ...current]);
      setBody("");
    } catch (cause) {
      const message =
        cause instanceof Error ? cause.message : "Unable to post comment";
      const looksForbidden = /403|forbidden/i.test(message);
      setSubmitError(
        looksForbidden
          ? `${message}. Project members can comment; Auditors must self-assign in Settings first.`
          : message,
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section aria-labelledby="work-item-comments-heading">
      <div className="mb-2.5 flex items-center justify-between gap-2">
        <h2
          id="work-item-comments-heading"
          className="m-0 text-xs font-bold uppercase tracking-wide text-text-tertiary"
        >
          Comments
        </h2>
        <span className="text-[11px] font-semibold text-text-tertiary">
          Newest first
        </span>
      </div>

      {loadState === "loading" && (
        <p role="status" className="m-0 text-sm text-text-tertiary">
          Loading comments…
        </p>
      )}
      {loadState === "error" && (
        <p role="alert" className="m-0 text-sm text-red-600">
          {loadError}
        </p>
      )}
      {loadState === "ready" && comments.length === 0 && (
        <p className="m-0 text-sm text-text-tertiary">No comments yet.</p>
      )}
      {loadState === "ready" && comments.length > 0 && (
        <ul className="m-0 flex list-none flex-col gap-3 p-0">
          {comments.map((comment) => (
            <li
              key={comment.id}
              className="rounded-lg border border-border bg-surface p-3"
            >
              <p className="m-0 text-sm font-semibold text-text">
                {comment.authorName || comment.authorEmail}
              </p>
              <p className="m-0 text-xs text-text-tertiary">
                {new Date(comment.createdAt).toLocaleString()}
              </p>
              <p className="mt-2 mb-0 whitespace-pre-wrap text-sm text-text-secondary">
                {comment.body}
              </p>
            </li>
          ))}
        </ul>
      )}

      <form onSubmit={submit} className="mt-4 flex flex-col gap-2">
        <label
          htmlFor="work-item-comment-body"
          className="text-xs font-semibold uppercase tracking-wide text-text-tertiary"
        >
          Add comment
        </label>
        <textarea
          id="work-item-comment-body"
          required
          maxLength={MAX_COMMENT_BODY}
          value={body}
          onChange={(event) => setBody(event.target.value)}
          rows={3}
          className="w-full resize-y rounded-lg border border-border bg-background px-3 py-2 text-[13px] text-text outline-none focus:border-brand"
          disabled={submitting || loadState === "loading"}
        />
        {submitError && (
          <p role="alert" className="m-0 text-sm text-red-600">
            {submitError}
          </p>
        )}
        <button
          type="submit"
          disabled={submitting || !body.trim() || loadState === "loading"}
          className="fl-cta self-start px-3 py-1.5 text-xs"
        >
          {submitting ? "Posting…" : "Post comment"}
        </button>
      </form>
    </section>
  );
}
