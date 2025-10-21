import { useState, useEffect } from "react";
import {
  Loader2,
  Film,
  Tv,
  Users,
  Clapperboard,
  HardDrive,
  Database,
  Activity,
  FolderTree,
  Download,
  Calendar,
  ArrowUpFromLine,
  Info,
} from "lucide-react";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import magnetAPI, {
  DownloadTask,
  TaskStatusSet,
  TorrentFile,
  DownloadTypeLabels,
  DownloadTypeSet,
} from "@/api/magnet";
import { Meta } from "@/api/meta";
import subscriptionAPI, { TorrentStatusSet } from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { useMobile } from "@/hooks/useMobile";
import { extractErrorMessage } from "@/utils/error";
import { FileTree } from "./file-tree";
import { renderTypeIcon, getTaskStatus, canTransfer } from "./magnet-utils";
import { TMDBInput } from "@/components/tmdb";

interface MagnetDetailDialogProps {
  taskID: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUpdated?: () => void;
}

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}

export function MagnetDetailDialog({
  taskID,
  open,
  onOpenChange,
  onUpdated,
}: MagnetDetailDialogProps) {
  const { toast } = useToast();
  const isMobile = useMobile();
  const [task, setTask] = useState<DownloadTask | null>(null);
  const [loading, setLoading] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [transferring, setTransferring] = useState(false);
  const [editedReleaseGroup, setEditedReleaseGroup] = useState("");
  const [editedTmdbID, setEditedTmdbID] = useState(0);

  // 获取任务详情
  const fetchTaskDetail = async () => {
    if (!taskID) return;

    try {
      setLoading(true);
      const data = await magnetAPI.getTask(taskID);
      setTask(data);
      setEditedReleaseGroup(data.meta.releaseGroup || "");
      setEditedTmdbID(data.meta.tmdbID || 0);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取任务详情失败",
        description,
        variant: "destructive",
      });
      onOpenChange(false);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && taskID) {
      fetchTaskDetail();
    }
  }, [open, taskID]);

  // 处理文件变更
  const handleFileChange = (
    fileName: string,
    updates: Partial<TorrentFile>
  ) => {
    if (!task) return;

    const updatedFiles = task.torrent.files.map((file) =>
      file.fileName === fileName ? { ...file, ...updates } : file
    );

    setTask({
      ...task,
      torrent: {
        ...task.torrent,
        files: updatedFiles,
      },
    });
  };

  // 处理初始化任务
  const handleInit = async () => {
    if (!taskID) return;

    try {
      setUpdating(true);
      await magnetAPI.initTask(taskID, editedTmdbID);
      toast({
        title: "解析成功",
        description: "种子解析已完成",
      });
      onUpdated?.();
      onOpenChange(false);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "解析失败",
        description,
        variant: "destructive",
      });
    } finally {
      setUpdating(false);
    }
  };

  // 处理TMDB元数据变化
  const handleTMDBMetaChange = (meta: Meta) => {
    setEditedTmdbID(meta.tmdbID);
    if (task) {
      setTask((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          meta: { ...prev.meta, ...meta },
        };
      });
    }
  };

  // 处理更新任务
  const handleUpdate = async (continueDownload?: boolean) => {
    if (!taskID || !task) return;

    // 验证 TMDB ID
    if (editedTmdbID && isNaN(editedTmdbID) && editedTmdbID === 0) {
      toast({
        title: "验证失败",
        description: "请填入正确的 TMDB ID",
        variant: "destructive",
      });
      return;
    }

    try {
      setUpdating(true);
      await magnetAPI.updateTask(taskID, {
        tmdbID: editedTmdbID,
        releaseGroup: editedReleaseGroup,
        torrent: task.torrent,
        continueDownload,
      });
      toast({
        title: "更新成功",
        description: continueDownload ? "任务已确认，开始下载" : "任务已更新",
      });
      onUpdated?.();
      onOpenChange(false);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "更新失败",
        description,
        variant: "destructive",
      });
    } finally {
      setUpdating(false);
    }
  };

  // 转移文件
  const handleTransfer = async () => {
    if (!task) return;

    try {
      setTransferring(true);
      await subscriptionAPI.transferTorrent(task.torrent.hash);
      toast({
        title: "转移成功",
        description: "文件转移已完成",
      });
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "转移失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      onUpdated?.();
      fetchTaskDetail(); // 刷新详情
      setTransferring(false);
    }
  };

  // 获取按钮文本和操作
  const getActionButton = () => {
    if (!task) return null;

    if (task.status === TaskStatusSet.WaitingForParsing) {
      return {
        text: "解析",
        action: handleInit,
        variant: "default" as const,
      };
    }

    if (task.status === TaskStatusSet.WaitingForConfirmation) {
      return {
        text: "确认",
        action: () => handleUpdate(true),
        variant: "default" as const,
      };
    }

    return {
      text: "更新",
      action: () => handleUpdate(),
      variant: "default" as const,
    };
  };

  const status = task
    ? getTaskStatus(task.status, task.downloadStatus)
    : { text: "", color: "", Icon: Activity };
  const actionButton = getActionButton();
  const isDownloading =
    task?.status === TaskStatusSet.InitSuccess &&
    task?.downloadStatus === TorrentStatusSet.Downloading;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={`${
          isMobile
            ? "w-[92vw] max-w-[92vw] h-[85vh] max-h-[85vh] p-4"
            : "max-w-4xl max-h-[90vh] p-6"
        } overflow-y-auto scrollbar-hide rounded-xl border-primary/20 bg-card/95 backdrop-blur-md`}
      >
        <DialogHeader>
          <DialogTitle
            className={`flex items-center gap-2 ${
              isMobile ? "text-lg" : "text-xl"
            }`}
          >
            {task && renderTypeIcon(task.downloadType)}
            <span className="anime-gradient-text">磁力任务详情</span>
          </DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : task ? (
          <div className={isMobile ? "space-y-4" : "space-y-6"}>
            {/* 第一部分：基础信息 */}
            <div className={isMobile ? "space-y-3" : "space-y-4"}>
              <h3
                className={`font-semibold flex items-center gap-2 ${
                  isMobile ? "text-base" : "text-lg"
                }`}
              >
                <Film
                  className={
                    isMobile ? "w-4 h-4 text-primary" : "w-5 h-5 text-primary"
                  }
                />
                <span className="anime-gradient-text">基础信息</span>
              </h3>

              {/* 种子名称卡片 */}
              <Card className="border-primary/20 bg-gradient-to-br from-primary/5 to-transparent anime-card">
                <CardContent className={isMobile ? "p-3" : "p-4"}>
                  <div className="flex items-start gap-2.5">
                    <div
                      className={`rounded-lg bg-primary/10 ${
                        isMobile ? "p-1.5" : "p-2"
                      }`}
                    >
                      {renderTypeIcon(task.downloadType)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p
                        className={`text-muted-foreground mb-1 ${
                          isMobile ? "text-[10px]" : "text-xs"
                        }`}
                      >
                        种子名称
                      </p>
                      <p
                        className={`font-medium break-all ${
                          isMobile ? "text-xs" : "text-sm"
                        }`}
                      >
                        {task.torrent.name}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* 信息网格 */}
              <div
                className={`grid grid-cols-1 md:grid-cols-2 ${
                  isMobile ? "gap-2.5" : "gap-3"
                }`}
              >
                {/* 字幕组 - 可编辑 */}
                <Card className="border-primary/10 hover:border-primary/30 transition-colors anime-card">
                  <CardContent className={isMobile ? "p-3" : "p-4"}>
                    <div
                      className={`flex items-start ${
                        isMobile ? "gap-2" : "gap-3"
                      }`}
                    >
                      <div
                        className={`rounded-lg bg-blue-500/10 flex-shrink-0 ${
                          isMobile ? "p-1.5" : "p-2"
                        }`}
                      >
                        <Users
                          className={
                            isMobile
                              ? "w-3.5 h-3.5 text-blue-500"
                              : "w-4 h-4 text-blue-500"
                          }
                        />
                      </div>
                      <div className="flex-1 min-w-0 space-y-1.5">
                        <Label
                          htmlFor="release-group"
                          className={`text-muted-foreground ${
                            isMobile ? "text-[10px]" : "text-xs"
                          }`}
                        >
                          字幕组
                        </Label>
                        <Input
                          id="release-group"
                          value={editedReleaseGroup}
                          onChange={(e) =>
                            setEditedReleaseGroup(e.target.value)
                          }
                          placeholder="请输入字幕组名称"
                          className={`rounded-lg border-blue-500/20 focus:border-blue-500 focus:ring-blue-500 ${
                            isMobile ? "h-7 text-xs" : "h-8"
                          }`}
                        />
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* 下载类型 */}
                <Card className="border-primary/10 hover:border-primary/30 transition-colors anime-card">
                  <CardContent className={isMobile ? "p-3" : "p-4"}>
                    <div
                      className={`flex items-center ${
                        isMobile ? "gap-2" : "gap-3"
                      }`}
                    >
                      <div
                        className={`rounded-lg bg-pink-500/10 ${
                          isMobile ? "p-1.5" : "p-2"
                        }`}
                      >
                        <Clapperboard
                          className={
                            isMobile
                              ? "w-3.5 h-3.5 text-pink-500"
                              : "w-4 h-4 text-pink-500"
                          }
                        />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p
                          className={`text-muted-foreground mb-1 ${
                            isMobile ? "text-[10px]" : "text-xs"
                          }`}
                        >
                          下载类型
                        </p>
                        <Badge
                          variant="outline"
                          className={`border-pink-500/30 text-pink-700 dark:text-pink-300 ${
                            isMobile ? "text-xs px-1.5 py-0" : ""
                          }`}
                        >
                          {DownloadTypeLabels[task.downloadType]}
                        </Badge>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* 种子大小 */}
                <Card className="border-primary/10 hover:border-primary/30 transition-colors anime-card">
                  <CardContent className={isMobile ? "p-3" : "p-4"}>
                    <div
                      className={`flex items-center ${
                        isMobile ? "gap-2" : "gap-3"
                      }`}
                    >
                      <div
                        className={`rounded-lg bg-green-500/10 ${
                          isMobile ? "p-1.5" : "p-2"
                        }`}
                      >
                        <HardDrive
                          className={
                            isMobile
                              ? "w-3.5 h-3.5 text-green-500"
                              : "w-4 h-4 text-green-500"
                          }
                        />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p
                          className={`text-muted-foreground mb-1 ${
                            isMobile ? "text-[10px]" : "text-xs"
                          }`}
                        >
                          种子大小
                        </p>
                        <p
                          className={`font-semibold text-green-700 dark:text-green-400 ${
                            isMobile ? "text-xs" : "text-sm"
                          }`}
                        >
                          {formatFileSize(task.torrent.size)}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* TMDB ID - 可编辑 */}
                <Card className="border-primary/10 hover:border-primary/30 transition-colors anime-card">
                  <CardContent className={isMobile ? "p-3" : "p-4"}>
                    <div
                      className={`flex items-start ${
                        isMobile ? "gap-2" : "gap-3"
                      }`}
                    >
                      <div
                        className={`rounded-lg bg-amber-500/10 flex-shrink-0 ${
                          isMobile ? "p-1.5" : "p-2"
                        }`}
                      >
                        <Database
                          className={
                            isMobile
                              ? "w-3.5 h-3.5 text-amber-500"
                              : "w-4 h-4 text-amber-500"
                          }
                        />
                      </div>
                      <div className="flex-1 min-w-0">
                        <TMDBInput
                          type={task.downloadType}
                          value={editedTmdbID}
                          onTMDBIDChange={(tmdbID) => setEditedTmdbID(tmdbID)}
                          onMetaChange={handleTMDBMetaChange}
                          label="TMDB ID"
                          placeholder="输入 TMDB ID 或点击搜索"
                        />
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* 元数据详情（如果有TMDB ID） */}
              {(editedTmdbID > 0 || task.meta.tmdbID > 0) && (
                <Card className="border-primary/20 bg-gradient-to-br from-amber-500/5 via-transparent to-transparent anime-card">
                  <CardContent className={isMobile ? "p-3" : "p-4"}>
                    <div className={isMobile ? "space-y-2.5" : "space-y-3"}>
                      <div className="flex items-center gap-2">
                        <div
                          className={`rounded-md bg-amber-500/10 ${
                            isMobile ? "p-1" : "p-1.5"
                          }`}
                        >
                          <Film
                            className={
                              isMobile
                                ? "w-3 h-3 text-amber-500"
                                : "w-3.5 h-3.5 text-amber-500"
                            }
                          />
                        </div>
                        <span
                          className={`font-medium text-amber-700 dark:text-amber-400 ${
                            isMobile ? "text-[10px]" : "text-xs"
                          }`}
                        >
                          TMDB 元数据
                        </span>
                      </div>
                      <div
                        className={`grid grid-cols-1 sm:grid-cols-2 ${
                          isMobile ? "gap-2 pl-1" : "gap-3 pl-2"
                        }`}
                      >
                        <div className="flex items-center gap-2">
                          <Tv
                            className={
                              isMobile
                                ? "w-3 h-3 text-muted-foreground"
                                : "w-3.5 h-3.5 text-muted-foreground"
                            }
                          />
                          <div className="flex-1 min-w-0">
                            <p
                              className={`text-muted-foreground ${
                                isMobile ? "text-[10px]" : "text-xs"
                              }`}
                            >
                              中文名
                            </p>
                            <p
                              className={`font-medium truncate ${
                                isMobile ? "text-xs" : "text-sm"
                              }`}
                            >
                              {task.meta.chineseName || "未知"}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Calendar
                            className={
                              isMobile
                                ? "w-3 h-3 text-muted-foreground"
                                : "w-3.5 h-3.5 text-muted-foreground"
                            }
                          />
                          <div className="flex-1 min-w-0">
                            <p
                              className={`text-muted-foreground ${
                                isMobile ? "text-[10px]" : "text-xs"
                              }`}
                            >
                              年份
                            </p>
                            <p
                              className={`font-medium ${
                                isMobile ? "text-xs" : "text-sm"
                              }`}
                            >
                              {task.meta.year || "未知"}
                            </p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>

            <Separator />

            {/* 第二部分：状态信息 */}
            <div className={isMobile ? "space-y-3" : "space-y-4"}>
              <h3
                className={`font-semibold flex items-center gap-2 ${
                  isMobile ? "text-base" : "text-lg"
                }`}
              >
                <Activity
                  className={
                    isMobile ? "w-4 h-4 text-primary" : "w-5 h-5 text-primary"
                  }
                />
                <span className="anime-gradient-text">状态信息</span>
              </h3>

              {/* 当前状态卡片 */}
              <Card className="border-primary/10 anime-card">
                <CardContent className={isMobile ? "p-3" : "p-4"}>
                  <div
                    className={`flex items-center ${
                      isMobile ? "gap-2" : "gap-3"
                    }`}
                  >
                    <div
                      className={`rounded-lg ${
                        status.color
                      } bg-opacity-10 flex-shrink-0 ${
                        isMobile ? "p-1.5" : "p-2"
                      }`}
                    >
                      <status.Icon
                        className={`${status.color.replace("bg-", "text-")} ${
                          isMobile ? "w-3.5 h-3.5" : "w-4 h-4"
                        }`}
                      />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p
                        className={`text-muted-foreground mb-1 ${
                          isMobile ? "text-[10px]" : "text-xs"
                        }`}
                      >
                        当前状态
                      </p>
                      <Badge
                        className={`${
                          status.color
                        } text-white w-fit flex items-center gap-1 ${
                          isMobile ? "text-xs px-1.5 py-0" : ""
                        }`}
                      >
                        {status.text}
                      </Badge>
                    </div>
                    {/* 转移按钮 */}
                    {task && canTransfer(task) && (
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              size={isMobile ? "sm" : "sm"}
                              onClick={handleTransfer}
                              disabled={updating || transferring}
                              className={`rounded-lg flex-shrink-0 ${
                                isMobile ? "text-xs px-2 h-7" : ""
                              }`}
                            >
                              {transferring && (
                                <Loader2
                                  className={`animate-spin ${
                                    isMobile ? "mr-1 h-3 w-3" : "mr-2 h-4 w-4"
                                  }`}
                                />
                              )}
                              {!transferring && (
                                <ArrowUpFromLine
                                  className={
                                    isMobile ? "mr-1 h-3 w-3" : "mr-2 h-4 w-4"
                                  }
                                />
                              )}
                              转移文件
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>转移文件到媒体库</TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* 下载进度卡片（仅下载中时显示） */}
              {isDownloading && (
                <Card className="border-primary/20 bg-gradient-to-br from-blue-500/5 to-transparent anime-card">
                  <CardContent
                    className={`${
                      isMobile ? "p-3 space-y-3" : "p-4 space-y-4"
                    }`}
                  >
                    {/* 下载速度 */}
                    <div
                      className={`flex items-center ${
                        isMobile ? "gap-2" : "gap-3"
                      }`}
                    >
                      <div
                        className={`rounded-lg bg-blue-500/10 ${
                          isMobile ? "p-1.5" : "p-2"
                        }`}
                      >
                        <Download
                          className={
                            isMobile
                              ? "w-3.5 h-3.5 text-blue-500"
                              : "w-4 h-4 text-blue-500"
                          }
                        />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p
                          className={`text-muted-foreground mb-1 ${
                            isMobile ? "text-[10px]" : "text-xs"
                          }`}
                        >
                          下载速度
                        </p>
                        <p
                          className={`font-semibold text-blue-700 dark:text-blue-400 ${
                            isMobile ? "text-xs" : "text-sm"
                          }`}
                        >
                          {formatFileSize(task.downloadSpeed)}/s
                        </p>
                      </div>
                    </div>

                    {/* 下载进度条 */}
                    <div className="space-y-2">
                      <div
                        className={`flex items-center justify-between ${
                          isMobile ? "text-[10px]" : "text-xs"
                        }`}
                      >
                        <span className="text-muted-foreground">下载进度</span>
                        <span className="font-medium text-blue-700 dark:text-blue-400">
                          {Math.round(task.progress * 100)}%
                        </span>
                      </div>
                      <Progress
                        value={task.progress * 100}
                        className={`bg-blue-500/10 ${
                          isMobile ? "h-1.5" : "h-2"
                        }`}
                      />
                      <div
                        className={`flex items-center justify-between text-muted-foreground ${
                          isMobile ? "text-[10px]" : "text-xs"
                        }`}
                      >
                        <span>
                          {formatFileSize(task.torrent.size * task.progress)}
                        </span>
                        <span>{formatFileSize(task.torrent.size)}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* 错误详情 */}
              {task.downloadStatusDetail && (
                <Alert
                  variant="destructive"
                  className="border-red-500/20 bg-red-500/5 anime-card"
                >
                  <Activity className={isMobile ? "h-3.5 w-3.5" : "h-4 w-4"} />
                  <AlertDescription
                    className={isMobile ? "text-xs" : "text-sm"}
                  >
                    <span className="font-medium block mb-1">状态详情</span>
                    {task.downloadStatusDetail}
                  </AlertDescription>
                </Alert>
              )}
            </div>

            <Separator />

            {/* 第三部分：文件信息 */}
            <div className={isMobile ? "space-y-3" : "space-y-4"}>
              <h3
                className={`font-semibold flex items-center gap-2 ${
                  isMobile ? "text-base" : "text-lg"
                }`}
              >
                <FolderTree
                  className={
                    isMobile ? "w-4 h-4 text-primary" : "w-5 h-5 text-primary"
                  }
                />
                <span className="anime-gradient-text">文件信息</span>
                <TooltipProvider>
                  <HybridTooltip>
                    <HybridTooltipTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-5 w-5 rounded-full"
                      >
                        <Info className="h-3.5 w-3.5 text-muted-foreground" />
                      </Button>
                    </HybridTooltipTrigger>
                    <HybridTooltipContent
                      className="max-w-xs whitespace-normal"
                      align="start"
                      side="top"
                    >
                      <div className="space-y-2 text-sm">
                        <p>• 点击文件左侧图标决定文件是否会转移到媒体库</p>
                        {task.downloadType === DownloadTypeSet.TV && (
                          <p>
                            • 点击文件的 SxxExx 标签可以修改识别的季度和集数信息
                          </p>
                        )}
                      </div>
                    </HybridTooltipContent>
                  </HybridTooltip>
                </TooltipProvider>
              </h3>
              <FileTree
                files={task.torrent.files || []}
                downloadType={task.downloadType}
                taskID={task.taskID}
                onFileChange={handleFileChange}
                onSubtitleTransferSuccess={fetchTaskDetail}
              />
            </div>
          </div>
        ) : null}

        <DialogFooter
          className={`${
            isMobile ? "gap-2 flex-col-reverse" : "gap-2 sm:gap-0"
          }`}
        >
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={updating || transferring}
            className={`rounded-xl ${isMobile ? "w-full" : ""}`}
          >
            取消
          </Button>
          {actionButton && (
            <Button
              variant={actionButton.variant}
              onClick={actionButton.action}
              disabled={updating || transferring}
              className={`rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button ${
                isMobile ? "w-full" : ""
              }`}
            >
              {updating && (
                <Loader2
                  className={`animate-spin ${
                    isMobile ? "mr-1.5 h-3.5 w-3.5" : "mr-2 h-4 w-4"
                  }`}
                />
              )}
              {actionButton.text}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
