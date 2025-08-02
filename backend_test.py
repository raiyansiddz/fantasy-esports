#!/usr/bin/env python3
"""
Fantasy Sports Backend API Testing Script
Tests the GoLang fantasy sports backend running on the configured URL
Focus: Fantasy Points Calculation Engine Testing
"""

import requests
import json
import sys
from datetime import datetime

# Get backend URL from environment - using localhost for GoLang backend
BACKEND_URL = "http://localhost:8080"

# Global admin token storage
ADMIN_TOKEN = None

def print_test_header(test_name):
    """Print formatted test header"""
    print(f"\n{'='*60}")
    print(f"TESTING: {test_name}")
    print(f"{'='*60}")

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

def test_add_match_event():
    """Test adding a match event with fantasy points calculation"""
    print_test_header("Add Match Event (Fantasy Points Engine)")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/events"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        payload = {
            "player_id": 1,
            "event_type": "kill",
            "points": 2.0,
            "round_number": 5,
            "timestamp": datetime.now().isoformat() + "Z",
            "description": "Entry frag",
            "additional_data": {}
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            # Check for NEW fantasy points calculation response
            message = data.get('message', '')
            fantasy_teams_affected = data.get('fantasy_teams_affected', 0)
            
            print(f"✅ Add match event PASSED")
            print(f"   Event ID: {data.get('event_id')}")
            print(f"   Player: {data.get('player_name')} ({data.get('team_name')})")
            print(f"   Event Type: {data.get('event_type')}")
            print(f"   Points: {data.get('points')}")
            print(f"   Message: {message}")
            print(f"   Fantasy Teams Affected: {fantasy_teams_affected}")
            
            # Verify NEW behavior vs old mock
            if "fantasy points recalculated" in message.lower():
                print("✅ NEW: Response shows 'fantasy points recalculated' message")
            else:
                print("⚠️  OLD: Response still shows mock message")
            
            if isinstance(fantasy_teams_affected, int) and fantasy_teams_affected != 1250:
                print(f"✅ NEW: Fantasy teams affected shows real number: {fantasy_teams_affected}")
            elif fantasy_teams_affected == 1250:
                print("⚠️  OLD: Fantasy teams affected shows hardcoded 1250")
            else:
                print(f"⚠️  UNKNOWN: Fantasy teams affected: {fantasy_teams_affected}")
            
            return True, data
        else:
            print("❌ Add match event FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Add match event ERROR: {str(e)}")
        return False, None

def test_recalculate_points():
    """Test manual fantasy points recalculation"""
    print_test_header("Recalculate Fantasy Points")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/recalculate-points"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        payload = {
            "force_recalculate": True,
            "notify_users": True,
            "recalculate_leaderboards": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            teams_affected = data.get('teams_affected', 0)
            leaderboards_updated = data.get('leaderboards_updated', 0)
            message = data.get('message', '')
            
            print(f"✅ Recalculate points PASSED")
            print(f"   Match ID: {data.get('match_id')}")
            print(f"   Force Recalculate: {data.get('force_recalculate')}")
            print(f"   Teams Affected: {teams_affected}")
            print(f"   Leaderboards Updated: {leaderboards_updated}")
            print(f"   Notifications Sent: {data.get('notifications_sent')}")
            print(f"   Message: {message}")
            
            # Verify NEW behavior vs old mock
            if teams_affected != 1500:
                print(f"✅ NEW: Teams affected shows real number: {teams_affected}")
            else:
                print("⚠️  OLD: Teams affected shows hardcoded 1500")
            
            if leaderboards_updated != 25:
                print(f"✅ NEW: Leaderboards updated shows real number: {leaderboards_updated}")
            else:
                print("⚠️  OLD: Leaderboards updated shows hardcoded 25")
            
            return True, data
        else:
            print("❌ Recalculate points FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Recalculate points ERROR: {str(e)}")
        return False, None

