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
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
import { Skeleton } from "@/components/ui/skeleton";
import { useToast } from "@/hooks/useToast";
import subscriptionAPI, {
  ListRecentUpdatedTorrentsReq,
  RecentUpdatedTorrent,
  TorrentStatusSet,
} from "@/api/subscription";
import { toZonedTime, format } from "date-fns-tz";
import { TruncatedText } from "@/components/common/truncate-rss-item";
import { formatDate } from "@/utils/time";
import { extractErrorMessage } from "@/utils/error";
import { renderDownloadStatus } from "@/utils/status";

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
  const pageSize = 5;
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
      const description = extractErrorMessage(error);
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
    const { color, text } = renderDownloadStatus(status);

    return (
      <Badge className={`${color} text-white whitespace-nowrap`}>{text}</Badge>
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[90dvh] sm:max-h-[80dvh] w-[95dvw] xs:w-[90dvw] overflow-y-auto">
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
            <TabsList className="w-full justify-center sm:justify-start mb-4 overflow-x-auto">
              {Object.entries(timeRanges).map(([key, range]) => (
                <TabsTrigger
                  key={key}
                  value={key}
                  className="min-w-fit text-sm px-2 py-1 sm:px-3"
                >
                  {range.label}
                </TabsTrigger>
              ))}
            </TabsList>

            {Object.keys(timeRanges).map((key) => (
              <TabsContent
                key={key}
                value={key}
                className="flex-grow overflow-auto min-h-[400px] max-h-[450px] transition-opacity duration-200"
              >
                {loading ? (
                  <div className="rounded-md border overflow-auto">
                    <Table>
                      <TableBody>
                        {Array(pageSize)
                          .fill(0)
                          .map((_, index) => (
                            <TableRow key={index}>
                              <TableCell className="w-14 sm:w-20 p-1 sm:p-2 hidden sm:table-cell">
                                <Skeleton className="w-12 h-16 sm:w-16 sm:h-20 rounded-sm" />
                              </TableCell>
                              <TableCell className="align-top py-2 px-2 sm:px-4">
                                <div className="flex flex-col space-y-2">
                                  <Skeleton className="h-4 sm:h-5 w-[96px] sm:w-[160px]" />
                                  <Skeleton className="h-3 sm:h-4 w-[150px] sm:w-[500px]" />
                                </div>
                              </TableCell>
                              <TableCell className="whitespace-nowrap text-center w-20 sm:w-40 text-xs sm:text-sm px-1 sm:px-4">
                                <Skeleton className="h-3 sm:h-4 w-[64px] sm:w-[96px] ml-auto" />
                              </TableCell>
                              <TableCell className="w-24 sm:w-32 text-center whitespace-nowrap px-1 sm:px-4">
                                <div className="ml-auto">
                                  <Skeleton className="h-5 sm:h-6 w-[48px] sm:w-[64px] ml-auto" />
                                </div>
                              </TableCell>
                            </TableRow>
                          ))}
                      </TableBody>
                    </Table>
                  </div>
                ) : torrents.length === 0 ? (
                  <div className="flex items-center justify-center py-8 sm:py-12 text-center px-4">
                    <img
                      src="/logo.png"
                      alt="暂无更新"
                      onError={(e) => {
                        // 图片加载失败时使用emoji作为备用
                        const target = e.target as HTMLImageElement;
                        target.onerror = null; // 防止循环触发错误
                        target.style.display = "none";
                      }}
                      className="w-32 sm:h-32 mb-4"
                    />
                    <div className="flex flex-col items-center justify-center">
                      <div className="text-lg sm:text-xl font-bold text-primary mt-2 sm:mt-4">
                        现在还没有更新哦
                      </div>
                      <div className="text-sm sm:text-base text-secondary-foreground mt-1 sm:mt-2">
                        稍后再来看看吧~
                      </div>
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
                              <TableCell className="align-top py-2 px-1 sm:px-4">
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
                                    <HybridTooltip>
                                      <HybridTooltipTrigger asChild>
                                        <div className="text-xs sm:text-sm text-muted-foreground mt-1 truncate max-w-[150px] sm:max-w-[500px] cursor-default">
                                          <TruncatedText
                                            text={torrent.rssItem}
                                          />
                                        </div>
                                      </HybridTooltipTrigger>
                                      <HybridTooltipContent
                                        side="top"
                                        className="max-w-[250px] sm:max-w-md"
                                      >
                                        <p className="sm:max-w-md break-words text-xs sm:text-sm">
                                          {torrent.rssItem}
                                        </p>
                                      </HybridTooltipContent>
                                    </HybridTooltip>
                                  </TooltipProvider>
                                </div>
                              </TableCell>
                              <TableCell className="whitespace-nowrap text-center w-20 sm:w-40 text-xs sm:text-sm px-1 sm:px-4">
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
                              <TableCell className="w-24 sm:w-32 text-center whitespace-nowrap px-0 xs:px-1 sm:px-4">
                                {torrent.statusDetail ? (
                                  <TooltipProvider>
                                    <HybridTooltip>
                                      <HybridTooltipTrigger asChild>
                                        <div className="inline-block whitespace-nowrap">
                                          {renderStatus(torrent.status)}
                                        </div>
                                      </HybridTooltipTrigger>
                                      <HybridTooltipContent
                                        side="bottom"
                                        className="max-w-[250px] sm:max-w-[150px] break-words whitespace-normal"
                                      >
                                        <p className="text-xs sm:text-sm">
                                          {torrent.statusDetail}
                                        </p>
                                      </HybridTooltipContent>
                                    </HybridTooltip>
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
