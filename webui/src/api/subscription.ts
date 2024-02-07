import { http } from "./index";

// RSS解析响应类型
export interface ParseRSSResponse {
  name: string;
  season: number;
  year: string;
  tmdbID: number;
  rssLink: string;
  releaseGroup: string;
  episodeTotalNum: number;
  airWeekday: number;
}

// 订阅请求类型
export interface SubscribeRequest {
  rssLink: string;
  season: number;
  includeRegs: string[];
  excludeRegs: string[];
  episodeOffset: number;
  priority: number;
  tmdbID: number;
  releaseGroup: string;
  episodeLocation: string;
  episodeTotalNum: number;
  airWeekday: number;
}

// 番剧详情接口
export interface Bangumi {
  subscriptionID: string;
  name: string;
  rssLink: string;
  active: boolean;
  includeRegs: string[];
  excludeRegs: string[];
  priority: number;
  episodeOffset: number;
  season: number;
  year: string;
  tmdbID: number;
  releaseGroup: string;
  episodeLocation: string;
  posterURL: string;
  airWeekday: number;
  episodeTotalNum: number;
  lastAirEpisode: number;
  backdropURL: string;
  overview: string;
  genres: string;
}

// 番剧基础信息类型
export interface ReleaseGroupSubscription {
  subscriptionID: string;
  releaseGroup: string;
  episodeTotalNum: number;
  lastAirEpisode: number;
  priority: number;
  active: boolean;
}

export interface BangumiBase {
  bangumiName: string;
  season: number;
  posterURL: string;
  airWeekday: number;
  releaseGroups: ReleaseGroupSubscription[];
}

// RSS匹配项（与后端types.go中的RSSMatch结构保持一致）
export interface RSSMatch {
  guid: string;
  match: boolean;
  processed: boolean;
  title?: string; // 添加title可选字段用于显示
}

// 文件信息
export interface TorrentFile {
  fileName: string;
  linkName?: string;
}

// 订阅日历项类型
export interface CalendarItem {
  bangumiName: string;
  posterURL: string;
  season: number;
}

// 订阅日历响应类型，根据星期分组
export interface SubscriptionCalendarResponse {
  [weekday: number]: CalendarItem[];
}

export const TorrentStatusSet = {
  Downloading: "downloading",
  Downloaded: "downloaded",
  Transferred: "transferred",
  TransferredError: "transferredError",
  DownloadError: "downloadError",
  DownloadPaused: "downloadPaused",
} as const;

type TorrentStatus = (typeof TorrentStatusSet)[keyof typeof TorrentStatusSet];

export interface Torrent {
  name: string;
  hash: string;
  status: TorrentStatus;
  statusDetail: string;
  downloadSpeed: number;
  progress: number;
  rssGUID: string;
}

export interface RecentUpdatedTorrent {
  posterURL: string;
  bangumiName: string;
  season: number;
  createdAt: string;
  rssItem: string;
  status: TorrentStatus;
  statusDetail: string;
}

export interface ListRecentUpdatedTorrentsResp {
  total: number;
  torrents: RecentUpdatedTorrent[];
}

export interface ListRecentUpdatedTorrentsReq {
  start_time: string;
  end_time: string;
  page: number;
  page_size: number;
}

export const subscriptionAPI = {
  // 解析RSS链接
  parseRSS: async (link: string): Promise<ParseRSSResponse> => {
    return http.get("/bangumis/rss", {
      params: { link },
    });
  },

  // 订阅番剧
  subscribe: async (data: SubscribeRequest): Promise<Bangumi> => {
    return http.post("/bangumis", data);
  },

  // 获取番剧基础信息列表
  listBangumisBase: async (params?: {
    fuzz_name?: string;
    active?: boolean;
    name?: string;
    season?: number;
  }): Promise<BangumiBase[]> => {
    return http.get("/bangumis/base", { params });
  },

  // 获取单个番剧详细信息
  getBangumi: async (id: string): Promise<Bangumi> => {
    return http.get(`/bangumis/${id}`);
  },

  // 获取番剧的RSS匹配情况
  getRSSMatch: async (id: string): Promise<RSSMatch[]> => {
    const response: RSSMatch[] = await http.get(`/bangumis/${id}/rss_match`);
    // 为每个RSSMatch添加title字段，使其在UI中显示更友好
    return response.map((item) => ({
      ...item,
      title: item.guid, // 使用guid作为显示文本
    }));
  },

  // 标记RSS记录的处理状态
  markRSSRecord: async (id: string, guids: string[], processed: boolean) => {
    return http.post(`/bangumis/${id}/rss_match`, {
      guids,
      processed,
    });
  },

  // 更新订阅配置
  updateSubscription: async (
    id: string,
    data: {
      active: boolean;
      includeRegs: string[];
      excludeRegs: string[];
      priority: number;
      episodeOffset: number;
      episodeLocation: string;
      episodeTotalNum: number;
      airWeekday: number;
    }
  ) => {
    return http.put(`/bangumis/${id}`, data);
  },

  // 处理番剧订阅（手动触发下载检查）
  handleSubscription: async (id: string) => {
    return http.post(`/bangumis/${id}/download`);
  },

  // 获取番剧的torrents
  getBangumiTorrents: async (id: string): Promise<Torrent[]> => {
    return http.get(`/bangumis/${id}/torrents`);
  },

  // 获取订阅日历
  getSubscriptionCalendar: async (): Promise<SubscriptionCalendarResponse> => {
    return http.get(`/bangumis/calendar`);
  },

  // 删除订阅
  deleteSubscription: async (
    id: string,
    deleteFiles: boolean = false
  ): Promise<void> => {
    return http.delete(`/bangumis/${id}`, {
      params: { delete_files: deleteFiles },
    });
  },

  // 删除种子
  deleteTorrent: async (hash: string, deleteOriginFiles: boolean) => {
    await http.delete(`/torrents/${hash}`, {
      params: {
        delete_origin_files: deleteOriginFiles,
      },
    });
  },

  // 转移种子
  transferTorrent: async (hash: string) => {
    await http.post(`/torrents/${hash}/transfer`);
  },

  // 获取种子文件
  getTorrentFiles: async (hash: string): Promise<TorrentFile[]> => {
    return await http.get(`/torrents/${hash}/files`);
  },

  // 获取近期更新的种子
  listRecentUpdatedTorrents: async (
    params: ListRecentUpdatedTorrentsReq
  ): Promise<ListRecentUpdatedTorrentsResp> => {
    return http.get(`/torrents/recent`, { params });
  },
};

export default subscriptionAPI;
