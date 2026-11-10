import Link from "next/link";

export function NoProjectSelected({
  title = "No Project selected",
  description = "Select a Project in the top navigation, or create one to continue.",
}: {
  title?: string;
  description?: string;
}) {
  return (
    <div className="fl-wash flex flex-1 flex-col items-center justify-center gap-3 px-8 py-16 text-center">
      <span className="fl-accent-bar fl-accent-bar-sm mb-1" aria-hidden />
      <h1 className="m-0 text-lg font-semibold tracking-tight text-text">
        {title}
      </h1>
      <p className="m-0 max-w-md text-sm text-text-tertiary">{description}</p>
      <Link href="/projects/new" className="fl-cta mt-2 px-4 py-2.5 text-xs">
        Create Project
      </Link>
    </div>
  );
}
