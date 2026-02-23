class NewsFeed extends HTMLElement {
    constructor() {
        super();
        this.allNewsData = [];
        this.currentCategory = 'All';
    }

    connectedCallback() {
        this.innerHTML = `
            <div id="news-container">
                <div id="news-header" style="position: relative; margin-bottom: 20px;">
                    <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
                        <div style="flex: 1;"></div>
                        <h2>The Daily Chronicle</h2>
                        <div style="flex: 1; text-align: right;">
                            <button id="refresh-news-btn" style="background: #333; color: #fff; border: none; padding: 6px 12px; cursor: pointer; font-size: 0.8em; text-transform: uppercase; border-radius: 4px; font-weight: bold; letter-spacing: 1px;">Refresh</button>
                        </div>
                    </div>
                    <div id="news-last-updated" style="font-family: 'Source Sans Pro', sans-serif; font-size: 0.9em; color: #666; margin-top: 10px; text-transform: uppercase; letter-spacing: 1px;"></div>
                </div>
                
                <div id="news-tabs">
                    <button class="news-tab active" data-category="All">All News</button>
                    <button class="news-tab" data-category="World">World</button>
                    <button class="news-tab" data-category="Tech">Tech & AI</button>
                    <button class="news-tab" data-category="Business">Business</button>
                </div>

                <div id="news-grid">Loading top stories...</div>
            </div>
        `;

        this.setupEventListeners();
        this.fetchNews();
    }

    setupEventListeners() {
        this.querySelector('#refresh-news-btn').addEventListener('click', () => this.refreshNews());
        
        this.querySelectorAll('.news-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                this.querySelectorAll('.news-tab').forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                this.currentCategory = tab.getAttribute('data-category');
                this.renderNews(this.currentCategory);
            });
        });
    }

    categorizeItem(item) {
        const src = (item.source || '').toLowerCase();
        const title = (item.title || '').toLowerCase();
        const summary = (item.summary || '').toLowerCase();
        const text = title + " " + summary;
        const cat = (item._category || '').toLowerCase();

        if (cat === 'tech') return 'Tech';
        if (cat === 'business') return 'Business';
        if (src.includes('nyt') || src.includes('bbc') || src.includes('world') || text.includes('president') || text.includes('war') || text.includes('gaza') || text.includes('ukraine')) return 'World';
        if (src.includes('ai4business') || src.includes('agenda digitale') || src.includes('zerouno') || src.includes('corcom') || text.includes('ai ') || text.includes('tech') || text.includes('digital') || text.includes('cyber') || text.includes('cloud') || text.includes('software')) return 'Tech';
        if (text.includes('market') || text.includes('economy') || text.includes('stock') || text.includes('trade')) return 'Business';
        return 'General';
    }

    renderNews(category) {
        const newsGrid = this.querySelector('#news-grid');
        newsGrid.innerHTML = '';
        
        const filtered = category === 'All' 
            ? this.allNewsData 
            : this.allNewsData.filter(item => item._category === category || (category === 'Business' && item._category === 'General'));

        if (filtered.length === 0) {
            newsGrid.innerHTML = '<p style="grid-column: 1/-1; text-align: center; font-style: italic; color: #666;">No news in this section today.</p>';
            return;
        }

        filtered.forEach(item => {
            const card = document.createElement('div');
            card.className = 'news-card';
            let dateStr = "";
            if (item.published) {
                try {
                    const date = new Date(item.published);
                    dateStr = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
                } catch(e) { dateStr = item.published; }
            }

            card.innerHTML = `
                <div class="news-meta">
                    <span class="news-source" style="color: ${this.getCategoryColor(item._category)}">${item._category} | ${item.source || 'Source'}</span>
                    <span class="news-date">${dateStr}</span>
                </div>
                <div class="news-title"><a href="${item.link}" target="_blank">${item.title}</a></div>
                <div class="news-snippet">${item.summary || ''}</div>
                <div class="news-footer" style="display: flex; gap: 10px; margin-top: 15px;">
                    <a href="${item.link}" target="_blank" class="read-more" style="flex: 1; text-align: center; background: #6c757d; color: white; padding: 8px; text-decoration: none; border-radius: 4px; font-size: 14px;">📖 Leggi Notizia</a>
                    <a href="#" class="create-post" data-link="${item.link}" data-title="${(item.title || '').replace(/"/g, '&quot;')}" data-summary="${(item.summary || '').replace(/"/g, '&quot;')}" style="flex: 1; text-align: center; background: #0077b5; color: white; padding: 8px; text-decoration: none; border-radius: 4px; font-size: 14px;">📝 Crea Post</a>
                </div>
            `;
            newsGrid.appendChild(card);
            
            // Add click handler for the new button
            const createBtn = card.querySelector('.create-post');
            createBtn.onclick = (e) => {
                e.preventDefault();
                const postData = {
                    url: createBtn.dataset.link,
                    text: `${createBtn.dataset.title}\n\n${createBtn.dataset.summary}`
                };
                sessionStorage.setItem('cms_post_data', JSON.stringify(postData));
                window.location.href = '/linkedin/cms';
            };
        });
    }

    getCategoryColor(cat) {
        switch(cat) {
            case 'World': return '#c62828';
            case 'Tech': return '#1976d2';
            case 'Business': return '#2e7d32';
            default: return '#555';
        }
    }

    async fetchNews() {
        try {
            console.log('Fetching news...');
            const response = await fetch('/api/news');
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            const data = await response.json();
            console.log('News data received:', data);
            
            if (!Array.isArray(data)) {
                throw new Error('Data is not an array');
            }

            this.allNewsData = data.map(item => ({
                ...item,
                _category: this.categorizeItem(item)
            })).sort((a, b) => {
                const dateA = a.published ? new Date(a.published) : new Date(0);
                const dateB = b.published ? new Date(b.published) : new Date(0);
                return dateB - dateA; 
            });

            // Update timestamp
            const now = new Date();
            const timeString = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            const dateString = now.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' });
            this.querySelector('#news-last-updated').textContent = `Last updated: ${dateString} at ${timeString}`;

            this.renderNews(this.currentCategory);
        } catch (error) {
            console.error('Error fetching news:', error);
            this.querySelector('#news-grid').innerHTML = `
                <div style="grid-column: 1/-1; text-align: center; color: red; padding: 20px; border: 1px solid red; background: #ffeeee;">
                    <h3>Error Loading News</h3>
                    <p>${error.message}</p>
                    <p>Check console for details.</p>
                </div>
            `;
        }
    }

    async refreshNews() {
        try {
            this.querySelector('#refresh-news-btn').textContent = 'Refreshing...';
            this.querySelector('#refresh-news-btn').disabled = true;
            
            await fetch('/api/news/refresh', { method: 'POST' });
            
            // Wait a moment for backend update
            setTimeout(() => {
                this.fetchNews();
                this.querySelector('#refresh-news-btn').textContent = 'Refresh';
                this.querySelector('#refresh-news-btn').disabled = false;
            }, 2000);
        } catch (error) {
            console.error('Error refreshing news:', error);
            this.querySelector('#refresh-news-btn').textContent = 'Error';
            setTimeout(() => {
                this.querySelector('#refresh-news-btn').textContent = 'Refresh';
                this.querySelector('#refresh-news-btn').disabled = false;
            }, 2000);
        }
    }
}

customElements.define('news-feed', NewsFeed);
