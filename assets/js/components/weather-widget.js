class WeatherWidget extends HTMLElement {
    constructor() {
        super();
        this.currentCity = 'Milan';
        this.attachShadow({ mode: 'open' });
    }

    connectedCallback() {
        this.render();
        this.fetchWeather(this.currentCity);
    }

    render() {
        this.shadowRoot.innerHTML = `
            <style>
                :host {
                    display: block;
                    font-family: 'Source Sans Pro', sans-serif;
                    margin-bottom: 20px;
                }
                .weather-container {
                    background: #fff;
                    border: 1px solid #ddd;
                    padding: 10px 15px;
                    border-radius: 8px;
                    display: flex;
                    align-items: center;
                    justify-content: space-between;
                    box-shadow: 0 2px 4px rgba(0,0,0,0.05);
                }
                .weather-info {
                    display: flex;
                    align-items: center;
                }
                .weather-icon {
                    font-size: 1.5em;
                    margin-right: 10px;
                }
                .weather-display {
                    font-weight: bold;
                    color: #333;
                }
                .controls {
                    display: flex;
                    gap: 5px;
                }
                input {
                    padding: 5px 8px;
                    border: 1px solid #ccc;
                    border-radius: 4px;
                    font-size: 0.9em;
                    width: 80px;
                }
                button {
                    background: #333;
                    color: white;
                    border: none;
                    padding: 5px 10px;
                    border-radius: 4px;
                    cursor: pointer;
                    font-size: 0.9em;
                }
                button:hover {
                    background: #555;
                }
            </style>
            <div class="weather-container">
                <div class="weather-info">
                    <span class="weather-icon">🌤️</span>
                    <div class="weather-display" id="weather-display">Loading weather for Milan...</div>
                </div>
                <div class="controls">
                    <input type="text" id="city-input" placeholder="City...">
                    <button id="update-btn">Update</button>
                </div>
            </div>
        `;

        this.setupEventListeners();
    }

    setupEventListeners() {
        const btn = this.shadowRoot.getElementById('update-btn');
        const input = this.shadowRoot.getElementById('city-input');

        const updateWeather = () => {
            const city = input.value.trim();
            if (city) {
                this.currentCity = city;
                this.fetchWeather(city);
            }
        };

        btn.addEventListener('click', updateWeather);
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') updateWeather();
        });
    }

    async fetchWeather(city) {
        const display = this.shadowRoot.getElementById('weather-display');
        display.textContent = `Loading weather for ${city}...`;
        display.style.opacity = '0.5';

        try {
                    const encodedCity = encodeURIComponent(city);
                    const controller = new AbortController();
                    const timeoutId = setTimeout(() => controller.abort(), 25000);
                    
                    const response = await fetch(`/api/weather?city=${encodedCity}`, {
                        signal: controller.signal
                    });
                    
                    clearTimeout(timeoutId);

                    if (!response.ok) throw new Error('Failed to fetch weather');
                    const text = await response.text();
                    
                    if (text.includes("Network error") || text.includes("Error fetching")) {
                         throw new Error(text);
                    }

                    display.textContent = text.trim();
                    display.style.opacity = '1';
                } catch (error) {
                    console.error('Error fetching weather:', error);
                    if (error.name === 'AbortError') {
                        display.textContent = `Request timed out for ${city}. Try again.`;
                    } else {
                        display.textContent = `Could not load weather for ${city}.`;
                    }
                    display.style.opacity = '1';
                }
    }
}

customElements.define('weather-widget', WeatherWidget);
