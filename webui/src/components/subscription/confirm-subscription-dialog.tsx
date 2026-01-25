import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
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
  Type,
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
import { subscriptionAPI, RSSMatch } from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { MatchInput } from "../common/match-input";
import { EpisodePositionInput } from "./episode-position-input";
import { TMDBInput } from "@/components/tmdb";
import { getSortedWeekDays, formatDate } from "@/utils/time";
import { ParseRSSResponse, SubscribeRequest } from "@/api/subscription";
import { extractErrorMessage } from "@/utils/error";
import { Meta } from "@/api/meta";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";

import { debounce } from "@/utils/debounce";
import { useCallback, useRef } from "react";

interface ConfirmSubscriptionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  parseRSSRsp: ParseRSSResponse;
  onSubscribed?: () => void;
}

// 扩展 SubscribeRequest 类型，添加表单所需的额外字段
interface BangumiFormState extends SubscribeRequest {
  name: string;
  year: string;
  backdropURL: string;
  posterURL: string;
}

export function ConfirmSubscriptionDialog({
  open,
  onOpenChange,
  parseRSSRsp,
  onSubscribed,
}: ConfirmSubscriptionDialogProps) {
  const { toast } = useToast();

  const [bangumiInfo, setBangumiInfo] = useState<BangumiFormState>({
    name: "",
    season: 0,
    year: "",
    tmdbID: 0,
    rssLink: "",
    releaseGroup: "",
    episodeTotalNum: 0,
    airWeekday: 0,
    includeRegs: [],
    excludeRegs: [],
    episodeOffset: 0,
    priority: 1,
    episodeLocation: "",
    backdropURL: "",
    posterURL: "",
  });

  // 监听 parseRSSRsp 变化
  useEffect(() => {
    if (parseRSSRsp && open) {
      setBangumiInfo((prev) => ({
        ...prev,
        name: parseRSSRsp.name,
        season: parseRSSRsp.season,
        year: parseRSSRsp.year,
        tmdbID: parseRSSRsp.tmdbID,
        rssLink: parseRSSRsp.rssLink,
        releaseGroup: parseRSSRsp.releaseGroup,
        episodeTotalNum: parseRSSRsp.episodeTotalNum,
        airWeekday: parseRSSRsp.airWeekday,
        backdropURL: parseRSSRsp.backdropURL,
        posterURL: parseRSSRsp.posterURL,
      }));
    }
  }, [open, parseRSSRsp]);

  const [fieldErrors, setFieldErrors] = useState({
    season: "",
    airWeekday: "",
    tmdbID: "",
    releaseGroup: "",
    episodeTotalNum: "",
  });

  const [previewMatches, setPreviewMatches] = useState<RSSMatch[]>([]);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);

  // 预览函数
  const preview = async (
    rssLink: string,
    includeRegs: string[],
    excludeRegs: string[]
  ) => {
    if (!rssLink) return;
    setIsPreviewLoading(true);
    try {
      const matches = await subscriptionAPI.previewRSSMatch({
        rssLink,
        includeRegs,
        excludeRegs,
      });
      setPreviewMatches(matches);
    } catch (error) {
      toast({
        title: "预览失败",
        description: extractErrorMessage(error),
        variant: "destructive",
      });
    } finally {
      setIsPreviewLoading(false);
    }
  };

  // 使用 useRef 保存防抖函数，确保每次渲染都使用同一个防抖实例
  const debouncedPreview = useRef(
    debounce(
      (rssLink: string, includeRegs: string[], excludeRegs: string[]) =>
        preview(rssLink, includeRegs, excludeRegs),
      500
    )
  ).current;

  // 手动触发预览（不防抖）
  const handlePreview = () => {
    preview(
      bangumiInfo.rssLink,
      bangumiInfo.includeRegs,
      bangumiInfo.excludeRegs
    );
  };

  // 监听匹配条件变化，自动触发预览
  useEffect(() => {
    if (open && bangumiInfo.rssLink) {
      debouncedPreview(
        bangumiInfo.rssLink,
        bangumiInfo.includeRegs,
        bangumiInfo.excludeRegs
      );
    }
  }, [
    bangumiInfo.includeRegs,
    bangumiInfo.excludeRegs,
    bangumiInfo.rssLink,
    open,
  ]);

  const validateField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "season":
        if (!value) error = "请填写季度";
        break;
      case "tmdbID":
        if (!value) error = "请填写TMDB ID";
        break;
      case "releaseGroup":
        if (!value.trim()) error = "请填写字幕组";
        break;
      case "episodeTotalNum":
        if (!value) error = "请填写总集数";
        break;
    }
    setFieldErrors((prev) => ({ ...prev, [field]: error }));
    return !error;
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setBangumiInfo({
        name: "",
        season: 0,
        year: "",
        tmdbID: 0,
        rssLink: "",
        releaseGroup: "",
        episodeTotalNum: 0,
        airWeekday: 0,
        includeRegs: [],
        excludeRegs: [],
        episodeOffset: 0,
        priority: 1,
        episodeLocation: "",
        backdropURL: "",
        posterURL: "",
      });
    }
    setFieldErrors({
      season: "",
      airWeekday: "",
      tmdbID: "",
      releaseGroup: "",
      episodeTotalNum: "",
    });
    onOpenChange(open);
  };

  const handleSubscribe = async () => {
    try {
      // 移除额外的字段，只保留 SubscribeRequest 需要的字段
      const { name, year, ...subscribeData } = bangumiInfo;
      await subscriptionAPI.subscribe(subscribeData);
      handleOpenChange(false);
      toast({
        title: "订阅成功",
        variant: "default",
      });
      if (onSubscribed) {
        onSubscribed();
      }
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "订阅失败",
        description: description,
        variant: "destructive",
      });
    }
  };

  const updateBangumiInfo = (field: keyof BangumiFormState, value: any) => {
    validateField(field, value);
    setBangumiInfo((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 处理TMDB ID变化（仅ID）
  const handleTMDBIDChange = (tmdbID: number) => {
    updateBangumiInfo("tmdbID", tmdbID);
  };

  // 处理TMDB元数据变化（完整信息）
  const handleTMDBMetaChange = (meta: Meta) => {
    setBangumiInfo((prev) => ({
      ...prev,
      tmdbID: meta.tmdbID,
      name: meta.chineseName,
      year: meta.year,
      season: meta.season,
      episodeTotalNum: meta.episodeTotalNum,
      airWeekday: meta.airWeekday || 0,
      posterURL: meta.posterURL,
      backdropURL: meta.backdropURL,
    }));

    // 验证相关字段
    validateField("tmdbID", meta.tmdbID);
    validateField("season", meta.season);
    validateField("episodeTotalNum", meta.episodeTotalNum);
  };

  return (
    <Dialog
      open={Boolean(open && parseRSSRsp.name)}
      onOpenChange={handleOpenChange}
    >
      <DialogContent className="max-w-lg w-[95dvw] max-h-[90dvh] overflow-y-auto rounded-2xl border-primary/20 bg-card/95 backdrop-blur-md sm:max-w-lg scrollbar-hide p-0 [&>button:last-child]:hidden">
        <div className="relative w-full overflow-hidden rounded-t-2xl">
          {/* 背景图 */}
          <div
            className="absolute inset-0 bg-cover bg-center"
            style={{
              backgroundImage: `url(${
                bangumiInfo.backdropURL || bangumiInfo.posterURL
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
                onClick={() => handleOpenChange(false)}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="flex gap-5 mt-4">
              {/* 海报 */}
              <div className="relative w-28 h-40 flex-shrink-0 rounded-xl bg-card/20 backdrop-blur-sm border-2 border-white/20 overflow-hidden shadow-2xl">
                {bangumiInfo.posterURL ? (
                  <img
                    src={bangumiInfo.posterURL}
                    className="w-full h-full object-cover"
                    alt={bangumiInfo.name}
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
                  {bangumiInfo.name}
                </h3>
                <div className="flex flex-wrap gap-2">
                  <span className="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold bg-white/20 text-white backdrop-blur-md border border-white/10">
                    {bangumiInfo.year || "未知年份"}
                  </span>
                  <span className="inline-flex items-center px-2.5 py-1 rounded-lg text-xs font-semibold bg-primary/40 text-white backdrop-blur-md border border-white/10">
                    第 {bangumiInfo.season} 季
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
                    value={bangumiInfo.name}
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
                    <Input
                      id="season"
                      type="number"
                      value={bangumiInfo.season}
                      onChange={(e) =>
                        updateBangumiInfo("season", parseInt(e.target.value))
                      }
                      className={cn(
                        "rounded-xl bg-muted/30 border-muted-foreground/10",
                        fieldErrors.season &&
                          "border-destructive focus-visible:ring-destructive"
                      )}
                    />
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
                      value={bangumiInfo.episodeTotalNum}
                      onChange={(e) =>
                        updateBangumiInfo(
                          "episodeTotalNum",
                          parseInt(e.target.value)
                        )
                      }
                      className={cn(
                        "rounded-xl bg-muted/30 border-muted-foreground/10",
                        fieldErrors.episodeTotalNum &&
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
                      value={bangumiInfo.airWeekday.toString()}
                      onValueChange={(value) =>
                        updateBangumiInfo("airWeekday", parseInt(value))
                      }
                    >
                      <SelectTrigger
                        id="airWeekDay"
                        className="rounded-xl bg-muted/30 border-muted-foreground/10"
                      >
                        <SelectValue placeholder="更新星期" />
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
                      value={bangumiInfo.releaseGroup}
                      onChange={(e) =>
                        updateBangumiInfo("releaseGroup", e.target.value)
                      }
                      placeholder="例如: VCB-Studio"
                      className={cn(
                        "rounded-xl bg-muted/30 border-muted-foreground/10",
                        fieldErrors.releaseGroup &&
                          "border-destructive focus-visible:ring-destructive"
                      )}
                    />
                  </div>
                </div>

                <Separator className="my-2 opacity-50" />

                <TMDBInput
                  type="tv"
                  value={bangumiInfo.tmdbID}
                  onTMDBIDChange={handleTMDBIDChange}
                  onMetaChange={handleTMDBMetaChange}
                  label="TMDB ID"
                  icon={<Search className="w-4 h-4" />}
                  className="mb-2"
                  error={fieldErrors.tmdbID}
                />
              </div>
            </TabsContent>

            <TabsContent value="filter" className="space-y-4 mt-0 outline-none">
              <MatchInput
                label="包含匹配"
                items={bangumiInfo.includeRegs}
                placeholder="添加包含匹配条件"
                onChange={(items) =>
                  setBangumiInfo((prev) => ({ ...prev, includeRegs: items }))
                }
              />
              <MatchInput
                label="排除匹配"
                items={bangumiInfo.excludeRegs}
                placeholder="添加排除匹配条件"
                onChange={(items) =>
                  setBangumiInfo((prev) => ({ ...prev, excludeRegs: items }))
                }
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
                      value={bangumiInfo.priority}
                      onChange={(e) =>
                        updateBangumiInfo("priority", parseInt(e.target.value))
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
                      value={bangumiInfo.episodeOffset}
                      onChange={(e) =>
                        updateBangumiInfo(
                          "episodeOffset",
                          parseInt(e.target.value)
                        )
                      }
                      className="rounded-xl bg-muted/30 border-muted-foreground/10"
                    />
                  </div>
                </div>

                <EpisodePositionInput
                  value={bangumiInfo.episodeLocation}
                  onChange={(value) =>
                    updateBangumiInfo("episodeLocation", value)
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
            onClick={() => handleOpenChange(false)}
            className="rounded-xl hover:bg-muted"
          >
            取消
          </Button>
          <Button
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
