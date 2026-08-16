import { CheckCircle2, Loader2 } from "lucide-react";

import type {
  BangumiDetail,
  ReleaseGroupCandidate,
  ResourceCandidate,
} from "@/api/discovery";
import { Button } from "@/components/ui/button";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";

import { orderReleaseGroups } from "../release-group-state";
import { getResourcesForReleaseGroup } from "../release-group-resources";
import { DiscoveryErrorState } from "./discovery-state-views";
import { PosterImage } from "./poster-image";
import { ResourceList } from "./resource-list";

interface BangumiInspectorProps {
  open: boolean;
  detail?: BangumiDetail;
  loading: boolean;
  error: string;
  isMovie: boolean;
  onOpenChange: (open: boolean) => void;
  onRetry: () => void;
  onAddGroup: (bangumiID: string, group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
  onAddResource: (resource: ResourceCandidate) => void;
}

export function BangumiInspector(props: BangumiInspectorProps) {
  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent
        side="right"
        className="w-full overflow-hidden p-0 motion-reduce:animate-none motion-reduce:transition-none sm:max-w-2xl"
      >
        <SheetHeader className="sr-only">
          <SheetTitle>番剧详情</SheetTitle>
          <SheetDescription>
            {props.isMovie
              ? "查看剧场版概况并下载资源"
              : "查看番剧概况、订阅字幕组并展开下载组内资源"}
          </SheetDescription>
        </SheetHeader>
        {props.loading && (
          <div className="flex h-full items-center justify-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin motion-reduce:animate-none" />
            正在载入番剧详情
          </div>
        )}
        {!props.loading && props.error && (
          <DiscoveryErrorState
            title="详情载入失败"
            description={props.error}
            onRetry={props.onRetry}
          />
        )}
        {!props.loading && props.detail && (
          <InspectorContent
            detail={props.detail}
            isMovie={props.isMovie}
            onSubscribeGroup={(group) =>
              props.onAddGroup(props.detail!.mikanBangumiID, group)
            }
            onViewSubscription={props.onViewSubscription}
            onDownloadResource={props.onAddResource}
          />
        )}
      </SheetContent>
    </Sheet>
  );
}

function InspectorContent({
  detail,
  isMovie,
  onSubscribeGroup,
  onViewSubscription,
  onDownloadResource,
}: {
  detail: BangumiDetail;
  isMovie: boolean;
  onSubscribeGroup: (group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
  onDownloadResource: (resource: ResourceCandidate) => void;
}) {
  const orderedReleaseGroups = orderReleaseGroups(detail.releaseGroups);
  return (
    <div className="flex h-full min-h-0 flex-col bg-background">
      <div className="flex-none border-b border-border/70 px-5 pb-5 pt-10 sm:px-6">
        <div className="flex gap-4">
          <div className="w-24 shrink-0">
            <PosterImage
              src={detail.posterURL}
              alt={detail.name}
              className="rounded-xl ring-1 ring-black/5 dark:ring-white/10"
            />
          </div>
          <div className="min-w-0 flex-1 self-end pb-1">
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-primary">
              Bangumi detail
            </p>
            <h2 className="mt-1 line-clamp-3 pr-7 text-2xl font-bold leading-tight tracking-tight">
              {detail.name}
            </h2>
            <div className="mt-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
              {detail.airStartDate && <span>{detail.airStartDate} 开播</span>}
              {detail.episodeTotalText && <span>{detail.episodeTotalText}</span>}
              <span>
                {detail.releaseGroups.length} 个
                {isMovie ? "资源组" : "字幕组"}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-5 py-6 sm:px-6">
        {detail.overview && (
          <section className="border-b border-border/70 pb-6">
            <h3 className="text-sm font-semibold">概况</h3>
            <p className="mt-3 whitespace-pre-line text-sm leading-7 text-foreground/80">
              {detail.overview}
            </p>
          </section>
        )}

        <section className={detail.overview ? "pt-6" : undefined}>
          <div className="mb-4 flex items-baseline justify-between gap-3">
            <div>
              <h3 className="text-base font-semibold">
                {isMovie ? "下载资源" : "字幕组"}
              </h3>
            </div>
            <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
              {detail.releaseGroups.length} 个
            </span>
          </div>

          {detail.releaseGroups.length > 0 ? (
            <Accordion type="multiple" className="space-y-3">
              {orderedReleaseGroups.map((group) => {
                const resources = getResourcesForReleaseGroup(
                  group,
                  detail.resources
                );
                return (
                  <AccordionItem
                    key={group.mikanSubgroupID}
                    value={group.mikanSubgroupID}
                    className="rounded-2xl border border-border/80 bg-card px-3"
                  >
                    <div className="flex items-center gap-2 [&>h3]:min-w-0 [&>h3]:flex-1">
                      <AccordionTrigger className="min-w-0 gap-3 py-3 text-left hover:no-underline">
                        <span className="min-w-0">
                          <span className="flex flex-wrap items-center gap-1.5">
                            <span className="min-w-0 truncate text-sm font-semibold">
                              {group.name}
                            </span>
                          </span>
                          <span className="mt-1 block text-xs font-normal text-muted-foreground">
                            {resources.length} 条资源
                          </span>
                        </span>
                      </AccordionTrigger>
                      {!isMovie ? (
                        <ReleaseGroupAction
                          group={group}
                          onSubscribe={onSubscribeGroup}
                          onViewSubscription={onViewSubscription}
                        />
                      ) : null}
                    </div>
                    <AccordionContent className="border-t border-border/70 pt-4 motion-reduce:animate-none">
                      <ResourceList
                        resources={resources}
                        onAddResource={onDownloadResource}
                      />
                    </AccordionContent>
                  </AccordionItem>
                );
              })}
            </Accordion>
          ) : (
            <p className="rounded-xl bg-muted/60 px-4 py-8 text-center text-sm text-muted-foreground">
              {isMovie ? "暂无可用资源" : "暂无可订阅字幕组"}
            </p>
          )}
        </section>
      </div>
    </div>
  );
}

function ReleaseGroupAction({
  group,
  onSubscribe,
  onViewSubscription,
}: {
  group: ReleaseGroupCandidate;
  onSubscribe: (group: ReleaseGroupCandidate) => void;
  onViewSubscription: (id: string) => void;
}) {
  if (!group.subscribed) {
    return (
      <Button
        type="button"
        size="sm"
        className="shrink-0 rounded-lg bg-gradient-to-r from-primary to-blue-500 text-white hover:opacity-90"
        aria-label={`订阅字幕组 ${group.name}`}
        onClick={() => onSubscribe(group)}
      >
        订阅
      </Button>
    );
  }

  if (group.subscriptionID) {
    return (
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="shrink-0 rounded-lg"
        aria-label={`查看 ${group.name} 的订阅`}
        onClick={() => onViewSubscription(group.subscriptionID!)}
      >
        <CheckCircle2 className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
        查看
      </Button>
    );
  }

  return (
    <span className="inline-flex shrink-0 items-center gap-1 text-xs font-medium text-emerald-600 dark:text-emerald-400">
      <CheckCircle2 className="h-4 w-4" />
      已订阅
    </span>
  );
}
