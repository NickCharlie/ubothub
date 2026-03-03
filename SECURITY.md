# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest  | Yes       |

## Reporting a Vulnerability

If you discover a security vulnerability in UBotHub, please report it responsibly.

### How to Report

1. **Do NOT** create a public GitHub issue for security vulnerabilities.
2. Send an email to **security@ubothub.com** (or open a private security advisory on GitHub).
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours of receiving the report.
- **Assessment**: Within 7 days, we will assess the severity and impact.
- **Fix**: Critical vulnerabilities will be patched within 14 days.
- **Disclosure**: We will coordinate with the reporter on public disclosure timing.

### Scope

The following are in scope for security reports:

- Authentication and authorization bypass
- SQL injection, XSS, CSRF, and SSRF
- Remote code execution
- Sensitive data exposure
- Insecure direct object references
- Server-side request forgery
- Privilege escalation
- Denial of service (application-level)

### Out of Scope

- Social engineering attacks
- Physical security
- Attacks requiring physical access to user devices
- Vulnerabilities in third-party dependencies (report to upstream)

## Security Best Practices

### For Contributors

- Never commit secrets, API keys, or credentials to the repository.
- Use parameterized queries (GORM) to prevent SQL injection.
- Validate all user input at system boundaries.
- Follow the principle of least privilege for database accounts.
- Keep dependencies updated and run govulncheck and pnpm audit regularly.

### For Deployers

- Always use HTTPS in production.
- Set strong JWT secrets (minimum 256-bit).
- Configure Redis with authentication and internal network binding.
- Use non-root users in Docker containers.
- Enable database SSL/TLS connections.
- Regularly rotate access tokens and credentials.

## Acknowledgments

We appreciate responsible security researchers who help keep UBotHub and its users safe.
