import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableRow } from "@/components/ui/table";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useToast } from "@/hooks/useToast";
import subscriptionAPI, {
  ListRecentUpdatedTorrentsReq,
  RecentUpdatedTorrent,
  TorrentStatusSet,
} from "@/api/subscription";
import { AxiosError } from "axios";
import { toZonedTime, format } from "date-fns-tz";

interface RecentUpdatesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

type TimeRange = {
  label: string;
  startTime: Date;
  endTime: Date;
};

export function RecentUpdatesDialog({
  open,
  onOpenChange,
}: RecentUpdatesDialogProps) {
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState("yesterday-to-now");
  const [torrents, setTorrents] = useState<RecentUpdatedTorrent[]>([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const pageSize = 10;
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone; // 获取当前浏览器的时区

  const getDayTime = (date: Date) => {
    const start = new Date(date.setHours(0, 0, 0, 0));
    const end = new Date(date);
    end.setHours(23, 59, 59, 999);
    return { start, end };
  };

  const getTimeRanges = (): Record<string, TimeRange> => {
    const now = new Date();
    const { start: todayStart, end: todayEnd } = getDayTime(now);

    const weekAgo = new Date(now);
    weekAgo.setDate(weekAgo.getDate() - 6);
    const { start: weekAgoStart } = getDayTime(weekAgo);

    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    const { start: yesterdayStart, end: yesterdayEnd } = getDayTime(yesterday);

    const pastThreeDays = new Date(now);
    pastThreeDays.setDate(pastThreeDays.getDate() - 2);
    const { start: pastThreeDaysStart } = getDayTime(pastThreeDays);
    return {
      "yesterday-to-now": {
        label: "昨天至今",
        startTime: yesterdayStart,
        endTime: todayEnd,
      },
      today: {
        label: "今天",
        startTime: todayStart,
        endTime: todayEnd,
      },
      yesterday: {
        label: "昨天",
        startTime: yesterdayStart,
        endTime: yesterdayEnd,
      },
      "three-days": {
        label: "近三天",
        startTime: pastThreeDaysStart,
        endTime: todayEnd,
      },
      week: {
        label: "最近一周",
        startTime: weekAgoStart,
        endTime: todayEnd,
      },
    };
  };

  const timeRanges = getTimeRanges();

  const toRFC3339WithTimezone = (date: Date) => {
    const zonedDate = toZonedTime(date, timeZone);
    return format(zonedDate, "yyyy-MM-dd'T'HH:mm:ssXXX", { timeZone });
  };

  const fetchTorrents = async (page: number = 1) => {
    const selectedRange = timeRanges[activeTab];

    if (!selectedRange) return;

    try {
      setLoading(true);

      const params: ListRecentUpdatedTorrentsReq = {
        start_time: toRFC3339WithTimezone(selectedRange.startTime),
        end_time: toRFC3339WithTimezone(selectedRange.endTime),
        page: page,
        page_size: pageSize,
      };

      const data = await subscriptionAPI.listRecentUpdatedTorrents(params);
      setTorrents(data.torrents);
      setTotal(data.total);
      setCurrentPage(page);
    } catch (error) {
      const description =
        (error as AxiosError<{ error: string }>)?.response?.data?.error ||
        "未知原因失败，请重试";
      toast({
        title: "获取近期更新失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  // 当对话框打开或tab变化时获取数据
  useEffect(() => {
    if (open) {
      setCurrentPage(1);
      fetchTorrents(1);
    }
  }, [open, activeTab]);

  const renderPagination = () => {
    const totalPages = Math.ceil(total / pageSize);
    if (totalPages <= 1) return null;

    return (
      <Pagination className="mt-4">
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              onClick={() => currentPage > 1 && fetchTorrents(currentPage - 1)}
              className={
                currentPage <= 1 ? "pointer-events-none opacity-50" : ""
              }
            />
          </PaginationItem>

          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
            // 显示当前页附近的页码
            let pageNum = i + 1;
            if (currentPage > 3 && totalPages > 5) {
              // 如果当前页大于3，调整显示的页码范围
              if (currentPage + 2 > totalPages) {
                // 如果当前页接近末尾，显示最后5页
                pageNum = totalPages - 4 + i;
              } else {
                // 否则当前页在中间，显示当前页前后2页
                pageNum = currentPage - 2 + i;
              }
            }

            return (
              <PaginationItem key={pageNum}>
                <PaginationLink
                  isActive={pageNum === currentPage}
                  onClick={() => fetchTorrents(pageNum)}
                >
                  {pageNum}
                </PaginationLink>
              </PaginationItem>
            );
          })}

          <PaginationItem>
            <PaginationNext
              onClick={() =>
                currentPage < totalPages && fetchTorrents(currentPage + 1)
              }
              className={
                currentPage >= totalPages
                  ? "pointer-events-none opacity-50"
                  : ""
              }
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    );
  };

  const renderStatus = (status: string) => {
    let color = "";
    let text = status;

    switch (status) {
      case TorrentStatusSet.Downloading:
        color = "bg-blue-500";
        text = "下载中";
        break;
      case TorrentStatusSet.Downloaded:
        color = "bg-green-500";
        text = "已下载";
        break;
      case TorrentStatusSet.Transferred:
        color = "bg-purple-500";
        text = "转移完成";
        break;
      case TorrentStatusSet.TransferredError:
        color = "bg-amber-500";
        text = "转移错误";
        break;
      case TorrentStatusSet.DownloadError:
        color = "bg-red-500";
        text = "下载错误";
        break;
      case TorrentStatusSet.DownloadPaused:
        color = "bg-gray-500";
        text = "下载暂停";
        break;
    }

    return (
      <Badge className={`${color} text-white whitespace-nowrap`}>{text}</Badge>
    );
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    const zonedDate = toZonedTime(date, timeZone);
    return format(zonedDate, "yyyy/MM/dd HH:mm", { timeZone });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[90vh] sm:max-h-[80vh] w-[95vw] sm:w-auto overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="text-xl sm:text-2xl font-bold">
            近期更新
          </DialogTitle>
        </DialogHeader>
        <DialogDescription></DialogDescription>

        <div className="flex-grow overflow-hidden flex flex-col">
          <Tabs
            defaultValue="yesterday-to-now"
            value={activeTab}
            onValueChange={setActiveTab}
            className="w-full"
          >
            <TabsList className="w-full justify-start mb-4 overflow-x-auto">
              {Object.entries(timeRanges).map(([key, range]) => (
                <TabsTrigger key={key} value={key} className="min-w-fit">
                  {range.label}
                </TabsTrigger>
              ))}
            </TabsList>

            {Object.keys(timeRanges).map((key) => (
              <TabsContent
                key={key}
                value={key}
                className="flex-grow overflow-auto"
              >
                {loading ? (
                  <div className="rounded-md border overflow-auto">
                    <Table>
                      <TableBody>
                        <TableRow>
                          <TableCell colSpan={4} className="text-center py-8">
                            加载中...
                          </TableCell>
                        </TableRow>
                      </TableBody>
                    </Table>
                  </div>
                ) : torrents.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <img
                      src="/empty-updates.png"
                      alt="暂无更新"
                      onError={(e) => {
                        // 图片加载失败时使用emoji作为备用
                        const target = e.target as HTMLImageElement;
                        target.onerror = null; // 防止循环触发错误
                        target.style.display = "none";
                      }}
                      className="w-32 h-32 mb-4"
                    />
                    <div className="text-xl sm:text-2xl font-bold text-primary mt-4">
                      现在还没有更新哦
                    </div>
                    <div className="text-secondary-foreground mt-2">
                      稍后再来看看吧~
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="rounded-md border overflow-auto">
                      <Table>
                        <TableBody>
                          {torrents.map((torrent, index) => (
                            <TableRow key={index} className="hover:bg-muted/30">
                              <TableCell className="w-14 sm:w-20 p-1 sm:p-2 hidden sm:table-cell">
                                <img
                                  src={
                                    torrent.posterURL ||
                                    "/placeholder.svg?height=80&width=60"
                                  }
                                  alt={torrent.bangumiName}
                                  className="w-12 h-16 sm:w-16 sm:h-20 object-cover rounded-sm"
                                />
                              </TableCell>
                              <TableCell className="align-top py-2">
                                <div className="flex flex-col h-full justify-between">
                                  <div className="font-medium text-sm sm:text-base">
                                    <span className="text-primary">
                                      {torrent.bangumiName}
                                    </span>
                                    <span className="text-secondary-foreground ml-1">
                                      (第{torrent.season}季)
                                    </span>
                                  </div>
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <div className="text-xs sm:text-sm text-muted-foreground mt-1 truncate max-w-[200px] sm:max-w-[500px] cursor-default">
                                          {torrent.rssItem}
                                        </div>
                                      </TooltipTrigger>
                                      <TooltipContent
                                        side="top"
                                        className="max-w-[250px] sm:max-w-md"
                                      >
                                        <p className="break-words text-xs sm:text-sm">
                                          {torrent.rssItem}
                                        </p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                </div>
                              </TableCell>
                              <TableCell className="whitespace-nowrap text-right w-24 sm:w-40 text-xs sm:text-sm">
                                <div className="flex flex-col sm:block">
                                  <span className="sm:inline">
                                    {
                                      formatDate(torrent.createdAt).split(
                                        " "
                                      )[0]
                                    }
                                  </span>
                                  <span className="sm:inline sm:ml-1">
                                    {
                                      formatDate(torrent.createdAt).split(
                                        " "
                                      )[1]
                                    }
                                  </span>
                                </div>
                              </TableCell>
                              <TableCell className="w-32 sm:w-32 text-right whitespace-nowrap">
                                {torrent.statusDetail ? (
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <div className="inline-block whitespace-nowrap">
                                          {renderStatus(torrent.status)}
                                        </div>
                                      </TooltipTrigger>
                                      <TooltipContent side="bottom">
                                        <p className="text-xs sm:text-sm">
                                          {torrent.statusDetail}
                                        </p>
                                      </TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                ) : (
                                  <div className="inline-block whitespace-nowrap">
                                    {renderStatus(torrent.status)}
                                  </div>
                                )}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                    {renderPagination()}
                  </>
                )}
              </TabsContent>
            ))}
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  );
}
