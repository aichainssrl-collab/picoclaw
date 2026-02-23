class CryptoTicker extends HTMLElement {
    constructor() {
        super();
    }

    connectedCallback() {
        this.innerHTML = `
            <div id="crypto-sidebar">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; border-bottom: 2px solid #eee; padding-bottom: 10px;">
                    <h2 style="margin: 0; font-size: 1.2em; font-family: 'Playfair Display', serif;">Crypto Market</h2>
                    <button id="refresh-crypto-btn" style="background: #333; color: #fff; border: none; padding: 5px 10px; cursor: pointer; font-size: 0.7em; text-transform: uppercase; font-weight: bold; border-radius: 4px; letter-spacing: 1px;">Refresh</button>
                </div>
                <div id="crypto-content">Loading...</div>
            </div>
        `;
        
        this.querySelector('#refresh-crypto-btn').addEventListener('click', () => this.refreshCrypto());
        this.fetchCrypto();
    }

    refreshCrypto() {
        const btn = this.querySelector('#refresh-crypto-btn');
        btn.disabled = true;
        btn.textContent = '...';
        
        fetch('/api/crypto/refresh', { method: 'POST' })
            .then(res => {
                if (!res.ok) throw new Error('Update failed');
                return res.json();
            })
            .then(() => {
                this.fetchCrypto();
            })
            .catch(err => {
                console.error(err);
                // alert('Failed to update crypto'); // Optional: don't annoy user if it fails silently
            })
            .finally(() => {
                btn.disabled = false;
                btn.textContent = 'Refresh';
            });
    }

    fetchCrypto() {
        console.log('Fetching crypto...');
        fetch('/api/crypto')
            .then(response => {
                if (!response.ok) throw new Error(`Network response was not ok: ${response.status}`);
                return response.json();
            })
            .then(data => {
                console.log('Crypto data received:', data);
                const cryptoContent = this.querySelector('#crypto-content');
                if (!data || !data.currencies) {
                    cryptoContent.innerHTML = '<p>No data available</p>';
                    return;
                }

                let html = '<table class="crypto-table"><thead><tr><th>Asset</th><th>Price (USD)</th><th>Price (BTC)</th><th>24h % (USD)</th><th>24h % (BTC)</th></tr></thead><tbody>';
        
        data.currencies.forEach(coin => {
            const change24h = coin.change_24h || '0%';
            const change24hBtc = coin.change_24h_btc || '0%';
            
            const isUp = change24h.includes('+');
            const changeClass = isUp ? 'price-up' : 'price-down';
            
            const isUpBtc = change24hBtc.includes('+');
            const changeClassBtc = isUpBtc ? 'price-up' : 'price-down';

            html += `
                <tr>
                    <td><strong>${coin.symbol || '?'}</strong></td>
                    <td>${coin.price || '-'}</td>
                    <td>${coin.price_btc || '-'}</td>
                    <td class="${changeClass}">${change24h}</td>
                    <td class="${changeClassBtc}">${change24hBtc}</td>
                </tr>
            `;
        });
                
                html += '</tbody></table>';
                
                html += `
                    <div class="market-summary">
                        <div><span>Total Cap:</span> <strong>${data.total_market_cap || '-'}</strong></div>
                        <div><span>BTC Dom:</span> <strong>${data.market_dominance_btc || '-'}</strong></div>
                        <div style="margin-top: 10px; font-size: 0.8em; text-align: right;">Updated: ${data.timestamp ? new Date(data.timestamp).toLocaleTimeString() : 'Unknown'}</div>
                    </div>
                `;

                cryptoContent.innerHTML = html;
            })
            .catch(err => {
                console.error('Failed to load crypto:', err);
                this.querySelector('#crypto-content').innerHTML = `
                    <div style="color: red; border: 1px solid red; padding: 10px; background: #ffeeee;">
                        <p>Failed to load market data</p>
                        <small>${err.message}</small>
                    </div>
                `;
            });
    }
}

customElements.define('crypto-ticker', CryptoTicker);
