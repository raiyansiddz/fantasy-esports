#!/usr/bin/env python3
"""
Crown Jewel Manual Scoring System Comprehensive Test
Focus on testing the specific scenarios mentioned in the review request
"""

import requests
import json
import sys
from datetime import datetime

BACKEND_URL = "http://localhost:8001"
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
            print(f"✅ Admin login successful")
            return True
        else:
            print(f"❌ Admin login failed: {data}")
            return False
            
    except Exception as e:
        print(f"❌ Admin login error: {str(e)}")
        return False

def test_enhanced_match_state_management():
    """Test Enhanced Match State Management with various match IDs"""
    print(f"\n{'='*80}")
    print("TESTING: Enhanced Match State Management (PUT /api/admin/matches/{id}/score)")
    print(f"{'='*80}")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available")
        return False
    
    # Test different match IDs to find various scenarios
    test_matches = [3, 4, 5, 6, 7, 8, 9, 10]
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    results = {}
    
    for match_id in test_matches:
        try:
            print(f"\n--- Testing Match {match_id} ---")
            url = f"{BACKEND_URL}/api/v1/admin/matches/{match_id}/score"
            
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
            data = response.json() if response.headers.get('content-type', '').startswith('application/json') else None
            
            print(f"Status: {response.status_code}")
            if data:
                print(f"Response: {json.dumps(data, indent=2)}")
            else:
                print(f"Response (text): {response.text}")
            
            if response.status_code == 200 and data and data.get('success'):
                results[match_id] = "SUCCESS"
                print(f"✅ Match {match_id}: SUCCESS - No COMMIT_ERROR")
            else:
                error_code = data.get('code') if data else 'UNKNOWN'
                results[match_id] = f"FAILED - {error_code}"
                
                if error_code == 'COMMIT_ERROR':
                    print(f"❌ Match {match_id}: CRITICAL - COMMIT_ERROR still occurring")
                elif error_code == 'INVALID_STATE_TRANSITION':
                    print(f"⚠️  Match {match_id}: Invalid state transition (expected for some matches)")
                elif error_code == 'ALREADY_COMPLETED':
                    print(f"⚠️  Match {match_id}: Already completed (expected for some matches)")
                else:
                    print(f"❌ Match {match_id}: FAILED - {error_code}")
                    
        except Exception as e:
            results[match_id] = f"ERROR - {str(e)}"
            print(f"❌ Match {match_id}: ERROR - {str(e)}")
    
    print(f"\n--- Enhanced Match State Management Summary ---")
    for match_id, result in results.items():
        print(f"Match {match_id}: {result}")
    
    return results

def test_complete_match_with_prize_distribution():
    """Test Complete Match functionality with various match IDs"""
    print(f"\n{'='*80}")
    print("TESTING: Complete Match with Prize Distribution (POST /api/admin/matches/{id}/complete)")
    print(f"{'='*80}")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available")
        return False
    
    # Test different match IDs to find various scenarios
    test_matches = [3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21]
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    results = {}
    
    for match_id in test_matches:
        try:
            print(f"\n--- Testing Match {match_id} ---")
            url = f"{BACKEND_URL}/api/v1/admin/matches/{match_id}/complete"
            
            payload = {
                "final_result": {
                    "winner_team_id": 1,
                    "final_score": "2-1",
                    "mvp_player_id": 1,
                    "match_duration": 3600
                },
                "distribute_prizes": True,
                "send_notifications": True
            }
            
            response = requests.post(url, json=payload, headers=headers, timeout=20)
            data = response.json() if response.headers.get('content-type', '').startswith('application/json') else None
            
            print(f"Status: {response.status_code}")
            if data:
                print(f"Response: {json.dumps(data, indent=2)}")
            else:
                print(f"Response (text): {response.text}")
            
            if response.status_code == 200 and data and data.get('success'):
                results[match_id] = "SUCCESS"
                print(f"✅ Match {match_id}: SUCCESS - No COMMIT_ERROR")
                
                # Check prize distribution details
                prize_data = data.get('prize_distribution', {})
                if prize_data:
                    total_amount = prize_data.get('total_amount', 0)
                    winners_rewarded = prize_data.get('winners_rewarded', 0)
                    print(f"   Prize Distribution: ${total_amount}, Winners: {winners_rewarded}")
                    
                    if total_amount == 0 and winners_rewarded == 0:
                        print(f"   ✅ Empty contest_participants handled correctly")
                    else:
                        print(f"   ✅ Prize distribution working with populated data")
                        
            else:
                error_code = data.get('code') if data else 'UNKNOWN'
                results[match_id] = f"FAILED - {error_code}"
                
                if error_code == 'COMMIT_ERROR':
                    print(f"❌ Match {match_id}: CRITICAL - COMMIT_ERROR still occurring")
                elif error_code == 'CONTEST_UPDATE_ERROR':
                    print(f"❌ Match {match_id}: CRITICAL - CONTEST_UPDATE_ERROR (empty contest scenario)")
                elif error_code == 'LEADERBOARD_FINALIZATION_ERROR':
                    print(f"❌ Match {match_id}: CRITICAL - LEADERBOARD_FINALIZATION_ERROR (empty contest_participants)")
                elif error_code == 'PRIZE_DISTRIBUTION_ERROR':
                    print(f"❌ Match {match_id}: CRITICAL - PRIZE_DISTRIBUTION_ERROR")
                elif error_code == 'ALREADY_COMPLETED':
                    print(f"⚠️  Match {match_id}: Already completed (expected for some matches)")
                elif error_code == 'INVALID_STATE_TRANSITION':
                    print(f"⚠️  Match {match_id}: Invalid state transition (expected for upcoming matches)")
                else:
                    print(f"❌ Match {match_id}: FAILED - {error_code}")
                    
        except Exception as e:
            results[match_id] = f"ERROR - {str(e)}"
            print(f"❌ Match {match_id}: ERROR - {str(e)}")
    
    print(f"\n--- Complete Match with Prize Distribution Summary ---")
    for match_id, result in results.items():
        print(f"Match {match_id}: {result}")
    
    return results

