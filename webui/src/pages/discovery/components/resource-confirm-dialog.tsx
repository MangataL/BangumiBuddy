import { Download } from "lucide-react";

import type { ResourceCandidate } from "@/api/discovery";
import type { DownloadType } from "@/api/magnet";
import { DownloadTypeSet } from "@/api/magnet";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface ResourceConfirmDialogProps {
  resource?: ResourceCandidate;
  submitting: boolean;
  downloadType: DownloadType;
  onDownloadTypeChange: (type: DownloadType) => void;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export function ResourceConfirmDialog(props: ResourceConfirmDialogProps) {
  return (
    <Dialog open={Boolean(props.resource)} onOpenChange={props.onOpenChange}>
      <DialogContent className="rounded-2xl motion-reduce:animate-none motion-reduce:transition-none">
        <DialogHeader>
          <DialogTitle>下载资源</DialogTitle>
          <DialogDescription>
            确认资源类型后创建下载任务。
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <p className="line-clamp-3 rounded-xl bg-muted/60 p-3 text-sm leading-6 text-muted-foreground">
            {props.resource?.title}
          </p>
          <div className="grid gap-2">
            <Label htmlFor="discovery-resource-type">资源类型</Label>
            <Select
              value={props.downloadType}
              onValueChange={(value) =>
                props.onDownloadTypeChange(value as DownloadType)
              }
              disabled={props.submitting}
            >
              <SelectTrigger
                id="discovery-resource-type"
                className="rounded-xl"
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={DownloadTypeSet.TV}>番剧</SelectItem>
                <SelectItem value={DownloadTypeSet.Movie}>剧场版</SelectItem>
              </SelectContent>
            </Select>
          </div>
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
            className="rounded-xl bg-gradient-to-r from-primary to-blue-500 text-white hover:opacity-90"
            disabled={props.submitting}
            onClick={props.onConfirm}
          >
            <Download className="h-4 w-4" />
            {props.submitting ? "创建中" : "开始下载"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
