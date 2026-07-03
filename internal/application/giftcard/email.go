package giftcard

import (
	"fmt"
	"html"
	"strings"
)

func formatMoney(amount float64, currency string) string {
	if strings.TrimSpace(currency) == "" {
		currency = "USD"
	}
	return fmt.Sprintf("%s %.2f", currency, amount)
}

func giftCardEmailHTML(title, intro, ctaLabel, ctaURL, footer string) string {
	safeTitle := html.EscapeString(title)
	safeIntro := html.EscapeString(intro)
	safeFooter := html.EscapeString(footer)
	safeCta := html.EscapeString(ctaLabel)

	ctaBlock := ""
	if strings.TrimSpace(ctaURL) != "" && strings.TrimSpace(ctaLabel) != "" {
		ctaBlock = fmt.Sprintf(
			`<p style="margin:28px 0 0;text-align:center;">
  <a href="%s" style="display:inline-block;background:#0a0a0b;color:#f5f3ef;text-decoration:none;font-weight:600;font-size:14px;letter-spacing:0.04em;padding:14px 28px;border-radius:999px;">%s</a>
</p>`,
			html.EscapeString(ctaURL),
			safeCta,
		)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="margin:0;padding:0;background:#f5f3ef;font-family:Georgia,'Times New Roman',serif;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f5f3ef;padding:32px 16px;">
    <tr><td align="center">
      <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e7e2d8;border-radius:20px;overflow:hidden;">
        <tr><td style="height:4px;background:linear-gradient(90deg,transparent,#c9a96e,transparent);"></td></tr>
        <tr><td style="padding:32px 32px 8px;text-align:center;">
          <div style="font-size:11px;letter-spacing:0.28em;text-transform:uppercase;color:#8b6914;">Luxe</div>
          <h1 style="margin:12px 0 0;font-size:28px;line-height:1.25;color:#0a0a0b;font-weight:600;">%s</h1>
        </td></tr>
        <tr><td style="padding:8px 32px 32px;font-size:16px;line-height:1.7;color:#3d3a36;font-family:ui-sans-serif,system-ui,sans-serif;">
          <p style="margin:0;">%s</p>
          %s
          <p style="margin:28px 0 0;font-size:13px;line-height:1.6;color:#6b6760;">%s</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, safeTitle, safeIntro, ctaBlock, safeFooter)
}

func giftCardReceivedEmail(senderName, recipientName, amountLabel, code, message, accountURL string) (string, string) {
	intro := fmt.Sprintf(
		"%s sent you a Luxe gift card worth %s. Sign in to your account to view balance and redeem at checkout.",
		senderName,
		amountLabel,
	)
	if strings.TrimSpace(message) != "" {
		intro += fmt.Sprintf(` Personal message: "%s"`, message)
	}
	intro += fmt.Sprintf(" Your gift code is %s.", code)

	subject := fmt.Sprintf("You received a %s Luxe gift card", amountLabel)
	body := giftCardEmailHTML(
		fmt.Sprintf("A gift for you, %s", recipientName),
		intro,
		"View my gift cards",
		accountURL+"?tab=giftCards",
		"Redeem this card on any marketplace order or Luxe Plus — balance applies automatically when you use the code at checkout.",
	)
	return subject, body
}

func giftCardSentEmail(senderName, recipientName, amountLabel, accountURL string) (string, string) {
	intro := fmt.Sprintf(
		"Your %s gift card for %s is ready. We'll notify them by email, and they can claim it from their Luxe account.",
		amountLabel,
		recipientName,
	)
	subject := fmt.Sprintf("Gift card sent to %s", recipientName)
	body := giftCardEmailHTML(
		fmt.Sprintf("Thank you, %s", senderName),
		intro,
		"View sent gift cards",
		accountURL+"?tab=giftCards",
		"You can track sent cards anytime from your account.",
	)
	return subject, body
}

func giftCardTransferredRecipientEmail(fromName, recipientName, amountLabel, accountURL string) (string, string) {
	intro := fmt.Sprintf(
		"%s gifted you a Luxe gift card with %s balance. It's now linked to your account.",
		fromName,
		amountLabel,
	)
	subject := fmt.Sprintf("You received a %s gift card from %s", amountLabel, fromName)
	body := giftCardEmailHTML(
		fmt.Sprintf("Enjoy, %s", recipientName),
		intro,
		"Open gift cards",
		accountURL+"?tab=giftCards",
		"Use your balance at checkout or for Luxe Plus membership.",
	)
	return subject, body
}
