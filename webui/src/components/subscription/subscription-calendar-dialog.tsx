import { useState, useEffect } from "react";
import { RefreshCw, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { subscriptionAPI, CalendarItem } from "@/api/subscription";
import { toast } from "@/components/ui/use-toast";
import { getWeekDayText, getSortedWeekDays } from "@/utils/weekday";

export interface SubscriptionCalendarDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// 获取排序后的星期数组，从周一开始
const getSortedWeekdayNumbers = (): number[] => {
  return getSortedWeekDays().map(([day]) => parseInt(day));
};

export function SubscriptionCalendarDialog({
  open,
  onOpenChange,
}: SubscriptionCalendarDialogProps) {
  const [calendarData, setCalendarData] = useState<
    Record<number, CalendarItem[]>
  >({});
  const [loading, setLoading] = useState(false);

  // 获取订阅日历数据
  const fetchCalendarData = async () => {
    if (!open) return;

    try {
      setLoading(true);
      const response = await subscriptionAPI.getSubscriptionCalendar();
      setCalendarData(response);
    } catch (error) {
      toast({
        title: "获取订阅日历失败",
        description: String(error),
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCalendarData();
  }, [open]);

  // 获取排序后的星期数组
  const sortedWeekdays = getSortedWeekdayNumbers();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl rounded-xl border-primary/20 bg-card/95 backdrop-blur-md">
        <DialogHeader>
          <DialogTitle className="text-xl anime-gradient-text">
            订阅日历
          </DialogTitle>
          <DialogDescription>按星期查看番剧更新时间表</DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="flex justify-center items-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : (
          <div className="grid grid-cols-7 gap-4 py-4">
            {sortedWeekdays.map((weekday) => (
              <div key={weekday} className="space-y-4">
                <h3 className="text-center font-medium">
                  {getWeekDayText(weekday)}
                </h3>
                <div className="grid grid-cols-1 gap-6">
                  {calendarData[weekday]?.map((anime, index) => (
                    <div
                      key={index}
                      className="flex flex-col items-center group"
                    >
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className="relative w-16 h-24 overflow-hidden rounded-xl animate-float anime-glow">
                              <img
                                src={anime.posterURL || "/placeholder.svg"}
                                alt={anime.bangumiName}
                                style={{ width: "100%", height: "100%" }}
                                className="object-cover group-hover:scale-105 transition-transform duration-200"
                              />
                              <div className="absolute top-0 right-0 bg-primary/80 backdrop-blur-sm text-primary-foreground text-[10px] font-semibold px-1.5 py-0.5 rounded-bl-md">
                                S{anime.season}
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>{anime.bangumiName}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>

                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className="mt-2 text-center w-full">
                              <span className="text-xs block truncate">
                                {anime.bangumiName}
                              </span>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>{anime.bangumiName}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                  ))}
                  {(!calendarData[weekday] ||
                    calendarData[weekday].length === 0) && (
                    <div className="flex justify-center items-center h-24 text-muted-foreground text-sm">
                      暂无更新
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}

        <div className="flex justify-end">
          <Button
            variant="outline"
            className="rounded-xl"
            onClick={fetchCalendarData}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="mr-2 h-4 w-4" />
            )}
            刷新
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
