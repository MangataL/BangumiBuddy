import { useRef, useState } from "react";
import { useNavigate } from "react-router-dom";

import {
  discoveryAPI,
  ReleaseGroupCandidate,
  ResourceCandidate,
  Season,
} from "@/api/discovery";
import magnetAPI, {
  DownloadType,
  DownloadTypeSet,
} from "@/api/magnet";
import { ParseRSSResponse } from "@/api/subscription";
import { ConfirmSubscriptionDialog } from "@/components/subscription/confirm-subscription-dialog";
import { useToast } from "@/hooks/useToast";
import { extractErrorMessage } from "@/utils/error";

import { BangumiInspector } from "@/pages/discovery/components/bangumi-inspector";
import { DiscoverySearchWorkspace } from "@/pages/discovery/components/discovery-search-workspace";
import { DiscoveryToolbar } from "./components/discovery-toolbar";
import { ManualTMDBParseDialog } from "./components/manual-tmdb-parse-dialog";
import { ResourceConfirmDialog } from "./components/resource-confirm-dialog";
import { ScheduleWorkspace } from "./components/schedule-workspace";
import { useBangumiDetail } from "./hooks/use-bangumi-detail";
import { useDiscoverySearch } from "./hooks/use-discovery-search";
import { useSeasonDiscovery } from "./hooks/use-season-discovery";
import { resolveDiscoveryBangumiName, isTMDBMatchFailure } from "./parse-fallback";
import { getDefaultSubscriptionPriority } from "./subscription-defaults";

const emptyParseRSS: ParseRSSResponse = {
  name: "",
  season: 1,
  year: "",
  tmdbID: 0,
  releaseGroup: "",
  episodeTotalNum: 0,
  airWeekday: 0,
  rssLink: "",
  posterURL: "",
  backdropURL: "",
};

interface ManualParseContext {
  bangumiID: string;
  group: ReleaseGroupCandidate;
  errorMessage: string;
  suggestedName: string;
}

