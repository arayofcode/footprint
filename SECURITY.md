# Security Policy

## Supported Versions

Only the latest version of Footprint is supported for security updates.

## Reporting a Vulnerability

If you discover a security vulnerability within this project, please open an issue or contact me directly if the issue is sensitive. I aim to acknowledge all reports within 48 hours.

---

## Threat Model

Footprint is a GitHub Action and local CLI tool that interacts with the GitHub API. Understanding the security boundaries is critical for safe usage.

### 1. GitHub Token Scope
*   **Threat**: Excessive permissions in the provided `GITHUB_TOKEN`.
*   **Risk**: If a vulnerability is found in the tool or its dependencies, an attacker could abuse the token to perform unauthorized actions on your behalf.
*   **Mitigation**: Always use the minimum required scope. Footprint primarily needs `pull-requests:read`, `issues:read`, and `contents:read`. If pushing artifacts back to the repository, `contents:write` is required.

### 2. Data Privacy
*   **Threat**: Accidental exposure of private data.
*   **Risk**: Footprint currently only indexes **public** contributions. HoIver, running it in a private repository context without care could potentially leak repository names or metadata into the generated artifacts (`summary.md`, `report.json`, `card.svg`).
*   **Mitigation**: Review the generated artifacts before making them public. Footprint is designed to showcase open-source impact; use it intentionally.

### 3. Dependency Supply Chain
*   **Threat**: Compromised third-party Go modules.
*   **Risk**: Malicious code execution during build or runtime.
*   **Mitigation**: I periodically update dependencies and use Go's checksum database (`go.sum`) to ensure integrity.

### 4. Injection Attacks
*   **Threat**: Maliciously crafted GitHub content (PR titles, comments).
*   **Risk**: XSS or command injection if the tool does not properly sanitize data before rendering into SVG or Markdown.
*   **Mitigation**: Footprint uses standard Go templating and `html.EscapeString` in the card renderer to sanitize all user-generated content displayed in the SVG.
