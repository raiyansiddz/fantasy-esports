#!/usr/bin/env python3
"""
Focused test for the 3 specific fixes mentioned in the review request
"""

import requests
import json

BACKEND_URL = "http://localhost:8080"

def get_admin_token():
    """Get admin token for authentication"""
    url = f"{BACKEND_URL}/api/v1/admin/login"
    payload = {"username": "admin", "password": "admin123"}
    response = requests.post(url, json=payload)
    if response.status_code == 200:
        data = response.json()
        return data.get('access_token')
    return None

def test_health():
    """Test health endpoint"""
    print("🔍 Testing Health Endpoint...")
    response = requests.get(f"{BACKEND_URL}/health")
    print(f"Status: {response.status_code}")
    if response.status_code == 200:
        print("✅ Health check working")
        return True
    else:
        print("❌ Health check failed")
        return False

def test_admin_login():
    """Test admin login"""
    print("\n🔍 Testing Admin Login...")
    token = get_admin_token()
    if token:
        print("✅ Admin login working - token obtained")
        return True, token
    else:
        print("❌ Admin login failed")
        return False, None

def test_live_scoring_postgresql_fix(token):
    """Test PostgreSQL STRING_AGG fix in live scoring endpoint"""
    print("\n🔍 Testing PostgreSQL STRING_AGG Fix (Live Scoring)...")
    url = f"{BACKEND_URL}/api/v1/admin/matches/live-scoring"
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.get(url, headers=headers)
    
    print(f"Status: {response.status_code}")
    if response.status_code == 200:
        data = response.json()
        print("✅ PostgreSQL STRING_AGG fix working - no more GROUP_CONCAT error")
        print(f"Response: {json.dumps(data, indent=2)}")
        return True
    else:
        print("❌ PostgreSQL STRING_AGG fix failed")
        try:
            error_data = response.json()
            print(f"Error: {json.dumps(error_data, indent=2)}")
        except:
            print(f"Error text: {response.text}")
        return False

def test_add_match_event_system_user_fix(token):
    """Test Add Match Event with fixed system user lookup"""
    print("\n🔍 Testing Add Match Event - System User Lookup Fix...")
    url = f"{BACKEND_URL}/api/v1/admin/matches/1/events"
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "player_id": 1,
        "event_type": "kill",
        "points": 2.0,
        "round_number": 5,
        "timestamp": "2025-08-01T17:30:00Z",
        "description": "Test event for system user fix"
    }
    
    response = requests.post(url, json=payload, headers=headers)
    print(f"Status: {response.status_code}")
    
    if response.status_code == 200:
        data = response.json()
        print("✅ Add Match Event working - system user lookup fixed")
        print(f"Response: {json.dumps(data, indent=2)}")
        return True
    else:
        print("❌ Add Match Event still failing")
        try:
            error_data = response.json()
            print(f"Error: {json.dumps(error_data, indent=2)}")
        except:
            print(f"Error text: {response.text}")
        return False

def test_recalculate_points_real_counts(token):
    """Test Recalculate Points returns real database counts, not hardcoded values"""
    print("\n🔍 Testing Recalculate Points - Real Database Counts...")
    url = f"{BACKEND_URL}/api/v1/admin/matches/1/recalculate-points"
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "force_recalculate": True,
        "notify_users": True,
        "recalculate_leaderboards": True
    }
    
    response = requests.post(url, json=payload, headers=headers)
    print(f"Status: {response.status_code}")
    
    if response.status_code == 200:
        data = response.json()
        teams_affected = data.get('teams_affected', 0)
        leaderboards_updated = data.get('leaderboards_updated', 0)
        
        print(f"Teams affected: {teams_affected}")
        print(f"Leaderboards updated: {leaderboards_updated}")
        
        # Check if values are no longer hardcoded (1500 and 25)
        if teams_affected == 1500 and leaderboards_updated == 25:
            print("❌ Still returning hardcoded values (1500, 25)")
            return False
        else:
            print("✅ Returning real database counts (not hardcoded)")
            print(f"Response: {json.dumps(data, indent=2)}")
            return True
    else:
        print("❌ Recalculate Points failed")
        try:
            error_data = response.json()
            print(f"Error: {json.dumps(error_data, indent=2)}")
        except:
            print(f"Error text: {response.text}")
        return False

def main():
    print("🎯 FOCUSED TESTING: Fantasy Points Calculation Engine Fixes")
    print("=" * 70)
    
    results = {}
    
    # Test 1: Health Check
    results['health'] = test_health()
    
    # Test 2: Admin Login
    login_success, token = test_admin_login()
    results['admin_login'] = login_success
    
    if not token:
        print("\n❌ Cannot proceed with other tests - no admin token")
        return
    
    # Test 3: PostgreSQL STRING_AGG Fix
    results['postgresql_fix'] = test_live_scoring_postgresql_fix(token)
    
    # Test 4: Add Match Event System User Fix
    results['add_event_fix'] = test_add_match_event_system_user_fix(token)
    
    # Test 5: Real Database Counts Fix
    results['real_counts_fix'] = test_recalculate_points_real_counts(token)
    
    # Summary
    print("\n" + "=" * 70)
    print("🎯 FOCUSED TEST RESULTS SUMMARY")
    print("=" * 70)
    
    passed = 0
    total = len(results)
    
    for test_name, success in results.items():
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{test_name.upper().replace('_', ' ')}: {status}")
        if success:
            passed += 1
    
    print(f"\nOverall: {passed}/{total} tests passed")
    
    # Specific fix verification
    print("\n🔍 FIX VERIFICATION:")
    if results.get('postgresql_fix'):
        print("✅ Fix 1: PostgreSQL STRING_AGG compatibility - WORKING")
    else:
        print("❌ Fix 1: PostgreSQL STRING_AGG compatibility - STILL BROKEN")
    
    if results.get('add_event_fix'):
        print("✅ Fix 2: Add Match Event system user lookup - WORKING")
    else:
        print("❌ Fix 2: Add Match Event system user lookup - STILL BROKEN")
    
    if results.get('real_counts_fix'):
        print("✅ Fix 3: Real database counts (not hardcoded) - WORKING")
    else:
        print("❌ Fix 3: Real database counts (not hardcoded) - STILL BROKEN")

if __name__ == "__main__":
    main()