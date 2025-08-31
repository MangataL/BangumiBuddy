module github.com/MangataL/BangumiBuddy

go 1.24.1

toolchain go1.24.8

// CGO 依赖说明：
// 字体子集化功能需要 HarfBuzz 库支持
// 开发环境安装：
//   macOS: brew install harfbuzz
//   Ubuntu/Debian: apt-get install libharfbuzz-dev
//   CentOS/RHEL: yum install harfbuzz-devel

require (
	github.com/Tnze/go.num/v2 v2.0.0-20191006170829-cb483d4c9152
	github.com/creasty/defaults v1.8.0
	github.com/cyruzin/golang-tmdb v1.6.3
	github.com/gin-gonic/gin v1.10.0
	github.com/glebarez/sqlite v1.11.0
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/mholt/archives v0.1.5
	github.com/mmcdole/gofeed v1.3.0
	github.com/nssteinbrenner/anitogo v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/samber/lo v1.49.1
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.10.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.37.0
	golang.org/x/image v0.32.0
	golang.org/x/sync v0.17.0
	gorm.io/gorm v1.25.7
	k8s.io/apimachinery v0.32.3
)

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/STARRY-S/zip v0.2.3 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/mikelolasagasti/xz v1.0.1 // indirect
	github.com/minio/minlz v1.0.1 // indirect
	github.com/nwaples/rardecode/v2 v2.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/sorairolake/lzip-go v0.3.8 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
)

require (
	github.com/PuerkitoBio/goquery v1.8.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/autobrr/go-qbittorrent v1.14.0
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/glebarez/go-sqlite v1.21.2
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mmcdole/goxpp v1.1.1-0.20240225020742-a0c311522b23 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/exp v0.0.0-20241108190413-2d47ceb2692f // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/sqlite v1.5.7
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1
)
