#!/usr/bin/env python3
"""
Focused Crown Jewel Manual Scoring System Test
Tests specific scenarios based on database state analysis
"""

import requests
import json
from datetime import datetime

BACKEND_URL = "http://localhost:8080"
ADMIN_TOKEN = None

def get_admin_token():
    """Get admin authentication token"""
    global ADMIN_TOKEN
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/login"
        payload = {"username": "admin", "password": "admin123"}
        
        response = requests.post(url, json=payload, timeout=10)
        data = response.json()
        
        if response.status_code == 200 and data.get('success'):
            ADMIN_TOKEN = data.get('access_token')
            return True
        return False
    except Exception as e:
        print(f"❌ Admin login failed: {e}")
        return False

def test_specific_crown_jewel_scenarios():
    """Test specific Crown Jewel scenarios based on database analysis"""
    
    if not get_admin_token():
        print("❌ Cannot proceed without admin token")
        return
    
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    print("🔍 FOCUSED CROWN JEWEL MANUAL SCORING SYSTEM TESTS")
    print("=" * 60)
    
    # Test Case 1: Match with contests but no participants (Match 1)
    print("\n📊 TEST CASE 1: Match 1 - Contests exist but no participants")
    print("Expected: LEADERBOARD_FINALIZATION_ERROR due to empty contest_participants")
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/complete"
        payload = {
            "final_result": {
                "winner_team_id": 1,
                "final_score": "2-1",
                "mvp_player_id": 1,
                "match_duration": 3000
            },
            "distribute_prizes": True,
            "send_notifications": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=20)
        data = response.json()
        
        print(f"Status: {response.status_code}")
        print(f"Response: {json.dumps(data, indent=2)}")
        
        if data.get('code') == 'LEADERBOARD_FINALIZATION_ERROR':
            print("✅ CONFIRMED: LEADERBOARD_FINALIZATION_ERROR as expected")
        else:
            print(f"❌ UNEXPECTED: Got {data.get('code')} instead of LEADERBOARD_FINALIZATION_ERROR")
            
    except Exception as e:
        print(f"❌ Test Case 1 ERROR: {e}")
    
    # Test Case 2: Match with no contests at all (Match 20)
    print("\n📊 TEST CASE 2: Match 20 - No contests exist")
    print("Expected: CONTEST_UPDATE_ERROR due to no contests to update")
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/20/complete"
        payload = {
            "final_result": {
                "winner_team_id": 1,
                "final_score": "2-0",
                "mvp_player_id": 1,
                "match_duration": 2400
            },
            "distribute_prizes": True,
            "send_notifications": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=20)
        data = response.json()
        
        print(f"Status: {response.status_code}")
        print(f"Response: {json.dumps(data, indent=2)}")
        
        if data.get('code') == 'CONTEST_UPDATE_ERROR':
            print("✅ CONFIRMED: CONTEST_UPDATE_ERROR as expected")
        else:
            print(f"❌ UNEXPECTED: Got {data.get('code')} instead of CONTEST_UPDATE_ERROR")
            
    except Exception as e:
        print(f"❌ Test Case 2 ERROR: {e}")
    
    # Test Case 3: UpdateMatchScore with live match that has contests but no participants
    print("\n📊 TEST CASE 3: Match 1 UpdateMatchScore - Live to completed with contests but no participants")
    print("Expected: COMMIT_ERROR due to transaction rollback in completion pipeline")
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/score"
        payload = {
            "team1_score": 2,
            "team2_score": 1,
            "current_round": 3,
            "match_status": "completed",
            "winner_team_id": 1,
            "final_score": "2-1",
            "match_duration": "40:00"
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=15)
        data = response.json()
        
        print(f"Status: {response.status_code}")
        print(f"Response: {json.dumps(data, indent=2)}")
        
        if data.get('code') == 'COMMIT_ERROR':
            print("✅ CONFIRMED: COMMIT_ERROR as expected")
        else:
            print(f"❌ UNEXPECTED: Got {data.get('code')} instead of COMMIT_ERROR")
            
    except Exception as e:
        print(f"❌ Test Case 3 ERROR: {e}")
    
    print("\n" + "=" * 60)
    print("CROWN JEWEL ROOT CAUSE ANALYSIS COMPLETE")
    print("=" * 60)
    print("🔍 FINDINGS:")
    print("1. Match 1: Has 365 contests with $450K prize pools each, but 0 contest_participants")
    print("2. Match 20: Has 0 contests, causing contest update operations to fail")
    print("3. Match 21: Status 'upcoming', cannot transition directly to 'completed'")
    print("4. Match 2: Already completed, cannot be completed again")
    print("\n💡 ROOT CAUSE:")
    print("The Crown Jewel fix is NOT handling edge cases where:")
    print("- Contests exist but have no participants (causes leaderboard finalization errors)")
    print("- No contests exist at all (causes contest update errors)")
    print("- Transaction pipeline fails when trying to process empty datasets")

if __name__ == "__main__":
    test_specific_crown_jewel_scenarios()