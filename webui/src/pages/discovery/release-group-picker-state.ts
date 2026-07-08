import type {
  BangumiCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";

import {
  getSubscribedReleaseGroupCount,
  hasLoadedNoReleaseGroups,
} from "./release-group-state";

export interface ReleaseGroupPickerState {
  disabled: boolean;
  label: string;
  subscribed: boolean;
}

export function getReleaseGroupPickerState(
  item: BangumiCandidate,
  summary: ReleaseGroupSummary | undefined,
  loading: boolean
): ReleaseGroupPickerState {
  const noReleaseGroups = hasLoadedNoReleaseGroups(item, summary);
  if (noReleaseGroups) {
    return { disabled: true, label: "暂无字幕组", subscribed: false };
  }
  if (loading && !summary) {
    return { disabled: false, label: "正在载入", subscribed: false };
  }
  const releaseGroupCount = summary?.releaseGroups.length ?? 0;
  const subscribedReleaseGroupCount =
    getSubscribedReleaseGroupCount(summary);
  if (subscribedReleaseGroupCount > 0) {
    return {
      disabled: false,
      label: `已订阅 ${subscribedReleaseGroupCount}/${releaseGroupCount}`,
      subscribed: true,
    };
  }
  if (summary) {
    return {
      disabled: false,
      label: `${releaseGroupCount} 个字幕组`,
      subscribed: false,
    };
  }
  return { disabled: false, label: "载入字幕组", subscribed: false };
}
