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
import { AxiosError } from "axios";
import { ConfirmSubscriptionDialog } from "./confirm-subscription-dialog";
import { ParseRSSResponse } from "@/api/subscription";

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
  const [parseDialogOpen, setParseDialogOpen] = useState(false);
  const [parseRSSRsp, setParseRSSRsp] = useState<ParseRSSResponse>({
    name: "",
    season: 1,
    year: "",
    tmdbID: 0,
    releaseGroup: "",
    episodeTotalNum: 0,
    airWeekday: 0,
    rssLink: "",
  });

  const handleOpenChange = (open: boolean) => {
    setSubscriptionUrl("");
    onOpenChange(open);
  };

  const handleParseSubscription = async () => {
    if (subscriptionUrl.trim()) {
      try {
        const rssInfo = await subscriptionAPI.parseRSS(subscriptionUrl.trim());
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
        }));
        handleOpenChange(false);
        setParseDialogOpen(true);
      } catch (error) {
        const description =
          (error as AxiosError<{ error: string }>)?.response?.data?.error ||
          "请检查RSS链接是否正确";
        toast({
          title: "解析RSS链接失败",
          description: description,
          variant: "destructive",
        });
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
              请输入RSS订阅链接，我们将解析订阅内容
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
              onClick={handleParseSubscription}
              className="rounded-xl bg-gradient-to-r from-primary to-blue-500"
            >
              解析
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
