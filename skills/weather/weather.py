#!/usr/bin/env python3
import sys
import random

def get_weather(city):
    conditions = ["Sunny", "Cloudy", "Rainy", "Partly Cloudy", "Stormy"]
    temp = random.randint(10, 35)
    condition = random.choice(conditions)
    return f"{condition}, {temp}°C in {city}"

if __name__ == "__main__":
    city = "Milan"
    if len(sys.argv) > 1:
        city = sys.argv[1]
    
    print(get_weather(city))