def test_recalculate_points_variations():
    """Test recalculate points with different parameters"""
    print_test_header("Recalculate Points - Parameter Variations")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    test_cases = [
        {
            "name": "Force=False, Notify=False",
            "payload": {"force_recalculate": False, "notify_users": False, "recalculate_leaderboards": True}
        },
        {
            "name": "Force=True, Notify=False", 
            "payload": {"force_recalculate": True, "notify_users": False, "recalculate_leaderboards": False}
        },
        {
            "name": "All False",
            "payload": {"force_recalculate": False, "notify_users": False, "recalculate_leaderboards": False}
        }
    ]
    
    results = {}
    
    for test_case in test_cases:
        try:
            url = f"{BACKEND_URL}/api/v1/admin/matches/1/recalculate-points"
            headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
            
            response = requests.post(url, json=test_case["payload"], headers=headers, timeout=10)
            data = response.json() if response.status_code == 200 else None
            
            if response.status_code == 200 and data and data.get('success'):
                results[test_case["name"]] = {
                    "success": True,
                    "teams_affected": data.get('teams_affected', 0),
                    "leaderboards_updated": data.get('leaderboards_updated', 0)
                }
                print(f"✅ {test_case['name']}: SUCCESS - Teams: {data.get('teams_affected')}, Leaderboards: {data.get('leaderboards_updated')}")
            else:
                results[test_case["name"]] = {"success": False}
                print(f"❌ {test_case['name']}: FAILED")
                
        except Exception as e:
            results[test_case["name"]] = {"success": False, "error": str(e)}
            print(f"❌ {test_case['name']}: ERROR - {str(e)}")
    
    return results

def test_multiple_match_events():
    """Test adding multiple match events to verify consistent behavior"""
    print_test_header("Multiple Match Events")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    events = [
        {"player_id": 2, "event_type": "death", "points": -1.0, "description": "Eliminated by enemy"},
        {"player_id": 3, "event_type": "assist", "points": 1.5, "description": "Assisted teammate"},
        {"player_id": 1, "event_type": "headshot", "points": 1.0, "description": "Precision shot"},
        {"player_id": 4, "event_type": "ace", "points": 8.0, "description": "Team wipe"}
    ]
    
    results = []
    
    for i, event in enumerate(events):
        try:
            url = f"{BACKEND_URL}/api/v1/admin/matches/1/events"
            headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
            payload = {
                **event,
                "round_number": 6 + i,
                "timestamp": datetime.now().isoformat() + "Z",
                "additional_data": {}
            }
            
            response = requests.post(url, json=payload, headers=headers, timeout=10)
            data = response.json() if response.status_code == 200 else None
            
            if response.status_code == 200 and data and data.get('success'):
                fantasy_teams = data.get('fantasy_teams_affected', 0)
                results.append({
                    "event_type": event["event_type"],
                    "success": True,
                    "fantasy_teams_affected": fantasy_teams,
                    "event_id": data.get('event_id')
                })
                print(f"✅ Event {i+1} ({event['event_type']}): SUCCESS - Teams affected: {fantasy_teams}")
            else:
                results.append({"event_type": event["event_type"], "success": False})
                print(f"❌ Event {i+1} ({event['event_type']}): FAILED")
                
        except Exception as e:
            results.append({"event_type": event["event_type"], "success": False, "error": str(e)})
            print(f"❌ Event {i+1} ({event['event_type']}): ERROR - {str(e)}")
    
    return results

def test_health_check():
    """Test the health check endpoint"""
    print_test_header("Health Check")
    
    try:
        url = f"{BACKEND_URL}/health"
        response = requests.get(url, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200:
            print("✅ Health check PASSED")
            return True, data
        else:
            print("❌ Health check FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Health check ERROR: {str(e)}")
        return False, None

def test_error_handling():
    """Test error handling for admin endpoints"""
    print_test_header("Error Handling Tests")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    results = {}
    
    # Test invalid match ID
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/99999/events"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        payload = {
            "player_id": 1,
            "event_type": "kill",
            "points": 2.0,
            "timestamp": datetime.now().isoformat() + "Z",
            "additional_data": {}
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        results['invalid_match'] = response.status_code in [400, 404, 500]
        print(f"Invalid match ID: {'✅ HANDLED' if results['invalid_match'] else '❌ NOT HANDLED'} (Status: {response.status_code})")
        
    except Exception as e:
        results['invalid_match'] = False
        print(f"Invalid match ID: ❌ ERROR - {str(e)}")
    
    # Test invalid player ID
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/events"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        payload = {
            "player_id": 99999,
            "event_type": "kill",
            "points": 2.0,
            "timestamp": datetime.now().isoformat() + "Z",
            "additional_data": {}
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        results['invalid_player'] = response.status_code in [400, 404, 500]
        print(f"Invalid player ID: {'✅ HANDLED' if results['invalid_player'] else '❌ NOT HANDLED'} (Status: {response.status_code})")
        
    except Exception as e:
        results['invalid_player'] = False
        print(f"Invalid player ID: ❌ ERROR - {str(e)}")
    
    # Test missing required fields
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/events"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        payload = {
            "event_type": "kill",
            "points": 2.0,
            "additional_data": {}
            # Missing player_id and timestamp
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=10)
        results['missing_fields'] = response.status_code == 400
        print(f"Missing required fields: {'✅ HANDLED' if results['missing_fields'] else '❌ NOT HANDLED'} (Status: {response.status_code})")
        
    except Exception as e:
        results['missing_fields'] = False
        print(f"Missing required fields: ❌ ERROR - {str(e)}")
    
    return results

