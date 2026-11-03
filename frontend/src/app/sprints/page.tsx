"use client";

import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { createSprint, deleteSprint, listSprints, updateSprint } from "@/api/sprints";
import { AppShell } from "@/components/layout/app-shell";
import { SprintActions } from "@/components/board/sprint-actions";
import type { Sprint } from "@/types/work-item";

function formatDate(value: string) {
  if (!value) return "Not set";
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export default function SprintsPage() {
  const [sprints, setSprints] = useState<Sprint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState("");
  const [team, setTeam] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [editingId, setEditingId] = useState<string>();
  const [savingId, setSavingId] = useState<string>();
  const [deletingId, setDeletingId] = useState<string>();

  useEffect(() => {
    let active = true;
    listSprints()
      .then((items) => {
        if (active) setSprints(items);
      })
      .catch((cause) => {
        if (active) {
          setError(cause instanceof Error ? cause.message : "Unable to load sprints");
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const handleCreate = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedName = name.trim();
    const trimmedTeam = team.trim();
    if (!trimmedName) {
      setError("Sprint name is required");
      return;
    }
    if (!trimmedTeam) {
      setError("Team is required");
      return;
    }
    if (startDate && endDate && startDate > endDate) {
      setError("End date must be on or after start date");
      return;
    }
    setCreating(true);
    setError(undefined);
    try {
      const sprint = await createSprint({
        name: trimmedName,
        team: trimmedTeam,
        startDate,
        endDate,
      });
      setSprints((current) => [...current, sprint]);
      setName("");
      setTeam("");
      setStartDate("");
      setEndDate("");
      setShowCreateForm(false);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to create sprint");
    } finally {
      setCreating(false);
    }
  };

  const handleUpdate = async (
    event: FormEvent<HTMLFormElement>,
    sprint: Sprint,
  ) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setSavingId(sprint.id);
    setError(undefined);
    try {
      const updated = await updateSprint(sprint, {
        name: String(form.get("name") ?? ""),
        team: String(form.get("team") ?? ""),
        startDate: String(form.get("startDate") ?? ""),
        endDate: String(form.get("endDate") ?? ""),
      });
      setSprints((current) =>
        current.map((item) => (item.id === updated.id ? updated : item)),
      );
      setEditingId(undefined);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to update sprint");
    } finally {
      setSavingId(undefined);
    }
  };

  const handleDelete = async (sprint: Sprint) => {
    if (!window.confirm(`Delete "${sprint.name}"? This cannot be undone.`)) return;
    setDeletingId(sprint.id);
    setError(undefined);
    try {
      await deleteSprint(sprint.id);
      setSprints((current) => current.filter((item) => item.id !== sprint.id));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to delete sprint");
    } finally {
      setDeletingId(undefined);
    }
  };

  return (
    <AppShell>
      <div className="flex items-center justify-between border-b border-border bg-surface px-8 py-5">
        <div>
          <p className="m-0 text-[11px] font-bold uppercase tracking-wider text-text-tertiary">Planning</p>
          <h1 className="m-0 mt-1 text-2xl font-bold tracking-tight">Sprints</h1>
        </div>
        <button
          type="button"
          onClick={() => {
            setShowCreateForm((current) => !current);
            setError(undefined);
          }}
          className="rounded-md bg-brand px-4 py-2.5 text-[13px] font-bold text-white shadow-sm hover:bg-brand-dark"
        >
          {showCreateForm ? "Cancel" : "Add sprint"}
        </button>
      </div>
      <div className="flex-1 overflow-y-auto px-8 py-6">
        {error && (
          <p className="mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        )}
        {loading ? (
          <p className="text-sm text-text-tertiary">Loading sprints...</p>
        ) : sprints.length === 0 ? (
          <p className="text-sm text-text-tertiary">No sprints found.</p>
        ) : (
          <div className="flex flex-col gap-3">
            {sprints.map((sprint) => (
              <div key={sprint.id} className="rounded-lg border border-border bg-surface shadow-[0_1px_2px_rgba(9,30,66,0.08)]">
                {editingId === sprint.id ? (
                  <form onSubmit={(event) => void handleUpdate(event, sprint)} className="p-5">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                        Name *
                        <input name="name" required defaultValue={sprint.name} className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand" />
                      </label>
                      <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                        Team *
                        <input name="team" required defaultValue={sprint.team} className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand" />
                      </label>
                      <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                        Start date
                        <input name="startDate" type="date" defaultValue={sprint.startDate} className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand" />
                      </label>
                      <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                        End date
                        <input name="endDate" type="date" defaultValue={sprint.endDate} className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand" />
                      </label>
                    </div>
                    <div className="mt-4 flex justify-end gap-2">
                      <button type="button" onClick={() => setEditingId(undefined)} className="rounded-lg border border-border px-3 py-2 text-xs font-bold text-text-secondary">Cancel</button>
                      <button type="submit" disabled={savingId === sprint.id} className="rounded-lg bg-brand px-3 py-2 text-xs font-bold text-white disabled:opacity-50">
                        {savingId === sprint.id ? "Saving..." : "Save changes"}
                      </button>
                    </div>
                  </form>
                ) : (
                  <div className="flex items-center justify-between px-5 py-4">
                    <div>
                      <div className="flex items-center gap-2">
                        <h2 className="m-0 text-sm font-bold text-text">{sprint.name}</h2>
                        <span className="rounded-full bg-brand-tint px-2 py-1 text-[11px] font-bold capitalize text-brand-dark">
                          {sprint.status}
                        </span>
                      </div>
                      <p className="m-0 mt-1 text-xs text-text-tertiary">
                        {formatDate(sprint.startDate)} – {formatDate(sprint.endDate)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {sprint.status !== "completed" && (
                        <SprintActions
                          sprint={sprint}
                          onUpdated={(updated) =>
                            setSprints((current) =>
                              current.map((item) =>
                                item.id === updated.id ? updated : item,
                              ),
                            )
                          }
                        />
                      )}
                      <button type="button" onClick={() => setEditingId(sprint.id)} className="rounded-lg border border-border px-3 py-2 text-xs font-bold text-text-secondary">Edit</button>
                      <button type="button" disabled={deletingId === sprint.id} onClick={() => void handleDelete(sprint)} className="rounded-lg border border-red-200 px-3 py-2 text-xs font-bold text-red-600 disabled:opacity-50">
                        {deletingId === sprint.id ? "Deleting..." : "Delete"}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
        {showCreateForm && (
          <form
            onSubmit={handleCreate}
            className="mb-5 rounded-lg border border-border bg-surface p-5 shadow-sm"
          >
            <h2 className="m-0 mb-4 text-sm font-bold">Create sprint</h2>
            <div className="grid gap-3 sm:grid-cols-2">
              <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                Name *
                <input
                  required
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Sprint 5"
                  className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
                />
              </label>
              <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                Team *
                <input
                  required
                  value={team}
                  onChange={(event) => setTeam(event.target.value)}
                  placeholder="Web team"
                  className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
                />
              </label>
              <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                Start date
                <input
                  type="date"
                  value={startDate}
                  onChange={(event) => setStartDate(event.target.value)}
                  className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
                />
              </label>
              <label className="flex flex-col gap-1 text-xs font-semibold text-text-secondary">
                End date
                <input
                  type="date"
                  value={endDate}
                  onChange={(event) => setEndDate(event.target.value)}
                  className="rounded-lg border border-border bg-background px-3 py-2 text-sm font-normal text-text outline-none focus:border-brand"
                />
              </label>
            </div>
            <div className="mt-4 flex justify-end">
              <button
                type="submit"
                disabled={creating}
                className="rounded-lg bg-brand px-4 py-2.5 text-[13px] font-bold text-white disabled:opacity-50"
              >
                {creating ? "Creating..." : "Create sprint"}
              </button>
            </div>
          </form>
        )}
      </div>
    </AppShell>
  );
}
