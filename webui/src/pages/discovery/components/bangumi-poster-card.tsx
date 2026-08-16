import type {
  BangumiCandidate,
  ReleaseGroupCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";
import { cn } from "@/lib/utils";

import {
  getSubscribedReleaseGroupCount,
  hasLoadedNoReleaseGroups,
} from "../release-group-state";
import { DiscoveryBangumiCard } from "./discovery-bangumi-card";
import { ReleaseGroupPicker } from "./release-group-picker";

interface BangumiPosterCardProps {
  item: BangumiCandidate;
  summary?: ReleaseGroupSummary;
  summaryLoading: boolean;
  summaryError?: string;
  isMovie?: boolean;
  onOpenDetail: (id: string, isMovie: boolean) => void;
  onEnsureSummary: (id: string) => void;
  onAddGroup: (bangumiID: string, group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
}

export function BangumiPosterCard(props: BangumiPosterCardProps) {
  const noReleaseGroups = hasLoadedNoReleaseGroups(
    props.item,
    props.summary
  );
  const subscribedCount = getSubscribedReleaseGroupCount(props.summary);

  return (
    <DiscoveryBangumiCard
      item={props.item}
      muted={noReleaseGroups}
      badge={
        !props.isMovie && subscribedCount > 0 ? (
          <span className="absolute left-2.5 top-2.5 z-[2] rounded-md bg-emerald-700/90 px-1.5 py-0.5 text-[10px] font-semibold tracking-wide text-white shadow-sm backdrop-blur-sm dark:bg-emerald-600/90">
            已订阅
          </span>
        ) : undefined
      }
      onOpenDetail={props.onOpenDetail}
    >
      {!props.isMovie ? (
        <div
          className={cn(
            "flex items-center",
            noReleaseGroups && "opacity-70"
          )}
        >
          <ReleaseGroupPicker
            item={props.item}
            summary={props.summary}
            loading={props.summaryLoading}
            error={props.summaryError}
            className="-ml-1.5"
            onEnsureSummary={() =>
              props.onEnsureSummary(props.item.mikanBangumiID)
            }
            onAdd={(group) =>
              props.onAddGroup(props.item.mikanBangumiID, group)
            }
            onViewSubscription={props.onViewSubscription}
          />
        </div>
      ) : null}
    </DiscoveryBangumiCard>
  );
}
