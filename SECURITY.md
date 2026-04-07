# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| latest (`master`) | :white_check_mark: |
| older tags | :x: |

## Reporting a Vulnerability

You may report a vulnerability either publicly or privately:

- **Public issue** — [Open an issue](../../issues/new) if you are comfortable disclosing the details publicly. Use the `security` label.
- **Private advisory** — Use GitHub's [private vulnerability reporting](../../security/advisories/new) if the vulnerability is sensitive and you prefer coordinated disclosure before it is publicly visible.

Include in your report:
- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept
- Affected versions
- Any suggested mitigations, if known

You can expect:
- **Acknowledgement** within 2 business days
- **Status update** within 7 days (confirmed, declined, or needs more info)
- **Patch and disclosure** coordinated with you once a fix is ready

When the vulnerability is fixed, a GitHub Security Advisory will be published and a CVE will be requested if warranted.

## Automated Security

- **Dependabot** is enabled and will automatically open pull requests to upgrade vulnerable dependencies (Go modules, npm packages, Docker base images).
- Dependabot PRs that pass CI are **automatically merged**.
- The repository owner is notified by GitHub whenever a security advisory is published or a Dependabot PR is opened.
