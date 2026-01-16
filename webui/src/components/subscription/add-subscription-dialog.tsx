import { useState } from "react";
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
import { subscriptionAPI } from "@/api/subscription";
import { useToast } from "@/hooks/useToast";
import { ConfirmSubscriptionDialog } from "./confirm-subscription-dialog";
import { ParseRSSResponse } from "@/api/subscription";
import { extractErrorMessage } from "@/utils/error";
import { TMDBInput } from "@/components/tmdb/input";

export interface AddSubscriptionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubscribed?: () => void;
}

export function AddSubscriptionDialog({
  open,
  onOpenChange,
  onSubscribed,
}: AddSubscriptionDialogProps) {
  const { toast } = useToast();
  const [subscriptionUrl, setSubscriptionUrl] = useState("");
  const [tmdbID, setTmdbID] = useState<number>(0);
  const [parseDialogOpen, setParseDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [parseRSSRsp, setParseRSSRsp] = useState<ParseRSSResponse>({
    name: "",
    season: 1,
    year: "",
    tmdbID: 0,
    releaseGroup: "",
    episodeTotalNum: 0,
    airWeekday: 0,
    rssLink: "",
    posterURL: "",
    backdropURL: "",
  });

  const handleOpenChange = (open: boolean) => {
    setSubscriptionUrl("");
    setTmdbID(0);
    onOpenChange(open);
  };

  const handleParseSubscription = async () => {
    if (subscriptionUrl.trim() || tmdbID !== 0) {
      try {
        setLoading(true);
        const rssInfo = await subscriptionAPI.parseRSS({
          rssLink: subscriptionUrl.trim(),
          tmdbID: tmdbID || undefined,
        });
        setParseRSSRsp((prev) => ({
          ...prev,
          name: rssInfo.name,
          season: rssInfo.season,
          year: rssInfo.year,
          tmdbID: rssInfo.tmdbID,
          episodeTotalNum: rssInfo.episodeTotalNum,
          releaseGroup: rssInfo.releaseGroup,
          airWeekday: rssInfo.airWeekday,
          rssLink: subscriptionUrl.trim(),
          posterURL: rssInfo.posterURL,
          backdropURL: rssInfo.backdropURL,
        }));
        handleOpenChange(false);
        setParseDialogOpen(true);
      } catch (error) {
        const description = extractErrorMessage(
          error,
          "未知原因失败，请检查RSS链接或TMDB ID是否正确"
        );
        toast({
          title: "解析失败",
          description: description,
          variant: "destructive",
        });
      } finally {
        setLoading(false);
      }
    }
  };

  return (
    <>
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="rounded-xl border-primary/20 bg-card/95 backdrop-blur-md">
          <DialogHeader>
            <DialogTitle className="text-xl anime-gradient-text">
              添加订阅
            </DialogTitle>
            <DialogDescription>
              请输入RSS订阅链接，默认解析出错时可以手动输入TMDB ID进行解析
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="subscription-url">订阅链接</Label>
              <Input
                id="subscription-url"
                placeholder="https://mikanani.me/RSS/Bangumi?bangumiId=xxx&subgroupid=xxx"
                value={subscriptionUrl}
                onChange={(e) => setSubscriptionUrl(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && subscriptionUrl.trim()) {
                    e.preventDefault();
                    handleParseSubscription();
                  }
                }}
                className="rounded-xl border-primary/20 focus:border-primary focus:ring-primary"
              />
            </div>
            <div className="grid gap-2">
              <TMDBInput
                value={tmdbID}
                onTMDBIDChange={(id) => setTmdbID(id)}
                onMetaChange={() => {}}
                type="tv"
              />
            </div>
          </div>
          <DialogFooter className="gap-2 sm:gap-0">
            <Button
              variant="outline"
              onClick={() => handleOpenChange(false)}
              className="rounded-xl"
              disabled={loading}
            >
              取消
            </Button>
            <Button
              onClick={handleParseSubscription}
              className="rounded-xl bg-gradient-to-r from-primary to-blue-500"
              disabled={loading || (!subscriptionUrl.trim() && tmdbID === 0)}
            >
              {loading ? "解析中..." : "解析"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmSubscriptionDialog
        open={parseDialogOpen}
        parseRSSRsp={parseRSSRsp}
        onOpenChange={setParseDialogOpen}
        onSubscribed={onSubscribed}
      />
    </>
  );
}
