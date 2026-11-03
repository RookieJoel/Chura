"use client";

import { useEffect, useState } from "react";
import { getWorkItem, listComments } from "@/api/work-items";
import { listMembers } from "@/api/members";
import { listSprints } from "@/api/sprints";
import { WorkItemDetailEditor } from "@/components/work-items/work-item-detail-editor";

export function WorkItemDetailPage({ id }: { id: string }) {
  const [item, setItem] = useState<Awaited<ReturnType<typeof getWorkItem>>>();
  const [comments, setComments] = useState<Awaited<ReturnType<typeof listComments>>>([]);
  const [members, setMembers] = useState<Awaited<ReturnType<typeof listMembers>>>([]);
  const [sprints, setSprints] = useState<Awaited<ReturnType<typeof listSprints>>>([]);
  const [error, setError] = useState<string>();

  useEffect(() => {
    let active = true;
    getWorkItem(id)
      .then(async (nextItem) => {
        if (!nextItem) throw new Error("Work Item not found");
        const [nextComments, nextMembers, nextSprints] = await Promise.all([
          listComments(nextItem.id),
          listMembers(),
          listSprints(),
        ]);
        if (!active) return;
        setItem(nextItem);
        setComments(nextComments);
        setMembers(nextMembers);
        setSprints(nextSprints);
      })
      .catch((cause) => {
        if (active) setError(cause instanceof Error ? cause.message : "Unable to load Work Item");
      });
    return () => {
      active = false;
    };
  }, [id]);

  if (error) return <p className="p-8 text-sm text-red-600">{error}</p>;
  if (!item) return <p className="p-8 text-sm text-text-tertiary">Loading Work Item...</p>;
  return <WorkItemDetailEditor initialItem={item} comments={comments} members={members} sprints={sprints} />;
}
