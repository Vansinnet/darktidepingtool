
        // AWS GameLift regions confirmed used by Fatshark's Darktide matchmaker.
        // Source: Darktide's own console logs (AppData\Roaming\Fatshark\Darktide\console_logs),
        // which log the exact region codes the matchmaker resolves latency against.
        // Fatshark's use of AWS GameLift is confirmed by the official AWS case study:
        // https://aws.amazon.com/solutions/case-studies/fatshark-case-study/
        // Latency is measured by pinging AWS DynamoDB public endpoints in each region —
        // a standard proxy technique, since DynamoDB and GameLift share the same AWS
        // regional backbone infrastructure.
        const regionGroups = [
            {
                label: "Europe",
                servers: [
                    { name: "Ireland (eu-west-1)",      endpoint: "https://dynamodb.eu-west-1.amazonaws.com" },
                    { name: "London (eu-west-2)",       endpoint: "https://dynamodb.eu-west-2.amazonaws.com" },
                    { name: "Frankfurt (eu-central-1)", endpoint: "https://dynamodb.eu-central-1.amazonaws.com" },
                ]
            },
            {
                label: "North America",
                servers: [
                    { name: "Ohio (us-east-2)",          endpoint: "https://dynamodb.us-east-2.amazonaws.com" },
                    { name: "Oregon (us-west-2)",        endpoint: "https://dynamodb.us-west-2.amazonaws.com" },
                    { name: "N. California (us-west-1)", endpoint: "https://dynamodb.us-west-1.amazonaws.com" },
                    { name: "Montreal (ca-central-1)",   endpoint: "https://dynamodb.ca-central-1.amazonaws.com" },
                ]
            },
            {
                label: "South America",
                servers: [
                    { name: "São Paulo (sa-east-1)", endpoint: "https://dynamodb.sa-east-1.amazonaws.com" },
                ]
            },
            {
                label: "Asia Pacific",
                servers: [
                    { name: "Tokyo (ap-northeast-1)",     endpoint: "https://dynamodb.ap-northeast-1.amazonaws.com" },
                    { name: "Seoul (ap-northeast-2)",     endpoint: "https://dynamodb.ap-northeast-2.amazonaws.com" },
                    { name: "Singapore (ap-southeast-1)", endpoint: "https://dynamodb.ap-southeast-1.amazonaws.com" },
                    { name: "Mumbai (ap-south-1)",        endpoint: "https://dynamodb.ap-south-1.amazonaws.com" },
                ]
            },
            {
                label: "Oceania",
                servers: [
                    { name: "Sydney (ap-southeast-2)", endpoint: "https://dynamodb.ap-southeast-2.amazonaws.com" },
                ]
            },
        ];

        const resultsDiv = document.getElementById('results');
        const startBtn = document.getElementById('startBtn');
        const clearBtn = document.getElementById('clearBtn');
        const statsPanel = document.getElementById('statsPanel');

        // Improved function to measure latency with multiple samples
        async function measureLatency(url) {
            const samples = [];
            const attempts = 7; // 7 samples balances accuracy with scan time across 13 regions
            let failedAttempts = 0;

            for (let i = 0; i < attempts; i++) {
                // AbortController lets us properly cancel the fetch on timeout,
                // instead of just ignoring the still-running request.
                const controller = new AbortController();
                const timeoutId = setTimeout(() => controller.abort(), 5000);

                try {
                    const start = performance.now();

                    // Use fetch with no-cors to measure HTTP request time.
                    // Add cache-busting parameter to force fresh requests.
                    await fetch(url + "?_=" + Math.random() + "&t=" + Date.now(), {
                        mode: 'no-cors',
                        cache: 'no-store',
                        redirect: 'follow',
                        signal: controller.signal
                    });

                    clearTimeout(timeoutId);
                    const elapsed = Math.round(performance.now() - start);
                    samples.push(Math.max(elapsed, 1)); // Ensure minimum 1ms

                } catch (error) {
                    clearTimeout(timeoutId);
                    // Track failed/timed-out attempts for packet loss estimation.
                    // Do NOT push 5000ms into samples — that would skew median/percentiles.
                    // Failures are accounted for separately via packetLoss.
                    failedAttempts++;
                }

                // Small delay between samples to avoid overwhelming the server
                if (i < attempts - 1) {
                    await new Promise(resolve => setTimeout(resolve, 100));
                }
            }

            // If every single attempt failed, return a sentinel value
            if (samples.length === 0) {
                return {
                    ping: 9999, median: 9999, average: 9999,
                    min: 9999, max: 9999, p10: 9999, p90: 9999,
                    stdDev: 0, jitter: 0,
                    samples: [],
                    packetLoss: 100,
                    failedAttempts: attempts,
                    totalAttempts: attempts
                };
            }

            // Sort once — used for median, percentiles, min, max
            samples.sort((a, b) => a - b);

            // Median: more robust than average against outliers
            const median = samples[Math.floor(samples.length / 2)];

            // Average kept for reference / stddev calculation
            const average = Math.round(samples.reduce((a, b) => a + b) / samples.length);

            const min = samples[0];
            const max = samples[samples.length - 1];

            // p10 and p90 give a clear picture of variability without extreme outliers
            const p10 = samples[Math.max(0, Math.floor(samples.length * 0.1))];
            const p90 = samples[Math.min(samples.length - 1, Math.floor(samples.length * 0.9))];

            // Standard deviation (jitter) — calculated against average
            const variance = samples.reduce((sum, val) => sum + Math.pow(val - average, 2), 0) / samples.length;
            const stdDev = Math.round(Math.sqrt(variance));

            // Packet loss: failed attempts / total attempts
            const packetLoss = Math.round((failedAttempts / attempts) * 100);

            return {
                ping: median,   // primary metric is now median
                samples: samples,
                median: median,
                average: average,
                min: min,
                max: max,
                p10: p10,
                p90: p90,
                stdDev: stdDev,
                jitter: stdDev,
                packetLoss: packetLoss,
                failedAttempts: failedAttempts,
                totalAttempts: attempts
            };
        }

        function getPingClass(ping) {
            if (ping <= 80) return 'ping-good';
            if (ping <= 120) return 'ping-ok';
            if (ping <= 180) return 'ping-poor';
            return 'ping-bad';
        }

        function getStabilityScore(stdDev, packetLoss) {
            // Calculate a stability score (0-100) based on jitter and packet loss
            // Lower jitter and packet loss = higher score

            // Normalize jitter (assume max useful jitter is 200ms)
            const jitterScore = Math.max(0, 100 - (stdDev / 2));

            // Normalize packet loss (0% loss = 100, any loss reduces score)
            const lossScore = Math.max(0, 100 - (packetLoss * 5));

            // Average the two components
            const finalScore = Math.round((jitterScore + lossScore) / 2);

            return Math.max(0, Math.min(100, finalScore));
        }

        function getJitterClass(stdDev) {
            if (stdDev <= 15) return 'jitter-excellent';
            if (stdDev <= 30) return 'jitter-good';
            if (stdDev <= 60) return 'jitter-poor';
            return 'jitter-bad';
        }

        function getStabilityDescription(score) {
            if (score >= 90) return 'EXCELLENT';
            if (score >= 75) return 'STABLE';
            if (score >= 60) return 'MODERATE';
            if (score >= 40) return 'UNSTABLE';
            return 'POOR';
        }

        function updateStats() {
            const rows = document.querySelectorAll('.row');
            const pings = [];

            rows.forEach(row => {
                const pingText = row.querySelector('.ping-value').innerText;
                const pingValue = parseInt(pingText);
                if (!isNaN(pingValue)) {
                    pings.push(pingValue);
                }
            });

            if (pings.length === 0) {
                statsPanel.style.display = 'none';
                return;
            }

            statsPanel.style.display = 'grid';

            const avgPing = Math.round(pings.reduce((a, b) => a + b) / pings.length);
            const bestPing = Math.min(...pings);
            const worstPing = Math.max(...pings);

            document.getElementById('avgPing').innerText = avgPing + ' ms';
            document.getElementById('bestServer').innerText = bestPing + ' ms';
            document.getElementById('worstServer').innerText = worstPing + ' ms';
        }

        clearBtn.addEventListener('click', () => {
            resultsDiv.innerHTML = '';
            statsPanel.style.display = 'none';
            startBtn.disabled = false;
            startBtn.innerText = '▶ START SCAN';
            const recommendation = document.querySelector('.recommendation');
            if (recommendation) {
                recommendation.remove();
            }
        });

        startBtn.addEventListener('click', async () => {
            startBtn.disabled = true;
            startBtn.innerText = '⏳ SCANNING...';
            clearBtn.disabled = true;
            resultsDiv.innerHTML = '<div class="loading">Scanning regions...</div>';

            let bestRegion = null;
            let bestPing = Infinity;
            let globalIndex = 0;
            const totalServers = regionGroups.reduce((sum, g) => sum + g.servers.length, 0);

            for (const group of regionGroups) {
                // Insert group header
                const header = document.createElement('div');
                header.className = 'region-group-header';
                header.innerText = '█ ' + group.label;
                resultsDiv.appendChild(header);

                for (let si = 0; si < group.servers.length; si++) {
                    const { name, endpoint } = group.servers[si];

                    // Create UI element
                    const row = document.createElement('div');
                    row.className = 'row';

                    const regionInfo = document.createElement('div');
                    regionInfo.className = 'region-info';

                    const nameSpan = document.createElement('div');
                    nameSpan.className = 'region-name';
                    nameSpan.innerText = name;

                    regionInfo.appendChild(nameSpan);

                    const pingDetails = document.createElement('div');
                    pingDetails.className = 'ping-details';

                    const pingSpan = document.createElement('div');
                    pingSpan.className = 'ping-value ping-testing';
                    pingSpan.innerText = 'Testing...';

                    const statsSpan = document.createElement('div');
                    statsSpan.className = 'ping-stats';
                    statsSpan.innerText = '';

                    pingDetails.appendChild(pingSpan);
                    pingDetails.appendChild(statsSpan);

                    row.appendChild(regionInfo);
                    row.appendChild(pingDetails);
                    resultsDiv.appendChild(row);

                    // Perform measurement
                    try {
                        const result = await measureLatency(endpoint);
                        pingSpan.innerText = result.median + ' ms';
                        pingSpan.className = 'ping-value ' + getPingClass(result.median);
                        statsSpan.innerText = '(median) p10: ' + result.p10 + 'ms  p90: ' + result.p90 + 'ms';

                        // Calculate stability score
                        const stabilityScore = getStabilityScore(result.jitter, result.packetLoss);
                        const jitterClass = getJitterClass(result.jitter);
                        const stabilityDesc = getStabilityDescription(stabilityScore);

                        // Create stability metric display
                        const stabilityMetric = document.createElement('div');
                        stabilityMetric.className = 'stability-metric';

                        const jitterIndicator = document.createElement('span');
                        jitterIndicator.className = 'jitter-indicator ' + jitterClass;
                        jitterIndicator.title = 'Jitter: ' + result.jitter + 'ms';

                        const stabilityLabel = document.createElement('span');
                        stabilityLabel.innerText = stabilityDesc + ' (' + stabilityScore + ')';

                        const stabilityBar = document.createElement('div');
                        stabilityBar.className = 'stability-bar';
                        const fill = document.createElement('div');
                        fill.className = 'stability-fill';
                        fill.style.width = stabilityScore + '%';
                        stabilityBar.appendChild(fill);

                        stabilityMetric.appendChild(jitterIndicator);
                        stabilityMetric.appendChild(stabilityLabel);
                        stabilityMetric.appendChild(stabilityBar);
                        pingDetails.appendChild(stabilityMetric);

                        // Create packet loss display
                        const packetLossDiv = document.createElement('div');
                        packetLossDiv.className = 'packet-loss';
                        if (result.packetLoss > 0) {
                            packetLossDiv.classList.add(result.packetLoss >= 30 ? 'critical' : 'warning');
                            packetLossDiv.innerText = '⚠ PACKET LOSS: ' + result.packetLoss + '%';
                        } else {
                            packetLossDiv.style.color = '#4d9d4d';
                            packetLossDiv.innerText = '✓ NO PACKET LOSS';
                        }
                        pingDetails.appendChild(packetLossDiv);

                        // Track best server (by median)
                        if (result.median < bestPing) {
                            bestPing = result.median;
                            bestRegion = group.label + ' — ' + name;
                        }
                    } catch (error) {
                        pingSpan.innerText = 'Error';
                        pingSpan.className = 'ping-value ping-bad';
                        statsSpan.innerText = '';
                    }

                    globalIndex++;
                    // Small delay between servers to avoid network congestion
                    if (globalIndex < totalServers) {
                        await new Promise(resolve => setTimeout(resolve, 200));
                    }
                }
            }

            // Remove loading message
            const loadingDiv = document.querySelector('.loading');
            if (loadingDiv) {
                loadingDiv.remove();
            }

            // Show recommendation
            const recommendationDiv = document.createElement('div');
            recommendationDiv.className = 'recommendation';
            recommendationDiv.innerHTML = `
                <div class="recommendation-title">LOWEST LATENCY REGION</div>
                <div class="recommendation-server">${bestRegion}</div>
                <div style="font-size: 0.85rem; color: #666; margin-top: 8px;">LATENCY: ${bestPing} ms</div>
            `;

            // Insert recommendation at the top
            resultsDiv.insertBefore(recommendationDiv, resultsDiv.firstChild);

            updateStats();

            startBtn.disabled = false;
            startBtn.innerText = '▶ SCAN AGAIN';
            clearBtn.disabled = false;
        });
    
