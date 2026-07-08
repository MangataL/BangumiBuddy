import { http } from "./index";
import type { ParseRSSResponse } from "./subscription";

export const SeasonSet = {
  Winter: "winter",
  Spring: "spring",
  Summer: "summer",
  Fall: "fall",
} as const;

export type Season = (typeof SeasonSet)[keyof typeof SeasonSet];

export interface BangumiCandidate {
  mikanBangumiID: string;
  name: string;
  posterURL: string;
  weekday: number;
  airStartDate: string;
  releaseGroupsKnownEmpty?: boolean;
}

export interface ReleaseGroupCandidate {
  mikanSubgroupID: string;
  name: string;
  subscribed: boolean;
  subscriptionID?: string;
  previousMaxPriority?: number;
}

export interface ResourceCandidate {
  title: string;
  mikanSubgroupID?: string;
  magnetLink: string;
  size: string;
  publishedAt?: string;
  releaseGroup: string;
  suggestedDownloadType: "tv" | "movie";
}

export interface SearchDiscoveryResp {
  bangumis: BangumiCandidate[];
  resources: ResourceCandidate[];
  bangumiTotal: number;
  resourceTotal: number;
  resourcePage: number;
  resourcePageSize: number;
}

export interface ReleaseGroupSummary {
  mikanBangumiID: string;
  releaseGroups: ReleaseGroupCandidate[];
}

export interface ReleaseGroupSummaryFailure {
  mikanBangumiID: string;
  message: string;
}

export interface BatchReleaseGroupsResp {
  items: ReleaseGroupSummary[];
  failures: ReleaseGroupSummaryFailure[];
}

export interface BangumiDetail {
  mikanBangumiID: string;
  name: string;
  posterURL: string;
  airStartDate: string;
  episodeTotalText: string;
  overview: string;
  releaseGroups: ReleaseGroupCandidate[];
  resources: ResourceCandidate[];
}

export interface CandidateRSSRequest {
  mikanBangumiID: string;
  mikanSubgroupID: string;
  tmdbID?: number;
}

export const DISCOVERY_SEARCH_PAGE_SIZE = 20;

export const discoveryAPI = {
  listBangumis: async (params: {
    year: number;
    season: Season;
  }): Promise<BangumiCandidate[]> => {
    return http.get("/discovery/mikan/bangumis", { params });
  },

  search: async (
    q: string,
    page = 1,
    pageSize = DISCOVERY_SEARCH_PAGE_SIZE
  ): Promise<SearchDiscoveryResp> => {
    return http.get("/discovery/mikan/search", {
      params: { q, page, pageSize },
    });
  },

  batchReleaseGroups: async (
    bangumiIDs: string[]
  ): Promise<BatchReleaseGroupsResp> => {
    return http.post("/discovery/mikan/bangumis/release-groups/batch", {
      bangumiIDs,
    });
  },

  getBangumi: async (id: string): Promise<BangumiDetail> => {
    return http.get(`/discovery/mikan/bangumis/${id}`);
  },

  parseCandidateRSS: async (
    data: CandidateRSSRequest
  ): Promise<ParseRSSResponse> => {
    return http.post("/discovery/mikan/rss/parse", data);
  },
};
