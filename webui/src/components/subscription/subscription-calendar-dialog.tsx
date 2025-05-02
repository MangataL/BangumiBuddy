import { useState, useEffect, useRef } from "react";
import { ChevronLeft, ChevronRight, Loader2 } from "lucide-react";
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
import { getWeekDayText, getSortedWeekDays } from "@/utils/time";
import { useMobile } from "@/hooks/useMobile";

export interface SubscriptionCalendarDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// 获取排序后的星期数组，从周一开始
const getSortedWeekdayNumbers = (): number[] => {
  return getSortedWeekDays().map(([day]) => parseInt(day));
};

// 获取当前星期几（0-6，0表示周日）
const getCurrentWeekday = (): number => {
  const date = new Date();
  return date.getDay();
};

export function SubscriptionCalendarDialog({
  open,
  onOpenChange,
}: SubscriptionCalendarDialogProps) {
  const [calendarData, setCalendarData] = useState<
    Record<number, CalendarItem[]>
  >({});
  const [loading, setLoading] = useState(false);

  // 是否为移动端
  const isMobile = useMobile();

  // 每页显示的星期数
  const weeksPerPage = isMobile ? 1 : 4;

  // 获取排序后的星期数组
  const sortedWeekdays = getSortedWeekdayNumbers();

  const [currentVisibleWeekdays, setCurrentVisibleWeekdays] = useState<
    number[]
  >([]);
  const currentDay = useRef(0);

  // 获取订阅日历数据
  const fetchCalendarData = async () => {
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
    if (!open) return;
    fetchCalendarData();
    if (isMobile) {
      // 移动端：从当天开始
      const today = getCurrentWeekday();
      const todayIndex = sortedWeekdays.findIndex((day) => day === today);
      currentDay.current = todayIndex;
      setCurrentVisibleWeekdays(visibleWeekdays(todayIndex));
    } else {
      // 桌面端：从前一天开始
      const today = getCurrentWeekday();
      const yesterdayIndex = sortedWeekdays.findIndex(
        (day) => day === (today === 0 ? 6 : today - 1)
      );
      currentDay.current = yesterdayIndex;
      setCurrentVisibleWeekdays(visibleWeekdays(yesterdayIndex));
    }
  }, [open]);

  // 计算当前显示的星期数组
  const visibleWeekdays = (startIndex: number) => {
    const result = [];
    for (let i = 0; i < weeksPerPage; i++) {
      const index = (startIndex + i) % sortedWeekdays.length;
      result.push(sortedWeekdays[index]);
    }
    return result;
  };

  // 前一页
  const prevPage = () => {
    const newIndex = currentDay.current - 1;
    currentDay.current =
      newIndex < 0 ? sortedWeekdays.length + newIndex : newIndex;
    setCurrentVisibleWeekdays(visibleWeekdays(currentDay.current));
  };

  // 后一页
  const nextPage = () => {
    const newIndex = currentDay.current + 1;
    currentDay.current = newIndex % sortedWeekdays.length;
    setCurrentVisibleWeekdays(visibleWeekdays(currentDay.current));
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[95dvw] max-h-[90dvh] sm:w-auto md:max-w-[95dvw] overflow-y-auto rounded-xl border-primary/20 bg-card/95 backdrop-blur-md scrollbar-hide">
        <DialogHeader className="flex flex-row justify-between items-center mb-4">
          <div>
            <DialogTitle className="text-xl anime-gradient-text">
              订阅日历
            </DialogTitle>
            <DialogDescription>按星期查看番剧更新</DialogDescription>
          </div>
        </DialogHeader>

        {loading ? (
          <div className="flex justify-center items-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : (
          <div className="relative px-10">
            <Button
              variant="ghost"
              size="icon"
              className="absolute left-0 top-1/2 -translate-y-1/2 z-10 h-10 w-10"
              onClick={prevPage}
            >
              <ChevronLeft className="h-6 w-6" />
              <span className="sr-only">上一页</span>
            </Button>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 pb-4 px-8 overflow-x-auto">
              {currentVisibleWeekdays.map((weekday) => (
                <div
                  key={`weekday-${weekday}-${currentDay.current}`}
                  className="space-y-4"
                >
                  <div className="border-b pb-2">
                    <h3 className="text-center font-medium">
                      {getWeekDayText(weekday)}
                    </h3>
                  </div>
                  <div className="space-y-4">
                    {calendarData[weekday]?.map((anime, index) => (
                      <div
                        key={index}
                        className="flex gap-3 items-center group"
                      >
                        <div className="relative w-16 h-24 overflow-hidden rounded-xl anime-glow flex-shrink-0">
                          <img
                            src={anime.posterURL}
                            alt={anime.bangumiName}
                            style={{ width: "100%", height: "100%" }}
                            className="object-cover group-hover:scale-105 transition-transform duration-200"
                          />
                          <div className="absolute top-0 right-0 bg-primary/80 backdrop-blur-sm text-primary-foreground text-[10px] font-semibold px-1.5 py-0.5 rounded-bl-md">
                            S{anime.season}
                          </div>
                        </div>
                        <div className="flex-1 overflow-hidden">
                          <p className="text-sm font-medium line-clamp-5 break-words">
                            {anime.bangumiName}
                          </p>
                        </div>
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

            <Button
              variant="ghost"
              size="icon"
              className="absolute right-0 top-1/2 -translate-y-1/2 z-10 h-10 w-10"
              onClick={nextPage}
            >
              <ChevronRight className="h-6 w-6" />
              <span className="sr-only">下一页</span>
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
