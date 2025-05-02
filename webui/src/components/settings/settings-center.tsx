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
import { TooltipProvider } from "@/components/ui/tooltip";
import {
  HybridTooltip,
  HybridTooltipContent,
  HybridTooltipTrigger,
} from "@/components/common/tooltip";
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
import { MatchInput } from "@/components/common/match-input";
import { Switch } from "@/components/ui/switch";
import { PasswordInput } from "../common/password-input";
import { extractErrorMessage } from "@/utils/error";

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

  // 下载设置表单验证
  const [downloadErrors, setDownloadErrors] = useState({
    tvSavePath: "",
    movieSavePath: "",
    host: "",
    username: "",
    password: "",
  });

  // TMDB配置状态
  const [tmdbConfig, setTmdbConfig] = useState<TMDBConfig>({
    tmdbToken: "",
    alternateURL: false,
  });

  // 元数据设置表单验证
  const [metadataErrors, setMetadataErrors] = useState({
    tmdbToken: "",
  });

  // 订阅配置状态
  const [subscriptionConfig, setSubscriptionConfig] =
    useState<SubscriptionConfig>({
      rssCheckInterval: 30,
      includeRegs: [],
      excludeRegs: [],
      autoStop: false,
    });

  // 订阅设置表单验证
  const [subscriptionErrors, setSubscriptionErrors] = useState({
    rssCheckInterval: "",
  });

  // 文件转移配置状态
  const [transferConfig, setTransferConfig] = useState<TransferConfig>({
    interval: 15,
    tvPath: "",
    tvFormat: "",
    transferType: "hardlink",
  });

  // 文件转移设置表单验证
  const [transferErrors, setTransferErrors] = useState({
    interval: "",
    tvPath: "",
    tvFormat: "",
    transferType: "",
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

  // 通知设置表单验证
  const [noticeErrors, setNoticeErrors] = useState({
    type: "",
    // Telegram 错误
    telegramToken: "",
    telegramChatID: "",
    // Email 错误
    emailHost: "",
    emailUsername: "",
    emailPassword: "",
    emailFrom: "",
    emailTo: "",
    // Bark 错误
    barkServerPath: "",
  });

  // 验证下载设置表单字段
  const validateDownloadField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "tvSavePath":
        if (!value.trim()) error = "请填写番剧保存位置";
        break;
      case "movieSavePath":
        if (!value.trim()) error = "请填写剧场版保存位置";
        break;
      case "host":
        if (!value.trim()) error = "请填写下载器地址";
        else if (!value.startsWith("http://") && !value.startsWith("https://"))
          error = "下载器地址必须以http://或https://开头";
        break;
      case "username":
        if (!value.trim()) error = "请填写用户名";
        break;
      case "password":
        if (!value.trim()) error = "请填写密码";
        break;
    }
    setDownloadErrors((prev) => ({ ...prev, [field]: error }));
    return !error;
  };

  // 更新下载管理器配置并验证
  const updateDownloadManagerConfig = (
    field: keyof DownloadManagerConfig,
    value: string
  ) => {
    validateDownloadField(field, value);
    setDownloadManagerConfig((prev) => ({
      ...prev,
      [field]: value,
    }));
    setDownloadManagerConfigChanged(true);
  };

  // 更新下载器配置并验证
  const updateDownloaderConfig = (field: string, value: string) => {
    validateDownloadField(field, value);
    setDownloaderConfig((prev) => ({
      ...prev,
      qbittorrent: {
        ...prev.qbittorrent,
        [field]: value,
      },
    }));
    setDownloaderConfigChanged(true);
  };

  // 验证订阅设置表单字段
  const validateSubscriptionField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "rssCheckInterval":
        if (!value || value <= 0) error = "检查间隔必须大于0";
        break;
    }
    setSubscriptionErrors((prev) => ({ ...prev, [field]: error }));
    return !error;
  };

  // 更新订阅配置并验证
  const updateSubscriptionConfig = (
    field: keyof SubscriptionConfig,
    value: any
  ) => {
    if (field === "rssCheckInterval") {
      validateSubscriptionField(field, value);
    }
    setSubscriptionConfig((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 验证元数据设置表单字段
  const validateMetadataField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "tmdbToken":
        if (!value.trim()) error = "请填写TMDB Token";
        break;
    }
    setMetadataErrors((prev) => ({ ...prev, [field]: error }));
    return !error;
  };

  // 更新TMDB配置并验证
  const updateTmdbConfig = (field: keyof TMDBConfig, value: any) => {
    if (field === "tmdbToken") {
      validateMetadataField(field, value);
    }
    setTmdbConfig((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 验证文件转移设置表单字段
  const validateTransferField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "interval":
        if (!value || value <= 0) error = "检查间隔必须大于0";
        break;
      case "tvPath":
        if (!value.trim()) error = "请填写番剧路径";
        break;
      case "tvFormat":
        if (!value.trim()) error = "请填写番剧重命名格式";
        else if (!value.includes("{name}")) error = "格式中必须包含{name}变量";
        break;
      case "transferType":
        if (!value.trim()) error = "请选择转移类型";
        break;
    }
    setTransferErrors((prev) => ({ ...prev, [field]: error }));
    return !error;
  };

  // 更新文件转移配置并验证
  const updateTransferConfig = (field: keyof TransferConfig, value: any) => {
    validateTransferField(field, value);
    setTransferConfig((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  // 验证通知设置表单字段
  const validateNoticeField = (field: string, value: any) => {
    let error = "";
    switch (field) {
      case "type":
        if (!value.trim()) error = "请选择通知类型";
        break;
      // Telegram 字段
      case "telegramToken":
        if (!value.trim()) error = "请填写Bot Token";
        else if (!value.includes(":")) error = "Bot Token格式不正确";
        break;
      case "telegramChatID":
        if (!value) error = "请填写Chat ID";
        break;

      // Email 字段
      case "emailHost":
        if (!value.trim()) error = "请填写SMTP服务器";
        break;
      case "emailUsername":
        if (!value.trim()) error = "请填写用户名";
        break;
      case "emailPassword":
        if (!value.trim()) error = "请填写密码";
        break;
      case "emailFrom":
        if (!value.trim()) error = "请填写发件人";
        break;
      case "emailTo":
        if (!value.trim()) error = "请填写收件人";
        break;

      // Bark 字段
      case "barkServerPath":
        if (!value.trim()) error = "请填写服务器通知地址";
        else if (!value.startsWith("http://") && !value.startsWith("https://"))
          error = "服务器通知地址必须以http://或https://开头";
        break;
    }
    setNoticeErrors((prev) => ({ ...prev, [field]: error }));
    return !error;
  };

  // 更新Telegram配置并验证
  const updateTelegramConfig = (
    field: keyof typeof noticeConfig.telegram,
    value: any
  ) => {
    const noticeField = field === "token" ? "telegramToken" : "telegramChatID";
    validateNoticeField(noticeField, value);
    setNoticeConfig((prev) => ({
      ...prev,
      telegram: {
        ...prev.telegram,
        [field]: value,
      },
    }));
  };

  // 更新Email配置并验证
  const updateEmailConfig = (
    field: keyof typeof noticeConfig.email,
    value: any
  ) => {
    const noticeField = `email${
      field.charAt(0).toUpperCase() + field.slice(1)
    }`;
    if (field !== "ssl") {
      validateNoticeField(noticeField, value);
    }
    setNoticeConfig((prev) => ({
      ...prev,
      email: {
        ...prev.email,
        [field]: value,
      },
    }));
  };

  // 更新Bark配置并验证
  const updateBarkConfig = (
    field: keyof typeof noticeConfig.bark,
    value: any
  ) => {
    if (field === "serverPath") {
      validateNoticeField("barkServerPath", value);
    }
    setNoticeConfig((prev) => ({
      ...prev,
      bark: {
        ...prev.bark,
        [field]: value,
      },
    }));
  };

  // 检查各tab表单是否有错误
  const hasDownloadErrors = () => {
    return Object.values(downloadErrors).some((error) => error !== "");
  };

  const hasMetadataErrors = () => {
    return Object.values(metadataErrors).some((error) => error !== "");
  };

  const hasSubscriptionErrors = () => {
    return Object.values(subscriptionErrors).some((error) => error !== "");
  };

  const hasTransferErrors = () => {
    return Object.values(transferErrors).some((error) => error !== "");
  };

  const hasNoticeErrors = () => {
    // 只检查当前启用的通知类型的错误
    if (!noticeConfig.enabled) return false;

    const relevantErrors: string[] = [];
    switch (noticeConfig.type) {
      case "telegram":
        relevantErrors.push(
          noticeErrors.telegramToken,
          noticeErrors.telegramChatID
        );
        break;
      case "email":
        relevantErrors.push(
          noticeErrors.emailHost,
          noticeErrors.emailUsername,
          noticeErrors.emailPassword,
          noticeErrors.emailFrom,
          noticeErrors.emailTo
        );
        break;
      case "bark":
        relevantErrors.push(noticeErrors.barkServerPath);
        break;
      default:
        relevantErrors.push(noticeErrors.type);
    }

    return relevantErrors.some((error) => error !== "");
  };

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

      // 初始验证所有表单字段
      validateInitialFields(
        downloader,
        manager,
        tmdb,
        subscription,
        transfer,
        notice
      );
    } catch (error) {
      const desc = extractErrorMessage(error);
      toast({
        title: "加载配置失败",
        description: desc,
        variant: "destructive",
      });
    }
  };

  // 初始验证所有配置字段
  const validateInitialFields = (
    downloader: DownloaderConfig,
    manager: DownloadManagerConfig,
    tmdb: TMDBConfig,
    subscription: SubscriptionConfig,
    transfer: TransferConfig,
    notice: NoticeConfig
  ) => {
    // 验证下载设置
    validateDownloadField("tvSavePath", manager.tvSavePath);
    validateDownloadField("movieSavePath", manager.movieSavePath);
    validateDownloadField("host", downloader.qbittorrent.host);
    validateDownloadField("username", downloader.qbittorrent.username);
    validateDownloadField("password", downloader.qbittorrent.password);

    // 验证元数据设置
    validateMetadataField("tmdbToken", tmdb.tmdbToken);

    // 验证订阅设置
    validateSubscriptionField(
      "rssCheckInterval",
      subscription.rssCheckInterval
    );

    // 验证文件转移设置
    validateTransferField("interval", transfer.interval);
    validateTransferField("tvPath", transfer.tvPath);
    validateTransferField("tvFormat", transfer.tvFormat);
    validateTransferField("transferType", transfer.transferType);

    // 验证通知设置
    if (notice.enabled) {
      validateNoticeSettings(notice);
    }
  };

  // 验证所有通知设置
  const validateNoticeSettings = (notice: NoticeConfig) => {
    validateNoticeField("type", notice.type);
    switch (notice.type) {
      case "telegram":
        validateNoticeField("telegramToken", notice.telegram.token);
        validateNoticeField("telegramChatID", notice.telegram.chatID);
        break;
      case "email":
        validateNoticeField("emailHost", notice.email.host);
        validateNoticeField("emailUsername", notice.email.username);
        validateNoticeField("emailPassword", notice.email.password);
        validateNoticeField("emailFrom", notice.email.from);
        validateNoticeField("emailTo", notice.email.to?.join(",") || "");
        break;
      case "bark":
        validateNoticeField("barkServerPath", notice.bark.serverPath);
        break;
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
      const desc = extractErrorMessage(error);
      toast({
        title: "下载器连接失败",
        description: desc,
        variant: "destructive",
      });
    } finally {
      setIsChecking(false);
    }
  };

  // 更新收件人列表
  const handleEmailRecipientsChange = (value: string) => {
    validateNoticeField("emailTo", value);
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
      const desc = extractErrorMessage(error);
      toast({
        title: "保存设置失败",
        description: desc,
        variant: "destructive",
      });
    }
  };

  // 更新通知点配置
  const updateNoticePoints = (
    field: keyof typeof noticeConfig.noticePoints,
    value: boolean
  ) => {
    setNoticeConfig((prev) => ({
      ...prev,
      noticePoints: {
        ...prev.noticePoints,
        [field]: value,
      },
    }));
  };

  // 更新通知类型
  const updateNoticeType = (type: "telegram" | "email" | "bark") => {
    setNoticeConfig((prev) => {
      const newConfig = {
        ...prev,
        type: type,
      };
      validateNoticeSettings(newConfig);
      return newConfig;
    });
  };

  // 更新通知启用状态
  const updateNoticeEnabled = (enabled: boolean) => {
    setNoticeConfig((prev) => {
      const newConfig = {
        ...prev,
        enabled: enabled,
      };
      validateNoticeSettings(newConfig);
      return newConfig;
    });
  };

  return (
    <div className="space-y-6 h-[calc(100dvh-6rem)] overflow-y-auto pb-10 scrollbar-hide">
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
            <Search className="icon-button" />
            <span className="hidden sm:inline">元数据设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="subscription"
            className="flex items-center gap-2 rounded-xl"
          >
            <Tv className="icon-button" />
            <span className="hidden sm:inline">订阅设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="transfer"
            className="flex items-center gap-2 rounded-xl"
          >
            <FolderSync className="icon-button" />
            <span className="hidden sm:inline">文件转移设置</span>
          </TabsTrigger>
          <TabsTrigger
            value="notice"
            className="flex items-center gap-2 rounded-xl"
          >
            <Bell className="icon-button" />
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
                      updateDownloadManagerConfig("tvSavePath", e.target.value);
                    }}
                    placeholder="/path/to/anime"
                    className={`rounded-xl ${
                      downloadErrors.tvSavePath ? "border-destructive" : ""
                    }`}
                  />
                  {downloadErrors.tvSavePath && (
                    <span className="text-sm text-destructive">
                      {downloadErrors.tvSavePath}
                    </span>
                  )}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="movie-path">剧场版保存位置</Label>
                  <Input
                    id="movie-path"
                    value={downloadManagerConfig.movieSavePath}
                    onChange={(e) => {
                      updateDownloadManagerConfig(
                        "movieSavePath",
                        e.target.value
                      );
                    }}
                    placeholder="/path/to/movies"
                    className={`rounded-xl ${
                      downloadErrors.movieSavePath ? "border-destructive" : ""
                    }`}
                  />
                  {downloadErrors.movieSavePath && (
                    <span className="text-sm text-destructive">
                      {downloadErrors.movieSavePath}
                    </span>
                  )}
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
                        updateDownloaderConfig("host", e.target.value);
                      }}
                      placeholder="http://localhost:8080"
                      className={`rounded-xl placeholder-gray-400 ${
                        downloadErrors.host ? "border-destructive" : ""
                      }`}
                    />
                    {downloadErrors.host && (
                      <span className="text-sm text-destructive">
                        {downloadErrors.host}
                      </span>
                    )}
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="downloader-username">用户名</Label>
                      <Input
                        id="downloader-username"
                        value={downloaderConfig.qbittorrent.username}
                        onChange={(e) => {
                          updateDownloaderConfig("username", e.target.value);
                        }}
                        placeholder="admin"
                        className={`rounded-xl placeholder-gray-400 ${
                          downloadErrors.username ? "border-destructive" : ""
                        }`}
                      />
                      {downloadErrors.username && (
                        <span className="text-sm text-destructive">
                          {downloadErrors.username}
                        </span>
                      )}
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="downloader-password">密码</Label>
                      <PasswordInput
                        id="downloader-password"
                        value={downloaderConfig.qbittorrent.password}
                        onChange={(e) => {
                          updateDownloaderConfig("password", e.target.value);
                        }}
                        className={`rounded-xl placeholder-gray-400 ${
                          downloadErrors.password ? "border-destructive" : ""
                        }`}
                      />
                      {downloadErrors.password && (
                        <span className="text-sm text-destructive">
                          {downloadErrors.password}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("download")}
                disabled={hasDownloadErrors()}
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
                    updateTmdbConfig("tmdbToken", e.target.value)
                  }
                  placeholder="输入您的TMDB Token"
                  className={`rounded-xl placeholder-gray-400 ${
                    metadataErrors.tmdbToken ? "border-destructive" : ""
                  }`}
                />
                {metadataErrors.tmdbToken && (
                  <span className="text-sm text-destructive">
                    {metadataErrors.tmdbToken}
                  </span>
                )}
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="use-alternative-url">使用备用地址</Label>
                  <TooltipProvider>
                    <HybridTooltip>
                      <HybridTooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-5 w-5 rounded-full"
                        >
                          <Info className="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                      </HybridTooltipTrigger>
                      <HybridTooltipContent>
                        <p>备用地址使用api.tmdb.org/3，无代理的情况下可选</p>
                      </HybridTooltipContent>
                    </HybridTooltip>
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
                disabled={hasMetadataErrors()}
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
                  placeholder="30"
                  type="number"
                  value={subscriptionConfig.rssCheckInterval}
                  onChange={(e) =>
                    updateSubscriptionConfig(
                      "rssCheckInterval",
                      parseInt(e.target.value)
                    )
                  }
                  className={`rounded-xl placeholder-gray-400 ${
                    subscriptionErrors.rssCheckInterval
                      ? "border-destructive"
                      : ""
                  }`}
                />
                {subscriptionErrors.rssCheckInterval && (
                  <span className="text-sm text-destructive">
                    {subscriptionErrors.rssCheckInterval}
                  </span>
                )}
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <Label htmlFor="auto-stop">自动停止订阅</Label>
                  <TooltipProvider>
                    <HybridTooltip>
                      <HybridTooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-5 w-5 rounded-full"
                        >
                          <Info className="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                      </HybridTooltipTrigger>
                      <HybridTooltipContent>
                        <p>
                          当下载到最后一集时，自动停止订阅（可以避免下载字幕组后续更新的合集），按需启动
                        </p>
                      </HybridTooltipContent>
                    </HybridTooltip>
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
                disabled={hasSubscriptionErrors()}
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
                    placeholder="1"
                    value={transferConfig.interval || ""}
                    min={1}
                    onChange={(e) =>
                      updateTransferConfig("interval", parseInt(e.target.value))
                    }
                    className={`rounded-xl placeholder-gray-400 ${
                      transferErrors.interval ? "border-destructive" : ""
                    }`}
                  />
                  {transferErrors.interval && (
                    <span className="text-sm text-destructive">
                      {transferErrors.interval}
                    </span>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="transfer-method">文件转移方式</Label>
                    <TooltipProvider>
                      <HybridTooltip>
                        <HybridTooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-5 w-5 rounded-full"
                          >
                            <Info className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        </HybridTooltipTrigger>
                        <HybridTooltipContent>
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
                          <p>
                            缺点：源文件删除后链接失效，需保持源文件。并且额外占用极少量的磁盘空间
                          </p>
                        </HybridTooltipContent>
                      </HybridTooltip>
                    </TooltipProvider>
                  </div>
                  <Select
                    value={transferConfig.transferType}
                    onValueChange={(value: "hardlink" | "softlink") => {
                      validateTransferField("transferType", value);
                      setTransferConfig((prev) => ({
                        ...prev,
                        transferType: value,
                      }));
                    }}
                  >
                    <SelectTrigger
                      id="transfer-method"
                      className={`rounded-xl ${
                        transferErrors.transferType ? "border-destructive" : ""
                      }`}
                    >
                      <SelectValue placeholder="选择文件转移方式" />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      <SelectItem value="hardlink">硬链接</SelectItem>
                      <SelectItem value="softlink">软链接</SelectItem>
                    </SelectContent>
                  </Select>
                  {transferErrors.transferType && (
                    <span className="text-sm text-destructive">
                      {transferErrors.transferType}
                    </span>
                  )}
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
                      updateTransferConfig("tvPath", e.target.value)
                    }
                    placeholder="/path/to/anime/library"
                    className={`rounded-xl placeholder-gray-400 ${
                      transferErrors.tvPath ? "border-destructive" : ""
                    }`}
                  />
                  {transferErrors.tvPath && (
                    <span className="text-sm text-destructive">
                      {transferErrors.tvPath}
                    </span>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="transfer-tv-format">番剧重命名格式</Label>
                    <TooltipProvider>
                      <HybridTooltip>
                        <HybridTooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-5 w-5 rounded-full"
                          >
                            <Info className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        </HybridTooltipTrigger>
                        <HybridTooltipContent>
                          <p>{`在文件转移时，会按照定义的格式对媒体库文件做重命名并转移`}</p>
                          <p>{`重命名格式本身是媒体库路径下的相对路径`}</p>
                          <p>{`支持使用占位符变量，当前支持的变量如下：`}</p>
                          <p>{`{name}=番剧名称`}</p>
                          <p>{`{year}=年份`}</p>
                          <p>{`{season}=季数`}</p>
                          <p>{`{episode}=集数`}</p>
                          <p>{`{season_episode}=季集数，等价于SXXEXX，与{season}和{episode}的区别是，如果季数或集数小于10，会自动补0`}</p>
                          <p>{`{origin_name}=原始文件名，不包含扩展名`}</p>
                          <p>{`{release_group}=压制组/字幕组`}</p>
                        </HybridTooltipContent>
                      </HybridTooltip>
                    </TooltipProvider>
                  </div>
                  <Input
                    id="transfer-tv-format"
                    value={transferConfig.tvFormat}
                    onChange={(e) =>
                      updateTransferConfig("tvFormat", e.target.value)
                    }
                    placeholder="{name}/Season {season}/{name} {season_episode}"
                    className={`rounded-xl placeholder-gray-400 ${
                      transferErrors.tvFormat ? "border-destructive" : ""
                    }`}
                  />
                  {transferErrors.tvFormat && (
                    <span className="text-sm text-destructive">
                      {transferErrors.tvFormat}
                    </span>
                  )}
                </div>
              </div>

              <Button
                className="w-full rounded-xl bg-gradient-to-r from-primary to-blue-500 anime-button"
                onClick={() => handleSaveSettings("transfer")}
                disabled={hasTransferErrors()}
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
                    onCheckedChange={(checked) => {
                      updateNoticeEnabled(checked);
                    }}
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
                          updateNoticeType(value)
                        }
                      >
                        <SelectTrigger
                          className={`rounded-xl ${
                            noticeErrors.type ? "border-destructive" : ""
                          }`}
                        >
                          <SelectValue placeholder="选择通知类型" />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                          <SelectItem value="telegram">Telegram</SelectItem>
                          <SelectItem value="email">邮件</SelectItem>
                          <SelectItem value="bark">Bark</SelectItem>
                        </SelectContent>
                      </Select>
                      {noticeErrors.type && (
                        <span className="text-sm text-destructive">
                          {noticeErrors.type}
                        </span>
                      )}
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
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
                                <p>从BotFather获取的Bot Token</p>
                                <p>
                                  格式: 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ
                                </p>
                              </HybridTooltipContent>
                            </HybridTooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="telegram-token"
                          value={noticeConfig.telegram.token}
                          onChange={(e) =>
                            updateTelegramConfig("token", e.target.value)
                          }
                          placeholder="输入您的Telegram Bot Token"
                          className={`rounded-xl placeholder-gray-400 ${
                            noticeErrors.telegramToken
                              ? "border-destructive"
                              : ""
                          }`}
                        />
                        {noticeErrors.telegramToken && (
                          <span className="text-sm text-destructive">
                            {noticeErrors.telegramToken}
                          </span>
                        )}
                      </div>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="telegram-chat-id">Chat ID</Label>
                          <TooltipProvider>
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
                                <p>Chat ID</p>
                                <p>可通过@userinfobot获取个人账号的Chat ID</p>
                                <p>群组可通过@RawDataBot获取</p>
                              </HybridTooltipContent>
                            </HybridTooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="telegram-chat-id"
                          type="number"
                          value={noticeConfig.telegram.chatID || ""}
                          onChange={(e) =>
                            updateTelegramConfig(
                              "chatID",
                              parseInt(e.target.value) || 0
                            )
                          }
                          placeholder="输入Chat ID"
                          className={`rounded-xl placeholder-gray-400 ${
                            noticeErrors.telegramChatID
                              ? "border-destructive"
                              : ""
                          }`}
                        />
                        {noticeErrors.telegramChatID && (
                          <span className="text-sm text-destructive">
                            {noticeErrors.telegramChatID}
                          </span>
                        )}
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
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
                                <p>邮件服务器地址，不包含端口</p>
                                <p>例如: smtp.gmail.com, smtp.qq.com</p>
                              </HybridTooltipContent>
                            </HybridTooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="email-server"
                          value={noticeConfig.email.host}
                          onChange={(e) =>
                            updateEmailConfig("host", e.target.value)
                          }
                          placeholder="smtp.example.com"
                          className={`rounded-xl placeholder-gray-400 ${
                            noticeErrors.emailHost ? "border-destructive" : ""
                          }`}
                        />
                        {noticeErrors.emailHost && (
                          <span className="text-sm text-destructive">
                            {noticeErrors.emailHost}
                          </span>
                        )}
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-center space-x-2">
                          <Switch
                            id="email-ssl"
                            checked={noticeConfig.email.ssl}
                            onCheckedChange={(checked) =>
                              updateEmailConfig("ssl", checked)
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
                              updateEmailConfig("username", e.target.value)
                            }
                            placeholder="username@example.com"
                            className={`rounded-xl placeholder-gray-400 ${
                              noticeErrors.emailUsername
                                ? "border-destructive"
                                : ""
                            }`}
                          />
                          {noticeErrors.emailUsername && (
                            <span className="text-sm text-destructive">
                              {noticeErrors.emailUsername}
                            </span>
                          )}
                        </div>
                        <div className="space-y-2">
                          <Label htmlFor="email-password">密码</Label>
                          <PasswordInput
                            id="email-password"
                            value={noticeConfig.email.password}
                            onChange={(e) =>
                              updateEmailConfig("password", e.target.value)
                            }
                            className={`rounded-xl placeholder-gray-400 ${
                              noticeErrors.emailPassword
                                ? "border-destructive"
                                : ""
                            }`}
                          />
                          {noticeErrors.emailPassword && (
                            <span className="text-sm text-destructive">
                              {noticeErrors.emailPassword}
                            </span>
                          )}
                        </div>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="email-from">发件人</Label>
                        <Input
                          id="email-from"
                          value={noticeConfig.email.from}
                          onChange={(e) =>
                            updateEmailConfig("from", e.target.value)
                          }
                          placeholder="user1@example.com"
                          className={`rounded-xl placeholder-gray-400 ${
                            noticeErrors.emailFrom ? "border-destructive" : ""
                          }`}
                        />
                        {noticeErrors.emailFrom && (
                          <span className="text-sm text-destructive">
                            {noticeErrors.emailFrom}
                          </span>
                        )}
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="email-to">收件人</Label>
                        <Input
                          id="email-to"
                          value={noticeConfig.email.to?.join(",") || ""}
                          onChange={(e) =>
                            handleEmailRecipientsChange(e.target.value)
                          }
                          placeholder="user1@example.com,user2@example.com"
                          className={`rounded-xl placeholder-gray-400 ${
                            noticeErrors.emailTo ? "border-destructive" : ""
                          }`}
                        />
                        {noticeErrors.emailTo && (
                          <span className="text-sm text-destructive">
                            {noticeErrors.emailTo}
                          </span>
                        )}
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
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
                                <p>
                                  Bark服务器通知地址，一般为https://api.day.app/your_device_key
                                </p>
                                <p>自定义服务器需包含http://或https://前缀</p>
                              </HybridTooltipContent>
                            </HybridTooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="bark-server-url"
                          value={noticeConfig.bark.serverPath}
                          onChange={(e) =>
                            updateBarkConfig("serverPath", e.target.value)
                          }
                          placeholder="https://api.day.app/your_device_key"
                          className={`rounded-xl placeholder-gray-400 ${
                            noticeErrors.barkServerPath
                              ? "border-destructive"
                              : ""
                          }`}
                        />
                        {noticeErrors.barkServerPath && (
                          <span className="text-sm text-destructive">
                            {noticeErrors.barkServerPath}
                          </span>
                        )}
                      </div>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="bark-sound">通知铃声</Label>
                          <TooltipProvider>
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
                                <p>通知铃声名称，留空则不启用铃声</p>
                                <p>支持系统铃声如alarm, bell, chime等</p>
                                <p>完整列表请参考Bark应用</p>
                              </HybridTooltipContent>
                            </HybridTooltip>
                          </TooltipProvider>
                        </div>
                        <Input
                          id="bark-sound"
                          value={noticeConfig.bark.sound}
                          onChange={(e) =>
                            updateBarkConfig("sound", e.target.value)
                          }
                          placeholder=""
                          className="rounded-xl placeholder-gray-400"
                        />
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Label htmlFor="bark-interruption">中断级别</Label>
                          <TooltipProvider>
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
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
                              </HybridTooltipContent>
                            </HybridTooltip>
                          </TooltipProvider>
                        </div>
                        <Select
                          value={noticeConfig.bark.interruption}
                          onValueChange={(value) =>
                            updateBarkConfig("interruption", value)
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
                              updateBarkConfig("autoSave", checked)
                            }
                          />
                          <Label htmlFor="bark-auto-save">自动保存通知</Label>
                          <TooltipProvider>
                            <HybridTooltip>
                              <HybridTooltipTrigger asChild>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-5 w-5 rounded-full"
                                >
                                  <Info className="h-3.5 w-3.5 text-muted-foreground" />
                                </Button>
                              </HybridTooltipTrigger>
                              <HybridTooltipContent>
                                <p>开启后，通知会自动保存到客户端</p>
                              </HybridTooltipContent>
                            </HybridTooltip>
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
                        <HybridTooltip>
                          <HybridTooltipTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-5 w-5 rounded-full"
                            >
                              <Info className="h-3.5 w-3.5 text-muted-foreground" />
                            </Button>
                          </HybridTooltipTrigger>
                          <HybridTooltipContent>
                            <p>选择需要接收通知的事件类型</p>
                            <br />
                            <p>番剧更新: 当发现新剧集时通知</p>
                            <p>下载完成: 当下载任务完成时通知</p>
                            <p>转移媒体库: 当文件成功转移到媒体库时通知</p>
                            <p>异常通知: 当系统发生错误时通知</p>
                          </HybridTooltipContent>
                        </HybridTooltip>
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
                            updateNoticePoints("subscriptionUpdated", checked)
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
                            updateNoticePoints("downloaded", checked)
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
                            updateNoticePoints("transferred", checked)
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
                            updateNoticePoints("error", checked)
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
                disabled={hasNoticeErrors()}
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
