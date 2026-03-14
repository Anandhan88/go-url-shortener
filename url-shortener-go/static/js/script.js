document.addEventListener('DOMContentLoaded', () => {

    // --- Hamburger Menu Toggle ---
    const hamburger = document.querySelector('.hamburger');
    const navMenu = document.querySelector('.nav-menu') || document.querySelector('.nav-links');
    if (hamburger && navMenu) {
        hamburger.addEventListener('click', () => {
            navMenu.classList.toggle('open');
            const icon = hamburger.querySelector('i');
            icon.classList.toggle('fa-bars');
            icon.classList.toggle('fa-xmark');
        });
        // Close menu when a link is clicked
        navMenu.querySelectorAll('a').forEach(link => {
            link.addEventListener('click', () => {
                navMenu.classList.remove('open');
                const icon = hamburger.querySelector('i');
                icon.classList.add('fa-bars');
                icon.classList.remove('fa-xmark');
            });
        });
    }

    // --- Smooth Scrolling for Nav Anchor Links only (Landing Page) ---
    const navLinks = document.querySelectorAll('.nav-menu a[href^="#"], .cta-btn[href^="#"], .nav-btn-outline[href^="#"]');
    navLinks.forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const targetId = this.getAttribute('href');
            if (targetId === '#') return;
            
            const targetElement = document.querySelector(targetId);
            if (targetElement) {
                targetElement.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // --- Index Page Logic ---
    const shortenBtn = document.getElementById('shorten-btn');
    if (shortenBtn) {
        shortenBtn.addEventListener('click', async () => {
            const longUrlInput = document.getElementById('long-url');
            const customKeywordInput = document.getElementById('custom-keyword');
            const resultContainer = document.getElementById('result-container');
            const shortUrlElem = document.getElementById('short-url');
            const errorElem = document.getElementById('error-message');
            const longUrl = longUrlInput.value.trim();
            const customCode = customKeywordInput ? customKeywordInput.value.trim() : '';

            // Reset UI
            resultContainer.classList.add('hidden');
            errorElem.classList.add('hidden');
            errorElem.textContent = '';

            if (!longUrl) {
                showError(errorElem, 'Please enter a valid URL.');
                return;
            }

            try {
                // Change UI state to loading
                const originalBtnText = shortenBtn.innerHTML;
                shortenBtn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Generating...';
                shortenBtn.disabled = true;

                const payload = { long_url: longUrl };
                if (customCode) {
                    payload.custom_code = customCode;
                }

                const response = await fetch('/api/shorten', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                const data = await response.json();

                // Restore Button State
                shortenBtn.innerHTML = originalBtnText;
                shortenBtn.disabled = false;

                if (!response.ok) {
                    throw new Error(data.error || 'Failed to shorten URL');
                }

                const generatedUrl = data.short_url;
                shortUrlElem.href = generatedUrl;
                shortUrlElem.textContent = generatedUrl;
                
                // Set the open link button href
                const openLinkBtn = document.getElementById('open-link-btn');
                if (openLinkBtn) {
                    openLinkBtn.href = generatedUrl;
                }

                resultContainer.classList.remove('hidden');
                
            } catch (error) {
                showError(errorElem, error.message);
            }
        });

        const copyBtn = document.getElementById('copy-btn');
        if (copyBtn) {
            copyBtn.addEventListener('click', () => {
                const shortUrl = document.getElementById('short-url').textContent;
                navigator.clipboard.writeText(shortUrl).then(() => {
                    const icon = copyBtn.querySelector('i');
                    icon.classList.remove('fa-copy', 'fa-regular');
                    icon.classList.add('fa-check', 'fa-solid', 'text-success');
                    
                    setTimeout(() => {
                        icon.classList.remove('fa-check', 'fa-solid', 'text-success');
                        icon.classList.add('fa-copy', 'fa-regular');
                    }, 2000);
                });
            });
        }
    }

    // --- Dashboard Page Logic ---
    const fetchAnalyticsBtn = document.getElementById('fetch-analytics-btn');
    let lineChartInstance = null;
    let pieChartInstance = null;

    if (fetchAnalyticsBtn) {
        fetchAnalyticsBtn.addEventListener('click', async () => {
            let shortCodeInput = document.getElementById('short-code-input').value.trim();
            const analyticsContent = document.getElementById('analytics-content');
            const errorElem = document.getElementById('analytics-error');

            analyticsContent.classList.add('hidden');
            errorElem.classList.add('hidden');

            if (!shortCodeInput) {
                showError(errorElem, 'Please enter your short code or paste the full URL.');
                return;
            }

            // If the user pastes a full URL (e.g., http://localhost:8080/abc123), extract just the last segment safely
            try {
                // If it's a valid URL string, parse it
                if (shortCodeInput.startsWith('http://') || shortCodeInput.startsWith('https://')) {
                    const parsedUrl = new URL(shortCodeInput);
                    // The pathname will be something like "/golangg"
                    shortCodeInput = parsedUrl.pathname.replace('/', '').trim();
                } else if (shortCodeInput.includes('/')) {
                    // Fallback for paths without protocol like localhost:8080/golangg
                    const parts = shortCodeInput.split('/');
                    shortCodeInput = parts[parts.length - 1]; 
                }
            } catch (e) {
                // Ignore parse errors and let the length validation handle it
            }

            // Allow codes between 3 and 20 characters (custom keywords are 3-20, random codes are 6)
            if(shortCodeInput.length < 3 || shortCodeInput.length > 20) {
                 showError(errorElem, 'Invalid code. Must be between 3 and 20 characters.');
                 return;
            }

            try {
                // Loading State
                const origText = fetchAnalyticsBtn.innerHTML;
                fetchAnalyticsBtn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> Loading...';
                fetchAnalyticsBtn.disabled = true;

                const response = await fetch(`/api/analytics/${shortCodeInput}`);
                
                // Restore button
                fetchAnalyticsBtn.innerHTML = origText;
                fetchAnalyticsBtn.disabled = false;

                if (!response.ok) {
                    if(response.status === 404) {
                         throw new Error(`Analytics not found for code: "${shortCodeInput}".`);
                    }
                    throw new Error('Failed to fetch analytics.');
                }

                const data = await response.json();
                
                // Update 4-Card Stats
                document.getElementById('total-clicks').textContent = data.total_clicks || 0;
                document.getElementById('unique-visitors').textContent = data.unique_visitors || 0;
                document.getElementById('avg-clicks').textContent = data.avg_clicks_per_day || 0;
                document.getElementById('days-active').textContent = data.days_active || 0;

                // Render Charts
                renderCharts(data);

                // Populate Recent Clicks Table
                populateTable(data.recent_clicks || []);

                // Show Content
                analyticsContent.classList.remove('hidden');

            } catch (error) {
                showError(errorElem, error.message);
                // Restore button on error as well
                fetchAnalyticsBtn.innerHTML = 'View Analytics';
                fetchAnalyticsBtn.disabled = false;
            }
        });
    }

    function showError(element, message) {
        element.innerHTML = `<i class="fa-solid fa-circle-exclamation"></i> ${message}`;
        element.classList.remove('hidden');
    }

    function populateTable(clicks) {
        const tbody = document.getElementById('recent-clicks-body');
        tbody.innerHTML = ''; // clear exiting

        if (clicks.length === 0) {
            tbody.innerHTML = '<tr><td colspan="3" class="text-center text-muted">No recent clicks found.</td></tr>';
            return;
        }

        clicks.forEach(click => {
            const tr = document.createElement('tr');
            
            // Format Time
            const dateObj = new Date(click.clicked_at);
            const timeStr = dateObj.toLocaleString([], {
                year: 'numeric', month: 'short', day: 'numeric', 
                hour: '2-digit', minute: '2-digit'
            });

            tr.innerHTML = `
                <td>${timeStr}</td>
                <td><span style="font-family: monospace; color: var(--text-muted);">${click.ip_address}</span></td>
                <td>
                    <span class="device-badge">
                        ${getDeviceIcon(click.device_type)} ${click.device_type}
                    </span>
                </td>
            `;
            tbody.appendChild(tr);
        });
    }

    function getDeviceIcon(type) {
        const t = type.toLowerCase();
        if(t.includes('mobile')) return '<i class="fa-solid fa-mobile-screen"></i>';
        if(t.includes('tablet')) return '<i class="fa-solid fa-tablet-screen-button"></i>';
        return '<i class="fa-solid fa-desktop"></i>';
    }

    function renderCharts(data) {
        const dates = Object.keys(data.clicks_over_time || {}).sort();
        const clicks = dates.map(date => data.clicks_over_time[date]);

        // Destroy previous instances
        if (lineChartInstance) lineChartInstance.destroy();
        if (pieChartInstance) pieChartInstance.destroy();

        // Line Chart
        if (dates.length > 0) {
            const ctxLine = document.getElementById('lineChart').getContext('2d');
            lineChartInstance = new Chart(ctxLine, {
                type: 'line',
                data: {
                    labels: dates,
                    datasets: [{
                        label: 'Clicks',
                        data: clicks,
                        borderColor: '#1f2937',        // Primary SaaS Color
                        backgroundColor: 'rgba(31, 41, 55, 0.05)',
                        borderWidth: 2,
                        pointBackgroundColor: '#1f2937',
                        pointBorderColor: '#ffffff',
                        pointBorderWidth: 2,
                        pointRadius: 4,
                        tension: 0.3,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: { display: false },
                        tooltip: { backgroundColor: '#111827', padding: 10, cornerRadius: 6 }
                    },
                    scales: {
                        x: { grid: { display: false } },
                        y: { 
                            beginAtZero: true, 
                            ticks: { precision: 0 },
                            border: { dash: [4, 4] },
                            grid: { color: '#e5e7eb' }
                        }
                    }
                }
            });
        }

        // Pie Chart
        const devices = Object.keys(data.device_types || {});
        const deviceData = devices.map(d => data.device_types[d]);
        // Professional SaaS Palette
        const pieColors = ['#1f2937', '#6b7280', '#e5e7eb'];

        if (devices.length > 0) {
            const ctxPie = document.getElementById('pieChart').getContext('2d');
            pieChartInstance = new Chart(ctxPie, {
                type: 'doughnut',
                data: {
                    labels: devices,
                    datasets: [{
                        data: deviceData,
                        backgroundColor: pieColors,
                        borderWidth: 0,
                        hoverOffset: 4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    cutout: '65%',
                    plugins: {
                        legend: { position: 'bottom', labels: { usePointStyle: true, padding: 20 } },
                        tooltip: { backgroundColor: '#111827', padding: 10, cornerRadius: 6 }
                    }
                }
            });
        }
    }
});
