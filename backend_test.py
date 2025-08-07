#!/usr/bin/env python3
"""
🎯 VERIFICATION TESTING AFTER ROUTE FIXES - 7 ADVANCED GAMING FEATURES
Fantasy Esports GoLang Backend - Post Binary Fix Validation

OBJECTIVE: Verify that missing gaming feature routes have been properly added 
and are now accessible after rebuilding the Go binary with latest source code.

BACKEND STATUS:
✅ New binary deployed: fantasy-esports-backend-fixed-v2
✅ Go 1.21.3 installed and used for compilation  
✅ Routes added to server.go and compiled into binary
✅ Backend running on http://localhost:8001

FOCUS: Test 7 gaming systems with emphasis on newly fixed routes:
1. Achievement System & Badge Management ✅ (Expected to work as before)
2. Friend System & Challenges 🔧 (Previous issues to verify)
3. Social Sharing Integration 🔧 (Previous validation issues)  
4. Advanced Game Analytics 🎯 (NEWLY FIXED - 7 metrics endpoints)
5. Player Performance Predictions 🎯 (NEWLY FIXED - 4 prediction endpoints)
6. Automated Tournament Brackets 🎯 (NEWLY FIXED - 5 bracket endpoints)
7. Advanced Fraud Detection 🎯 (NEWLY FIXED - 4 additional admin endpoints)

SUCCESS CRITERIA: Previously 404 endpoints should now return proper responses
TARGET: >70% success rate (improvement from previous 15.8%)
"""

import requests
import json
import time
import uuid
import random
from typing import Dict, Any, Optional, Tuple, List
from datetime import datetime, timedelta

class AdvancedGamingFeaturesTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        self.created_resources = {
            "achievements": [],
            "friends": [],
            "challenges": [],
            "shares": [],
            "brackets": [],
            "predictions": [],
            "fraud_reports": []
        }
        
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
        """Authenticate as admin user with multiple methods"""
        try:
            auth_methods = [
                {"username": "admin", "password": "admin123"},
                {"email": "admin@fantasy-esports.com", "password": "admin123"},
                {"username": "admin", "password": "password"},
            ]
            
            for auth_data in auth_methods:
                response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=auth_data)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get("success") and "access_token" in data:
                        self.admin_token = data["access_token"]
                        self.log_test("Admin Authentication", True, f"Successfully authenticated with {auth_data}")
                        return True
            
            self.log_test("Admin Authentication", False, f"All authentication methods failed. Last status: {response.status_code}")
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
                    self.log_test("User Authentication", True, "Successfully authenticated user with mobile +919876543210")
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

    def test_achievement_system(self) -> bool:
        """Test comprehensive achievement system functionality"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT")
        print("-" * 60)
        
        system_success = True
        
        # Test 1: Admin - Create Achievement
        self.set_admin_headers()
        achievement_data = {
            "name": "First Victory",
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
                if data.get("success") and "data" in data:
                    achievement_id = data["data"].get("id")
                    if achievement_id:
                        self.created_resources["achievements"].append(achievement_id)
                        self.log_test("Achievement Creation", True, f"Created achievement ID: {achievement_id}")
                    else:
                        success = False
                        self.log_test("Achievement Creation", False, "Response missing achievement ID")
                else:
                    success = False
                    self.log_test("Achievement Creation", False, "Invalid response structure")
            else:
                self.log_test("Achievement Creation", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Achievement Creation", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Admin - List Achievements
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/achievements")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    achievements = data.get("data", [])
                    self.log_test("Admin List Achievements", True, f"Found {len(achievements)} achievements")
                else:
                    success = False
                    self.log_test("Admin List Achievements", False, "Response missing success field")
            else:
                self.log_test("Admin List Achievements", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Admin List Achievements", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: User - Get Available Achievements
        self.set_user_headers()
        try:
            response = self.session.get(f"{self.base_url}/api/v1/achievements")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    achievements = data.get("data", [])
                    self.log_test("User Available Achievements", True, f"User can see {len(achievements)} achievements")
                else:
                    success = False
                    self.log_test("User Available Achievements", False, "Response missing success field")
            else:
                self.log_test("User Available Achievements", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("User Available Achievements", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: User - Get My Achievements
        try:
            response = self.session.get(f"{self.base_url}/api/v1/achievements/my")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    my_achievements = data.get("data", [])
                    self.log_test("User My Achievements", True, f"User has {len(my_achievements)} earned achievements")
                else:
                    success = False
                    self.log_test("User My Achievements", False, "Response missing success field")
            else:
                self.log_test("User My Achievements", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("User My Achievements", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Achievement Progress Tracking
        if self.created_resources["achievements"]:
            achievement_id = self.created_resources["achievements"][0]
            try:
                response = self.session.get(f"{self.base_url}/api/v1/achievements/{achievement_id}/progress")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        progress = data.get("data", {})
                        self.log_test("Achievement Progress Tracking", True, f"Progress tracking working for achievement {achievement_id}")
                    else:
                        success = False
                        self.log_test("Achievement Progress Tracking", False, "Response missing success field")
                else:
                    self.log_test("Achievement Progress Tracking", False, f"Status: {response.status_code}")
                    system_success = False
                    
            except Exception as e:
                self.log_test("Achievement Progress Tracking", False, f"Exception: {str(e)}")
                system_success = False

        return system_success

    # ========================= SYSTEM 2: FRIEND SYSTEM & CHALLENGES =========================

    def test_friend_system(self) -> bool:
        """Test comprehensive friend system and challenges"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES")
        print("-" * 60)
        
        system_success = True
        self.set_user_headers()
        
        # Test 1: Add Friend by Username
        friend_data = {"username": "rajesh_kumar", "message": "Let's compete in Fantasy Esports!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("Add Friend by Username", True, "Friend request sent successfully")
                else:
                    success = False
                    self.log_test("Add Friend by Username", False, "Response missing success field")
            else:
                self.log_test("Add Friend by Username", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Add Friend by Username", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Get Friends List
        try:
            response = self.session.get(f"{self.base_url}/api/v1/friends")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    friends = data.get("data", [])
                    self.log_test("Get Friends List", True, f"User has {len(friends)} friends/requests")
                else:
                    success = False
                    self.log_test("Get Friends List", False, "Response missing success field")
            else:
                self.log_test("Get Friends List", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Friends List", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Create Friend Challenge
        challenge_data = {
            "friend_id": "friend_user_123",
            "contest_id": "contest_456",
            "entry_fee": 50,
            "challenge_message": "Ready for an epic battle in BGMI tournament?",
            "expires_at": (datetime.now() + timedelta(days=7)).isoformat()
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/challenges", json=challenge_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    challenge_id = data["data"].get("id")
                    if challenge_id:
                        self.created_resources["challenges"].append(challenge_id)
                        self.log_test("Create Friend Challenge", True, f"Challenge created with ID: {challenge_id}")
                    else:
                        success = False
                        self.log_test("Create Friend Challenge", False, "Response missing challenge ID")
                else:
                    success = False
                    self.log_test("Create Friend Challenge", False, "Invalid response structure")
            else:
                self.log_test("Create Friend Challenge", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Create Friend Challenge", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Get Challenges List
        try:
            response = self.session.get(f"{self.base_url}/api/v1/challenges")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    challenges = data.get("data", [])
                    self.log_test("Get Challenges List", True, f"User has {len(challenges)} challenges")
                else:
                    success = False
                    self.log_test("Get Challenges List", False, "Response missing success field")
            else:
                self.log_test("Get Challenges List", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Challenges List", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Friend Activities Feed
        try:
            response = self.session.get(f"{self.base_url}/api/v1/friends/activities")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    activities = data.get("data", [])
                    self.log_test("Friend Activities Feed", True, f"Found {len(activities)} friend activities")
                else:
                    success = False
                    self.log_test("Friend Activities Feed", False, "Response missing success field")
            else:
                self.log_test("Friend Activities Feed", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Friend Activities Feed", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 3: SOCIAL SHARING INTEGRATION =========================

    def test_social_sharing(self) -> bool:
        """Test comprehensive social sharing integration"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION")
        print("-" * 60)
        
        system_success = True
        self.set_user_headers()
        
        # Test 1: Create Shareable Content
        share_data = {
            "content_type": "team_victory",
            "content_id": "team_789",
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team in Fantasy Esports! 🏆",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook", "whatsapp", "instagram"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    share_id = data["data"].get("id")
                    if share_id:
                        self.created_resources["shares"].append(share_id)
                        self.log_test("Create Shareable Content", True, f"Share created with ID: {share_id}")
                    else:
                        success = False
                        self.log_test("Create Shareable Content", False, "Response missing share ID")
                else:
                    success = False
                    self.log_test("Create Shareable Content", False, "Invalid response structure")
            else:
                self.log_test("Create Shareable Content", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Create Shareable Content", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Generate Team Share URLs
        team_id = "team_456"
        try:
            response = self.session.get(f"{self.base_url}/api/v1/share/teams/{team_id}/urls")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    urls = data.get("data", {})
                    platforms = ["twitter", "facebook", "whatsapp", "instagram"]
                    found_platforms = [p for p in platforms if p in urls]
                    self.log_test("Generate Team Share URLs", True, f"Generated URLs for {len(found_platforms)} platforms: {found_platforms}")
                else:
                    success = False
                    self.log_test("Generate Team Share URLs", False, "Response missing success field")
            else:
                self.log_test("Generate Team Share URLs", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Generate Team Share URLs", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Generate Contest Win Share URLs
        contest_id = "contest_789"
        try:
            response = self.session.get(f"{self.base_url}/api/v1/share/contests/{contest_id}/urls")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    urls = data.get("data", {})
                    platforms = ["twitter", "facebook", "whatsapp", "instagram"]
                    found_platforms = [p for p in platforms if p in urls]
                    self.log_test("Generate Contest Win Share URLs", True, f"Generated URLs for {len(found_platforms)} platforms: {found_platforms}")
                else:
                    success = False
                    self.log_test("Generate Contest Win Share URLs", False, "Response missing success field")
            else:
                self.log_test("Generate Contest Win Share URLs", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Generate Contest Win Share URLs", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Generate Achievement Share URLs
        achievement_id = "achievement_123"
        try:
            response = self.session.get(f"{self.base_url}/api/v1/share/achievements/{achievement_id}/urls")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    urls = data.get("data", {})
                    platforms = ["twitter", "facebook", "whatsapp", "instagram"]
                    found_platforms = [p for p in platforms if p in urls]
                    self.log_test("Generate Achievement Share URLs", True, f"Generated URLs for {len(found_platforms)} platforms: {found_platforms}")
                else:
                    success = False
                    self.log_test("Generate Achievement Share URLs", False, "Response missing success field")
            else:
                self.log_test("Generate Achievement Share URLs", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Generate Achievement Share URLs", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Track Share Click
        if self.created_resources["shares"]:
            share_id = self.created_resources["shares"][0]
            try:
                response = self.session.post(f"{self.base_url}/api/v1/share/{share_id}/click")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        self.log_test("Track Share Click", True, f"Share click tracked for ID: {share_id}")
                    else:
                        success = False
                        self.log_test("Track Share Click", False, "Response missing success field")
                else:
                    self.log_test("Track Share Click", False, f"Status: {response.status_code}")
                    system_success = False
                    
            except Exception as e:
                self.log_test("Track Share Click", False, f"Exception: {str(e)}")
                system_success = False

        # Test 6: Get User Sharing History
        try:
            response = self.session.get(f"{self.base_url}/api/v1/share/my")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    shares = data.get("data", [])
                    self.log_test("Get User Sharing History", True, f"User has {len(shares)} shares in history")
                else:
                    success = False
                    self.log_test("Get User Sharing History", False, "Response missing success field")
            else:
                self.log_test("Get User Sharing History", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get User Sharing History", False, f"Exception: {str(e)}")
            system_success = False

        # Test 7: Admin - Social Sharing Analytics
        self.set_admin_headers()
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/social/analytics")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    analytics = data.get("data", {})
                    self.log_test("Admin Social Sharing Analytics", True, "Social sharing analytics accessible")
                else:
                    success = False
                    self.log_test("Admin Social Sharing Analytics", False, "Response missing success field")
            else:
                self.log_test("Admin Social Sharing Analytics", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Admin Social Sharing Analytics", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 4: ADVANCED GAME ANALYTICS (7 NEWLY FIXED METRICS) =========================

    def test_advanced_game_analytics(self) -> bool:
        """Test all 7 newly fixed game analytics metrics endpoints"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS (7 NEWLY FIXED METRICS)")
        print("-" * 60)
        
        system_success = True
        self.set_admin_headers()
        
        game_id = 1  # Using integer game ID as mentioned in review
        
        # Test 1: Player Efficiency Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/player-efficiency")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Player Efficiency Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Player Efficiency Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Player Efficiency Metric", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Team Synergy Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/team-synergy")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Team Synergy Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Team Synergy Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Team Synergy Metric", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Strategic Diversity Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/strategic-diversity")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Strategic Diversity Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Strategic Diversity Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Strategic Diversity Metric", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Comeback Potential Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/comeback-potential")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Comeback Potential Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Comeback Potential Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Comeback Potential Metric", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Clutch Performance Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/clutch-performance")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Clutch Performance Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Clutch Performance Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Clutch Performance Metric", False, f"Exception: {str(e)}")
            system_success = False

        # Test 6: Consistency Index Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/consistency-index")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Consistency Index Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Consistency Index Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Consistency Index Metric", False, f"Exception: {str(e)}")
            system_success = False

        # Test 7: Adaptability Score Metric (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/analytics/games/{game_id}/adaptability-score")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                self.log_test("Adaptability Score Metric", True, f"✅ FIXED: Endpoint accessible (was 404)")
            else:
                self.log_test("Adaptability Score Metric", False, f"Status: {response.status_code} - Expected 200 after fix")
                system_success = False
                
        except Exception as e:
            self.log_test("Adaptability Score Metric", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 5: TOURNAMENT BRACKETS (4 TYPES) =========================

    def test_tournament_brackets(self) -> bool:
        """Test all 4 tournament bracket types"""
        print("\n🏆 TESTING TOURNAMENT BRACKETS (4 TYPES)")
        print("-" * 60)
        
        system_success = True
        self.set_admin_headers()
        
        # Test 1: Get Available Bracket Types
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/brackets/types")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    bracket_types = data.get("data", [])
                    expected_types = ["single_elimination", "double_elimination", "round_robin", "swiss_system"]
                    found_types = [bt for bt in expected_types if bt in [b.get("type") for b in bracket_types]]
                    self.log_test("Available Bracket Types", True, f"Found {len(found_types)}/4 bracket types: {found_types}")
                else:
                    success = False
                    self.log_test("Available Bracket Types", False, "Response missing success field")
            else:
                self.log_test("Available Bracket Types", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Available Bracket Types", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Create Single Elimination Bracket
        bracket_data = {
            "tournament_id": "tournament_123",
            "bracket_type": "single_elimination",
            "name": "BGMI Championship - Single Elimination",
            "max_participants": 16,
            "seeding_method": "random",
            "settings": {
                "best_of": 3,
                "allow_third_place": True
            }
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/tournaments/brackets", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    bracket_id = data["data"].get("id")
                    if bracket_id:
                        self.created_resources["brackets"].append(bracket_id)
                        self.log_test("Create Single Elimination Bracket", True, f"Created bracket ID: {bracket_id}")
                    else:
                        success = False
                        self.log_test("Create Single Elimination Bracket", False, "Response missing bracket ID")
                else:
                    success = False
                    self.log_test("Create Single Elimination Bracket", False, "Invalid response structure")
            else:
                self.log_test("Create Single Elimination Bracket", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Create Single Elimination Bracket", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Create Double Elimination Bracket
        bracket_data["bracket_type"] = "double_elimination"
        bracket_data["name"] = "BGMI Championship - Double Elimination"
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/tournaments/brackets", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    bracket_id = data["data"].get("id")
                    if bracket_id:
                        self.created_resources["brackets"].append(bracket_id)
                        self.log_test("Create Double Elimination Bracket", True, f"Created bracket ID: {bracket_id}")
                    else:
                        success = False
                        self.log_test("Create Double Elimination Bracket", False, "Response missing bracket ID")
                else:
                    success = False
                    self.log_test("Create Double Elimination Bracket", False, "Invalid response structure")
            else:
                self.log_test("Create Double Elimination Bracket", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Create Double Elimination Bracket", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Create Round Robin Bracket
        bracket_data["bracket_type"] = "round_robin"
        bracket_data["name"] = "BGMI Championship - Round Robin"
        bracket_data["max_participants"] = 8  # Smaller for round robin
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/tournaments/brackets", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    bracket_id = data["data"].get("id")
                    if bracket_id:
                        self.created_resources["brackets"].append(bracket_id)
                        self.log_test("Create Round Robin Bracket", True, f"Created bracket ID: {bracket_id}")
                    else:
                        success = False
                        self.log_test("Create Round Robin Bracket", False, "Response missing bracket ID")
                else:
                    success = False
                    self.log_test("Create Round Robin Bracket", False, "Invalid response structure")
            else:
                self.log_test("Create Round Robin Bracket", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Create Round Robin Bracket", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Create Swiss System Bracket
        bracket_data["bracket_type"] = "swiss_system"
        bracket_data["name"] = "BGMI Championship - Swiss System"
        bracket_data["max_participants"] = 16
        bracket_data["settings"]["rounds"] = 5
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/tournaments/brackets", json=bracket_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    bracket_id = data["data"].get("id")
                    if bracket_id:
                        self.created_resources["brackets"].append(bracket_id)
                        self.log_test("Create Swiss System Bracket", True, f"Created bracket ID: {bracket_id}")
                    else:
                        success = False
                        self.log_test("Create Swiss System Bracket", False, "Response missing bracket ID")
                else:
                    success = False
                    self.log_test("Create Swiss System Bracket", False, "Invalid response structure")
            else:
                self.log_test("Create Swiss System Bracket", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Create Swiss System Bracket", False, f"Exception: {str(e)}")
            system_success = False

        # Test 6: Get Tournament Brackets
        tournament_id = "tournament_123"
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/tournaments/{tournament_id}/brackets")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    brackets = data.get("data", [])
                    self.log_test("Get Tournament Brackets", True, f"Tournament has {len(brackets)} brackets")
                else:
                    success = False
                    self.log_test("Get Tournament Brackets", False, "Response missing success field")
            else:
                self.log_test("Get Tournament Brackets", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Tournament Brackets", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 6: PLAYER PERFORMANCE PREDICTIONS (AI-BASED) =========================

    def test_player_predictions(self) -> bool:
        """Test AI-based player performance predictions"""
        print("\n🤖 TESTING PLAYER PERFORMANCE PREDICTIONS (AI-BASED)")
        print("-" * 60)
        
        system_success = True
        
        # Test 1: Admin - Generate Match Predictions
        self.set_admin_headers()
        match_id = "match_456"
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/matches/{match_id}/generate-predictions")
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    predictions = data.get("data", {})
                    
                    # Check for AI prediction factors
                    expected_factors = ["recent_form", "head_to_head", "team_strength", "map_performance", "team_morale"]
                    found_factors = [factor for factor in expected_factors if factor in predictions]
                    
                    self.log_test("Generate Match Predictions", True, 
                                f"AI predictions generated with {len(found_factors)}/5 factors: {found_factors}")
                else:
                    success = False
                    self.log_test("Generate Match Predictions", False, "Response missing success field")
            else:
                self.log_test("Generate Match Predictions", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Generate Match Predictions", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: User - Get Match Predictions
        self.set_user_headers()
        try:
            response = self.session.get(f"{self.base_url}/api/v1/matches/{match_id}/predictions")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    predictions = data.get("data", {})
                    
                    # Check for prediction components
                    if "confidence_score" in predictions and "predicted_outcome" in predictions:
                        confidence = predictions.get("confidence_score", 0)
                        self.log_test("Get Match Predictions", True, f"Predictions available with {confidence}% confidence")
                    else:
                        success = False
                        self.log_test("Get Match Predictions", False, "Missing prediction components")
                else:
                    success = False
                    self.log_test("Get Match Predictions", False, "Response missing success field")
            else:
                self.log_test("Get Match Predictions", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Match Predictions", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Admin - Update Prediction Accuracy
        self.set_admin_headers()
        accuracy_data = {
            "actual_outcome": "team_a_win",
            "actual_score": {"team_a": 16, "team_b": 12},
            "prediction_accuracy": 85.5
        }
        
        try:
            response = self.session.put(f"{self.base_url}/api/v1/admin/matches/{match_id}/update-accuracy", json=accuracy_data)
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("Update Prediction Accuracy", True, "Prediction accuracy updated successfully")
                else:
                    success = False
                    self.log_test("Update Prediction Accuracy", False, "Response missing success field")
            else:
                self.log_test("Update Prediction Accuracy", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Update Prediction Accuracy", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Admin - Get Prediction Analytics
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/predictions/analytics")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    analytics = data.get("data", {})
                    
                    # Check for analytics components
                    expected_components = ["overall_accuracy", "prediction_count", "accuracy_by_game", "confidence_distribution"]
                    found_components = [comp for comp in expected_components if comp in analytics]
                    
                    self.log_test("Get Prediction Analytics", True, 
                                f"Analytics available with {len(found_components)}/4 components: {found_components}")
                else:
                    success = False
                    self.log_test("Get Prediction Analytics", False, "Response missing success field")
            else:
                self.log_test("Get Prediction Analytics", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Prediction Analytics", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 7: ADVANCED FRAUD DETECTION SYSTEM =========================

    def test_fraud_detection(self) -> bool:
        """Test advanced fraud detection system"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION SYSTEM")
        print("-" * 60)
        
        system_success = True
        
        # Test 1: Public - Report Suspicious Activity
        # Remove auth headers for public endpoint
        original_headers = self.session.headers.copy()
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']
        
        fraud_report = {
            "report_type": "suspicious_betting_pattern",
            "user_id": "user_789",
            "description": "User placing unusually large bets with perfect win rate",
            "evidence": {
                "bet_amounts": [5000, 7500, 10000],
                "win_rate": 100,
                "time_pattern": "always_bets_just_before_match_start"
            },
            "reporter_contact": "security@fantasy-esports.com"
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/fraud/report", json=fraud_report)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    report_id = data.get("data", {}).get("id")
                    if report_id:
                        self.created_resources["fraud_reports"].append(report_id)
                        self.log_test("Report Suspicious Activity", True, f"Fraud report created with ID: {report_id}")
                    else:
                        self.log_test("Report Suspicious Activity", True, "Fraud report submitted successfully")
                else:
                    success = False
                    self.log_test("Report Suspicious Activity", False, "Response missing success field")
            else:
                self.log_test("Report Suspicious Activity", False, f"Status: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Report Suspicious Activity", False, f"Exception: {str(e)}")
            system_success = False

        # Restore headers
        self.session.headers.clear()
        self.session.headers.update(original_headers)

        # Test 2: Admin - Get Fraud Alerts
        self.set_admin_headers()
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    alerts = data.get("data", [])
                    self.log_test("Get Fraud Alerts", True, f"Found {len(alerts)} fraud alerts")
                else:
                    success = False
                    self.log_test("Get Fraud Alerts", False, "Response missing success field")
            else:
                self.log_test("Get Fraud Alerts", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Fraud Alerts", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Admin - Get Fraud Statistics
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/statistics")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    stats = data.get("data", {})
                    
                    # Check for fraud statistics components
                    expected_stats = ["total_reports", "active_alerts", "resolved_cases", "detection_accuracy"]
                    found_stats = [stat for stat in expected_stats if stat in stats]
                    
                    self.log_test("Get Fraud Statistics", True, 
                                f"Statistics available with {len(found_stats)}/4 components: {found_stats}")
                else:
                    success = False
                    self.log_test("Get Fraud Statistics", False, "Response missing success field")
            else:
                self.log_test("Get Fraud Statistics", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Get Fraud Statistics", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Admin - Update Alert Status
        # Create a mock alert ID for testing
        alert_id = "alert_123"
        status_data = {
            "status": "investigating",
            "notes": "Reviewing user betting patterns and account activity",
            "assigned_to": "security_team_lead"
        }
        
        try:
            response = self.session.put(f"{self.base_url}/api/v1/admin/fraud/alerts/{alert_id}/status", json=status_data)
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("Update Alert Status", True, f"Alert {alert_id} status updated to investigating")
                else:
                    success = False
                    self.log_test("Update Alert Status", False, "Response missing success field")
            else:
                self.log_test("Update Alert Status", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Update Alert Status", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Fraud Webhook (Public endpoint)
        # Remove auth headers for public endpoint
        original_headers = self.session.headers.copy()
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']
        
        webhook_data = {
            "event_type": "suspicious_activity_detected",
            "user_id": "user_456",
            "detection_algorithm": "ml_betting_pattern_analyzer",
            "confidence_score": 92.5,
            "details": {
                "anomaly_type": "unusual_win_rate",
                "risk_level": "high",
                "recommended_action": "account_review"
            }
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/fraud/webhook", json=webhook_data)
            success = response.status_code in [200, 201]
            
            if success:
                data = response.json()
                if data.get("success"):
                    self.log_test("Fraud Webhook", True, "Fraud webhook processed successfully")
                else:
                    success = False
                    self.log_test("Fraud Webhook", False, "Response missing success field")
            else:
                self.log_test("Fraud Webhook", False, f"Status: {response.status_code}")
                system_success = False
                
        except Exception as e:
            self.log_test("Fraud Webhook", False, f"Exception: {str(e)}")
            system_success = False

        # Restore headers
        self.session.headers.clear()
        self.session.headers.update(original_headers)

        return system_success

    # ========================= COMPREHENSIVE TEST RUNNER =========================

    def run_comprehensive_gaming_features_tests(self):
        """Run all 7 Advanced Gaming Features tests"""
        print("🎯 STARTING COMPREHENSIVE ADVANCED GAMING FEATURES TESTING")
        print("Fantasy Esports GoLang Backend - All 7 Systems Validation")
        print("=" * 80)
        
        # Authentication Setup
        print("\n🔐 AUTHENTICATION SETUP")
        print("-" * 40)
        
        admin_auth = self.authenticate_admin()
        user_auth = self.authenticate_user()
        
        if not admin_auth:
            print("❌ Admin authentication failed. Some tests will be skipped.")
        
        if not user_auth:
            print("❌ User authentication failed. Some tests will be skipped.")
        
        if not admin_auth and not user_auth:
            print("❌ Both authentications failed. Cannot proceed with testing.")
            return
        
        # Run all 7 system tests
        system_results = {}
        
        system_results["Achievement System"] = self.test_achievement_system()
        system_results["Friend System"] = self.test_friend_system()
        system_results["Social Sharing"] = self.test_social_sharing()
        system_results["Advanced Analytics"] = self.test_advanced_game_analytics()
        system_results["Tournament Brackets"] = self.test_tournament_brackets()
        system_results["Player Predictions"] = self.test_player_predictions()
        system_results["Fraud Detection"] = self.test_fraud_detection()
        
        # Generate comprehensive summary
        self.generate_comprehensive_summary(system_results)

    def generate_comprehensive_summary(self, system_results: Dict[str, bool]):
        """Generate comprehensive test summary for all 7 systems"""
        print("\n" + "=" * 80)
        print("📊 COMPREHENSIVE ADVANCED GAMING FEATURES TEST SUMMARY")
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
        total_systems = len(system_results)
        
        for system_name, system_success in system_results.items():
            status = "✅ WORKING" if system_success else "❌ ISSUES FOUND"
            print(f"  {system_name}: {status}")
            if system_success:
                working_systems += 1
        
        system_success_rate = (working_systems / total_systems * 100) if total_systems > 0 else 0
        print(f"\nSystems Working: {working_systems}/{total_systems} ({system_success_rate:.1f}%)")
        
        # Detailed test breakdown
        print("\n📋 DETAILED TEST BREAKDOWN:")
        print("-" * 40)
        
        system_test_counts = {}
        for result in self.test_results:
            test_name = result["test"]
            
            # Categorize tests by system
            if "Achievement" in test_name:
                system = "Achievement System"
            elif "Friend" in test_name or "Challenge" in test_name:
                system = "Friend System"
            elif "Share" in test_name or "Social" in test_name:
                system = "Social Sharing"
            elif "Analytics" in test_name or "Metric" in test_name:
                system = "Advanced Analytics"
            elif "Bracket" in test_name or "Tournament" in test_name:
                system = "Tournament Brackets"
            elif "Prediction" in test_name:
                system = "Player Predictions"
            elif "Fraud" in test_name:
                system = "Fraud Detection"
            else:
                system = "Authentication"
            
            if system not in system_test_counts:
                system_test_counts[system] = {"passed": 0, "total": 0}
            
            system_test_counts[system]["total"] += 1
            if result["success"]:
                system_test_counts[system]["passed"] += 1
        
        for system, counts in system_test_counts.items():
            rate = (counts["passed"] / counts["total"] * 100) if counts["total"] > 0 else 0
            print(f"  {system}: {counts['passed']}/{counts['total']} passed ({rate:.1f}%)")
        
        # Failed tests details
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print("\n❌ FAILED TESTS DETAILS:")
            print("-" * 40)
            for result in failed_results:
                print(f"  • {result['test']}: {result['details']}")
        
        # Created resources summary
        print(f"\n📝 CREATED TEST RESOURCES:")
        print("-" * 40)
        total_resources = 0
        for resource_type, ids in self.created_resources.items():
            if ids:
                print(f"  {resource_type}: {len(ids)} items created")
                total_resources += len(ids)
        
        if total_resources == 0:
            print("  No resources were created during testing")
        
        # Overall assessment
        print("\n" + "=" * 80)
        print("🎯 FINAL ASSESSMENT")
        print("=" * 80)
        
        if success_rate >= 90:
            print("🎉 EXCELLENT: All 7 Advanced Gaming Features are working excellently!")
            print("   The Fantasy Esports backend has production-ready gaming functionality.")
        elif success_rate >= 75:
            print("✅ GOOD: Most Advanced Gaming Features are working well with minor issues.")
            print("   The majority of gaming functionality is production-ready.")
        elif success_rate >= 50:
            print("⚠️  MODERATE: Some Advanced Gaming Features have issues that need attention.")
            print("   Several gaming systems need fixes before production deployment.")
        else:
            print("❌ CRITICAL: Advanced Gaming Features have significant issues requiring immediate attention.")
            print("   Major problems found in multiple gaming systems.")
        
        # Recommendations
        print(f"\n💡 RECOMMENDATIONS:")
        print("-" * 40)
        
        if system_success_rate == 100:
            print("  • All 7 gaming systems are functional - ready for production!")
            print("  • Consider performance optimization and load testing")
            print("  • Implement monitoring and alerting for production deployment")
        elif system_success_rate >= 75:
            print("  • Focus on fixing the failing systems identified above")
            print("  • Most systems are ready for production use")
            print("  • Consider gradual rollout of working features")
        else:
            print("  • Significant development work needed on multiple systems")
            print("  • Review implementation approach for failing systems")
            print("  • Consider prioritizing the most critical gaming features")
        
        print(f"\n🔧 TESTING COMPLETED: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 80)

if __name__ == "__main__":
    tester = AdvancedGamingFeaturesTester()
    tester.run_comprehensive_gaming_features_tests()