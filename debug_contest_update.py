#!/usr/bin/env python3
"""
Debug script to check contest update issues
"""

import requests
import json

BACKEND_URL = "http://localhost:8001"
ADMIN_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbl9pZCI6MSwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOiJzdXBlcl9hZG1pbiIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJpc3MiOiJmYW50YXN5LWVzcG9ydHMtYWRtaW4iLCJleHAiOjE3NTQxNDc0MTMsImlhdCI6MTc1NDEzMzAxM30.vAi2iwAazv2gqX2QqL_D96MtWJImLJcn1GNRbhe6IH8"

def test_simple_complete_match():
    """Test a simple match completion to debug the issue"""
    print("Testing simple match completion...")
    
    url = f"{BACKEND_URL}/api/v1/admin/matches/30/complete"
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    payload = {
        "final_result": {
            "winner_team_id": 1,
            "final_score": "2-0",
            "mvp_player_id": 1,
            "match_duration": 2400
        },
        "distribute_prizes": False,  # Disable prize distribution to isolate the issue
        "send_notifications": False  # Disable notifications to isolate the issue
    }
    
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
        print(f"Status Code: {response.status_code}")
        print(f"Response: {response.text}")
        
        if response.status_code != 200:
            print("❌ Match completion failed")
        else:
            print("✅ Match completion succeeded")
            
    except Exception as e:
        print(f"❌ Error: {str(e)}")

if __name__ == "__main__":
    test_simple_complete_match()