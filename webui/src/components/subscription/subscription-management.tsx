import type React from "react";

import { cn } from "@/lib/utils";

import { useState, useEffect, useCallback } from "react";
import {
  Plus,
  Calendar,
  RefreshCw,
  Sparkles,
  Clock,
  Filter,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/useToast";

import { AddSubscriptionDialog } from "./add-subscription-dialog";
import { SubscriptionCalendarDialog } from "./subscription-calendar-dialog";
import { SubscriptionDetailDialog } from "./subscription-detail-dialog";
import { RecentUpdatesDialog } from "./recent-updates-dialog";
import subscriptionAPI, {
  BangumiBase,
  ReleaseGroupSubscription,
} from "@/api/subscription";
import { extractErrorMessage } from "@/utils/error";

export default function SubscriptionManagement() {
  const { toast } = useToast();
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [calendarDialogOpen, setCalendarDialogOpen] = useState(false);
  const [recentUpdatesDialogOpen, setRecentUpdatesDialogOpen] = useState(false);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedBangumi, setSelectedBangumi] = useState<BangumiBase | null>(
    null
  );
  const [selectedSubscriptionID, setSelectedSubscriptionID] =
    useState<string>("");
  const [bangumis, setBangumis] = useState<BangumiBase[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshSuccess, setRefreshSuccess] = useState(false);
  const [showActiveOnly, setShowActiveOnly] = useState(() => {
    const saved = localStorage.getItem("subscription-show-active-only");
    return saved ? JSON.parse(saved) : false;
  });

  // 获取番剧列表
  const fetchBangumis = useCallback(async () => {
    try {
      setLoading(true);
      setRefreshSuccess(false);
      const params = showActiveOnly ? { active: true } : undefined;
      const data = await subscriptionAPI.listBangumisBase(params);
      setBangumis(data);
      setRefreshSuccess(true);
      // 短暂显示刷新成功状态
      setTimeout(() => setRefreshSuccess(false), 500);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "获取番剧列表失败",
        description: description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }, [showActiveOnly]);

  // 保存过滤状态到 sessionStorage
  useEffect(() => {
    localStorage.setItem(
      "subscription-show-active-only",
      JSON.stringify(showActiveOnly)
    );
  }, [showActiveOnly]);

  // 初始化加载
  useEffect(() => {
    fetchBangumis();
  }, [fetchBangumis]);

  const handleBangumiClick = (bangumi: BangumiBase) => {
    if (bangumi.releaseGroups.length > 0) {
      if (!selectedSubscriptionID) {
        setSelectedSubscriptionID(bangumi.releaseGroups[0].subscriptionID);
      }
      setSelectedBangumi(bangumi);
      setDetailDialogOpen(true);
    } else {
      // 正常不会走到这里，都会有订阅信息
      toast({
        title: "该番剧没有订阅",
        description: "请添加订阅",
        variant: "destructive",
      });
    }
  };

  const handleReleaseGroupClick = (
    bangumi: BangumiBase,
    releaseGroup: ReleaseGroupSubscription,
    e: React.MouseEvent
  ) => {
    e.stopPropagation();
    setSelectedBangumi(bangumi);
    setSelectedSubscriptionID(releaseGroup.subscriptionID);
    setDetailDialogOpen(true);
  };

  const handleDetailClose = () => {
    setDetailDialogOpen(false);
    setSelectedBangumi(null);
    setSelectedSubscriptionID("");
    fetchBangumis();
  };

  const addBreakOpportunities = (str: string) => {
    // 使用零宽空格字符，允许在 & 符号处换行
    return str.replace(/&/g, "\u200B&");
  };

  return (
    <div className="flex flex-col h-[calc(100dvh-4rem)] hide-scrollbar">
      <div className="flex-none space-y-6 pb-6">
        <div className="flex flex-row items-center justify-between gap-1.5 xs:gap-4">
          <div className="flex items-center gap-1.5 xs:gap-3 flex-shrink-0">
            <h1 className="text-xl xs:text-2xl sm:text-3xl font-bold anime-gradient-text flex items-center gap-1.5 xs:gap-2">
              <Sparkles className="h-4 w-4 xs:h-5 xs:w-5 sm:h-6 sm:w-6 text-primary animate-pulse flex-shrink-0" />
              <span className="flex flex-row whitespace-nowrap">订阅管理</span>
            </h1>
            <Button
              variant={showActiveOnly ? "default" : "outline"}
              size="sm"
              className={cn(
                "rounded-xl anime-button transition-all duration-300 h-7 w-7 xs:h-8 xs:w-auto xs:px-3 p-0 xs:p-2 flex-shrink-0",
                showActiveOnly &&
                  "bg-gradient-to-r from-primary to-blue-500 hover:opacity-90 shadow-md scale-105"
              )}
              onClick={() => setShowActiveOnly(!showActiveOnly)}
            >
              <Filter
                className={cn(
                  "h-3.5 w-3.5 xs:h-4 xs:w-4 transition-transform duration-300",
                  showActiveOnly && "text-white"
                )}
              />
              <span className="hidden sm:inline ml-1 text-xs whitespace-nowrap">
                只看启用
              </span>
            </Button>
          </div>
          <div className="flex gap-1.5 xs:gap-2 justify-end flex-shrink-0">
            <Button
              className="rounded-xl anime-button bg-gradient-to-r from-primary to-blue-500 hover:opacity-90 p-2 sm:px-3 sm:py-2 aspect-square sm:aspect-auto"
              onClick={() => setAddDialogOpen(true)}
            >
              <Plus className="icon-button" />
              <span className="hidden sm:inline">添加订阅</span>
            </Button>
            <AddSubscriptionDialog
              open={addDialogOpen}
              onOpenChange={setAddDialogOpen}
              onSubscribed={fetchBangumis}
            />

            <Button
              variant="outline"
              className="rounded-xl anime-button p-2 sm:px-3 sm:py-2 aspect-square sm:aspect-auto"
              onClick={() => setCalendarDialogOpen(true)}
            >
              <Calendar className="icon-button" />
              <span className="hidden sm:inline">订阅日历</span>
            </Button>
            <SubscriptionCalendarDialog
              open={calendarDialogOpen}
              onOpenChange={setCalendarDialogOpen}
            />

            <Button
              variant="outline"
              className="rounded-xl anime-button p-2 sm:px-3 sm:py-2 aspect-square sm:aspect-auto"
              onClick={() => setRecentUpdatesDialogOpen(true)}
            >
              <Clock className="icon-button" />
              <span className="hidden sm:inline">近期更新</span>
            </Button>
            <RecentUpdatesDialog
              open={recentUpdatesDialogOpen}
              onOpenChange={setRecentUpdatesDialogOpen}
            />

            <Button
              variant="outline"
              className={cn(
                "rounded-xl anime-button p-2 sm:px-3 sm:py-2 aspect-square sm:aspect-auto",
                refreshSuccess &&
                  "bg-green-100 border-green-500 text-green-600 transition-colors duration-500"
              )}
              onClick={fetchBangumis}
              disabled={loading}
            >
              <RefreshCw
                className={cn(
                  "icon-button",
                  loading && "animate-spin",
                  refreshSuccess &&
                    "text-green-500 animate-[spin_1s_ease-in-out]"
                )}
              />
              <span className="hidden sm:inline">
                {refreshSuccess ? "刷新成功" : "刷新"}
              </span>
            </Button>
          </div>
        </div>
      </div>
      <div
        className={cn(
          "flex-grow border rounded-xl bg-background/50 overflow-auto",
          refreshSuccess && "animate-[pulse_0.5s_ease-in-out]"
        )}
      >
        <div className="flex flex-wrap gap-4 p-4 pb-10">
          {bangumis.map((bangumi) => (
            <div
              key={bangumi.bangumiName + bangumi.season}
              className="relative"
            >
              <Card
                className={cn(
                  "anime-card overflow-hidden border-primary/10 cursor-pointer transition-all duration-300 hover:scale-105 w-[130px] xs:w-[160px]",
                  selectedBangumi?.bangumiName === bangumi.bangumiName &&
                    selectedBangumi?.season === bangumi.season &&
                    "scale-105 border-primary/30 shadow-lg"
                )}
                onClick={() => handleBangumiClick(bangumi)}
              >
                <div className="flex flex-col h-full">
                  <div className="relative w-full h-[195px] xs:h-[240px] overflow-hidden group">
                    <img
                      src={
                        bangumi.posterURL ||
                        "/placeholder.svg?height=300&width=200"
                      }
                      alt={bangumi.bangumiName}
                      style={{ width: "100%", height: "100%" }}
                      className={cn(
                        "object-cover absolute top-0 left-0 transition-transform duration-500 group-hover:scale-110",
                        !bangumi.releaseGroups.some((group) => group.active) &&
                          "filter grayscale"
                      )}
                    />
                    <div className="absolute top-2 right-2 z-20">
                      <Badge className="rounded-full bg-primary/80 px-2 py-1 text-xs font-bold">
                        S{bangumi.season}
                      </Badge>
                    </div>

                    {!bangumi.releaseGroups.some((group) => group.active) && (
                      <div className="absolute bottom-2 left-2 z-20">
                        <Badge
                          variant="outline"
                          className="rounded-full bg-background/70 px-2 text-xs"
                        >
                          已禁用
                        </Badge>
                      </div>
                    )}

                    <div className="absolute inset-0 bg-black/70 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-center p-4 overflow-y-auto">
                      <div className="space-y-2">
                        {bangumi.releaseGroups.map((group, index) => (
                          <button
                            key={index}
                            onClick={(e) =>
                              handleReleaseGroupClick(bangumi, group, e)
                            }
                            className={cn(
                              "rounded-lg px-2 py-0.5 text-xs transition-all duration-200 w-full break-words whitespace-normal",
                              selectedBangumi?.bangumiName ===
                                bangumi.bangumiName &&
                                selectedBangumi?.season === bangumi.season &&
                                selectedSubscriptionID === group.subscriptionID
                                ? group.active
                                  ? "bg-primary/20 text-primary border-2 border-primary font-medium shadow-sm"
                                  : "bg-white text-gray-700 border-2 border-blue-300 font-medium shadow-sm"
                                : group.active
                                ? "bg-primary/10 text-primary border border-primary/20"
                                : "bg-white text-muted-foreground border border-muted"
                            )}
                          >
                            {addBreakOpportunities(group.releaseGroup)} (
                            {group.lastAirEpisode}/{group.episodeTotalNum})
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>

                  <CardContent className="p-2">
                    <h3 className="font-semibold text-sm line-clamp-1">
                      {bangumi.bangumiName}
                    </h3>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      第 {bangumi.season} 季
                    </p>
                  </CardContent>
                </div>
              </Card>
            </div>
          ))}
        </div>
      </div>

      <SubscriptionDetailDialog
        open={detailDialogOpen}
        onOpenChange={setDetailDialogOpen}
        selectedSubscriptionID={selectedSubscriptionID}
        releaseGroups={selectedBangumi?.releaseGroups || []}
        onClose={handleDetailClose}
      />
    </div>
  );
}
