import { useMemo, useState, type CSSProperties } from "react";
import { ChevronLeft, ChevronRight, Loader2 } from "lucide-react";

import type {
  BangumiCandidate,
  ReleaseGroupCandidate,
  ReleaseGroupSummary,
} from "@/api/discovery";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

import {
  filterScheduleBangumis,
  formatWeekdayLabel,
  groupBangumisByWeekday,
  ScheduleFilter,
  weekdayOptions,
} from "../discovery-options";
import { discoveryPosterGridClassName } from "../poster-grid";
import { getAdjacentWeekdayValue } from "../weekday-schedule";
import { BangumiPosterCard } from "./bangumi-poster-card";
import { BroadcastDayStrip } from "./broadcast-day-strip";
import {
  DiscoveryErrorState,
  EmptyDiscoveryState,
  PosterGridSkeleton,
} from "./discovery-state-views";

interface ScheduleWorkspaceProps {
  bangumis: BangumiCandidate[];
  activeBangumis: BangumiCandidate[];
  activeWeekday: number;
  currentWeekday: number;
  loading: boolean;
  error: string;
  summaries: Record<string, ReleaseGroupSummary>;
  summaryLoadingIDs: string[];
  summaryErrors: Record<string, string>;
  onActiveWeekdayChange: (weekday: number) => void;
  onBrowseMovies: () => void;
  onRetry: () => void;
  onOpenDetail: (id: string) => void;
  onEnsureSummary: (id: string) => void;
  onAddGroup: (bangumiID: string, group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
}

const filters: { value: ScheduleFilter; label: string }[] = [
  { value: "all", label: "全部" },
  { value: "available", label: "可订阅" },
  { value: "subscribed", label: "已订阅" },
];

export function ScheduleWorkspace(props: ScheduleWorkspaceProps) {
  const [filter, setFilter] = useState<ScheduleFilter>("all");
  const groups = useMemo(
    () => groupBangumisByWeekday(props.bangumis),
    [props.bangumis]
  );
  const counts = useMemo(
    () =>
      Object.fromEntries(
        Object.entries(groups).map(([weekday, items]) => [
          Number(weekday),
          items.length,
        ])
      ),
    [groups]
  );
  const visibleBangumis = useMemo(
    () =>
      filterScheduleBangumis(
        props.activeBangumis,
        props.summaries,
        filter
      ),
    [filter, props.activeBangumis, props.summaries]
  );
  const isMovie = props.activeWeekday === 7;

  const shiftWeekday = (direction: -1 | 1) => {
    const next = getAdjacentWeekdayValue(
      weekdayOptions,
      props.activeWeekday,
      direction
    );
    if (next !== undefined) props.onActiveWeekdayChange(next);
  };

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-2.5 sm:gap-5">
      <div className="flex-none">
        <BroadcastDayStrip
          activeWeekday={props.activeWeekday}
          currentWeekday={props.currentWeekday}
          counts={counts}
          onChange={props.onActiveWeekdayChange}
          onBrowseMovies={props.onBrowseMovies}
        />
      </div>

      <div className="flex min-h-0 flex-1 flex-col">
        <div className="flex flex-none items-center justify-between gap-2 border-b border-border/70 pb-2.5 sm:items-end sm:pb-4">
          {/* Desktop keeps the weekday heading; mobile relies on the selected chip. */}
          <div className="hidden min-w-0 sm:block">
            <div className="flex flex-wrap items-baseline gap-2">
              <h2 className="text-xl font-bold tracking-tight">
                {isMovie
                  ? "剧场版"
                  : formatWeekdayLabel(props.activeWeekday)}
              </h2>
              <span className="text-sm tabular-nums text-muted-foreground">
                {props.activeBangumis.length} 部
              </span>
              {props.loading && props.bangumis.length > 0 && (
                <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground motion-reduce:animate-none" />
              )}
            </div>
          </div>

          <div className="flex w-full items-center justify-between gap-2 sm:w-auto sm:justify-end">
            <div
              className="flex rounded-lg bg-muted p-0.5 sm:rounded-xl sm:p-1"
              aria-label="筛选番剧"
            >
              {filters.map((item) => (
                <button
                  key={item.value}
                  type="button"
                  className={cn(
                    "rounded-md px-2.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary sm:rounded-lg sm:px-3 sm:py-1.5",
                    filter === item.value
                      ? "bg-background text-foreground shadow-sm"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                  aria-pressed={filter === item.value}
                  onClick={() => setFilter(item.value)}
                >
                  {item.label}
                </button>
              ))}
            </div>
            <div className="flex items-center gap-1.5 sm:gap-1">
              {props.loading && props.bangumis.length > 0 && (
                <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground motion-reduce:animate-none sm:hidden" />
              )}
              {!isMovie && (
                <div className="hidden items-center gap-1 sm:flex">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 rounded-lg"
                    aria-label="上一个放送日"
                    onClick={() => shiftWeekday(-1)}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 rounded-lg"
                    aria-label="下一个放送日"
                    onClick={() => shiftWeekday(1)}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto pt-3 sm:pt-5">
          {props.loading && props.bangumis.length === 0 ? (
            <PosterGridSkeleton />
          ) : props.error && props.bangumis.length === 0 ? (
            <DiscoveryErrorState
              description={props.error}
              onRetry={props.onRetry}
            />
          ) : visibleBangumis.length === 0 ? (
            <EmptyDiscoveryState
              title={
                props.activeBangumis.length === 0
                  ? "这一天暂无放送"
                  : "没有符合筛选条件的番剧"
              }
              description={
                props.activeBangumis.length === 0
                  ? "可以切换到其他放送日，或使用上方搜索查找番剧。"
                  : "切换到“全部”即可重新查看这一天的完整放送表。"
              }
            />
          ) : (
            <>
              {props.error && (
                <div className="mb-4 rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
                  刷新失败，当前仍显示上一次结果：{props.error}
                </div>
              )}
              <div className={`${discoveryPosterGridClassName} pb-6`}>
                {visibleBangumis.map((item, index) => (
                  <div
                    key={item.mikanBangumiID}
                    className="discovery-poster-enter"
                    style={
                      { "--poster-index": index } as CSSProperties
                    }
                  >
                    <BangumiPosterCard
                      item={item}
                      summary={props.summaries[item.mikanBangumiID]}
                      summaryLoading={props.summaryLoadingIDs.includes(
                        item.mikanBangumiID
                      )}
                      summaryError={props.summaryErrors[item.mikanBangumiID]}
                      onOpenDetail={props.onOpenDetail}
                      onEnsureSummary={props.onEnsureSummary}
                      onAddGroup={props.onAddGroup}
                      onViewSubscription={props.onViewSubscription}
                    />
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </section>
  );
}
