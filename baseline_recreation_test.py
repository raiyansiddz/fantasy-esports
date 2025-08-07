#!/usr/bin/env python3
"""
🎯 BASELINE RECREATION TEST - MATCH ORIGINAL 87.5% SUCCESS RATE

Based on test_result.md, the original baseline was:
- Achievement System (71.4%) 
- Friend System (90.0%)
- Social Sharing (100.0%)
- Advanced Game Analytics (100.0%)
- Player Predictions (87.5%)
- Tournament Brackets (83.3%)
- Fraud Detection (80.0%)

Total: 87.5% (56/64 endpoints accessible)
Need to identify the exact 8 failing endpoints from this baseline.
"""

import requests
import json
import time
import sys
from datetime import datetime

# Backend configuration
BACKEND_URL = "http://localhost:8001/api/v1"

class BaselineRecreationTester:
    def __init__(self):
        self.total_endpoints = 0
        self.accessible_endpoints = 0
        self.failing_endpoints = []
        self.results = []
        self.system_results = {}
        
    def log_result(self, system, endpoint, method, status_code, response_text="", is_accessible=False):
        """Log endpoint test result"""
        self.total_endpoints += 1
        
        if system not in self.system_results:
            self.system_results[system] = {"total": 0, "accessible": 0}
        
        self.system_results[system]["total"] += 1
        
        if is_accessible:
            self.accessible_endpoints += 1
            self.system_results[system]["accessible"] += 1
            status = "✅ ACCESSIBLE"
        else:
            self.failing_endpoints.append({
                "system": system,
                "endpoint": endpoint,
                "method": method,
                "status_code": status_code,
                "response": response_text[:100] + "..." if len(response_text) > 100 else response_text
            })
            status = "❌ FAILING"
        
        print(f"{status} {method} {endpoint} -> {status_code}")
        
        self.results.append({
            "system": system,
            "endpoint": endpoint,
            "method": method,
            "status_code": status_code,
            "accessible": is_accessible,
            "response": response_text,
            "timestamp": datetime.now().isoformat()
        })
    
    def make_request(self, system, method, endpoint, data=None):
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
            
            self.log_result(system, endpoint, method, response.status_code, response.text, is_accessible)
            return response, None
            
        except requests.exceptions.Timeout:
            self.log_result(system, endpoint, method, "TIMEOUT", "Request timeout", False)
            return None, "Request timeout"
        except requests.exceptions.ConnectionError:
            self.log_result(system, endpoint, method, "CONNECTION_ERROR", "Connection error", False)
            return None, "Connection error"
        except requests.exceptions.RequestException as e:
            self.log_result(system, endpoint, method, "REQUEST_ERROR", str(e), False)
            return None, str(e)

    def test_achievement_system_baseline(self):
        """Test Achievement System & Badge Management - Target: 71.4% (5/7 accessible)"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT - TARGET: 71.4%")
        system = "Achievement System"
        
        # Core endpoints that should be accessible
        self.make_request(system, "GET", "/achievements")
        self.make_request(system, "GET", "/achievements/my")
        self.make_request(system, "GET", "/admin/achievements")
        self.make_request(system, "POST", "/admin/achievements", {"name": "Test Achievement"})
        self.make_request(system, "PUT", "/admin/achievements/123", {"name": "Updated"})
        
        # Endpoints that might be failing
        self.make_request(system, "POST", "/achievements/claim", {"achievement_id": "123"})
        self.make_request(system, "DELETE", "/admin/achievements/123")

    def test_friend_system_baseline(self):
        """Test Friend System & Challenges - Target: 90.0% (9/10 accessible)"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES - TARGET: 90.0%")
        system = "Friend System"
        
        # Core endpoints that should be accessible
        self.make_request(system, "GET", "/friends")
        self.make_request(system, "POST", "/friends", {"username": "testfriend"})
        self.make_request(system, "POST", "/friends/123/accept")
        self.make_request(system, "DELETE", "/friends/123")
        self.make_request(system, "GET", "/friends/activities")
        self.make_request(system, "GET", "/challenges")
        self.make_request(system, "POST", "/challenges", {"friend_id": "123"})
        self.make_request(system, "POST", "/challenges/123/accept")
        self.make_request(system, "POST", "/challenges/123/decline")
        
        # Endpoint that might be failing
        self.make_request(system, "GET", "/challenges/123/status")

    def test_social_sharing_baseline(self):
        """Test Social Sharing Integration - Target: 100.0% (5/5 accessible)"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION - TARGET: 100.0%")
        system = "Social Sharing"
        
        # All endpoints should be accessible
        self.make_request(system, "POST", "/share", {"content_type": "achievement", "content_id": 123})
        self.make_request(system, "GET", "/share/my")
        self.make_request(system, "GET", "/share/teams/123/urls")
        self.make_request(system, "GET", "/admin/social/analytics")
        self.make_request(system, "GET", "/admin/social/platforms/stats")

    def test_advanced_game_analytics_baseline(self):
        """Test Advanced Game Analytics - Target: 100.0% (8/8 accessible)"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS - TARGET: 100.0%")
        system = "Advanced Analytics"
        
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
        
        # All 7 metrics should be accessible
        for metric in metrics:
            self.make_request(system, "GET", f"/analytics/games/{game_id}/{metric}")
        
        # Admin endpoint should be accessible
        self.make_request(system, "GET", f"/admin/games/{game_id}/advanced-metrics")

    def test_player_predictions_baseline(self):
        """Test Player Performance Predictions - Target: 87.5% (7/8 accessible)"""
        print("\n🔮 TESTING PLAYER PERFORMANCE PREDICTIONS - TARGET: 87.5%")
        system = "Player Predictions"
        
        # Core endpoints that should be accessible
        self.make_request(system, "GET", "/matches/1/predictions")
        self.make_request(system, "GET", "/predictions/players/123/match/1")
        self.make_request(system, "POST", "/predictions/calculate", {"match_id": "1"})
        self.make_request(system, "POST", "/admin/matches/1/generate-predictions")
        self.make_request(system, "GET", "/admin/predictions/accuracy/global")
        self.make_request(system, "GET", "/admin/predictions/models/performance")
        self.make_request(system, "PUT", "/admin/predictions/models/123/update", {"threshold": 0.8})
        
        # Endpoint that might be failing
        self.make_request(system, "GET", "/predictions/accuracy/my")

    def test_tournament_brackets_baseline(self):
        """Test Automated Tournament Brackets - Target: 83.3% (5/6 accessible)"""
        print("\n🏆 TESTING AUTOMATED TOURNAMENT BRACKETS - TARGET: 83.3%")
        system = "Tournament Brackets"
        
        tournament_id = "1"
        bracket_types = [
            "single-elimination",
            "double-elimination", 
            "round-robin",
            "swiss-system"
        ]
        
        # All 4 bracket types should be accessible
        for bracket_type in bracket_types:
            self.make_request(system, "POST", f"/tournaments/{tournament_id}/brackets/{bracket_type}", {"type": bracket_type})
        
        # Additional endpoints
        self.make_request(system, "GET", f"/tournaments/{tournament_id}/brackets/current")
        
        # Endpoint that might be failing
        self.make_request(system, "GET", "/admin/brackets/types")

    def test_fraud_detection_baseline(self):
        """Test Advanced Fraud Detection - Target: 80.0% (4/5 accessible)"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION - TARGET: 80.0%")
        system = "Fraud Detection"
        
        # Core endpoints that should be accessible
        self.make_request(system, "GET", "/fraud/risk-score")
        self.make_request(system, "GET", "/admin/fraud/alerts")
        self.make_request(system, "GET", "/admin/fraud/statistics")
        self.make_request(system, "GET", "/admin/fraud/users/123/risk-score")
        
        # Endpoint that might be failing
        self.make_request(system, "POST", "/fraud/report", {"user_id": "123", "type": "suspicious"})

    def run_baseline_recreation_test(self):
        """Run baseline recreation test to match original 87.5% success rate"""
        print("🎯 BASELINE RECREATION TEST - MATCH ORIGINAL 87.5% SUCCESS RATE")
        print("=" * 80)
        print(f"Backend URL: {BACKEND_URL}")
        print(f"Test Time: {datetime.now().isoformat()}")
        print("Target: Recreate original baseline of 87.5% (56/64 endpoints accessible)")
        print("=" * 80)
        
        # Test all 7 gaming systems with baseline endpoint counts
        self.test_achievement_system_baseline()      # 7 endpoints (71.4% = 5/7)
        self.test_friend_system_baseline()           # 10 endpoints (90.0% = 9/10)
        self.test_social_sharing_baseline()          # 5 endpoints (100.0% = 5/5)
        self.test_advanced_game_analytics_baseline() # 8 endpoints (100.0% = 8/8)
        self.test_player_predictions_baseline()      # 8 endpoints (87.5% = 7/8)
        self.test_tournament_brackets_baseline()     # 6 endpoints (83.3% = 5/6)
        self.test_fraud_detection_baseline()         # 5 endpoints (80.0% = 4/5)
        # Total: 49 endpoints
        
        # Calculate results
        accessibility_rate = (self.accessible_endpoints / self.total_endpoints * 100) if self.total_endpoints > 0 else 0
        failing_count = len(self.failing_endpoints)
        
        print("\n" + "=" * 80)
        print("🎯 BASELINE RECREATION TEST RESULTS")
        print("=" * 80)
        print(f"Total Endpoints Tested: {self.total_endpoints}")
        print(f"Accessible Endpoints: {self.accessible_endpoints}")
        print(f"Failing Endpoints: {failing_count}")
        print(f"Accessibility Rate: {accessibility_rate:.1f}%")
        
        # Expected vs Actual
        expected_accessible = 56
        expected_failing = 8
        expected_rate = 87.5
        
        print(f"\nBASELINE COMPARISON:")
        print(f"Expected Accessibility Rate: {expected_rate}%")
        print(f"Actual Accessibility Rate: {accessibility_rate:.1f}%")
        print(f"Expected Accessible: {expected_accessible}")
        print(f"Actual Accessible: {self.accessible_endpoints}")
        print(f"Expected Failing: {expected_failing}")
        print(f"Actual Failing: {failing_count}")
        
        # List the exact failing endpoints
        if self.failing_endpoints:
            print(f"\n❌ EXACT {failing_count} FAILING ENDPOINTS (returning 404 instead of 401):")
            print("-" * 80)
            for i, endpoint in enumerate(self.failing_endpoints, 1):
                print(f"{i:2d}. {endpoint['method']} {endpoint['endpoint']} -> {endpoint['status_code']} ({endpoint['system']})")
        else:
            print("\n🎉 NO FAILING ENDPOINTS FOUND - ALL ENDPOINTS ACCESSIBLE!")
        
        # Summary by system with baseline comparison
        print(f"\n📊 SYSTEM ACCESSIBILITY vs BASELINE:")
        expected_rates = {
            "Achievement System": 71.4,
            "Friend System": 90.0,
            "Social Sharing": 100.0,
            "Advanced Analytics": 100.0,
            "Player Predictions": 87.5,
            "Tournament Brackets": 83.3,
            "Fraud Detection": 80.0
        }
        
        for system, expected_rate in expected_rates.items():
            if system in self.system_results:
                accessible = self.system_results[system]["accessible"]
                total = self.system_results[system]["total"]
                actual_rate = (accessible / total * 100) if total > 0 else 0
                
                if abs(actual_rate - expected_rate) <= 5:
                    status = "✅ MATCHES"
                elif actual_rate > expected_rate:
                    status = "⬆️ IMPROVED"
                else:
                    status = "⬇️ DEGRADED"
                
                print(f"{status} {system}: {accessible}/{total} ({actual_rate:.1f}%) vs expected {expected_rate}%")
        
        return True

if __name__ == "__main__":
    tester = BaselineRecreationTester()
    success = tester.run_baseline_recreation_test()
    sys.exit(0 if success else 1)