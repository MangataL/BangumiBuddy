import { http } from "./index";

// 下载器配置类型
export interface DownloaderConfig {
  downloadType: string;
  qbittorrent: {
    host: string;
    username: string;
    password: string;
  };
}

// 下载管理器配置类型
export interface DownloadManagerConfig {
  tvSavePath: string;
  movieSavePath: string;
}

// TMDB配置类型
export interface TMDBConfig {
  tmdbToken: string;
  alternateURL: boolean;
}

// 订阅配置类型
export interface SubscriptionConfig {
  rssCheckInterval: number;
  includeRegs: string[];
  excludeRegs: string[];
  autoStop: boolean;
}

// 文件转移配置类型
export interface TransferConfig {
  interval: number;
  tvPath: string;
  tvFormat: string;
  transferType: string;
}

// 通知配置类型
export interface NoticeConfig {
  enabled: boolean;
  type: string;
  telegram: {
    token: string;
    chatID: number;
  };
  email: {
    host: string;
    username: string;
    password: string;
    from: string;
    to: string[] | null;
    ssl: boolean;
  };
  bark: {
    serverPath: string;
    sound: string;
    interruption: string;
    autoSave: boolean;
  };
  noticePoints: {
    subscriptionUpdated: boolean;
    downloaded: boolean;
    transferred: boolean;
    error: boolean;
  };
}

export interface QBittorrentConfig {
  host: string;
  username: string;
  password: string;
}

// API函数
export const configAPI = {
  // 下载器配置
  getDownloaderConfig: (): Promise<DownloaderConfig> =>
    http.get("/config/download/downloader") as Promise<DownloaderConfig>,
  setDownloaderConfig: (config: DownloaderConfig): Promise<void> =>
    http.put("/config/download/downloader", config) as Promise<void>,

  // 下载管理器配置
  getDownloadManagerConfig: (): Promise<DownloadManagerConfig> =>
    http.get("/config/download/manager") as Promise<DownloadManagerConfig>,
  setDownloadManagerConfig: (config: DownloadManagerConfig): Promise<void> =>
    http.put("/config/download/manager", config) as Promise<void>,

  // TMDB配置
  getTMDBConfig: (): Promise<TMDBConfig> =>
    http.get("/config/tmdb") as Promise<TMDBConfig>,
  setTMDBConfig: (config: TMDBConfig): Promise<void> =>
    http.put("/config/tmdb", config) as Promise<void>,

  // 订阅配置
  getSubscriptionConfig: (): Promise<SubscriptionConfig> =>
    http.get("/config/subscriber") as Promise<SubscriptionConfig>,
  setSubscriptionConfig: (config: SubscriptionConfig): Promise<void> =>
    http.put("/config/subscriber", config) as Promise<void>,

  // 文件转移配置
  getTransferConfig: (): Promise<TransferConfig> =>
    http.get("/config/transfer") as Promise<TransferConfig>,
  setTransferConfig: (config: TransferConfig): Promise<void> =>
    http.put("/config/transfer", config) as Promise<void>,

  // 通知配置
  getNoticeConfig: (): Promise<NoticeConfig> =>
    http.get("/config/notice") as Promise<NoticeConfig>,
  setNoticeConfig: (config: NoticeConfig): Promise<void> =>
    http.put("/config/notice", config) as Promise<void>,

  // 检查qBittorrent连通性
  checkQBittorrentConnection: (config: QBittorrentConfig): Promise<void> =>
    http.post("/downloader/qbittorrent/check", config) as Promise<void>,
};
