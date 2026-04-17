package service

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/bytedance/gopkg/util/gopool"
)

func paymentMethodLabel(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "stripe":
		return "Stripe"
	case "epay":
		return "易支付"
	case "creem":
		return "Creem"
	case "waffo":
		return "Waffo"
	case "":
		return "在线支付"
	default:
		return method
	}
}

func paymentActionURL() string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	if base == "" {
		return ""
	}
	return base + "/topup"
}

func buildTopUpVars(topUp *model.TopUp, user *model.User, forAdmin bool) map[string]string {
	completedAt := time.Unix(topUp.CompleteTime, 0).Format("2006-01-02 15:04:05")
	if topUp.CompleteTime == 0 {
		completedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	username := ""
	userId := 0
	if user != nil {
		username = user.Username
		userId = user.Id
	}

	rows := []common.EmailTemplateRow{
		{Label: "订单编号", Value: html.EscapeString(topUp.TradeNo)},
		{Label: "支付方式", Value: html.EscapeString(paymentMethodLabel(topUp.PaymentMethod))},
		{Label: "支付金额", Value: fmt.Sprintf("%.2f", topUp.Money)},
		{Label: "充值额度", Value: fmt.Sprintf("%d", topUp.Amount)},
		{Label: "完成时间", Value: html.EscapeString(completedAt)},
	}
	if strings.TrimSpace(username) != "" {
		rows = append([]common.EmailTemplateRow{
			{Label: "下单用户", Value: html.EscapeString(strings.TrimSpace(username))},
		}, rows...)
	}

	heading := "充值已到账"
	intro := "你的充值已经到账，感谢支持。"
	actionLabel := "查看账户"
	if forAdmin {
		heading = "一笔新的充值"
		intro = fmt.Sprintf("%s 刚完成了一笔充值。", username)
		actionLabel = "查看后台"
	}

	return map[string]string{
		"system_name":           html.EscapeString(systemNameOrDefault()),
		"server_address":        html.EscapeString(strings.TrimRight(system_setting.ServerAddress, "/")),
		"heading":               html.EscapeString(heading),
		"intro":                 html.EscapeString(intro),
		"trade_no":              html.EscapeString(topUp.TradeNo),
		"payment_method":        html.EscapeString(paymentMethodLabel(topUp.PaymentMethod)),
		"money":                 fmt.Sprintf("%.2f", topUp.Money),
		"amount":                fmt.Sprintf("%d", topUp.Amount),
		"username":              html.EscapeString(strings.TrimSpace(username)),
		"user_id":               fmt.Sprintf("%d", userId),
		"completed_at":          html.EscapeString(completedAt),
		"info_table":            common.RenderInfoTableHTML(rows),
		"content_preview":       "",
		"content_preview_block": "",
		"action_url":            html.EscapeString(paymentActionURL()),
		"action_label":          html.EscapeString(actionLabel),
	}
}

// NotifyTopUpSuccess 异步在支付成功后发送邮件（用户 / 管理员 两条通路各自可开关）
func NotifyTopUpSuccess(topUp *model.TopUp) {
	if topUp == nil {
		return
	}
	if !common.PaymentNotifyUserEnabled && !common.PaymentNotifyAdminEnabled {
		return
	}

	gopool.Go(func() {
		user, err := model.GetUserById(topUp.UserId, false)
		if err != nil {
			common.SysLog(fmt.Sprintf("topup notify: failed to load user %d (trade_no=%s): %s", topUp.UserId, topUp.TradeNo, err.Error()))
			return
		}

		// 通知下单用户
		if common.PaymentNotifyUserEnabled {
			userSetting := user.GetSetting()
			userEmail := strings.TrimSpace(user.Email)
			if strings.TrimSpace(userSetting.NotificationEmail) != "" {
				userEmail = strings.TrimSpace(userSetting.NotificationEmail)
			}
			if userEmail != "" {
				vars := buildTopUpVars(topUp, user, false)
				subject, body := RenderEmailByKey(constant.EmailTemplateKeyPaymentSuccessUser, vars)
				if subject != "" && body != "" {
					notify := dto.NewNotify(dto.NotifyTypePaymentSuccess, subject, body, nil)
					if err := NotifyUser(user.Id, userEmail, userSetting, notify); err != nil {
						common.SysLog(fmt.Sprintf("topup notify: failed to notify user %d (trade_no=%s): %s", user.Id, topUp.TradeNo, err.Error()))
					}
				}
			}
		}

		// 通知管理员
		if common.PaymentNotifyAdminEnabled {
			recipients := parseAdminEmails(common.PaymentAdminEmail)
			if len(recipients) == 0 {
				return
			}
			vars := buildTopUpVars(topUp, user, true)
			subject, body := RenderEmailByKey(constant.EmailTemplateKeyPaymentSuccessAdmin, vars)
			if subject == "" || body == "" {
				return
			}
			for _, to := range recipients {
				if err := common.SendEmail(subject, to, body); err != nil {
					common.SysLog(fmt.Sprintf("topup notify: failed to send admin email to %s (trade_no=%s): %s", to, topUp.TradeNo, err.Error()))
				}
			}
		}
	})
}
