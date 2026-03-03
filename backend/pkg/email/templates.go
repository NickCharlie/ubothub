package email

import "fmt"

// VerificationEmail generates an email verification HTML body.
func VerificationEmail(username, verifyURL string) (subject string, body string) {
	subject = "Verify Your UBotHub Account"
	body = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#333">
<h2 style="color:#1a1a2e">Welcome to UBotHub!</h2>
<p>Hi %s,</p>
<p>Thank you for registering. Please verify your email address by clicking the button below:</p>
<p style="text-align:center;margin:30px 0">
  <a href="%s" style="background:#4361ee;color:#fff;text-decoration:none;padding:12px 30px;border-radius:8px;font-weight:600;display:inline-block">Verify Email</a>
</p>
<p>Or copy and paste this link into your browser:</p>
<p style="word-break:break-all;color:#666;font-size:14px">%s</p>
<p style="color:#999;font-size:13px;margin-top:30px">This link expires in 24 hours. If you did not create an account, please ignore this email.</p>
<hr style="border:none;border-top:1px solid #eee;margin:20px 0">
<p style="color:#999;font-size:12px">UBotHub — Open Bot Avatar Platform</p>
</body>
</html>`, username, verifyURL, verifyURL)
	return
}

// PasswordResetEmail generates a password reset HTML body.
func PasswordResetEmail(username, resetURL string) (subject string, body string) {
	subject = "Reset Your UBotHub Password"
	body = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#333">
<h2 style="color:#1a1a2e">Password Reset Request</h2>
<p>Hi %s,</p>
<p>We received a request to reset your password. Click the button below to set a new password:</p>
<p style="text-align:center;margin:30px 0">
  <a href="%s" style="background:#e63946;color:#fff;text-decoration:none;padding:12px 30px;border-radius:8px;font-weight:600;display:inline-block">Reset Password</a>
</p>
<p>Or copy and paste this link into your browser:</p>
<p style="word-break:break-all;color:#666;font-size:14px">%s</p>
<p style="color:#999;font-size:13px;margin-top:30px">This link expires in 1 hour. If you did not request a password reset, please ignore this email.</p>
<hr style="border:none;border-top:1px solid #eee;margin:20px 0">
<p style="color:#999;font-size:12px">UBotHub — Open Bot Avatar Platform</p>
</body>
</html>`, username, resetURL, resetURL)
	return
}
