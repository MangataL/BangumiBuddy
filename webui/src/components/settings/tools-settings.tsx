import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Loader2, CheckCircle } from "lucide-react";
import { useToast } from "@/hooks/useToast";
import { configAPI, type SubtitleOperatorConfig } from "@/api/config";
import { extractErrorMessage } from "@/utils/error";
import { SubtitleConfig } from "./subtitle-config";

interface ToolsSettingsProps {
  onFontStatsUpdate?: () => void;
}

export function ToolsSettings({ onFontStatsUpdate }: ToolsSettingsProps = {}) {
  const { toast } = useToast();
  const [subtitleConfig, setSubtitleConfig] = useState<SubtitleOperatorConfig>({
    useOTF: false,
    useSimilarFont: false,
    useSystemFontsDir: false,
    coverExistSubFont: false,
    generateNewFile: false,
    checkGlyphs: false,
  });
  const [loading, setLoading] = useState(false);

  // 加载字幕操作器配置
  const loadSubtitleConfig = async () => {
    try {
      setLoading(true);
      const config = await configAPI.getSubtitleOperatorConfig();
      setSubtitleConfig(config);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "加载字幕配置失败",
        description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  // 保存字幕操作器配置
  const saveSubtitleConfig = async () => {
    try {
      setLoading(true);
      await configAPI.setSubtitleOperatorConfig(subtitleConfig);
      toast({
        title: "保存成功",
        description: "字幕操作器配置已保存",
        variant: "default",
      });
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "保存失败",
        description,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSubtitleConfig();
  }, []);

  return (
    <Card className="border-primary/10 rounded-xl overflow-hidden">
      <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
        <CardTitle className="text-xl anime-gradient-text">工具设置</CardTitle>
        <CardDescription>配置字幕子集化和字体库管理</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4 p-6">
        {/* 字幕子集化配置 */}
        <SubtitleConfig
          subtitleConfig={subtitleConfig}
          setSubtitleConfig={setSubtitleConfig}
          loading={loading}
          onFontStatsUpdate={onFontStatsUpdate}
        />

        {/* 保存按钮 */}
        <div className="flex gap-3 pt-4">
          <Button
            onClick={saveSubtitleConfig}
            disabled={loading}
            className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                保存中
              </>
            ) : (
              <>
                <CheckCircle className="h-4 w-4 mr-2" />
                保存工具设置
              </>
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