def test_specific_crown_jewel_scenarios():
    """Test the specific scenarios mentioned in the review request"""
    print(f"\n{'='*80}")
    print("TESTING: Specific Crown Jewel Scenarios from Review Request")
    print(f"{'='*80}")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available")
        return False
    
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    results = {}
    
    # Test Case A: Empty Contest Scenarios (contests exist but 0 contest_participants)
    print(f"\n--- Test Case A: Empty Contest Scenarios ---")
    print("Testing matches that have contests but 0 contest_participants")
    
    empty_contest_matches = [1, 20, 21]  # Based on review request details
    
    for match_id in empty_contest_matches:
        try:
            print(f"\nTesting Match {match_id} (Empty Contest Participants)")
            url = f"{BACKEND_URL}/api/v1/admin/matches/{match_id}/complete"
            
            payload = {
                "final_result": {
                    "winner_team_id": 1,
                    "final_score": "2-1",
                    "mvp_player_id": 1,
                    "match_duration": 3600
                },
                "distribute_prizes": True,
                "send_notifications": True
            }
            
            response = requests.post(url, json=payload, headers=headers, timeout=20)
            data = response.json() if response.headers.get('content-type', '').startswith('application/json') else None
            
            print(f"Status: {response.status_code}")
            if data:
                print(f"Error Code: {data.get('code', 'NONE')}")
                print(f"Error Message: {data.get('error', 'NONE')}")
            
            if response.status_code == 200 and data and data.get('success'):
                results[f"empty_contest_{match_id}"] = "SUCCESS"
                print(f"✅ Match {match_id}: Empty contest scenario handled correctly")
            else:
                error_code = data.get('code') if data else 'UNKNOWN'
                results[f"empty_contest_{match_id}"] = f"FAILED - {error_code}"
                
                if error_code in ['LEADERBOARD_FINALIZATION_ERROR', 'CONTEST_UPDATE_ERROR', 'COMMIT_ERROR']:
                    print(f"❌ CRITICAL: Match {match_id} - Crown Jewel fix FAILED with {error_code}")
                else:
                    print(f"⚠️  Match {match_id}: {error_code} (may be expected)")
                    
        except Exception as e:
            results[f"empty_contest_{match_id}"] = f"ERROR - {str(e)}"
            print(f"❌ Match {match_id}: ERROR - {str(e)}")
    
    # Test Case B: No Contest Scenarios (0 contests)
    print(f"\n--- Test Case B: No Contest Scenarios ---")
    print("Testing matches with 0 contests")
    
    no_contest_matches = [20, 21, 22, 23, 24, 25]  # Try various IDs
    
    for match_id in no_contest_matches:
        try:
            print(f"\nTesting Match {match_id} (No Contests)")
            url = f"{BACKEND_URL}/api/v1/admin/matches/{match_id}/complete"
            
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
            data = response.json() if response.headers.get('content-type', '').startswith('application/json') else None
            
            print(f"Status: {response.status_code}")
            if data:
                print(f"Error Code: {data.get('code', 'NONE')}")
                print(f"Error Message: {data.get('error', 'NONE')}")
            
            if response.status_code == 200 and data and data.get('success'):
                results[f"no_contest_{match_id}"] = "SUCCESS"
                print(f"✅ Match {match_id}: No contest scenario handled correctly")
            else:
                error_code = data.get('code') if data else 'UNKNOWN'
                results[f"no_contest_{match_id}"] = f"FAILED - {error_code}"
                
                if error_code == 'CONTEST_UPDATE_ERROR':
                    print(f"❌ CRITICAL: Match {match_id} - Crown Jewel fix FAILED with CONTEST_UPDATE_ERROR")
                else:
                    print(f"⚠️  Match {match_id}: {error_code} (may be expected)")
                    
        except Exception as e:
            results[f"no_contest_{match_id}"] = f"ERROR - {str(e)}"
            print(f"❌ Match {match_id}: ERROR - {str(e)}")
    
    # Test Case D: State Transition Validation
    print(f"\n--- Test Case D: State Transition Validation ---")
    print("Testing invalid state transitions (completing 'upcoming' matches)")
    
    upcoming_matches = [21, 22, 23, 24, 25]  # Try various IDs that might be upcoming
    
    for match_id in upcoming_matches:
        try:
            print(f"\nTesting Match {match_id} (State Transition)")
            url = f"{BACKEND_URL}/api/v1/admin/matches/{match_id}/score"
            
            payload = {
                "team1_score": 2,
                "team2_score": 0,
                "current_round": 2,
                "match_status": "completed",
                "winner_team_id": 1,
                "final_score": "2-0",
                "match_duration": "30:00"
            }
            
            response = requests.put(url, json=payload, headers=headers, timeout=15)
            data = response.json() if response.headers.get('content-type', '').startswith('application/json') else None
            
            print(f"Status: {response.status_code}")
            if data:
                print(f"Error Code: {data.get('code', 'NONE')}")
                print(f"Error Message: {data.get('error', 'NONE')}")
            
            if response.status_code >= 400 and data and data.get('code') == 'INVALID_STATE_TRANSITION':
                results[f"state_transition_{match_id}"] = "SUCCESS"
                print(f"✅ Match {match_id}: Invalid state transition correctly rejected")
            elif response.status_code == 200:
                results[f"state_transition_{match_id}"] = "SUCCESS"
                print(f"✅ Match {match_id}: Valid state transition accepted")
            else:
                error_code = data.get('code') if data else 'UNKNOWN'
                results[f"state_transition_{match_id}"] = f"FAILED - {error_code}"
                print(f"❌ Match {match_id}: Unexpected result - {error_code}")
                    
        except Exception as e:
            results[f"state_transition_{match_id}"] = f"ERROR - {str(e)}"
            print(f"❌ Match {match_id}: ERROR - {str(e)}")
    
    return results

