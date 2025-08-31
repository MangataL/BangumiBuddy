import {
  Film,
  Tv,
  Download,
  Clock,
  CircleX,
  UserRoundCheck,
} from "lucide-react";
import { TaskStatusSet, DownloadTask } from "@/api/magnet";
import { torrentCanTransfer } from "@/utils/util";
import { renderDownloadStatus } from "@/utils/status";

// 类型图标渲染
export const renderTypeIcon = (type: string) => {
  if (type === "tv") {
    return <Tv className="w-5 h-5 text-primary" />;
  }
  return <Film className="w-5 h-5 text-primary" />;
};

// 任务状态信息
export interface StatusInfo {
  color: string;
  text: string;
  Icon: React.ComponentType<{ className?: string }>;
}

// 获取任务状态信息（包含图标）
export const getTaskStatus = (
  taskStatus: string,
  downloadStatus?: string
): StatusInfo => {
  let color = "";
  let text = "";
  let Icon = Download;

  if (taskStatus !== TaskStatusSet.InitSuccess) {
    switch (taskStatus) {
      case TaskStatusSet.WaitingForParsing:
        color = "bg-yellow-500";
        text = "待解析";
        Icon = Clock;
        break;
      case TaskStatusSet.WaitingForConfirmation:
        color = "bg-orange-500";
        text = "待确认";
        Icon = UserRoundCheck;
        break;
      default:
        color = "bg-gray-400";
        text = "未知状态";
        Icon = CircleX;
    }
  } else {
    return renderDownloadStatus(downloadStatus || "");
  }

  return { color, text, Icon };
};

// 判断任务是否可以转移
export const canTransfer = (task: DownloadTask) => {
  return (
    task.status === TaskStatusSet.InitSuccess &&
    torrentCanTransfer(task.downloadStatus)
  );
};
