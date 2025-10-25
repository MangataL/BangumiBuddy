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
import {
  configAPI,
  type SubtitleOperatorConfig,
  type ScraperConfig,
} from "@/api/config";
import { extractErrorMessage } from "@/utils/error";
import { SubtitleConfig } from "./subtitle-config";
import { ScraperConfigComponent } from "./scraper-config";

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
  const [scraperConfig, setScraperConfig] = useState<ScraperConfig>({
    enable: false,
    checkInterval: 24,
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

  // 加载刮削器配置
  const loadScraperConfig = async () => {
    try {
      const config = await configAPI.getScraperConfig();
      setScraperConfig(config);
    } catch (error) {
      const description = extractErrorMessage(error);
      toast({
        title: "加载刮削配置失败",
        description,
        variant: "destructive",
      });
    }
  };

  // 保存所有配置
  const saveAllConfigs = async () => {
    try {
      setLoading(true);
      await Promise.all([
        configAPI.setSubtitleOperatorConfig(subtitleConfig),
        configAPI.setScraperConfig(scraperConfig),
      ]);
      toast({
        title: "保存成功",
        description: "工具配置已保存",
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
    loadScraperConfig();
  }, []);

  return (
    <Card className="border-primary/10 rounded-xl overflow-hidden">
      <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
        <CardTitle className="text-xl anime-gradient-text">工具设置</CardTitle>
        <CardDescription>
          配置字幕子集化、字体库管理和媒体库刮削
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6 p-6">
        {/* 字幕子集化配置 */}
        <SubtitleConfig
          subtitleConfig={subtitleConfig}
          setSubtitleConfig={setSubtitleConfig}
          loading={loading}
          onFontStatsUpdate={onFontStatsUpdate}
        />

        {/* 媒体库刮削配置 */}
        <ScraperConfigComponent
          scraperConfig={scraperConfig}
          setScraperConfig={setScraperConfig}
          loading={loading}
        />

        {/* 保存按钮 */}
        <div className="flex gap-3 pt-4">
          <Button
            onClick={saveAllConfigs}
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
