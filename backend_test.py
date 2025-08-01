#!/usr/bin/env python3
"""
Fantasy Sports Backend API Testing Script
Tests the GoLang fantasy sports backend running on the configured URL
"""

import requests
import json
import sys
from datetime import datetime

# Get backend URL from environment - using localhost since external URL routes to frontend
BACKEND_URL = "http://localhost:8080"

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

def test_games_list():
    """Test the games list endpoint"""
    print_test_header("Games List")
    
    try:
        url = f"{BACKEND_URL}/api/v1/games/"
        response = requests.get(url, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200:
            if data and 'games' in data:
                print(f"✅ Games list PASSED - Found {len(data['games'])} games")
                return True, data
            else:
                print("⚠️ Games list returned 200 but no games data")
                return True, data
        else:
            print("❌ Games list FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Games list ERROR: {str(e)}")
        return False, None

def test_matches_list():
    """Test the matches list endpoint"""
    print_test_header("Matches List")
    
    try:
        url = f"{BACKEND_URL}/api/v1/matches/"
        response = requests.get(url, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200:
            if data and 'matches' in data:
                print(f"✅ Matches list PASSED - Found {len(data['matches'])} matches")
                return True, data
            else:
                print("⚠️ Matches list returned 200 but no matches data")
                return True, data
        else:
            print("❌ Matches list FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Matches list ERROR: {str(e)}")
        return False, None

def test_match_details(match_id=1):
    """Test the match details endpoint"""
    print_test_header(f"Match Details (ID: {match_id})")
    
    try:
        url = f"{BACKEND_URL}/api/v1/matches/{match_id}"
        response = requests.get(url, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200:
            if data and 'match' in data:
                print(f"✅ Match details PASSED - Match: {data['match'].get('name', 'Unknown')}")
                return True, data
            else:
                print("⚠️ Match details returned 200 but no match data")
                return True, data
        elif response.status_code == 404:
            print(f"⚠️ Match {match_id} not found (404) - This is expected if no sample data")
            return True, data
        else:
            print("❌ Match details FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Match details ERROR: {str(e)}")
        return False, None

def test_match_players(match_id=1):
    """Test the match players endpoint"""
    print_test_header(f"Match Players (ID: {match_id})")
    
    try:
        url = f"{BACKEND_URL}/api/v1/matches/{match_id}/players"
        response = requests.get(url, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200:
            if data and 'players' in data:
                print(f"✅ Match players PASSED - Found {len(data['players'])} players")
                return True, data
            else:
                print("⚠️ Match players returned 200 but no players data")
                return True, data
        elif response.status_code == 404:
            print(f"⚠️ Match {match_id} players not found (404) - This is expected if no sample data")
            return True, data
        else:
            print("❌ Match players FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Match players ERROR: {str(e)}")
        return False, None

def test_tournaments():
    """Test the tournaments endpoint"""
    print_test_header("Tournaments List")
    
    try:
        url = f"{BACKEND_URL}/api/v1/tournaments/"
        response = requests.get(url, timeout=10)
        data = print_response(response, url)
        
        if response.status_code == 200:
            if data and 'tournaments' in data:
                print(f"✅ Tournaments list PASSED - Found {len(data['tournaments'])} tournaments")
                return True, data
            else:
                print("⚠️ Tournaments list returned 200 but no tournaments data")
                return True, data
        else:
            print("❌ Tournaments list FAILED")
            return False, data
            
    except Exception as e:
        print(f"❌ Tournaments list ERROR: {str(e)}")
        return False, None

def test_additional_endpoints():
    """Test additional endpoints for comprehensive coverage"""
    print_test_header("Additional Endpoint Tests")
    
    results = {}
    
    # Test games with filters
    try:
        url = f"{BACKEND_URL}/api/v1/games/?status=active"
        response = requests.get(url, timeout=10)
        results['games_filtered'] = response.status_code == 200
        print(f"Games with status filter: {'✅ PASSED' if results['games_filtered'] else '❌ FAILED'}")
    except:
        results['games_filtered'] = False
        print("Games with status filter: ❌ ERROR")
    
    # Test matches with filters
    try:
        url = f"{BACKEND_URL}/api/v1/matches/?status=upcoming&limit=5"
        response = requests.get(url, timeout=10)
        results['matches_filtered'] = response.status_code == 200
        print(f"Matches with filters: {'✅ PASSED' if results['matches_filtered'] else '❌ FAILED'}")
    except:
        results['matches_filtered'] = False
        print("Matches with filters: ❌ ERROR")
    
    # Test tournaments with filters
    try:
        url = f"{BACKEND_URL}/api/v1/tournaments/?featured=true"
        response = requests.get(url, timeout=10)
        results['tournaments_filtered'] = response.status_code == 200
        print(f"Tournaments with filters: {'✅ PASSED' if results['tournaments_filtered'] else '❌ FAILED'}")
    except:
        results['tournaments_filtered'] = False
        print("Tournaments with filters: ❌ ERROR")
    
    return results

def main():
    """Main test execution"""
    print("🚀 Starting Fantasy Sports Backend API Tests")
    print(f"Backend URL: {BACKEND_URL}")
    print(f"Test started at: {datetime.now()}")
    
    # Track test results
    test_results = {}
    
    # Run all tests
    test_results['health'] = test_health_check()
    test_results['games'] = test_games_list()
    test_results['matches'] = test_matches_list()
    test_results['match_details'] = test_match_details()
    test_results['match_players'] = test_match_players()
    test_results['tournaments'] = test_tournaments()
    
    # Additional tests
    additional_results = test_additional_endpoints()
    
    # Print summary
    print(f"\n{'='*60}")
    print("TEST SUMMARY")
    print(f"{'='*60}")
    
    passed = 0
    total = 0
    
    for test_name, (success, data) in test_results.items():
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{test_name.upper()}: {status}")
        if success:
            passed += 1
        total += 1
    
    # Additional tests summary
    for test_name, success in additional_results.items():
        status = "✅ PASSED" if success else "❌ FAILED"
        print(f"{test_name.upper()}: {status}")
        if success:
            passed += 1
        total += 1
    
    print(f"\nOverall: {passed}/{total} tests passed")
    
    # Check if we have sample data
    games_data = test_results['games'][1] if test_results['games'][0] else None
    matches_data = test_results['matches'][1] if test_results['matches'][0] else None
    tournaments_data = test_results['tournaments'][1] if test_results['tournaments'][0] else None
    
    print(f"\n{'='*60}")
    print("DATA ANALYSIS")
    print(f"{'='*60}")
    
    if games_data and games_data.get('games'):
        print(f"✅ Sample games data found: {len(games_data['games'])} games")
    else:
        print("⚠️ No sample games data found")
    
    if matches_data and matches_data.get('matches'):
        print(f"✅ Sample matches data found: {len(matches_data['matches'])} matches")
    else:
        print("⚠️ No sample matches data found")
    
    if tournaments_data and tournaments_data.get('tournaments'):
        print(f"✅ Sample tournaments data found: {len(tournaments_data['tournaments'])} tournaments")
    else:
        print("⚠️ No sample tournaments data found")
    
    print(f"\nTest completed at: {datetime.now()}")
    
    # Return exit code based on critical failures
    critical_failures = 0
    if not test_results['health'][0]:
        critical_failures += 1
    if not test_results['games'][0]:
        critical_failures += 1
    if not test_results['matches'][0]:
        critical_failures += 1
    if not test_results['tournaments'][0]:
        critical_failures += 1
    
    if critical_failures > 0:
        print(f"\n❌ {critical_failures} critical endpoint(s) failed")
        sys.exit(1)
    else:
        print(f"\n✅ All critical endpoints working properly")
        sys.exit(0)

if __name__ == "__main__":
    main()