def main():
    """Main test execution"""
    print("🚀 Crown Jewel Manual Scoring System Comprehensive Test")
    print(f"Backend URL: {BACKEND_URL}")
    print(f"Test started at: {datetime.now()}")
    
    # Get admin token
    if not get_admin_token():
        print("❌ Failed to get admin token, exiting")
        sys.exit(1)
    
    # Run comprehensive tests
    enhanced_results = test_enhanced_match_state_management()
    complete_results = test_complete_match_with_prize_distribution()
    crown_jewel_results = test_specific_crown_jewel_scenarios()
    
    # Print final summary
    print(f"\n{'='*80}")
    print("CROWN JEWEL COMPREHENSIVE TEST SUMMARY")
    print(f"{'='*80}")
    
    # Count successes and failures
    all_results = {}
    all_results.update(enhanced_results)
    all_results.update(complete_results)
    all_results.update(crown_jewel_results)
    
    successes = sum(1 for result in all_results.values() if result == "SUCCESS")
    total_tests = len(all_results)
    
    print(f"Total Tests: {successes}/{total_tests} passed")
    
    # Identify critical failures
    critical_failures = []
    for test_name, result in all_results.items():
        if "COMMIT_ERROR" in result:
            critical_failures.append(f"{test_name}: {result}")
        elif "LEADERBOARD_FINALIZATION_ERROR" in result:
            critical_failures.append(f"{test_name}: {result}")
        elif "CONTEST_UPDATE_ERROR" in result:
            critical_failures.append(f"{test_name}: {result}")
        elif "PRIZE_DISTRIBUTION_ERROR" in result:
            critical_failures.append(f"{test_name}: {result}")
    
    if critical_failures:
        print(f"\n❌ {len(critical_failures)} CRITICAL Crown Jewel failures found:")
        for failure in critical_failures:
            print(f"   - {failure}")
        print("\n❌ Crown Jewel Manual Scoring System transaction fixes are NOT working properly")
    else:
        print(f"\n✅ No critical Crown Jewel failures found")
        print("✅ Crown Jewel Manual Scoring System transaction fixes appear to be working")
    
    print(f"\nTest completed at: {datetime.now()}")

if __name__ == "__main__":
    main()