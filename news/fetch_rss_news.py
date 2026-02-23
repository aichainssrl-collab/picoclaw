#!/usr/bin/env python3
import json
import datetime
import os

# Dummy script to simulate fetching news
# In a real scenario, this would fetch RSS feeds

news_items = [
    {
        "title": "New AI Models Released",
        "link": "https://example.com/ai-news",
        "published": datetime.datetime.now().isoformat(),
        "source": "TechCrunch",
        "summary": "Major tech companies release new AI models today.",
        "_category": "Tech"
    },
    {
        "title": "Global Markets Update",
        "link": "https://example.com/markets",
        "published": datetime.datetime.now().isoformat(),
        "source": "Bloomberg",
        "summary": "Stock markets show positive trends.",
        "_category": "Business"
    }
]

file_path = os.path.join(os.path.dirname(__file__), "news.json")
with open(file_path, "w") as f:
    json.dump(news_items, f, indent=4)

print("News updated successfully")
