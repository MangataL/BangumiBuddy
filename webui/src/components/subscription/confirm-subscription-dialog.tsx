import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import {
  Info,
  Filter,
  Settings,
  Tv,
  Calendar,
  Hash,
  Users,
  Flag,
  MapPin,
  Clock,
  ArrowUpCircle,
  Search,
  X,
  Loader2,
  RefreshCw,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import {
  subscriptionAPI,
  type RSSMatch,
  type ParseRSSResponse,
  type SubscribeRequest,
} from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { MatchInput } from "../common/match-input";
import { EpisodePositionInput } from "./episode-position-input";
import { TMDBInput } from "@/components/tmdb";
import { getSortedWeekDays, formatDate, isValidWeekDay } from "@/utils/time";
import { extractErrorMessage } from "@/utils/error";
import { cn } from "@/lib/utils";
import { useParseSubscription } from "./use-parse-subscription";

interface ConfirmSubscriptionDialogProps {
  onClose: () => void;
  parseRSSRsp: ParseRSSResponse;
  defaultPriority?: number;
  onSubscribed?: () => void;
}

type SubscriptionForm = Omit<SubscribeRequest, "airWeekday"> & {
  airWeekday: number | null;
};

function positiveInteger(value: number) {
  return Number.isSafeInteger(value) && value > 0;
}

function regularSeasons(result?: ParseRSSResponse) {
  return (result?.seasons ?? []).filter((season) =>
    positiveInteger(season.number)
  );
}

function parsedSeasonFields(result: ParseRSSResponse) {
  const seasons = regularSeasons(result);
  const validSeason =
    positiveInteger(result.season) &&
    (seasons.length === 0 ||
      seasons.some((season) => season.number === result.season));
  return {
    tmdbID: result.tmdbID,
    season: validSeason ? result.season : 0,
    episodeTotalNum: validSeason ? result.episodeTotalNum : 0,
    airWeekday: validSeason ? result.airWeekday : null,
  };
}

export function ConfirmSubscriptionDialog({
  onClose,
  parseRSSRsp,
  defaultPriority = 1,
  onSubscribed,
}: ConfirmSubscriptionDialogProps) {
  const { toast } = useToast();
  const [metadata, setMetadata] = useState<ParseRSSResponse | undefined>(
    parseRSSRsp
  );
  const [subscription, setSubscription] = useState<SubscriptionForm>(() => ({
    ...parsedSeasonFields(parseRSSRsp),
    rssLink: parseRSSRsp.rssLink,
    releaseGroup: parseRSSRsp.releaseGroup,
    includeRegs: [],
    excludeRegs: [],
    episodeOffset: 0,
    priority: Math.max(1, defaultPriority),
    episodeLocation: "",
  }));
  const [submitting, setSubmitting] = useState(false);
  const { parse, cancel, loading } = useParseSubscription((result) => {
    setMetadata(result);
    setSubscription((prev) => ({ ...prev, ...parsedSeasonFields(result) }));
  });
  const seasons = regularSeasons(metadata);
  const selectedSeason = seasons.find(
    (season) => season.number === subscription.season
  );
  const posterURL = selectedSeason
    ? selectedSeason.posterURL
    : metadata?.posterURL;
  const year = selectedSeason
    ? selectedSeason.airDate.slice(0, 4)
    : metadata?.year;
  const validSeason =
    positiveInteger(subscription.season) &&
    (seasons.length === 0 || Boolean(selectedSeason));
  const validEpisodes = positiveInteger(subscription.episodeTotalNum);
  const validGroup = Boolean(subscription.releaseGroup.trim());
  const validID =
    positiveInteger(subscription.tmdbID) &&
    metadata?.tmdbID === subscription.tmdbID;
  const canSubscribe =
    validID &&
    validSeason &&
    validEpisodes &&
    validGroup &&
    Boolean(subscription.rssLink.trim()) &&
    Number.isSafeInteger(subscription.priority) &&
    Number.isSafeInteger(subscription.episodeOffset) &&
    subscription.airWeekday !== null &&
    isValidWeekDay(subscription.airWeekday) &&
    !loading &&
    !submitting;

  const [previewMatches, setPreviewMatches] = useState<RSSMatch[]>([]);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);

  const previewRequest = useRef(0);
  const { rssLink, includeRegs, excludeRegs } = subscription;
  const handlePreview = useCallback(async () => {
    if (!rssLink) return;
    const request = ++previewRequest.current;
    setIsPreviewLoading(true);
    try {
      const matches = await subscriptionAPI.previewRSSMatch({
        rssLink,
        includeRegs,
        excludeRegs,
      });
      if (request === previewRequest.current) setPreviewMatches(matches);
    } catch (error) {
      if (request === previewRequest.current) {
        toast({
          title: "预览失败",
          description: extractErrorMessage(error),
          variant: "destructive",
        });
      }
    } finally {
      if (request === previewRequest.current) setIsPreviewLoading(false);
    }
  }, [rssLink, includeRegs, excludeRegs, toast]);

  useEffect(() => {
    const timer = setTimeout(handlePreview, 500);
    return () => {
      clearTimeout(timer);
      previewRequest.current++;
    };
  }, [handlePreview]);

  const handleSubscribe = async () => {
    const { airWeekday } = subscription;
    if (!canSubscribe || airWeekday === null) return;
    setSubmitting(true);
    try {
      await subscriptionAPI.subscribe({ ...subscription, airWeekday });
      onClose();
      toast({ title: "订阅成功" });
      onSubscribed?.();
    } catch (error) {
      toast({
        title: "订阅失败",
        description: extractErrorMessage(error),
        variant: "destructive",
      });
    } finally {
      setSubmitting(false);
    }
  };

  function updateSubscription<K extends keyof SubscriptionForm>(
    field: K,
    value: SubscriptionForm[K]
  ) {
    setSubscription((prev) => ({ ...prev, [field]: value }));
  }

  const changeTMDBID = (tmdbID: number) => {
    cancel();
    setMetadata(undefined);
    setSubscription((prev) => ({
      ...prev,
      tmdbID,
      season: 0,
      episodeTotalNum: 0,
      airWeekday: null,
    }));
  };

  const confirmTMDBID = (tmdbID: number) => {
    changeTMDBID(tmdbID);
    void parse(rssLink, tmdbID);
  };

  return (
    <Dialog
      open
      onOpenChange={() => {
        if (!submitting) onClose();
      }}
    >
      <DialogContent className="max-w-lg w-[95dvw] max-h-[90dvh] overflow-y-auto rounded-2xl border-primary/20 bg-card/95 backdrop-blur-md sm:max-w-lg scrollbar-hide p-0 [&>button:last-child]:hidden">
        <div className="relative w-full overflow-hidden rounded-t-2xl">
          {/* 背景图 */}
          <div
            className="absolute inset-0 bg-cover bg-center"
            style={{
              backgroundImage: `url(${
                metadata?.backdropURL || posterURL || ""
              })`,
              backgroundPosition: "center",
              filter: "blur(4px) brightness(0.5)",
              transform: "scale(1.1)",
            }}
          />
          {/* 渐变遮罩 */}
          <div className="absolute inset-0 bg-gradient-to-b from-black/40 via-transparent to-card" />

          <div className="relative p-6 pt-8">
            <div className="flex justify-between items-start mb-4">
              <div className="space-y-1">
                <DialogTitle className="text-xl font-bold text-white/95">
                  订阅确认
                </DialogTitle>
                <DialogDescription className="text-white/60 text-xs">
                  请核对并完善番剧订阅信息
                </DialogDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 rounded-full border border-white/20 bg-black/10 text-white/70 hover:bg-black/30 hover:text-white backdrop-blur-md transition-all -mr-2 -mt-2"
                onClick={onClose}
                aria-label="关闭订阅确认"
                disabled={submitting}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="flex gap-5 mt-4">
              {/* 海报 */}
              <div className="relative w-28 h-40 flex-shrink-0 rounded-xl bg-card/20 backdrop-blur-sm border-2 border-white/20 overflow-hidden shadow-2xl">
                {posterURL ? (
                  <img
                    src={posterURL}
                    className="w-full h-full object-cover"
                    alt={metadata?.name ?? ""}
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-white/40">
                    <Tv className="w-10 h-10" />
                  </div>
                )}
              </div>

              {/* 番剧简要信息 */}
              <div className="flex-1 flex flex-col justify-end pb-1">
                <h3 className="text-2xl font-bold text-white mb-3 line-clamp-2 leading-tight drop-shadow-md">
                  {metadata?.name ?? ""}
                </h3>
                <div className="flex flex-wrap gap-2">
                  <span className="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold bg-white/20 text-white backdrop-blur-md border border-white/10">
                    {year || "未知年份"}
                  </span>
                  <span className="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold bg-primary/40 text-white backdrop-blur-md border border-white/10">
                    {validSeason ? `第 ${subscription.season} 季` : "待选季"}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="px-6 py-4">
          <Tabs
            defaultValue="basic"
            className="w-full"
            onValueChange={(val) => {
              if (val === "filter") {
                handlePreview();
              }
            }}
          >
            <TabsList className="grid w-full grid-cols-3 mb-6 bg-muted/50 p-1 rounded-xl">
              <TabsTrigger
                value="basic"
                className="rounded-lg data-[state=active]:bg-background data-[state=active]:shadow-sm"
              >
                <Info className="w-4 h-4 mr-2 text-primary" />
                <span>基础信息</span>
              </TabsTrigger>
              <TabsTrigger
                value="filter"
                className="rounded-lg data-[state=active]:bg-background data-[state=active]:shadow-sm"
              >
                <Filter className="w-4 h-4 mr-2 text-blue-500" />
                <span>过滤规则</span>
              </TabsTrigger>
              <TabsTrigger
                value="advanced"
                className="rounded-lg data-[state=active]:bg-background data-[state=active]:shadow-sm"
              >
                <Settings className="w-4 h-4 mr-2 text-orange-500" />
                <span>高级设置</span>
              </TabsTrigger>
            </TabsList>

            <TabsContent value="basic" className="space-y-4 mt-0 outline-none">
              <div className="grid gap-4">
                <div className="grid gap-2">
                  <Label className="flex items-center gap-2 text-muted-foreground mb-1">
                    <Tv className="w-4 h-4" /> 番剧名称
                  </Label>
                  <Input
                    value={metadata?.name ?? ""}
                    readOnly
                    className="rounded-xl bg-muted/30 border-muted-foreground/10 cursor-default focus-visible:ring-0"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label
                      htmlFor="season"
                      className="flex items-center gap-2 text-muted-foreground mb-1"
                    >
                      <Clock className="w-4 h-4" /> 季度
                    </Label>
                    {seasons.length > 0 ? (
                      <Select
                        value={selectedSeason?.number.toString() ?? ""}
                        onValueChange={(value) => {
                          const season = seasons.find(
                            (item) => item.number === Number(value)
                          );
                          if (season) {
                            setSubscription((prev) => ({
                              ...prev,
                              season: season.number,
                              episodeTotalNum: season.episodeTotalNum,
                              airWeekday: season.airWeekday,
                            }));
                          }
                        }}
                      >
                        <SelectTrigger
                          id="season"
                          aria-invalid={!validSeason}
                          className="rounded-xl bg-muted/30 border-muted-foreground/10"
                        >
                          <SelectValue placeholder="请选择季度" />
                        </SelectTrigger>
                        <SelectContent>
                          {seasons.map((season) => (
                            <SelectItem
                              key={season.number}
                              value={String(season.number)}
                            >
                              第 {season.number} 季
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <Input
                        id="season"
                        type="number"
                        min="1"
                        placeholder={
                          metadata ? "请输入季号" : "请先确认 TMDB ID"
                        }
                        disabled={!metadata}
                        value={subscription.season || ""}
                        onChange={(e) =>
                          updateSubscription("season", Number(e.target.value))
                        }
                        aria-invalid={!validSeason}
                        className="rounded-xl bg-muted/30 border-muted-foreground/10"
                      />
                    )}
                  </div>
                  <div className="grid gap-2">
                    <Label
                      htmlFor="episodeTotalNum"
                      className="flex items-center gap-2 text-muted-foreground mb-1"
                    >
                      <Hash className="w-4 h-4" /> 总集数
                    </Label>
                    <Input
                      id="episodeTotalNum"
                      type="number"
                      min="1"
                      value={subscription.episodeTotalNum || ""}
                      disabled={
                        !metadata || (seasons.length > 0 && !selectedSeason)
                      }
                      aria-invalid={!validEpisodes}
                      onChange={(e) =>
                        updateSubscription(
                          "episodeTotalNum",
                          Number(e.target.value)
                        )
                      }
                      className={cn(
                        "rounded-xl bg-muted/30 border-muted-foreground/10",
                        !validEpisodes &&
                          "border-destructive focus-visible:ring-destructive"
                      )}
                    />
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label
                      htmlFor="airWeekDay"
                      className="flex items-center gap-2 text-muted-foreground mb-1"
                    >
                      <Calendar className="w-4 h-4" /> 更新时间
                    </Label>
                    <Select
                      value={subscription.airWeekday?.toString() ?? ""}
                      onValueChange={(value) =>
                        updateSubscription("airWeekday", parseInt(value))
                      }
                    >
                      <SelectTrigger
                        id="airWeekDay"
                        className="rounded-xl bg-muted/30 border-muted-foreground/10"
                      >
                        <SelectValue placeholder="请选择更新星期" />
                      </SelectTrigger>
                      <SelectContent>
                        {getSortedWeekDays().map(([value, label]) => (
                          <SelectItem key={value} value={value}>
                            {label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label
                      htmlFor="subgroup"
                      className="flex items-center gap-2 text-muted-foreground mb-1"
                    >
                      <Users className="w-4 h-4" /> 字幕组
                    </Label>
                    <Input
                      id="subgroup"
                      value={subscription.releaseGroup}
                      onChange={(e) =>
                        updateSubscription("releaseGroup", e.target.value)
                      }
                      placeholder="例如: VCB-Studio"
                      className={cn(
                        "rounded-xl bg-muted/30 border-muted-foreground/10",
                        !validGroup &&
                          "border-destructive focus-visible:ring-destructive"
                      )}
                    />
                  </div>
                </div>

                <Separator className="my-2 opacity-50" />

                <TMDBInput
                  type="tv"
                  value={subscription.tmdbID}
                  onTMDBIDChange={changeTMDBID}
                  onConfirmID={confirmTMDBID}
                  resolving={loading}
                  label="TMDB ID"
                  icon={<Search className="w-4 h-4" />}
                  className="mb-2"
                  error={
                    !metadata && !loading
                      ? "请确认 TMDB ID 以加载季度"
                      : undefined
                  }
                />
              </div>
            </TabsContent>

            <TabsContent value="filter" className="space-y-4 mt-0 outline-none">
              <MatchInput
                label="包含匹配"
                items={subscription.includeRegs}
                placeholder="添加包含匹配条件"
                onChange={(items) => updateSubscription("includeRegs", items)}
              />
              <MatchInput
                label="排除匹配"
                items={subscription.excludeRegs}
                placeholder="添加排除匹配条件"
                onChange={(items) => updateSubscription("excludeRegs", items)}
              />

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label className="flex items-center gap-2 text-muted-foreground">
                    <Filter className="w-4 h-4" /> 预览结果
                  </Label>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handlePreview}
                    disabled={isPreviewLoading}
                    className="h-7 text-xs"
                  >
                    {isPreviewLoading ? (
                      <Loader2 className="w-3 h-3 animate-spin mr-1" />
                    ) : (
                      <RefreshCw className="w-3 h-3 mr-1" />
                    )}
                    <span className="hidden sm:inline">刷新预览</span>
                  </Button>
                </div>
                <div className="rounded-xl border border-muted-foreground/10 bg-muted/30 p-2 h-64 overflow-y-auto scrollbar-thin scrollbar-thumb-primary/10 scrollbar-track-transparent">
                  <div className="flex flex-col gap-2">
                    {previewMatches.map((m, i) => (
                      <div
                        key={i}
                        className={cn(
                          "relative overflow-hidden rounded-lg border p-3 text-sm transition-all duration-200",
                          m.match
                            ? "bg-green-500/5 border-green-500/20 hover:border-green-500/40 hover:bg-green-500/10"
                            : "bg-background/40 border-transparent opacity-60 hover:opacity-100 hover:bg-background/60"
                        )}
                      >
                        {/* 匹配状态指示条 */}
                        {m.match && (
                          <div className="absolute left-0 top-0 bottom-0 w-1 bg-green-500" />
                        )}

                        <div
                          className={cn(
                            "flex flex-col gap-2",
                            m.match ? "pl-2" : ""
                          )}
                        >
                          {/* 标题 - 自动换行，最多显示2行 */}
                          <div className="font-medium leading-relaxed break-all line-clamp-2 text-foreground/90">
                            {m.guid}
                          </div>

                          {/* 底部信息栏 */}
                          <div className="flex items-center justify-between mt-1">
                            <span className="text-xs text-muted-foreground font-mono">
                              {formatDate(m.publishedAt)}
                            </span>
                            {m.match ? (
                              <CheckCircle2 className="w-5 h-5 text-green-500" />
                            ) : (
                              <XCircle className="w-5 h-5 text-muted-foreground/30" />
                            )}
                          </div>
                        </div>
                      </div>
                    ))}

                    {previewMatches.length === 0 && !isPreviewLoading && (
                      <div className="h-48 flex flex-col items-center justify-center text-muted-foreground gap-3">
                        <div className="p-3 rounded-full bg-muted/50">
                          <Search className="w-6 h-6 opacity-40" />
                        </div>
                        <span className="text-sm opacity-60">
                          点击右上角刷新预览
                        </span>
                      </div>
                    )}

                    {isPreviewLoading && previewMatches.length === 0 && (
                      <div className="h-48 flex flex-col items-center justify-center text-muted-foreground gap-3">
                        <Loader2 className="w-8 h-8 animate-spin text-primary/60" />
                        <span className="text-sm opacity-60">
                          正在获取 RSS 数据...
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </TabsContent>

            <TabsContent
              value="advanced"
              className="space-y-4 mt-0 outline-none"
            >
              <div className="grid gap-6">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label
                      htmlFor="priority"
                      className="flex items-center gap-2 text-muted-foreground mb-1"
                    >
                      <Flag className="w-4 h-4" /> 优先级
                    </Label>
                    <Input
                      id="priority"
                      type="number"
                      value={
                        Number.isFinite(subscription.priority)
                          ? subscription.priority
                          : ""
                      }
                      onChange={(e) =>
                        updateSubscription("priority", e.target.valueAsNumber)
                      }
                      className="rounded-xl bg-muted/30 border-muted-foreground/10"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label
                      htmlFor="episodeOffset"
                      className="flex items-center gap-2 text-muted-foreground mb-1"
                    >
                      <ArrowUpCircle className="w-4 h-4" /> 集数偏移
                    </Label>
                    <Input
                      id="episodeOffset"
                      type="number"
                      value={
                        Number.isFinite(subscription.episodeOffset)
                          ? subscription.episodeOffset
                          : ""
                      }
                      onChange={(e) =>
                        updateSubscription(
                          "episodeOffset",
                          e.target.valueAsNumber
                        )
                      }
                      className="rounded-xl bg-muted/30 border-muted-foreground/10"
                    />
                  </div>
                </div>

                <EpisodePositionInput
                  value={subscription.episodeLocation}
                  onChange={(value) =>
                    updateSubscription("episodeLocation", value)
                  }
                  icon={<MapPin className="w-4 h-4" />}
                  inputClassName="bg-muted/30 border-muted-foreground/10"
                />
              </div>
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter className="px-6 py-4 bg-muted/30 gap-2 sm:gap-0 border-t border-muted-foreground/5 rounded-b-2xl">
          <Button
            variant="ghost"
            onClick={onClose}
            disabled={submitting}
            className="rounded-xl hover:bg-muted"
          >
            取消
          </Button>
          <Button
            disabled={!canSubscribe}
            onClick={handleSubscribe}
            className="rounded-xl px-8 bg-gradient-to-r from-primary to-blue-500 hover:opacity-90 transition-all shadow-md shadow-primary/20"
          >
            确认订阅
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
