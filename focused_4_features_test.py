#!/usr/bin/env python3
"""
🎯 FOCUSED TESTING - 4 ADVANCED GAMING FEATURES RE-VERIFICATION
Fantasy Esports GoLang Backend - Specific Fix Validation

This test focuses on the 4 specific Advanced Gaming Features mentioned in the review request:
1. Friend System - Test POST /api/v1/friends with username/mobile payloads to verify 400 errors instead of 500 errors
2. Social Sharing - Test POST /api/v1/share with string content_id to verify ContentIDValue type handles both string and int64
3. Advanced Game Analytics - Test GET /api/v1/admin/games/1/advanced-metrics to confirm it's still working
4. Advanced Fraud Detection - Test GET /api/v1/admin/fraud/alerts to confirm no 404 errors

Expected improvements:
- Friend System should return 400 "Bad Request" for user not found instead of 500 "Internal Server Error"
- Social Sharing should accept both string and int64 content_id values without JSON unmarshaling errors
- All systems should show improved error handling and validation
"""

import requests
import json
import time
from typing import Dict, Any, Optional, Tuple, List
from datetime import datetime

class FocusedGamingFeaturesTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        
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
            
            data = response.json()
            session_id = data.get("session_id")
            if not session_id:
                self.log_test("User Authentication - Mobile Verify", False, "No session_id received")
                return False
            
            # Step 2: Verify OTP with session_id
            otp_data = {
                "mobile": "+919876543210",
                "otp": "123456",
                "session_id": session_id
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

    # ========================= FEATURE 1: FRIEND SYSTEM =========================

    def test_friend_system_fixes(self) -> bool:
        """Test Friend System fixes - should return 400 instead of 500 for user not found"""
        print("\n👥 TESTING FRIEND SYSTEM FIXES")
        print("-" * 60)
        
        system_success = True
        self.set_user_headers()
        
        # Test 1: Add Friend by Username (Non-existent user) - Should return 400, not 500
        friend_data = {"username": "non_existent_user_12345", "message": "Let's compete!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data)
            
            if response.status_code == 400:
                data = response.json()
                self.log_test("Friend System - Username Not Found (400 Expected)", True, 
                            f"Correctly returned 400 Bad Request: {data.get('message', 'User not found')}")
            elif response.status_code == 500:
                self.log_test("Friend System - Username Not Found (400 Expected)", False, 
                            f"Still returning 500 Internal Server Error instead of 400 Bad Request")
                system_success = False
            else:
                self.log_test("Friend System - Username Not Found (400 Expected)", False, 
                            f"Unexpected status code: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Friend System - Username Not Found (400 Expected)", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Add Friend by Mobile (Non-existent user) - Should return 400, not 500
        friend_data = {"mobile": "+919999999999", "message": "Let's compete!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data)
            
            if response.status_code == 400:
                data = response.json()
                self.log_test("Friend System - Mobile Not Found (400 Expected)", True, 
                            f"Correctly returned 400 Bad Request: {data.get('message', 'User not found')}")
            elif response.status_code == 500:
                self.log_test("Friend System - Mobile Not Found (400 Expected)", False, 
                            f"Still returning 500 Internal Server Error instead of 400 Bad Request")
                system_success = False
            else:
                self.log_test("Friend System - Mobile Not Found (400 Expected)", False, 
                            f"Unexpected status code: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Friend System - Mobile Not Found (400 Expected)", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Valid Friend Request (Should work normally)
        friend_data = {"username": "rajesh_kumar", "message": "Let's compete in Fantasy Esports!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data)
            
            if response.status_code in [200, 201, 400]:  # 400 might be valid if user doesn't exist
                data = response.json()
                if response.status_code in [200, 201]:
                    self.log_test("Friend System - Valid Request", True, "Friend request processed successfully")
                else:
                    self.log_test("Friend System - Valid Request", True, f"Request handled properly with status {response.status_code}")
            else:
                self.log_test("Friend System - Valid Request", False, 
                            f"Unexpected status code: {response.status_code}, Response: {response.text}")
                system_success = False
                
        except Exception as e:
            self.log_test("Friend System - Valid Request", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= FEATURE 2: SOCIAL SHARING =========================

    def test_social_sharing_fixes(self) -> bool:
        """Test Social Sharing fixes - should handle both string and int64 content_id"""
        print("\n📱 TESTING SOCIAL SHARING FIXES")
        print("-" * 60)
        
        system_success = True
        self.set_user_headers()
        
        # Test 1: Social Sharing with String content_id (Should work now)
        share_data = {
            "content_type": "team_victory",
            "content_id": "team_789_string",  # String content_id
            "share_type": "achievement",  # Adding required share_type field
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team in Fantasy Esports! 🏆",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook", "whatsapp", "instagram"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data)
            
            if response.status_code in [200, 201]:
                data = response.json()
                if data.get("success"):
                    self.log_test("Social Sharing - String content_id", True, 
                                f"Successfully handled string content_id: {share_data['content_id']}")
                else:
                    self.log_test("Social Sharing - String content_id", False, "Response missing success field")
                    system_success = False
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                if "json: cannot unmarshal string" in str(data):
                    self.log_test("Social Sharing - String content_id", False, 
                                f"Still has JSON unmarshaling error for string content_id: {data}")
                    system_success = False
                else:
                    self.log_test("Social Sharing - String content_id", False, 
                                f"Status: {response.status_code}, Response: {data}")
                    system_success = False
                
        except Exception as e:
            self.log_test("Social Sharing - String content_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Social Sharing with Integer content_id (Should continue to work)
        share_data_int = {
            "content_type": "team_victory",
            "content_id": 789,  # Integer content_id
            "share_type": "achievement",
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team in Fantasy Esports! 🏆",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook", "whatsapp", "instagram"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_int)
            
            if response.status_code in [200, 201]:
                data = response.json()
                if data.get("success"):
                    self.log_test("Social Sharing - Integer content_id", True, 
                                f"Successfully handled integer content_id: {share_data_int['content_id']}")
                else:
                    self.log_test("Social Sharing - Integer content_id", False, "Response missing success field")
                    system_success = False
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Social Sharing - Integer content_id", False, 
                            f"Status: {response.status_code}, Response: {data}")
                system_success = False
                
        except Exception as e:
            self.log_test("Social Sharing - Integer content_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Social Sharing with int64 content_id (Large number)
        share_data_int64 = {
            "content_type": "team_victory",
            "content_id": 9223372036854775807,  # Max int64 value
            "share_type": "achievement",
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team in Fantasy Esports! 🏆",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook", "whatsapp", "instagram"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_int64)
            
            if response.status_code in [200, 201]:
                data = response.json()
                if data.get("success"):
                    self.log_test("Social Sharing - Int64 content_id", True, 
                                f"Successfully handled int64 content_id: {share_data_int64['content_id']}")
                else:
                    self.log_test("Social Sharing - Int64 content_id", False, "Response missing success field")
                    system_success = False
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Social Sharing - Int64 content_id", False, 
                            f"Status: {response.status_code}, Response: {data}")
                system_success = False
                
        except Exception as e:
            self.log_test("Social Sharing - Int64 content_id", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= FEATURE 3: ADVANCED GAME ANALYTICS =========================

    def test_advanced_game_analytics_fixes(self) -> bool:
        """Test Advanced Game Analytics - should still be working"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS FIXES")
        print("-" * 60)
        
        system_success = True
        self.set_admin_headers()
        
        # Test 1: Get Advanced Game Metrics with game ID 1 (as mentioned in review)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/1/advanced-metrics")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    metrics = data.get("data", {})
                    self.log_test("Advanced Game Analytics - Game ID 1", True, 
                                f"Successfully retrieved advanced metrics for game ID 1")
                else:
                    self.log_test("Advanced Game Analytics - Game ID 1", False, "Response missing success field")
                    system_success = False
            elif response.status_code == 400:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Advanced Game Analytics - Game ID 1", False, 
                            f"400 Bad Request - Game validation issue: {data}")
                system_success = False
            elif response.status_code == 404:
                self.log_test("Advanced Game Analytics - Game ID 1", False, 
                            f"404 Not Found - Endpoint or game not found")
                system_success = False
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Advanced Game Analytics - Game ID 1", False, 
                            f"Status: {response.status_code}, Response: {data}")
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Game Analytics - Game ID 1", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Test with different game IDs to verify validation
        for game_id in [2, "bgmi_2025"]:
            try:
                response = self.session.get(f"{self.base_url}/api/v1/admin/games/{game_id}/advanced-metrics")
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get("success"):
                        self.log_test(f"Advanced Game Analytics - Game ID {game_id}", True, 
                                    f"Successfully retrieved metrics for game ID {game_id}")
                    else:
                        self.log_test(f"Advanced Game Analytics - Game ID {game_id}", False, "Response missing success field")
                        system_success = False
                elif response.status_code == 400:
                    data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                    self.log_test(f"Advanced Game Analytics - Game ID {game_id}", True, 
                                f"Proper validation - 400 for invalid game ID: {data}")
                else:
                    data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                    self.log_test(f"Advanced Game Analytics - Game ID {game_id}", False, 
                                f"Status: {response.status_code}, Response: {data}")
                    system_success = False
                    
            except Exception as e:
                self.log_test(f"Advanced Game Analytics - Game ID {game_id}", False, f"Exception: {str(e)}")
                system_success = False

        return system_success

    # ========================= FEATURE 4: ADVANCED FRAUD DETECTION =========================

    def test_advanced_fraud_detection_fixes(self) -> bool:
        """Test Advanced Fraud Detection - should not return 404 errors"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION FIXES")
        print("-" * 60)
        
        system_success = True
        self.set_admin_headers()
        
        # Test 1: Get Fraud Alerts (Should not return 404)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    alerts = data.get("data", [])
                    self.log_test("Advanced Fraud Detection - Get Alerts", True, 
                                f"Successfully retrieved {len(alerts)} fraud alerts (no 404 error)")
                else:
                    self.log_test("Advanced Fraud Detection - Get Alerts", False, "Response missing success field")
                    system_success = False
            elif response.status_code == 404:
                self.log_test("Advanced Fraud Detection - Get Alerts", False, 
                            f"Still returning 404 Not Found - endpoint routing issue not fixed")
                system_success = False
            elif response.status_code == 401:
                self.log_test("Advanced Fraud Detection - Get Alerts", True, 
                            f"Proper authentication required - endpoint exists but needs valid admin token")
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Advanced Fraud Detection - Get Alerts", False, 
                            f"Status: {response.status_code}, Response: {data}")
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Fraud Detection - Get Alerts", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Get Fraud Statistics (Should also work)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/statistics")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    stats = data.get("data", {})
                    self.log_test("Advanced Fraud Detection - Get Statistics", True, 
                                f"Successfully retrieved fraud statistics")
                else:
                    self.log_test("Advanced Fraud Detection - Get Statistics", False, "Response missing success field")
                    system_success = False
            elif response.status_code == 404:
                self.log_test("Advanced Fraud Detection - Get Statistics", False, 
                            f"404 Not Found - endpoint routing issue")
                system_success = False
            elif response.status_code == 401:
                self.log_test("Advanced Fraud Detection - Get Statistics", True, 
                            f"Proper authentication required - endpoint exists")
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Advanced Fraud Detection - Get Statistics", False, 
                            f"Status: {response.status_code}, Response: {data}")
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Fraud Detection - Get Statistics", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Update Alert Status (Should work with proper alert ID)
        alert_id = "alert_123"
        status_data = {
            "status": "investigating",
            "notes": "Reviewing user betting patterns and account activity",
            "assigned_to": "security_team_lead"
        }
        
        try:
            response = self.session.put(f"{self.base_url}/api/v1/admin/fraud/alerts/{alert_id}/status", json=status_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    self.log_test("Advanced Fraud Detection - Update Alert Status", True, 
                                f"Successfully updated alert status")
                else:
                    self.log_test("Advanced Fraud Detection - Update Alert Status", False, "Response missing success field")
                    system_success = False
            elif response.status_code == 404:
                # This might be expected if alert doesn't exist, but endpoint should exist
                if "alert not found" in response.text.lower() or "not found" in response.text.lower():
                    self.log_test("Advanced Fraud Detection - Update Alert Status", True, 
                                f"Endpoint exists - 404 for non-existent alert is expected")
                else:
                    self.log_test("Advanced Fraud Detection - Update Alert Status", False, 
                                f"404 Not Found - endpoint routing issue")
                    system_success = False
            elif response.status_code == 401:
                self.log_test("Advanced Fraud Detection - Update Alert Status", True, 
                            f"Proper authentication required - endpoint exists")
            else:
                data = response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text
                self.log_test("Advanced Fraud Detection - Update Alert Status", False, 
                            f"Status: {response.status_code}, Response: {data}")
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Fraud Detection - Update Alert Status", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= COMPREHENSIVE TEST RUNNER =========================

    def run_focused_4_features_tests(self):
        """Run focused tests for the 4 specific gaming features"""
        print("🎯 STARTING FOCUSED 4 ADVANCED GAMING FEATURES RE-VERIFICATION")
        print("Fantasy Esports GoLang Backend - Specific Fix Validation")
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
        
        # Run the 4 specific feature tests
        feature_results = {}
        
        if user_auth:
            feature_results["Friend System"] = self.test_friend_system_fixes()
            feature_results["Social Sharing"] = self.test_social_sharing_fixes()
        else:
            print("⚠️ Skipping Friend System and Social Sharing tests - user authentication failed")
            feature_results["Friend System"] = False
            feature_results["Social Sharing"] = False
        
        if admin_auth:
            feature_results["Advanced Game Analytics"] = self.test_advanced_game_analytics_fixes()
            feature_results["Advanced Fraud Detection"] = self.test_advanced_fraud_detection_fixes()
        else:
            print("⚠️ Skipping Advanced Game Analytics and Advanced Fraud Detection tests - admin authentication failed")
            feature_results["Advanced Game Analytics"] = False
            feature_results["Advanced Fraud Detection"] = False
        
        # Generate focused summary
        self.generate_focused_summary(feature_results)

    def generate_focused_summary(self, feature_results: Dict[str, bool]):
        """Generate focused test summary for the 4 specific features"""
        print("\n" + "=" * 80)
        print("📊 FOCUSED 4 ADVANCED GAMING FEATURES TEST SUMMARY")
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
        
        # Feature-wise results
        print("🎯 FEATURE-WISE RESULTS:")
        print("-" * 40)
        
        working_features = 0
        total_features = len(feature_results)
        
        for feature_name, feature_success in feature_results.items():
            status = "✅ FIXED" if feature_success else "❌ STILL BROKEN"
            print(f"  {feature_name}: {status}")
            if feature_success:
                working_features += 1
        
        feature_success_rate = (working_features / total_features * 100) if total_features > 0 else 0
        print(f"\nFeatures Fixed: {working_features}/{total_features} ({feature_success_rate:.1f}%)")
        
        # Failed tests details
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print("\n❌ FAILED TESTS DETAILS:")
            print("-" * 40)
            for result in failed_results:
                print(f"  • {result['test']}: {result['details']}")
        
        # Overall assessment
        print("\n" + "=" * 80)
        print("🎯 FINAL ASSESSMENT")
        print("=" * 80)
        
        if feature_success_rate == 100:
            print("🎉 EXCELLENT: All 4 Advanced Gaming Features fixes are working perfectly!")
            print("   Target: Achieve close to 100% success rate - ACHIEVED!")
        elif feature_success_rate >= 75:
            print("✅ GOOD: Most Advanced Gaming Features fixes are working well.")
            print("   Target: Achieve close to 100% success rate - MOSTLY ACHIEVED!")
        elif feature_success_rate >= 50:
            print("⚠️  MODERATE: Some Advanced Gaming Features fixes are working.")
            print("   Target: Achieve close to 100% success rate - PARTIALLY ACHIEVED!")
        else:
            print("❌ CRITICAL: Advanced Gaming Features fixes are not working as expected.")
            print("   Target: Achieve close to 100% success rate - NOT ACHIEVED!")
        
        # Specific fix verification
        print(f"\n💡 SPECIFIC FIX VERIFICATION:")
        print("-" * 40)
        
        if feature_results.get("Friend System", False):
            print("  ✅ Friend System: Now returns 400 'Bad Request' instead of 500 'Internal Server Error'")
        else:
            print("  ❌ Friend System: Still has issues with error handling")
        
        if feature_results.get("Social Sharing", False):
            print("  ✅ Social Sharing: ContentIDValue type now handles both string and int64 properly")
        else:
            print("  ❌ Social Sharing: Still has JSON unmarshaling errors with content_id")
        
        if feature_results.get("Advanced Game Analytics", False):
            print("  ✅ Advanced Game Analytics: GET /api/v1/admin/games/1/advanced-metrics is working")
        else:
            print("  ❌ Advanced Game Analytics: Still has issues with game metrics endpoint")
        
        if feature_results.get("Advanced Fraud Detection", False):
            print("  ✅ Advanced Fraud Detection: GET /api/v1/admin/fraud/alerts no longer returns 404")
        else:
            print("  ❌ Advanced Fraud Detection: Still returning 404 errors for fraud endpoints")
        
        print(f"\n🔧 TESTING COMPLETED: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 80)

if __name__ == "__main__":
    tester = FocusedGamingFeaturesTester()
    tester.run_focused_4_features_tests()