#!/usr/bin/env python3
"""
🎯 FOCUSED TESTING FOR 4 SPECIFIC FAILING ADVANCED GAMING FEATURES
Fantasy Esports GoLang Backend - Detailed Error Analysis

This focused test suite validates the 4 specific failing systems:
1. Friend System - Test POST /api/v1/friends endpoint with different payloads
2. Social Sharing - Test POST /api/v1/share endpoint for content_id validation issues  
3. Advanced Game Analytics - Test GET /api/v1/admin/games/1/advanced-metrics for 400 errors
4. Advanced Fraud Detection - Test GET /api/v1/admin/fraud/alerts for 404 errors and admin auth

Focus: Identify exact error messages, HTTP status codes, validation errors, and authentication issues
"""

import requests
import json
import time
from typing import Dict, Any, Optional, Tuple, List

class FocusedGamingFeaturesTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = "", response_data: Any = None, status_code: int = None):
        """Log test results with comprehensive details"""
        result = {
            "test": test_name,
            "success": success,
            "details": details,
            "response_data": response_data,
            "status_code": status_code,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        self.test_results.append(result)
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status}: {test_name}")
        if status_code:
            print(f"   Status Code: {status_code}")
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

    def clear_headers(self):
        """Clear authorization headers"""
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']

    # ========================= SYSTEM 1: FRIEND SYSTEM DETAILED TESTING =========================

    def test_friend_system_detailed(self) -> bool:
        """Test Friend System with different payloads to identify exact errors"""
        print("\n👥 DETAILED TESTING: FRIEND SYSTEM")
        print("-" * 60)
        
        system_success = True
        self.set_user_headers()
        
        # Test 1: Add Friend by friend_id (as mentioned in review request)
        print("🔍 Testing POST /api/v1/friends with friend_id payload...")
        friend_data_1 = {"friend_id": "user_123", "message": "Let's compete in Fantasy Esports!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_1)
            success = response.status_code in [200, 201]
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Add Friend by friend_id", success, 
                         f"Payload: {friend_data_1}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Add Friend by friend_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Add Friend by username (as mentioned in review request)
        print("🔍 Testing POST /api/v1/friends with username payload...")
        friend_data_2 = {"username": "rajesh_kumar", "message": "Let's compete!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_2)
            success = response.status_code in [200, 201]
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Add Friend by username", success, 
                         f"Payload: {friend_data_2}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Add Friend by username", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Add Friend by mobile (as mentioned in review request)
        print("🔍 Testing POST /api/v1/friends with mobile payload...")
        friend_data_3 = {"mobile": "+919876543211", "message": "Let's be friends!"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_3)
            success = response.status_code in [200, 201]
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Add Friend by mobile", success, 
                         f"Payload: {friend_data_3}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Add Friend by mobile", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Test with missing required fields
        print("🔍 Testing POST /api/v1/friends with missing fields...")
        friend_data_4 = {"message": "Hello"}  # Missing friend identifier
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_4)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should fail with validation error
            success = response.status_code == 400
            self.log_test("Add Friend with missing fields", success, 
                         f"Expected 400 validation error. Payload: {friend_data_4}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Add Friend with missing fields", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Test with invalid data types
        print("🔍 Testing POST /api/v1/friends with invalid data types...")
        friend_data_5 = {"friend_id": 123, "message": ["invalid", "array"]}  # Invalid types
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_5)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should fail with validation error
            success = response.status_code == 400
            self.log_test("Add Friend with invalid data types", success, 
                         f"Expected 400 validation error. Payload: {friend_data_5}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Add Friend with invalid data types", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 2: SOCIAL SHARING DETAILED TESTING =========================

    def test_social_sharing_detailed(self) -> bool:
        """Test Social Sharing with focus on content_id validation issues"""
        print("\n📱 DETAILED TESTING: SOCIAL SHARING")
        print("-" * 60)
        
        system_success = True
        self.set_user_headers()
        
        # Test 1: Create share with string content_id (as mentioned in review request)
        print("🔍 Testing POST /api/v1/share with string content_id...")
        share_data_1 = {
            "content_type": "team_victory",
            "content_id": "team_789",  # String content_id
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team!",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook", "whatsapp", "instagram"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_1)
            success = response.status_code in [200, 201]
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Create Share with string content_id", success, 
                         f"Payload: {share_data_1}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Create Share with string content_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Create share with integer content_id (as mentioned in review request - expects int64)
        print("🔍 Testing POST /api/v1/share with integer content_id...")
        share_data_2 = {
            "content_type": "team_victory",
            "content_id": 789,  # Integer content_id
            "title": "My Dream Team Won Big!",
            "description": "Just won ₹5,000 with my BGMI dream team!",
            "image_url": "https://example.com/team-victory.jpg",
            "platforms": ["twitter", "facebook"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_2)
            success = response.status_code in [200, 201]
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Create Share with integer content_id", success, 
                         f"Payload: {share_data_2}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Create Share with integer content_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Test with missing required fields
        print("🔍 Testing POST /api/v1/share with missing required fields...")
        share_data_3 = {
            "content_type": "team_victory",
            # Missing content_id
            "title": "My Dream Team Won Big!",
            "platforms": ["twitter"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_3)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should fail with validation error
            success = response.status_code == 400
            self.log_test("Create Share with missing content_id", success, 
                         f"Expected 400 validation error. Payload: {share_data_3}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Create Share with missing content_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Test with invalid content_type
        print("🔍 Testing POST /api/v1/share with invalid content_type...")
        share_data_4 = {
            "content_type": "invalid_type",
            "content_id": "123",
            "title": "Test Share",
            "platforms": ["twitter"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_4)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should fail with validation error
            success = response.status_code == 400
            self.log_test("Create Share with invalid content_type", success, 
                         f"Expected 400 validation error. Payload: {share_data_4}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Create Share with invalid content_type", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Test with invalid platforms array
        print("🔍 Testing POST /api/v1/share with invalid platforms...")
        share_data_5 = {
            "content_type": "team_victory",
            "content_id": "123",
            "title": "Test Share",
            "platforms": ["invalid_platform", "another_invalid"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_5)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should fail with validation error
            success = response.status_code == 400
            self.log_test("Create Share with invalid platforms", success, 
                         f"Expected 400 validation error. Payload: {share_data_5}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Create Share with invalid platforms", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 3: ADVANCED GAME ANALYTICS DETAILED TESTING =========================

    def test_advanced_game_analytics_detailed(self) -> bool:
        """Test Advanced Game Analytics with focus on 400 errors"""
        print("\n📊 DETAILED TESTING: ADVANCED GAME ANALYTICS")
        print("-" * 60)
        
        system_success = True
        self.set_admin_headers()
        
        # Test 1: Test with game ID "1" (as mentioned in review request)
        print("🔍 Testing GET /api/v1/admin/games/1/advanced-metrics...")
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/1/advanced-metrics")
            success = response.status_code == 200
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Advanced Metrics for game ID 1", success, 
                         f"Testing specific game ID from review request", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Metrics for game ID 1", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Test with realistic game ID
        print("🔍 Testing GET /api/v1/admin/games/bgmi_2025/advanced-metrics...")
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/bgmi_2025/advanced-metrics")
            success = response.status_code == 200
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Advanced Metrics for game ID bgmi_2025", success, 
                         f"Testing with realistic game ID", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Metrics for game ID bgmi_2025", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Test with non-existent game ID
        print("🔍 Testing GET /api/v1/admin/games/nonexistent/advanced-metrics...")
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/nonexistent/advanced-metrics")
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should return 404 or 400
            success = response.status_code in [400, 404]
            self.log_test("Advanced Metrics for non-existent game", success, 
                         f"Expected 400/404 error for non-existent game", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Metrics for non-existent game", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Test with query parameters
        print("🔍 Testing GET /api/v1/admin/games/1/advanced-metrics with query params...")
        
        try:
            params = {"metric": "player_efficiency", "days": 30}
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/1/advanced-metrics", params=params)
            success = response.status_code == 200
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Advanced Metrics with query params", success, 
                         f"Testing with query parameters: {params}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Metrics with query params", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Test without admin authentication
        print("🔍 Testing GET /api/v1/admin/games/1/advanced-metrics without admin auth...")
        self.clear_headers()
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/1/advanced-metrics")
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should return 401
            success = response.status_code == 401
            self.log_test("Advanced Metrics without admin auth", success, 
                         f"Expected 401 unauthorized error", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Advanced Metrics without admin auth", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 4: ADVANCED FRAUD DETECTION DETAILED TESTING =========================

    def test_advanced_fraud_detection_detailed(self) -> bool:
        """Test Advanced Fraud Detection with focus on 404 errors and admin auth issues"""
        print("\n🛡️ DETAILED TESTING: ADVANCED FRAUD DETECTION")
        print("-" * 60)
        
        system_success = True
        
        # Test 1: Test GET /api/v1/admin/fraud/alerts with admin auth (as mentioned in review request)
        print("🔍 Testing GET /api/v1/admin/fraud/alerts with admin auth...")
        self.set_admin_headers()
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            success = response.status_code == 200
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Get Fraud Alerts with admin auth", success, 
                         f"Testing specific endpoint from review request", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Get Fraud Alerts with admin auth", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Test without admin authentication (as mentioned in review request)
        print("🔍 Testing GET /api/v1/admin/fraud/alerts without admin auth...")
        self.clear_headers()
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            # This should return 401
            success = response.status_code == 401
            self.log_test("Get Fraud Alerts without admin auth", success, 
                         f"Expected 401 unauthorized error", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Get Fraud Alerts without admin auth", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Test fraud statistics endpoint
        print("🔍 Testing GET /api/v1/admin/fraud/statistics with admin auth...")
        self.set_admin_headers()
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/statistics")
            success = response.status_code == 200
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Get Fraud Statistics with admin auth", success, 
                         f"Testing fraud statistics endpoint", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Get Fraud Statistics with admin auth", False, f"Exception: {str(e)}")
            system_success = False

        # Test 4: Test update alert status endpoint
        print("🔍 Testing PUT /api/v1/admin/fraud/alerts/123/status with admin auth...")
        
        status_data = {
            "status": "investigating",
            "notes": "Reviewing user activity patterns",
            "assigned_to": "security_team"
        }
        
        try:
            response = self.session.put(f"{self.base_url}/api/v1/admin/fraud/alerts/123/status", json=status_data)
            success = response.status_code == 200
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Update Alert Status with admin auth", success, 
                         f"Testing alert status update. Payload: {status_data}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Update Alert Status with admin auth", False, f"Exception: {str(e)}")
            system_success = False

        # Test 5: Test public fraud report endpoint (should not require admin auth)
        print("🔍 Testing POST /api/v1/fraud/report without auth...")
        self.clear_headers()
        
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
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = response_text
            
            self.log_test("Report Fraud without auth", success, 
                         f"Testing public fraud report. Payload: {fraud_report}", 
                         response_json, response.status_code)
            
            if not success:
                system_success = False
                
        except Exception as e:
            self.log_test("Report Fraud without auth", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= COMPREHENSIVE TEST RUNNER =========================

    def run_focused_tests(self):
        """Run focused tests for the 4 specific failing systems"""
        print("🎯 STARTING FOCUSED TESTING FOR 4 SPECIFIC FAILING ADVANCED GAMING FEATURES")
        print("Fantasy Esports GoLang Backend - Detailed Error Analysis")
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
        
        # Run focused tests for the 4 specific systems
        system_results = {}
        
        system_results["Friend System"] = self.test_friend_system_detailed()
        system_results["Social Sharing"] = self.test_social_sharing_detailed()
        system_results["Advanced Game Analytics"] = self.test_advanced_game_analytics_detailed()
        system_results["Advanced Fraud Detection"] = self.test_advanced_fraud_detection_detailed()
        
        # Generate focused summary
        self.generate_focused_summary(system_results)

    def generate_focused_summary(self, system_results: Dict[str, bool]):
        """Generate focused test summary for the 4 specific systems"""
        print("\n" + "=" * 80)
        print("📊 FOCUSED TEST SUMMARY - 4 SPECIFIC FAILING SYSTEMS")
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
        
        # Detailed error analysis
        print("\n🔍 DETAILED ERROR ANALYSIS:")
        print("-" * 40)
        
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            for result in failed_results:
                print(f"\n❌ {result['test']}:")
                print(f"   Status Code: {result.get('status_code', 'N/A')}")
                print(f"   Details: {result['details']}")
                if result['response_data']:
                    print(f"   Response: {result['response_data']}")
        else:
            print("🎉 No failed tests found!")
        
        # Root cause analysis
        print(f"\n🔧 ROOT CAUSE ANALYSIS:")
        print("-" * 40)
        
        # Analyze common error patterns
        status_codes = {}
        for result in failed_results:
            code = result.get('status_code', 'Unknown')
            if code not in status_codes:
                status_codes[code] = []
            status_codes[code].append(result['test'])
        
        if status_codes:
            for code, tests in status_codes.items():
                print(f"  HTTP {code} errors: {len(tests)} tests")
                for test in tests:
                    print(f"    - {test}")
        
        # Final assessment
        print("\n" + "=" * 80)
        print("🎯 FINAL ASSESSMENT FOR 4 SPECIFIC SYSTEMS")
        print("=" * 80)
        
        if success_rate >= 90:
            print("🎉 EXCELLENT: The 4 specific systems are working well!")
        elif success_rate >= 75:
            print("✅ GOOD: Most functionality working with minor issues.")
        elif success_rate >= 50:
            print("⚠️  MODERATE: Several issues found that need attention.")
        else:
            print("❌ CRITICAL: Significant issues found requiring immediate fixes.")
        
        print(f"\n🔧 TESTING COMPLETED: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 80)

if __name__ == "__main__":
    tester = FocusedGamingFeaturesTester()
    tester.run_focused_tests()