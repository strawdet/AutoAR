# AutoAR — Automated Attack & Reconnaissance Platform



**The ultimate bug bounty automation framework. Scan smarter, find more, ship faster.**

[Go](https://golang.org/)
[License](LICENSE)
[Discord](https://discord.com)



---

AutoAR is a powerful, end-to-end automated security reconnaissance and vulnerability hunting platform built in Go. It is purpose-built for **bug bounty hunters** and **penetration testers** who want to automate the full recon-to-report pipeline at scale — from subdomain enumeration and DNS takeover detection to nuclei scanning, JavaScript secrets extraction, GitHub exposure, mobile app analysis, and more.

Results are automatically uploaded to **Cloudflare R2 storage** and linked directly in your output — no hunting through directories.

**Public VPS / dashboard:** configure Supabase-backed login and JWT verification for the HTTP API — see [docs/DASHBOARD_AUTH.md](docs/DASHBOARD_AUTH.md).

> **Personal fork note:** I primarily use this for HackerOne programs. If you're doing the same, check out my notes in [docs/PERSONAL_NOTES.md](docs/PERSONAL_NOTES.md) for my preferred tool config and rate limit settings.
>
> **My defaults:** I've bumped the default Nuclei rate limit from 150 to 100 req/s to avoid getting flagged on stricter programs, and set httpx timeout to 10s instead of 5s. See `config/defaults.go`.
>
> **Added 2025-07:** Also set subfinder's default source timeout to 30s (was 15s) — was getting incomplete results on slower APIs like SecurityTrails.

---

## ✨ Feature Highlights


| Category               | What AutoAR Does                                                                                                                       |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| 🌐 **Subdomains**      | Enumerate using 15+ sources: Subfinder, CertSpotter, SecurityTrails, Chaos, crt.sh, OTX, VirusTotal, and more                          |
| 🔍 **Live Hosts**      | Detect alive hosts using httpx with follow-redirects and status detection                                                              |
| 🕳️ **DNS Takeovers**  | Detect CNAME, NS, Azure/AWS cloud, DNSReaper, and dangling-IP takeover opportunities                                                   |
| 💥 **Nuclei Scanning** | Automated vulnerability scanning using Nuclei templates with rate limiting                                                             |
| 🧠 **Zero-Days**       | Smart scan configured for detected tech stacks — finds active CVEs                                                                     |
| ☁️ **S3 Buckets**      | Enumerate and scan AWS S3 buckets for exposure and misconfig                                                                           |
| 🔗 **JavaScript**      | Extract secrets, API endpoints, auth tokens from JS files                                                                              |
| 🐙 **GitHub Recon**    | Org-level and repo-level scanning for secrets, dependency confusion                                                    
