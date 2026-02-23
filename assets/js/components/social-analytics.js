class SocialAnalytics extends HTMLElement {
    constructor() {
        super();
    }

    connectedCallback() {
        this.innerHTML = `
            <div id="analytics-container" style="display: block;">
                <h2 style="font-family: 'Playfair Display', serif; margin-bottom: 20px;">Social Media Analytics</h2>
                <div id="analytics-content">Loading latest metrics...</div>
            </div>

            <!-- Edit Modal -->
            <div id="editModal" class="modal">
                <div class="modal-content">
                    <span class="close">&times;</span>
                    <h2 style="font-family: 'Playfair Display', serif; margin-top: 0;">Edit Analytics</h2>
                    <form id="editForm">
                        <div class="form-group">
                            <label for="edit-date">Date</label>
                            <input type="date" id="edit-date" name="date" readonly style="background-color: #eee;">
                        </div>
                        <div class="form-group">
                            <label for="edit-fb">Facebook Followers</label>
                            <input type="number" id="edit-fb" name="facebook_followers">
                        </div>
                        <div class="form-group">
                            <label for="edit-fb-posts">Facebook Posts</label>
                            <input type="number" id="edit-fb-posts" name="facebook_posts">
                        </div>
                        <div class="form-group">
                            <label for="edit-ig">Instagram Followers</label>
                            <input type="number" id="edit-ig" name="instagram_followers">
                        </div>
                        <div class="form-group">
                            <label for="edit-ig-posts">Instagram Posts</label>
                            <input type="number" id="edit-ig-posts" name="instagram_posts">
                        </div>
                        <div class="form-group">
                            <label for="edit-li">LinkedIn Followers</label>
                            <input type="number" id="edit-li" name="linkedin_followers">
                        </div>
                        <div class="form-group">
                            <label for="edit-tw">Twitter Followers</label>
                            <input type="number" id="edit-tw" name="twitter_followers">
                        </div>
                        <div class="form-group">
                            <label for="edit-tw-posts">Twitter Posts</label>
                            <input type="number" id="edit-tw-posts" name="twitter_posts">
                        </div>
                        <div class="form-group">
                            <label for="edit-yt">YouTube Subscribers</label>
                            <input type="number" id="edit-yt" name="youtube_subscribers">
                        </div>
                        <div class="form-group">
                            <label for="edit-yt-videos">YouTube Videos</label>
                            <input type="number" id="edit-yt-videos" name="youtube_videos">
                        </div>
                        <button type="submit" class="btn-submit">Save Changes</button>
                    </form>
                </div>
            </div>
        `;

        this.setupModal();
        this.loadAnalytics();
    }

    setupModal() {
        const modal = this.querySelector('#editModal');
        const span = this.querySelector('.close');
        const form = this.querySelector('#editForm');

        span.onclick = () => modal.style.display = "none";
        window.onclick = (event) => {
            if (event.target == modal) modal.style.display = "none";
        }

        form.onsubmit = async (e) => {
            e.preventDefault();
            const data = {
                date: form.date.value,
                facebook_followers: parseInt(form.facebook_followers.value),
                facebook_posts: parseInt(form.facebook_posts.value),
                instagram_followers: parseInt(form.instagram_followers.value),
                instagram_posts: parseInt(form.instagram_posts.value),
                linkedin_followers: parseInt(form.linkedin_followers.value),
                twitter_followers: parseInt(form.twitter_followers.value),
                twitter_posts: parseInt(form.twitter_posts.value),
                youtube_subscribers: parseInt(form.youtube_subscribers.value),
                youtube_videos: parseInt(form.youtube_videos.value)
            };

            try {
                const response = await fetch('/api/analytics', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (!response.ok) throw new Error('Failed to update analytics');
                modal.style.display = "none";
                this.loadAnalytics();
                alert('Analytics updated successfully!');
            } catch (error) {
                alert('Failed to update analytics: ' + error.message);
            }
        }
    }

    openEditModal(day) {
        const modal = this.querySelector('#editModal');
        this.querySelector('#edit-date').value = day.date;
        this.querySelector('#edit-fb').value = day.facebook_followers || 0;
        this.querySelector('#edit-fb-posts').value = day.facebook_posts || 0;
        this.querySelector('#edit-ig').value = day.instagram_followers || 0;
        this.querySelector('#edit-ig-posts').value = day.instagram_posts || 0;
        this.querySelector('#edit-li').value = day.linkedin_followers || 0;
        this.querySelector('#edit-tw').value = day.twitter_followers || 0;
        this.querySelector('#edit-tw-posts').value = day.twitter_posts || 0;
        this.querySelector('#edit-yt').value = day.youtube_subscribers || 0;
        this.querySelector('#edit-yt-videos').value = day.youtube_videos || 0;
        modal.style.display = "block";
    }

    async loadAnalytics() {
        const container = this.querySelector('#analytics-content');
        try {
            const response = await fetch('/api/analytics');
            if (!response.ok) throw new Error('Failed to fetch analytics');
            const data = await response.json();
            
            if (!data || data.length === 0) {
                container.innerHTML = '<p>No analytics data available yet.</p>';
                return;
            }

            data.sort((a, b) => new Date(b.date) - new Date(a.date));

            let html = `
                <table class="analytics-table">
                    <thead>
                        <tr>
                            <th>Date</th>
                            <th>Facebook</th>
                            <th>FB Posts</th>
                            <th>Instagram</th>
                            <th>IG Posts</th>
                            <th>Twitter</th>
                            <th>TW Posts</th>
                            <th>YouTube</th>
                            <th>YT Videos</th>
                            <th>Action</th>
                        </tr>
                    </thead>
                    <tbody>
            `;

            data.forEach((day, index) => {
                html += `
                    <tr>
                        <td>${day.date}</td>
                        <td>${day.facebook_followers || 0}</td>
                        <td>${day.facebook_posts || 0}</td>
                        <td>${day.instagram_followers || 0}</td>
                        <td>${day.instagram_posts || 0}</td>
                        <td>${day.twitter_followers || 0}</td>
                        <td>${day.twitter_posts || 0}</td>
                        <td>${day.youtube_subscribers || 0}</td>
                        <td>${day.youtube_videos || 0}</td>
                        <td>
                            <button class="edit-btn" data-index="${index}" style="padding: 5px 10px; font-size: 0.8em;">Edit</button>
                        </td>
                    </tr>
                `;
            });

            html += `</tbody></table>`;
            container.innerHTML = html;

            this.querySelectorAll('.edit-btn').forEach(btn => {
                btn.onclick = () => this.openEditModal(data[btn.dataset.index]);
            });

        } catch (error) {
            console.error(error);
            container.innerHTML = '<p style="color:red">Error loading analytics data.</p>';
        }
    }
}

customElements.define('social-analytics', SocialAnalytics);