export default function DiscoveryPage() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const seasonDiscovery = useSeasonDiscovery();
  const search = useDiscoverySearch();
  const detail = useBangumiDetail();
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [parseRSSResponse, setParseRSSResponse] =
    useState<ParseRSSResponse>(emptyParseRSS);
  const [defaultSubscriptionPriority, setDefaultSubscriptionPriority] =
    useState(1);
  const [resource, setResource] = useState<ResourceCandidate>();
  const [resourceSubmitting, setResourceSubmitting] = useState(false);
  const [downloadType, setDownloadType] = useState<DownloadType>(
    DownloadTypeSet.TV
  );
  const [manualParse, setManualParse] = useState<ManualParseContext>();
  const [manualParseSubmitting, setManualParseSubmitting] = useState(false);
  const candidateRequestIDRef = useRef(0);
  const pendingSubscriptionBangumiIDRef = useRef<string | undefined>(
    undefined
  );
  const searchMode = Boolean(search.submittedQuery);

  const invalidateCandidateRequest = () => {
    candidateRequestIDRef.current += 1;
  };

  const parseCandidate = async (
    bangumiID: string,
    group: ReleaseGroupCandidate,
    tmdbID?: number
  ) => {
    const requestID = candidateRequestIDRef.current + 1;
    candidateRequestIDRef.current = requestID;
    pendingSubscriptionBangumiIDRef.current = undefined;
    try {
      const response = await discoveryAPI.parseCandidateRSS({
        mikanBangumiID: bangumiID,
        mikanSubgroupID: group.mikanSubgroupID,
        tmdbID,
      });
      if (requestID !== candidateRequestIDRef.current) return;
      pendingSubscriptionBangumiIDRef.current = bangumiID;
      setDefaultSubscriptionPriority(getDefaultSubscriptionPriority(group));
      setParseRSSResponse(response);
      setManualParse(undefined);
      setConfirmOpen(true);
    } catch (error) {
      if (requestID !== candidateRequestIDRef.current) return;
      const errorMessage = extractErrorMessage(error);
      if (tmdbID || !isTMDBMatchFailure(errorMessage)) {
        showError(toast, "字幕组解析失败", error);
        return;
      }
      setManualParse({
        bangumiID,
        group,
        errorMessage,
        suggestedName: resolveDiscoveryBangumiName({
          bangumiID,
          detail: detail.detail,
          bangumis: seasonDiscovery.bangumis,
          searchBangumis: search.data.bangumis,
          errorMessage,
        }),
      });
    }
  };

  const retryParseWithTMDB = async (tmdbID: number) => {
    if (!manualParse || manualParseSubmitting) return;
    setManualParseSubmitting(true);
    try {
      await parseCandidate(manualParse.bangumiID, manualParse.group, tmdbID);
    } finally {
      setManualParseSubmitting(false);
    }
  };

  const openResourceDialog = (nextResource: ResourceCandidate) => {
    setDownloadType(nextResource.suggestedDownloadType);
    setResource(nextResource);
    setResourceSubmitting(false);
  };

  const addResource = async () => {
    if (!resource || resourceSubmitting) return;
    setResourceSubmitting(true);
    try {
      const task = await magnetAPI.addTask(resource.magnetLink, downloadType);
      setResource(undefined);
      navigate(`/download?taskID=${task.taskID}`);
    } catch (error) {
      showError(toast, "创建下载任务失败", error);
    } finally {
      setResourceSubmitting(false);
    }
  };

  const setYear = (year: number) => {
    invalidateCandidateRequest();
    detail.close();
    search.clear();
    seasonDiscovery.setYear(year);
  };

  const setSeason = (season: Season) => {
    invalidateCandidateRequest();
    detail.close();
    search.clear();
    seasonDiscovery.setSeason(season);
  };

  const submitSearch = () => {
    invalidateCandidateRequest();
    detail.close();
    void search.submit();
  };

  const browseMovies = () => {
    invalidateCandidateRequest();
    detail.close();
    seasonDiscovery.setActiveWeekday(7);
  };

  const closeSearch = () => {
    invalidateCandidateRequest();
    detail.close();
    search.clear();
  };

  const refreshCurrent = () => {
    if (searchMode) {
      void search.searchPage(search.data.resourcePage);
      return;
    }
    void seasonDiscovery.refresh();
  };

  const handleSubscribed = () => {
    const targetID =
      pendingSubscriptionBangumiIDRef.current ?? detail.selectedID;
    pendingSubscriptionBangumiIDRef.current = undefined;
    if (!targetID) return;
    if (detail.selectedID === targetID) {
      void detail.open(targetID);
    }
    void seasonDiscovery.refreshReleaseGroups(targetID);
  };

  const openDetail = (id: string) => {
    invalidateCandidateRequest();
    void detail.open(id);
  };

  const changeActiveWeekday = (weekday: number) => {
    invalidateCandidateRequest();
    detail.close();
    seasonDiscovery.setActiveWeekday(weekday);
  };

  return (
    <div className="flex h-full min-h-0 flex-col gap-3 overflow-hidden sm:gap-5">
      <DiscoveryToolbar
        year={seasonDiscovery.year}
        season={seasonDiscovery.season}
        query={search.query}
        loading={searchMode ? search.loading : seasonDiscovery.loading}
        onYearChange={setYear}
        onSeasonChange={setSeason}
        onQueryChange={search.setQuery}
        onSearch={submitSearch}
        onRefresh={refreshCurrent}
      />

      {searchMode ? (
        <DiscoverySearchWorkspace
          query={search.submittedQuery}
          data={search.data}
          loading={search.loading}
          error={search.error}
          onClear={closeSearch}
          onRetry={() => void search.searchPage(search.data.resourcePage)}
          onPageChange={(page) => void search.searchPage(page)}
          onOpenDetail={openDetail}
          onAddResource={openResourceDialog}
        />
      ) : (
        <ScheduleWorkspace
          bangumis={seasonDiscovery.bangumis}
          activeBangumis={seasonDiscovery.activeBangumis}
          activeWeekday={seasonDiscovery.activeWeekday}
          currentWeekday={seasonDiscovery.currentWeekday}
          loading={seasonDiscovery.loading}
          error={seasonDiscovery.error}
          summaries={seasonDiscovery.summaries}
          summaryLoadingIDs={seasonDiscovery.summaryLoadingIDs}
          summaryErrors={seasonDiscovery.summaryErrors}
          onActiveWeekdayChange={changeActiveWeekday}
          onBrowseMovies={browseMovies}
          onRetry={() => void seasonDiscovery.refresh()}
          onOpenDetail={openDetail}
          onEnsureSummary={(id) =>
            void seasonDiscovery.retryReleaseGroups(id)
          }
          onAddGroup={parseCandidate}
          onViewSubscription={(id) => navigate(`/?subscriptionID=${id}`)}
        />
      )}

      <BangumiInspector
        open={Boolean(detail.selectedID)}
        detail={detail.detail}
        loading={detail.loading}
        error={detail.error}
        onOpenChange={(open) => {
          if (!open) {
            invalidateCandidateRequest();
            detail.close();
          }
        }}
        onRetry={() => void detail.retry()}
        onAddGroup={parseCandidate}
        onViewSubscription={(id) => navigate(`/?subscriptionID=${id}`)}
        onAddResource={openResourceDialog}
      />

      <ResourceConfirmDialog
        resource={resource}
        submitting={resourceSubmitting}
        downloadType={downloadType}
        onDownloadTypeChange={setDownloadType}
        onOpenChange={(open) => {
          if (!open && !resourceSubmitting) setResource(undefined);
        }}
        onConfirm={addResource}
      />

      <ManualTMDBParseDialog
        open={Boolean(manualParse)}
        errorMessage={manualParse?.errorMessage ?? ""}
        suggestedName={manualParse?.suggestedName ?? ""}
        releaseGroupName={manualParse?.group.name ?? ""}
        submitting={manualParseSubmitting}
        onOpenChange={(open) => {
          if (!open && !manualParseSubmitting) setManualParse(undefined);
        }}
        onConfirm={(tmdbID) => void retryParseWithTMDB(tmdbID)}
      />

      <ConfirmSubscriptionDialog
        open={confirmOpen}
        parseRSSRsp={parseRSSResponse}
        defaultPriority={defaultSubscriptionPriority}
        onOpenChange={setConfirmOpen}
        onSubscribed={handleSubscribed}
      />
    </div>
  );
}

function showError(
  toast: ReturnType<typeof useToast>["toast"],
  title: string,
  error: unknown
) {
  toast({
    title,
    description: extractErrorMessage(error),
    variant: "destructive",
  });
}
