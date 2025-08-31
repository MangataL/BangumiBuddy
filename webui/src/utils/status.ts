import {
  CheckCircle2,
  PauseCircle,
  ChevronsDown,
  CircleArrowDown,
  CircleAlert,
  CircleX,
} from "lucide-react";
import { TorrentStatusSet } from "@/api/subscription";

export interface StatusInfo {
  color: string;
  text: string;
  Icon: React.ComponentType<{ className?: string }>;
}

export const renderDownloadStatus = (downloadStatus: string): StatusInfo => {
  let color = "";
  let text = "";
  let Icon = CircleX;
  switch (downloadStatus) {
    case TorrentStatusSet.Downloading:
      color = "bg-blue-500";
      text = "下载中";
      Icon = ChevronsDown;
      break;
    case TorrentStatusSet.Downloaded:
      color = "bg-purple-500";
      text = "已下载";
      Icon = CircleArrowDown;
      break;
    case TorrentStatusSet.Transferred:
      color = "bg-green-500";
      text = "转移完成";
      Icon = CheckCircle2;
      break;
    case TorrentStatusSet.TransferredError:
      color = "bg-red-500";
      text = "转移错误";
      Icon = CircleAlert;
      break;
    case TorrentStatusSet.DownloadError:
      color = "bg-red-500";
      text = "下载错误";
      Icon = CircleArrowDown;
      break;
    case TorrentStatusSet.DownloadPaused:
      color = "bg-gray-500";
      text = "下载暂停";
      Icon = PauseCircle;
      break;
    default:
      color = "bg-gray-400";
      text = "未知状态";
      Icon = CircleX;
  }
  return { color, text, Icon };
};