def test_enhanced_match_state_management():
    """Test Enhanced Match State Management with complex state validation"""
    print_test_header("Enhanced Match State Management")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    try:
        url = f"{BACKEND_URL}/api/v1/admin/matches/1/score"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        
        # Test Case 1: Valid state transition with correct request format for best-of-3 match
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
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            print(f"✅ Enhanced Match State Management PASSED")
            print(f"   Match Status: {data.get('status')}")
            print(f"   Final Score: {data.get('final_score')}")
            print(f"   Winner Team: {data.get('winner_team')}")
            print(f"   State Transition: {data.get('state_transition')}")
            print(f"   Score Validation: {data.get('score_validation')}")
            
            # Check for complex state management features
            if data.get('state_transition') == 'valid':
                print("✅ State transition validation working")
            if data.get('score_validation'):
                print("✅ Score validation working")
            if data.get('completion_data'):
                print("✅ Match completion logic working")
                
            return True, data
        else:
            print("❌ Enhanced Match State Management FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Enhanced Match State Management ERROR: {str(e)}")
        return False, None

def test_complete_match_with_prize_distribution():
    """Test Complete Match functionality with real prize distribution - Crown Jewel Fix"""
    print_test_header("Complete Match with Prize Distribution - Crown Jewel Transaction Fix")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    try:
        # Test with match ID that likely has empty contest_participants table
        url = f"{BACKEND_URL}/api/v1/admin/matches/10/complete"
        headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
        
        # Use correct request format based on CompleteMatchRequest model
        payload = {
            "final_result": {
                "winner_team_id": 2,
                "final_score": "2-1",
                "mvp_player_id": 3,
                "match_duration": 3600
            },
            "distribute_prizes": True,
            "send_notifications": True
        }
        
        response = requests.post(url, json=payload, headers=headers, timeout=20)
        data = print_response(response, url)
        
        if response.status_code == 200 and data and data.get('success'):
            print(f"✅ Complete Match with Prize Distribution PASSED - NO COMMIT_ERROR")
            print(f"   Match ID: {data.get('match_id')}")
            print(f"   Winner Team: {data.get('winner_team')}")
            print(f"   MVP Player: {data.get('mvp_player')}")
            print(f"   Fantasy Teams Finalized: {data.get('fantasy_teams_finalized', 0)}")
            print(f"   Leaderboards Finalized: {data.get('leaderboards_finalized', 0)}")
            print(f"   Contests Updated: {data.get('contests_updated', 0)}")
            print(f"   Notifications Sent: {data.get('notifications_sent', 0)}")
            print(f"   Statistics Updated: {data.get('statistics_updated', False)}")
            
            # Check for Crown Jewel fix - should handle empty contest_participants gracefully
            prize_data = data.get('prize_distribution', {})
            if prize_data:
                total_amount = prize_data.get('total_amount', 0)
                winners_rewarded = prize_data.get('winners_rewarded', 0)
                contests_processed = prize_data.get('contests_processed', 0)
                
                print(f"✅ Crown Jewel Fix Working: Prize distribution completed without transaction errors")
                print(f"   Total Amount: ${total_amount}")
                print(f"   Winners Rewarded: {winners_rewarded}")
                print(f"   Contests Processed: {contests_processed}")
                
                if total_amount == 0 and winners_rewarded == 0:
                    print("✅ Empty contest_participants handled correctly - zero distributions returned")
                else:
                    print("✅ Prize distribution working with populated data")
            
            return True, data
        else:
            error_code = data.get('code') if data else 'UNKNOWN'
            if error_code == 'COMMIT_ERROR':
                print("❌ CRITICAL: Crown Jewel fix FAILED - Still getting COMMIT_ERROR")
            elif error_code == 'PRIZE_DISTRIBUTION_ERROR':
                print("❌ CRITICAL: Crown Jewel fix FAILED - Still getting PRIZE_DISTRIBUTION_ERROR")
            else:
                print(f"❌ Complete Match with Prize Distribution FAILED - Error: {error_code}")
            return False, data
            
    except Exception as e:
        print(f"❌ Complete Match with Prize Distribution ERROR: {str(e)}")
        return False, None

