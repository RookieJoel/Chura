import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getWorkItem, listComments } from "@/api/work-items";
import { listMembers } from "@/api/members";
import { WorkItemDetailEditor } from "@/components/work-items/work-item-detail-editor";

export async function generateMetadata(
  props: PageProps<"/work-items/[id]">,
): Promise<Metadata> {
  const { id } = await props.params;
  const item = await getWorkItem(id);
  return { title: item ? `${item.id} · Chura` : "Work Item · Chura" };
}

export default async function WorkItemDetailPage(
  props: PageProps<"/work-items/[id]">,
) {
  const { id } = await props.params;
  const item = await getWorkItem(id);
  if (!item) notFound();

  const [comments, members] = await Promise.all([
    listComments(item.id),
    listMembers(),
  ]);

  return (
    <WorkItemDetailEditor
      initialItem={item}
      comments={comments}
      members={members}
    />
  );
}
