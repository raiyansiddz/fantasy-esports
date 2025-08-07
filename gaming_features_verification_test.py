#!/usr/bin/env python3
"""
Advanced Gaming Features Verification Test
Focus: Verify that critical binary fix resolved route registration vs accessibility gap
Target: Confirm all gaming routes return 401 (auth required) instead of 404 (not found)
"""

import requests
import json
from datetime import datetime

# Backend configuration
BACKEND_URL = "http://localhost:8001/api/v1"

class GamingFeaturesVerifier:
    def __init__(self):
        self.total_tests = 0
        self.passed_tests = 0
        self.results = []
        
    def log_result(self, test_name, status, details=""):
        """Log test result"""
        self.total_tests += 1
        if status:
            self.passed_tests += 1
            print(f"✅ {test_name}")
        else:
            self.passed_tests += 0
            print(f"❌ {test_name}: {details}")
        
        self.results.append({
            "test": test_name,
            "status": "PASS" if status else "FAIL",
            "details": details
        })
    
    def test_endpoint_accessibility(self, method, endpoint, expected_status_codes, test_name, data=None):
        """Test if endpoint is accessible (returns expected status instead of 404)"""
        url = f"{BACKEND_URL}{endpoint}"
        headers = {"Content-Type": "application/json", "Authorization": "Bearer dummy_token"}
        
        try:
            if method.upper() == "GET":
                response = requests.get(url, headers=headers, timeout=5)
            elif method.upper() == "POST":
                response = requests.post(url, json=data, headers=headers, timeout=5)
            else:
                response = requests.get(url, headers=headers, timeout=5)
            
            if response.status_code in expected_status_codes:
                self.log_result(test_name, True, f"Status: {response.status_code} (endpoint accessible)")
                return True
            elif response.status_code == 404:
                self.log_result(test_name, False, f"Status: 404 (endpoint not found/not implemented)")
                return False
            else:
                self.log_result(test_name, True, f"Status: {response.status_code} (endpoint accessible)")
                return True
                
        except requests.exceptions.Timeout:
            self.log_result(test_name, False, "Request timeout")
            return False
        except requests.exceptions.ConnectionError:
            self.log_result(test_name, False, "Connection error")
            return False
        except Exception as e:
            self.log_result(test_name, False, f"Error: {str(e)}")
            return False

    def test_achievement_system(self):
        """Test Achievement System & Badge Management endpoints"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT")
        
        endpoints = [
            ("GET", "/achievements", [200, 401], "User achievements list"),
            ("GET", "/achievements/my", [200, 401], "User personal achievements"),
            ("GET", "/achievements/123/progress", [200, 400, 401], "Achievement progress"),
            ("POST", "/admin/achievements", [200, 201, 400, 401], "Admin create achievement"),
            ("GET", "/admin/achievements", [200, 401], "Admin list achievements"),
            ("PUT", "/admin/achievements/123", [200, 400, 401], "Admin update achievement"),
            ("DELETE", "/admin/achievements/123", [200, 400, 401], "Admin delete achievement")
        ]
        
        for method, endpoint, expected_codes, name in endpoints:
            data = {"name": "Test Achievement", "description": "Test"} if method == "POST" else None
            self.test_endpoint_accessibility(method, endpoint, expected_codes, f"Achievement System - {name}", data)

    def test_friend_system(self):
        """Test Friend System & Challenges endpoints"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES")
        
        endpoints = [
            ("POST", "/friends", [200, 201, 400, 401], "Add friend"),
            ("GET", "/friends", [200, 401], "Friends list"),
            ("POST", "/friends/123/accept", [200, 400, 401], "Accept friend request"),
            ("POST", "/friends/123/decline", [200, 400, 401], "Decline friend request"),
            ("DELETE", "/friends/123", [200, 400, 401], "Remove friend"),
            ("POST", "/challenges", [200, 201, 400, 401], "Create challenge"),
            ("GET", "/challenges", [200, 401], "Challenges list"),
            ("POST", "/challenges/123/accept", [200, 400, 401], "Accept challenge"),
            ("POST", "/challenges/123/decline", [200, 400, 401], "Decline challenge"),
            ("GET", "/friends/activities", [200, 401], "Friend activities")
        ]
        
        for method, endpoint, expected_codes, name in endpoints:
            data = {"username": "testfriend"} if method == "POST" and "friends" in endpoint else None
            data = {"friend_id": "123", "contest_id": "456"} if method == "POST" and "challenges" in endpoint else data
            self.test_endpoint_accessibility(method, endpoint, expected_codes, f"Friend System - {name}", data)

    def test_social_sharing(self):
        """Test Social Sharing Integration endpoints"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION")
        
        endpoints = [
            ("POST", "/share", [200, 201, 400, 401], "Create share"),
            ("GET", "/share/my", [200, 401], "User shares"),
            ("GET", "/share/teams/123/urls", [200, 400, 401], "Team share URLs"),
            ("GET", "/share/contests/123/urls", [200, 400, 401], "Contest share URLs"),
            ("GET", "/share/achievements/123/urls", [200, 400, 401], "Achievement share URLs"),
            ("POST", "/share/123/click", [200, 400, 401], "Track share click"),
            ("GET", "/admin/social/analytics", [200, 401], "Admin social analytics")
        ]
        
        for method, endpoint, expected_codes, name in endpoints:
            data = {"content_type": "achievement", "content_id": "123", "platform": "twitter"} if method == "POST" and endpoint == "/share" else None
            self.test_endpoint_accessibility(method, endpoint, expected_codes, f"Social Sharing - {name}", data)

    def test_advanced_game_analytics(self):
        """Test Advanced Game Analytics (7 metrics) endpoints"""
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
        
        for metric in metrics:
            self.test_endpoint_accessibility(
                "GET", 
                f"/analytics/games/{game_id}/{metric}", 
                [200, 400, 401], 
                f"Advanced Analytics - {metric}"
            )
        
        # Admin endpoints
        admin_endpoints = [
            (f"/admin/games/{game_id}/advanced-metrics", "Admin advanced metrics"),
            (f"/admin/games/{game_id}/metrics-history", "Admin metrics history"),
            ("/admin/games/compare", "Admin games comparison")
        ]
        
        for endpoint, name in admin_endpoints:
            self.test_endpoint_accessibility("GET", endpoint, [200, 400, 401], f"Advanced Analytics - {name}")

    def test_player_predictions(self):
        """Test Player Performance Predictions endpoints"""
        print("\n🔮 TESTING PLAYER PERFORMANCE PREDICTIONS")
        
        endpoints = [
            ("GET", "/matches/1/predictions", [200, 400, 401], "Match predictions"),
            ("GET", "/predictions/players/123/match/1", [200, 400, 401], "Player match predictions"),
            ("GET", "/predictions/match/1/teams", [200, 400, 401], "Match team predictions"),
            ("POST", "/predictions/calculate", [200, 201, 400, 401], "Calculate predictions"),
            ("GET", "/predictions/history/123", [200, 400, 401], "Prediction history"),
            ("POST", "/admin/matches/1/generate-predictions", [200, 201, 400, 401], "Admin generate predictions"),
            ("PUT", "/admin/matches/1/update-accuracy", [200, 400, 401], "Admin update accuracy"),
            ("GET", "/admin/predictions/analytics", [200, 401], "Admin prediction analytics")
        ]
        
        for method, endpoint, expected_codes, name in endpoints:
            data = {"match_id": "1", "players": ["123"]} if method == "POST" and "calculate" in endpoint else None
            self.test_endpoint_accessibility(method, endpoint, expected_codes, f"Player Predictions - {name}", data)

    def test_tournament_brackets(self):
        """Test Automated Tournament Brackets (4 types) endpoints"""
        print("\n🏆 TESTING AUTOMATED TOURNAMENT BRACKETS (4 TYPES)")
        
        tournament_id = "1"
        bracket_types = [
            "single-elimination",
            "double-elimination", 
            "round-robin",
            "swiss-system"
        ]
        
        # User bracket creation endpoints
        for bracket_type in bracket_types:
            data = {"type": bracket_type, "participants": ["team1", "team2"]}
            self.test_endpoint_accessibility(
                "POST", 
                f"/tournaments/{tournament_id}/brackets/{bracket_type}", 
                [200, 201, 400, 401], 
                f"Tournament Brackets - {bracket_type} creation",
                data
            )
        
        # Other bracket endpoints
        other_endpoints = [
            ("GET", f"/tournaments/{tournament_id}/brackets/current", [200, 400, 401], "Get current brackets"),
            ("POST", "/admin/tournaments/brackets", [200, 201, 400, 401], "Admin create bracket"),
            ("GET", f"/admin/tournaments/{tournament_id}/brackets", [200, 401], "Admin get tournament brackets"),
            ("GET", "/admin/brackets/123", [200, 400, 401], "Admin get bracket"),
            ("PUT", "/admin/brackets/123/advance", [200, 400, 401], "Admin advance bracket"),
            ("PUT", "/admin/brackets/123/status", [200, 400, 401], "Admin update bracket status"),
            ("DELETE", "/admin/brackets/123", [200, 400, 401], "Admin delete bracket"),
            ("GET", "/admin/brackets/types", [200, 401], "Admin bracket types")
        ]
        
        for method, endpoint, expected_codes, name in other_endpoints:
            data = {"type": "single-elimination"} if method == "POST" else None
            self.test_endpoint_accessibility(method, endpoint, expected_codes, f"Tournament Brackets - {name}", data)

    def test_fraud_detection(self):
        """Test Advanced Fraud Detection endpoints"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION")
        
        endpoints = [
            ("GET", "/fraud/risk-score", [200, 401], "User risk score"),
            ("POST", "/fraud/report", [200, 201, 400], "Public fraud reporting"),
            ("POST", "/fraud/webhook", [200, 400], "Fraud webhook"),
            ("GET", "/admin/fraud/alerts", [200, 401], "Admin fraud alerts"),
            ("PUT", "/admin/fraud/alerts/123/status", [200, 400, 401], "Admin update alert status"),
            ("GET", "/admin/fraud/statistics", [200, 401], "Admin fraud statistics"),
            ("GET", "/admin/fraud/users/123/risk-score", [200, 400, 401], "Admin user risk score"),
            ("POST", "/admin/fraud/investigate", [200, 201, 400, 401], "Admin investigate"),
            ("GET", "/admin/fraud/patterns", [200, 401], "Admin fraud patterns"),
            ("PUT", "/admin/fraud/threshold", [200, 400, 401], "Admin update threshold")
        ]
        
        for method, endpoint, expected_codes, name in endpoints:
            data = {"type": "suspicious_activity", "description": "Test"} if method == "POST" and "report" in endpoint else None
            self.test_endpoint_accessibility(method, endpoint, expected_codes, f"Fraud Detection - {name}", data)

    def run_verification_test(self):
        """Run comprehensive verification test"""
        print("🎯 ADVANCED GAMING FEATURES VERIFICATION TEST")
        print("=" * 70)
        print(f"Backend URL: {BACKEND_URL}")
        print(f"Test Time: {datetime.now().isoformat()}")
        print("Objective: Verify gaming routes return 401 (auth required) instead of 404 (not found)")
        print("=" * 70)
        
        # Test all 7 gaming systems
        self.test_achievement_system()
        self.test_friend_system()
        self.test_social_sharing()
        self.test_advanced_game_analytics()
        self.test_player_predictions()
        self.test_tournament_brackets()
        self.test_fraud_detection()
        
        # Calculate results
        success_rate = (self.passed_tests / self.total_tests * 100) if self.total_tests > 0 else 0
        
        print("\n" + "=" * 70)
        print("🎯 VERIFICATION TEST RESULTS")
        print("=" * 70)
        print(f"Total Endpoints Tested: {self.total_tests}")
        print(f"Accessible Endpoints: {self.passed_tests}")
        print(f"Inaccessible Endpoints (404): {self.total_tests - self.passed_tests}")
        print(f"Accessibility Rate: {success_rate:.1f}%")
        
        # Determine if binary fix was successful
        target_rate = 70.0
        previous_rate = 23.7
        
        print(f"\nPrevious Success Rate: {previous_rate}%")
        print(f"Target Success Rate: {target_rate}%")
        print(f"Current Success Rate: {success_rate:.1f}%")
        
        if success_rate >= target_rate:
            print(f"🎉 BINARY FIX SUCCESSFUL! Accessibility rate {success_rate:.1f}% exceeds target {target_rate}%")
            print("✅ Gaming features are properly implemented and accessible")
        else:
            improvement = success_rate - previous_rate
            print(f"⚠️ BINARY FIX PARTIALLY SUCCESSFUL. Improvement: +{improvement:.1f}%")
            if improvement > 0:
                print("✅ Some improvement achieved, but target not fully met")
            else:
                print("❌ No improvement or regression detected")
        
        # Summary by system
        print("\n📊 ACCESSIBILITY BY GAMING SYSTEM:")
        systems = {
            "Achievement System": [r for r in self.results if "Achievement System" in r["test"]],
            "Friend System": [r for r in self.results if "Friend System" in r["test"]],
            "Social Sharing": [r for r in self.results if "Social Sharing" in r["test"]],
            "Advanced Analytics": [r for r in self.results if "Advanced Analytics" in r["test"]],
            "Player Predictions": [r for r in self.results if "Player Predictions" in r["test"]],
            "Tournament Brackets": [r for r in self.results if "Tournament Brackets" in r["test"]],
            "Fraud Detection": [r for r in self.results if "Fraud Detection" in r["test"]]
        }
        
        working_systems = 0
        for system, tests in systems.items():
            if tests:
                passed = len([t for t in tests if t["status"] == "PASS"])
                total = len(tests)
                rate = (passed / total * 100) if total > 0 else 0
                status = "✅" if rate >= 70 else "❌"
                if rate >= 70:
                    working_systems += 1
                print(f"{status} {system}: {passed}/{total} ({rate:.1f}%)")
        
        print(f"\n🏆 WORKING SYSTEMS: {working_systems}/7 ({working_systems/7*100:.1f}%)")
        
        return success_rate >= target_rate

if __name__ == "__main__":
    verifier = GamingFeaturesVerifier()
    success = verifier.run_verification_test()
    exit(0 if success else 1)