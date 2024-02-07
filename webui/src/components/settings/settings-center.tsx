import { useEffect, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Download,
  Tv,
  FolderSync,
  Sparkles,
  Info,
  Search,
  CheckCircle,
  Bell,
} from "lucide-react";
import { useToast } from "@/hooks/useToast";
import {
  configAPI,
  type DownloaderConfig,
  type DownloadManagerConfig,
  type TMDBConfig,
  type SubscriptionConfig,
  type TransferConfig,
  type NoticeConfig,
} from "@/api/config";
import { AxiosError } from "axios";
import { MatchInput } from "@/components/common/match-input";
import { Switch } from "@/components/ui/switch";
import { PasswordInput } from "../common/password-input";

export default function SettingsCenter() {
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState("download");

  // 下载配置状态
  const [downloaderConfig, setDownloaderConfig] = useState<DownloaderConfig>({
    downloadType: "qbittorrent",
    qbittorrent: {
      host: "",
      username: "",
      password: "",
    },
  });
  const [downloadManagerConfig, setDownloadManagerConfig] =
    useState<DownloadManagerConfig>({
      tvSavePath: "",
      movieSavePath: "",
    });
  const [downloaderConfigChanged, setDownloaderConfigChanged] = useState(false);
  const [downloadManagerConfigChanged, setDownloadManagerConfigChanged] =
    useState(false);
  const [connectionStatus, setConnectionStatus] = useState<
    "none" | "success" | "error"
  >("none");
  const [isChecking, setIsChecking] = useState(false);

  // TMDB配置状态
  const [tmdbConfig, setTmdbConfig] = useState<TMDBConfig>({
    tmdbToken: "",
    alternateURL: false,
  });

  // 订阅配置状态
  const [subscriptionConfig, setSubscriptionConfig] =
    useState<SubscriptionConfig>({
      rssCheckInterval: 30,
      includeRegs: [],
      excludeRegs: [],
      autoStop: false,
    });

  // 文件转移配置状态
  const [transferConfig, setTransferConfig] = useState<TransferConfig>({
    interval: 15,
    tvPath: "",
    tvFormat: "",
    transferType: "hardlink",
  });

  // 通知配置状态
  const [noticeConfig, setNoticeConfig] = useState<NoticeConfig>({
    enabled: false,
    type: "email",
    telegram: {
      token: "",
      chatID: 0,
    },
    email: {
      host: "",
      username: "",
      password: "",
      from: "",
      to: [],
      ssl: true,
    },
    bark: {
      serverPath: "",
      sound: "",
      interruption: "active",
      autoSave: true,
    },
    noticePoints: {
      subscriptionUpdated: true,
      downloaded: true,
      transferred: true,
      error: true,
    },
  });

  // 加载配置函数
  const loadConfigs = async () => {
    try {
      const [downloader, manager, tmdb, subscription, transfer, notice] =
        await Promise.all([
          configAPI.getDownloaderConfig(),
          configAPI.getDownloadManagerConfig(),
          configAPI.getTMDBConfig(),
          configAPI.getSubscriptionConfig(),
          configAPI.getTransferConfig(),
          configAPI.getNoticeConfig(),
        ]);

      setDownloaderConfig(downloader);
      setDownloadManagerConfig(manager);
      setTmdbConfig(tmdb);
      setSubscriptionConfig(subscription);
      setTransferConfig(transfer);
      setNoticeConfig(notice);
    } catch (error) {
      const desc =
        (error as AxiosError<{ error: string }>)?.response?.data?.error ||
        "请检查网络连接后重试";
      toast({
        title: "加载配置失败",
        description: desc,
        variant: "destructive",
      });
    }
  };

  // 加载配置
  useEffect(() => {
    loadConfigs();
  }, []);

  // 检查qBittorrent连通性
  const checkQBittorrentConnection = async () => {
    if (isChecking) return;
    setIsChecking(true);
    try {
      await configAPI.checkQBittorrentConnection({
        host: downloaderConfig.qbittorrent.host,
        username: downloaderConfig.qbittorrent.username,
        password: downloaderConfig.qbittorrent.password,
      });
      setConnectionStatus("success");
      toast({
        title: "连接成功",
        description: "成功连接到qBittorrent",
      });
    } catch (error) {
      setConnectionStatus("error");
      const desc =
        (error as AxiosError<{ error: string }>)?.response?.data?.error ||
        "请检查网络连接后重试";
      toast({
        title: "连接失败",
        description: desc,
        variant: "destructive",
      });
    } finally {
      setIsChecking(false);
    }
  };

  // 更新收件人列表
  const handleEmailRecipientsChange = (value: string) => {
    // 将逗号分隔的邮件地址转换为数组
    const recipients = value
      .split(",")
      .map((email) => email.trim())
      .filter(Boolean);
    setNoticeConfig((prev) => ({
      ...prev,
      email: {
        ...prev.email,
        to: recipients,
      },
    }));
  };

  // 保存配置
  const handleSaveSettings = async (tab: string) => {
    try {
      switch (tab) {
        case "download":
          if (downloaderConfigChanged) {
            await configAPI.setDownloaderConfig(downloaderConfig);
            const newDownloaderConfig = await configAPI.getDownloaderConfig();
            setDownloaderConfig(newDownloaderConfig);
            setDownloaderConfigChanged(false);
          }
          if (downloadManagerConfigChanged) {
            await configAPI.setDownloadManagerConfig(downloadManagerConfig);
            const newManagerConfig = await configAPI.getDownloadManagerConfig();
            setDownloadManagerConfig(newManagerConfig);
            setDownloadManagerConfigChanged(false);
          }
          break;
        case "metadata":
          await configAPI.setTMDBConfig(tmdbConfig);
          const newTmdbConfig = await configAPI.getTMDBConfig();
          setTmdbConfig(newTmdbConfig);
          break;
        case "subscription":
          await configAPI.setSubscriptionConfig(subscriptionConfig);
          const newSubscriptionConfig = await configAPI.getSubscriptionConfig();
          setSubscriptionConfig(newSubscriptionConfig);
          break;
        case "transfer":
          await configAPI.setTransferConfig(transferConfig);
          const newTransferConfig = await configAPI.getTransferConfig();
          setTransferConfig(newTransferConfig);
          break;
        case "notice":
          await configAPI.setNoticeConfig(noticeConfig);
          const newNoticeConfig = await configAPI.getNoticeConfig();
          setNoticeConfig(newNoticeConfig);
          break;
      }

      toast({
        title: "设置已保存",
        description: `已成功保存${
          tab === "download"
            ? "下载"
            : tab === "metadata"
            ? "元数据"
            : tab === "subscription"
            ? "订阅"
            : tab === "transfer"
            ? "文件转移"
            : "通知"
        }设置`,
      });
    } catch (error) {
      const desc =
        (error as AxiosError<{ error: string }>)?.response?.data?.error ||
        "请检查网络连接后重试";
      console.error(desc);
      toast({
        title: "保存设置失败",
        description: desc,
        variant: "destructive",
      });
    }
  };

  return (
    <div className="space-y-6 h-[calc(100vh-6rem)] overflow-y-auto pb-10">
      <div>
        <h1 className="text-3xl font-bold anime-gradient-text flex items-center gap-2">
          <Sparkles className="h-6 w-6 text-primary animate-pulse" />
          设置中心
        </h1>
        <p className="text-muted-foreground">
          管理您的下载器、订阅和文件转移设置
        </p>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-4"
      >
        <TabsList className="grid w-full grid-cols-5 rounded-xl p-1">
          <TabsTrigger
            value="download"
            className="flex items-center gap-2 rounded-xl"
          >
            <Download className="h-4 w-4" />
            <span className="hidden sm:inline">下载设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="metadata"
            className="flex items-center gap-2 rounded-xl"
          >
            <Search className="h-4 w-4" />
            <span className="hidden sm:inline">元数据设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="subscription"
            className="flex items-center gap-2 rounded-xl"
          >
            <Tv className="h-4 w-4" />
            <span className="hidden sm:inline">订阅设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="transfer"
            className="flex items-center gap-2 rounded-xl"
          >
            <FolderSync className="h-4 w-4" />
            <span className="hidden sm:inline">文件转移设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="notice"
            className="flex items-center gap-2 rounded-xl"
          >
            <Bell className="h-4 w-4" />
            <span className="hidden sm:inline">通知设置</span>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="download">
          <Card className="border-primary/10 rounded-xl overflow-hidden">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
              <CardTitle className="text-xl anime-gradient-text">
                下载设置
              </CardTitle>
              <CardDescription>配置文件保存位置与下载器</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6 p-6">
              {/* 文件保存配置 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">文件保存配置</h3>
                <div className="space-y-2">
                  <Label htmlFor="anime-path">番剧保存位置</Label>
                  <Input
                    id="anime-path"
                    value={downloadManagerConfig.tvSavePath}
                    onChange={(e) => {
                      setDownloadManagerConfig((prev) => ({
                        ...prev,
                        tvSavePath: e.target.value,
                      }));
                      setDownloadManagerConfigChanged(true);
                    }}
                    placeholder="/path/to/anime"
                    className="rounded-xl"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="movie-path">剧场版保存位置</Label>
                  <Input
                    id="movie-path"
                    value={downloadManagerConfig.movieSavePath}
                    onChange={(e) => {
                      setDownloadManagerConfig((prev) => ({
                        ...prev,
                        movieSavePath: e.target.value,
                      }));
                      setDownloadManagerConfigChanged(true);
                    }}
                    placeholder="/path/to/movies"
                    className="rounded-xl"
                  />
                </div>
              </div>

              {/* 下载器配置 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">下载器配置</h3>
                <div className="space-y-2">
                  <Label htmlFor="downloader-type">下载器类型</Label>
                  <div className="flex items-center gap-2">
                    <Select
                      value={downloaderConfig.downloadType}
                      onValueChange={(value: "qbittorrent") => {
                        setDownloaderConfig((prev) => ({
                          ...prev,
                          downloadType: value,
                        }));
                        setDownloaderConfigChanged(true);
                        setConnectionStatus("none");
                      }}
                    >
                      <SelectTrigger
                        id="downloader-type"
                        className="rounded-xl w-[180px]"
                      >
                        <SelectValue placeholder="选择下载器类型" />
                      </SelectTrigger>
                      <SelectContent className="rounded-xl">
                        <SelectItem value="qbittorrent">qBittorrent</SelectItem>
                      </SelectContent>
                    </Select>
                    {downloaderConfig.downloadType === "qbittorrent" && (
                      <Button
                        variant="outline"
                        className={`rounded-xl flex items-center gap-2 min-w-[120px] ${
                          connectionStatus === "success"
                            ? "border-green-500 text-green-500"
                            : connectionStatus === "error"
                            ? "border-destructive text-destructive"
                            : ""
                        }`}
                        onClick={checkQBittorrentConnection}
                        disabled={isChecking}
                      >
                        {connectionStatus === "success" ? (
                          <CheckCircle className="h-4 w-4" />
                        ) : connectionStatus === "error" ? (
                          <CheckCircle className="h-4 w-4" />
                        ) : (
                          <CheckCircle
                            className={`h-4 w-4 ${
                              isChecking ? "animate-spin" : ""
                            }`}
                          />
                        )}
                        <span>测试连接</span>
                      </Button>
                    )}
                  </div>
                </div>

                <div className="space-y-4 pl-4 border-l-2 border-primary/10">
                  <div className="space-y-2">
                    <Label htmlFor="downloader-url">下载器地址</Label>
                    <Input
                      id="downloader-url"
                      value={downloaderConfig.qbittorrent.host}
                      onChange={(e) => {
                        setDownloaderConfig((prev) => ({
                          ...prev,
                          qbittorrent: {
                            ...prev.qbittorrent,
                            host: e.target.value,
                          },
                        }));
                        setDownloaderConfigChanged(true);
                      }}
                      placeholder="http://localhost:8080"
                      className="rounded-xl placeholder-gray-400"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="downloader-username">用户名</Label>
                      <Input
                        id="downloader-username"
                        value={downloaderConfig.qbittorrent.username}
                        onChange={(e) => {
                          setDownloaderConfig((prev) => ({
                            ...prev,
                            qbittorrent: {
                              ...prev.qbittorrent,
                              username: e.target.value,
                            },
                          }));
                          setDownloaderConfigChanged(true);
                        }}
                        placeholder="admin"
                        className="rounded-xl placeholder-gray-400"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="downloader-password">密码</Label>
                      <PasswordInput
                        id="downloader-password"
                        value={downloaderConfig.qbittorrent.password}
                        onChange={(e) => {
                          setDownloaderConfig((prev) => ({
                            ...prev,
                            qbittorrent: {
                              ...prev.qbittorrent,
                              password: e.target.value,
                            },
                          }));
                          setDownloaderConfigChanged(true);
                        }}
                        className="rounded-xl placeholder-gray-400"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("download")}
              >
                保存下载设置
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="metadata">
          <Card className="border-primary/10 rounded-xl overflow-hidden">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
              <CardTitle className="text-xl anime-gradient-text">
                元数据设置
              </CardTitle>
              <CardDescription>配置元数据获取方式</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 p-6">
              <div className="space-y-2">
                <Label htmlFor="tmdb-token">TMDB Token</Label>
                <Input
                  id="tmdb-token"
                  value={tmdbConfig.tmdbToken}
                  onChange={(e) =>
                    setTmdbConfig((prev) => ({
                      ...prev,
                      tmdbToken: e.target.value,
                    }))
                  }
                  placeholder="输入您的TMDB Token"
                  className="rounded-xl placeholder-gray-400"
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="use-alternative-url">使用备用地址</Label>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-5 w-5 rounded-full"
                        >
                          <Info className="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>备用地址使用api.tmdb.org/3，无代理的情况下可选</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
                <Select
                  value={tmdbConfig.alternateURL ? "true" : "false"}
                  onValueChange={(value) =>
                    setTmdbConfig((prev) => ({
                      ...prev,
                      alternateURL: value === "true",
                    }))
                  }
                >
                  <SelectTrigger
                    id="use-alternative-url"
                    className="rounded-xl"
                  >
                    <SelectValue placeholder="选择是否使用备用地址" />
                  </SelectTrigger>
                  <SelectContent className="rounded-xl">
                    <SelectItem value="true">是</SelectItem>
                    <SelectItem value="false">否</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("metadata")}
              >
                保存元数据设置
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="subscription">
          <Card className="border-primary/10 rounded-xl overflow-hidden">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
              <CardTitle className="text-xl anime-gradient-text">
                订阅设置
              </CardTitle>
              <CardDescription>配置订阅检查间隔和匹配规则</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4 p-6">
              <div className="space-y-2">
                <Label htmlFor="check-interval">检查间隔（分钟）</Label>
                <Input
                  id="check-interval"
                  type="number"
                  value={subscriptionConfig.rssCheckInterval}
                  onChange={(e) =>
                    setSubscriptionConfig((prev) => ({
                      ...prev,
                      rssCheckInterval: parseInt(e.target.value) || 30,
                    }))
                  }
                  className="rounded-xl placeholder-gray-400"
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="auto-stop">自动停止订阅</Label>
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-5 w-5 rounded-full"
                        >
                          <Info className="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>
                          当下载到最后一集时，自动停止订阅（可以避免下载字幕组后续更新的合集），按需启动
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    id="auto-stop"
                    checked={subscriptionConfig.autoStop}
                    onCheckedChange={(checked) =>
                      setSubscriptionConfig((prev) => ({
                        ...prev,
                        autoStop: checked,
                      }))
                    }
                  />
                </div>
              </div>

              {/* 全局包含匹配 */}
              <MatchInput
                label="全局包含匹配"
                items={subscriptionConfig.includeRegs}
                placeholder="添加全局包含匹配条件"
                onChange={(items) =>
                  setSubscriptionConfig((prev) => ({
                    ...prev,
                    includeRegs: items,
                  }))
                }
              />

              {/* 全局排除匹配 */}
              <MatchInput
                label="全局排除匹配"
                items={subscriptionConfig.excludeRegs}
                placeholder="添加全局排除匹配条件"
                onChange={(items) =>
                  setSubscriptionConfig((prev) => ({
                    ...prev,
                    excludeRegs: items,
                  }))
                }
              />

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("subscription")}
              >
                保存订阅设置
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="transfer">
          <Card className="border-primary/10 rounded-xl overflow-hidden">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
              <CardTitle className="text-xl anime-gradient-text">
                文件转移设置
              </CardTitle>
              <CardDescription>配置文件转移方式和媒体库位置</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6 p-6">
              {/* 基础配置 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">基础配置</h3>
                <div className="space-y-2">
                  <Label htmlFor="transfer-interval">检查间隔（分钟）</Label>
                  <Input
                    id="transfer-interval"
                    type="number"
                    value={transferConfig.interval}
                    onChange={(e) =>
                      setTransferConfig((prev) => ({
                        ...prev,
                        interval: parseInt(e.target.value) || 15,
                      }))
                    }
                    className="rounded-xl placeholder-gray-400"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="transfer-method">文件转移方式</Label>
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-5 w-5 rounded-full"
                          >
                            <Info className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>
                            硬链接：同一文件的多个引用，占用相同空间，删除源文件不影响
                          </p>
                          <p>优点：不占用额外空间，源文件删除后仍可访问</p>
                          <p>
                            缺点：必须在同一文件系统内，无法跨磁盘。Docker运行时需要直接映射源目录和目的目录或媒体库目录的上级目录
                          </p>
                          <br />
                          <p>软链接：类似快捷方式，指向原始文件</p>
                          <p>优点：可跨磁盘/文件系统</p>
                          <p>缺点：源文件删除后链接失效，需保持源文件。并且额外占用极少量的磁盘空间</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                  <Select
                    value={transferConfig.transferType}
                    onValueChange={(value: "hardlink") =>
                      setTransferConfig((prev) => ({
                        ...prev,
                        transferType: value,
                      }))
                    }
                  >
                    <SelectTrigger id="transfer-method" className="rounded-xl">
                      <SelectValue placeholder="选择文件转移方式" />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      <SelectItem value="hardlink">硬链接</SelectItem>
                      <SelectItem value="softlink">软链接</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* 媒体库配置 */}
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">媒体库配置</h3>
                <div className="space-y-2">
                  <Label htmlFor="transfer-anime-path">番剧路径</Label>
                  <Input
                    id="transfer-anime-path"
                    value={transferConfig.tvPath}
                    onChange={(e) =>
                      setTransferConfig((prev) => ({
                        ...prev,
                        tvPath: e.target.value,
                      }))
                    }
                    placeholder="/path/to/anime/library"
                    className="rounded-xl placeholder-gray-400"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="transfer-tv-format">番剧重命名格式</Label>
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-5 w-5 rounded-full"
                          >
                            <Info className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>{`在文件转移时，会按照定义的格式对媒体库文件做重命名并转移`}</p>
                          <p>{`重命名格式本身是媒体库路径下的相对路径`}</p>
                          <p>{`支持使用占位符变量，当前支持的变量如下：`}</p>
                          <p>{`{name}=番剧名称`}</p>
                          <p>{`{year}=年份`}</p>
                          <p>{`{season}=季数`}</p>
                          <p>{`{episode}=集数`}</p>
                          <p>{`{season_episode}=季集数，等价于S{season}E{episode}`}</p>
                          <p>{`{origin_name}=原始文件名，不包含扩展名`}</p>
                          <p>{`{release_group}=压制组/字幕组`}</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                  <Input
                    id="transfer-tv-format"
                    value={transferConfig.tvFormat}
                    onChange={(e) =>
                      setTransferConfig((prev) => ({
                        ...prev,
                        tvFormat: e.target.value,
                      }))
                    }
                    placeholder="{name}/Season {season}/{name} - S{season}E{episode}.{ext}"
                    className="rounded-xl placeholder-gray-400"
                  />
                </div>
              </div>

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("transfer")}
              >
                保存文件转移设置
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notice">
          <Card className="border-primary/10 rounded-xl overflow-hidden">
            <CardHeader className="bg-gradient-to-r from-primary/5 to-blue-500/5">
              <CardTitle className="text-xl anime-gradient-text">
                通知设置
              </CardTitle>
              <CardDescription>配置通知方式和相关参数</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6 p-6">
              {/* 通知开关 */}
              <div className="space-y-4">
                <div className="flex items-center space-x-2">
                  <Switch
                    id="notice-enabled"
                    checked={noticeConfig.enabled}
                    onCheckedChange={(checked) =>
                      setNoticeConfig((prev) => ({
                        ...prev,
                        enabled: checked,
                      }))
                    }
                  />
                  <Label htmlFor="notice-enabled" className="font-medium">
                    启用通知
                  </Label>
                </div>
              </div>

              {noticeConfig.enabled && (
                <>
                  {/* 通知类型选择 */}
                  <div className="space-y-4">
                    <h3 className="text-lg font-semibold">通知类型</h3>
                    <div className="space-y-2">
                      <Select
                        value={noticeConfig.type}
                        onValueChange={(value: "telegram" | "email" | "bark") =>
                          setNoticeConfig((prev) => ({
                            ...prev,
                            type: value,
                          }))
                        }
                      >
                        <SelectTrigger className="rounded-xl">
                          <SelectValue placeholder="选择通知类型" />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                          <SelectItem value="telegram">Telegram</SelectItem>
                          <SelectItem value="email">邮件</SelectItem>
                          <SelectItem value="bark">Bark</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* Telegram设置 */}
                  {noticeConfig.type === "telegram" && (
                    <div className="space-y-4 pl-4 border-l-2 border-primary/10">
                      <h3 className="text-lg font-semibold">Telegram设置</h3>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="telegram-token">Bot Token</Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>从BotFather获取的Bot Token</p>
                                <p>
                                  格式: 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ
                                </p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="telegram-token"
                          value={noticeConfig.telegram.token}
                          onChange={(e) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              telegram: {
                                ...prev.telegram,
                                token: e.target.value,
                              },
                            }))
                          }
                          placeholder="输入您的Telegram Bot Token"
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="telegram-chat-id">Chat ID</Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>目标聊天的Chat ID</p>
                                <p>可通过@userinfobot获取个人账号的Chat ID</p>
                                <p>群组可通过@RawDataBot获取</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="telegram-chat-id"
                          type="number"
                          value={noticeConfig.telegram.chatID || ""}
                          onChange={(e) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              telegram: {
                                ...prev.telegram,
                                chatID: parseInt(e.target.value) || 0,
                              },
                            }))
                          }
                          placeholder="输入目标聊天ID"
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>
                    </div>
                  )}

                  {/* 邮件设置 */}
                  {noticeConfig.type === "email" && (
                    <div className="space-y-4 pl-4 border-l-2 border-primary/10">
                      <h3 className="text-lg font-semibold">邮件设置</h3>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="email-server">SMTP服务器</Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>邮件服务器地址，不包含端口</p>
                                <p>例如: smtp.gmail.com, smtp.qq.com</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="email-server"
                          value={noticeConfig.email.host}
                          onChange={(e) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              email: {
                                ...prev.email,
                                host: e.target.value,
                              },
                            }))
                          }
                          placeholder="smtp.example.com"
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-center space-x-2">
                          <Switch
                            id="email-ssl"
                            checked={noticeConfig.email.ssl}
                            onCheckedChange={(checked) =>
                              setNoticeConfig((prev) => ({
                                ...prev,
                                email: {
                                  ...prev.email,
                                  ssl: checked,
                                },
                              }))
                            }
                          />
                          <Label htmlFor="email-ssl">使用SSL</Label>
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-2">
                          <Label htmlFor="email-username">用户名</Label>
                          <Input
                            id="email-username"
                            value={noticeConfig.email.username}
                            onChange={(e) =>
                              setNoticeConfig((prev) => ({
                                ...prev,
                                email: {
                                  ...prev.email,
                                  username: e.target.value,
                                },
                              }))
                            }
                            placeholder="username@example.com"
                            className="rounded-xl placeholder-gray-400"
                          />
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="email-password">密码</Label>
                          <PasswordInput
                            id="email-password"
                            value={noticeConfig.email.password}
                            onChange={(e) =>
                              setNoticeConfig((prev) => ({
                                ...prev,
                                email: {
                                  ...prev.email,
                                  password: e.target.value,
                                },
                              }))
                            }
                            className="rounded-xl placeholder-gray-400"
                          />
                        </div>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="email-from">发件人</Label>
                        <Input
                          id="email-from"
                          value={noticeConfig.email.from}
                          onChange={(e) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              email: {
                                ...prev.email,
                                from: e.target.value,
                              },
                            }))
                          }
                          placeholder="BangumiBuddy <noreply@example.com>"
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="email-to">收件人</Label>
                        <Input
                          id="email-to"
                          value={noticeConfig.email.to.join(", ")}
                          onChange={(e) =>
                            handleEmailRecipientsChange(e.target.value)
                          }
                          placeholder="user1@example.com, user2@example.com"
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>
                    </div>
                  )}

                  {/* Bark设置 */}
                  {noticeConfig.type === "bark" && (
                    <div className="space-y-4 pl-4 border-l-2 border-primary/10">
                      <h3 className="text-lg font-semibold">Bark设置</h3>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="bark-server-url">
                            服务器通知地址
                          </Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>
                                  Bark服务器通知地址，一般为https://api.day.app/your_device_key
                                </p>
                                <p>自定义服务器需包含http://或https://前缀</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="bark-server-url"
                          value={noticeConfig.bark.serverPath}
                          onChange={(e) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              bark: {
                                ...prev.bark,
                                serverPath: e.target.value,
                              },
                            }))
                          }
                          placeholder="https://api.day.app/your_device_key"
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="bark-sound">通知铃声</Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>通知铃声名称，留空则不启用铃声</p>
                                <p>支持系统铃声如alarm, bell, chime等</p>
                                <p>完整列表请参考Bark应用</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="bark-sound"
                          value={noticeConfig.bark.sound}
                          onChange={(e) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              bark: {
                                ...prev.bark,
                                sound: e.target.value,
                              },
                            }))
                          }
                          placeholder=""
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="bark-interruption">中断级别</Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>通知中断级别，控制通知显示方式</p>
                                <p>
                                  passive: 仅将通知添加到通知列表，不会亮屏提醒
                                </p>
                                <p>active: 默认值，系统会立即亮屏显示通知</p>
                                <p>
                                  timeSensitive:
                                  时效性通知，可在专注状态下显示通知
                                </p>
                                <p>critical: 重要警告, 在静音模式下也会响铃</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                        <Select
                          value={noticeConfig.bark.interruption}
                          onValueChange={(value) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              bark: {
                                ...prev.bark,
                                interruption: value,
                              },
                            }))
                          }
                        >
                          <SelectTrigger className="rounded-xl">
                            <SelectValue placeholder="选择中断级别" />
                          </SelectTrigger>
                          <SelectContent className="rounded-xl">
                            <SelectItem value="active">active</SelectItem>
                            <SelectItem value="passive">passive</SelectItem>
                            <SelectItem value="timeSensitive">
                              timeSensitive
                            </SelectItem>
                            <SelectItem value="critical">critical</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-center space-x-2">
                          <Switch
                            id="bark-auto-save"
                            checked={noticeConfig.bark.autoSave}
                            onCheckedChange={(checked) =>
                              setNoticeConfig((prev) => ({
                                ...prev,
                                bark: {
                                  ...prev.bark,
                                  autoSave: checked,
                                },
                              }))
                            }
                          />
                          <Label htmlFor="bark-auto-save">自动保存通知</Label>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>开启后，通知会自动保存到客户端</p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* 通知点配置 */}
                  <div className="space-y-4 mt-6">
                    <div className="flex items-center gap-2">
                      <h3 className="text-lg font-semibold">事件订阅</h3>
                      <TooltipProvider>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-5 w-5 rounded-full"
                            >
                              <Info className="h-3.5 w-3.5 text-muted-foreground" />
                            </Button>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>选择需要接收通知的事件类型</p>
                            <br />
                            <p>番剧更新: 当发现新剧集时通知</p>
                            <p>下载完成: 当下载任务完成时通知</p>
                            <p>转移媒体库: 当文件成功转移到媒体库时通知</p>
                            <p>异常通知: 当系统发生错误时通知</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </div>
                    <div className="space-y-4 pl-4 border-l-2 border-primary/10">
                      <div className="flex items-center justify-between">
                        <Label htmlFor="notice-subscription" className="flex-1">
                          番剧更新
                        </Label>
                        <Switch
                          id="notice-subscription"
                          checked={
                            noticeConfig.noticePoints.subscriptionUpdated
                          }
                          onCheckedChange={(checked) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              noticePoints: {
                                ...prev.noticePoints,
                                subscriptionUpdated: checked,
                              },
                            }))
                          }
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <Label htmlFor="notice-downloaded" className="flex-1">
                          下载完成
                        </Label>
                        <Switch
                          id="notice-downloaded"
                          checked={noticeConfig.noticePoints.downloaded}
                          onCheckedChange={(checked) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              noticePoints: {
                                ...prev.noticePoints,
                                downloaded: checked,
                              },
                            }))
                          }
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <Label htmlFor="notice-transferred" className="flex-1">
                          转移媒体库
                        </Label>
                        <Switch
                          id="notice-transferred"
                          checked={noticeConfig.noticePoints.transferred}
                          onCheckedChange={(checked) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              noticePoints: {
                                ...prev.noticePoints,
                                transferred: checked,
                              },
                            }))
                          }
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <Label htmlFor="notice-error" className="flex-1">
                          异常通知
                        </Label>
                        <Switch
                          id="notice-error"
                          checked={noticeConfig.noticePoints.error}
                          onCheckedChange={(checked) =>
                            setNoticeConfig((prev) => ({
                              ...prev,
                              noticePoints: {
                                ...prev.noticePoints,
                                error: checked,
                              },
                            }))
                          }
                        />
                      </div>
                    </div>
                  </div>
                </>
              )}

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("notice")}
              >
                保存通知设置
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
