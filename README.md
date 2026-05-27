# Darktide Ping Tool

A browser-based latency tester for Warhammer 40,000: Darktide server regions. Measures round-trip time from your machine to each confirmed AWS region used by Fatshark's game servers, helping you identify the lowest-latency region for matchmaking.

**[→ Open the tool: darktidepingtool.pages.dev](https://darktidepingtool.pages.dev/)**

![Warhammer 40k: Darktide](https://img.shields.io/badge/Warhammer%2040k-Darktide-cc0000?style=flat-square)
![Open Source](https://img.shields.io/badge/open%20source-yes-4dff4d?style=flat-square)

## Features

- Tests all 13 confirmed Darktide server regions, grouped by continent
- 7 samples per region for statistical accuracy
- Reports median, p10, p90, jitter, and packet loss per region
- Highlights the lowest-latency region on completion
- Runs entirely client-side — no data is transmitted externally

## Server regions

Region list extracted directly from Darktide's own matchmaker logs (`AppData\Roaming\Fatshark\Darktide\console_logs`), which record the exact AWS region codes the matchmaker resolves latency against before placing a player into a session.

| Region | AWS Region |
|---|---|
| Europe — Ireland | eu-west-1 |
| Europe — London | eu-west-2 |
| Europe — Frankfurt | eu-central-1 |
| North America East — Ohio | us-east-2 |
| North America West — Oregon | us-west-2 |
| North America West — N. California | us-west-1 |
| Canada — Montreal | ca-central-1 |
| South America — São Paulo | sa-east-1 |
| East Asia — Tokyo | ap-northeast-1 |
| East Asia — Seoul | ap-northeast-2 |
| Southeast Asia — Singapore | ap-southeast-1 |
| South Asia — Mumbai | ap-south-1 |
| Oceania — Sydney | ap-southeast-2 |

## How it works

Fatshark hosts Darktide entirely on AWS GameLift, as confirmed by their [AWS re:Invent 2023 session (GAM305)](https://www.youtube.com/watch?v=ZgGYtuNSX4M). This tool measures latency by sending HTTP requests to public AWS DynamoDB endpoints in each region. DynamoDB and GameLift share the same AWS regional backbone, so the measured latency is a reliable proxy for actual in-game connection quality.

The tool cannot ping Fatshark's game servers directly — GameLift instances are not publicly addressable. AWS does provide dedicated UDP ping beacons (`gamelift-ping.{region}.api.aws`, port 7770) specifically for measuring GameLift latency, but browsers cannot send raw UDP packets.

## Usage

Open `index.html` in any modern browser and click **START SCAN**. No installation or dependencies required.

## Files

| File | Description |
|---|---|
| `index.html` | Page markup and layout |
| `app.js` | All JavaScript — scan logic, latency measurement, UI |
| `_headers` | Cloudflare Pages HTTP security headers |

## Security

The project ships a `_headers` file for Cloudflare Pages with a strict Content Security Policy:

- `script-src 'self'` — only `app.js` may execute, no inline scripts
- `connect-src https://*.amazonaws.com` — fetch calls limited to AWS DynamoDB endpoints
- `frame-ancestors 'none'` — page cannot be embedded in an iframe
- `default-src 'none'` — all other resource types blocked by default

## License

Open source.
