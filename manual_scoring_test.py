#!/usr/bin/env python3
"""
Manual Scoring System Testing Script
Tests the newly implemented Manual Scoring System features in the GoLang fantasy sports backend
Focus: Enhanced Match State Management and Complete Match with Prize Distribution
"""

import requests
import json
import sys
from datetime import datetime

# Backend URL
BACKEND_URL = "http://localhost:8080"

# Global admin token storage
ADMIN_TOKEN = None

def print_test_header(test_name):
    """Print formatted test header"""
    print(f"\n{'='*80}")
    print(f"TESTING: {test_name}")
    print(f"{'='*80}")

def print_response(response, endpoint):
    """Print formatted response details"""
    print(f"\nEndpoint: {endpoint}")
    print(f"Status Code: {response.status_code}")
    print(f"Response Headers: {dict(response.headers)}")
    
    try:
        response_json = response.json()
        print(f"Response Body: {json.dumps(response_json, indent=2)}")
        return response_json
    except:
        print(f"Response Body (text): {response.text}")
        return None

def test_admin_login():
    """Test admin login to get authentication token"""
    print_test_header("Admin Login")
    
    global ADMIN_TOKEN
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/login"
        payload = {
            "username": "admin",
            "password": "admin123"
        }
        
        response = requests.post(url, json=payload, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            ADMIN_TOKEN = data.get('access_token')
            admin_user = data.get('admin_user', {})
            print(f"✅ Admin login PASSED - Token obtained for user: {admin_user.get('username', 'Unknown')}")
            print(f"   Admin Role: {admin_user.get('role', 'Unknown')}")
            return True, data
        else:
            print("❌ Admin login FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Admin login ERROR: {str(e)}")
        return False, None

def test_enhanced_match_state_management():
    """Test Enhanced Match State Management endpoint"""
    print_test_header("Enhanced Match State Management - PUT /api/v1/admin/matches/{id}/score")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    test_results = {}
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    # Test 1: Valid state transition (upcoming -> live)
    print("\n--- Test 1: Valid State Transition (upcoming -> live) ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/score"
        payload = {
            "team1_score": 0,
            "team2_score": 0,
            "current_round": 1,
            "match_status": "live",
            "winner_team_id": None,
            "final_score": "0-0",
            "match_duration": "00:05:30"
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            test_results['valid_transition_live'] = True
            print("✅ Valid state transition (upcoming -> live) PASSED")
            print(f"   Match Status: {data.get('status')}")
            print(f"   State Transition: {data.get('state_transition')}")
            print(f"   Score Validation: {data.get('score_validation')}")
        else:
            test_results['valid_transition_live'] = False
            print("❌ Valid state transition (upcoming -> live) FAILED")
            
    except Exception as e:
        test_results['valid_transition_live'] = False
        print(f"❌ Valid state transition ERROR: {str(e)}")
    
    # Test 2: Valid state transition (live -> completed)
    print("\n--- Test 2: Valid State Transition (live -> completed) ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/score"
        payload = {
            "team1_score": 13,
            "team2_score": 8,
            "current_round": 21,
            "match_status": "completed",
            "winner_team_id": 1,
            "final_score": "13-8",
            "match_duration": "00:45:30"
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            test_results['valid_transition_completed'] = True
            print("✅ Valid state transition (live -> completed) PASSED")
            print(f"   Winner Team: {data.get('winner_team')}")
            print(f"   Final Score: {data.get('final_score')}")
            print(f"   Completion Data: {data.get('completion_data') is not None}")
        else:
            test_results['valid_transition_completed'] = False
            print("❌ Valid state transition (live -> completed) FAILED")
            
    except Exception as e:
        test_results['valid_transition_completed'] = False
        print(f"❌ Valid state transition ERROR: {str(e)}")
    
    # Test 3: Invalid state transition (completed -> live should fail)
    print("\n--- Test 3: Invalid State Transition (completed -> live should fail) ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/score"
        payload = {
            "team1_score": 5,
            "team2_score": 3,
            "current_round": 8,
            "match_status": "live",
            "winner_team_id": None,
            "final_score": "5-3",
            "match_duration": "00:15:30"
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 400 and data and not data.get('success'):
            test_results['invalid_transition'] = True
            print("✅ Invalid state transition (completed -> live) correctly REJECTED")
            print(f"   Error Code: {data.get('code')}")
            print(f"   Error Message: {data.get('error')}")
        else:
            test_results['invalid_transition'] = False
            print("❌ Invalid state transition should have been rejected but wasn't")
            
    except Exception as e:
        test_results['invalid_transition'] = False
        print(f"❌ Invalid state transition test ERROR: {str(e)}")
    
    # Test 4: Score validation for best-of matches
    print("\n--- Test 4: Score Validation for Best-of Matches ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/2/score"  # Different match ID
        payload = {
            "team1_score": 2,
            "team2_score": 1,
            "current_round": 3,
            "match_status": "live",
            "winner_team_id": None,
            "final_score": "2-1",
            "match_duration": "01:30:45"
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            test_results['score_validation'] = True
            print("✅ Score validation for best-of matches PASSED")
            print(f"   Score Validation: {data.get('score_validation')}")
        else:
            test_results['score_validation'] = False
            print("❌ Score validation for best-of matches FAILED")
            
    except Exception as e:
        test_results['score_validation'] = False
        print(f"❌ Score validation test ERROR: {str(e)}")
    
    # Test 5: Error handling for non-existent matches
    print("\n--- Test 5: Error Handling for Non-existent Matches ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/99999/score"
        payload = {
            "team1_score": 1,
            "team2_score": 0,
            "current_round": 1,
            "match_status": "live",
            "winner_team_id": None,
            "final_score": "1-0",
            "match_duration": "00:10:00"
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 404 and data and not data.get('success'):
            test_results['nonexistent_match'] = True
            print("✅ Non-existent match error handling PASSED")
            print(f"   Error Code: {data.get('code')}")
            print(f"   Error Message: {data.get('error')}")
        else:
            test_results['nonexistent_match'] = False
            print("❌ Non-existent match should return 404 error")
            
    except Exception as e:
        test_results['nonexistent_match'] = False
        print(f"❌ Non-existent match test ERROR: {str(e)}")
    
    return test_results

def test_complete_match_with_prize_distribution():
    """Test Complete Match with Prize Distribution endpoint"""
    print_test_header("Complete Match with Prize Distribution - POST /api/v1/admin/matches/{id}/complete")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    test_results = {}
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    # Test 1: Match completion with prize distribution enabled
    print("\n--- Test 1: Match Completion with Prize Distribution Enabled ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/3/complete"  # Use different match ID
        payload = {
            "final_result": {
                "winner_team_id": 5,
                "final_score": "16-14",
                "mvp_player_id": 12,
                "match_duration": 2700  # 45 minutes in seconds
            },
            "distribute_prizes": True,
            "send_notifications": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            test_results['completion_with_prizes'] = True
            print("✅ Match completion with prize distribution PASSED")
            print(f"   Winner Team: {data.get('winner_team')}")
            print(f"   MVP Player: {data.get('mvp_player')}")
            print(f"   Fantasy Teams Finalized: {data.get('fantasy_teams_finalized')}")
            print(f"   Leaderboards Finalized: {data.get('leaderboards_finalized')}")
            print(f"   Contests Updated: {data.get('contests_updated')}")
            print(f"   Notifications Sent: {data.get('notifications_sent')}")
            print(f"   Statistics Updated: {data.get('statistics_updated')}")
            print(f"   Prizes Distributed: {data.get('prizes_distributed')}")
            print(f"   Prize Distribution Details: {data.get('prize_distribution') is not None}")
        else:
            test_results['completion_with_prizes'] = False
            print("❌ Match completion with prize distribution FAILED")
            
    except Exception as e:
        test_results['completion_with_prizes'] = False
        print(f"❌ Match completion with prizes ERROR: {str(e)}")
    
    # Test 2: Match completion without prize distribution
    print("\n--- Test 2: Match Completion without Prize Distribution ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/4/complete"  # Use different match ID
        payload = {
            "final_result": {
                "winner_team_id": 7,
                "final_score": "13-11",
                "mvp_player_id": 15,
                "match_duration": 2400  # 40 minutes in seconds
            },
            "distribute_prizes": False,
            "send_notifications": False
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            test_results['completion_without_prizes'] = True
            print("✅ Match completion without prize distribution PASSED")
            print(f"   Prizes Distributed: {data.get('prizes_distributed')}")
            print(f"   Notifications Sent: {data.get('notifications_sent')}")
            print(f"   Prize Distribution Details: {data.get('prize_distribution')}")
        else:
            test_results['completion_without_prizes'] = False
            print("❌ Match completion without prize distribution FAILED")
            
    except Exception as e:
        test_results['completion_without_prizes'] = False
        print(f"❌ Match completion without prizes ERROR: {str(e)}")
    
    # Test 3: Completion of already completed match (should fail)
    print("\n--- Test 3: Completion of Already Completed Match (should fail) ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/3/complete"  # Same match as Test 1
        payload = {
            "final_result": {
                "winner_team_id": 6,
                "final_score": "16-10",
                "mvp_player_id": 18,
                "match_duration": 2100
            },
            "distribute_prizes": True,
            "send_notifications": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 400 and data and not data.get('success'):
            test_results['already_completed'] = True
            print("✅ Already completed match correctly REJECTED")
            print(f"   Error Code: {data.get('code')}")
            print(f"   Error Message: {data.get('error')}")
        else:
            test_results['already_completed'] = False
            print("❌ Already completed match should have been rejected")
            
    except Exception as e:
        test_results['already_completed'] = False
        print(f"❌ Already completed match test ERROR: {str(e)}")
    
    # Test 4: Comprehensive response structure validation
    print("\n--- Test 4: Comprehensive Response Structure Validation ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/5/complete"  # Use different match ID
        payload = {
            "final_result": {
                "winner_team_id": 9,
                "final_score": "16-12",
                "mvp_player_id": 21,
                "match_duration": 3000
            },
            "distribute_prizes": True,
            "send_notifications": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            # Check for all expected response fields
            expected_fields = [
                'success', 'match_id', 'winner_team', 'mvp_player', 'final_score',
                'match_duration', 'fantasy_teams_finalized', 'leaderboards_finalized',
                'contests_updated', 'notifications_sent', 'statistics_updated',
                'prizes_distributed', 'completion_timestamp', 'message'
            ]
            
            missing_fields = [field for field in expected_fields if field not in data]
            
            if not missing_fields:
                test_results['response_structure'] = True
                print("✅ Comprehensive response structure PASSED")
                print(f"   All expected fields present: {len(expected_fields)} fields")
            else:
                test_results['response_structure'] = False
                print(f"❌ Missing response fields: {missing_fields}")
        else:
            test_results['response_structure'] = False
            print("❌ Comprehensive response structure test FAILED")
            
    except Exception as e:
        test_results['response_structure'] = False
        print(f"❌ Response structure test ERROR: {str(e)}")
    
    return test_results

def test_error_handling_scenarios():
    """Test various error handling scenarios"""
    print_test_header("Error Handling Scenarios")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    test_results = {}
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    # Test 1: Invalid request format for UpdateMatchScore
    print("\n--- Test 1: Invalid Request Format for UpdateMatchScore ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/score"
        payload = {
            "invalid_field": "invalid_value"
            # Missing required fields
        }
        
        response = requests.put(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 400 and data and not data.get('success'):
            test_results['invalid_request_format'] = True
            print("✅ Invalid request format correctly handled")
        else:
            test_results['invalid_request_format'] = False
            print("❌ Invalid request format should return 400 error")
            
    except Exception as e:
        test_results['invalid_request_format'] = False
        print(f"❌ Invalid request format test ERROR: {str(e)}")
    
    # Test 2: Invalid request format for CompleteMatch
    print("\n--- Test 2: Invalid Request Format for CompleteMatch ---")
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/complete"
        payload = {
            "invalid_field": "invalid_value"
            # Missing required final_result field
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 400 and data and not data.get('success'):
            test_results['invalid_complete_format'] = True
            print("✅ Invalid complete request format correctly handled")
        else:
            test_results['invalid_complete_format'] = False
            print("❌ Invalid complete request format should return 400 error")
            
    except Exception as e:
        test_results['invalid_complete_format'] = False
        print(f"❌ Invalid complete request format test ERROR: {str(e)}")
    
    return test_results

def main():
    """Main test execution for Manual Scoring System"""
    print("🚀 Starting Manual Scoring System Tests")
    print(f"Backend URL: {BACKEND_URL}")
    print(f"Test started at: {datetime.now()}")
    
    # Track test results
    all_results = {}
    
    # Run tests in order
    all_results['admin_login'] = test_admin_login()
    
    if ADMIN_TOKEN:
        all_results['enhanced_match_state'] = test_enhanced_match_state_management()
        all_results['complete_match_prizes'] = test_complete_match_with_prize_distribution()
        all_results['error_handling'] = test_error_handling_scenarios()
    else:
        print("❌ Cannot run tests without admin token")
        sys.exit(1)
    
    # Print comprehensive summary
    print(f"\n{'='*80}")
    print("MANUAL SCORING SYSTEM TEST SUMMARY")
    print(f"{'='*80}")
    
    # Admin login summary
    login_success, _ = all_results['admin_login']
    print(f"Admin Login: {'✅ PASSED' if login_success else '❌ FAILED'}")
    
    # Enhanced Match State Management summary
    if isinstance(all_results['enhanced_match_state'], dict):
        state_results = all_results['enhanced_match_state']
        print(f"\nEnhanced Match State Management:")
        print(f"  Valid Transition (live): {'✅ PASSED' if state_results.get('valid_transition_live') else '❌ FAILED'}")
        print(f"  Valid Transition (completed): {'✅ PASSED' if state_results.get('valid_transition_completed') else '❌ FAILED'}")
        print(f"  Invalid Transition Rejection: {'✅ PASSED' if state_results.get('invalid_transition') else '❌ FAILED'}")
        print(f"  Score Validation: {'✅ PASSED' if state_results.get('score_validation') else '❌ FAILED'}")
        print(f"  Non-existent Match Handling: {'✅ PASSED' if state_results.get('nonexistent_match') else '❌ FAILED'}")
    
    # Complete Match with Prize Distribution summary
    if isinstance(all_results['complete_match_prizes'], dict):
        complete_results = all_results['complete_match_prizes']
        print(f"\nComplete Match with Prize Distribution:")
        print(f"  Completion with Prizes: {'✅ PASSED' if complete_results.get('completion_with_prizes') else '❌ FAILED'}")
        print(f"  Completion without Prizes: {'✅ PASSED' if complete_results.get('completion_without_prizes') else '❌ FAILED'}")
        print(f"  Already Completed Rejection: {'✅ PASSED' if complete_results.get('already_completed') else '❌ FAILED'}")
        print(f"  Response Structure: {'✅ PASSED' if complete_results.get('response_structure') else '❌ FAILED'}")
    
    # Error Handling summary
    if isinstance(all_results['error_handling'], dict):
        error_results = all_results['error_handling']
        print(f"\nError Handling:")
        print(f"  Invalid UpdateMatchScore Format: {'✅ PASSED' if error_results.get('invalid_request_format') else '❌ FAILED'}")
        print(f"  Invalid CompleteMatch Format: {'✅ PASSED' if error_results.get('invalid_complete_format') else '❌ FAILED'}")
    
    # Overall assessment
    print(f"\n{'='*80}")
    print("OVERALL ASSESSMENT")
    print(f"{'='*80}")
    
    # Count passed tests
    total_tests = 0
    passed_tests = 0
    
    if login_success:
        passed_tests += 1
    total_tests += 1
    
    if isinstance(all_results['enhanced_match_state'], dict):
        state_results = all_results['enhanced_match_state']
        for result in state_results.values():
            total_tests += 1
            if result:
                passed_tests += 1
    
    if isinstance(all_results['complete_match_prizes'], dict):
        complete_results = all_results['complete_match_prizes']
        for result in complete_results.values():
            total_tests += 1
            if result:
                passed_tests += 1
    
    if isinstance(all_results['error_handling'], dict):
        error_results = all_results['error_handling']
        for result in error_results.values():
            total_tests += 1
            if result:
                passed_tests += 1
    
    print(f"Total Tests: {passed_tests}/{total_tests} passed ({(passed_tests/total_tests)*100:.1f}%)")
    print(f"Test completed at: {datetime.now()}")
    
    # Determine critical failures
    critical_failures = []
    
    if not login_success:
        critical_failures.append("Admin login failed")
    
    if isinstance(all_results['enhanced_match_state'], dict):
        state_results = all_results['enhanced_match_state']
        if not state_results.get('valid_transition_live'):
            critical_failures.append("Valid state transitions not working")
        if not state_results.get('invalid_transition'):
            critical_failures.append("Invalid state transition validation not working")
    
    if isinstance(all_results['complete_match_prizes'], dict):
        complete_results = all_results['complete_match_prizes']
        if not complete_results.get('completion_with_prizes'):
            critical_failures.append("Match completion with prizes not working")
        if not complete_results.get('already_completed'):
            critical_failures.append("Already completed match validation not working")
    
    if critical_failures:
        print(f"\n❌ CRITICAL ISSUES FOUND:")
        for failure in critical_failures:
            print(f"   - {failure}")
        print(f"\n❌ Manual Scoring System has {len(critical_failures)} critical issue(s)")
        sys.exit(1)
    else:
        print(f"\n✅ All critical functionality working - Manual Scoring System is operational")
        sys.exit(0)

if __name__ == "__main__":
    main()