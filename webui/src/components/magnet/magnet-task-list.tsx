import { useState, useEffect, useCallback, useRef } from "react";
import { Trash2, ArrowUpFromLine, Loader2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Checkbox } from "@/components/ui/checkbox";
import magnetAPI, { DownloadTask, TaskStatusSet } from "@/api/magnet";
import subscriptionAPI, { TorrentStatusSet } from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { useMobile } from "@/hooks/useMobile";
import { extractErrorMessage } from "@/utils/error";
import { MagnetDetailDialog } from "./magnet-detail-dialog";
import { renderTypeIcon, getTaskStatus, canTransfer } from "./magnet-utils";

// 格式化文件大小
function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}

interface MagnetTaskListProps {
  refresh: boolean;
  onRefreshComplete?: () => void;
  pauseRefresh?: boolean;
  openTaskID?: string | null;
}

export function MagnetTaskList({
  refresh,
  onRefreshComplete,
  pauseRefresh = false,
  openTaskID,
}: MagnetTaskListProps) {
  const { toast } = useToast();
  const isMobile = useMobile();
  const [tasks, setTasks] = useState<DownloadTask[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [deleteTaskInfo, setDeleteTaskInfo] = useState<{
    taskID: string;
    taskName: string;
  } | null>(null);
  const [showDetailDialog, setShowDetailDialog] = useState(false);
  const [selectedTaskID, setSelectedTaskID] = useState<string | null>(null);
  const [refreshInterval, setRefreshInterval] = useState<number>(10000);
  const [isUpdating, setIsUpdating] = useState(false);

  // 添加定时器引用
  const refreshTimerRef = useRef<NodeJS.Timeout | null>(null);

  // 清理定时器的函数
  const clearRefreshTimer = () => {
    if (refreshTimerRef.current) {
      clearInterval(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }
  };

  // 启动定时器的函数
  const startRefreshTimer = (interval?: number) => {
    clearRefreshTimer();
    const actualInterval = interval ?? refreshInterval;
    refreshTimerRef.current = setInterval(() => {
      updateTasks();
    }, actualInterval);
  };

  // 检查是否有需要刷新的任务（下载中或暂停）
  const hasTasksNeedingRefresh = (tasks: DownloadTask[]) => {
    return tasks.some(
      (task) =>
        task.status === TaskStatusSet.InitSuccess &&
        (task.downloadStatus === TorrentStatusSet.Downloading ||
          task.downloadStatus === TorrentStatusSet.DownloadPaused)
    );
  };

  // 更新刷新间隔
  const updateRefreshInterval = (tasks: DownloadTask[]) => {
    const needsRefresh = hasTasksNeedingRefresh(tasks);

    if (!needsRefresh) {
      // 没有需要刷新的任务，停止定时器
      clearRefreshTimer();
      return;
    }

    // 有下载中的任务用1秒间隔，否则用10秒
    const hasDownloading = tasks.some(
      (task) =>
        task.status === TaskStatusSet.InitSuccess &&
        task.downloadStatus === TorrentStatusSet.Downloading
    );
    const newInterval = hasDownloading ? 1000 : 10000;

    if (newInterval !== refreshInterval) {
      setRefreshInterval(newInterval);
    }

    // 如果定时器没有运行，启动它
    if (!refreshTimerRef.current) {
      startRefreshTimer(newInterval);
    } else if (newInterval !== refreshInterval) {
      // 间隔变化时重置定时器
      startRefreshTimer(newInterval);
    }
  };

  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      const response = await magnetAPI.listTasks({
        page: currentPage,
        page_size: pageSize,
      });
      const data = Array.isArray(response.tasks) ? response.tasks : [];
      setTasks(data);
      setTotal(response.total);
      // 检查并更新刷新间隔
      updateRefreshInterval(data);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取磁力任务失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
      onRefreshComplete?.();
    }
  }, [currentPage, pageSize, toast, onRefreshComplete]);

  // 修改 updateTasks 用于定时刷新
  const updateTasks = async () => {
    if (isUpdating) return;

    try {
      setIsUpdating(true);
      const response = await magnetAPI.listTasks({
        page: currentPage,
        page_size: pageSize,
      });
      const data = Array.isArray(response.tasks) ? response.tasks : [];

      setTasks((prevTasks) => {
        // 如果是首次加载数据，直接返回新数据
        if (prevTasks.length === 0) {
          updateRefreshInterval(data);
          return data;
        }

        // 否则，保持现有的 DOM 结构，只更新变化的数据
        const newTasks = data.map((newTask) => {
          const existingTask = prevTasks.find(
            (t) => t.taskID === newTask.taskID
          );
          if (existingTask) {
            return {
              ...existingTask,
              status: newTask.status,
              downloadStatus: newTask.downloadStatus,
              progress: newTask.progress,
              downloadSpeed: newTask.downloadSpeed,
            };
          }
          return newTask;
        });

        // 检查并更新刷新间隔
        updateRefreshInterval(newTasks);
        return newTasks;
      });

      setTotal(response.total);
    } catch (error) {
      console.error("Failed to update tasks:", error);
    } finally {
      setIsUpdating(false);
    }
  };

  // 删除任务
  const handleDelete = async (taskID: string, deleteFiles: boolean = false) => {
    try {
      await magnetAPI.deleteTask(taskID, deleteFiles);
      toast({
        title: "删除成功",
        description: `磁力任务已删除${deleteFiles ? "（包含相关文件）" : ""}`,
      });
      fetchTasks();
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "删除失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setShowDeleteDialog(false);
      setDeleteTaskInfo(null);
    }
  };

  // 显示删除确认对话框
  const showDeleteConfirmation = (taskID: string, taskName: string) => {
    setDeleteTaskInfo({ taskID, taskName });
    setShowDeleteDialog(true);
  };

  // 显示任务详情
  const showTaskDetail = (taskID: string) => {
    setSelectedTaskID(taskID);
    setShowDetailDialog(true);
  };

  // 转移文件
  const handleTransfer = async (hash: string) => {
    try {
      await subscriptionAPI.transferTorrent(hash);
      toast({
        title: "转移成功",
        description: "文件转移已完成",
      });
      fetchTasks();
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "转移失败",
        description: description,
        variant: "destructive",
      });
    }
  };

  // 在组件卸载时清理定时器
  useEffect(() => {
    return () => clearRefreshTimer();
  }, []);

  // 启动自动刷新定时器
  useEffect(() => {
    clearRefreshTimer();

    // 立即加载一次数据，fetchTasks 会根据任务状态决定是否启动定时器
    fetchTasks();

    return () => clearRefreshTimer();
  }, [currentPage, pageSize]);

  useEffect(() => {
    if (refresh) {
      fetchTasks();
    }
  }, [refresh, fetchTasks]);

  // 监听 pauseRefresh、对话框状态变化，控制定时器
  useEffect(() => {
    // 当添加对话框、详情对话框或删除对话框打开时，暂停定时器
    const shouldPause = pauseRefresh || showDetailDialog || showDeleteDialog;

    if (shouldPause) {
      // 暂停时停止定时器
      clearRefreshTimer();
    } else {
      // 恢复时根据当前任务状态决定是否启动定时器
      if (tasks.length > 0) {
        updateRefreshInterval(tasks);
      }
    }
  }, [pauseRefresh, showDetailDialog, showDeleteDialog]);

  // 监听 openTaskID 变化，自动打开任务详情
  useEffect(() => {
    if (openTaskID) {
      setSelectedTaskID(openTaskID);
      setShowDetailDialog(true);
    }
  }, [openTaskID]);

  // 渲染分页
  const totalPages = Math.ceil(total / pageSize);
  const renderPagination = () => {
    const pages: number[] = [];
    const maxVisiblePages = 5;
    let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2));
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

    if (endPage - startPage < maxVisiblePages - 1) {
      startPage = Math.max(1, endPage - maxVisiblePages + 1);
    }

    for (let i = startPage; i <= endPage; i++) {
      pages.push(i);
    }

    return (
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              className={
                currentPage === 1
                  ? "pointer-events-none opacity-50"
                  : "cursor-pointer"
              }
            />
          </PaginationItem>

          {startPage > 1 && (
            <>
              <PaginationItem>
                <PaginationLink
                  onClick={() => setCurrentPage(1)}
                  className="cursor-pointer"
                >
                  1
                </PaginationLink>
              </PaginationItem>
              {startPage > 2 && (
                <PaginationItem>
                  <PaginationEllipsis />
                </PaginationItem>
              )}
            </>
          )}

          {pages.map((page) => (
            <PaginationItem key={page}>
              <PaginationLink
                onClick={() => setCurrentPage(page)}
                isActive={currentPage === page}
                className="cursor-pointer"
              >
                {page}
              </PaginationLink>
            </PaginationItem>
          ))}

          {endPage < totalPages && (
            <>
              {endPage < totalPages - 1 && (
                <PaginationItem>
                  <PaginationEllipsis />
                </PaginationItem>
              )}
              <PaginationItem>
                <PaginationLink
                  onClick={() => setCurrentPage(totalPages)}
                  className="cursor-pointer"
                >
                  {totalPages}
                </PaginationLink>
              </PaginationItem>
            </>
          )}

          <PaginationItem>
            <PaginationNext
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              className={
                currentPage === totalPages
                  ? "pointer-events-none opacity-50"
                  : "cursor-pointer"
              }
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    );
  };

  if (loading && tasks.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    );
  }

  // 判断是否显示下载进度
  const isDownloading = (task: DownloadTask) => {
    return (
      task.status === TaskStatusSet.InitSuccess &&
      task.downloadStatus === TorrentStatusSet.Downloading &&
      task.progress > 0
    );
  };

  const renderContent = () => (
    <div className="space-y-4 flex flex-col h-full">
      {/* 任务列表 */}
      <div className="flex-1 space-y-3 overflow-auto">
        {tasks.map((task) => (
          <Card
            key={task.taskID}
            className="anime-card hover:shadow-lg transition-shadow cursor-pointer"
            onClick={() => showTaskDetail(task.taskID)}
          >
            <CardContent className={isMobile ? "p-3" : "p-4"}>
              <div className="flex items-start gap-3">
                {/* 种子信息 */}
                <div className="flex-1 min-w-0 space-y-2">
                  {/* 种子名称 + 图标 */}
                  <div className="flex items-start gap-2">
                    <div className="flex-shrink-0 mt-0.5">
                      {renderTypeIcon(task.downloadType)}
                    </div>
                    <div
                      className={`font-semibold flex-1 min-w-0 ${
                        isMobile ? "text-sm line-clamp-2" : "text-base"
                      }`}
                    >
                      {task.torrent.name}
                    </div>
                  </div>

                  {/* 状态标签 + 元数据标签 */}
                  <div className="flex items-center gap-1.5 flex-wrap">
                    {(() => {
                      const { color, text, Icon } = getTaskStatus(
                        task.status,
                        task.downloadStatus
                      );

                      const badge = (
                        <Badge
                          className={`${color} text-white whitespace-nowrap flex items-center gap-1 ${
                            isMobile ? "text-xs px-1.5 py-0" : ""
                          }`}
                        >
                          <Icon
                            className={isMobile ? "w-2.5 h-2.5" : "w-3 h-3"}
                          />
                          {text}
                        </Badge>
                      );

                      // 如果是转移错误且有错误详情，使用 Tooltip 包裹
                      if (
                        task.status === TaskStatusSet.InitSuccess &&
                        task.downloadStatus ===
                          TorrentStatusSet.TransferredError &&
                        task.downloadStatusDetail
                      ) {
                        return (
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>{badge}</TooltipTrigger>
                              <TooltipContent className="max-w-md">
                                {task.downloadStatusDetail}
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        );
                      }

                      return badge;
                    })()}
                    {task.meta?.chineseName && (
                      <Badge
                        variant="secondary"
                        className={`whitespace-nowrap ${
                          isMobile ? "text-xs px-1.5 py-0" : ""
                        }`}
                      >
                        {task.meta.chineseName}
                      </Badge>
                    )}
                    {task.meta?.year && (
                      <Badge
                        variant="secondary"
                        className={`whitespace-nowrap ${
                          isMobile ? "text-xs px-1.5 py-0" : ""
                        }`}
                      >
                        {task.meta.year}
                      </Badge>
                    )}
                    {task.meta?.releaseGroup && (
                      <Badge
                        variant="secondary"
                        className={`whitespace-nowrap ${
                          isMobile ? "text-xs px-1.5 py-0" : ""
                        }`}
                      >
                        {task.meta.releaseGroup}
                      </Badge>
                    )}
                  </div>

                  {/* 下载进度条（仅下载中时显示） */}
                  {isDownloading(task) && (
                    <>
                      <Progress value={task.progress * 100} className="h-1.5" />
                      <div
                        className={`flex items-center justify-between text-muted-foreground ${
                          isMobile ? "text-[10px]" : "text-xs"
                        }`}
                      >
                        <span>{Math.round(task.progress * 100)}%</span>
                        {isMobile ? (
                          <span>{formatFileSize(task.downloadSpeed)}/s</span>
                        ) : (
                          <span>
                            {formatFileSize(task.downloadSpeed)}/s{" "}
                            {formatFileSize(task.torrent.size * task.progress)}{" "}
                            / {formatFileSize(task.torrent.size)}
                          </span>
                        )}
                      </div>
                    </>
                  )}

                  {/* 种子大小 */}
                  {!isDownloading(task) && (
                    <div
                      className={`text-muted-foreground ${
                        isMobile ? "text-[10px]" : "text-xs"
                      }`}
                    >
                      {task.torrent.size > 0
                        ? formatFileSize(task.torrent.size)
                        : "未知大小"}
                    </div>
                  )}
                </div>

                {/* 操作按钮 */}
                <div className="flex-shrink-0 flex flex-col gap-1.5">
                  {/* 转移按钮 */}
                  {canTransfer(task) && (
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            size="icon"
                            variant="ghost"
                            className={`rounded-full ${
                              isMobile ? "h-7 w-7" : "h-8 w-8"
                            }`}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleTransfer(task.torrent.hash);
                            }}
                          >
                            <ArrowUpFromLine
                              className={isMobile ? "h-3.5 w-3.5" : "h-4 w-4"}
                            />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>转移文件</TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  )}

                  {/* 删除按钮 */}
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          size="icon"
                          variant="ghost"
                          className={`rounded-full hover:bg-destructive/10 ${
                            isMobile ? "h-7 w-7" : "h-8 w-8"
                          }`}
                          onClick={(e) => {
                            e.stopPropagation();
                            showDeleteConfirmation(
                              task.taskID,
                              task.torrent.name
                            );
                          }}
                        >
                          <Trash2
                            className={`text-muted-foreground hover:text-destructive ${
                              isMobile ? "h-3.5 w-3.5" : "h-4 w-4"
                            }`}
                          />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>删除任务</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* 分页控件 */}
      <div
        className={`flex-shrink-0 flex items-center border-t ${
          isMobile ? "flex-col gap-3 py-3" : "flex-row justify-between py-4"
        }`}
      >
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">每页显示</span>
          <Select
            value={pageSize.toString()}
            onValueChange={(value) => {
              setPageSize(Number(value));
              setCurrentPage(1);
            }}
          >
            <SelectTrigger className="w-20">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="10">10</SelectItem>
              <SelectItem value="20">20</SelectItem>
              <SelectItem value="50">50</SelectItem>
            </SelectContent>
          </Select>
          <span className="text-sm text-muted-foreground">条</span>
        </div>

        {totalPages > 1 && <div>{renderPagination()}</div>}
      </div>
    </div>
  );

  // 删除确认对话框
  const DeleteConfirmDialog = () => {
    // 使用 ref 替代 state 避免重渲染
    const deleteFilesRef = useRef(false);
    // 添加一个状态钩子来强制重新渲染
    const [, forceUpdate] = useState({});

    // 强制组件更新的辅助函数
    const forceRender = () => forceUpdate({});

    return (
      <AlertDialog
        open={showDeleteDialog}
        onOpenChange={(open) => {
          if (!loading) {
            setShowDeleteDialog(open);
            if (!open) {
              setDeleteTaskInfo(null);
            }
          }
        }}
      >
        <AlertDialogContent
          onClick={(e) => e.stopPropagation()}
          className="w-[95vw] max-w-md p-4 md:p-6"
        >
          <AlertDialogHeader>
            <AlertDialogTitle className="text-lg md:text-xl">
              确认删除任务
            </AlertDialogTitle>
            <AlertDialogDescription className="text-sm mt-2">
              你确定要删除磁力任务 <strong>{deleteTaskInfo?.taskName}</strong>{" "}
              吗？此操作无法撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>

          <div
            className="flex items-center space-x-2 py-4"
            onClick={(e) => e.stopPropagation()}
          >
            <Checkbox
              id="delete-files-option"
              checked={deleteFilesRef.current}
              onCheckedChange={(checked) => {
                deleteFilesRef.current = checked as boolean;
                // 强制重新渲染复选框
                forceRender();
              }}
            />
            <div
              className="cursor-pointer select-none text-sm"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                deleteFilesRef.current = !deleteFilesRef.current;
                forceRender();
              }}
            >
              同时删除已下载的文件及其相关文件
            </div>
          </div>

          <AlertDialogFooter className="flex-col space-y-2 sm:space-y-0 sm:flex-row">
            <AlertDialogCancel className="mt-2 sm:mt-0" disabled={loading}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={() =>
                deleteTaskInfo &&
                handleDelete(deleteTaskInfo.taskID, deleteFilesRef.current)
              }
              disabled={loading}
              className="bg-destructive hover:bg-destructive/90"
            >
              {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    );
  };

  return (
    <>
      {renderContent()}
      <DeleteConfirmDialog />
      <MagnetDetailDialog
        taskID={selectedTaskID}
        open={showDetailDialog}
        onOpenChange={setShowDetailDialog}
        onUpdated={fetchTasks}
      />
    </>
  );
}
