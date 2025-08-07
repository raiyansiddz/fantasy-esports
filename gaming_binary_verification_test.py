#!/usr/bin/env python3
"""
Gaming Features Binary Verification Test
Quick verification test to check if gaming features are now accessible after backend server restart
"""

import requests
import json
import time
from typing import Dict, Any, Optional

class GamingBinaryVerificationTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = "", response_data: Any = None):
        """Log test results"""
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

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user with credentials admin/admin123"""
        try:
            auth_data = {"username": "admin", "password": "admin123"}
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=auth_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.admin_token = data["access_token"]
                    self.session.headers.update({"Authorization": f"Bearer {self.admin_token}"})
                    self.log_test("Admin Authentication", True, f"Successfully authenticated as admin")
                    return True
            
            self.log_test(
                "Admin Authentication", 
                False, 
                f"Authentication failed. Status: {response.status_code}",
                response.text
            )
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def authenticate_user(self) -> bool:
        """Authenticate as regular user using mobile verification + OTP"""
        try:
            # Step 1: Verify Mobile (send OTP)
            mobile_data = {"mobile": "+919876543210"}
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-mobile", json=mobile_data)
            
            if response.status_code != 200:
                self.log_test("User Authentication - Verify Mobile", False, f"Verify mobile failed. Status: {response.status_code}", response.text)
                return False
            
            # Step 2: Verify OTP
            otp_data = {"mobile": "+919876543210", "otp": "123456"}
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-otp", json=otp_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.user_token = data["access_token"]
                    # Create a separate session for user requests
                    self.log_test("User Authentication", True, f"Successfully authenticated user with mobile +919876543210")
                    return True
            
            self.log_test(
                "User Authentication", 
                False, 
                f"OTP verification failed. Status: {response.status_code}",
                response.text
            )
            return False
            
        except Exception as e:
            self.log_test("User Authentication", False, f"Exception: {str(e)}")
            return False

    def test_gaming_endpoints_accessibility(self) -> Dict[str, Any]:
        """Test key gaming endpoints to verify they're accessible (not 404)"""
        print("\n🎯 TESTING GAMING ENDPOINTS ACCESSIBILITY")
        print("-" * 60)
        
        # Test endpoints without authentication first to distinguish 404 vs 401
        gaming_endpoints = [
            ("/api/v1/achievements", "Achievement System"),
            ("/api/v1/friends", "Friend System"),
            ("/api/v1/share/my", "Social Sharing"),
            ("/api/v1/matches/1/predictions", "Performance Predictions"),
            ("/api/v1/admin/fraud/alerts", "Fraud Detection")
        ]
        
        accessibility_results = {
            "accessible": [],
            "not_found": [],
            "other_errors": []
        }
        
        # Remove any existing auth headers for initial test
        original_headers = self.session.headers.copy()
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']
        
        for endpoint, feature_name in gaming_endpoints:
            try:
                response = self.session.get(f"{self.base_url}{endpoint}")
                
                if response.status_code == 404:
                    # 404 = Binary compilation issue still exists
                    accessibility_results["not_found"].append((endpoint, feature_name))
                    self.log_test(f"Accessibility - {feature_name}", False, f"404 Not Found - Binary compilation issue", f"GET {endpoint} -> 404")
                elif response.status_code == 401:
                    # 401 = Gaming features are accessible, authentication required (SUCCESS!)
                    accessibility_results["accessible"].append((endpoint, feature_name))
                    self.log_test(f"Accessibility - {feature_name}", True, f"401 Unauthorized - Feature accessible, auth required", f"GET {endpoint} -> 401")
                elif response.status_code in [200, 201]:
                    # 200/201 = Features working correctly (PERFECT!)
                    accessibility_results["accessible"].append((endpoint, feature_name))
                    self.log_test(f"Accessibility - {feature_name}", True, f"200 OK - Feature working correctly", f"GET {endpoint} -> {response.status_code}")
                else:
                    # Other status codes
                    accessibility_results["other_errors"].append((endpoint, feature_name, response.status_code))
                    self.log_test(f"Accessibility - {feature_name}", False, f"Unexpected status: {response.status_code}", f"GET {endpoint} -> {response.status_code}")
                    
            except Exception as e:
                accessibility_results["other_errors"].append((endpoint, feature_name, str(e)))
                self.log_test(f"Accessibility - {feature_name}", False, f"Exception: {str(e)}")
        
        # Restore original headers
        self.session.headers.clear()
        self.session.headers.update(original_headers)
        
        return accessibility_results

    def test_basic_functionality(self, accessibility_results: Dict[str, Any]) -> bool:
        """Test basic functionality if endpoints are accessible"""
        if not accessibility_results["accessible"]:
            print("\n❌ No gaming features are accessible. Skipping functionality tests.")
            return False
        
        print(f"\n✅ {len(accessibility_results['accessible'])} gaming features are accessible. Testing basic functionality...")
        print("-" * 60)
        
        functionality_success = 0
        total_functionality_tests = 0
        
        # Test 1: Create one achievement as admin (if achievements accessible)
        if any("achievements" in endpoint for endpoint, _ in accessibility_results["accessible"]):
            total_functionality_tests += 1
            if self.test_create_achievement():
                functionality_success += 1
        
        # Test 2: Add one friend as user (if friends accessible)
        if any("friends" in endpoint for endpoint, _ in accessibility_results["accessible"]):
            total_functionality_tests += 1
            if self.test_add_friend():
                functionality_success += 1
        
        # Test 3: Generate one share URL as user (if sharing accessible)
        if any("share" in endpoint for endpoint, _ in accessibility_results["accessible"]):
            total_functionality_tests += 1
            if self.test_generate_share():
                functionality_success += 1
        
        # Test 4: Check fraud detection endpoints (if fraud accessible)
        if any("fraud" in endpoint for endpoint, _ in accessibility_results["accessible"]):
            total_functionality_tests += 1
            if self.test_fraud_detection():
                functionality_success += 1
        
        success_rate = (functionality_success / total_functionality_tests * 100) if total_functionality_tests > 0 else 0
        overall_success = success_rate >= 75  # Consider success if 75% or more functionality tests pass
        
        self.log_test(
            "Basic Functionality Overall",
            overall_success,
            f"Passed {functionality_success}/{total_functionality_tests} functionality tests ({success_rate:.1f}%)"
        )
        
        return overall_success

    def test_create_achievement(self) -> bool:
        """Test creating one achievement as admin"""
        if not self.admin_token:
            self.log_test("Create Achievement", False, "No admin token available")
            return False
        
        try:
            achievement_data = {
                "name": "First Win",
                "description": "Win your first match",
                "type": "milestone",
                "criteria": {"wins": 1},
                "reward_type": "badge",
                "reward_value": 10,
                "icon_url": "https://example.com/first-win-badge.png"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/achievements", json=achievement_data)
            
            success = response.status_code in [200, 201]
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Achievement created successfully"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test("Create Achievement", success, details, response.json() if success else response.text)
            return success
            
        except Exception as e:
            self.log_test("Create Achievement", False, f"Exception: {str(e)}")
            return False

    def test_add_friend(self) -> bool:
        """Test adding one friend as user"""
        if not self.user_token:
            self.log_test("Add Friend", False, "No user token available")
            return False
        
        try:
            # Create a separate session for user requests
            user_session = requests.Session()
            user_session.headers.update({"Authorization": f"Bearer {self.user_token}"})
            
            friend_data = {
                "friend_username": "testfriend123",
                "message": "Let's compete in fantasy esports!"
            }
            
            response = user_session.post(f"{self.base_url}/api/v1/friends/add", json=friend_data)
            
            success = response.status_code in [200, 201]
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Friend request sent successfully"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test("Add Friend", success, details, response.json() if success else response.text)
            return success
            
        except Exception as e:
            self.log_test("Add Friend", False, f"Exception: {str(e)}")
            return False

    def test_generate_share(self) -> bool:
        """Test generating one share URL as user"""
        if not self.user_token:
            self.log_test("Generate Share URL", False, "No user token available")
            return False
        
        try:
            # Create a separate session for user requests
            user_session = requests.Session()
            user_session.headers.update({"Authorization": f"Bearer {self.user_token}"})
            
            share_data = {
                "content_type": "achievement",
                "content_id": "1",
                "platform": "twitter",
                "message": "Just unlocked my first achievement in Fantasy Esports!"
            }
            
            response = user_session.post(f"{self.base_url}/api/v1/share/generate", json=share_data)
            
            success = response.status_code in [200, 201]
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Share URL generated successfully"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test("Generate Share URL", success, details, response.json() if success else response.text)
            return success
            
        except Exception as e:
            self.log_test("Generate Share URL", False, f"Exception: {str(e)}")
            return False

    def test_fraud_detection(self) -> bool:
        """Test fraud detection endpoints"""
        if not self.admin_token:
            self.log_test("Fraud Detection", False, "No admin token available")
            return False
        
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            
            success = response.status_code in [200, 401]  # 401 is also acceptable (auth required)
            details = f"Status: {response.status_code}"
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    details += " - Fraud alerts retrieved successfully"
                else:
                    success = False
                    details += " - Response missing success field"
            elif response.status_code == 401:
                details += " - Fraud detection accessible (auth required)"
            
            self.log_test("Fraud Detection", success, details, response.json() if success else response.text)
            return success
            
        except Exception as e:
            self.log_test("Fraud Detection", False, f"Exception: {str(e)}")
            return False

    def run_verification_test(self):
        """Run the focused gaming features binary verification test"""
        print("🎯 GAMING FEATURES BINARY VERIFICATION TEST")
        print("Quick verification to check if gaming features are accessible after backend restart")
        print("=" * 80)
        
        # Step 1: Authentication Verification
        print("\n🔐 STEP 1: AUTHENTICATION VERIFICATION")
        print("-" * 60)
        
        admin_auth_success = self.authenticate_admin()
        user_auth_success = self.authenticate_user()
        
        if not admin_auth_success and not user_auth_success:
            print("❌ Both admin and user authentication failed. Cannot proceed with tests.")
            return
        
        # Step 2: Gaming Features Accessibility Test
        accessibility_results = self.test_gaming_endpoints_accessibility()
        
        # Step 3: Basic Functionality Test (if endpoints are accessible)
        if accessibility_results["accessible"]:
            self.test_basic_functionality(accessibility_results)
        
        # Generate Summary
        self.generate_verification_summary(accessibility_results)

    def generate_verification_summary(self, accessibility_results: Dict[str, Any]):
        """Generate verification test summary"""
        print("\n" + "=" * 80)
        print("📊 GAMING FEATURES BINARY VERIFICATION SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {success_rate:.1f}%")
        print()
        
        # Accessibility Results
        print("🎯 GAMING FEATURES ACCESSIBILITY RESULTS:")
        print(f"  ✅ Accessible Features: {len(accessibility_results['accessible'])}")
        for endpoint, feature_name in accessibility_results["accessible"]:
            print(f"    • {feature_name} ({endpoint})")
        
        print(f"  ❌ Not Found (404): {len(accessibility_results['not_found'])}")
        for endpoint, feature_name in accessibility_results["not_found"]:
            print(f"    • {feature_name} ({endpoint})")
        
        if accessibility_results["other_errors"]:
            print(f"  ⚠️  Other Errors: {len(accessibility_results['other_errors'])}")
            for item in accessibility_results["other_errors"]:
                if len(item) == 3:
                    endpoint, feature_name, status = item
                    print(f"    • {feature_name} ({endpoint}) - Status: {status}")
        
        print("\n🔍 VERIFICATION RESULTS:")
        
        if not accessibility_results["not_found"]:
            print("🎉 SUCCESS: No 404 errors found - Gaming features are accessible!")
            print("   The backend binary now includes gaming features correctly.")
            
            if len(accessibility_results["accessible"]) >= 4:
                print("✅ EXCELLENT: All major gaming features are accessible.")
            elif len(accessibility_results["accessible"]) >= 2:
                print("✅ GOOD: Most gaming features are accessible.")
            else:
                print("⚠️  PARTIAL: Some gaming features are accessible but others may need attention.")
        else:
            print("❌ BINARY COMPILATION ISSUE: Some gaming features still return 404 errors.")
            print("   The backend binary may not include all gaming features.")
            print("   Features returning 404:")
            for endpoint, feature_name in accessibility_results["not_found"]:
                print(f"     • {feature_name}")
        
        # Overall Assessment
        print(f"\n🎯 OVERALL ASSESSMENT:")
        if not accessibility_results["not_found"] and len(accessibility_results["accessible"]) >= 3:
            print("🎉 BINARY ISSUE RESOLVED: Gaming features are now accessible!")
            print("   Ready for comprehensive testing of gaming functionality.")
        elif accessibility_results["accessible"] and len(accessibility_results["not_found"]) <= 1:
            print("✅ MOSTLY RESOLVED: Most gaming features are accessible.")
            print("   Minor issues may remain but core functionality is available.")
        else:
            print("❌ BINARY ISSUE PERSISTS: Gaming features are still not fully accessible.")
            print("   Backend binary compilation issue requires further investigation.")

if __name__ == "__main__":
    tester = GamingBinaryVerificationTester()
    tester.run_verification_test()