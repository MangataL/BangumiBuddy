import { useState } from "react";
import { Tv, Film } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useToast } from "@/hooks/useToast";
import magnetAPI, {
  type DownloadType,
  DownloadTypeSet,
  DownloadTypeLabels,
} from "@/api/magnet";
import { extractErrorMessage } from "@/utils/error";

export interface AddMagnetDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (taskID: string) => void;
}

export function AddMagnetDialog({
  open,
  onOpenChange,
  onSuccess,
}: AddMagnetDialogProps) {
  const { toast } = useToast();
  const [magnetLink, setMagnetLink] = useState("");
  const [contentType, setContentType] = useState<DownloadType>(
    DownloadTypeSet.TV
  );
  const [loading, setLoading] = useState(false);

  const handleOpenChange = (open: boolean) => {
    setMagnetLink("");
    setContentType(DownloadTypeSet.TV);
    onOpenChange(open);
  };

  const handleAddMagnet = async () => {
    if (magnetLink.trim()) {
      try {
        setLoading(true);
        const response = await magnetAPI.addTask(
          magnetLink.trim(),
          contentType
        );
        toast({
          title: "添加成功",
          description: "磁力任务已添加",
        });
        handleOpenChange(false);
        // 调用成功回调，传递任务ID
        onSuccess?.(response.taskID);
      } catch (error) {
        const description = extractErrorMessage(
          error,
          "请检查磁力链接是否正确"
        );
        toast({
          title: "添加磁力任务失败",
          description: description,
          variant: "destructive",
        });
      } finally {
        setLoading(false);
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="rounded-xl border-primary/20 bg-card/95 backdrop-blur-md">
        <DialogHeader>
          <DialogTitle className="text-xl anime-gradient-text">
            添加磁力任务
          </DialogTitle>
          <DialogDescription>请输入磁力链接并选择下载类型</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="magnet-link">磁力链接</Label>
            <Input
              id="magnet-link"
              placeholder="magnet:?xt=urn:btih:..."
              value={magnetLink}
              onChange={(e) => setMagnetLink(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && magnetLink.trim()) {
                  e.preventDefault();
                  handleAddMagnet();
                }
              }}
              className="rounded-xl border-primary/20 focus:border-primary focus:ring-primary"
            />
          </div>
          <div className="grid gap-2">
            <Label>下载类型</Label>
            <RadioGroup
              value={contentType}
              onValueChange={(value) => setContentType(value as DownloadType)}
              className="flex flex-col gap-3"
            >
              <div className="flex items-center space-x-2">
                <RadioGroupItem value={DownloadTypeSet.TV} id="tv" />
                <Label
                  htmlFor={DownloadTypeSet.TV}
                  className="cursor-pointer font-normal flex items-center gap-2"
                >
                  <Tv className="h-4 w-4" />
                  {DownloadTypeLabels[DownloadTypeSet.TV]}
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value={DownloadTypeSet.Movie} id="movie" />
                <Label
                  htmlFor={DownloadTypeSet.Movie}
                  className="cursor-pointer font-normal flex items-center gap-2"
                >
                  <Film className="h-4 w-4" />
                  {DownloadTypeLabels[DownloadTypeSet.Movie]}
                </Label>
              </div>
            </RadioGroup>
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
            onClick={handleAddMagnet}
            className="rounded-xl bg-gradient-to-r from-primary to-blue-500"
            disabled={loading || !magnetLink.trim()}
          >
            {loading ? "添加中..." : "添加"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
