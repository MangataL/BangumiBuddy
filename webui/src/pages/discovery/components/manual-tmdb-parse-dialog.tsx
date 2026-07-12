import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { TMDBInput } from "@/components/tmdb";

interface ManualTMDBParseDialogProps {
  open: boolean;
  errorMessage: string;
  suggestedName: string;
  releaseGroupName: string;
  submitting: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (tmdbID: number) => void;
}

export function ManualTMDBParseDialog(props: ManualTMDBParseDialogProps) {
  const [tmdbID, setTmdbID] = useState(0);

  useEffect(() => {
    if (props.open) {
      setTmdbID(0);
    }
  }, [props.open]);

  return (
    <Dialog
      open={props.open}
      onOpenChange={(open) => {
        if (!open && props.submitting) return;
        props.onOpenChange(open);
      }}
    >
      <DialogContent className="rounded-2xl motion-reduce:animate-none motion-reduce:transition-none">
        <DialogHeader>
          <DialogTitle>字幕组解析失败</DialogTitle>
          <DialogDescription>
            自动匹配 TMDB 失败时，可手动搜索并指定番剧信息后继续添加订阅。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2 rounded-xl bg-muted/60 p-3 text-sm leading-6 text-muted-foreground">
            {props.releaseGroupName ? (
              <p>
                字幕组：
                <span className="text-foreground">{props.releaseGroupName}</span>
              </p>
            ) : null}
            {props.suggestedName ? (
              <p>
                识别名称：
                <span className="text-foreground">{props.suggestedName}</span>
              </p>
            ) : null}
            <p className="text-destructive">{props.errorMessage}</p>
          </div>

          <TMDBInput
            type="tv"
            value={tmdbID}
            onTMDBIDChange={setTmdbID}
            initialSearchName={props.suggestedName}
            label="TMDB ID"
            placeholder="输入 TMDB ID 或点击搜索"
          />
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            className="rounded-xl"
            disabled={props.submitting}
            onClick={() => props.onOpenChange(false)}
          >
            取消
          </Button>
          <Button
            type="button"
            className="rounded-xl"
            disabled={props.submitting || tmdbID <= 0}
            onClick={() => props.onConfirm(tmdbID)}
          >
            {props.submitting ? "解析中..." : "继续解析"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
