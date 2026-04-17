package common

import (
	"fmt"
	"html"
	"strings"
)

// EmailTemplateRow 表示邮件正文里的一条"键：值"行
type EmailTemplateRow struct {
	Label string
	Value string
}

// EmailTemplateData 通用邮件数据
type EmailTemplateData struct {
	// Heading 顶部大标题，例如 "收到新的工单" / "支付成功"
	Heading string
	// Intro 正文开头说明文字（纯文本，会被 HTML 转义）
	Intro string
	// Rows 键值信息表；Value 已视为安全 HTML（调用方应自行转义）
	Rows []EmailTemplateRow
	// PreviewTitle 可选；若非空，会在 Rows 下方显示 "内容预览" 区块
	PreviewTitle string
	// PreviewHTML 预览内容（已转义过的 HTML）
	PreviewHTML string
	// ActionLabel + ActionURL 可选跳转按钮
	ActionLabel string
	ActionURL   string
}

// RenderEmailTemplate 生成统一风格的 HTML 邮件正文（Apple 风：留白 + 细边 + 黑底按钮）。
//
// 所有系统邮件通知建议使用此模板，保证品牌风格一致。
func RenderEmailTemplate(data EmailTemplateData) string {
	systemName := SystemName
	if systemName == "" {
		systemName = "New API"
	}

	var sb strings.Builder
	sb.WriteString(`<div style="background-color:#fafafa;padding:40px 16px;font-family:-apple-system,BlinkMacSystemFont,'SF Pro Text','Helvetica Neue',Arial,'PingFang SC','Microsoft YaHei',sans-serif;color:#1d1d1f;line-height:1.55;">`)
	sb.WriteString(`<div style="max-width:560px;margin:0 auto;background-color:#ffffff;border:1px solid #ebebeb;border-radius:14px;padding:40px 40px 32px;">`)

	// Heading
	sb.WriteString(fmt.Sprintf(
		`<h1 style="margin:0 0 8px;font-size:24px;font-weight:600;letter-spacing:-0.01em;color:#1d1d1f;">%s</h1>`,
		html.EscapeString(data.Heading),
	))

	// Intro
	if data.Intro != "" {
		sb.WriteString(fmt.Sprintf(
			`<p style="margin:0 0 28px;font-size:14px;color:#8e8e93;">%s</p>`,
			html.EscapeString(data.Intro),
		))
	} else {
		sb.WriteString(`<div style="height:20px;"></div>`)
	}

	// Rows table
	sb.WriteString(RenderInfoTableHTML(data.Rows))

	// Preview block
	if data.PreviewHTML != "" {
		sb.WriteString(RenderPreviewBlockHTML(data.PreviewTitle, data.PreviewHTML))
	}

	// Action button
	if data.ActionURL != "" && data.ActionLabel != "" {
		sb.WriteString(fmt.Sprintf(
			`<div style="margin:32px 0 8px;"><a href="%s" style="display:inline-block;padding:11px 22px;background-color:#1d1d1f;color:#ffffff;text-decoration:none;border-radius:10px;font-size:14px;font-weight:500;letter-spacing:0.01em;">%s</a></div>`,
			html.EscapeString(data.ActionURL),
			html.EscapeString(data.ActionLabel),
		))
	}

	sb.WriteString(`</div>`)

	// Footer: 仅展示站点名作为署名，不再写"系统自动发送"
	sb.WriteString(fmt.Sprintf(
		`<p style="max-width:560px;margin:20px auto 0;text-align:center;color:#a1a1a6;font-size:12px;letter-spacing:0.02em;">— %s —</p>`,
		html.EscapeString(systemName),
	))
	sb.WriteString(`</div>`)

	return sb.String()
}

// EscapeAndBreak 先 HTML 转义再把 \n 换成 <br/>，常用于用户输入的预览
func EscapeAndBreak(s string) string {
	escaped := html.EscapeString(s)
	return strings.ReplaceAll(escaped, "\n", "<br/>")
}

// RenderInfoTableHTML 把一组键值对渲染成邮件里使用的信息表 HTML。Value 视为已转义过的安全 HTML。
//
// 样式：左侧 label 用 #8e8e93，右侧 value 用 #1d1d1f 靠右对齐，每行下细分割线，参考 Apple 账单表格。
func RenderInfoTableHTML(rows []EmailTemplateRow) string {
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`<table style="width:100%;border-collapse:collapse;font-size:14px;margin:0 0 8px;">`)
	for i, r := range rows {
		border := "border-top:1px solid #f0f0f0;"
		if i == 0 {
			border = ""
		}
		sb.WriteString(fmt.Sprintf(
			`<tr><td style="padding:12px 0;color:#8e8e93;vertical-align:top;white-space:nowrap;%s">%s</td><td style="padding:12px 0;color:#1d1d1f;text-align:right;word-break:break-all;%s">%s</td></tr>`,
			border, html.EscapeString(r.Label),
			border, r.Value,
		))
	}
	sb.WriteString(`</table>`)
	return sb.String()
}

// RenderPreviewBlockHTML 将预览正文包装成带标题的预览块；若 previewHTML 为空，返回空串。
func RenderPreviewBlockHTML(title, previewHTML string) string {
	if previewHTML == "" {
		return ""
	}
	if title == "" {
		title = "内容预览"
	}
	return fmt.Sprintf(
		`<p style="margin:24px 0 8px;color:#8e8e93;font-size:13px;">%s</p><div style="padding:16px;background-color:#f5f5f7;border-radius:10px;line-height:1.6;color:#1d1d1f;font-size:14px;">%s</div>`,
		html.EscapeString(title), previewHTML,
	)
}

// RenderPlaceholders 将模板里的 {{key}} 替换为 vars[key] 对应的值。
//
// 说明：
//   - key 支持字母、数字、下划线、点；两侧允许空白，如 {{ user_name }}
//   - 未命中的占位符保持原样，便于管理员在自定义模板里识别错别字
//   - 变量值**不做 HTML 转义**。调用方在准备 vars 时需要自行处理 (EscapeAndBreak / html.EscapeString)，
//     以便模板作者可以故意插入 HTML 片段（例如预览块）
func RenderPlaceholders(tpl string, vars map[string]string) string {
	if tpl == "" || len(vars) == 0 {
		return tpl
	}
	var sb strings.Builder
	sb.Grow(len(tpl))
	i := 0
	for i < len(tpl) {
		// 查找 {{
		idx := strings.Index(tpl[i:], "{{")
		if idx < 0 {
			sb.WriteString(tpl[i:])
			break
		}
		sb.WriteString(tpl[i : i+idx])
		start := i + idx
		end := strings.Index(tpl[start+2:], "}}")
		if end < 0 {
			// 没有闭合，原样保留
			sb.WriteString(tpl[start:])
			break
		}
		rawKey := tpl[start+2 : start+2+end]
		key := strings.TrimSpace(rawKey)
		if val, ok := vars[key]; ok {
			sb.WriteString(val)
		} else {
			// 原样保留整个 {{...}}
			sb.WriteString(tpl[start : start+2+end+2])
		}
		i = start + 2 + end + 2
	}
	return sb.String()
}
