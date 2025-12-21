import { TorrentStatusSet } from "@/api/subscription";

// 判断任务是否可以转移
export const torrentCanTransfer = (downloadStatus: string) => {
  return (
    downloadStatus === TorrentStatusSet.Downloaded ||
    downloadStatus === TorrentStatusSet.Transferred ||
    downloadStatus === TorrentStatusSet.TransferredError
  );
};

// 格式化文件大小
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}