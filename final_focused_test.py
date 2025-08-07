#!/usr/bin/env python3
"""
🎯 FINAL FOCUSED TESTING FOR 4 SPECIFIC FAILING ADVANCED GAMING FEATURES
Fantasy Esports GoLang Backend - Root Cause Analysis

This test focuses on the exact issues mentioned in the review request:
1. Friend System - POST /api/v1/friends endpoint with different payloads (friend_id, username, mobile) - 500 errors
2. Social Sharing - POST /api/v1/share endpoint - content_id validation issues and data type mismatches 
3. Advanced Game Analytics - GET /api/v1/admin/games/1/advanced-metrics - 400 errors
4. Advanced Fraud Detection - GET /api/v1/admin/fraud/alerts - 404 errors and admin auth issues

Focus: Show exact error messages, HTTP status codes, validation errors, authentication issues, request/response details
"""

import requests
import json
import time
from typing import Dict, Any, Optional, Tuple, List

class FinalFocusedTester:
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
        if response_data:
            print(f"   Response: {json.dumps(response_data, indent=2) if isinstance(response_data, dict) else response_data}")
        print()

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user"""
        try:
            auth_data = {"username": "admin", "password": "admin123"}
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=auth_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.admin_token = data["access_token"]
                    self.log_test("Admin Authentication", True, f"Successfully authenticated")
                    return True
            
            self.log_test("Admin Authentication", False, f"Failed with status: {response.status_code}")
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def create_test_user(self) -> bool:
        """Create a test user for authentication"""
        try:
            # Set admin headers
            if self.admin_token:
                self.session.headers.update({"Authorization": f"Bearer {self.admin_token}"})
            
            # Try to create a user via mobile verification (simulate the flow)
            # First, let's try the mobile verification endpoint
            mobile_data = {"mobile": "+919876543210"}
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-mobile", json=mobile_data)
            
            if response.status_code == 200:
                # Now try OTP verification
                otp_data = {
                    "mobile": "+919876543210",
                    "otp": "123456",
                    "name": "Test User",
                    "email": "testuser@example.com",
                    "referral_code": ""
                }
                response = self.session.post(f"{self.base_url}/api/v1/auth/verify-otp", json=otp_data)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get("success") and "access_token" in data:
                        self.user_token = data["access_token"]
                        self.log_test("Create Test User", True, "User created and authenticated successfully")
                        return True
                else:
                    self.log_test("Create Test User", False, f"OTP verification failed: {response.status_code}", response.json() if response.text else None)
            else:
                self.log_test("Create Test User", False, f"Mobile verification failed: {response.status_code}", response.json() if response.text else None)
            
            return False
            
        except Exception as e:
            self.log_test("Create Test User", False, f"Exception: {str(e)}")
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

    # ========================= SYSTEM 1: FRIEND SYSTEM - EXACT ISSUE TESTING =========================

    def test_friend_system_exact_issues(self) -> bool:
        """Test Friend System - Focus on 500 errors with different payloads"""
        print("\n👥 TESTING FRIEND SYSTEM - EXACT ISSUES FROM REVIEW REQUEST")
        print("-" * 70)
        
        system_success = True
        
        if not self.user_token:
            print("⚠️  No user token available, testing with admin token...")
            self.set_admin_headers()
        else:
            self.set_user_headers()
        
        # Test 1: friend_id payload (as mentioned in review request)
        print("🔍 Testing POST /api/v1/friends with friend_id payload (expecting 500 errors)...")
        friend_data_1 = {"friend_id": "user_123"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_1)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check if this is the 500 error mentioned in review
            is_500_error = response.status_code == 500
            self.log_test("Friend System - friend_id payload", not is_500_error, 
                         f"Testing for 500 errors mentioned in review. Payload: {friend_data_1}", 
                         response_json, response.status_code)
            
            if is_500_error:
                system_success = False
                print(f"   🚨 CONFIRMED 500 ERROR: {response_json}")
                
        except Exception as e:
            self.log_test("Friend System - friend_id payload", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: username payload (as mentioned in review request)
        print("🔍 Testing POST /api/v1/friends with username payload (expecting 500 errors)...")
        friend_data_2 = {"username": "testuser123"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_2)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check if this is the 500 error mentioned in review
            is_500_error = response.status_code == 500
            self.log_test("Friend System - username payload", not is_500_error, 
                         f"Testing for 500 errors mentioned in review. Payload: {friend_data_2}", 
                         response_json, response.status_code)
            
            if is_500_error:
                system_success = False
                print(f"   🚨 CONFIRMED 500 ERROR: {response_json}")
                
        except Exception as e:
            self.log_test("Friend System - username payload", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: mobile payload (as mentioned in review request)
        print("🔍 Testing POST /api/v1/friends with mobile payload (expecting 500 errors)...")
        friend_data_3 = {"mobile": "+919876543211"}
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/friends", json=friend_data_3)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check if this is the 500 error mentioned in review
            is_500_error = response.status_code == 500
            self.log_test("Friend System - mobile payload", not is_500_error, 
                         f"Testing for 500 errors mentioned in review. Payload: {friend_data_3}", 
                         response_json, response.status_code)
            
            if is_500_error:
                system_success = False
                print(f"   🚨 CONFIRMED 500 ERROR: {response_json}")
                
        except Exception as e:
            self.log_test("Friend System - mobile payload", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 2: SOCIAL SHARING - EXACT ISSUE TESTING =========================

    def test_social_sharing_exact_issues(self) -> bool:
        """Test Social Sharing - Focus on content_id validation issues and data type mismatches"""
        print("\n📱 TESTING SOCIAL SHARING - EXACT ISSUES FROM REVIEW REQUEST")
        print("-" * 70)
        
        system_success = True
        
        if not self.user_token:
            print("⚠️  No user token available, testing with admin token...")
            self.set_admin_headers()
        else:
            self.set_user_headers()
        
        # Test 1: content_id validation issues (string vs int64)
        print("🔍 Testing POST /api/v1/share - content_id validation issues (string vs int64)...")
        share_data_1 = {
            "content_type": "team_victory",
            "content_id": "team_789",  # String content_id (review mentions expects int64)
            "title": "My Dream Team Won!",
            "description": "Just won big!",
            "platforms": ["twitter", "facebook"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_1)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check for validation issues mentioned in review
            has_validation_issue = response.status_code in [400, 422, 500]
            self.log_test("Social Sharing - string content_id validation", not has_validation_issue, 
                         f"Testing content_id validation (string vs int64). Payload: {share_data_1}", 
                         response_json, response.status_code)
            
            if has_validation_issue:
                system_success = False
                print(f"   🚨 CONFIRMED VALIDATION ISSUE: {response_json}")
                
        except Exception as e:
            self.log_test("Social Sharing - string content_id validation", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Data type mismatches
        print("🔍 Testing POST /api/v1/share - data type mismatches...")
        share_data_2 = {
            "content_type": "team_victory",
            "content_id": 789,  # Integer content_id (as expected)
            "title": "My Dream Team Won!",
            "description": "Just won big!",
            "platforms": ["twitter", "facebook"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_2)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check for data type issues
            has_data_type_issue = response.status_code in [400, 422, 500]
            self.log_test("Social Sharing - integer content_id", not has_data_type_issue, 
                         f"Testing with integer content_id. Payload: {share_data_2}", 
                         response_json, response.status_code)
            
            if has_data_type_issue:
                system_success = False
                print(f"   🚨 CONFIRMED DATA TYPE ISSUE: {response_json}")
                
        except Exception as e:
            self.log_test("Social Sharing - integer content_id", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Missing required fields
        print("🔍 Testing POST /api/v1/share - missing required fields...")
        share_data_3 = {
            "content_type": "team_victory",
            # Missing content_id
            "title": "My Dream Team Won!",
            "platforms": ["twitter"]
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/share", json=share_data_3)
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # This should return validation error
            is_validation_error = response.status_code == 400
            self.log_test("Social Sharing - missing content_id", is_validation_error, 
                         f"Expected 400 validation error. Payload: {share_data_3}", 
                         response_json, response.status_code)
            
            if not is_validation_error:
                system_success = False
                
        except Exception as e:
            self.log_test("Social Sharing - missing content_id", False, f"Exception: {str(e)}")
            system_success = False

        return system_success

    # ========================= SYSTEM 3: ADVANCED GAME ANALYTICS - EXACT ISSUE TESTING =========================

    def test_advanced_game_analytics_exact_issues(self) -> bool:
        """Test Advanced Game Analytics - Focus on 400 errors"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS - EXACT ISSUES FROM REVIEW REQUEST")
        print("-" * 70)
        
        system_success = True
        self.set_admin_headers()
        
        # Test 1: GET /api/v1/admin/games/1/advanced-metrics (exact endpoint from review)
        print("🔍 Testing GET /api/v1/admin/games/1/advanced-metrics - 400 errors...")
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/games/1/advanced-metrics")
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check for 400 errors mentioned in review
            has_400_error = response.status_code == 400
            self.log_test("Advanced Game Analytics - game ID 1", not has_400_error, 
                         f"Testing exact endpoint from review request", 
                         response_json, response.status_code)
            
            if has_400_error:
                system_success = False
                print(f"   🚨 CONFIRMED 400 ERROR: {response_json}")
            elif response.status_code == 200:
                print(f"   ✅ SUCCESS: Analytics data returned successfully")
                
        except Exception as e:
            self.log_test("Advanced Game Analytics - game ID 1", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Test with different game IDs to identify pattern
        print("🔍 Testing with different game IDs to identify 400 error pattern...")
        
        test_game_ids = ["2", "bgmi", "valorant", "nonexistent", "0", "-1"]
        
        for game_id in test_game_ids:
            try:
                response = self.session.get(f"{self.base_url}/api/v1/admin/games/{game_id}/advanced-metrics")
                
                response_text = response.text
                try:
                    response_json = response.json()
                except:
                    response_json = {"raw_response": response_text}
                
                # Log the result for pattern analysis
                has_400_error = response.status_code == 400
                self.log_test(f"Advanced Game Analytics - game ID {game_id}", not has_400_error, 
                             f"Pattern analysis for game ID: {game_id}", 
                             response_json, response.status_code)
                
                if has_400_error:
                    print(f"   🚨 400 ERROR for game ID '{game_id}': {response_json}")
                    
            except Exception as e:
                self.log_test(f"Advanced Game Analytics - game ID {game_id}", False, f"Exception: {str(e)}")

        return system_success

    # ========================= SYSTEM 4: ADVANCED FRAUD DETECTION - EXACT ISSUE TESTING =========================

    def test_advanced_fraud_detection_exact_issues(self) -> bool:
        """Test Advanced Fraud Detection - Focus on 404 errors and admin auth issues"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION - EXACT ISSUES FROM REVIEW REQUEST")
        print("-" * 70)
        
        system_success = True
        
        # Test 1: GET /api/v1/admin/fraud/alerts (exact endpoint from review)
        print("🔍 Testing GET /api/v1/admin/fraud/alerts - 404 errors and admin auth...")
        self.set_admin_headers()
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Check for 404 errors mentioned in review
            has_404_error = response.status_code == 404
            self.log_test("Advanced Fraud Detection - admin fraud alerts", not has_404_error, 
                         f"Testing exact endpoint from review request with admin auth", 
                         response_json, response.status_code)
            
            if has_404_error:
                system_success = False
                print(f"   🚨 CONFIRMED 404 ERROR: {response_json}")
            elif response.status_code == 200:
                print(f"   ✅ SUCCESS: Fraud alerts returned successfully")
                
        except Exception as e:
            self.log_test("Advanced Fraud Detection - admin fraud alerts", False, f"Exception: {str(e)}")
            system_success = False

        # Test 2: Test admin auth issues
        print("🔍 Testing GET /api/v1/admin/fraud/alerts - admin auth issues...")
        self.clear_headers()  # Remove admin auth
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            
            response_text = response.text
            try:
                response_json = response.json()
            except:
                response_json = {"raw_response": response_text}
            
            # Should return 401 without admin auth
            is_auth_error = response.status_code == 401
            self.log_test("Advanced Fraud Detection - no admin auth", is_auth_error, 
                         f"Expected 401 without admin auth", 
                         response_json, response.status_code)
            
            if not is_auth_error:
                system_success = False
                print(f"   🚨 AUTH ISSUE: Expected 401 but got {response.status_code}")
                
        except Exception as e:
            self.log_test("Advanced Fraud Detection - no admin auth", False, f"Exception: {str(e)}")
            system_success = False

        # Test 3: Test other fraud endpoints for 404 pattern
        print("🔍 Testing other fraud endpoints for 404 error pattern...")
        self.set_admin_headers()  # Restore admin auth
        
        fraud_endpoints = [
            "/api/v1/admin/fraud/statistics",
            "/api/v1/admin/fraud/alerts/123/status",
            "/api/v1/fraud/report"
        ]
        
        for endpoint in fraud_endpoints:
            try:
                if "status" in endpoint:
                    # PUT request for status update
                    response = self.session.put(f"{self.base_url}{endpoint}", json={"status": "investigating"})
                elif "report" in endpoint:
                    # POST request for fraud report (public endpoint)
                    self.clear_headers()
                    response = self.session.post(f"{self.base_url}{endpoint}", json={
                        "report_type": "suspicious_activity",
                        "description": "Test report"
                    })
                    self.set_admin_headers()
                else:
                    # GET request
                    response = self.session.get(f"{self.base_url}{endpoint}")
                
                response_text = response.text
                try:
                    response_json = response.json()
                except:
                    response_json = {"raw_response": response_text}
                
                # Check for 404 errors
                has_404_error = response.status_code == 404
                self.log_test(f"Fraud Detection - {endpoint}", not has_404_error, 
                             f"Pattern analysis for endpoint: {endpoint}", 
                             response_json, response.status_code)
                
                if has_404_error:
                    print(f"   🚨 404 ERROR for endpoint '{endpoint}': {response_json}")
                    
            except Exception as e:
                self.log_test(f"Fraud Detection - {endpoint}", False, f"Exception: {str(e)}")

        return system_success

    # ========================= COMPREHENSIVE TEST RUNNER =========================

    def run_final_focused_tests(self):
        """Run final focused tests for the 4 specific failing systems"""
        print("🎯 FINAL FOCUSED TESTING FOR 4 SPECIFIC FAILING ADVANCED GAMING FEATURES")
        print("Fantasy Esports GoLang Backend - Root Cause Analysis")
        print("=" * 80)
        
        # Authentication Setup
        print("\n🔐 AUTHENTICATION SETUP")
        print("-" * 40)
        
        admin_auth = self.authenticate_admin()
        if not admin_auth:
            print("❌ Admin authentication failed. Cannot proceed with testing.")
            return
        
        # Try to create a test user
        user_created = self.create_test_user()
        if not user_created:
            print("⚠️  User creation failed. Will test with admin token where possible.")
        
        # Run focused tests for the 4 specific systems
        system_results = {}
        
        system_results["Friend System"] = self.test_friend_system_exact_issues()
        system_results["Social Sharing"] = self.test_social_sharing_exact_issues()
        system_results["Advanced Game Analytics"] = self.test_advanced_game_analytics_exact_issues()
        system_results["Advanced Fraud Detection"] = self.test_advanced_fraud_detection_exact_issues()
        
        # Generate final summary
        self.generate_final_summary(system_results)

    def generate_final_summary(self, system_results: Dict[str, bool]):
        """Generate final test summary with root cause analysis"""
        print("\n" + "=" * 80)
        print("📊 FINAL TEST SUMMARY - ROOT CAUSE ANALYSIS")
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
        
        # Root cause analysis by HTTP status codes
        print(f"\n🔧 ROOT CAUSE ANALYSIS BY HTTP STATUS CODES:")
        print("-" * 50)
        
        status_codes = {}
        for result in self.test_results:
            if not result["success"]:
                code = result.get('status_code', 'Unknown')
                if code not in status_codes:
                    status_codes[code] = []
                status_codes[code].append({
                    'test': result['test'],
                    'details': result['details'],
                    'response': result['response_data']
                })
        
        if status_codes:
            for code, tests in status_codes.items():
                print(f"\n  HTTP {code} ERRORS ({len(tests)} tests):")
                for test_info in tests:
                    print(f"    - {test_info['test']}")
                    if test_info['response'] and isinstance(test_info['response'], dict):
                        error_msg = test_info['response'].get('error', 'No error message')
                        print(f"      Error: {error_msg}")
        
        # Specific findings for each system
        print(f"\n🔍 SPECIFIC FINDINGS FOR EACH SYSTEM:")
        print("-" * 50)
        
        print("\n1. FRIEND SYSTEM:")
        friend_tests = [r for r in self.test_results if "Friend System" in r['test']]
        if friend_tests:
            for test in friend_tests:
                if not test['success']:
                    print(f"   ❌ {test['test']}: HTTP {test.get('status_code', 'N/A')}")
                    if test['response_data'] and isinstance(test['response_data'], dict):
                        print(f"      Error: {test['response_data'].get('error', 'No error message')}")
        
        print("\n2. SOCIAL SHARING:")
        sharing_tests = [r for r in self.test_results if "Social Sharing" in r['test']]
        if sharing_tests:
            for test in sharing_tests:
                if not test['success']:
                    print(f"   ❌ {test['test']}: HTTP {test.get('status_code', 'N/A')}")
                    if test['response_data'] and isinstance(test['response_data'], dict):
                        print(f"      Error: {test['response_data'].get('error', 'No error message')}")
        
        print("\n3. ADVANCED GAME ANALYTICS:")
        analytics_tests = [r for r in self.test_results if "Advanced Game Analytics" in r['test']]
        if analytics_tests:
            for test in analytics_tests:
                if not test['success']:
                    print(f"   ❌ {test['test']}: HTTP {test.get('status_code', 'N/A')}")
                    if test['response_data'] and isinstance(test['response_data'], dict):
                        print(f"      Error: {test['response_data'].get('error', 'No error message')}")
        
        print("\n4. ADVANCED FRAUD DETECTION:")
        fraud_tests = [r for r in self.test_results if "Advanced Fraud Detection" in r['test'] or "Fraud Detection" in r['test']]
        if fraud_tests:
            for test in fraud_tests:
                if not test['success']:
                    print(f"   ❌ {test['test']}: HTTP {test.get('status_code', 'N/A')}")
                    if test['response_data'] and isinstance(test['response_data'], dict):
                        print(f"      Error: {test['response_data'].get('error', 'No error message')}")
        
        # Final recommendations
        print(f"\n💡 RECOMMENDATIONS FOR MAIN AGENT:")
        print("-" * 50)
        
        if system_success_rate == 100:
            print("  ✅ All 4 systems are working correctly!")
            print("  ✅ No critical issues found in the specific failing features.")
        elif system_success_rate >= 75:
            print("  ⚠️  Most systems working with minor issues.")
            print("  🔧 Focus on fixing the remaining validation and error handling issues.")
        else:
            print("  ❌ Multiple critical issues found requiring immediate attention:")
            print("  🔧 Review authentication flows and token management")
            print("  🔧 Fix validation logic for request payloads")
            print("  🔧 Ensure proper error handling and status codes")
            print("  🔧 Verify admin authentication middleware is working correctly")
        
        print(f"\n🔧 TESTING COMPLETED: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 80)

if __name__ == "__main__":
    tester = FinalFocusedTester()
    tester.run_final_focused_tests()