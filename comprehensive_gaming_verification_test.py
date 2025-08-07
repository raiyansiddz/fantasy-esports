#!/usr/bin/env python3
"""
🎯 COMPREHENSIVE VERIFICATION TESTING - 7 ADVANCED GAMING FEATURES
Fantasy Esports GoLang Backend - Baseline Establishment Testing

This test suite verifies the current state of all 7 Advanced Gaming Features
after Go binary is confirmed running on http://localhost:8001 to establish
baseline before implementing fixes.

Focus Areas:
1. Achievement System & Badge Management
2. Friend System & Challenges  
3. Social Sharing Integration
4. Advanced Game Analytics (7 Metrics)
5. Player Performance Predictions
6. Automated Tournament Brackets
7. Advanced Fraud Detection

Testing with proper authentication, valid data formats, and comprehensive error analysis.
"""

import requests
import json
import time
import uuid
from typing import Dict, Any, Optional, Tuple, List
from datetime import datetime, timedelta

class ComprehensiveGamingVerificationTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        self.system_results = {}
        
    def log_test(self, test_name: str, success: bool, details: str = "", response_data: Any = None):
        """Log test results with comprehensive details"""
        result = {
            "test": test_name,
            "success": success,
            "details": details,
            "response_data": response_data,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        self.test_results.append(result)
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status}: {test_name}")
        if details:
            print(f"   Details: {details}")
        if not success and response_data:
            print(f"   Response: {response_data}")
        print()

    # ========================= AUTHENTICATION SETUP =========================

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user"""
        try:
            auth_data = {"username": "admin", "password": "admin123"}
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=auth_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.admin_token = data["access_token"]
                    self.log_test("Admin Authentication", True, f"Successfully authenticated as admin")
                    return True
            
            self.log_test("Admin Authentication", False, f"Authentication failed. Status: {response.status_code}")
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def authenticate_user(self) -> bool:
        """Authenticate as regular user with mobile OTP"""
        try:
            # Step 1: Verify mobile number
            mobile_data = {"mobile": "+919876543210"}
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-mobile", json=mobile_data)
            
            if response.status_code != 200:
                self.log_test("User Authentication - Mobile Verify", False, f"Mobile verification failed: {response.status_code}")
                return False
            
            # Step 2: Verify OTP
            otp_data = {
                "mobile": "+919876543210",
                "otp": "123456",
                "name": "Arjun Sharma",
                "email": "arjun.sharma@gmail.com",
                "referral_code": ""
            }
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-otp", json=otp_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.user_token = data["access_token"]
                    self.log_test("User Authentication", True, "Successfully authenticated user")
                    return True
            
            self.log_test("User Authentication", False, f"OTP verification failed: {response.status_code}")
            return False
            
        except Exception as e:
            self.log_test("User Authentication", False, f"Exception: {str(e)}")
            return False

    def set_admin_headers(self):
        """Set admin authorization headers"""
        if self.admin_token:
            self.session.headers.update({"Authorization": f"Bearer {self.admin_token}"})

    def set_user_headers(self):
        """Set user authorization headers"""
        if self.user_token:
            self.session.headers.update({"Authorization": f"Bearer {self.user_token}"})

    # ========================= SYSTEM 1: ACHIEVEMENT SYSTEM & BADGE MANAGEMENT =========================

    def test_achievement_system(self) -> Dict[str, Any]:
        """Test Achievement System & Badge Management endpoints"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT")
        print("-" * 60)
        
        results = {"system_name": "Achievement System", "tests": [], "working": False}
        
        # Test 1: GET /api/v1/achievements (list achievements)
        self.set_user_headers()
        try:
            response = self.session.get(f"{self.base_url}/api/v1/achievements")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                achievements = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/achievements", True, f"Found {len(achievements)} achievements")
                results["tests"].append({"endpoint": "GET /api/v1/achievements", "success": True, "details": f"Found {len(achievements)} achievements"})
            else:
                self.log_test("GET /api/v1/achievements", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/achievements", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/achievements", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/achievements", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: POST /api/v1/admin/achievements (create achievement)
        self.set_admin_headers()
        achievement_data = {
            "name": "First Victory Champion",
            "description": "Win your first contest in Fantasy Esports",
            "category": "contest",
            "trigger_type": "contest_win",
            "trigger_criteria": {"min_contests": 1},
            "reward_type": "badge_and_bonus",
            "reward_amount": 100,
            "badge_icon": "trophy-gold",
            "badge_color": "#FFD700",
            "is_active": True,
            "rarity": "common"
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/achievements", json=achievement_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    achievement_id = data.get("data", {}).get("id")
                    self.log_test("POST /api/v1/admin/achievements", True, f"Created achievement ID: {achievement_id}")
                    results["tests"].append({"endpoint": "POST /api/v1/admin/achievements", "success": True, "details": f"Created achievement ID: {achievement_id}"})
                else:
                    self.log_test("POST /api/v1/admin/achievements", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/admin/achievements", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/admin/achievements", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/admin/achievements", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/admin/achievements", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/admin/achievements", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: GET /api/v1/admin/achievements (admin list)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/achievements")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                achievements = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/admin/achievements", True, f"Admin can see {len(achievements)} achievements")
                results["tests"].append({"endpoint": "GET /api/v1/admin/achievements", "success": True, "details": f"Admin can see {len(achievements)} achievements"})
            else:
                self.log_test("GET /api/v1/admin/achievements", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/admin/achievements", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/admin/achievements", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/admin/achievements", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: GET /api/v1/users/{user_id}/achievements (user achievements)
        user_id = "1"  # Using integer user ID
        try:
            response = self.session.get(f"{self.base_url}/api/v1/users/{user_id}/achievements")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                achievements = data.get("data", []) if data.get("success") else []
                self.log_test(f"GET /api/v1/users/{user_id}/achievements", True, f"User {user_id} has {len(achievements)} achievements")
                results["tests"].append({"endpoint": f"GET /api/v1/users/{user_id}/achievements", "success": True, "details": f"User {user_id} has {len(achievements)} achievements"})
            else:
                self.log_test(f"GET /api/v1/users/{user_id}/achievements", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": f"GET /api/v1/users/{user_id}/achievements", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/users/{user_id}/achievements", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/users/{user_id}/achievements", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= SYSTEM 2: FRIEND SYSTEM & CHALLENGES =========================

    def test_friend_system(self) -> Dict[str, Any]:
        """Test Friend System & Challenges endpoints"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES")
        print("-" * 60)
        
        results = {"system_name": "Friend System", "tests": [], "working": False}
        self.set_user_headers()
        
        # Test 1: POST /api/v1/friends/add (add friend by username/mobile)
        friend_data = {"username": "rajesh_kumar", "message": "Let's compete in Fantasy Esports!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends/add", json=friend_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("POST /api/v1/friends/add (username)", True, "Friend request sent successfully")
                    results["tests"].append({"endpoint": "POST /api/v1/friends/add (username)", "success": True, "details": "Friend request sent successfully"})
                else:
                    self.log_test("POST /api/v1/friends/add (username)", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/friends/add (username)", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/friends/add (username)", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/friends/add (username)", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/friends/add (username)", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/friends/add (username)", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: POST /api/v1/friends/add (add friend by mobile)
        friend_data_mobile = {"mobile": "+919876543211", "message": "Let's compete together!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends/add", json=friend_data_mobile)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("POST /api/v1/friends/add (mobile)", True, "Friend request sent successfully")
                    results["tests"].append({"endpoint": "POST /api/v1/friends/add (mobile)", "success": True, "details": "Friend request sent successfully"})
                else:
                    self.log_test("POST /api/v1/friends/add (mobile)", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/friends/add (mobile)", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/friends/add (mobile)", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/friends/add (mobile)", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/friends/add (mobile)", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/friends/add (mobile)", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: GET /api/v1/friends (list friends)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/friends")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                friends = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/friends", True, f"User has {len(friends)} friends/requests")
                results["tests"].append({"endpoint": "GET /api/v1/friends", "success": True, "details": f"User has {len(friends)} friends/requests"})
            else:
                self.log_test("GET /api/v1/friends", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/friends", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/friends", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/friends", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: POST /api/v1/friends/{friend_id}/challenge (create challenge)
        friend_id = "2"  # Using integer friend ID
        challenge_data = {
            "contest_id": 1,  # Using integer contest ID
            "entry_fee": 50,
            "challenge_message": "Ready for an epic battle in BGMI tournament?",
            "expires_at": (datetime.now() + timedelta(days=7)).isoformat()
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends/{friend_id}/challenge", json=challenge_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    challenge_id = data.get("data", {}).get("id")
                    self.log_test(f"POST /api/v1/friends/{friend_id}/challenge", True, f"Challenge created with ID: {challenge_id}")
                    results["tests"].append({"endpoint": f"POST /api/v1/friends/{friend_id}/challenge", "success": True, "details": f"Challenge created with ID: {challenge_id}"})
                else:
                    self.log_test(f"POST /api/v1/friends/{friend_id}/challenge", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"POST /api/v1/friends/{friend_id}/challenge", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"POST /api/v1/friends/{friend_id}/challenge", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"POST /api/v1/friends/{friend_id}/challenge", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"POST /api/v1/friends/{friend_id}/challenge", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"POST /api/v1/friends/{friend_id}/challenge", "success": False, "details": f"Exception: {str(e)}"})

        # Test 5: GET /api/v1/challenges (list challenges)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/challenges")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                challenges = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/challenges", True, f"User has {len(challenges)} challenges")
                results["tests"].append({"endpoint": "GET /api/v1/challenges", "success": True, "details": f"User has {len(challenges)} challenges"})
            else:
                self.log_test("GET /api/v1/challenges", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/challenges", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/challenges", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/challenges", "success": False, "details": f"Exception: {str(e)}"})

        # Test 6: GET /api/v1/friends/activity (activity feed)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/friends/activity")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                activities = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/friends/activity", True, f"Found {len(activities)} friend activities")
                results["tests"].append({"endpoint": "GET /api/v1/friends/activity", "success": True, "details": f"Found {len(activities)} friend activities"})
            else:
                self.log_test("GET /api/v1/friends/activity", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/friends/activity", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/friends/activity", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/friends/activity", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= SYSTEM 3: SOCIAL SHARING INTEGRATION =========================

    def test_social_sharing(self) -> Dict[str, Any]:
        """Test Social Sharing Integration endpoints"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION")
        print("-" * 60)
        
        results = {"system_name": "Social Sharing", "tests": [], "working": False}
        self.set_user_headers()
        
        # Test 1: POST /api/v1/social/share (create share)
        share_data = {
            "content_type": "team_victory",
            "content_id": 789,  # Using integer instead of string
            "share_type": "manual",  # Adding required share_type field
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team in Fantasy Esports! 🏆",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook", "whatsapp", "instagram"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/social/share", json=share_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    share_id = data.get("data", {}).get("id")
                    self.log_test("POST /api/v1/social/share", True, f"Share created with ID: {share_id}")
                    results["tests"].append({"endpoint": "POST /api/v1/social/share", "success": True, "details": f"Share created with ID: {share_id}"})
                else:
                    self.log_test("POST /api/v1/social/share", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/social/share", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/social/share", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/social/share", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/social/share", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/social/share", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: GET /api/v1/social/shares (list shares)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/social/shares")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                shares = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/social/shares", True, f"User has {len(shares)} shares")
                results["tests"].append({"endpoint": "GET /api/v1/social/shares", "success": True, "details": f"User has {len(shares)} shares"})
            else:
                self.log_test("GET /api/v1/social/shares", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/social/shares", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/social/shares", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/social/shares", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: GET /api/v1/admin/social/analytics (admin analytics)
        self.set_admin_headers()
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/social/analytics")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    analytics = data.get("data", {})
                    self.log_test("GET /api/v1/admin/social/analytics", True, "Social sharing analytics accessible")
                    results["tests"].append({"endpoint": "GET /api/v1/admin/social/analytics", "success": True, "details": "Social sharing analytics accessible"})
                else:
                    self.log_test("GET /api/v1/admin/social/analytics", False, "Response missing success field")
                    results["tests"].append({"endpoint": "GET /api/v1/admin/social/analytics", "success": False, "details": "Response missing success field"})
            else:
                self.log_test("GET /api/v1/admin/social/analytics", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/admin/social/analytics", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/admin/social/analytics", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/admin/social/analytics", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: POST /api/v1/social/click-track (click tracking)
        self.set_user_headers()
        click_data = {
            "share_id": 1,  # Using integer share ID
            "platform": "twitter",
            "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            "ip_address": "192.168.1.100"
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/social/click-track", json=click_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("POST /api/v1/social/click-track", True, "Click tracking recorded successfully")
                    results["tests"].append({"endpoint": "POST /api/v1/social/click-track", "success": True, "details": "Click tracking recorded successfully"})
                else:
                    self.log_test("POST /api/v1/social/click-track", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/social/click-track", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/social/click-track", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/social/click-track", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/social/click-track", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/social/click-track", "success": False, "details": f"Exception: {str(e)}"})

        # Test 5: GET /api/v1/social/platforms (supported platforms)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/social/platforms")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                platforms = data.get("data", []) if data.get("success") else []
                self.log_test("GET /api/v1/social/platforms", True, f"Found {len(platforms)} supported platforms")
                results["tests"].append({"endpoint": "GET /api/v1/social/platforms", "success": True, "details": f"Found {len(platforms)} supported platforms"})
            else:
                self.log_test("GET /api/v1/social/platforms", False, f"Status: {response.status_code}")
                results["tests"].append({"endpoint": "GET /api/v1/social/platforms", "success": False, "details": f"Status: {response.status_code}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/social/platforms", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/social/platforms", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= SYSTEM 4: ADVANCED GAME ANALYTICS (7 METRICS) =========================

    def test_advanced_game_analytics(self) -> Dict[str, Any]:
        """Test Advanced Game Analytics (7 Metrics) endpoints"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS (7 METRICS)")
        print("-" * 60)
        
        results = {"system_name": "Advanced Game Analytics", "tests": [], "working": False}
        self.set_admin_headers()
        
        game_id = 1  # Using integer game ID instead of string
        
        # Test 1: GET /api/v1/analytics/games/{game_id}/player-efficiency
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/player-efficiency")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    efficiency = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/player-efficiency", True, "Player efficiency metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/player-efficiency", "success": True, "details": "Player efficiency metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/player-efficiency", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/player-efficiency", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/player-efficiency", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/player-efficiency", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/player-efficiency", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/player-efficiency", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: GET /api/v1/analytics/games/{game_id}/team-synergy
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/team-synergy")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    synergy = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/team-synergy", True, "Team synergy metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/team-synergy", "success": True, "details": "Team synergy metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/team-synergy", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/team-synergy", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/team-synergy", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/team-synergy", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/team-synergy", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/team-synergy", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: GET /api/v1/analytics/games/{game_id}/strategic-diversity
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/strategic-diversity")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    diversity = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", True, "Strategic diversity metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", "success": True, "details": "Strategic diversity metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/strategic-diversity", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: GET /api/v1/analytics/games/{game_id}/comeback-potential
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/comeback-potential")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    comeback = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/comeback-potential", True, "Comeback potential metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/comeback-potential", "success": True, "details": "Comeback potential metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/comeback-potential", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/comeback-potential", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/comeback-potential", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/comeback-potential", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/comeback-potential", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/comeback-potential", "success": False, "details": f"Exception: {str(e)}"})

        # Test 5: GET /api/v1/analytics/games/{game_id}/clutch-performance
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/clutch-performance")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    clutch = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/clutch-performance", True, "Clutch performance metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/clutch-performance", "success": True, "details": "Clutch performance metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/clutch-performance", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/clutch-performance", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/clutch-performance", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/clutch-performance", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/clutch-performance", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/clutch-performance", "success": False, "details": f"Exception: {str(e)}"})

        # Test 6: GET /api/v1/analytics/games/{game_id}/consistency-index
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/consistency-index")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    consistency = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/consistency-index", True, "Consistency index metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/consistency-index", "success": True, "details": "Consistency index metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/consistency-index", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/consistency-index", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/consistency-index", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/consistency-index", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/consistency-index", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/consistency-index", "success": False, "details": f"Exception: {str(e)}"})

        # Test 7: GET /api/v1/analytics/games/{game_id}/adaptability-score
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/adaptability-score")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    adaptability = data.get("data", {})
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/adaptability-score", True, "Adaptability score metrics available")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/adaptability-score", "success": True, "details": "Adaptability score metrics available"})
                else:
                    self.log_test(f"GET /api/v1/analytics/games/{game_id}/adaptability-score", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/adaptability-score", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/analytics/games/{game_id}/adaptability-score", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/adaptability-score", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/analytics/games/{game_id}/adaptability-score", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/analytics/games/{game_id}/adaptability-score", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= SYSTEM 5: PLAYER PERFORMANCE PREDICTIONS =========================

    def test_player_predictions(self) -> Dict[str, Any]:
        """Test Player Performance Predictions endpoints"""
        print("\n🤖 TESTING PLAYER PERFORMANCE PREDICTIONS")
        print("-" * 60)
        
        results = {"system_name": "Player Performance Predictions", "tests": [], "working": False}
        self.set_admin_headers()
        
        player_id = 1  # Using integer player ID
        match_id = 1   # Using integer match ID
        
        # Test 1: GET /api/v1/predictions/players/{player_id}/match/{match_id}
        try:
            response = self.session.get(f"{self.base_url}/api/v1/predictions/players/{player_id}/match/{match_id}")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    predictions = data.get("data", {})
                    self.log_test(f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", True, "Player match predictions available")
                    results["tests"].append({"endpoint": f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", "success": True, "details": "Player match predictions available"})
                else:
                    self.log_test(f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/predictions/players/{player_id}/match/{match_id}", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: GET /api/v1/predictions/match/{match_id}/teams
        try:
            response = self.session.get(f"{self.base_url}/api/v1/predictions/match/{match_id}/teams")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    team_predictions = data.get("data", {})
                    self.log_test(f"GET /api/v1/predictions/match/{match_id}/teams", True, "Team predictions available")
                    results["tests"].append({"endpoint": f"GET /api/v1/predictions/match/{match_id}/teams", "success": True, "details": "Team predictions available"})
                else:
                    self.log_test(f"GET /api/v1/predictions/match/{match_id}/teams", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/predictions/match/{match_id}/teams", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/predictions/match/{match_id}/teams", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/predictions/match/{match_id}/teams", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/predictions/match/{match_id}/teams", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/predictions/match/{match_id}/teams", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: POST /api/v1/predictions/calculate
        prediction_data = {
            "match_id": match_id,
            "player_ids": [player_id, 2, 3],
            "factors": ["recent_form", "head_to_head", "team_strength", "map_performance", "team_morale"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/predictions/calculate", json=prediction_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    calculations = data.get("data", {})
                    self.log_test("POST /api/v1/predictions/calculate", True, "Prediction calculations completed")
                    results["tests"].append({"endpoint": "POST /api/v1/predictions/calculate", "success": True, "details": "Prediction calculations completed"})
                else:
                    self.log_test("POST /api/v1/predictions/calculate", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/predictions/calculate", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/predictions/calculate", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/predictions/calculate", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/predictions/calculate", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/predictions/calculate", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: GET /api/v1/predictions/history/{player_id}
        try:
            response = self.session.get(f"{self.base_url}/api/v1/predictions/history/{player_id}")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    history = data.get("data", [])
                    self.log_test(f"GET /api/v1/predictions/history/{player_id}", True, f"Player has {len(history)} prediction history entries")
                    results["tests"].append({"endpoint": f"GET /api/v1/predictions/history/{player_id}", "success": True, "details": f"Player has {len(history)} prediction history entries"})
                else:
                    self.log_test(f"GET /api/v1/predictions/history/{player_id}", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/predictions/history/{player_id}", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/predictions/history/{player_id}", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/predictions/history/{player_id}", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/predictions/history/{player_id}", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/predictions/history/{player_id}", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= SYSTEM 6: AUTOMATED TOURNAMENT BRACKETS =========================

    def test_tournament_brackets(self) -> Dict[str, Any]:
        """Test Automated Tournament Brackets endpoints"""
        print("\n🏆 TESTING AUTOMATED TOURNAMENT BRACKETS")
        print("-" * 60)
        
        results = {"system_name": "Automated Tournament Brackets", "tests": [], "working": False}
        self.set_admin_headers()
        
        tournament_id = 1  # Using integer tournament ID
        
        # Test 1: POST /api/v1/tournaments/{id}/brackets/single-elimination
        bracket_data = {
            "name": "BGMI Championship - Single Elimination",
            "max_participants": 16,
            "seeding_method": "random",
            "settings": {
                "best_of": 3,
                "allow_third_place": True
            }
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/tournaments/{tournament_id}/brackets/single-elimination", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    bracket_id = data.get("data", {}).get("id")
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", True, f"Single elimination bracket created with ID: {bracket_id}")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", "success": True, "details": f"Single elimination bracket created with ID: {bracket_id}"})
                else:
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/single-elimination", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: POST /api/v1/tournaments/{id}/brackets/double-elimination
        bracket_data["name"] = "BGMI Championship - Double Elimination"
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/tournaments/{tournament_id}/brackets/double-elimination", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    bracket_id = data.get("data", {}).get("id")
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", True, f"Double elimination bracket created with ID: {bracket_id}")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", "success": True, "details": f"Double elimination bracket created with ID: {bracket_id}"})
                else:
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/double-elimination", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: POST /api/v1/tournaments/{id}/brackets/round-robin
        bracket_data["name"] = "BGMI Championship - Round Robin"
        bracket_data["max_participants"] = 8  # Smaller for round robin
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/tournaments/{tournament_id}/brackets/round-robin", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    bracket_id = data.get("data", {}).get("id")
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", True, f"Round robin bracket created with ID: {bracket_id}")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", "success": True, "details": f"Round robin bracket created with ID: {bracket_id}"})
                else:
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/round-robin", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: POST /api/v1/tournaments/{id}/brackets/swiss-system
        bracket_data["name"] = "BGMI Championship - Swiss System"
        bracket_data["max_participants"] = 16
        bracket_data["settings"]["rounds"] = 5
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/tournaments/{tournament_id}/brackets/swiss-system", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    bracket_id = data.get("data", {}).get("id")
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", True, f"Swiss system bracket created with ID: {bracket_id}")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", "success": True, "details": f"Swiss system bracket created with ID: {bracket_id}"})
                else:
                    self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"POST /api/v1/tournaments/{tournament_id}/brackets/swiss-system", "success": False, "details": f"Exception: {str(e)}"})

        # Test 5: GET /api/v1/tournaments/{id}/brackets/current
        try:
            response = self.session.get(f"{self.base_url}/api/v1/tournaments/{tournament_id}/brackets/current")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    current_bracket = data.get("data", {})
                    self.log_test(f"GET /api/v1/tournaments/{tournament_id}/brackets/current", True, "Current tournament bracket available")
                    results["tests"].append({"endpoint": f"GET /api/v1/tournaments/{tournament_id}/brackets/current", "success": True, "details": "Current tournament bracket available"})
                else:
                    self.log_test(f"GET /api/v1/tournaments/{tournament_id}/brackets/current", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/tournaments/{tournament_id}/brackets/current", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/tournaments/{tournament_id}/brackets/current", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/tournaments/{tournament_id}/brackets/current", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/tournaments/{tournament_id}/brackets/current", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/tournaments/{tournament_id}/brackets/current", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= SYSTEM 7: ADVANCED FRAUD DETECTION =========================

    def test_fraud_detection(self) -> Dict[str, Any]:
        """Test Advanced Fraud Detection endpoints"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION")
        print("-" * 60)
        
        results = {"system_name": "Advanced Fraud Detection", "tests": [], "working": False}
        self.set_admin_headers()
        
        # Test 1: GET /api/v1/admin/fraud/alerts
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    alerts = data.get("data", [])
                    self.log_test("GET /api/v1/admin/fraud/alerts", True, f"Found {len(alerts)} fraud alerts")
                    results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/alerts", "success": True, "details": f"Found {len(alerts)} fraud alerts"})
                else:
                    self.log_test("GET /api/v1/admin/fraud/alerts", False, "Response missing success field")
                    results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/alerts", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("GET /api/v1/admin/fraud/alerts", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/alerts", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/admin/fraud/alerts", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/alerts", "success": False, "details": f"Exception: {str(e)}"})

        # Test 2: GET /api/v1/admin/fraud/users/{user_id}/risk-score
        user_id = 1  # Using integer user ID
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/users/{user_id}/risk-score")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    risk_score = data.get("data", {})
                    self.log_test(f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", True, f"User {user_id} risk score available")
                    results["tests"].append({"endpoint": f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", "success": True, "details": f"User {user_id} risk score available"})
                else:
                    self.log_test(f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", False, "Response missing success field")
                    results["tests"].append({"endpoint": f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test(f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test(f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": f"GET /api/v1/admin/fraud/users/{user_id}/risk-score", "success": False, "details": f"Exception: {str(e)}"})

        # Test 3: POST /api/v1/admin/fraud/investigate
        investigation_data = {
            "user_id": user_id,
            "investigation_type": "suspicious_betting_pattern",
            "priority": "high",
            "notes": "User showing unusual win rate patterns and large bet amounts",
            "assigned_to": "security_team_lead"
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/fraud/investigate", json=investigation_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    investigation_id = data.get("data", {}).get("id")
                    self.log_test("POST /api/v1/admin/fraud/investigate", True, f"Investigation created with ID: {investigation_id}")
                    results["tests"].append({"endpoint": "POST /api/v1/admin/fraud/investigate", "success": True, "details": f"Investigation created with ID: {investigation_id}"})
                else:
                    self.log_test("POST /api/v1/admin/fraud/investigate", False, "Response missing success field")
                    results["tests"].append({"endpoint": "POST /api/v1/admin/fraud/investigate", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("POST /api/v1/admin/fraud/investigate", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "POST /api/v1/admin/fraud/investigate", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("POST /api/v1/admin/fraud/investigate", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "POST /api/v1/admin/fraud/investigate", "success": False, "details": f"Exception: {str(e)}"})

        # Test 4: GET /api/v1/admin/fraud/patterns
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/patterns")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    patterns = data.get("data", [])
                    self.log_test("GET /api/v1/admin/fraud/patterns", True, f"Found {len(patterns)} fraud patterns")
                    results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/patterns", "success": True, "details": f"Found {len(patterns)} fraud patterns"})
                else:
                    self.log_test("GET /api/v1/admin/fraud/patterns", False, "Response missing success field")
                    results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/patterns", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("GET /api/v1/admin/fraud/patterns", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/patterns", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("GET /api/v1/admin/fraud/patterns", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "GET /api/v1/admin/fraud/patterns", "success": False, "details": f"Exception: {str(e)}"})

        # Test 5: PUT /api/v1/admin/fraud/threshold
        threshold_data = {
            "risk_threshold": 75.0,
            "alert_threshold": 85.0,
            "auto_block_threshold": 95.0,
            "detection_sensitivity": "high"
        }
        
        try:
            response = self.session.put(f"{self.base_url}/api/v1/admin/fraud/threshold", json=threshold_data)
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("PUT /api/v1/admin/fraud/threshold", True, "Fraud detection thresholds updated")
                    results["tests"].append({"endpoint": "PUT /api/v1/admin/fraud/threshold", "success": True, "details": "Fraud detection thresholds updated"})
                else:
                    self.log_test("PUT /api/v1/admin/fraud/threshold", False, "Response missing success field")
                    results["tests"].append({"endpoint": "PUT /api/v1/admin/fraud/threshold", "success": False, "details": "Response missing success field"})
            else:
                error_msg = response.text
                self.log_test("PUT /api/v1/admin/fraud/threshold", False, f"Status: {response.status_code}, Error: {error_msg}")
                results["tests"].append({"endpoint": "PUT /api/v1/admin/fraud/threshold", "success": False, "details": f"Status: {response.status_code}, Error: {error_msg}"})
                
        except Exception as e:
            self.log_test("PUT /api/v1/admin/fraud/threshold", False, f"Exception: {str(e)}")
            results["tests"].append({"endpoint": "PUT /api/v1/admin/fraud/threshold", "success": False, "details": f"Exception: {str(e)}"})

        # Determine if system is working
        passed_tests = sum(1 for test in results["tests"] if test["success"])
        total_tests = len(results["tests"])
        results["working"] = passed_tests >= (total_tests * 0.5)  # 50% threshold
        results["success_rate"] = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        return results

    # ========================= COMPREHENSIVE TEST RUNNER =========================

    def run_comprehensive_verification_tests(self):
        """Run comprehensive verification tests for all 7 gaming systems"""
        print("🎯 COMPREHENSIVE VERIFICATION TESTING - 7 ADVANCED GAMING FEATURES")
        print("Fantasy Esports GoLang Backend - Baseline Establishment")
        print("=" * 80)
        
        # Authentication Setup
        print("\n🔐 AUTHENTICATION SETUP")
        print("-" * 40)
        
        admin_auth = self.authenticate_admin()
        user_auth = self.authenticate_user()
        
        if not admin_auth:
            print("❌ Admin authentication failed. Admin-only tests will be skipped.")
        
        if not user_auth:
            print("❌ User authentication failed. User-only tests will be skipped.")
        
        if not admin_auth and not user_auth:
            print("❌ Both authentications failed. Cannot proceed with testing.")
            return
        
        # Run all 7 system tests
        self.system_results["Achievement System"] = self.test_achievement_system()
        self.system_results["Friend System"] = self.test_friend_system()
        self.system_results["Social Sharing"] = self.test_social_sharing()
        self.system_results["Advanced Game Analytics"] = self.test_advanced_game_analytics()
        self.system_results["Player Performance Predictions"] = self.test_player_predictions()
        self.system_results["Automated Tournament Brackets"] = self.test_tournament_brackets()
        self.system_results["Advanced Fraud Detection"] = self.test_fraud_detection()
        
        # Generate comprehensive summary
        self.generate_verification_summary()

    def generate_verification_summary(self):
        """Generate comprehensive verification summary"""
        print("\n" + "=" * 80)
        print("📊 COMPREHENSIVE VERIFICATION TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"Total Tests Executed: {total_tests}")
        print(f"Tests Passed: {passed_tests} ✅")
        print(f"Tests Failed: {failed_tests} ❌")
        print(f"Overall Success Rate: {success_rate:.1f}%")
        print()
        
        # System-wise results
        print("🎯 SYSTEM-WISE RESULTS:")
        print("-" * 40)
        
        working_systems = 0
        total_systems = len(self.system_results)
        
        for system_name, system_data in self.system_results.items():
            status = "✅ WORKING" if system_data["working"] else "❌ ISSUES FOUND"
            success_rate_sys = system_data["success_rate"]
            print(f"  {system_name}: {status} ({success_rate_sys:.1f}% success rate)")
            if system_data["working"]:
                working_systems += 1
        
        system_success_rate = (working_systems / total_systems * 100) if total_systems > 0 else 0
        print(f"\nSystems Working: {working_systems}/{total_systems} ({system_success_rate:.1f}%)")
        
        # Detailed endpoint breakdown
        print("\n📋 DETAILED ENDPOINT BREAKDOWN:")
        print("-" * 40)
        
        for system_name, system_data in self.system_results.items():
            print(f"\n{system_name}:")
            for test in system_data["tests"]:
                status = "✅" if test["success"] else "❌"
                print(f"  {status} {test['endpoint']}: {test['details']}")
        
        # Failed tests summary
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print("\n❌ CRITICAL ISSUES FOUND:")
            print("-" * 40)
            
            # Group by system
            system_failures = {}
            for result in failed_results:
                test_name = result["test"]
                
                # Categorize by system
                if "Achievement" in test_name:
                    system = "Achievement System"
                elif "Friend" in test_name or "Challenge" in test_name:
                    system = "Friend System"
                elif "Social" in test_name or "Share" in test_name:
                    system = "Social Sharing"
                elif "Analytics" in test_name or "Metric" in test_name:
                    system = "Advanced Game Analytics"
                elif "Prediction" in test_name:
                    system = "Player Performance Predictions"
                elif "Tournament" in test_name or "Bracket" in test_name:
                    system = "Automated Tournament Brackets"
                elif "Fraud" in test_name:
                    system = "Advanced Fraud Detection"
                else:
                    system = "Authentication"
                
                if system not in system_failures:
                    system_failures[system] = []
                system_failures[system].append(result)
            
            for system, failures in system_failures.items():
                print(f"\n{system} Issues:")
                for failure in failures:
                    print(f"  • {failure['test']}: {failure['details']}")
        
        # Overall assessment
        print("\n" + "=" * 80)
        print("🎯 BASELINE ESTABLISHMENT COMPLETE")
        print("=" * 80)
        
        if success_rate >= 70:
            print("✅ GOOD BASELINE: Most gaming features are working well.")
            print("   The majority of endpoints are functional and ready for production.")
        elif success_rate >= 50:
            print("⚠️  MODERATE BASELINE: Some gaming features need attention.")
            print("   Several systems have issues that should be addressed.")
        else:
            print("❌ CRITICAL BASELINE: Gaming features have significant issues.")
            print("   Major problems found across multiple systems.")
        
        print(f"\n📈 BASELINE METRICS:")
        print(f"   • Overall Success Rate: {success_rate:.1f}%")
        print(f"   • Working Systems: {working_systems}/{total_systems}")
        print(f"   • Total Endpoints Tested: {total_tests}")
        print(f"   • Critical Issues: {failed_tests}")
        
        print(f"\n🔧 TESTING COMPLETED: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 80)

if __name__ == "__main__":
    tester = ComprehensiveGamingVerificationTester()
    tester.run_comprehensive_verification_tests()