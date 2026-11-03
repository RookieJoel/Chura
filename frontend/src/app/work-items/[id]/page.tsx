import { WorkItemDetailPage } from "@/components/work-items/work-item-detail-page";

export default async function WorkItemDetailRoute(
  props: PageProps<"/work-items/[id]">,
) {
  const { id } = await props.params;
  return <WorkItemDetailPage id={id} />;
}
