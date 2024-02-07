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
import { subscriptionAPI } from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { AxiosError } from "axios";
import { MatchInput } from "../common/match-input";
import { EpisodePositionInput } from "./episode-position-input";
import { getSortedWeekDays } from "@/utils/weekday";
import { ParseRSSResponse, SubscribeRequest } from "@/api/subscription";

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
  });

  // 监听 parseRSSRsp 变化
  useEffect(() => {
    if (parseRSSRsp) {
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
      }));
    }
  }, [open]);

  const [fieldErrors, setFieldErrors] = useState({
    season: "",
    airWeekday: "",
    tmdbID: "",
    releaseGroup: "",
    episodeTotalNum: "",
  });

  const validateField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "season":
        if (!value) error = "请填写季度";
        break;
      case "airWeekday":
        if (value === 0) error = "请选择更新星期";
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
      const description =
        (error as AxiosError<{ error: string }>).response?.data?.error ||
        "订阅失败，请重试";
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

  return (
    <Dialog
      open={Boolean(open && parseRSSRsp.name)}
      onOpenChange={handleOpenChange}
    >
      <DialogContent className="max-w-md rounded-xl border-primary/20 bg-card/95 backdrop-blur-md">
        <DialogHeader>
          <DialogTitle className="text-xl anime-gradient-text">
            订阅信息
          </DialogTitle>
          <DialogDescription>请确认或修改以下订阅信息</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label>番剧名称</Label>
            <Input value={bangumiInfo.name} className="rounded-xl" readOnly />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid content-start gap-2">
              <Label htmlFor="season">季度</Label>
              <Input
                id="season"
                type="number"
                value={bangumiInfo.season}
                onChange={(e) =>
                  updateBangumiInfo("season", parseInt(e.target.value))
                }
                className={`rounded-xl ${
                  fieldErrors.season ? "border-destructive" : ""
                }`}
              />
              {fieldErrors.season && (
                <span className="text-sm text-destructive">
                  {fieldErrors.season}
                </span>
              )}
            </div>
            <div className="grid content-start gap-2">
              <Label htmlFor="year">年份</Label>
              <Input
                id="year"
                type="number"
                value={bangumiInfo.year}
                className="rounded-xl"
                readOnly
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid content-start gap-2">
              <Label htmlFor="episodeTotalNum">总集数</Label>
              <Input
                id="episodeTotalNum"
                type="number"
                min="1"
                value={bangumiInfo.episodeTotalNum}
                onChange={(e) =>
                  updateBangumiInfo("episodeTotalNum", parseInt(e.target.value))
                }
                className={`rounded-xl ${
                  fieldErrors.episodeTotalNum ? "border-destructive" : ""
                }`}
              />
              {fieldErrors.episodeTotalNum && (
                <span className="text-sm text-destructive">
                  {fieldErrors.episodeTotalNum}
                </span>
              )}
            </div>
            <div className="grid content-start gap-2">
              <Label htmlFor="airWeekDay">更新星期</Label>
              <Select
                value={bangumiInfo.airWeekday.toString()}
                onValueChange={(value) =>
                  updateBangumiInfo("airWeekday", parseInt(value))
                }
              >
                <SelectTrigger
                  id="airWeekDay"
                  className={`rounded-xl ${
                    fieldErrors.airWeekday ? "border-destructive" : ""
                  }`}
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
              {fieldErrors.airWeekday && (
                <span className="text-sm text-destructive">
                  {fieldErrors.airWeekday}
                </span>
              )}
            </div>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="tmdbid">TMDBID</Label>
            <Input
              id="tmdbid"
              value={bangumiInfo.tmdbID}
              onChange={(e) =>
                updateBangumiInfo("tmdbID", parseInt(e.target.value))
              }
              className={`rounded-xl ${
                fieldErrors.tmdbID ? "border-destructive" : ""
              }`}
            />
            {fieldErrors.tmdbID && (
              <span className="text-sm text-destructive">
                {fieldErrors.tmdbID}
              </span>
            )}
          </div>
          <div className="grid gap-2">
            <Label htmlFor="subgroup">字幕组</Label>
            <Input
              id="subgroup"
              value={bangumiInfo.releaseGroup}
              onChange={(e) =>
                updateBangumiInfo("releaseGroup", e.target.value)
              }
              className={`rounded-xl ${
                fieldErrors.releaseGroup ? "border-destructive" : ""
              }`}
            />
            {fieldErrors.releaseGroup && (
              <span className="text-sm text-destructive">
                {fieldErrors.releaseGroup}
              </span>
            )}
          </div>

          <MatchInput
            label="包含匹配"
            items={bangumiInfo.includeRegs}
            placeholder="添加包含匹配条件"
            onChange={(items) =>
              setBangumiInfo((prev) => ({
                ...prev,
                includeRegs: items,
              }))
            }
          />

          <MatchInput
            label="排除匹配"
            items={bangumiInfo.excludeRegs}
            placeholder="添加排除匹配条件"
            onChange={(items) =>
              setBangumiInfo((prev) => ({
                ...prev,
                excludeRegs: items,
              }))
            }
          />

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="priority">优先级</Label>
              <Input
                id="priority"
                type="number"
                value={bangumiInfo.priority}
                onChange={(e) =>
                  updateBangumiInfo("priority", parseInt(e.target.value))
                }
                className="rounded-xl"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="episodeOffset">集数偏移</Label>
              <Input
                id="episodeOffset"
                type="number"
                className="rounded-xl"
                value={bangumiInfo.episodeOffset}
                onChange={(e) =>
                  updateBangumiInfo("episodeOffset", parseInt(e.target.value))
                }
              />
            </div>
          </div>

          <EpisodePositionInput
            value={bangumiInfo.episodeLocation}
            onChange={(value) => updateBangumiInfo("episodeLocation", value)}
          />
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            className="rounded-xl"
          >
            取消
          </Button>
          <Button
            onClick={handleSubscribe}
            className="rounded-xl bg-gradient-to-r from-primary to-blue-500"
          >
            订阅
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
