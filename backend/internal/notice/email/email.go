package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

// notifier 实现notice.Notifier接口，通过邮件发送通知
type notifier struct {
	mu  sync.Mutex
	cfg Config
}

// Config 邮件通知配置
type Config struct {
	Host     string   `mapstructure:"host" json:"host"`
	Username string   `mapstructure:"username" json:"username"`
	Password string   `mapstructure:"password" json:"password"`
	From     string   `mapstructure:"from" json:"from"`
	To       []string `mapstructure:"to" json:"to"`
	SSL      bool     `mapstructure:"ssl" json:"ssl"`
}

// NewEmailNotifier 创建新的邮件通知器实例
func NewEmailNotifier(cfg Config) notice.Notifier {
	return &notifier{
		cfg: cfg,
	}
}

// 获取SMTP认证和服务器地址
func (n *notifier) getAuth() (smtp.Auth, string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	auth := smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.Host)
	port := "25"
	if n.cfg.SSL {
		port = "587"
	}
	addr := fmt.Sprintf("%s:%s", n.cfg.Host, port)

	return auth, addr
}

// 发送邮件
func (n *notifier) sendEmail(subject, htmlBody string) error {
	auth, addr := n.getAuth()

	// 设置邮件头
	headers := make(map[string]string)
	headers["From"] = n.cfg.From
	headers["To"] = strings.Join(n.cfg.To, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + htmlBody

	// 发送邮件
	return smtp.SendMail(
		addr,
		auth,
		n.cfg.From,
		n.cfg.To,
		[]byte(message),
	)
}

// NoticeSubscriptionUpdated 实现Notifier接口，通知订阅更新状态
func (n *notifier) NoticeSubscriptionUpdated(ctx context.Context, req notice.NoticeSubscriptionUpdatedReq) error {
	subject := fmt.Sprintf("番剧订阅更新：%s", req.BangumiName)

	var statusMsg, statusColor string
	if req.Error != nil {
		statusMsg = "下载失败"
		statusColor = "#FF3B30" // 红色
	} else {
		statusMsg = "开始下载..."
		statusColor = "#34C759" // 绿色
	}

	// 构建错误信息HTML
	var errorHtml string
	if req.Error != nil {
		errorHtml = fmt.Sprintf(`<div style="margin-top: 20px; color: #FF3B30; background-color: #FFEBE9; padding: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
			<strong>错误详情:</strong> %s
		</div>`, req.Error.Error())
	} else {
		errorHtml = ""
	}

	// 构建HTML邮件内容
	htmlBody := fmt.Sprintf(`
	<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 30px; border-radius: 10px; box-shadow: 0 4px 10px rgba(0,0,0,0.1); background-color: #ffffff;">
		<div style="text-align: center; margin-bottom: 25px;">
			<h1 style="color: #0A84FF; margin: 0; font-size: 24px; font-weight: 600;">番剧订阅更新通知</h1>
			<div style="width: 50px; height: 3px; background-color: #0A84FF; margin: 15px auto;"></div>
		</div>
		
		<table style="width: 100%%; border-collapse: collapse; margin-bottom: 25px;">
			<tr>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; width: 30%%;"><strong style="color: #555;">番剧:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; color: #333;">%s</td>
			</tr>
			<tr style="background-color: #f9f9f9;">
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">季度:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; color: #333;">第%d季</td>
			</tr>
			<tr>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">字幕组:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; color: #333;">%s</td>
			</tr>
			<tr style="background-color: #f9f9f9;">
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">RSS订阅项:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><code style="background-color: #f0f0f0; padding: 3px 6px; border-radius: 4px; font-size: 13px; color: #333;">%s</code></td>
			</tr>
		</table>
		
		<div style="margin: 25px 0; text-align: center; background-color: %s; color: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1);">
			<strong style="font-size: 16px;">%s</strong>
		</div>
		
		%s
		
		<div style="margin-top: 35px; padding-top: 20px; border-top: 1px solid #eaeaea; font-size: 13px; color: #999; text-align: center;">
			<p>此邮件由 BangumiBuddy 系统自动发送，请勿回复</p>
			<p style="margin-top: 5px; font-size: 12px;">© %d BangumiBuddy</p>
		</div>
	</div>
	`,
		req.BangumiName, req.Season, req.ReleaseGroup, req.RSSGUID,
		statusColor, statusMsg, errorHtml, time.Now().Year())

	// 如果有海报，添加海报图片链接
	if req.Poster != "" {
		htmlBody = strings.Replace(htmlBody,
			`<div style="text-align: center; margin-bottom: 25px;">`,
			fmt.Sprintf(`<div style="text-align: center; margin-bottom: 25px;">
				<img src="%s" alt="番剧海报" style="max-width: 180px; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.2); margin-bottom: 20px;">`, req.Poster), 1)
	}

	return n.sendEmail(subject, htmlBody)
}

// NoticeDownloaded 实现Notifier接口，通知资源下载状态
func (n *notifier) NoticeDownloaded(ctx context.Context, req notice.NoticeDownloadedReq) error {
	var subject string
	title := ""
	if req.RSSGUID != "" {
		title = "番剧"
	} else {
		title = "磁力任务"
	}
	if req.Failed {
		subject = fmt.Sprintf("%s下载失败：%s", title, req.TorrentName)
	} else {
		subject = fmt.Sprintf("%s下载完成：%s", title, req.TorrentName)
	}

	var statusMsg, statusColor, detailsHtml string
	if req.Failed {
		statusMsg = "下载失败"
		statusColor = "#FF3B30" // 红色
		detailsHtml = fmt.Sprintf(`
		<div style="margin-top: 20px; color: #FF3B30; background-color: #FFEBE9; padding: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
			<strong>错误详情:</strong> %s
		</div>`, req.FailDetail)
	} else {
		statusMsg = "下载成功"
		statusColor = "#34C759" // 绿色

		sizeInfo := utils.FormatFileSize(req.Size)
		costInfo := utils.FormatDuration(req.Cost)
		speedInfo := utils.CalculateAverageSpeed(req.Size, req.Cost)

		detailsHtml = fmt.Sprintf(`
		<div style="margin-top: 25px; background-color: #F5F9FF; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.05);">
			<h3 style="margin-top: 0; margin-bottom: 15px; color: #0A84FF; font-size: 16px;">下载详情</h3>
			<div style="display: flex; justify-content: space-between; margin-bottom: 15px;">
				<div style="flex: 1; text-align: center; padding: 10px; border-right: 1px solid #E0E7FF;">
					<div style="font-size: 12px; color: #666; margin-bottom: 5px;">文件大小</div>
					<div style="font-size: 16px; color: #333; font-weight: 600;">%s</div>
				</div>
				<div style="flex: 1; text-align: center; padding: 10px; border-right: 1px solid #E0E7FF;">
					<div style="font-size: 12px; color: #666; margin-bottom: 5px;">耗时</div>
					<div style="font-size: 16px; color: #333; font-weight: 600;">%s</div>
				</div>
				<div style="flex: 1; text-align: center; padding: 10px;">
					<div style="font-size: 12px; color: #666; margin-bottom: 5px;">平均速度</div>
					<div style="font-size: 16px; color: #333; font-weight: 600;">%s</div>
				</div>
			</div>
		</div>`, sizeInfo, costInfo, speedInfo)
	}

	// 如果 RSSGUID 为空，则不展示对应信息
	rssRow := ""
	if strings.TrimSpace(req.RSSGUID) != "" {
		rssRow = fmt.Sprintf(`
			<tr style="background-color: #f9f9f9;">
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">RSS订阅项:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; word-break: break-all;"><code style="background-color: #f0f0f0; padding: 3px 6px; border-radius: 4px; font-size: 13px; color: #333;">%s</code></td>
			</tr>
		`, req.RSSGUID)
	}

	// 构建HTML邮件内容
	htmlBody := fmt.Sprintf(`
	<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 30px; border-radius: 10px; box-shadow: 0 4px 10px rgba(0,0,0,0.1); background-color: #ffffff;">
		<div style="text-align: center; margin-bottom: 25px;">
			<h1 style="color: #0A84FF; margin: 0; font-size: 24px; font-weight: 600;">番剧下载通知</h1>
			<div style="width: 50px; height: 3px; background-color: #0A84FF; margin: 15px auto;"></div>
		</div>
		
		<table style="width: 100%%; border-collapse: collapse; margin-bottom: 25px;">
			<tr>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; width: 30%%;"><strong style="color: #555;">文件名:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; word-break: break-all;"><code style="background-color: #f0f0f0; padding: 3px 6px; border-radius: 4px; font-size: 13px; color: #333;">%s</code></td>
			</tr>
			%s
		</table>
		
		<div style="margin: 25px 0; text-align: center; background-color: %s; color: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1);">
			<strong style="font-size: 16px;">%s</strong>
		</div>
		
		%s
		
		<div style="margin-top: 35px; padding-top: 20px; border-top: 1px solid #eaeaea; font-size: 13px; color: #999; text-align: center;">
			<p>此邮件由 BangumiBuddy 系统自动发送，请勿回复</p>
			<p style="margin-top: 5px; font-size: 12px;">© %d BangumiBuddy</p>
		</div>
	</div>
	`, req.TorrentName, rssRow, statusColor, statusMsg, detailsHtml, time.Now().Year())

	return n.sendEmail(subject, htmlBody)
}

// NoticeSubscriptionTransferred 实现Notifier接口，通知资源转移状态
func (n *notifier) NoticeSubscriptionTransferred(ctx context.Context, req notice.NoticeSubscriptionTransferredReq) error {
	var subject string
	if req.Error == nil {
		subject = fmt.Sprintf("番剧转移成功：%s", req.BangumiName)
	} else {
		subject = fmt.Sprintf("番剧转移失败：%s", req.BangumiName)
	}

	var statusMsg, statusColor, detailsHtml string
	if req.Error != nil {
		statusMsg = "转移媒体库失败"
		statusColor = "#FF3B30" // 红色

		errorMsg := "未知错误"
		if req.Error != nil {
			errorMsg = req.Error.Error()
		}

		detailsHtml = fmt.Sprintf(`
		<div style="margin-top: 20px; color: #FF3B30; background-color: #FFEBE9; padding: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
			<strong>错误详情:</strong> %s
		</div>`, errorMsg)
	} else {
		statusMsg = "转移媒体库成功"
		statusColor = "#34C759" // 绿色

		detailsHtml = fmt.Sprintf(`
		<div style="margin-top: 20px; background-color: #F5F9FF; padding: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.05);">
			<h3 style="margin-top: 0; margin-bottom: 10px; color: #0A84FF; font-size: 16px;">番剧媒体库信息</h3>
			<code style="display: block; background-color: #f0f0f0; padding: 10px; border-radius: 5px; font-size: 13px; color: #333; word-break: break-all; overflow-wrap: break-word;">%s</code>
		</div>`, req.MediaFilePath)
	}

	// 构建HTML邮件内容
	htmlBody := fmt.Sprintf(`
	<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 30px; border-radius: 10px; box-shadow: 0 4px 10px rgba(0,0,0,0.1); background-color: #ffffff;">
		<div style="text-align: center; margin-bottom: 25px;">
			<h1 style="color: #0A84FF; margin: 0; font-size: 24px; font-weight: 600;">番剧转移媒体库通知</h1>
			<div style="width: 50px; height: 3px; background-color: #0A84FF; margin: 15px auto;"></div>
		</div>
		
		<table style="width: 100%%; border-collapse: collapse; margin-bottom: 25px;">
			<tr>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; width: 30%%;"><strong style="color: #555;">番剧:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; color: #333;">%s</td>
			</tr>
			<tr style="background-color: #f9f9f9;">
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">季度:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; color: #333;">第%d季</td>
			</tr>
			<tr>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">字幕组:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; color: #333;">%s</td>
			</tr>
			<tr style="background-color: #f9f9f9;">
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">文件名:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; word-break: break-all;"><code style="background-color: #f0f0f0; padding: 3px 6px; border-radius: 4px; font-size: 13px; color: #333;">%s</code></td>
			</tr>
			<tr>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea;"><strong style="color: #555;">RSS订阅项:</strong></td>
				<td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; word-break: break-all;"><code style="background-color: #f0f0f0; padding: 3px 6px; border-radius: 4px; font-size: 13px; color: #333;">%s</code></td>
			</tr>
		</table>
		
		<div style="margin: 25px 0; text-align: center; background-color: %s; color: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1);">
			<strong style="font-size: 16px;">%s</strong>
		</div>
		
		%s
		
		<div style="margin-top: 35px; padding-top: 20px; border-top: 1px solid #eaeaea; font-size: 13px; color: #999; text-align: center;">
			<p>此邮件由 BangumiBuddy 系统自动发送，请勿回复</p>
			<p style="margin-top: 5px; font-size: 12px;">© %d BangumiBuddy</p>
		</div>
	</div>
	`, req.BangumiName, req.Season, req.ReleaseGroup, req.FileName, req.RSSGUID, statusColor, statusMsg, detailsHtml, time.Now().Year())

	// 如果有海报，添加海报图片链接
	if req.Poster != "" {
		htmlBody = strings.Replace(htmlBody,
			`<div style="text-align: center; margin-bottom: 25px;">`,
			fmt.Sprintf(`<div style="text-align: center; margin-bottom: 25px;">
				<img src="%s" alt="番剧海报" style="max-width: 180px; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.2); margin-bottom: 20px;">`, req.Poster), 1)
	}

	return n.sendEmail(subject, htmlBody)
}

// NoticeTaskTransferred 实现Notifier接口，通知任务转移状态
func (n *notifier) NoticeTaskTransferred(ctx context.Context, req notice.NoticeTaskTransferredReq) error {
	var subject string
	var statusMsg, statusColor, detailsHtml string

	// 判断转移状态
	successCount := len(req.MediaFilePaths)
	hasFailures := req.Error != nil

	if successCount == 0 && hasFailures {
		// 全部转移失败
		subject = fmt.Sprintf("磁力任务转移失败：%s", req.BangumiName)
		statusMsg = "全部转移失败"
		statusColor = "#FF3B30" // 红色
		detailsHtml = fmt.Sprintf(`
        <div style="margin-top: 20px; color: #FF3B30; background-color: #FFEBE9; padding: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
            <strong>错误详情:</strong> %s
        </div>`, req.Error.Error())
	} else if successCount > 0 && !hasFailures {
		// 全部转移成功
		subject = fmt.Sprintf("磁力任务转移成功：%s", req.BangumiName)
		statusMsg = fmt.Sprintf("全部转移成功 (%d个文件)", successCount)
		statusColor = "#34C759" // 绿色

		// 构建成功文件列表
		fileListHtml := `
        <div style="margin-top: 25px; background-color: #F5F9FF; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.05);">
            <h3 style="margin-top: 0; margin-bottom: 15px; color: #0A84FF; font-size: 16px;">转移成功的文件</h3>`

		for originFile, mediaFile := range req.MediaFilePaths {
			fileListHtml += fmt.Sprintf(`
            <div style="margin-bottom: 10px; padding: 10px; border-left: 3px solid #34C759; background-color: #ffffff;">
                <div style="font-size: 13px; color: #666; margin-bottom: 3px;">原始文件:</div>
                <code style="display: block; background-color: #f0f0f0; padding: 5px; border-radius: 4px; font-size: 12px; color: #333; word-break: break-all; margin-bottom: 8px;">%s</code>
                <div style="font-size: 13px; color: #666; margin-bottom: 3px;">媒体库路径:</div>
                <code style="display: block; background-color: #f0f0f0; padding: 5px; border-radius: 4px; font-size: 12px; color: #333; word-break: break-all;">%s</code>
            </div>`, originFile, mediaFile)
		}
		fileListHtml += `</div>`
		detailsHtml = fileListHtml
	} else {
		// 部分成功，部分失败
		subject = fmt.Sprintf("磁力任务转移部分成功：%s", req.BangumiName)
		statusMsg = fmt.Sprintf("%d个文件转移成功", successCount)
		statusColor = "#FF9500" // 橙色

		detailsHtml = ""

		// 添加成功文件列表
		if successCount > 0 {
			detailsHtml += `
        <div style="margin-top: 25px; background-color: #F5F9FF; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.05);">
            <h3 style="margin-top: 0; margin-bottom: 15px; color: #34C759; font-size: 16px;">转移成功的文件</h3>`

			for originFile, mediaFile := range req.MediaFilePaths {
				detailsHtml += fmt.Sprintf(`
            <div style="margin-bottom: 10px; padding: 10px; border-left: 3px solid #34C759; background-color: #ffffff;">
                <div style="font-size: 13px; color: #666; margin-bottom: 3px;">原始文件:</div>
                <code style="display: block; background-color: #f0f0f0; padding: 5px; border-radius: 4px; font-size: 12px; color: #333; word-break: break-all; margin-bottom: 8px;">%s</code>
                <div style="font-size: 13px; color: #666; margin-bottom: 3px;">媒体库路径:</div>
                <code style="display: block; background-color: #f0f0f0; padding: 5px; border-radius: 4px; font-size: 12px; color: #333; word-break: break-all;">%s</code>
            </div>`, originFile, mediaFile)
			}
			detailsHtml += `</div>`
		}

		// 添加失败信息
		if hasFailures {
			detailsHtml += fmt.Sprintf(`
        <div style="margin-top: 20px; color: #FF3B30; background-color: #FFEBE9; padding: 15px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
            <strong>转移失败详情:</strong> %s
        </div>`, req.Error.Error())
		}
	}

	htmlBody := fmt.Sprintf(`
    <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 30px; border-radius: 10px; box-shadow: 0 4px 10px rgba(0,0,0,0.1); background-color: #ffffff;">
        <div style="text-align: center; margin-bottom: 25px;">
            <h1 style="color: #0A84FF; margin: 0; font-size: 24px; font-weight: 600;">磁力任务转移通知</h1>
            <div style="width: 50px; height: 3px; background-color: #0A84FF; margin: 15px auto;"></div>
        </div>

        <table style="width: 100%%; border-collapse: collapse; margin-bottom: 25px;">
            <tr>
                <td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; width: 30%%;"><strong style="color: #555;">种子名:</strong></td>
                <td style="padding: 12px 15px; border-bottom: 1px solid #eaeaea; word-break: break-all;"><code style="background-color: #f0f0f0; padding: 3px 6px; border-radius: 4px; font-size: 13px; color: #333;">%s</code></td>
            </tr>
        </table>

        <div style="margin: 25px 0; text-align: center; background-color: %s; color: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 5px rgba(0,0,0,0.1);">
            <strong style="font-size: 16px;">%s</strong>
        </div>

        %s

        <div style="margin-top: 35px; padding-top: 20px; border-top: 1px solid #eaeaea; font-size: 13px; color: #999; text-align: center;">
            <p>此邮件由 BangumiBuddy 系统自动发送，请勿回复</p>
            <p style="margin-top: 5px; font-size: 12px;">© %d BangumiBuddy</p>
        </div>
    </div>
    `, req.TorrentName, statusColor, statusMsg, detailsHtml, time.Now().Year())

	return n.sendEmail(subject, htmlBody)
}
