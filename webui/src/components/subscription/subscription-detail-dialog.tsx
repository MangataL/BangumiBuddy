import type React from "react";
import { useState, useEffect, useRef, useMemo } from "react";
import {
  X,
  SettingsIcon,
  Eye,
  FileBox,
  Star,
  Save,
  CircleSlash,
  CheckSquare,
  Info,
  Play,
  Trash2,
  Loader2,
  CheckCircle2,
  PauseCircle,
  CircleArrowDown,
  CircleAlert,
  CircleX,
  CircleArrowRight,
  Sparkles,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent } from "@/components/ui/tabs";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Progress } from "@/components/ui/progress";
import { useToast } from "@/hooks/useToast";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import subscriptionAPI, {
  type Bangumi,
  type RSSMatch,
  type TorrentFile,
  ReleaseGroupSubscription,
  type Torrent,
  TorrentStatusSet,
} from "@/api/subscription";
import { MatchInput } from "../common/match-input";
import { EpisodePositionInput } from "./episode-position-input";
import { extractErrorMessage } from "@/utils/error";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
import { Checkbox } from "@/components/ui/checkbox";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { formatDate } from "@/utils/time";
import { useMobile } from "@/hooks/useMobile";
import { TruncatedText } from "@/components/common/truncate-rss-item";
import { torrentCanTransfer } from "@/utils/util";

export interface SubscriptionInit {
  id: string;
  releaseGroups: ReleaseGroupSubscription[];
}

interface SubscriptionDetailDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedSubscriptionID: string;
  releaseGroups: ReleaseGroupSubscription[];
  onClose: () => void;
}

