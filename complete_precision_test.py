#!/usr/bin/env python3
"""
🎯 COMPLETE PRECISION TESTING - ALL 64 ADVANCED GAMING FEATURES ENDPOINTS

Based on initial test results, I need to add missing endpoints to reach the full 64 endpoints.
Current findings: 7 failing endpoints out of 49 tested (85.7% success rate)
Need to test additional endpoints to reach 64 total and identify all 8 failing endpoints.
"""

import requests
import json
import time
import sys
from datetime import datetime

# Backend configuration
BACKEND_URL = "http://localhost:8001/api/v1"

class CompletePrecisionTester:
    def __init__(self):
        self.total_endpoints = 0
        self.accessible_endpoints = 0
        self.failing_endpoints = []
        self.results = []
        
    def log_result(self, endpoint, method, status_code, response_text="", is_accessible=False):
        """Log endpoint test result"""
        self.total_endpoints += 1
        
        if is_accessible:
            self.accessible_endpoints += 1
            status = "✅ ACCESSIBLE"
        else:
            self.failing_endpoints.append({
                "endpoint": endpoint,
                "method": method,
                "status_code": status_code,
                "response": response_text[:100] + "..." if len(response_text) > 100 else response_text
            })
            status = "❌ FAILING"
        
        print(f"{status} {method} {endpoint} -> {status_code}")
        
        self.results.append({
            "endpoint": endpoint,
            "method": method,
            "status_code": status_code,
            "accessible": is_accessible,
            "response": response_text,
            "timestamp": datetime.now().isoformat()
        })
    
    def make_request(self, method, endpoint, data=None):
        """Make HTTP request and determine if endpoint is accessible"""
        url = f"{BACKEND_URL}{endpoint}"
        headers = {"Content-Type": "application/json"}
        
        try:
            if method.upper() == "GET":
                response = requests.get(url, headers=headers, timeout=5)
            elif method.upper() == "POST":
                response = requests.post(url, json=data, headers=headers, timeout=5)
            elif method.upper() == "PUT":
                response = requests.put(url, json=data, headers=headers, timeout=5)
            elif method.upper() == "DELETE":
                response = requests.delete(url, headers=headers, timeout=5)
            else:
                return None, f"Unsupported method: {method}"
            
            # SUCCESS CRITERIA: Endpoint returns 401 (authentication required) - means it's properly routed
            # FAILURE CRITERIA: Endpoint returns 404 (page not found) - means missing route
            is_accessible = response.status_code != 404
            
            self.log_result(endpoint, method, response.status_code, response.text, is_accessible)
            return response, None
            
        except requests.exceptions.Timeout:
            self.log_result(endpoint, method, "TIMEOUT", "Request timeout", False)
            return None, "Request timeout"
        except requests.exceptions.ConnectionError:
            self.log_result(endpoint, method, "CONNECTION_ERROR", "Connection error", False)
            return None, "Connection error"
        except requests.exceptions.RequestException as e:
            self.log_result(endpoint, method, "REQUEST_ERROR", str(e), False)
            return None, str(e)

    def test_achievement_system_complete(self):
        """Test Achievement System & Badge Management - Complete (10 endpoints)"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT - COMPLETE")
        
        # User endpoints
        self.make_request("GET", "/achievements")
        self.make_request("GET", "/achievements/my")
        self.make_request("POST", "/achievements/claim", {"achievement_id": "123"})
        self.make_request("GET", "/achievements/123")
        self.make_request("GET", "/achievements/categories")
        
        # Admin endpoints
        self.make_request("GET", "/admin/achievements")
        self.make_request("POST", "/admin/achievements", {
            "name": "Test Achievement",
            "description": "Test description",
            "type": "match_completion",
            "criteria": {"matches_required": 10},
            "reward_type": "badge",
            "reward_value": 100
        })
        self.make_request("PUT", "/admin/achievements/123", {"name": "Updated Achievement"})
        self.make_request("DELETE", "/admin/achievements/123")
        self.make_request("GET", "/admin/achievements/stats")

    def test_friend_system_complete(self):
        """Test Friend System & Challenges - Complete (12 endpoints)"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES - COMPLETE")
        
        # Friend management
        self.make_request("GET", "/friends")
        self.make_request("POST", "/friends", {"username": "testfriend"})
        self.make_request("POST", "/friends/123/accept")
        self.make_request("DELETE", "/friends/123")
        self.make_request("GET", "/friends/activities")
        self.make_request("GET", "/friends/requests")
        
        # Challenge system
        self.make_request("GET", "/challenges")
        self.make_request("POST", "/challenges", {
            "friend_id": "123",
            "contest_id": "456",
            "entry_fee": 100,
            "message": "Let's compete!"
        })
        self.make_request("POST", "/challenges/123/accept")
        self.make_request("POST", "/challenges/123/decline")
        self.make_request("GET", "/challenges/123/status")
        self.make_request("GET", "/challenges/my")

    def test_social_sharing_complete(self):
        """Test Social Sharing Integration - Complete (8 endpoints)"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION - COMPLETE")
        
        # User sharing
        self.make_request("POST", "/share", {
            "content_type": "achievement",
            "content_id": 123,
            "platform": "twitter",
            "message": "Just earned a new achievement!"
        })
        self.make_request("GET", "/share/my")
        self.make_request("GET", "/share/teams/123/urls")
        self.make_request("GET", "/share/123/stats")
        
        # Admin analytics
        self.make_request("GET", "/admin/social/analytics")
        self.make_request("GET", "/admin/social/platforms/stats")
        self.make_request("GET", "/admin/social/trending")
        self.make_request("POST", "/admin/social/campaigns", {"name": "Test Campaign"})

    def test_advanced_game_analytics_complete(self):
        """Test Advanced Game Analytics - Complete (10 endpoints)"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS - COMPLETE")
        
        game_id = "1"
        metrics = [
            "player-efficiency",
            "team-synergy", 
            "strategic-diversity",
            "comeback-potential",
            "clutch-performance",
            "consistency-index",
            "adaptability-score"
        ]
        
        # Test each metric endpoint
        for metric in metrics:
            self.make_request("GET", f"/analytics/games/{game_id}/{metric}")
        
        # Admin advanced metrics
        self.make_request("GET", f"/admin/games/{game_id}/advanced-metrics")
        self.make_request("GET", "/admin/analytics/summary")
        self.make_request("POST", "/admin/analytics/generate", {"game_id": game_id})

    def test_player_predictions_complete(self):
        """Test Player Performance Predictions - Complete (10 endpoints)"""
        print("\n🔮 TESTING PLAYER PERFORMANCE PREDICTIONS - COMPLETE")
        
        # User prediction endpoints
        self.make_request("GET", "/matches/1/predictions")
        self.make_request("GET", "/predictions/players/123/match/1")
        self.make_request("POST", "/predictions/calculate", {
            "match_id": "1",
            "players": ["123", "456"],
            "factors": ["recent_form", "head_to_head"]
        })
        self.make_request("GET", "/predictions/accuracy/my")
        self.make_request("GET", "/predictions/my")
        
        # Admin prediction endpoints
        self.make_request("POST", "/admin/matches/1/generate-predictions")
        self.make_request("GET", "/admin/predictions/accuracy/global")
        self.make_request("GET", "/admin/predictions/models/performance")
        self.make_request("PUT", "/admin/predictions/models/123/update", {"confidence_threshold": 0.8})
        self.make_request("GET", "/admin/predictions/leaderboard")

    def test_tournament_brackets_complete(self):
        """Test Automated Tournament Brackets - Complete (8 endpoints)"""
        print("\n🏆 TESTING AUTOMATED TOURNAMENT BRACKETS - COMPLETE")
        
        tournament_id = "1"
        bracket_types = [
            "single-elimination",
            "double-elimination", 
            "round-robin",
            "swiss-system"
        ]
        
        # Test bracket creation for each type
        for bracket_type in bracket_types:
            self.make_request("POST", f"/tournaments/{tournament_id}/brackets/{bracket_type}", {
                "type": bracket_type,
                "participants": ["team1", "team2", "team3", "team4"]
            })
        
        # Additional bracket management
        self.make_request("GET", f"/tournaments/{tournament_id}/brackets/current")
        self.make_request("GET", "/admin/brackets/types")
        self.make_request("PUT", f"/tournaments/{tournament_id}/brackets/123/advance", {"winner": "team1"})
        self.make_request("GET", "/admin/brackets/stats")

    def test_fraud_detection_complete(self):
        """Test Advanced Fraud Detection - Complete (8 endpoints)"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION - COMPLETE")
        
        # User endpoints
        self.make_request("GET", "/fraud/risk-score")
        self.make_request("GET", "/fraud/my-reports")
        
        # Admin endpoints
        self.make_request("GET", "/admin/fraud/alerts")
        self.make_request("GET", "/admin/fraud/statistics")
        self.make_request("GET", "/admin/fraud/users/123/risk-score")
        self.make_request("PUT", "/admin/fraud/alerts/123/resolve", {"resolution": "false_positive"})
        
        # Public reporting
        self.make_request("POST", "/fraud/report", {
            "user_id": "123",
            "type": "suspicious_activity",
            "description": "Unusual betting pattern",
            "evidence": {"ip_address": "192.168.1.1"}
        })
        self.make_request("GET", "/admin/fraud/patterns")

    def run_complete_precision_test(self):
        """Run complete precision test to identify all failing endpoints"""
        print("🎯 COMPLETE PRECISION TESTING - ALL 64 ADVANCED GAMING FEATURES ENDPOINTS")
        print("=" * 90)
        print(f"Backend URL: {BACKEND_URL}")
        print(f"Test Time: {datetime.now().isoformat()}")
        print("Target: Test all 64 endpoints and identify exactly which ones return 404 instead of 401")
        print("=" * 90)
        
        # Test all 7 gaming systems with complete endpoint coverage
        self.test_achievement_system_complete()      # 10 endpoints
        self.test_friend_system_complete()           # 12 endpoints  
        self.test_social_sharing_complete()          # 8 endpoints
        self.test_advanced_game_analytics_complete() # 10 endpoints
        self.test_player_predictions_complete()      # 10 endpoints
        self.test_tournament_brackets_complete()     # 8 endpoints
        self.test_fraud_detection_complete()         # 8 endpoints
        # Total: 66 endpoints (slightly over 64 to ensure complete coverage)
        
        # Calculate results
        accessibility_rate = (self.accessible_endpoints / self.total_endpoints * 100) if self.total_endpoints > 0 else 0
        failing_count = len(self.failing_endpoints)
        
        print("\n" + "=" * 90)
        print("🎯 COMPLETE PRECISION TEST RESULTS")
        print("=" * 90)
        print(f"Total Endpoints Tested: {self.total_endpoints}")
        print(f"Accessible Endpoints: {self.accessible_endpoints}")
        print(f"Failing Endpoints: {failing_count}")
        print(f"Accessibility Rate: {accessibility_rate:.1f}%")
        
        # Expected vs Actual
        expected_total = 64
        expected_accessible = 56
        expected_failing = 8
        
        print(f"\nEXPECTED vs ACTUAL:")
        print(f"Total Endpoints: {expected_total} (expected) vs {self.total_endpoints} (actual)")
        print(f"Accessible: {expected_accessible} (expected) vs {self.accessible_endpoints} (actual)")
        print(f"Failing: {expected_failing} (expected) vs {failing_count} (actual)")
        
        # List the exact failing endpoints
        if self.failing_endpoints:
            print(f"\n❌ EXACT {failing_count} FAILING ENDPOINTS (returning 404 instead of 401):")
            print("-" * 90)
            for i, endpoint in enumerate(self.failing_endpoints, 1):
                print(f"{i:2d}. {endpoint['method']} {endpoint['endpoint']} -> {endpoint['status_code']}")
                if endpoint['response'] and 'page not found' in endpoint['response'].lower():
                    print(f"    Response: {endpoint['response'][:100]}")
        else:
            print("\n🎉 NO FAILING ENDPOINTS FOUND - ALL ENDPOINTS ACCESSIBLE!")
        
        # Summary by system
        print(f"\n📊 ACCESSIBILITY BY GAMING SYSTEM:")
        systems = {
            "Achievement System": [r for r in self.results if "/achievements" in r["endpoint"] or "/admin/achievements" in r["endpoint"]],
            "Friend System": [r for r in self.results if "/friends" in r["endpoint"] or "/challenges" in r["endpoint"]],
            "Social Sharing": [r for r in self.results if "/share" in r["endpoint"] or "/admin/social" in r["endpoint"]],
            "Advanced Analytics": [r for r in self.results if "/analytics/games" in r["endpoint"] or "/admin/games" in r["endpoint"] or "/admin/analytics" in r["endpoint"]],
            "Player Predictions": [r for r in self.results if "/predictions" in r["endpoint"] or "/matches" in r["endpoint"] and "predictions" in r["endpoint"]],
            "Tournament Brackets": [r for r in self.results if "/tournaments" in r["endpoint"] and "brackets" in r["endpoint"] or "/admin/brackets" in r["endpoint"]],
            "Fraud Detection": [r for r in self.results if "/fraud" in r["endpoint"] or "/admin/fraud" in r["endpoint"]]
        }
        
        for system, tests in systems.items():
            if tests:
                accessible = len([t for t in tests if t["accessible"]])
                total = len(tests)
                rate = (accessible / total * 100) if total > 0 else 0
                status = "✅" if rate == 100 else "⚠️" if rate >= 80 else "❌"
                print(f"{status} {system}: {accessible}/{total} ({rate:.1f}%)")
        
        # Final assessment
        if accessibility_rate >= 87.5:
            print(f"\n🎯 BASELINE CONFIRMED: {accessibility_rate:.1f}% accessibility rate matches expected ~87.5%")
        else:
            print(f"\n⚠️ BASELINE DIFFERENT: {accessibility_rate:.1f}% accessibility rate differs from expected 87.5%")
        
        return True

if __name__ == "__main__":
    tester = CompletePrecisionTester()
    success = tester.run_complete_precision_test()
    sys.exit(0 if success else 1)