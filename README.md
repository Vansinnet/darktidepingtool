# Darktide Ping Tool

A browser-based latency tester for Warhammer 40,000: Darktide server regions. Measures round-trip time from your machine to each confirmed AWS region used by Fatshark's game servers, helping you identify the lowest-latency region for matchmaking.

**[→ Open the tool: darktidepingtool.pages.dev](https://darktidepingtool.pages.dev/)**

![Warhammer 40k: Darktide](https://img.shields.io/badge/Warhammer%2040k-Darktide-cc0000?style=flat-square)
![Open Source](https://img.shields.io/badge/open%20source-yes-4dff4d?style=flat-square)

## Features

- Tests all 8 verified Darktide server regions
- 10 samples per region for statistical accuracy
- Reports median, p10, p90, jitter, and packet loss per region
- Highlights the lowest-latency region on completion
- Runs entirely client-side — no data is transmitted externally

## Server regions

| Region | AWS Region |
|---|---|
| North America West | us-west-2 (Oregon) |
| North America East | us-east-2 (Ohio) |
| Canada | ca-central-1 (Montreal) |
| South America | sa-east-1 (São Paulo) |
| Europe | eu-central-1 (Frankfurt) |
| Southeast Asia | ap-southeast-1 (Singapore) |
| East Asia | ap-northeast-1 (Tokyo) |
| Oceania | ap-southeast-2 (Sydney) |

Regions were verified using community-documented matchmaker logs and player discussions on the Fatshark forums and Steam. Regions without independent confirmation (Mumbai, Seoul, Hong Kong, South Africa, N. California) are excluded.

## How it works

Fatshark hosts Darktide entirely on AWS GameLift, as confirmed by an [official AWS case study](https://aws.amazon.com/solutions/case-studies/fatshark-case-study/). This tool measures latency by sending HTTP requests to public AWS DynamoDB endpoints in each region. DynamoDB and GameLift share the same AWS regional backbone, so the measured latency is a reliable proxy for actual in-game connection quality.

## Usage

Open `index.html` in any modern browser and click **START SCAN**. No installation or dependencies required.

## License

Open source.
