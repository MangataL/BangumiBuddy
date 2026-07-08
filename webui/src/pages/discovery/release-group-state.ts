import type {
  BangumiCandidate,
  ReleaseGroupCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";

export function orderReleaseGroups(
  groups: ReleaseGroupCandidate[]
): ReleaseGroupCandidate[] {
  return [
    ...groups.filter((group) => group.subscribed),
    ...groups.filter((group) => !group.subscribed),
  ];
}

export function getSubscribedReleaseGroupCount(
  summary: ReleaseGroupSummary | undefined
): number {
  return (
    summary?.releaseGroups.filter((group) => group.subscribed).length ?? 0
  );
}

export function hasLoadedNoReleaseGroups(
  item: BangumiCandidate,
  summary: ReleaseGroupSummary | undefined
): boolean {
  return (
    item.releaseGroupsKnownEmpty === true ||
    summary?.releaseGroups.length === 0
  );
}
