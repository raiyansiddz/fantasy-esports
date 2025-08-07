#!/usr/bin/env python3
import requests

# Test a simple endpoint
url = "http://localhost:8001/api/v1/achievements"
try:
    response = requests.get(url, timeout=10)
    print(f"Status: {response.status_code}")
    print(f"Response: {response.text}")
    print(f"Response object: {response}")
except Exception as e:
    print(f"Error: {e}")
    print(f"Error type: {type(e)}")