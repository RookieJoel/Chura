"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import { createProject } from "@/api/projects";
import { getTemplate, listTemplates, type ProjectTemplate } from "@/api/templates";
import { useProjects } from "@/components/layout/project-provider";

export function ProjectCreationForm() {
  const router = useRouter();
  const { registerCreatedProject } = useProjects();
  const [templates, setTemplates] = useState<ProjectTemplate[]>([]);
  const [templatesLoading, setTemplatesLoading] = useState(true);
  const [templatesError, setTemplatesError] = useState<string | null>(null);
  const [selectedTemplateId, setSelectedTemplateId] = useState("");
  const [selectedTemplate, setSelectedTemplate] =
    useState<ProjectTemplate | null>(null);
  const [templateLoading, setTemplateLoading] = useState(false);
  const [templateError, setTemplateError] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const templateRequest = useRef(0);

  const loadTemplates = useCallback(async () => {
    setTemplatesLoading(true);
    setTemplatesError(null);
    try {
      setTemplates(await listTemplates());
    } catch (reason) {
      setTemplatesError(
        reason instanceof Error ? reason.message : "Unable to load templates",
      );
    } finally {
      setTemplatesLoading(false);
    }
  }, []);

  useEffect(() => {
    let active = true;
    listTemplates()
      .then((result) => {
        if (active) setTemplates(result);
      })
      .catch((reason: unknown) => {
        if (active) {
          setTemplatesError(
            reason instanceof Error ? reason.message : "Unable to load templates",
          );
        }
      })
      .finally(() => {
        if (active) setTemplatesLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const handleTemplateChange = async (templateId: string) => {
    const requestId = templateRequest.current + 1;
    templateRequest.current = requestId;
    setSelectedTemplateId(templateId);
    setSelectedTemplate(null);
    setTemplateError(null);
    setSubmitError(null);
    if (!templateId) return;

    setTemplateLoading(true);
    try {
      const template = await getTemplate(templateId);
      if (templateRequest.current === requestId) setSelectedTemplate(template);
    } catch (reason) {
      if (templateRequest.current === requestId) {
        setTemplateError(
          reason instanceof Error
            ? reason.message
            : "Unable to load the selected template",
        );
      }
    } finally {
      if (templateRequest.current === requestId) setTemplateLoading(false);
    }
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const projectName = name.trim();
    const projectDescription = description.trim();
    if (!projectName || !selectedTemplate) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const project = await createProject({
        name: projectName,
        description: projectDescription,
        template_id: selectedTemplate.id,
      });
      registerCreatedProject(project);
      router.push(`/projects/${encodeURIComponent(project.id)}`);
    } catch (reason) {
      setSubmitError(
        reason instanceof Error ? reason.message : "Unable to create Project",
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="mx-auto w-full max-w-2xl p-6">
      <Link
        href="/backlog"
        className="text-sm font-semibold text-brand hover:underline"
      >
        Back to Projects
      </Link>
      <h1 className="mb-2 mt-5 text-2xl font-bold text-text">Create Project</h1>
      <p className="mb-6 text-sm text-text-secondary">
        Choose a template to set up this Project. You will be its initial owner.
      </p>

      <form
        className="space-y-5 rounded-lg border border-border bg-surface p-5"
        onSubmit={handleSubmit}
      >
        <div className="space-y-1.5">
          <label htmlFor="project-name" className="text-sm font-semibold text-text">
            Project name
          </label>
          <input
            id="project-name"
            name="name"
            required
            maxLength={100}
            value={name}
            onChange={(event) => setName(event.target.value)}
            className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-text outline-none focus:border-brand"
          />
        </div>

        <div className="space-y-1.5">
          <label
            htmlFor="project-description"
            className="text-sm font-semibold text-text"
          >
            Description <span className="font-normal text-text-tertiary">(optional)</span>
          </label>
          <textarea
            id="project-description"
            name="description"
            maxLength={2000}
            rows={4}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            className="w-full resize-y rounded-md border border-border bg-background px-3 py-2 text-sm text-text outline-none focus:border-brand"
          />
        </div>

        <div className="space-y-1.5">
          <label
            htmlFor="project-template"
            className="text-sm font-semibold text-text"
          >
            Project template
          </label>
          {templatesLoading ? (
            <p role="status" className="text-sm text-text-secondary">
              Loading templates…
            </p>
          ) : templatesError ? (
            <div role="alert" className="space-y-2 text-sm text-red-600">
              <p>{templatesError}</p>
              <button
                type="button"
                onClick={() => void loadTemplates()}
                className="font-semibold underline"
              >
                Retry loading templates
              </button>
            </div>
          ) : templates.length === 0 ? (
            <p role="status" className="text-sm text-text-secondary">
              No templates are currently available. You cannot create a Project
              yet.
            </p>
          ) : (
            <select
              id="project-template"
              name="template_id"
              required
              value={selectedTemplateId}
              onChange={(event) => void handleTemplateChange(event.target.value)}
              className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-text outline-none focus:border-brand"
            >
              <option value="">Choose a template</option>
              {templates.map((template) => (
                <option key={template.id} value={template.id}>
                  {template.name}
                </option>
              ))}
            </select>
          )}
          {templateLoading && (
            <p role="status" className="text-sm text-text-secondary">
              Loading selected template…
            </p>
          )}
          {templateError && (
            <div role="alert" className="space-y-2 text-sm text-red-600">
              <p>{templateError}</p>
              <button
                type="button"
                onClick={() => void handleTemplateChange(selectedTemplateId)}
                className="font-semibold underline"
              >
                Retry selected template
              </button>
            </div>
          )}
          {selectedTemplate && (
            <p className="text-sm text-text-secondary">
              Template mode: {selectedTemplate.mode}. Your Project role will be{" "}
              <span className="font-semibold">{selectedTemplate.creator_role}</span>.
            </p>
          )}
        </div>

        {submitError && (
          <p role="alert" className="text-sm text-red-600">
            {submitError}
          </p>
        )}

        <div className="flex items-center justify-end gap-3 border-t border-border pt-4">
          <Link
            href="/backlog"
            className="rounded-md px-3 py-2 text-sm font-semibold text-text-secondary hover:bg-background"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={
              submitting ||
              templatesLoading ||
              templates.length === 0 ||
              !name.trim() ||
              !selectedTemplate ||
              templateLoading
            }
            className="rounded-md bg-brand px-4 py-2 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50"
          >
            {submitting ? "Creating…" : "Create Project"}
          </button>
        </div>
      </form>
    </section>
  );
}
