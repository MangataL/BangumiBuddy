import { http } from "./index";
import { type TorrentStatus } from "./subscription";

export const DownloadTypeSet = {
  TV: "tv",
  Movie: "movie",
} as const;

export type DownloadType =
  (typeof DownloadTypeSet)[keyof typeof DownloadTypeSet];

export const DownloadTypeLabels = {
  [DownloadTypeSet.TV]: "番剧",
  [DownloadTypeSet.Movie]: "剧场版",
} as const;

export interface TorrentFileMeta {
  mediaType: DownloadType;
  chineseName: string;
  year: string;
  tmdbID: number;
}

export interface TorrentFile {
  fileName: string;
  season: number;
  episode: number;
  media: boolean;
  download: boolean;
  linkFile: string;
  meta?: TorrentFileMeta;
}

export interface Torrent {
  hash: string;
  name: string;
  files: TorrentFile[];
  size: number;
}

export interface Meta {
  chineseName: string;
  year: string;
  tmdbID: number;
  releaseGroup: string;
}

export const TaskStatusSet = {
  WaitingForParsing: "waiting for parsing",
  WaitingForConfirmation: "waiting for confirmation",
  InitSuccess: "init success",
} as const;

export type TaskStatus = (typeof TaskStatusSet)[keyof typeof TaskStatusSet];

export interface Task {
  taskID: string;
  magnetLink: string;
  torrent: Torrent;
  createdAt: string;
  downloadType: DownloadType;
  meta: Meta;
  taskStatus: TaskStatus;
}

export interface DownloadTask {
  taskID: string;
  torrent: Torrent;
  magnetLink: string;
  createdAt: string;
  downloadType: DownloadType;
  meta: Meta;
  status: TaskStatus;
  downloadStatus: TorrentStatus | "";
  downloadStatusDetail: string;
  downloadSpeed: number;
  progress: number;
}

export interface AddTaskRequest {
  magnetLink: string;
  type: DownloadType;
}

export interface ListTasksReq {
  page: number;
  page_size: number;
}

export interface ListTasksResp {
  total: number;
  tasks: DownloadTask[];
}

export interface AddSubtitlesRequest {
  subtitleDir: string;
  dstDir: string;
  episodeLocation?: string;
  episodeOffset?: number;
}

export interface AddSubtitlesResponse {
  successCount: number;
}

export interface ListDirsResp {
  dirs: DirInfo[];
  files: FileInfo[];
  filePathSplit: string;
  fileRoots: string[];
}

export interface DirInfo {
  path: string;
  name: string;
  hasDir: boolean;
  subtitleCount: number;
}

export interface FileInfo {
  path: string;
  name: string;
  size: number;
}

const magnetAPI = {
  // 添加磁力任务
  addTask: async (magnetLink: string, type: DownloadType): Promise<Task> => {
    return http.post("/magnets", {
      magnetLink,
      type,
    });
  },

  // 列出磁力任务
  listTasks: async (params: ListTasksReq): Promise<ListTasksResp> => {
    return http.get("/magnets", {
      params,
    });
  },

  // 获取磁力任务
  getTask: async (taskID: string): Promise<DownloadTask> => {
    return http.get(`/magnets/${taskID}`);
  },

  // 初始化磁力任务
  initTask: (taskID: string, tmdbID: number) => {
    return http.put<Task>(`/magnet/init/${taskID}?tmdb_id=${tmdbID}`);
  },

  // 更新磁力任务
  updateTask: (
    taskID: string,
    data: {
      tmdbID?: number;
      releaseGroup?: string;
      torrent?: Torrent;
      continueDownload?: boolean;
    }
  ) => {
    return http.put<void>(`/magnets/${taskID}`, data);
  },

  // 删除磁力任务
  deleteTask: (taskID: string, deleteFiles: boolean = false) => {
    return http.delete<void>(`/magnets/${taskID}`, {
      params: { delete_files: deleteFiles },
    });
  },

  // 添加字幕
  addSubtitles: (
    taskID: string,
    data: {
      subtitleFiles: Record<string, string>;
      preserveOriginal?: boolean;
    }
  ): Promise<AddSubtitlesResponse> => {
    return http.post<AddSubtitlesResponse>(
      `/magnets/${taskID}/subtitles`,
      data
    );
  },

  // 预览添加字幕
  previewAddSubtitles: (
    taskID: string,
    data: {
      subtitlePath: string;
      dstPath: string;
      episodeLocation?: string;
      episodeOffset?: number;
      season?: number;
      extensionLevel?: number;
    }
  ): Promise<PreviewAddSubtitlesResponse> => {
    return http.post<PreviewAddSubtitlesResponse>(
      `/magnets/${taskID}/subtitles/preview`,
      data
    );
  },

  // 查找任务中相似的文件
  findSimilarFiles: (taskID: string, filePath: string): Promise<string[]> => {
    return http.get(`/magnets/${taskID}/files`, {
      params: { similar_file_path: filePath },
    });
  },

  // 列出目录（返回目录路径数组）
  listDirs: async (path: string): Promise<ListDirsResp> => {
    return http.get("/utils/dirs", {
      params: { path },
    });
  },
};

export interface PreviewAddSubtitlesResponse {
  subtitleFiles: Record<string, AddSubtitlesResult>;
}

export interface AddSubtitlesResult {
  subtitleFile: string;
  newFileName?: string;
  targetPath?: string;
  mediaFileName?: string;
  error?: string;
}

export default magnetAPI;
