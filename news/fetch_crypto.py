#!/usr/bin/env python3
import json
import datetime
import os
import random

def get_random_change():
    change = random.uniform(-5.0, 5.0)
    sign = "+" if change >= 0 else ""
    return f"{sign}{change:.2f}%"

def get_random_price(base):
    variation = random.uniform(-0.02, 0.02)
    return base * (1 + variation)

crypto_data = {
    "total_market_cap": "$2.1T",
    "market_dominance_btc": "52.4%",
    "timestamp": datetime.datetime.now().isoformat(),
    "currencies": [
        {
            "symbol": "BTC",
            "price": f"${get_random_price(65000):,.2f}",
            "price_btc": "1.00",
            "change_24h": get_random_change(),
            "change_24h_btc": "0.00%"
        },
        {
            "symbol": "ETH",
            "price": f"${get_random_price(3500):,.2f}",
            "price_btc": "0.054",
            "change_24h": get_random_change(),
            "change_24h_btc": get_random_change()
        },
        {
            "symbol": "SOL",
            "price": f"${get_random_price(145):,.2f}",
            "price_btc": "0.0022",
            "change_24h": get_random_change(),
            "change_24h_btc": get_random_change()
        }
    ]
}

file_path = os.path.join(os.path.dirname(__file__), "crypto.json")
with open(file_path, "w") as f:
    json.dump(crypto_data, f, indent=4)

print("Crypto data updated")