def test_state_transition_validation():
    """Test various state transition scenarios"""
    print_test_header("State Transition Validation Tests")
    
    if not ADMIN_TOKEN:
        print("❌ No admin token available - skipping test")
        return False, None
    
    test_cases = [
        {
            "name": "Invalid transition: completed to live",
            "match_id": "1",
            "payload": {
                "match_status": "live", 
                "team1_score": 0, 
                "team2_score": 0,
                "current_round": 1,
                "final_score": "0-0",
                "match_duration": "00:00"
            },
            "should_fail": True
        },
        {
            "name": "Valid transition: upcoming to live",
            "match_id": "4",
            "payload": {
                "match_status": "live", 
                "team1_score": 0, 
                "team2_score": 0,
                "current_round": 1,
                "final_score": "0-0",
                "match_duration": "00:00"
            },
            "should_fail": False
        },
        {
            "name": "Invalid score format",
            "match_id": "5",
            "payload": {
                "match_status": "completed", 
                "team1_score": -1, 
                "team2_score": 5,
                "current_round": 10,
                "final_score": "-1-5",
                "match_duration": "20:00"
            },
            "should_fail": True
        }
    ]
    
    results = {}
    headers = {"Authorization": f"Bearer {ADMIN_TOKEN}"}
    
    for test_case in test_cases:
        try:
            url = f"{BACKEND_URL}/api/v1/admin/matches/{test_case['match_id']}/score"
            response = requests.put(url, json=test_case["payload"], headers=headers, timeout=10)
            
            if test_case["should_fail"]:
                success = response.status_code >= 400
                status = "✅ CORRECTLY REJECTED" if success else "❌ SHOULD HAVE FAILED"
            else:
                success = response.status_code == 200
                status = "✅ PASSED" if success else "❌ FAILED"
            
            results[test_case["name"]] = success
            print(f"{status}: {test_case['name']} (Status: {response.status_code})")
            
        except Exception as e:
            results[test_case["name"]] = False
            print(f"❌ ERROR: {test_case['name']} - {str(e)}")
    
    return results

def main():
    """Main test execution for Manual Scoring System (Crown Jewel) Features"""
    print("🚀 Starting Manual Scoring System (Crown Jewel) Tests")
    print(f"Backend URL: {BACKEND_URL}")
    print(f"Test started at: {datetime.now()}")
    
    # Track test results
    test_results = {}
    
    # Run prerequisite tests
    test_results['health'] = test_health_check()
    test_results['admin_login'] = test_admin_login()
    
    # Run Crown Jewel feature tests
    if ADMIN_TOKEN:
        test_results['enhanced_match_state'] = test_enhanced_match_state_management()
        test_results['complete_match_prizes'] = test_complete_match_with_prize_distribution()
        state_validation_results = test_state_transition_validation()
    else:
        test_results['enhanced_match_state'] = (False, "No admin token")
        test_results['complete_match_prizes'] = (False, "No admin token")
        state_validation_results = {}
    
    # Print summary
    print(f"\n{'='*60}")
    print("MANUAL SCORING SYSTEM (CROWN JEWEL) TEST SUMMARY")
    print(f"{'='*60}")
    
    passed = 0
    total = 0
    
    for test_name, (success, data) in test_results.items():
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{test_name.upper()}: {status}")
        if success:
            passed += 1
        total += 1
    
    print(f"\nCore Tests: {passed}/{total} passed")
    
    # State validation tests summary
    print(f"\n{'='*60}")
    print("STATE VALIDATION TESTS")
    print(f"{'='*60}")
    
    validation_passed = sum(1 for result in state_validation_results.values() if result)
    validation_total = len(state_validation_results)
    print(f"State Validation Tests: {validation_passed}/{validation_total} passed")
    
    # Overall summary
    overall_passed = passed + validation_passed
    overall_total = total + validation_total
    
    print(f"\n{'='*60}")
    print("OVERALL SUMMARY")
    print(f"{'='*60}")
    print(f"Total Tests: {overall_passed}/{overall_total} passed")
    
    print(f"\nTest completed at: {datetime.now()}")
    
    # Determine if critical functionality is working
    critical_failures = 0
    if not test_results.get('admin_login', (False, None))[0]:
        critical_failures += 1
        print("❌ CRITICAL: Admin login failed")
    if not test_results.get('enhanced_match_state', (False, None))[0]:
        critical_failures += 1
        print("❌ CRITICAL: Enhanced Match State Management failed")
    if not test_results.get('complete_match_prizes', (False, None))[0]:
        critical_failures += 1
        print("❌ CRITICAL: Complete Match with Prize Distribution failed")
    
    if critical_failures > 0:
        print(f"\n❌ {critical_failures} critical test(s) failed - Manual Scoring System has issues")
        sys.exit(1)
    else:
        print(f"\n✅ All critical tests passed - Manual Scoring System is working properly")
        sys.exit(0)

if __name__ == "__main__":
    main()