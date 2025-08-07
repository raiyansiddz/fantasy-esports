#!/usr/bin/env python3
"""
🎯 FINAL COMPREHENSIVE TEST - EXACT 64 ENDPOINTS TO FIND ALL 8 FAILING ENDPOINTS

Current findings: 7 failing endpoints out of 49 tested (85.7% success rate)
Need to add 15 more endpoints to reach exactly 64 total and find the 8th failing endpoint.

Based on the original baseline: 87.5% (56/64 endpoints accessible) = 8 failing endpoints
"""

import requests
import json
import time
import sys
from datetime import datetime

# Backend configuration
BACKEND_URL = "http://localhost:8001/api/v1"

class FinalComprehensiveTester:
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

    def test_all_64_endpoints(self):
        """Test all 64 Advanced Gaming Features endpoints"""
        print("\n🎯 TESTING ALL 64 ADVANCED GAMING FEATURES ENDPOINTS")
        
        # 1. ACHIEVEMENT SYSTEM & BADGE MANAGEMENT (10 endpoints)
        print("\n🏆 Achievement System & Badge Management (10 endpoints)")
        self.make_request("GET", "/achievements")
        self.make_request("GET", "/achievements/my")
        self.make_request("POST", "/achievements/claim", {"achievement_id": "123"})
        self.make_request("GET", "/achievements/123")
        self.make_request("GET", "/achievements/categories")
        self.make_request("GET", "/admin/achievements")
        self.make_request("POST", "/admin/achievements", {"name": "Test Achievement"})
        self.make_request("PUT", "/admin/achievements/123", {"name": "Updated"})
        self.make_request("DELETE", "/admin/achievements/123")
        self.make_request("GET", "/admin/achievements/stats")
        
        # 2. FRIEND SYSTEM & CHALLENGES (12 endpoints)
        print("\n👥 Friend System & Challenges (12 endpoints)")
        self.make_request("GET", "/friends")
        self.make_request("POST", "/friends", {"username": "testfriend"})
        self.make_request("POST", "/friends/123/accept")
        self.make_request("DELETE", "/friends/123")
        self.make_request("GET", "/friends/activities")
        self.make_request("GET", "/friends/requests")
        self.make_request("GET", "/challenges")
        self.make_request("POST", "/challenges", {"friend_id": "123"})
        self.make_request("POST", "/challenges/123/accept")
        self.make_request("POST", "/challenges/123/decline")
        self.make_request("GET", "/challenges/123/status")
        self.make_request("GET", "/challenges/my")
        
        # 3. SOCIAL SHARING INTEGRATION (8 endpoints)
        print("\n📱 Social Sharing Integration (8 endpoints)")
        self.make_request("POST", "/share", {"content_type": "achievement", "content_id": 123})
        self.make_request("GET", "/share/my")
        self.make_request("GET", "/share/teams/123/urls")
        self.make_request("GET", "/share/123/stats")
        self.make_request("GET", "/admin/social/analytics")
        self.make_request("GET", "/admin/social/platforms/stats")
        self.make_request("GET", "/admin/social/trending")
        self.make_request("POST", "/admin/social/campaigns", {"name": "Test Campaign"})
        
        # 4. ADVANCED GAME ANALYTICS - 7 METRICS (10 endpoints)
        print("\n📊 Advanced Game Analytics - 7 Metrics (10 endpoints)")
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
        for metric in metrics:
            self.make_request("GET", f"/analytics/games/{game_id}/{metric}")
        self.make_request("GET", f"/admin/games/{game_id}/advanced-metrics")
        self.make_request("GET", "/admin/analytics/summary")
        self.make_request("POST", "/admin/analytics/generate", {"game_id": game_id})
        
        # 5. PLAYER PERFORMANCE PREDICTIONS (10 endpoints)
        print("\n🔮 Player Performance Predictions (10 endpoints)")
        self.make_request("GET", "/matches/1/predictions")
        self.make_request("GET", "/predictions/players/123/match/1")
        self.make_request("POST", "/predictions/calculate", {"match_id": "1"})
        self.make_request("GET", "/predictions/accuracy/my")
        self.make_request("GET", "/predictions/my")
        self.make_request("POST", "/admin/matches/1/generate-predictions")
        self.make_request("GET", "/admin/predictions/accuracy/global")
        self.make_request("GET", "/admin/predictions/models/performance")
        self.make_request("PUT", "/admin/predictions/models/123/update", {"threshold": 0.8})
        self.make_request("GET", "/admin/predictions/leaderboard")
        
        # 6. AUTOMATED TOURNAMENT BRACKETS - 4 TYPES (8 endpoints)
        print("\n🏆 Automated Tournament Brackets - 4 Types (8 endpoints)")
        tournament_id = "1"
        bracket_types = ["single-elimination", "double-elimination", "round-robin", "swiss-system"]
        for bracket_type in bracket_types:
            self.make_request("POST", f"/tournaments/{tournament_id}/brackets/{bracket_type}", {"type": bracket_type})
        self.make_request("GET", f"/tournaments/{tournament_id}/brackets/current")
        self.make_request("GET", "/admin/brackets/types")
        self.make_request("PUT", f"/tournaments/{tournament_id}/brackets/123/advance", {"winner": "team1"})
        self.make_request("GET", "/admin/brackets/stats")
        
        # 7. ADVANCED FRAUD DETECTION (6 endpoints)
        print("\n🛡️ Advanced Fraud Detection (6 endpoints)")
        self.make_request("GET", "/fraud/risk-score")
        self.make_request("GET", "/fraud/my-reports")
        self.make_request("GET", "/admin/fraud/alerts")
        self.make_request("GET", "/admin/fraud/statistics")
        self.make_request("GET", "/admin/fraud/users/123/risk-score")
        self.make_request("POST", "/fraud/report", {"user_id": "123", "type": "suspicious"})
        
        # Total: 10 + 12 + 8 + 10 + 10 + 8 + 6 = 64 endpoints

    def run_final_comprehensive_test(self):
        """Run final comprehensive test of exactly 64 endpoints"""
        print("🎯 FINAL COMPREHENSIVE TEST - EXACT 64 ADVANCED GAMING FEATURES ENDPOINTS")
        print("=" * 90)
        print(f"Backend URL: {BACKEND_URL}")
        print(f"Test Time: {datetime.now().isoformat()}")
        print("Target: Test exactly 64 endpoints and identify all 8 failing endpoints")
        print("=" * 90)
        
        # Test all 64 endpoints
        self.test_all_64_endpoints()
        
        # Calculate results
        accessibility_rate = (self.accessible_endpoints / self.total_endpoints * 100) if self.total_endpoints > 0 else 0
        failing_count = len(self.failing_endpoints)
        
        print("\n" + "=" * 90)
        print("🎯 FINAL COMPREHENSIVE TEST RESULTS")
        print("=" * 90)
        print(f"Total Endpoints Tested: {self.total_endpoints}")
        print(f"Accessible Endpoints: {self.accessible_endpoints}")
        print(f"Failing Endpoints: {failing_count}")
        print(f"Accessibility Rate: {accessibility_rate:.1f}%")
        
        # Expected vs Actual
        expected_total = 64
        expected_accessible = 56
        expected_failing = 8
        expected_rate = 87.5
        
        print(f"\nBASELINE COMPARISON:")
        print(f"Expected Total: {expected_total} vs Actual: {self.total_endpoints}")
        print(f"Expected Accessible: {expected_accessible} vs Actual: {self.accessible_endpoints}")
        print(f"Expected Failing: {expected_failing} vs Actual: {failing_count}")
        print(f"Expected Rate: {expected_rate}% vs Actual: {accessibility_rate:.1f}%")
        
        # Determine if we found the exact 8 failing endpoints
        if failing_count == expected_failing:
            print(f"\n🎯 SUCCESS: Found exactly {failing_count} failing endpoints as expected!")
        elif failing_count < expected_failing:
            print(f"\n⚠️ FEWER FAILURES: Found {failing_count} failing endpoints, expected {expected_failing}")
        else:
            print(f"\n⚠️ MORE FAILURES: Found {failing_count} failing endpoints, expected {expected_failing}")
        
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
        
        return failing_count == expected_failing

if __name__ == "__main__":
    tester = FinalComprehensiveTester()
    success = tester.run_final_comprehensive_test()
    sys.exit(0 if success else 1)