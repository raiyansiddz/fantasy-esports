#!/usr/bin/env python3
"""
🎯 PRECISION TESTING - IDENTIFY THE REMAINING 8 FAILING ENDPOINTS FOR 100% SUCCESS

OBJECTIVE: Test all 64 Advanced Gaming Features endpoints to identify exactly which 8 endpoints 
are still returning 404 (not found) instead of 401 (auth required), so we can fix them to achieve 100% success rate.

CURRENT STATUS: 
- Previous success rate: 87.5% (56/64 endpoints accessible)
- Target: 100% (64/64 endpoints accessible)
- Need to fix: 8 remaining endpoints (12.5%)

TESTING SCOPE - ALL 7 ADVANCED GAMING SYSTEMS:
1. Achievement System & Badge Management
2. Friend System & Challenges  
3. Social Sharing Integration
4. Advanced Game Analytics (7 metrics)
5. Player Performance Predictions
6. Automated Tournament Brackets (4 types)
7. Advanced Fraud Detection
"""

import requests
import json
import time
import sys
from datetime import datetime

# Backend configuration
BACKEND_URL = "http://localhost:8001/api/v1"

class PrecisionGameTester:
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

    def test_achievement_system(self):
        """Test Achievement System & Badge Management (7 endpoints)"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT")
        
        # User endpoints
        self.make_request("GET", "/achievements")
        self.make_request("GET", "/achievements/my")
        self.make_request("POST", "/achievements/claim", {"achievement_id": "123"})
        
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

    def test_friend_system(self):
        """Test Friend System & Challenges (10 endpoints)"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES")
        
        # Friend management
        self.make_request("GET", "/friends")
        self.make_request("POST", "/friends", {"username": "testfriend"})
        self.make_request("POST", "/friends/123/accept")
        self.make_request("DELETE", "/friends/123")
        self.make_request("GET", "/friends/activities")
        
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

    def test_social_sharing(self):
        """Test Social Sharing Integration (5 endpoints)"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION")
        
        # User sharing
        self.make_request("POST", "/share", {
            "content_type": "achievement",
            "content_id": 123,
            "platform": "twitter",
            "message": "Just earned a new achievement!"
        })
        self.make_request("GET", "/share/my")
        self.make_request("GET", "/share/teams/123/urls")
        
        # Admin analytics
        self.make_request("GET", "/admin/social/analytics")
        self.make_request("GET", "/admin/social/platforms/stats")

    def test_advanced_game_analytics(self):
        """Test Advanced Game Analytics - 7 metrics (8 endpoints)"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS (7 METRICS)")
        
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

    def test_player_predictions(self):
        """Test Player Performance Predictions (8 endpoints)"""
        print("\n🔮 TESTING PLAYER PERFORMANCE PREDICTIONS")
        
        # User prediction endpoints
        self.make_request("GET", "/matches/1/predictions")
        self.make_request("GET", "/predictions/players/123/match/1")
        self.make_request("POST", "/predictions/calculate", {
            "match_id": "1",
            "players": ["123", "456"],
            "factors": ["recent_form", "head_to_head"]
        })
        self.make_request("GET", "/predictions/accuracy/my")
        
        # Admin prediction endpoints
        self.make_request("POST", "/admin/matches/1/generate-predictions")
        self.make_request("GET", "/admin/predictions/accuracy/global")
        self.make_request("GET", "/admin/predictions/models/performance")
        self.make_request("PUT", "/admin/predictions/models/123/update", {"confidence_threshold": 0.8})

    def test_tournament_brackets(self):
        """Test Automated Tournament Brackets - 4 types (6 endpoints)"""
        print("\n🏆 TESTING AUTOMATED TOURNAMENT BRACKETS (4 TYPES)")
        
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

    def test_fraud_detection(self):
        """Test Advanced Fraud Detection (5 endpoints)"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION")
        
        # User endpoints
        self.make_request("GET", "/fraud/risk-score")
        
        # Admin endpoints
        self.make_request("GET", "/admin/fraud/alerts")
        self.make_request("GET", "/admin/fraud/statistics")
        self.make_request("GET", "/admin/fraud/users/123/risk-score")
        
        # Public reporting
        self.make_request("POST", "/fraud/report", {
            "user_id": "123",
            "type": "suspicious_activity",
            "description": "Unusual betting pattern",
            "evidence": {"ip_address": "192.168.1.1"}
        })

    def run_precision_test(self):
        """Run precision test to identify exact failing endpoints"""
        print("🎯 PRECISION TESTING - IDENTIFY THE REMAINING 8 FAILING ENDPOINTS")
        print("=" * 80)
        print(f"Backend URL: {BACKEND_URL}")
        print(f"Test Time: {datetime.now().isoformat()}")
        print("Target: Identify exactly which 8 endpoints return 404 instead of 401")
        print("=" * 80)
        
        # Test all 7 gaming systems
        self.test_achievement_system()      # 7 endpoints
        self.test_friend_system()           # 10 endpoints  
        self.test_social_sharing()          # 5 endpoints
        self.test_advanced_game_analytics() # 8 endpoints
        self.test_player_predictions()      # 8 endpoints
        self.test_tournament_brackets()     # 6 endpoints
        self.test_fraud_detection()         # 5 endpoints
        
        # Calculate results
        accessibility_rate = (self.accessible_endpoints / self.total_endpoints * 100) if self.total_endpoints > 0 else 0
        failing_count = len(self.failing_endpoints)
        
        print("\n" + "=" * 80)
        print("🎯 PRECISION TEST RESULTS")
        print("=" * 80)
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
            print("-" * 80)
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
            "Advanced Analytics": [r for r in self.results if "/analytics/games" in r["endpoint"] or "/admin/games" in r["endpoint"]],
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
        
        return failing_count == expected_failing

if __name__ == "__main__":
    tester = PrecisionGameTester()
    success = tester.run_precision_test()
    sys.exit(0 if success else 1)