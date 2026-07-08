import { useState, type CSSProperties } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Loader2,
} from "lucide-react";

import type {
  BangumiCandidate,
  ResourceCandidate,
} from "@/api/discovery";
import { Button } from "@/components/ui/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";

import type { DiscoverySearchState } from "../hooks/use-discovery-search";
import { formatDate } from "../discovery-options";
import { discoveryPosterGridClassName } from "../poster-grid";
import { getSearchPageCount } from "../search-pagination";
import { DiscoveryBangumiCard } from "./discovery-bangumi-card";
import {
  DiscoveryErrorState,
  EmptyDiscoveryState,
  PosterGridSkeleton,
} from "./discovery-state-views";
import { ResourceList } from "./resource-list";

interface DiscoverySearchWorkspaceProps {
  query: string;
  data: DiscoverySearchState;
  loading: boolean;
  error: string;
  onClear: () => void;
  onRetry: () => void;
  onPageChange: (page: number) => void;
  onOpenDetail: (id: string) => void;
  onAddResource: (resource: ResourceCandidate) => void;
}

export function DiscoverySearchWorkspace(
  props: DiscoverySearchWorkspaceProps
) {
  const [tab, setTab] = useState("all");
  const hasData =
    props.data.bangumiTotal > 0 || props.data.resourceTotal > 0;
  const pageCount = getSearchPageCount(
    props.data.resourceTotal,
    props.data.resourcePageSize
  );

  return (
    <section className="flex min-h-0 flex-1 flex-col">
      <div className="flex flex-none flex-col gap-4 border-b border-border/70 pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex min-w-0 items-start gap-3">
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-9 w-9 shrink-0 rounded-xl"
            aria-label="返回季度放送表"
            onClick={props.onClear}
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-primary">
              Search results
            </p>
            <h2 className="mt-1 truncate text-xl font-bold tracking-tight">
              「{props.query}」
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {props.data.bangumiTotal} 部番剧 · {props.data.resourceTotal} 条资源
            </p>
          </div>
        </div>
        {props.loading && hasData && (
          <span className="inline-flex items-center gap-2 text-xs text-muted-foreground">
            <Loader2 className="h-3.5 w-3.5 animate-spin motion-reduce:animate-none" />
            正在更新结果
          </span>
        )}
      </div>

      {props.loading && !hasData ? (
        <div className="min-h-0 flex-1 overflow-y-auto pt-5">
          <PosterGridSkeleton count={8} />
        </div>
      ) : props.error && !hasData ? (
        <DiscoveryErrorState
          title="搜索失败"
          description={props.error}
          onRetry={props.onRetry}
        />
      ) : !hasData ? (
        <EmptyDiscoveryState
          title="没有找到相关内容"
          description="尝试使用更短的番剧名称、日文标题或字幕组名称。"
        />
      ) : (
        <Tabs
          value={tab}
          onValueChange={setTab}
          className="flex min-h-0 flex-1 flex-col pt-4"
        >
          <TabsList className="grid w-full max-w-sm flex-none grid-cols-3 rounded-xl">
            <TabsTrigger value="all" className="rounded-lg">
              全部
            </TabsTrigger>
            <TabsTrigger value="bangumi" className="rounded-lg">
              番剧
            </TabsTrigger>
            <TabsTrigger value="resource" className="rounded-lg">
              资源
            </TabsTrigger>
          </TabsList>
          <div className="min-h-0 flex-1 overflow-y-auto pt-5">
            {props.error && (
              <div className="mb-4 rounded-xl border border-destructive/20 bg-destructive/5 px-4 py-3 text-sm text-destructive">
                {props.error}
              </div>
            )}
            <TabsContent value="all" className="mt-0 space-y-8">
              {props.data.bangumis.length > 0 && (
                <SearchSection
                  title="番剧"
                  count={props.data.bangumiTotal}
                >
                  <SearchBangumiGrid
                    bangumis={props.data.bangumis}
                    onOpenDetail={props.onOpenDetail}
                  />
                </SearchSection>
              )}
              {props.data.resources.length > 0 && (
                <SearchSection
                  title="资源"
                  count={props.data.resourceTotal}
                >
                  <ResourceList
                    resources={props.data.resources}
                    onAddResource={props.onAddResource}
                  />
                  <ResourcePagination
                    page={props.data.resourcePage}
                    pageCount={pageCount}
                    loading={props.loading}
                    onPageChange={props.onPageChange}
                  />
                </SearchSection>
              )}
            </TabsContent>
            <TabsContent value="bangumi" className="mt-0">
              {props.data.bangumis.length > 0 ? (
                <SearchBangumiGrid
                  bangumis={props.data.bangumis}
                  onOpenDetail={props.onOpenDetail}
                />
              ) : (
                <EmptyDiscoveryState
                  title="没有番剧结果"
                  description="可以切换到资源页签继续查找单集或合集。"
                />
              )}
            </TabsContent>
            <TabsContent value="resource" className="mt-0">
              <ResourceList
                resources={props.data.resources}
                onAddResource={props.onAddResource}
              />
              <ResourcePagination
                page={props.data.resourcePage}
                pageCount={pageCount}
                loading={props.loading}
                onPageChange={props.onPageChange}
              />
            </TabsContent>
          </div>
        </Tabs>
      )}
    </section>
  );
}

function SearchSection({
  title,
  count,
  children,
}: {
  title: string;
  count: number;
  children: React.ReactNode;
}) {
  return (
    <section>
      <div className="mb-4 flex items-baseline gap-2">
        <h3 className="text-base font-semibold">{title}</h3>
        <span className="text-xs tabular-nums text-muted-foreground">
          {count}
        </span>
      </div>
      {children}
    </section>
  );
}

function SearchBangumiGrid({
  bangumis,
  onOpenDetail,
}: {
  bangumis: BangumiCandidate[];
  onOpenDetail: (id: string) => void;
}) {
  return (
    <div className={discoveryPosterGridClassName}>
      {bangumis.map((item, index) => (
        <div
          key={item.mikanBangumiID}
          className="discovery-poster-enter"
          style={{ "--poster-index": index } as CSSProperties}
        >
          <DiscoveryBangumiCard item={item} onOpenDetail={onOpenDetail}>
            <p className="flex min-w-0 items-center gap-1.5 text-[11px] leading-5 text-muted-foreground sm:text-xs">
              <CalendarDays className="h-3.5 w-3.5 shrink-0 opacity-70" />
              <span className="truncate">
                {item.airStartDate
                  ? `${formatDate(item.airStartDate)} 开播`
                  : "开播日期未公布"}
              </span>
            </p>
          </DiscoveryBangumiCard>
        </div>
      ))}
    </div>
  );
}

function ResourcePagination({
  page,
  pageCount,
  loading,
  onPageChange,
}: {
  page: number;
  pageCount: number;
  loading: boolean;
  onPageChange: (page: number) => void;
}) {
  if (pageCount <= 1) return null;
  return (
    <nav
      aria-label="资源结果分页"
      className="mt-6 flex items-center justify-between border-t border-border/70 pt-4"
    >
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="rounded-lg"
        disabled={loading || page <= 1}
        onClick={() => onPageChange(page - 1)}
      >
        <ChevronLeft className="h-4 w-4" />
        上一页
      </Button>
      <span className="text-xs tabular-nums text-muted-foreground">
        {page} / {pageCount}
      </span>
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="rounded-lg"
        disabled={loading || page >= pageCount}
        onClick={() => onPageChange(page + 1)}
      >
        下一页
        <ChevronRight className="h-4 w-4" />
      </Button>
    </nav>
  );
}
