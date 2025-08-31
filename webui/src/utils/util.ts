import { TorrentStatusSet } from "@/api/subscription";

// 判断任务是否可以转移
export const torrentCanTransfer = (downloadStatus: string) => {
  return (
    downloadStatus === TorrentStatusSet.Downloaded ||
    downloadStatus === TorrentStatusSet.Transferred ||
    downloadStatus === TorrentStatusSet.TransferredError
  );
};
