package utils

import (
	"fmt"
	"html"
	"strings"
)

// BrandedEmailContent holds the pieces of a Luxe transactional email.
type BrandedEmailContent struct {
	Preheader string
	Eyebrow   string
	Title     string
	Intro     string
	// BodyHTML is optional extra HTML (already escaped / trusted fragments only).
	BodyHTML  string
	CTALabel  string
	CTAURL    string
	Footer    string
	LogoURL   string
	BrandName string
}

func trimFrontendURL(frontendURL string) string {
	return strings.TrimRight(strings.TrimSpace(frontendURL), "/")
}

// LogoURLFromFrontend returns an absolute logo URL for email clients.
func LogoURLFromFrontend(frontendURL string) string {
	base := trimFrontendURL(frontendURL)
	if base == "" {
		return ""
	}
	return base + "/apple-touch-icon.png"
}

// BrandedEmailHTML renders a Luxe-themed HTML email (table layout for clients).
func BrandedEmailHTML(c BrandedEmailContent) string {
	brand := strings.TrimSpace(c.BrandName)
	if brand == "" {
		brand = "Luxe"
	}
	safeBrand := html.EscapeString(brand)
	safePreheader := html.EscapeString(c.Preheader)
	safeEyebrow := html.EscapeString(c.Eyebrow)
	if safeEyebrow == "" {
		safeEyebrow = safeBrand
	}
	safeTitle := html.EscapeString(c.Title)
	safeIntro := html.EscapeString(c.Intro)
	safeFooter := html.EscapeString(c.Footer)
	safeCta := html.EscapeString(c.CTALabel)
	safeCtaURL := html.EscapeString(c.CTAURL)

	logoBlock := ""
	if strings.TrimSpace(c.LogoURL) != "" {
		logoBlock = fmt.Sprintf(
			`<img src="%s" width="56" height="56" alt="%s" style="display:block;margin:0 auto 16px;border-radius:14px;border:0;" />`,
			html.EscapeString(c.LogoURL),
			safeBrand,
		)
	}

	ctaBlock := ""
	if strings.TrimSpace(c.CTAURL) != "" && strings.TrimSpace(c.CTALabel) != "" {
		ctaBlock = fmt.Sprintf(
			`<p style="margin:28px 0 0;text-align:center;">
  <a href="%s" style="display:inline-block;background:#0a0a0b;color:#f5f3ef;text-decoration:none;font-weight:600;font-size:14px;letter-spacing:0.04em;padding:14px 28px;border-radius:999px;">%s</a>
</p>
<p style="margin:16px 0 0;font-size:12px;line-height:1.6;color:#6b6760;word-break:break-all;font-family:ui-sans-serif,system-ui,sans-serif;">
  Or paste this link into your browser:<br>
  <a href="%s" style="color:#8b6914;text-decoration:underline;">%s</a>
</p>`,
			safeCtaURL,
			safeCta,
			safeCtaURL,
			safeCtaURL,
		)
	}

	bodyExtra := c.BodyHTML
	if bodyExtra != "" {
		bodyExtra = `<div style="margin:20px 0 0;">` + bodyExtra + `</div>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta name="color-scheme" content="light">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background:#f5f3ef;font-family:Georgia,'Times New Roman',serif;">
  <div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;">%s</div>
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f5f3ef;padding:32px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e7e2d8;border-radius:20px;overflow:hidden;">
        <tr><td style="height:4px;background:linear-gradient(90deg,transparent,#c9a96e,transparent);"></td></tr>
        <tr><td style="padding:36px 32px 8px;text-align:center;">
          %s
          <div style="font-size:11px;letter-spacing:0.28em;text-transform:uppercase;color:#8b6914;">%s</div>
          <h1 style="margin:14px 0 0;font-size:28px;line-height:1.25;color:#0a0a0b;font-weight:600;">%s</h1>
        </td></tr>
        <tr><td style="padding:12px 32px 36px;font-size:16px;line-height:1.7;color:#3d3a36;font-family:ui-sans-serif,system-ui,sans-serif;">
          <p style="margin:0;">%s</p>
          %s
          %s
          <p style="margin:28px 0 0;font-size:13px;line-height:1.6;color:#6b6760;">%s</p>
        </td></tr>
      </table>
      <p style="margin:20px 0 0;font-size:11px;line-height:1.5;color:#9a958c;font-family:ui-sans-serif,system-ui,sans-serif;text-align:center;">
        © %s · This is an automated message — please do not reply.
      </p>
    </td></tr>
  </table>
</body>
</html>`,
		safeTitle,
		safePreheader,
		logoBlock,
		safeEyebrow,
		safeTitle,
		safeIntro,
		bodyExtra,
		ctaBlock,
		safeFooter,
		safeBrand,
	)
}

// PasswordResetEmail builds subject + HTML body for a password reset message.
func PasswordResetEmail(frontendURL, token string) (subject, body string) {
	base := trimFrontendURL(frontendURL)
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", base, token)
	subject = "Reset your Luxe password"
	body = BrandedEmailHTML(BrandedEmailContent{
		Preheader: "Reset your password — this link expires in 1 hour.",
		Eyebrow:   "Account security",
		Title:     "Reset your password",
		Intro:     "We received a request to reset the password for your Luxe account. Click the button below to choose a new password.",
		BodyHTML: `<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#faf8f4;border:1px solid #e7e2d8;border-radius:12px;">
  <tr><td style="padding:14px 16px;font-size:13px;line-height:1.55;color:#6b6760;">
    For your security, this link expires in <strong style="color:#0a0a0b;">1 hour</strong>.
    If you did not request a password reset, you can safely ignore this email — your password will stay the same.
  </td></tr>
</table>`,
		CTALabel:  "Reset password",
		CTAURL:    resetURL,
		Footer:    "Never share this link with anyone. Luxe will never ask you for your password by email.",
		LogoURL:   LogoURLFromFrontend(frontendURL),
		BrandName: "Luxe",
	})
	return subject, body
}

// VerificationEmail builds subject + HTML body for email address verification.
func VerificationEmail(frontendURL, token string) (subject, body string) {
	base := trimFrontendURL(frontendURL)
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", base, token)
	subject = "Verify your Luxe email"
	body = BrandedEmailHTML(BrandedEmailContent{
		Preheader: "Confirm your email to finish setting up your Luxe account.",
		Eyebrow:   "Welcome to Luxe",
		Title:     "Verify your email",
		Intro:     "Thanks for joining Luxe. Please confirm your email address so we can secure your account and keep you updated.",
		BodyHTML: `<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#faf8f4;border:1px solid #e7e2d8;border-radius:12px;">
  <tr><td style="padding:14px 16px;font-size:13px;line-height:1.55;color:#6b6760;">
    This verification link expires in <strong style="color:#0a0a0b;">24 hours</strong>.
  </td></tr>
</table>`,
		CTALabel:  "Verify email",
		CTAURL:    verifyURL,
		Footer:    "If you did not create a Luxe account, you can ignore this message.",
		LogoURL:   LogoURLFromFrontend(frontendURL),
		BrandName: "Luxe",
	})
	return subject, body
}