export function SubscriptionDetailDialog({
  open,
  onOpenChange,
  selectedSubscriptionID,
  releaseGroups,
  onClose,
}: SubscriptionDetailDialogProps) {
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState("config");
  const [selectedSubGroup, setSelectedSubGroup] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [bangumiDetails, setBangumiDetails] = useState<Bangumi | null>(null);
  const [rssMatches, setRssMatches] = useState<RSSMatch[]>([]);
  const [subscriptionID, setSubscriptionID] = useState<string>("");
  const [selectedItems, setSelectedItems] = useState<string[]>([]);
  const [configLoaded, setConfigLoaded] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [curReleaseGroups, setCurReleaseGroups] =
    useState<ReleaseGroupSubscription[]>(releaseGroups);
  const [torrents, setTorrents] = useState<Torrent[]>([]);
  const [refreshInterval, setRefreshInterval] = useState<number>(10000);
  const [showFileDetails, setShowFileDetails] = useState(false);
  const [selectedTorrent, setSelectedTorrent] = useState<Torrent | null>(null);
  const isMobile = useMobile();

  // 添加定时器引用
  const refreshTimerRef = useRef<NodeJS.Timeout | null>(null);

  // 状态图标的 Tooltip 组件
  const StatusTooltip = ({
    children,
    content,
    className = "w-8 h-8",
  }: {
    children: React.ReactNode;
    content: string;
    className?: string;
  }) => (
    <TooltipProvider>
      <HybridTooltip>
        <HybridTooltipTrigger>
          <div className={cn("flex items-center justify-center", className)}>
            {children}
          </div>
        </HybridTooltipTrigger>
        <HybridTooltipContent className="max-w-[200px] whitespace-normal break-words">
          {content}
        </HybridTooltipContent>
      </HybridTooltip>
    </TooltipProvider>
  );

  // 转移按钮组件
  const TransferButton = ({ hash }: { hash: string }) => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            size="icon"
            variant="ghost"
            className="h-8 w-8 rounded-full"
            onClick={() => handleRetryTransfer(hash)}
          >
            <CircleArrowRight className="h-4 w-4" />
          </Button>
        </TooltipTrigger>
        <TooltipContent>转移文件</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );

  // 清理定时器的函数
  const clearRefreshTimer = () => {
    if (refreshTimerRef.current) {
      clearInterval(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }
  };

  // 在组件卸载时清理定时器
  useEffect(() => {
    return () => clearRefreshTimer();
  }, []);

  // 检查是否有正在下载的种子
  const hasDownloadingTorrents = (torrents: Torrent[]) => {
    return torrents.some((t) => t.status === TorrentStatusSet.Downloading);
  };

  // 更新刷新间隔
  const updateRefreshInterval = (torrents: Torrent[]) => {
    const newInterval = hasDownloadingTorrents(torrents) ? 1000 : 10000;
    if (newInterval !== refreshInterval) {
      setRefreshInterval(newInterval);
      // 如果间隔发生变化，重置定时器
      if (refreshTimerRef.current) {
        clearInterval(refreshTimerRef.current);
        refreshTimerRef.current = setInterval(() => {
          updateTorrents();
        }, newInterval);
      }
    }
  };

  // 修改 updateTorrents
  const updateTorrents = async () => {
    if (!subscriptionID || isUpdating) return;

    try {
      setIsUpdating(true);
      const torrentList = await subscriptionAPI.getBangumiTorrents(
        subscriptionID
      );
      setTorrents((prevTorrents) => {
        // 如果是首次加载数据，直接返回新数据
        if (prevTorrents.length === 0) {
          updateRefreshInterval(torrentList);
          return torrentList;
        }

        // 检查是否有任何种子从非Transferred状态变为Transferred状态
        const hasNewTransferred = torrentList.some((newTorrent) => {
          const existingTorrent = prevTorrents.find(
            (t) => t.hash === newTorrent.hash
          );
          return (
            existingTorrent &&
            existingTorrent.status !== TorrentStatusSet.Transferred &&
            newTorrent.status === TorrentStatusSet.Transferred
          );
        });

        // 如果有新的已转移种子，更新番剧信息和字幕组信息
        if (hasNewTransferred) {
          updateReleaseGroups(subscriptionID);
        }

        // 否则，保持现有的 DOM 结构，只更新变化的数据
        const newTorrents = torrentList.map((newTorrent) => {
          const existingTorrent = prevTorrents.find(
            (t) => t.hash === newTorrent.hash
          );
          if (existingTorrent) {
            return {
              ...existingTorrent,
              status: newTorrent.status,
              progress: newTorrent.progress,
              downloadSpeed: newTorrent.downloadSpeed,
              statusDetail: newTorrent.statusDetail,
            };
          }
          return newTorrent;
        });

        // 检查并更新刷新间隔
        updateRefreshInterval(newTorrents);
        return newTorrents;
      });
    } catch (error) {
      console.error("Failed to update torrents:", error);
    } finally {
      setIsUpdating(false);
    }
  };

  // 修改定时器相关的 useEffect
  useEffect(() => {
    clearRefreshTimer();

    // 立即加载一次数据
    loadTabData();

    if (activeTab === "files" && subscriptionID) {
      // 设置定时器，初始使用10秒的间隔，后续会根据状态自动调整
      refreshTimerRef.current = setInterval(() => {
        updateTorrents();
      }, refreshInterval);
    }

    return () => clearRefreshTimer();
  }, [activeTab, subscriptionID]);

  // 当subscription变化时自动选择字幕组和设置subscriptionID
  useEffect(() => {
    if (!open) {
      return;
    }
    if (selectedSubscriptionID) {
      setSubscriptionID(selectedSubscriptionID);
    }
    setCurReleaseGroups(releaseGroups);
    if (releaseGroups && releaseGroups.length > 0) {
      setSelectedSubGroup(
        releaseGroups.find(
          (group) => group.subscriptionID === selectedSubscriptionID
        )?.releaseGroup as string
      );
    } else {
      setSelectedSubGroup(null);
    }
  }, [open]);

  // 内部处理字幕组选择
  const handleReleaseGroupClick = (
    groupName: string,
    subscriptionID: string,
    e: React.MouseEvent
  ) => {
    e.stopPropagation();
    setRssMatches([]);
    setLoading(true);
    setSelectedSubGroup(groupName);
    setSubscriptionID(subscriptionID);
    setConfigLoaded(false);
  };

  // 重置状态函数
  const resetState = () => {
    clearRefreshTimer();
    setActiveTab("config");
    setRssMatches([]);
    setSubscriptionID("");
    setSelectedItems([]);
    setConfigLoaded(false);
    setTorrents([]);
  };

  // 增强版onClose处理函数
  const handleClose = () => {
    resetState();
    onClose();
    onOpenChange(false);
  };

  const fetchBangumiDetails = async (subscriptionID: string) => {
    try {
      const data = await subscriptionAPI.getBangumi(subscriptionID);
      setBangumiDetails(data);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取番剧详情失败",
        description: description,
        variant: "destructive",
      });
    }
  };

  // 修改 loadTabData
  const loadTabData = async () => {
    if (!subscriptionID) return;

    setLoading(true);
    try {
      if (activeTab !== "preview") {
        setSelectedItems([]);
      }
      if (activeTab !== "files") {
        setTorrents([]);
      }
      if (activeTab === "preview") {
        // 获取RSS匹配数据
        const rssMatchData = await subscriptionAPI.getRSSMatch(subscriptionID);
        setRssMatches(rssMatchData);
      } else if (activeTab === "files") {
        // 首次加载时获取种子列表
        await updateTorrents();
      } else {
        if (!configLoaded) {
          fetchBangumiDetails(subscriptionID);
          setConfigLoaded(true);
        }
      }
    } catch (error) {
      toast({
        title: `加载${activeTab === "preview" ? "预览" : "文件"}数据失败`,
        description: String(error),
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const updateReleaseGroups = async (subscriptionID: string) => {
    if (!bangumiDetails) return;
    const bangumiBases = await subscriptionAPI.listBangumisBase({
      name: bangumiDetails.name,
      season: bangumiDetails.season,
    });
    if (bangumiBases.length === 0) {
      handleClose();
      return;
    }
    const bangumiBase = bangumiBases[0];

    const updatedGroup = curReleaseGroups.find(
      (group) => group.subscriptionID === subscriptionID
    );
    if (!updatedGroup) {
      return;
    }

    setCurReleaseGroups(bangumiBase.releaseGroups);

    // 如果当前选中的字幕组就是刚修改的，也需要更新selectedSubGroup确保UI刷新
    if (selectedSubGroup === updatedGroup.releaseGroup) {
      // 短暂设置为null然后重新设置回来，强制刷新
      setSelectedSubGroup(null);
      setTimeout(() => {
        setSelectedSubGroup(updatedGroup.releaseGroup);
      }, 0);
    }
  };

  const handleSaveConfig = async () => {
    if (!bangumiDetails || !subscriptionID || !selectedSubGroup) return;

    setLoading(true);
    try {
      await subscriptionAPI.updateSubscription(subscriptionID, {
        active: bangumiDetails.active,
        includeRegs: bangumiDetails.includeRegs,
        excludeRegs: bangumiDetails.excludeRegs,
        priority: bangumiDetails.priority,
        episodeOffset: bangumiDetails.episodeOffset,
        episodeLocation: bangumiDetails.episodeLocation,
        episodeTotalNum: bangumiDetails.episodeTotalNum || 0,
        airWeekday: bangumiDetails.airWeekday || 0,
      });

      toast({
        title: "配置已保存",
        description: `已成功保存 ${selectedSubGroup} 的配置信息`,
      });
      updateReleaseGroups(subscriptionID);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "保存配置失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleRetryTransfer = async (hash: string) => {
    try {
      await subscriptionAPI.transferTorrent(hash);
      toast({
        title: "文件转移重试",
        description: "已重新转移文件",
      });
      // 重新加载文件列表
      loadTabData();
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "文件转移重试失败",
        description: description,
        variant: "destructive",
      });
    }
  };

  const handleDeleteFile = async (
    name: string,
    hash: string,
    deleteSourceFile: boolean
  ) => {
    try {
      await subscriptionAPI.deleteTorrent(hash, deleteSourceFile);
      toast({
        title: "文件删除成功",
        description: `已删除 ${name} 的 ${
          deleteSourceFile ? "源文件和媒体文件" : "媒体文件"
        }`,
      });
      // 重新加载文件列表
      loadTabData();
    } catch (error) {
      toast({
        title: "文件删除失败",
        description: String(error),
        variant: "destructive",
      });
    }
  };

  // 添加格式化下载速度的函数
  const formatSpeed = (bytesPerSecond: number): string => {
    if (bytesPerSecond === 0) return "0 B/s";
    const units = ["B/s", "KiB/s", "MiB/s", "GiB/s"];
    const exponent = Math.min(
      Math.floor(Math.log(bytesPerSecond) / Math.log(1024)),
      units.length - 1
    );
    const value = (bytesPerSecond / Math.pow(1024, exponent)).toFixed(2);
    return `${value} ${units[exponent]}`;
  };

  // 处理立即执行按钮点击
  const handleRunTaskNow = async () => {
    if (!subscriptionID) return;

    setLoading(true);
    try {
      // 触发RSS检测
      await subscriptionAPI.handleSubscription(subscriptionID);

      // 重新加载RSS预览数据
      const rssMatchData = await subscriptionAPI.getRSSMatch(subscriptionID);

      // 更新状态
      setRssMatches(rssMatchData);

      toast({
        title: "RSS检测下载",
        description: "已完成RSS订阅项检测下载",
      });
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "RSS检测失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  // 处理批量标记
  const handleBatchMark = async (processed: boolean) => {
    if (!subscriptionID || selectedItems.length === 0) return;

    setLoading(true);
    try {
      await subscriptionAPI.markRSSRecord(
        subscriptionID,
        selectedItems,
        processed
      );

      // 更新本地状态
      setRssMatches((prevMatches) =>
        prevMatches.map((item) =>
          selectedItems.includes(item.guid) ? { ...item, processed } : item
        )
      );

      setSelectedItems([]); // 清空选择

      toast({
        title: "批量标记已更新",
        description: `已将 ${selectedItems.length} 个项目标记为${
          processed ? "已处理" : "未处理"
        }`,
      });
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "批量标记更新失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  // 处理删除订阅
  const handleDeleteSubscription = async (
    id: string = subscriptionID,
    shouldDeleteFiles: boolean = false
  ) => {
    if (!id || !selectedSubGroup || !bangumiDetails || !curReleaseGroups)
      return;

    setLoading(true);
    try {
      await subscriptionAPI.deleteSubscription(id, shouldDeleteFiles);

      // 显示成功消息
      toast({
        title: "删除成功",
        description: `已删除 ${bangumiDetails.name} 的 ${selectedSubGroup} 订阅`,
      });
      // 关闭删除确认对话框
      setShowDeleteDialog(false);

      updateReleaseGroups(subscriptionID);
      if (curReleaseGroups.length !== 0) {
        setSubscriptionID(curReleaseGroups[0].subscriptionID);
        setConfigLoaded(false);
      }
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "删除失败",
        description: description,
        variant: "destructive",
      });
      setShowDeleteDialog(false);
    } finally {
      setLoading(false);
    }
  };

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
          }
        }}
      >
        <AlertDialogContent
          onClick={(e) => e.stopPropagation()}
          className="w-[95vw] max-w-md p-4 md:p-6"
        >
          <AlertDialogHeader>
            <AlertDialogTitle className="text-lg md:text-xl">
              确认删除订阅
            </AlertDialogTitle>
            <AlertDialogDescription className="text-sm mt-2 break-words">
              你确定要删除{" "}
              <strong className="break-all">{selectedSubGroup}</strong>{" "}
              订阅吗？此操作无法撤销。
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
              同时删除已下载的文件
            </div>
          </div>

          <AlertDialogFooter className="flex-col space-y-2 sm:space-y-0 sm:flex-row">
            <AlertDialogCancel className="mt-2 sm:mt-0" disabled={loading}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={() =>
                handleDeleteSubscription(subscriptionID, deleteFilesRef.current)
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

  // 文件详情对话框组件
  const FileDetailsDialog = () => {
    const [files, setFiles] = useState<TorrentFile[]>([]);
    // 添加加载状态
    const [detailsLoading, setDetailsLoading] = useState(false);

    // 每当showFileDetails或selectedTorrent变化时，获取最新数据
    useEffect(() => {
      if (!showFileDetails || !selectedTorrent) return;

      const fetchLatestTorrentData = async () => {
        try {
          setDetailsLoading(true);
          // 直接从API获取最新数据
          const fileList = await subscriptionAPI.getTorrentFiles(
            selectedTorrent.hash
          );
          setFiles(fileList);
        } catch (error) {
          console.error("Failed to fetch latest torrent data:", error);
        } finally {
          setDetailsLoading(false);
        }
      };

      fetchLatestTorrentData();

      // 添加效果以在对话框打开/关闭时管理定时器
      if (showFileDetails) {
        clearRefreshTimer();
      } else if (activeTab === "files" && subscriptionID) {
        clearRefreshTimer(); // 先清除，确保没有多个定时器
        refreshTimerRef.current = setInterval(() => {
          updateTorrents();
        }, refreshInterval);
      }

      return () => {
        // 组件卸载时清理
        if (refreshTimerRef.current) {
          clearInterval(refreshTimerRef.current);
          refreshTimerRef.current = null;
        }
      };
    }, [showFileDetails, selectedTorrent, subscriptionID]);

    if (files.length === 0) return null;

    return (
      <Dialog open={showFileDetails} onOpenChange={setShowFileDetails}>
        <DialogContent className="w-[95dvw] sm:max-w-[700px] max-h-[80dvh] overflow-auto p-4 md:p-6 scrollbar-hide rounded-xl">
          <DialogHeader>
            <DialogTitle className="text-lg md:text-xl">
              文件转移详情
            </DialogTitle>
            <DialogDescription></DialogDescription>
          </DialogHeader>

          {detailsLoading ? (
            <div className="flex justify-center items-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
          ) : files.length === 0 ? (
            <div className="text-sm text-muted-foreground p-4 text-center">
              无文件信息
            </div>
          ) : (
            <ScrollArea className="max-h-[60dvh]">
              <div className="space-y-3">
                {files.map((file, index) => {
                  const hasLinkName =
                    file.linkName && file.linkName.trim() !== "";
                  return (
                    <Card
                      key={index}
                      className={cn(
                        "anime-card transition-all duration-200",
                        hasLinkName
                          ? "border-green-500/30 bg-gradient-to-r from-green-500/5 to-transparent hover:border-green-500/50"
                          : "border-gray-500/30 bg-gradient-to-r from-gray-500/5 to-transparent hover:border-gray-500/50"
                      )}
                    >
                      <CardContent className="p-3 md:p-4">
                        <div className="space-y-3">
                          {/* 季度和集数标签 - 突出显示识别信息 */}
                          <div className="flex items-center gap-2 flex-wrap">
                            <div className="flex items-center gap-1.5 px-2 py-1 bg-gradient-to-r from-primary/20 to-blue-500/20 rounded-lg border border-primary/30">
                              <Sparkles className="h-3 w-3 text-primary" />
                              <span className="text-xs font-medium text-muted-foreground">
                                识别季集:
                              </span>
                            </div>
                            <Badge
                              variant="outline"
                              className="bg-primary/20 text-primary border-primary/40 px-2.5 py-1 text-xs font-semibold shadow-sm"
                            >
                              第 {file.season} 季
                            </Badge>
                            <Badge
                              variant="outline"
                              className="bg-blue-500/20 text-blue-500 border-blue-500/40 px-2.5 py-1 text-xs font-semibold shadow-sm"
                            >
                              第 {file.episode.toString().padStart(2, "0")} 集
                            </Badge>
                          </div>

                          {/* 原始文件和媒体库文件 - 左右排布 */}
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 md:gap-4">
                            {/* 原始文件 */}
                            <div className="space-y-1.5">
                              <div className="flex items-center gap-2">
                                <span className="text-xs font-medium text-muted-foreground">
                                  原始文件:
                                </span>
                              </div>
                              <div className="break-all text-xs md:text-sm bg-muted/50 rounded-lg px-3 py-2 font-mono min-h-[2.5rem] flex items-center">
                                <TruncatedText text={file.fileName} />
                              </div>
                            </div>

                            {/* 媒体库文件 */}
                            <div className="space-y-1.5">
                              <div className="flex items-center gap-2">
                                <span className="text-xs font-medium text-muted-foreground">
                                  媒体库文件:
                                </span>
                              </div>
                              <div
                                className={cn(
                                  "break-all text-xs md:text-sm rounded-lg px-3 py-2 font-mono min-h-[2.5rem] flex items-center",
                                  hasLinkName
                                    ? "bg-primary/10 text-primary"
                                    : "bg-gray-500/10 text-gray-600 dark:text-gray-400 italic"
                                )}
                              >
                                {hasLinkName ? (
                                  <TruncatedText text={file.linkName || ""} />
                                ) : (
                                  <span className="text-muted-foreground text-xs">
                                    存在更高优先级的文件，未转移或被覆盖
                                  </span>
                                )}
                              </div>
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </ScrollArea>
          )}
        </DialogContent>
      </Dialog>
    );
  };

  return (
    <>
      <div
        className={cn(
          "fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/20",
          !open && "hidden"
        )}
        onClick={(e) => {
          e.stopPropagation();
          handleClose();
        }}
      >
        <div
          className="animate-accordion-down shadow-xl subscription-detail-card bg-background rounded-xl scrollbar-hide"
          style={{
            width: "min(92dvw, 1000px)",
            maxHeight: "92dvh",
            overflow: "auto",
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <Card className="border-primary/10 rounded-xl overflow-hidden">
            {/* 背景图和信息区域 */}
            <div className="relative">
              {/* 背景图 */}
              <div
                className="absolute inset-0 bg-cover bg-center"
                style={{
                  backgroundImage: `url(${bangumiDetails?.backdropURL})`,
                  backgroundPosition: "top",
                  filter: "blur(5px)",
                  opacity: 0.7,
                }}
              />

              {/* 内容区域 */}
              <div className="relative flex p-4 md:p-6 pb-6 md:pb-8">
                {/* 海报 */}
                <div className="w-[120px] h-[180px] md:w-[180px] md:h-[270px] rounded-lg overflow-hidden shadow-lg mr-4 md:mr-6 flex-shrink-0">
                  <img
                    src={bangumiDetails?.posterURL}
                    alt={bangumiDetails?.name}
                    className="w-full h-full object-cover"
                  />
                </div>

                {/* 番剧信息 */}
                <div className="min-w-0 flex-1 flex flex-col md:justify-between">
                  <div>
                    {/* 标题和关闭按钮 */}
                    <div className="flex justify-between items-start">
                      <h3 className="text-lg md:text-3xl font-bold text-foreground truncate pr-2">
                        {bangumiDetails?.name}{" "}
                        <span className="text-muted-foreground text-sm md:text-base">
                          {bangumiDetails?.year
                            ? `(${bangumiDetails.year})`
                            : ""}
                        </span>
                      </h3>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="rounded-full text-foreground hover:text-foreground hover:bg-muted"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleClose();
                        }}
                      >
                        <X className="h-5 w-5" />
                      </Button>
                    </div>

                    {/* 标签 */}
                    {bangumiDetails?.genres && (
                      <div className="flex flex-wrap gap-1 md:gap-2 my-2 md:my-3">
                        {bangumiDetails.genres.split(",").map((tag, index) => (
                          <Badge
                            key={index}
                            variant="secondary"
                            className="bg-primary/20 text-primary border-none px-2 md:px-3 py-0.5 md:py-1 text-xs md:text-sm"
                          >
                            {tag.trim()}
                          </Badge>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* 剧情简介 */}
                  <div className="mt-3" style={{ marginBottom: "0" }}>
                    <h4 className="font-bold text-foreground mb-2">剧情简介</h4>
                    {bangumiDetails?.overview && (
                      <p
                        className={cn(
                          "overview-text",
                          bangumiDetails.overview.length > 300
                            ? "text-[0.7rem] md:text-[0.85rem]"
                            : "text-[0.75rem] md:text-[0.95rem]",
                          "leading-[1.4] md:leading-[1.5]",
                          "max-h-[5em] md:max-h-[7em]"
                        )}
                        style={{
                          minHeight: "2em",
                          height: "auto",
                          overflow: "auto",
                          paddingRight: "10px",
                        }}
                      >
                        {bangumiDetails.overview}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            </div>

            <CardContent className="p-6">
              <div className="mb-5">
                <div className="flex flex-wrap justify-between items-center mb-3">
                  <h4 className="font-medium flex items-center gap-1 mb-2 md:mb-0">
                    <Star className="h-4 w-4 text-blue-400" />
                    字幕组
                  </h4>
                  {selectedSubGroup && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-destructive hover:text-destructive hover:bg-destructive/10 text-xs md:text-sm"
                      onClick={() => setShowDeleteDialog(true)}
                      disabled={loading}
                    >
                      <Trash2 className="h-3 w-3 md:h-4 md:w-4 mr-1" />
                      删除订阅
                    </Button>
                  )}
                </div>
                <div className="flex flex-wrap gap-2 overflow-x-auto pb-2">
                  {curReleaseGroups.map((group, index) => (
                    <button
                      key={index}
                      onClick={(e) =>
                        handleReleaseGroupClick(
                          group.releaseGroup,
                          group.subscriptionID,
                          e
                        )
                      }
                      className={cn(
                        "rounded-lg px-3 py-1 text-xs transition-all duration-200",
                        selectedSubGroup === group.releaseGroup
                          ? group.active
                            ? "bg-primary/20 text-primary border-2 border-primary font-medium shadow-sm"
                            : "bg-white text-gray-700 border-2 border-blue-300 font-medium shadow-sm"
                          : group.active
                          ? "bg-primary/10 text-primary border border-primary/20"
                          : "bg-white text-muted-foreground border border-muted"
                      )}
                    >
                      {group.releaseGroup} ({group.lastAirEpisode}/
                      {group.episodeTotalNum})
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex w-full">
                <div className="flex flex-col w-full md:flex-row">
                  <div className="w-full md:w-48 mb-4 md:mb-0 md:mr-6 md:border-r md:pr-6">
                    <div className="space-y-3">
                      <h4 className="text-sm font-medium text-muted-foreground mb-3">
                        详情选项
                      </h4>
                      <div className="flex flex-col space-y-2">
                        <button
                          onClick={() => setActiveTab("config")}
                          className={cn(
                            "flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-all duration-300 transform",
                            activeTab === "config"
                              ? "bg-gradient-to-r from-primary/20 to-blue-500/20 text-primary shadow-sm translate-x-1"
                              : "text-muted-foreground hover:bg-muted hover:translate-x-1"
                          )}
                        >
                          <SettingsIcon
                            className={cn(
                              "mr-2 h-4 w-4 transition-transform duration-300",
                              activeTab === "config"
                                ? "text-primary scale-110"
                                : "text-muted-foreground"
                            )}
                          />
                          配置信息
                        </button>
                        <button
                          onClick={() => setActiveTab("preview")}
                          className={cn(
                            "flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-all duration-300 transform",
                            activeTab === "preview"
                              ? "bg-gradient-to-r from-primary/20 to-blue-500/20 text-primary shadow-sm translate-x-1"
                              : "text-muted-foreground hover:bg-muted hover:translate-x-1"
                          )}
                        >
                          <Eye
                            className={cn(
                              "mr-2 h-4 w-4 transition-transform duration-300",
                              activeTab === "preview"
                                ? "text-primary scale-110"
                                : "text-muted-foreground"
                            )}
                          />
                          订阅预览
                        </button>
                        <button
                          onClick={() => setActiveTab("files")}
                          className={cn(
                            "flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-all duration-300 transform",
                            activeTab === "files"
                              ? "bg-gradient-to-r from-primary/20 to-blue-500/20 text-primary shadow-sm translate-x-1"
                              : "text-muted-foreground hover:bg-muted hover:translate-x-1"
                          )}
                        >
                          <FileBox
                            className={cn(
                              "mr-2 h-4 w-4 transition-transform duration-300",
                              activeTab === "files"
                                ? "text-primary scale-110"
                                : "text-muted-foreground"
                            )}
                          />
                          文件管理
                        </button>
                      </div>
                    </div>
                  </div>

                  <div
                    className="flex-1 min-w-0"
                    style={{ minHeight: "400px" }}
                  >
                    <Tabs
                      value={activeTab}
                      defaultValue="config"
                      className="animate-tab-change"
                    >
                      <TabsContent
                        value="config"
                        className="mt-0 ml-0 min-h-[550px] animate-in fade-in duration-300"
                      >
                        {loading && (
                          <div className="flex justify-center items-center h-[50dvh] md:h-[550px]">
                            加载中...
                          </div>
                        )}
                        {!loading && bangumiDetails && (
                          <div className="space-y-4">
                            <div className="flex items-center justify-between">
                              <Label
                                htmlFor={`enabled-${subscriptionID}-${selectedSubGroup}`}
                              >
                                订阅状态
                              </Label>
                              <Switch
                                id={`enabled-${subscriptionID}-${selectedSubGroup}`}
                                checked={bangumiDetails.active}
                                onCheckedChange={(active) =>
                                  setBangumiDetails((prev) => ({
                                    ...prev!,
                                    active,
                                  }))
                                }
                              />
                            </div>

                            <MatchInput
                              label="包含匹配"
                              items={bangumiDetails.includeRegs}
                              placeholder="添加包含匹配条件"
                              onChange={(items) =>
                                setBangumiDetails((prev) => ({
                                  ...prev!,
                                  includeRegs: items,
                                }))
                              }
                            />

                            <MatchInput
                              label="排除匹配"
                              items={bangumiDetails.excludeRegs}
                              placeholder="添加排除匹配条件"
                              onChange={(items) =>
                                setBangumiDetails((prev) => ({
                                  ...prev!,
                                  excludeRegs: items,
                                }))
                              }
                            />

                            <div className="grid grid-cols-2 gap-4">
                              <div className="space-y-2">
                                <Label
                                  htmlFor={`priority-${subscriptionID}-${selectedSubGroup}`}
                                >
                                  优先级
                                </Label>
                                <Input
                                  id={`priority-${subscriptionID}-${selectedSubGroup}`}
                                  type="number"
                                  value={bangumiDetails.priority}
                                  onChange={(e) =>
                                    setBangumiDetails((prev) => ({
                                      ...prev!,
                                      priority: parseInt(e.target.value),
                                    }))
                                  }
                                  className="rounded-xl"
                                />
                              </div>
                              <div className="space-y-2">
                                <Label
                                  htmlFor={`offset-${subscriptionID}-${selectedSubGroup}`}
                                >
                                  集数偏移
                                </Label>
                                <Input
                                  id={`offset-${subscriptionID}-${selectedSubGroup}`}
                                  type="number"
                                  value={bangumiDetails.episodeOffset}
                                  onChange={(e) =>
                                    setBangumiDetails((prev) => ({
                                      ...prev!,
                                      episodeOffset: Number(e.target.value),
                                    }))
                                  }
                                  className="rounded-xl"
                                />
                              </div>
                              <div className="space-y-2">
                                <Label
                                  htmlFor={`total-${subscriptionID}-${selectedSubGroup}`}
                                >
                                  总集数
                                </Label>
                                <Input
                                  id={`total-${subscriptionID}-${selectedSubGroup}`}
                                  type="number"
                                  value={bangumiDetails?.episodeTotalNum || 0}
                                  onChange={(e) =>
                                    setBangumiDetails((prev) => ({
                                      ...prev!,
                                      episodeTotalNum:
                                        Number(e.target.value) * 1,
                                    }))
                                  }
                                  className="rounded-xl"
                                />
                              </div>
                              <div className="space-y-2">
                                <Label
                                  htmlFor={`weekday-${subscriptionID}-${selectedSubGroup}`}
                                >
                                  更新星期
                                </Label>
                                <select
                                  id={`weekday-${subscriptionID}-${selectedSubGroup}`}
                                  value={bangumiDetails?.airWeekday || 0}
                                  onChange={(e) =>
                                    setBangumiDetails((prev) => ({
                                      ...prev!,
                                      airWeekday: Number(e.target.value),
                                    }))
                                  }
                                  className="flex h-10 w-full rounded-xl border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                >
                                  <option value={0}>周日</option>
                                  <option value={1}>周一</option>
                                  <option value={2}>周二</option>
                                  <option value={3}>周三</option>
                                  <option value={4}>周四</option>
                                  <option value={5}>周五</option>
                                  <option value={6}>周六</option>
                                </select>
                              </div>
                            </div>

                            <EpisodePositionInput
                              value={bangumiDetails.episodeLocation}
                              onChange={(value) =>
                                setBangumiDetails((prev) => ({
                                  ...prev!,
                                  episodeLocation: value,
                                }))
                              }
                            />

                            <Button
                              className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button mt-8"
                              onClick={handleSaveConfig}
                              disabled={loading}
                            >
                              <Save className="mr-2 h-4 w-4" />
                              保存配置
                            </Button>
                          </div>
                        )}
                      </TabsContent>

                      <TabsContent
                        value="preview"
                        className="mt-0 ml-0 min-h-[550px] animate-in fade-in duration-300"
                      >
                        <div className="space-y-4">
                          <div className="flex flex-row justify-between items-center">
                            <div className="flex items-center gap-1">
                              <h4 className="font-bold text-sm md:text-base">
                                {selectedSubGroup}
                                <span className="hidden md:inline">
                                  {" "}
                                  RSS订阅项
                                </span>
                              </h4>
                              <HoverCard>
                                <HoverCardTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-4 w-4 sm:h-6 sm:w-6 rounded-full hover:bg-muted p-0 ml-0.5 mr-1.5"
                                  >
                                    <Info className="h-4 w-4 text-muted-foreground" />
                                  </Button>
                                </HoverCardTrigger>
                                <HoverCardContent
                                  className="w-[400px]"
                                  align="start"
                                  side="right"
                                >
                                  <div className="space-y-2">
                                    <h4 className="font-medium">状态说明</h4>
                                    <div className="text-sm space-y-1.5 text-muted-foreground">
                                      <div className="flex items-center gap-3">
                                        <Badge
                                          variant="outline"
                                          className="w-[58px] justify-center rounded-full bg-primary/10 text-primary border-primary/20"
                                        >
                                          匹配
                                        </Badge>
                                        <span>
                                          表示该项目符合您设置的匹配规则
                                        </span>
                                      </div>
                                      <div className="flex items-center gap-3">
                                        <Badge
                                          variant="outline"
                                          className="w-[58px] justify-center rounded-full bg-muted text-muted-foreground"
                                        >
                                          不匹配
                                        </Badge>
                                        <span>
                                          表示该项目不符合匹配规则，将被忽略
                                        </span>
                                      </div>
                                      <div className="flex items-center gap-3">
                                        <Badge
                                          variant="outline"
                                          className="w-[58px] justify-center rounded-full bg-green-500/10 text-green-500 border-green-500/20"
                                        >
                                          已处理
                                        </Badge>
                                        <span>
                                          表示该项目已被下载或手动标记为已处理
                                        </span>
                                      </div>
                                      <div className="flex items-center gap-3">
                                        <Badge
                                          variant="outline"
                                          className="w-[58px] justify-center rounded-full bg-yellow-500/10 text-yellow-500 border-yellow-500/20"
                                        >
                                          未处理
                                        </Badge>
                                        <span>
                                          表示该项目尚未处理，可能会在下次检测时下载
                                        </span>
                                      </div>
                                    </div>
                                  </div>
                                </HoverCardContent>
                              </HoverCard>
                            </div>
                            <div className="flex items-center gap-1 shrink-0">
                              {selectedItems.length > 0 && (
                                <>
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    className="rounded-xl text-xs sm:text-sm sm:flex-none whitespace-nowrap"
                                    onClick={() => handleBatchMark(true)}
                                    disabled={loading}
                                  >
                                    <CheckSquare className="h-4 w-4 -mr-1" />
                                    <span className="hidden sm:inline -mr-1">
                                      标记已处理
                                    </span>
                                    <span>({selectedItems.length})</span>
                                  </Button>
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    className="rounded-xl text-xs sm:text-sm sm:flex-none whitespace-nowrap"
                                    onClick={() => handleBatchMark(false)}
                                    disabled={loading}
                                  >
                                    <CircleSlash className="h-4 w-4 -mr-1" />
                                    <span className="hidden sm:inline -mr-1">
                                      标记未处理
                                    </span>
                                    <span>({selectedItems.length})</span>
                                  </Button>
                                </>
                              )}
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger asChild>
                                    <Button
                                      size="sm"
                                      className="rounded-xl anime-button text-xs sm:text-sm whitespace-nowrap"
                                      onClick={handleRunTaskNow}
                                      disabled={loading}
                                    >
                                      <Play
                                        className={cn(
                                          "h-4 w-4 sm:h-4 sm:w-4 mr-1",
                                          loading && "animate-spin"
                                        )}
                                      />
                                      <span className="hidden sm:inline">
                                        立即执行
                                      </span>
                                    </Button>
                                  </TooltipTrigger>
                                  <TooltipContent
                                    side="top"
                                    className="text-xs max-w-[150px] whitespace-normal break-words"
                                  >
                                    立即检查RSS链接并下载未处理但已匹配的种子
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                            </div>
                          </div>

                          {loading && (
                            <div className="flex justify-center items-center h-[450px]">
                              加载中...
                            </div>
                          )}
                          {!loading && rssMatches.length > 0 && (
                            <div className="relative">
                              <ScrollArea className="h-[50dvh] md:h-[450px]">
                                <div className="space-y-2">
                                  {/* 全选/全不选按钮区域 */}
                                  <div className="flex gap-2 mb-2">
                                    <Button
                                      size="sm"
                                      variant="outline"
                                      className="rounded-xl text-xs"
                                      onClick={() =>
                                        setSelectedItems(
                                          rssMatches.map((item) => item.guid)
                                        )
                                      }
                                      disabled={rssMatches.length === 0}
                                    >
                                      全选
                                    </Button>
                                    {selectedItems.length > 0 && (
                                      <Button
                                        size="sm"
                                        variant="outline"
                                        className="rounded-xl text-xs"
                                        onClick={() => setSelectedItems([])}
                                      >
                                        全不选
                                      </Button>
                                    )}
                                  </div>
                                  {/* 订阅项内容块 */}
                                  {rssMatches.map((item, index) => (
                                    <div
                                      key={index}
                                      className="rounded-xl border p-2 text-sm transition-all duration-200"
                                    >
                                      <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-2">
                                          <Checkbox
                                            id={`select-${item.guid}`}
                                            checked={selectedItems.includes(
                                              item.guid
                                            )}
                                            onCheckedChange={(checked) => {
                                              setSelectedItems((prev) =>
                                                checked
                                                  ? [...prev, item.guid]
                                                  : prev.filter(
                                                      (id) => id !== item.guid
                                                    )
                                              );
                                            }}
                                          />
                                          <TooltipProvider>
                                            <HybridTooltip delayDuration={200}>
                                              <HybridTooltipTrigger asChild>
                                                <div className="flex-1 flex items-center gap-2 min-w-0">
                                                  <span
                                                    className={cn(
                                                      "flex w-full max-w-[200px] sm:max-w-[500px]",
                                                      item.processed &&
                                                        "text-muted-foreground"
                                                    )}
                                                  >
                                                    <TruncatedText
                                                      text={item.guid}
                                                    />
                                                  </span>
                                                  {!isMobile && (
                                                    <div className="w-32 text-xs text-muted-foreground text-center m-2 shrink-0">
                                                      {formatDate(
                                                        item.publishedAt
                                                      )}
                                                    </div>
                                                  )}
                                                </div>
                                              </HybridTooltipTrigger>
                                              <HybridTooltipContent
                                                side="top"
                                                align="start"
                                                sideOffset={5}
                                                className="break-words text-xs max-w-[300px] whitespace-normal"
                                              >
                                                {item.guid}
                                                {isMobile &&
                                                  item.publishedAt && (
                                                    <div className="mt-1 text-xs text-muted-foreground">
                                                      发布时间：
                                                      {formatDate(
                                                        item.publishedAt
                                                      )}
                                                    </div>
                                                  )}
                                              </HybridTooltipContent>
                                            </HybridTooltip>
                                          </TooltipProvider>
                                        </div>
                                        <div className="flex items-center gap-2 flex-shrink-0">
                                          <Badge
                                            variant="outline"
                                            className={cn(
                                              "w-[58px] justify-center rounded-full",
                                              item.match
                                                ? "bg-primary/10 text-primary border-primary/20"
                                                : "bg-muted text-muted-foreground"
                                            )}
                                          >
                                            {item.match ? "匹配" : "不匹配"}
                                          </Badge>
                                          <Badge
                                            variant="outline"
                                            className={cn(
                                              "w-[58px] justify-center rounded-full",
                                              item.processed
                                                ? "bg-green-500/10 text-green-500 border-green-500/20"
                                                : "bg-yellow-500/10 text-yellow-500 border-yellow-500/20"
                                            )}
                                          >
                                            {item.processed
                                              ? "已处理"
                                              : "未处理"}
                                          </Badge>
                                        </div>
                                      </div>
                                    </div>
                                  ))}
                                </div>
                              </ScrollArea>
                            </div>
                          )}
                          {!loading && rssMatches.length === 0 && (
                            <div className="flex justify-center items-center h-[450px] text-muted-foreground">
                              无RSS订阅项，请稍后再试
                            </div>
                          )}
                        </div>
                      </TabsContent>

                      <TabsContent
                        value="files"
                        className="mt-0 ml-0 min-h-[550px] animate-in fade-in duration-300"
                      >
                        <div className="space-y-4">
                          <div className="flex items-center gap-2">
                            <h4 className="font-medium">
                              {selectedSubGroup} 媒体文件
                            </h4>
                            <HoverCard>
                              <HoverCardTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-6 w-6 rounded-full hover:bg-muted"
                                >
                                  <Info className="h-4 w-4 text-muted-foreground" />
                                </Button>
                              </HoverCardTrigger>
                              <HoverCardContent
                                className="w-[320px]"
                                align="start"
                                side="right"
                              >
                                <div className="space-y-2">
                                  <h4 className="font-medium">状态图标说明</h4>
                                  <div className="text-sm space-y-2 text-muted-foreground">
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 flex items-center justify-center">
                                        <CircleArrowDown className="w-5 h-5 text-green-500" />
                                      </div>
                                      <span>已下载完成</span>
                                    </div>
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 flex items-center justify-center">
                                        <CheckCircle2 className="w-5 h-5 text-green-500" />
                                      </div>
                                      <span>已处理完成</span>
                                    </div>
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 flex items-center justify-center">
                                        <CircleX className="w-5 h-5 text-yellow-500" />
                                      </div>
                                      <span>未知状态，程序异常</span>
                                    </div>
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 flex items-center justify-center">
                                        <PauseCircle className="w-5 h-5 text-gray-500" />
                                      </div>
                                      <span>下载已暂停</span>
                                    </div>
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 flex items-center justify-center">
                                        <CircleArrowDown className="w-5 h-5 text-destructive" />
                                      </div>
                                      <span>
                                        下载失败，鼠标悬浮在图标上查看原因
                                      </span>
                                    </div>
                                    <div className="flex items-center gap-3">
                                      <div className="w-8 h-8 flex items-center justify-center">
                                        <CircleAlert className="w-5 h-5 text-destructive" />
                                      </div>
                                      <span>
                                        转移失败，鼠标悬浮在图标上查看原因
                                      </span>
                                    </div>
                                  </div>
                                </div>
                              </HoverCardContent>
                            </HoverCard>
                          </div>

                          {loading && (
                            <div className="flex justify-center items-center h-[50dvh] md:h-[500px]">
                              加载中...
                            </div>
                          )}
                          {!loading && torrents.length > 0 && (
                            <ScrollArea className="h-[50dvh] md:h-[500px]">
                              <div className="space-y-2">
                                {torrents.map((torrent) => (
                                  <div
                                    key={torrent.hash}
                                    className="rounded-xl border p-2 text-sm"
                                  >
                                    <div className="flex flex-col md:flex-row md:items-center md:justify-between">
                                      <TooltipProvider>
                                        <HybridTooltip>
                                          <HybridTooltipTrigger className="flex-1 mb-2 md:mb-0">
                                            <div className="text-left">
                                              <div className="line-clamp-1 text-primary font-medium">
                                                <TruncatedText
                                                  text={torrent.rssGUID}
                                                />
                                              </div>
                                              <div className="line-clamp-1 text-xs text-muted-foreground bg-muted/50 rounded px-1 py-0.5 mt-0.5">
                                                <TruncatedText
                                                  text={torrent.name}
                                                />
                                              </div>
                                              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                <span>
                                                  {formatDate(
                                                    torrent.createdAt
                                                  )}
                                                </span>
                                                {torrent.collection ? (
                                                  <Badge
                                                    variant="outline"
                                                    className="bg-emerald-500/10 text-emerald-500 dark:text-emerald-400 border-emerald-500/20 px-1.5 py-0 text-xs"
                                                  >
                                                    合集
                                                  </Badge>
                                                ) : (
                                                  torrent.season !==
                                                    undefined &&
                                                  torrent.episode !==
                                                    undefined && (
                                                    <Badge
                                                      variant="outline"
                                                      className="bg-emerald-500/10 text-emerald-500 dark:text-emerald-400 border-emerald-500/20 px-1.5 py-0 text-xs font-mono"
                                                    >
                                                      S
                                                      {torrent.season
                                                        .toString()
                                                        .padStart(2, "0")}
                                                      E
                                                      {torrent.episode
                                                        .toString()
                                                        .padStart(2, "0")}
                                                    </Badge>
                                                  )
                                                )}
                                              </div>
                                            </div>
                                          </HybridTooltipTrigger>
                                          <HybridTooltipContent
                                            side="top"
                                            className="max-w-[300px] md:max-w-[500px] w-auto"
                                          >
                                            <div className="space-y-1">
                                              <div className="font-medium">
                                                RSS订阅项:
                                              </div>
                                              <div className="text-xs break-all">
                                                {torrent.rssGUID}
                                              </div>
                                              <div className="font-medium mt-2">
                                                种子名:
                                              </div>
                                              <div className="text-xs break-all">
                                                {torrent.name}
                                              </div>
                                            </div>
                                          </HybridTooltipContent>
                                        </HybridTooltip>
                                      </TooltipProvider>
                                      <div className="flex items-center justify-end gap-1">
                                        {torrent.status ===
                                        TorrentStatusSet.Downloading ? (
                                          <div className="flex items-center w-40">
                                            <div className="w-24 mr-1">
                                              <Progress
                                                value={torrent.progress * 100}
                                                className="h-2 rounded-full bg-secondary"
                                              />
                                            </div>
                                            <span className="text-xs text-muted-foreground whitespace-nowrap w-20 text-right">
                                              {formatSpeed(
                                                torrent.downloadSpeed
                                              )}
                                            </span>
                                          </div>
                                        ) : torrent.status ===
                                          TorrentStatusSet.DownloadPaused ? (
                                          <StatusTooltip content="已暂停">
                                            <PauseCircle className="w-5 h-5 text-gray-500" />
                                          </StatusTooltip>
                                        ) : torrent.status ===
                                          TorrentStatusSet.DownloadError ? (
                                          <StatusTooltip
                                            content={`下载失败：${torrent.statusDetail}`}
                                          >
                                            <CircleArrowDown className="w-5 h-5 text-destructive" />
                                          </StatusTooltip>
                                        ) : torrent.status ===
                                          TorrentStatusSet.TransferredError ? (
                                          <div className="flex items-center gap-2">
                                            <StatusTooltip
                                              content={`转移失败：${torrent.statusDetail}`}
                                            >
                                              <CircleAlert className="w-5 h-5 text-destructive" />
                                            </StatusTooltip>
                                          </div>
                                        ) : torrent.status ===
                                          TorrentStatusSet.Downloaded ? (
                                          <StatusTooltip content="已下载">
                                            <CircleArrowDown className="w-5 h-5 text-green-500" />
                                          </StatusTooltip>
                                        ) : torrent.status ===
                                          TorrentStatusSet.Transferred ? (
                                          <TooltipProvider>
                                            <Tooltip>
                                              <TooltipTrigger asChild>
                                                <Button
                                                  variant="ghost"
                                                  size="icon"
                                                  className="w-8 h-8 rounded-full hover:bg-primary/10 group"
                                                  onClick={() => {
                                                    setSelectedTorrent(torrent);
                                                    setShowFileDetails(true);
                                                  }}
                                                >
                                                  <CheckCircle2 className="w-5 h-5 text-green-500 group-hover:hidden" />
                                                  <Eye className="w-5 h-5 text-green-500 hidden group-hover:block" />
                                                </Button>
                                              </TooltipTrigger>
                                              <TooltipContent>
                                                已完成处理，点击查看转移详情
                                              </TooltipContent>
                                            </Tooltip>
                                          </TooltipProvider>
                                        ) : (
                                          <StatusTooltip content="未知">
                                            <CircleX className="w-5 h-5 text-yellow-500" />
                                          </StatusTooltip>
                                        )}
                                        {torrentCanTransfer(torrent.status) && (
                                          <TransferButton hash={torrent.hash} />
                                        )}
                                        <DropdownMenu>
                                          <DropdownMenuTrigger asChild>
                                            <Button
                                              size="icon"
                                              variant="ghost"
                                              className="h-6 w-6 rounded-full"
                                            >
                                              <X className="h-3 w-3" />
                                            </Button>
                                          </DropdownMenuTrigger>
                                          <DropdownMenuContent
                                            align="end"
                                            className="w-48 rounded-xl"
                                          >
                                            <DropdownMenuItem
                                              onClick={() =>
                                                handleDeleteFile(
                                                  torrent.name,
                                                  torrent.hash,
                                                  false
                                                )
                                              }
                                            >
                                              删除媒体库文件
                                            </DropdownMenuItem>
                                            <DropdownMenuItem
                                              onClick={() =>
                                                handleDeleteFile(
                                                  torrent.name,
                                                  torrent.hash,
                                                  true
                                                )
                                              }
                                            >
                                              删除源文件与媒体库文件
                                            </DropdownMenuItem>
                                          </DropdownMenuContent>
                                        </DropdownMenu>
                                      </div>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            </ScrollArea>
                          )}
                          {!loading && torrents.length === 0 && (
                            <div className="flex justify-center items-center h-[50dvh] md:h-[500px] text-muted-foreground">
                              无媒体文件
                            </div>
                          )}
                        </div>
                      </TabsContent>
                    </Tabs>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <DeleteConfirmDialog />
      <FileDetailsDialog />
    </>
  );
}
