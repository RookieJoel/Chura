"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";
import {
  addProjectMember,
  assignProjectRole,
  deleteProject,
  getProject,
  listProjectAuditors,
  removeProjectMember,
  selfAssignAuditor,
  updateProject,
  type ProjectAuditorAssignment,
  type ProjectDetails,
  type ProjectMember,
} from "@/api/projects";
import { getTemplate, listTemplates, type ProjectTemplate } from "@/api/templates";
import { useProjects } from "@/components/layout/project-provider";
import { Modal } from "@/components/modal";
import { WorkItemAvatar } from "@/components/work-items/avatar";
import { userFacingError } from "@/lib/user-facing-error";

interface ProjectOverviewResult {
  projectId: string;
  project?: ProjectDetails;
  template?: ProjectTemplate;
  error?: string;
  templateError?: string;
}

type LoadState = "loading" | "ready" | "error";

const newMemberKey = (email: string) => `new:${email.toLowerCase()}`;

/** One row of the members list as it will look once Save settings is pressed. */
interface DraftMemberRow {
  key: string;
  /** Existing member; absent for a member still to be added. */
  member?: ProjectMember;
  email: string;
  label: string;
  role: string;
  removed: boolean;
}

export function ProjectOverview({ projectId }: { projectId: string }) {
  const router = useRouter();
  const {
    projects,
    loadState,
    selectProject,
    selectedProjectId,
    currentUser,
    identityError,
    updateSelectedProject,
    removeDeletedProject,
  } = useProjects();
  const [result, setResult] = useState<ProjectOverviewResult | null>(null);
  const [email, setEmail] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  // Shown beside Save settings, where the owner is looking when a save fails.
  const [settingsError, setSettingsError] = useState<string | null>(null);
  // Unsaved member changes. Roles are keyed by user id, or by
  // newMemberKey(email) for members not added yet.
  const [pendingRoles, setPendingRoles] = useState<Record<string, string>>({});
  const [pendingAdds, setPendingAdds] = useState<string[]>([]);
  const [pendingRemovals, setPendingRemovals] = useState<string[]>([]);
  // Rows highlighted after a save was refused for a missing role.
  const [missingRoleKeys, setMissingRoleKeys] = useState<string[]>([]);
  const [pendingAction, setPendingAction] = useState<string | null>(null);
  const [auditorResult, setAuditorResult] = useState<{
    projectId: string;
    items?: ProjectAuditorAssignment[];
    error?: string;
  } | null>(null);
  const [settingsName, setSettingsName] = useState("");
  const [settingsDescription, setSettingsDescription] = useState("");
  const [settingsTemplateId, setSettingsTemplateId] = useState("");
  const [availableTemplates, setAvailableTemplates] = useState<ProjectTemplate[]>(
    [],
  );
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [confirmLeave, setConfirmLeave] = useState(false);

  const isAuditor = !!currentUser?.roles.includes("auditor");

  useEffect(() => {
    let active = true;
    getProject(projectId)
      .then((details) => {
        if (!active) return;
        setResult({ projectId, project: details });
        setSettingsName(details.name);
        setSettingsDescription(details.description);
        setSettingsTemplateId(details.template_id);
        return getTemplate(details.template_id)
          .then((selectedTemplate) => {
            if (active) {
              setResult((current) =>
                current?.projectId === projectId
                  ? { ...current, template: selectedTemplate }
                  : current,
              );
            }
          })
          .catch((reason: unknown) => {
            if (active) {
              setResult((current) =>
                current?.projectId === projectId
                  ? {
                      ...current,
                      templateError: userFacingError(
                        reason,
                        "Unable to load the Project template.",
                      ),
                    }
                  : current,
              );
            }
          });
      })
      .catch((reason: unknown) => {
        if (!active) return;
        setResult({
          projectId,
          error: userFacingError(reason, "Unable to load this Project."),
        });
      });
    return () => {
      active = false;
    };
  }, [projectId]);

  useEffect(() => {
    let active = true;
    listProjectAuditors(projectId)
      .then((items) => {
        if (!active) return;
        setAuditorResult({ projectId, items });
      })
      .catch((reason: unknown) => {
        if (!active) return;
        setAuditorResult({
          projectId,
          error: userFacingError(
            reason,
            "Unable to load Auditor assignments.",
          ),
        });
      });
    return () => {
      active = false;
    };
  }, [projectId]);

  useEffect(() => {
    let active = true;
    listTemplates()
      .then((templates) => {
        if (active) setAvailableTemplates(templates);
      })
      .catch(() => {
        if (active) setAvailableTemplates([]);
      });
    return () => {
      active = false;
    };
  }, []);

  const currentResult = result?.projectId === projectId ? result : null;
  const project = currentResult?.project ?? null;
  const template = currentResult?.template ?? null;
  const error = currentResult?.error ?? null;
  const templateError = currentResult?.templateError ?? null;
  const loading = !currentResult;
  const currentAuditors =
    auditorResult?.projectId === projectId ? auditorResult : null;
  const auditors = currentAuditors?.items ?? [];
  const auditorsState: LoadState = !currentAuditors
    ? "loading"
    : currentAuditors.error
      ? "error"
      : "ready";
  const auditorsError = currentAuditors?.error ?? null;
  const isOwner =
    !!currentUser &&
    !!template &&
    project?.members.some(
      (member) =>
        member.id === currentUser.user_id && member.role === template.creator_role,
    );
  const canManage =
    isOwner && selectedProjectId === projectId && loadState === "ready";
  const assignedAuditor =
    !!currentUser &&
    auditors.some(
      (assignment) => assignment.auditor_user_id === currentUser.user_id,
    );

  useEffect(() => {
    if (project && projects.some((item) => item.id === project.id)) {
      selectProject(project.id);
    }
  }, [project, projects, selectProject]);
  // Member changes are drafts until Save settings is pressed.
  function submitAddMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!project) return;
    const value = email.trim();
    if (!value) return;
    const lower = value.toLowerCase();
    const existing = project.members.find(
      (member) => member.email.toLowerCase() === lower,
    );
    if (existing && pendingRemovals.includes(existing.id)) {
      // Adding back someone marked for removal just undoes the removal.
      setPendingRemovals((current) => current.filter((id) => id !== existing.id));
    } else if (
      existing ||
      pendingAdds.some((added) => added.toLowerCase() === lower)
    ) {
      setActionError("That person is already a member of this Project.");
      return;
    } else {
      setPendingAdds((current) => [...current, value]);
    }
    setEmail("");
    setActionError(null);
    setSettingsError(null);
  }

  function setDraftRole(key: string, role: string) {
    setPendingRoles((current) => ({ ...current, [key]: role }));
    if (role) setMissingRoleKeys((current) => current.filter((k) => k !== key));
  }

  function toggleRemoval(userId: string) {
    setPendingRemovals((current) =>
      current.includes(userId)
        ? current.filter((id) => id !== userId)
        : [...current, userId],
    );
    setMissingRoleKeys((current) => current.filter((k) => k !== userId));
  }

  function cancelAdd(addedEmail: string) {
    setPendingAdds((current) => current.filter((item) => item !== addedEmail));
    const key = newMemberKey(addedEmail);
    setPendingRoles((current) => {
      const next = { ...current };
      delete next[key];
      return next;
    });
    setMissingRoleKeys((current) => current.filter((k) => k !== key));
  }

  async function assignSelf() {
    setPendingAction("assign");
    setActionError(null);
    try {
      const assignment = await selfAssignAuditor(projectId);
      setAuditorResult((current) => {
        if (current?.projectId !== projectId) {
          return { projectId, items: [assignment] };
        }
        const items = current.items ?? [];
        if (
          items.some(
            (item) => item.auditor_user_id === assignment.auditor_user_id,
          )
        ) {
          return { projectId, items };
        }
        return { projectId, items: [...items, assignment] };
      });
    } catch (reason) {
      setActionError(
        userFacingError(reason, "Unable to assign yourself as Auditor."),
      );
    } finally {
      setPendingAction(null);
    }
  }

  /**
   * The settings form and members list as they will be once saved: the
   * template preview (owners move to the new owner role, everyone else
   * starts without a role), role picks, additions and removals.
   */
  function draftFor(current: ProjectDetails) {
    const displayedTemplate =
      template?.id === current.template_id ? template : null;
    const selectedTemplatePreview =
      availableTemplates.find((item) => item.id === settingsTemplateId) ??
      (template?.id === (settingsTemplateId || current.template_id)
        ? template
        : null);
    const templateChanging =
      !!settingsTemplateId && settingsTemplateId !== current.template_id;
    const pendingTemplate = templateChanging ? selectedTemplatePreview : null;
    const baseRole = (member: ProjectMember) => {
      if (!templateChanging) return member.role;
      const isOwner = displayedTemplate
        ? member.role === displayedTemplate.creator_role
        : member.id === current.created_by;
      return isOwner ? (pendingTemplate?.creator_role ?? "") : "";
    };
    const rows: DraftMemberRow[] = [
      ...current.members.map((member) => ({
        key: member.id,
        member,
        email: member.email,
        label: member.name || member.email || member.id,
        role: member.id in pendingRoles ? pendingRoles[member.id] : baseRole(member),
        removed: pendingRemovals.includes(member.id),
      })),
      ...pendingAdds.map((added) => ({
        key: newMemberKey(added),
        email: added,
        label: added,
        role: pendingRoles[newMemberKey(added)] ?? "",
        removed: false,
      })),
    ];
    const rolesChanged = rows.some(
      (row) => row.member && !row.removed && row.role !== row.member.role,
    );
    const dirty =
      settingsName !== current.name ||
      settingsDescription !== current.description ||
      templateChanging ||
      pendingAdds.length > 0 ||
      pendingRemovals.length > 0 ||
      rolesChanged;
    return {
      displayedTemplate,
      selectedTemplatePreview,
      templateChanging,
      roleOptionsTemplate: pendingTemplate ?? displayedTemplate,
      rows,
      dirty,
    };
  }

  function clearMemberDrafts() {
    setPendingRoles({});
    setPendingAdds([]);
    setPendingRemovals([]);
    setMissingRoleKeys([]);
  }

  /** Shows the server's state after a save, keeping only what is still unsaved. */
  function showSavedProject(saved: ProjectDetails, desired: DraftMemberRow[]) {
    setResult((current) =>
      current?.projectId === projectId ? { ...current, project: saved } : current,
    );
    updateSelectedProject(saved);
    const byEmail = new Map(saved.members.map((m) => [m.email.toLowerCase(), m]));
    const byId = new Map(saved.members.map((m) => [m.id, m]));
    const nextRoles: Record<string, string> = {};
    const nextAdds: string[] = [];
    for (const row of desired) {
      if (row.removed) continue;
      const member = row.member
        ? byId.get(row.member.id)
        : byEmail.get(row.email.toLowerCase());
      if (!row.member && !member) {
        nextAdds.push(row.email);
        if (row.role) nextRoles[newMemberKey(row.email)] = row.role;
      } else if (member && member.role !== row.role) {
        nextRoles[member.id] = row.role;
      }
    }
    setPendingRoles(nextRoles);
    setPendingAdds(nextAdds);
    setPendingRemovals(
      desired
        .filter((row) => row.removed && row.member && byId.has(row.member.id))
        .map((row) => row.key),
    );
  }

  async function saveSettings(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!project) return;
    const draft = draftFor(project);
    setActionError(null);

    // Every member who stays must end up with a Project Role.
    const missing = draft.rows.filter((row) => !row.removed && !row.role);
    if (missing.length > 0) {
      setMissingRoleKeys(missing.map((row) => row.key));
      setSettingsError(
        `"No Project Role" isn't allowed. Choose a Project Role for ${missing
          .map((row) => row.label)
          .join(", ")} before saving.`,
      );
      return;
    }
    setMissingRoleKeys([]);
    setSettingsError(null);
    setPendingAction("settings");

    const selfId = currentUser?.user_id;
    const selfLast = <T extends { key: string }>(rows: T[]) =>
      [...rows].sort((a, b) => Number(a.key === selfId) - Number(b.key === selfId));
    let latest = project;
    let saved = false;
    let step = "the Project settings";
    try {
      // 1. Name, description and template.
      if (
        settingsName !== project.name ||
        settingsDescription !== project.description ||
        draft.templateChanging
      ) {
        latest = await updateProject(project.id, {
          name: settingsName,
          description: settingsDescription,
          ...(draft.templateChanging ? { template_id: settingsTemplateId } : {}),
        });
        saved = true;
      }
      // 2. Removals of other members (removing yourself happens last of all).
      for (const row of draft.rows) {
        if (!row.removed || !row.member || row.key === selfId) continue;
        step = `removing ${row.label}`;
        latest = await removeProjectMember(project.id, row.member.id);
        saved = true;
      }
      // 3. Additions.
      for (const row of draft.rows) {
        if (row.member) continue;
        step = `adding ${row.label}`;
        latest = await addProjectMember(project.id, row.email);
        saved = true;
      }
      // 4. Roles, compared with what the server now holds (your own last).
      const byEmail = new Map(latest.members.map((m) => [m.email.toLowerCase(), m]));
      const roleChanges = selfLast(
        draft.rows.flatMap((row) => {
          const member = row.removed ? undefined : byEmail.get(row.email.toLowerCase());
          return member && member.role !== row.role
            ? [{ key: member.id, member, label: row.label, role: row.role }]
            : [];
        }),
      );
      for (const change of roleChanges) {
        step = `setting the Project Role for ${change.label}`;
        latest = await assignProjectRole(project.id, change.member.id, change.role);
        saved = true;
      }
      // 5. Removing yourself: you can no longer manage the Project afterwards.
      if (selfId && draft.rows.some((row) => row.removed && row.key === selfId)) {
        step = "removing you from the Project";
        await removeProjectMember(project.id, selfId);
        clearMemberDrafts();
        setPendingAction(null);
        router.push("/backlog");
        return;
      }
    } catch (reason) {
      if (saved) showSavedProject(latest, draft.rows);
      const why = userFacingError(reason, "please try again.");
      setSettingsError(
        saved
          ? `Some changes were saved, but ${step} failed: ${why} Press Save settings to retry the rest.`
          : `Unable to save: ${why}`,
      );
      setPendingAction(null);
      return;
    }

    // Everything saved.
    showSavedProject(latest, draft.rows);
    clearMemberDrafts();
    setSettingsName(latest.name);
    setSettingsDescription(latest.description);
    setSettingsTemplateId(latest.template_id);
    try {
      if (latest.template_id !== template?.id) {
        try {
          const nextTemplate = await getTemplate(latest.template_id);
          setResult((current) =>
            current?.projectId === projectId
              ? { ...current, template: nextTemplate, templateError: undefined }
              : current,
          );
        } catch (reason) {
          // Do not keep evaluating roles against the old template.
          setResult((current) =>
            current?.projectId === projectId
              ? {
                  ...current,
                  template: undefined,
                  templateError: userFacingError(
                    reason,
                    "Project settings were saved, but the new template could not be loaded. Reload the page.",
                  ),
                }
              : current,
          );
          return;
        }
      }
      router.push("/board");
    } finally {
      setPendingAction(null);
    }
  }

  async function confirmPermanentDelete() {
    if (!project) return;
    setPendingAction("delete");
    setActionError(null);
    try {
      await deleteProject(project.id);
      removeDeletedProject(project.id);
      setConfirmDelete(false);
      router.push("/backlog");
    } catch (reason) {
      setActionError(
        userFacingError(reason, "Unable to delete this Project."),
      );
    } finally {
      setPendingAction(null);
    }
  }

  if (loading || loadState === "loading") {
    return (
      <section className="flex-1 overflow-y-auto p-6">
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

  // Nothing below is saved until Save settings is pressed.
  const {
    selectedTemplatePreview,
    roleOptionsTemplate,
    rows: draftRows,
    dirty: settingsDirty,
  } = draftFor(project);
  const displayedTemplateName =
    selectedTemplatePreview?.name ??
    (settingsTemplateId || project.template_id);
  const memberCount = draftRows.filter((row) => !row.removed).length;

  function discardSettingsChanges() {
    if (!project) return;
    setSettingsName(project.name);
    setSettingsDescription(project.description);
    setSettingsTemplateId(project.template_id);
    clearMemberDrafts();
    setActionError(null);
    setSettingsError(null);
  }

  function leaveSettings() {
    discardSettingsChanges();
    router.push("/board");
  }

  function cancelAndLeave() {
    if (settingsDirty) {
      setConfirmLeave(true);
    } else {
      leaveSettings();
    }
  }

  return (
    <section className="flex-1 p-6">
      <div className="mx-auto w-full max-w-3xl">
        {/* Owners leave through "Cancel and back to Board" beside Save. */}
        {!canManage && (
          <button
            type="button"
            onClick={leaveSettings}
            className="mb-4 text-sm font-semibold text-brand-dark hover:underline"
          >
            Back to Board
          </button>
        )}
        <p className="text-xs font-bold uppercase tracking-wide text-text-tertiary">
          Project settings
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
        <div className="mt-6 rounded-sm border border-border bg-surface p-5">
          {canManage ? (
            <>
              <label
                htmlFor="project-settings-template"
                className="block text-xs font-semibold uppercase text-text-tertiary"
              >
                Template
              </label>
              <select
                id="project-settings-template"
                value={settingsTemplateId}
                disabled={pendingAction !== null}
                onChange={(event) => {
                  const nextTemplateId = event.target.value;
                  setSettingsTemplateId(nextTemplateId);
                  // Picks were made against the previous preview's roles.
                  setPendingRoles({});
                  setMissingRoleKeys([]);
                  setActionError(null);
                  setSettingsError(null);
                }}
                className="mt-1 w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-semibold text-text"
              >
                {availableTemplates.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </select>
            </>
          ) : (
            <>
              <p className="text-xs font-semibold uppercase text-text-tertiary">
                Template
              </p>
              <p className="mt-1 text-sm font-semibold text-text">
                {displayedTemplateName}
              </p>
              <p className="mt-1 text-xs text-text-tertiary">
                Defines roles and work item types for this Project.
              </p>
            </>
          )}
        </div>
        {templateError && (
          <p role="alert" className="mt-3 text-sm text-red-600">
            Project Template data unavailable: {templateError}
          </p>
        )}
        {actionError && (
          <p role="alert" className="mt-3 text-sm text-red-600">
            {actionError}
          </p>
        )}

        {isAuditor && (
          <section aria-labelledby="project-auditors-heading" className="mt-8">
            <h2
              id="project-auditors-heading"
              className="text-lg font-bold text-text"
            >
              Auditor assignment
            </h2>
            {auditorsState === "loading" && (
              <p role="status" className="mt-3 text-sm text-text-tertiary">
                Loading Auditor assignments…
              </p>
            )}
            {auditorsState === "error" && (
              <p role="alert" className="mt-3 text-sm text-red-600">
                {auditorsError}
              </p>
            )}
            {auditorsState === "ready" && (
              <>
                {assignedAuditor ? (
                  <p className="mt-3 text-sm text-text-secondary">
                    You are assigned to this Project. Assignment cannot be
                    removed until the Project is permanently deleted.
                  </p>
                ) : (
                  <button
                    type="button"
                    disabled={pendingAction !== null}
                    onClick={() => void assignSelf()}
                    className="fl-cta mt-3 px-3 py-2 text-sm"
                  >
                    Assign myself as Auditor
                  </button>
                )}
                {auditors.length === 0 ? (
                  <p className="mt-3 text-sm text-text-tertiary">
                    No Auditors are assigned yet.
                  </p>
                ) : (
                  <ul className="mt-3 list-disc space-y-1 pl-5 text-sm text-text-secondary">
                    {auditors.map((assignment) => (
                      <li key={assignment.auditor_user_id}>
                        {assignment.auditor_name || assignment.auditor_email}{" "}
                        <span className="text-text-tertiary">
                          ({new Date(assignment.assigned_at).toLocaleString()})
                        </span>
                      </li>
                    ))}
                  </ul>
                )}
              </>
            )}
          </section>
        )}

        <section aria-labelledby="project-members-heading" className="mt-8">
          <div className="flex items-center justify-between">
            <h2
              id="project-members-heading"
              className="text-lg font-bold text-text"
            >
              Project members
            </h2>
            <span className="text-xs text-text-tertiary">
              {memberCount} {memberCount === 1 ? "member" : "members"}
            </span>
          </div>
          {canManage && (
            <form
              onSubmit={submitAddMember}
              className="mt-4 flex flex-wrap gap-2"
            >
              <label htmlFor="project-member-email" className="sr-only">
                Member email
              </label>
              <input
                id="project-member-email"
                type="email"
                required
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="Add member by email"
                className="min-w-0 flex-1 rounded-md border border-border bg-surface px-3 py-2 text-sm text-text"
              />
              <button
                type="submit"
                disabled={pendingAction !== null}
                className="fl-cta px-3 py-2 text-sm"
              >
                Add member
              </button>
            </form>
          )}
          {identityError && (
            <p role="alert" className="mt-3 text-sm text-red-600">
              Unable to verify your identity. Member management is unavailable.
            </p>
          )}
          {draftRows.length === 0 ? (
            <p className="mt-4 rounded-lg border border-border bg-surface p-4 text-sm text-text-tertiary">
              No Project members.
            </p>
          ) : (
            <ul className="mt-4 divide-y divide-border rounded-lg border border-border bg-surface">
              {draftRows.map((row) => {
                const missingRole = missingRoleKeys.includes(row.key);
                return (
                  <li
                    key={row.key}
                    data-missing-role={missingRole || undefined}
                    className={`flex flex-wrap items-center gap-3 p-4 ${
                      missingRole
                        ? "bg-red-50 ring-2 ring-inset ring-red-400"
                        : row.removed
                          ? "opacity-60"
                          : ""
                    }`}
                  >
                    <WorkItemAvatar
                      member={
                        row.member ?? {
                          id: row.key,
                          name: row.label,
                          initials: row.label.slice(0, 2).toUpperCase(),
                          color: "#64748B",
                        }
                      }
                      size="md"
                    />
                    <div className="min-w-0 flex-1">
                      <p
                        className={`truncate text-sm font-semibold text-text ${
                          row.removed ? "line-through" : ""
                        }`}
                      >
                        {row.label}
                      </p>
                      <p className="truncate text-xs text-text-tertiary">
                        {row.email}
                        {!row.member && " · Will be added when you save"}
                        {row.removed && " · Will be removed when you save"}
                      </p>
                    </div>
                    {canManage ? (
                      <>
                        {!row.removed && (
                          <>
                            <label
                              htmlFor={`project-role-${row.key}`}
                              className="sr-only"
                            >
                              Project Role for {row.label}
                            </label>
                            <select
                              id={`project-role-${row.key}`}
                              aria-label={`Project Role for ${row.label}`}
                              aria-invalid={missingRole || undefined}
                              value={row.role}
                              disabled={pendingAction !== null}
                              onChange={(event) =>
                                setDraftRole(row.key, event.target.value)
                              }
                              className={`rounded-md border bg-surface px-2 py-1.5 text-xs text-text ${
                                missingRole
                                  ? "border-red-500 ring-2 ring-red-300"
                                  : "border-border"
                              }`}
                            >
                              <option value="">No Project Role</option>
                              {(roleOptionsTemplate?.roles ?? []).map((role) => (
                                <option key={role} value={role}>
                                  {role}
                                </option>
                              ))}
                            </select>
                          </>
                        )}
                        <button
                          type="button"
                          disabled={pendingAction !== null}
                          onClick={() =>
                            row.member
                              ? toggleRemoval(row.member.id)
                              : cancelAdd(row.email)
                          }
                          className="rounded-md px-2 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50 disabled:opacity-50"
                        >
                          {row.removed ? "Undo" : "Remove"}
                        </button>
                      </>
                    ) : (
                      <span className="rounded-full bg-background px-2.5 py-1 text-xs text-text-secondary">
                        {row.role || "No Project Role"}
                      </span>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
          {!isOwner && !identityError && currentUser && (
            <p className="mt-3 text-xs text-text-tertiary">
              Only a Project owner can manage members.
            </p>
          )}
        </section>

        {canManage && (
          <section aria-labelledby="project-settings-heading" className="mt-8">
            <h2
              id="project-settings-heading"
              className="text-lg font-bold text-text"
            >
              Project settings
            </h2>
            <form onSubmit={saveSettings} className="mt-4 space-y-3">
              <div>
                <label
                  htmlFor="project-settings-name"
                  className="block text-xs font-semibold uppercase text-text-tertiary"
                >
                  Name
                </label>
                <input
                  id="project-settings-name"
                  required
                  maxLength={100}
                  value={settingsName}
                  onChange={(event) => setSettingsName(event.target.value)}
                  className="mt-1 w-full rounded-md border border-border bg-surface px-3 py-2 text-sm text-text"
                />
              </div>
              <div>
                <label
                  htmlFor="project-settings-description"
                  className="block text-xs font-semibold uppercase text-text-tertiary"
                >
                  Description
                </label>
                <textarea
                  id="project-settings-description"
                  maxLength={2000}
                  value={settingsDescription}
                  onChange={(event) =>
                    setSettingsDescription(event.target.value)
                  }
                  rows={3}
                  className="mt-1 w-full rounded-md border border-border bg-surface px-3 py-2 text-sm text-text"
                />
              </div>
              <div className="flex flex-wrap items-center gap-3 pt-1">
                <button
                  type="button"
                  disabled={pendingAction !== null}
                  onClick={cancelAndLeave}
                  className="border border-border px-3 py-2 text-sm font-semibold text-text-secondary hover:border-accent hover:text-accent disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Cancel and back to Board
                </button>
                <button
                  type="submit"
                  disabled={pendingAction !== null || !settingsDirty}
                  className="fl-cta px-3 py-2 text-sm"
                >
                  Save settings
                </button>
              </div>
              {settingsError && (
                <p role="alert" className="text-sm text-red-600">
                  {settingsError}
                </p>
              )}
            </form>
            <div className="mt-6 border-t border-border pt-4">
              <h3 className="text-sm font-bold text-text">Delete Project</h3>
              <p className="mt-1 text-sm text-text-secondary">
                Permanently deletes this Project and its memberships, Auditor
                assignments, feedback, Sprints, and Work Items. This cannot be
                undone.
              </p>
              <button
                type="button"
                disabled={pendingAction !== null}
                onClick={() => setConfirmDelete(true)}
                className="mt-3 rounded-md border border-red-200 px-3 py-2 text-sm font-semibold text-red-600 hover:bg-red-50 disabled:opacity-50"
              >
                Delete Project
              </button>
            </div>
          </section>
        )}
      </div>

      {confirmLeave && (
        <Modal title="Unsaved changes" onClose={() => setConfirmLeave(false)}>
          <div className="space-y-4 px-5 py-4">
            <p className="text-sm text-text-secondary">
              You have changes to this Project that haven&apos;t been saved.
              Leave and discard them?
            </p>
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setConfirmLeave(false)}
                className="rounded-md border border-border px-3 py-2 text-sm font-semibold text-text-secondary"
              >
                Keep editing
              </button>
              <button
                type="button"
                onClick={() => {
                  setConfirmLeave(false);
                  leaveSettings();
                }}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white"
              >
                Discard and leave
              </button>
            </div>
          </div>
        </Modal>
      )}
      {confirmDelete && (
        <Modal title="Delete Project" onClose={() => setConfirmDelete(false)}>
          <div className="space-y-4 px-5 py-4">
            <p className="text-sm text-text-secondary">
              Permanently delete <strong>{project.name}</strong>? This cannot be
              undone. All Project-owned data will be removed.
            </p>
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setConfirmDelete(false)}
                className="rounded-md border border-border px-3 py-2 text-sm font-semibold text-text-secondary"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={pendingAction !== null}
                onClick={() => void confirmPermanentDelete()}
                className="rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white disabled:opacity-50"
              >
                Delete permanently
              </button>
            </div>
          </div>
        </Modal>
      )}
    </section>
  );
}
