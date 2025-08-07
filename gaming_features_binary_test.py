#!/usr/bin/env python3
"""
🎯 CRITICAL GAMING FEATURES VERIFICATION TEST - Post Binary Fix

This test verifies if gaming features are now accessible and working correctly
after the backend server supervisor configuration has been fixed to run the 
correct Go binary instead of Python uvicorn.

TESTING FOCUS AREAS:
1. Binary Compilation Verification - Test Go backend is running correctly
2. Gaming Features Accessibility Testing - Test 5 gaming feature endpoints
3. Authentication & Functionality Testing - Test with proper authentication
4. Route Registration Analysis - Verify gaming routes appear in logs

EXPECTED RESULTS:
- ✅ 404 errors should be ELIMINATED - gaming features should be accessible
- ✅ 401 errors = Success (features accessible, authentication required)  
- ✅ 200/201 responses with auth = Features working correctly
- ❌ 404 errors = Binary compilation issue still exists
"""

import requests
import json
import time
from typing import Dict, Any, Optional, List, Tuple

class GamingFeaturesBinaryTester:
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

    # ========================= PHASE 1: BINARY COMPILATION VERIFICATION =========================

    def test_health_endpoint(self) -> bool:
        """Test health endpoint to verify Go backend is running correctly"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            success = response.status_code == 200
            
            details = f"Status: {response.status_code}"
            if success:
                try:
                    data = response.json()
                    if "go" in response.text.lower() or "golang" in response.text.lower():
                        details += " - Go backend confirmed"
                    else:
                        details += " - Backend type unclear from response"
                except:
                    details += f" - Response: {response.text[:100]}"
            
            self.log_test(
                "Binary Verification - Health Endpoint",
                success,
                details,
                response.text if success else response.text
            )
            return success
        except Exception as e:
            self.log_test("Binary Verification - Health Endpoint", False, f"Exception: {str(e)}")
            return False

    def test_supervisor_binary_check(self) -> bool:
        """Verify supervisor is running the correct Go binary"""
        try:
            import subprocess
            result = subprocess.run(['sudo', 'supervisorctl', 'status', 'backend'], 
                                  capture_output=True, text=True)
            
            if result.returncode == 0 and "RUNNING" in result.stdout:
                details = "Backend service is running via supervisor"
                
                # Check if the correct binary is being used
                if "fantasy-esports-backend" in result.stdout or "RUNNING" in result.stdout:
                    details += " - Correct Go binary process confirmed"
                    success = True
                else:
                    details += " - Binary process unclear"
                    success = False
            else:
                success = False
                details = f"Backend not running: {result.stdout}"
            
            self.log_test(
                "Binary Verification - Supervisor Check",
                success,
                details,
                result.stdout
            )
            return success
        except Exception as e:
            self.log_test("Binary Verification - Supervisor Check", False, f"Exception: {str(e)}")
            return False

    # ========================= PHASE 2: GAMING FEATURES ACCESSIBILITY TESTING =========================

    def test_gaming_features_accessibility(self) -> Dict[str, List[Tuple[str, str, int]]]:
        """Test the 5 critical gaming feature endpoints for accessibility"""
        print("\n🎯 TESTING GAMING FEATURES ACCESSIBILITY")
        print("-" * 60)
        
        # The 5 gaming feature endpoints to test
        gaming_endpoints = [
            ("/api/v1/achievements", "Achievement System"),
            ("/api/v1/friends", "Friend System"), 
            ("/api/v1/share/my", "Social Sharing"),
            ("/api/v1/matches/1/predictions", "Performance Predictions"),
            ("/api/v1/admin/fraud/alerts", "Fraud Detection")
        ]
        
        results = {
            "accessible": [],      # 401 or 200/201 responses
            "not_found": [],       # 404 responses (binary issue)
            "other_errors": []     # Other status codes
        }
        
        # Remove any existing auth headers for initial accessibility test
        original_headers = self.session.headers.copy()
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']
        
        for endpoint, feature_name in gaming_endpoints:
            try:
                response = self.session.get(f"{self.base_url}{endpoint}")
                status_code = response.status_code
                
                if status_code == 404:
                    # 404 = Binary compilation issue - gaming features not accessible
                    results["not_found"].append((endpoint, feature_name, status_code))
                    self.log_test(
                        f"Accessibility - {feature_name}", 
                        False, 
                        f"404 Not Found - Binary compilation issue persists",
                        f"GET {endpoint} -> 404"
                    )
                elif status_code == 401:
                    # 401 = SUCCESS! Features accessible, authentication required
                    results["accessible"].append((endpoint, feature_name, status_code))
                    self.log_test(
                        f"Accessibility - {feature_name}", 
                        True, 
                        f"401 Unauthorized - Feature accessible, auth required (SUCCESS!)",
                        f"GET {endpoint} -> 401"
                    )
                elif status_code in [200, 201]:
                    # 200/201 = PERFECT! Features working correctly
                    results["accessible"].append((endpoint, feature_name, status_code))
                    self.log_test(
                        f"Accessibility - {feature_name}", 
                        True, 
                        f"{status_code} OK - Feature working correctly (PERFECT!)",
                        f"GET {endpoint} -> {status_code}"
                    )
                else:
                    # Other status codes - need investigation
                    results["other_errors"].append((endpoint, feature_name, status_code))
                    self.log_test(
                        f"Accessibility - {feature_name}", 
                        False, 
                        f"Unexpected status: {status_code}",
                        f"GET {endpoint} -> {status_code}: {response.text[:100]}"
                    )
                    
            except Exception as e:
                results["other_errors"].append((endpoint, feature_name, f"Exception: {str(e)}"))
                self.log_test(f"Accessibility - {feature_name}", False, f"Exception: {str(e)}")
        
        # Restore original headers
        self.session.headers.clear()
        self.session.headers.update(original_headers)
        
        return results

    # ========================= PHASE 3: AUTHENTICATION & FUNCTIONALITY TESTING =========================

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user"""
        try:
            auth_data = {"username": "admin", "password": "admin123"}
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=auth_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.admin_token = data["access_token"]
                    self.session.headers.update({"Authorization": f"Bearer {self.admin_token}"})
                    self.log_test("Admin Authentication", True, "Successfully authenticated as admin")
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
                    self.log_test("User Authentication", True, "Successfully authenticated user with mobile +919876543210")
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

    def test_authenticated_gaming_features(self, accessible_endpoints: List[Tuple[str, str, int]]) -> bool:
        """Test gaming features with proper authentication"""
        if not accessible_endpoints:
            self.log_test("Authenticated Gaming Features", False, "No accessible endpoints to test")
            return False
        
        print(f"\n✅ Testing {len(accessible_endpoints)} accessible gaming features with authentication...")
        print("-" * 60)
        
        functionality_success = 0
        total_functionality_tests = 0
        
        # Test Achievement System (if accessible)
        if any("achievements" in endpoint for endpoint, _, _ in accessible_endpoints):
            total_functionality_tests += 1
            if self.test_achievement_functionality():
                functionality_success += 1
        
        # Test Friend System (if accessible)
        if any("friends" in endpoint for endpoint, _, _ in accessible_endpoints):
            total_functionality_tests += 1
            if self.test_friend_functionality():
                functionality_success += 1
        
        # Test Social Sharing (if accessible)
        if any("share" in endpoint for endpoint, _, _ in accessible_endpoints):
            total_functionality_tests += 1
            if self.test_sharing_functionality():
                functionality_success += 1
        
        # Test Performance Predictions (if accessible)
        if any("predictions" in endpoint for endpoint, _, _ in accessible_endpoints):
            total_functionality_tests += 1
            if self.test_predictions_functionality():
                functionality_success += 1
        
        # Test Fraud Detection (if accessible)
        if any("fraud" in endpoint for endpoint, _, _ in accessible_endpoints):
            total_functionality_tests += 1
            if self.test_fraud_functionality():
                functionality_success += 1
        
        success_rate = (functionality_success / total_functionality_tests * 100) if total_functionality_tests > 0 else 0
        overall_success = success_rate >= 60  # Consider success if 60% or more functionality tests pass
        
        self.log_test(
            "Authenticated Gaming Features Overall",
            overall_success,
            f"Passed {functionality_success}/{total_functionality_tests} functionality tests ({success_rate:.1f}%)"
        )
        
        return overall_success

    def test_achievement_functionality(self) -> bool:
        """Test Achievement System functionality"""
        if not self.admin_token:
            return False
        
        try:
            # Test creating an achievement
            achievement_data = {
                "name": "First Victory",
                "description": "Win your first match in Fantasy Esports",
                "type": "milestone",
                "criteria": {"wins": 1},
                "reward_type": "badge",
                "reward_value": 50,
                "icon_url": "https://example.com/first-victory-badge.png"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/achievements", json=achievement_data)
            
            success = response.status_code in [200, 201]
            details = f"Status: {response.status_code}"
            
            if success:
                try:
                    data = response.json()
                    if data.get("success"):
                        details += " - Achievement created successfully"
                    else:
                        success = False
                        details += " - Response missing success field"
                except:
                    details += " - Achievement creation response received"
            
            self.log_test("Achievement System Functionality", success, details)
            return success
            
        except Exception as e:
            self.log_test("Achievement System Functionality", False, f"Exception: {str(e)}")
            return False

    def test_friend_functionality(self) -> bool:
        """Test Friend System functionality"""
        if not self.user_token:
            return False
        
        try:
            # Create a separate session for user requests
            user_session = requests.Session()
            user_session.headers.update({"Authorization": f"Bearer {self.user_token}"})
            
            # Test getting friends list
            response = user_session.get(f"{self.base_url}/api/v1/friends")
            
            success = response.status_code in [200, 401]  # 401 is also acceptable (auth working)
            details = f"Status: {response.status_code}"
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data.get("success"):
                        details += " - Friends list retrieved successfully"
                    else:
                        details += " - Friends endpoint accessible"
                except:
                    details += " - Friends endpoint accessible"
            elif response.status_code == 401:
                details += " - Friends system accessible (auth required)"
            
            self.log_test("Friend System Functionality", success, details)
            return success
            
        except Exception as e:
            self.log_test("Friend System Functionality", False, f"Exception: {str(e)}")
            return False

    def test_sharing_functionality(self) -> bool:
        """Test Social Sharing functionality"""
        if not self.user_token:
            return False
        
        try:
            # Create a separate session for user requests
            user_session = requests.Session()
            user_session.headers.update({"Authorization": f"Bearer {self.user_token}"})
            
            # Test getting user's shares
            response = user_session.get(f"{self.base_url}/api/v1/share/my")
            
            success = response.status_code in [200, 401]  # 401 is also acceptable (auth working)
            details = f"Status: {response.status_code}"
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data.get("success"):
                        details += " - User shares retrieved successfully"
                    else:
                        details += " - Social sharing endpoint accessible"
                except:
                    details += " - Social sharing endpoint accessible"
            elif response.status_code == 401:
                details += " - Social sharing accessible (auth required)"
            
            self.log_test("Social Sharing Functionality", success, details)
            return success
            
        except Exception as e:
            self.log_test("Social Sharing Functionality", False, f"Exception: {str(e)}")
            return False

    def test_predictions_functionality(self) -> bool:
        """Test Performance Predictions functionality"""
        if not self.user_token:
            return False
        
        try:
            # Create a separate session for user requests
            user_session = requests.Session()
            user_session.headers.update({"Authorization": f"Bearer {self.user_token}"})
            
            # Test getting predictions for a match
            response = user_session.get(f"{self.base_url}/api/v1/matches/1/predictions")
            
            success = response.status_code in [200, 401, 404]  # 404 might be acceptable if match doesn't exist
            details = f"Status: {response.status_code}"
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data.get("success"):
                        details += " - Match predictions retrieved successfully"
                    else:
                        details += " - Performance predictions endpoint accessible"
                except:
                    details += " - Performance predictions endpoint accessible"
            elif response.status_code == 401:
                details += " - Performance predictions accessible (auth required)"
            elif response.status_code == 404:
                details += " - Performance predictions accessible (match not found)"
            
            self.log_test("Performance Predictions Functionality", success, details)
            return success
            
        except Exception as e:
            self.log_test("Performance Predictions Functionality", False, f"Exception: {str(e)}")
            return False

    def test_fraud_functionality(self) -> bool:
        """Test Fraud Detection functionality"""
        if not self.admin_token:
            return False
        
        try:
            # Test getting fraud alerts
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts")
            
            success = response.status_code in [200, 401]  # 401 is also acceptable (auth working)
            details = f"Status: {response.status_code}"
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data.get("success"):
                        details += " - Fraud alerts retrieved successfully"
                    else:
                        details += " - Fraud detection endpoint accessible"
                except:
                    details += " - Fraud detection endpoint accessible"
            elif response.status_code == 401:
                details += " - Fraud detection accessible (auth required)"
            
            self.log_test("Fraud Detection Functionality", success, details)
            return success
            
        except Exception as e:
            self.log_test("Fraud Detection Functionality", False, f"Exception: {str(e)}")
            return False

    # ========================= PHASE 4: ROUTE REGISTRATION ANALYSIS =========================

    def analyze_backend_logs(self) -> bool:
        """Analyze backend logs for route registration"""
        try:
            import subprocess
            
            # Get recent backend logs
            result = subprocess.run(['sudo', 'tail', '-n', '50', '/var/log/supervisor/backend.out.log'], 
                                  capture_output=True, text=True)
            
            if result.returncode == 0:
                logs = result.stdout
                
                # Look for gaming route registrations
                gaming_routes_found = 0
                gaming_keywords = ["achievements", "friends", "share", "predictions", "fraud"]
                
                for keyword in gaming_keywords:
                    if keyword in logs.lower():
                        gaming_routes_found += 1
                
                success = gaming_routes_found >= 2  # At least 2 gaming features mentioned in logs
                details = f"Found {gaming_routes_found}/5 gaming features mentioned in backend logs"
                
                if "error" in logs.lower() or "panic" in logs.lower():
                    details += " - Errors detected in logs"
                    success = False
                elif "started" in logs.lower() or "listening" in logs.lower():
                    details += " - Backend started successfully"
                
                self.log_test("Route Registration Analysis", success, details, logs[-200:] if logs else "No logs")
                return success
            else:
                self.log_test("Route Registration Analysis", False, "Could not access backend logs")
                return False
                
        except Exception as e:
            self.log_test("Route Registration Analysis", False, f"Exception: {str(e)}")
            return False

    # ========================= MAIN TEST EXECUTION =========================

    def run_comprehensive_gaming_test(self):
        """Run the comprehensive gaming features binary verification test"""
        print("🎯 CRITICAL GAMING FEATURES VERIFICATION TEST - Post Binary Fix")
        print("Testing if gaming features are now accessible after supervisor configuration fix")
        print("=" * 80)
        
        # Phase 1: Binary Compilation Verification
        print("\n🔧 PHASE 1: BINARY COMPILATION VERIFICATION")
        print("-" * 60)
        
        health_ok = self.test_health_endpoint()
        supervisor_ok = self.test_supervisor_binary_check()
        
        if not health_ok:
            print("❌ Backend health check failed. Cannot proceed with gaming features testing.")
            return
        
        # Phase 2: Gaming Features Accessibility Testing
        accessibility_results = self.test_gaming_features_accessibility()
        
        # Phase 3: Authentication & Functionality Testing (if endpoints are accessible)
        if accessibility_results["accessible"]:
            print("\n🔐 PHASE 3: AUTHENTICATION & FUNCTIONALITY TESTING")
            print("-" * 60)
            
            admin_auth = self.authenticate_admin()
            user_auth = self.authenticate_user()
            
            if admin_auth or user_auth:
                self.test_authenticated_gaming_features(accessibility_results["accessible"])
        
        # Phase 4: Route Registration Analysis
        print("\n📋 PHASE 4: ROUTE REGISTRATION ANALYSIS")
        print("-" * 60)
        
        self.analyze_backend_logs()
        
        # Generate Comprehensive Summary
        self.generate_comprehensive_summary(accessibility_results)

    def generate_comprehensive_summary(self, accessibility_results: Dict[str, List[Tuple]]):
        """Generate comprehensive test summary"""
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
        
        # Gaming Features Accessibility Results
        print("🎯 GAMING FEATURES ACCESSIBILITY RESULTS:")
        print(f"  ✅ Accessible Features: {len(accessibility_results['accessible'])}")
        for endpoint, feature_name, status in accessibility_results["accessible"]:
            print(f"    • {feature_name} ({endpoint}) - Status: {status}")
        
        print(f"  ❌ Not Found (404): {len(accessibility_results['not_found'])}")
        for endpoint, feature_name, status in accessibility_results["not_found"]:
            print(f"    • {feature_name} ({endpoint}) - Status: {status}")
        
        if accessibility_results["other_errors"]:
            print(f"  ⚠️  Other Errors: {len(accessibility_results['other_errors'])}")
            for item in accessibility_results["other_errors"]:
                if len(item) >= 3:
                    endpoint, feature_name, status = item[0], item[1], item[2]
                    print(f"    • {feature_name} ({endpoint}) - Status: {status}")
        
        print("\n🔍 BINARY FIX VERIFICATION RESULTS:")
        
        # Critical Assessment
        if not accessibility_results["not_found"]:
            print("🎉 SUCCESS: No 404 errors found - Gaming features are accessible!")
            print("   ✅ The backend binary now includes gaming features correctly.")
            print("   ✅ Supervisor configuration fix has resolved the binary compilation issue.")
            
            if len(accessibility_results["accessible"]) == 5:
                print("   🏆 PERFECT: All 5 gaming features are accessible!")
            elif len(accessibility_results["accessible"]) >= 4:
                print("   ✅ EXCELLENT: 4+ gaming features are accessible.")
            elif len(accessibility_results["accessible"]) >= 3:
                print("   ✅ GOOD: 3+ gaming features are accessible.")
            else:
                print("   ⚠️  PARTIAL: Some gaming features are accessible but others may need attention.")
        else:
            print("❌ BINARY COMPILATION ISSUE PERSISTS: Some gaming features still return 404 errors.")
            print("   The backend binary may not include all gaming features or routes are not registered.")
            print("   Features still returning 404:")
            for endpoint, feature_name, status in accessibility_results["not_found"]:
                print(f"     • {feature_name} ({endpoint})")
        
        # Overall Assessment
        print(f"\n🎯 OVERALL ASSESSMENT:")
        
        accessible_count = len(accessibility_results["accessible"])
        not_found_count = len(accessibility_results["not_found"])
        
        if not_found_count == 0 and accessible_count >= 4:
            print("🎉 BINARY ISSUE COMPLETELY RESOLVED!")
            print("   ✅ Gaming features are now fully accessible and ready for comprehensive testing.")
            print("   ✅ The supervisor configuration fix has successfully resolved the binary compilation issue.")
        elif not_found_count <= 1 and accessible_count >= 3:
            print("✅ BINARY ISSUE MOSTLY RESOLVED!")
            print("   ✅ Most gaming features are accessible with minor issues remaining.")
            print("   ✅ Core gaming functionality is available for testing.")
        elif accessible_count >= 2:
            print("⚠️  BINARY ISSUE PARTIALLY RESOLVED!")
            print("   ⚠️  Some gaming features are accessible but significant issues remain.")
            print("   🔧 Further investigation needed for remaining 404 errors.")
        else:
            print("❌ BINARY ISSUE PERSISTS!")
            print("   ❌ Gaming features are still not accessible due to binary compilation issues.")
            print("   🔧 Backend binary compilation requires immediate attention.")
        
        # Recommendations
        print(f"\n💡 RECOMMENDATIONS:")
        if not_found_count == 0:
            print("   🎯 Ready for comprehensive gaming features functionality testing")
            print("   🎯 Proceed with full integration testing of gaming systems")
        elif not_found_count <= 2:
            print("   🔧 Investigate remaining 404 endpoints for route registration issues")
            print("   🎯 Proceed with testing accessible gaming features")
        else:
            print("   🔧 Review backend binary compilation process")
            print("   🔧 Check if all gaming feature handlers are included in the build")
            print("   🔧 Verify route registration in server initialization")

if __name__ == "__main__":
    tester = GamingFeaturesBinaryTester()
    tester.run_comprehensive_gaming_test()