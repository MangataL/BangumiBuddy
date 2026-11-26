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
import { ConfirmSubscriptionDialog } from "./confirm-subscription-dialog";
import type { ParseRSSResponse } from "@/api/subscription";
import { useParseSubscription } from "./use-parse-subscription";
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
  const [parsed, setParsed] = useState<ParseRSSResponse>();

  return (
    <>
      {open && (
        <ParseSubscriptionDialog
          onClose={() => onOpenChange(false)}
          onParsed={(result) => {
            setParsed(result);
            onOpenChange(false);
          }}
        />
      )}
      {parsed && (
        <ConfirmSubscriptionDialog
          parseRSSRsp={parsed}
          onClose={() => setParsed(undefined)}
          onSubscribed={onSubscribed}
        />
      )}
    </>
  );
}

function ParseSubscriptionDialog({
  onClose,
  onParsed,
}: {
  onClose: () => void;
  onParsed: (result: ParseRSSResponse) => void;
}) {
  const [subscriptionUrl, setSubscriptionUrl] = useState("");
  const [tmdbID, setTmdbID] = useState(0);
  const { parse, cancel, loading } = useParseSubscription(onParsed);
  const canParse = (Boolean(subscriptionUrl.trim()) || tmdbID > 0) && !loading;

  const handleParseSubscription = () => {
    if (canParse) void parse(subscriptionUrl.trim(), tmdbID || undefined);
  };

  return (
    <Dialog open onOpenChange={onClose}>
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
              onChange={(e) => {
                cancel();
                setSubscriptionUrl(e.target.value);
              }}
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
              onTMDBIDChange={(id) => {
                if (id !== tmdbID) cancel();
                setTmdbID(id);
              }}
              type="tv"
            />
          </div>
        </div>
        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={onClose} className="rounded-xl">
            取消
          </Button>
          <Button
            onClick={handleParseSubscription}
            className="rounded-xl bg-gradient-to-r from-primary to-blue-500"
            disabled={!canParse}
          >
            {loading ? "解析中..." : "解析"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
