#!/usr/bin/env python3
"""
Backend API Testing Script for Fantasy Esports Platform
Testing GoLang Fantasy Esports backend notification system fixes after critical backend configuration issue:

FOCUS: Re-test the notification system fixes after fixing the critical backend configuration issue
PRIORITY 1: Statistics Filtering Fix - Test ALL statistics filtering scenarios to confirm SQL syntax errors are resolved
PRIORITY 2: Enhanced Validation Fixes - Re-test validation order fixes

Expected: 
- All statistics filtering tests should return 200 status with proper JSON response (not 500 SQL syntax errors)
- Validation should return correct error messages in proper order
"""

import requests
import json
import sys
import time
from typing import Dict, Any, List, Optional

class NotificationSystemTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.api_base = f"{base_url}/api/v1"
        self.session = requests.Session()
        self.test_results = []
        self.admin_token = None
        
    def log_test(self, test_name: str, passed: bool, details: str, response_data: Optional[Dict] = None):
        """Log test results"""
        result = {
            "test": test_name,
            "passed": passed,
            "details": details,
            "response_data": response_data,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        self.test_results.append(result)
        
        status = "✅ PASS" if passed else "❌ FAIL"
        print(f"{status} | {test_name}")
        print(f"      Details: {details}")
        if response_data:
            print(f"      Response: {json.dumps(response_data, indent=2)[:200]}...")
        print()

    def test_health_check(self):
        """Test if backend is running"""
        try:
            response = self.session.get(f"{self.base_url}/health", timeout=10)
            if response.status_code == 200:
                data = response.json()
                self.log_test(
                    "Backend Health Check",
                    True,
                    f"Backend is running. Status: {data.get('status', 'unknown')}",
                    data
                )
                return True
            else:
                self.log_test(
                    "Backend Health Check",
                    False,
                    f"Backend returned status {response.status_code}",
                    {"status_code": response.status_code}
                )
                return False
        except Exception as e:
            self.log_test(
                "Backend Health Check",
                False,
                f"Backend connection failed: {str(e)}",
                {"error": str(e)}
            )
            return False

    def test_admin_login(self):
        """Test admin authentication to get token for protected endpoints"""
        try:
            payload = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.api_base}/admin/login", json=payload, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                if data.get('success') and data.get('access_token'):
                    self.admin_token = data.get('access_token')
                    self.log_test(
                        "Admin Login Authentication",
                        True,
                        f"✅ Admin login successful. Token obtained.",
                        {"status_code": 200, "has_token": True}
                    )
                    return True
                else:
                    self.log_test(
                        "Admin Login Authentication",
                        False,
                        "Login response missing success or access_token",
                        data
                    )
                    return False
            else:
                self.log_test(
                    "Admin Login Authentication",
                    False,
                    f"Admin login failed with status {response.status_code}",
                    {"status_code": response.status_code, "response": response.text[:200]}
                )
                return False
                
        except Exception as e:
            self.log_test(
                "Admin Login Authentication",
                False,
                f"Admin login request failed: {str(e)}",
                {"error": str(e)}
            )
            return False

    def test_statistics_filtering_fixes(self):
        """Test PRIORITY 1: Statistics Filtering Fix - ALL scenarios to confirm SQL syntax errors are resolved"""
        if not self.admin_token:
            self.log_test(
                "Statistics Filtering Tests",
                False,
                "Cannot test statistics endpoints - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test scenarios from the review request
        test_scenarios = [
            {"name": "No filters", "params": {}},
            {"name": "Channel SMS filter", "params": {"channel": "sms"}},
            {"name": "Channel Email filter", "params": {"channel": "email"}},
            {"name": "Provider Fast2SMS filter", "params": {"provider": "fast2sms"}},
            {"name": "Provider SMTP filter", "params": {"provider": "smtp"}},
            {"name": "Combined channel=sms&provider=fast2sms", "params": {"channel": "sms", "provider": "fast2sms"}},
            {"name": "Combined channel=email&provider=smtp", "params": {"channel": "email", "provider": "smtp"}},
            {"name": "Days=7 filter", "params": {"days": "7"}},
            {"name": "Days=30 filter", "params": {"days": "30"}},
            {"name": "Combined channel=sms&provider=fast2sms&days=7", "params": {"channel": "sms", "provider": "fast2sms", "days": "7"}}
        ]
        
        success_count = 0
        total_count = len(test_scenarios)
        
        for scenario in test_scenarios:
            try:
                response = self.session.get(
                    f"{self.api_base}/admin/stats/notifications", 
                    headers=headers, 
                    params=scenario["params"], 
                    timeout=10
                )
                
                if response.status_code == 200:
                    try:
                        data = response.json()
                        if data.get('success') and 'stats' in data:
                            self.log_test(
                                f"Statistics Filtering - {scenario['name']}",
                                True,
                                f"✅ FIXED: Statistics filtering working correctly (status: 200, success: true)",
                                {"status_code": 200, "params": scenario["params"], "has_stats": True}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Statistics Filtering - {scenario['name']}",
                                False,
                                f"❌ Returns 200 but missing success or stats field",
                                {"status_code": 200, "params": scenario["params"], "response": data}
                            )
                    except Exception as json_err:
                        self.log_test(
                            f"Statistics Filtering - {scenario['name']}",
                            False,
                            f"❌ Returns 200 but invalid JSON: {str(json_err)}",
                            {"status_code": 200, "params": scenario["params"], "json_error": str(json_err)}
                        )
                elif response.status_code == 500:
                    # Check if it's the SQL syntax error
                    try:
                        error_data = response.json()
                        error_msg = error_data.get('error', '')
                        if 'syntax error at or near' in error_msg or 'pq:' in error_msg:
                            self.log_test(
                                f"Statistics Filtering - {scenario['name']}",
                                False,
                                f"❌ STILL BROKEN: SQL syntax error not fixed - {error_msg[:100]}",
                                {"status_code": 500, "params": scenario["params"], "sql_error": True}
                            )
                        else:
                            self.log_test(
                                f"Statistics Filtering - {scenario['name']}",
                                False,
                                f"❌ Server error (different from SQL syntax): {error_msg[:100]}",
                                {"status_code": 500, "params": scenario["params"], "error": error_msg}
                            )
                    except:
                        self.log_test(
                            f"Statistics Filtering - {scenario['name']}",
                            False,
                            f"❌ Server error 500 with unparseable response",
                            {"status_code": 500, "params": scenario["params"], "response": response.text[:200]}
                        )
                else:
                    self.log_test(
                        f"Statistics Filtering - {scenario['name']}",
                        False,
                        f"❌ Unexpected status: {response.status_code}",
                        {"status_code": response.status_code, "params": scenario["params"], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Statistics Filtering - {scenario['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "params": scenario["params"]}
                )
        
        # Overall statistics filtering summary
        overall_success = success_count == total_count
        self.log_test(
            "Statistics Filtering Summary",
            overall_success,
            f"{'✅ ALL statistics filtering scenarios working correctly (SQL syntax errors FIXED)' if overall_success else f'❌ {total_count - success_count}/{total_count} statistics filtering scenarios still failing'} (Success rate: {success_count}/{total_count})",
            {"working_count": success_count, "total_count": total_count, "sql_fix_verified": overall_success}
        )
        
        return overall_success

    def test_enhanced_validation_fixes(self):
        """Test PRIORITY 2: Enhanced Validation Fixes - Re-test validation order fixes"""
        if not self.admin_token:
            self.log_test(
                "Enhanced Validation Tests",
                False,
                "Cannot test validation endpoints - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test single notification validation scenarios
        single_validation_tests = [
            {
                "name": "Invalid SMS recipient (too short)",
                "payload": {"channel": "sms", "recipient": "123", "body": "Test message"},
                "expected_error": "invalid phone number"
            },
            {
                "name": "Invalid SMS recipient (bad format)",
                "payload": {"channel": "sms", "recipient": "abc123", "body": "Test message"},
                "expected_error": "phone number should start with"
            },
            {
                "name": "Invalid email recipient (no @)",
                "payload": {"channel": "email", "recipient": "testuser.com", "body": "Test message"},
                "expected_error": "invalid email format"
            },
            {
                "name": "Invalid email recipient (no .)",
                "payload": {"channel": "email", "recipient": "test@user", "body": "Test message"},
                "expected_error": "invalid email format"
            },
            {
                "name": "Invalid push token (too short)",
                "payload": {"channel": "push", "recipient": "abc123", "body": "Test message"},
                "expected_error": "invalid push token"
            }
        ]
        
        success_count = 0
        total_count = len(single_validation_tests)
        
        for test in single_validation_tests:
            try:
                response = self.session.post(
                    f"{self.api_base}/notify/send", 
                    headers=headers, 
                    json=test["payload"], 
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_msg = data.get('error', '').lower()
                        expected_error = test["expected_error"].lower()
                        
                        if expected_error in error_msg:
                            self.log_test(
                                f"Single Validation - {test['name']}",
                                True,
                                f"✅ FIXED: Correct validation error returned - '{data.get('error')}'",
                                {"status_code": 400, "expected": test["expected_error"], "actual": data.get('error')}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Single Validation - {test['name']}",
                                False,
                                f"❌ Wrong validation error - Expected: '{test['expected_error']}', Got: '{data.get('error')}'",
                                {"status_code": 400, "expected": test["expected_error"], "actual": data.get('error')}
                            )
                    except Exception as json_err:
                        self.log_test(
                            f"Single Validation - {test['name']}",
                            False,
                            f"❌ Returns 400 but invalid JSON: {str(json_err)}",
                            {"status_code": 400, "json_error": str(json_err)}
                        )
                else:
                    self.log_test(
                        f"Single Validation - {test['name']}",
                        False,
                        f"❌ Expected 400 validation error, got {response.status_code}",
                        {"status_code": response.status_code, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Single Validation - {test['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e)}
                )
        
        # Test bulk notification validation scenarios
        bulk_validation_tests = [
            {
                "name": "Too many recipients (>1000)",
                "payload": {
                    "channel": "sms", 
                    "recipients": [f"+91900000{i:04d}" for i in range(1001)],  # 1001 recipients
                    "body": "Test message"
                },
                "expected_error": "maximum 1000 recipients allowed"
            },
            {
                "name": "Invalid recipient in bulk list",
                "payload": {
                    "channel": "email", 
                    "recipients": ["valid@email.com", "invalid-email"],
                    "body": "Test message"
                },
                "expected_error": "invalid recipient"
            }
        ]
        
        for test in bulk_validation_tests:
            try:
                response = self.session.post(
                    f"{self.api_base}/notify/bulk", 
                    headers=headers, 
                    json=test["payload"], 
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_msg = data.get('error', '').lower()
                        expected_error = test["expected_error"].lower()
                        
                        if expected_error in error_msg:
                            self.log_test(
                                f"Bulk Validation - {test['name']}",
                                True,
                                f"✅ FIXED: Correct validation error returned - '{data.get('error')}'",
                                {"status_code": 400, "expected": test["expected_error"], "actual": data.get('error')}
                            )
                            success_count += 1
                            total_count += 1
                        else:
                            self.log_test(
                                f"Bulk Validation - {test['name']}",
                                False,
                                f"❌ Wrong validation error - Expected: '{test['expected_error']}', Got: '{data.get('error')}'",
                                {"status_code": 400, "expected": test["expected_error"], "actual": data.get('error')}
                            )
                            total_count += 1
                    except Exception as json_err:
                        self.log_test(
                            f"Bulk Validation - {test['name']}",
                            False,
                            f"❌ Returns 400 but invalid JSON: {str(json_err)}",
                            {"status_code": 400, "json_error": str(json_err)}
                        )
                        total_count += 1
                else:
                    self.log_test(
                        f"Bulk Validation - {test['name']}",
                        False,
                        f"❌ Expected 400 validation error, got {response.status_code}",
                        {"status_code": response.status_code, "response": response.text[:200]}
                    )
                    total_count += 1
                    
            except Exception as e:
                self.log_test(
                    f"Bulk Validation - {test['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e)}
                )
                total_count += 1
        
        # Overall validation summary
        overall_success = success_count == total_count
        self.log_test(
            "Enhanced Validation Summary",
            overall_success,
            f"{'✅ ALL validation scenarios working correctly (validation order FIXED)' if overall_success else f'❌ {total_count - success_count}/{total_count} validation scenarios still failing'} (Success rate: {success_count}/{total_count})",
            {"working_count": success_count, "total_count": total_count, "validation_fix_verified": overall_success}
        )
        
        return overall_success

    def test_working_notification_endpoints(self):
        """Test that previously working notification endpoints remain functional"""
        if not self.admin_token:
            self.log_test(
                "Working Notification Endpoints Test",
                False,
                "Cannot test notification endpoints - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        working_endpoints = [
            {"method": "GET", "path": "/admin/templates", "name": "Get Templates"},
            {"method": "GET", "path": "/admin/config/notifications", "name": "Get Config", "params": {"provider": "smtp", "channel": "email"}}
        ]
        
        success_count = 0
        total_count = len(working_endpoints)
        
        for endpoint in working_endpoints:
            try:
                if endpoint["method"] == "GET":
                    params = endpoint.get("params", {})
                    response = self.session.get(f"{self.api_base}{endpoint['path']}", headers=headers, params=params, timeout=10)
                else:
                    response = self.session.post(f"{self.api_base}{endpoint['path']}", headers=headers, json={}, timeout=10)
                
                if response.status_code in [200, 201]:
                    try:
                        data = response.json()
                        if data.get('success'):
                            self.log_test(
                                f"Working Endpoint - {endpoint['name']}",
                                True,
                                f"✅ {endpoint['name']} working correctly (status: {response.status_code})",
                                {"status_code": response.status_code, "endpoint": endpoint['path']}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Working Endpoint - {endpoint['name']}",
                                False,
                                f"❌ {endpoint['name']} returns success=false",
                                {"status_code": response.status_code, "endpoint": endpoint['path'], "response": data}
                            )
                    except Exception as json_err:
                        self.log_test(
                            f"Working Endpoint - {endpoint['name']}",
                            False,
                            f"❌ {endpoint['name']} returns invalid JSON: {str(json_err)}",
                            {"status_code": response.status_code, "endpoint": endpoint['path'], "json_error": str(json_err)}
                        )
                else:
                    self.log_test(
                        f"Working Endpoint - {endpoint['name']}",
                        False,
                        f"❌ {endpoint['name']} returned unexpected status: {response.status_code}",
                        {"status_code": response.status_code, "endpoint": endpoint['path'], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Working Endpoint - {endpoint['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "endpoint": endpoint['path']}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Working Notification Endpoints Summary",
            overall_success,
            f"{'✅ All previously working endpoints remain functional' if overall_success else f'❌ {total_count - success_count}/{total_count} previously working endpoints failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def run_all_tests(self):
        """Run all tests and generate summary"""
        print("=" * 80)
        print("🧪 NOTIFICATION SYSTEM FIXES TESTING")
        print("Testing GoLang Fantasy Esports backend notification system fixes")
        print("Focus: Re-test notification system fixes after critical backend configuration issue")
        print("=" * 80)
        print()
        
        # Test 1: Health check
        if not self.test_health_check():
            print("❌ Backend is not running. Cannot proceed with tests.")
            return False
        
        # Test 2: Admin authentication
        print("🔍 Testing Admin Authentication")
        if not self.test_admin_login():
            print("❌ Admin authentication failed. Cannot test protected endpoints.")
            return False
        
        # Test 3: PRIORITY 1 - Statistics Filtering Fixes
        print("🔍 Testing PRIORITY 1: Statistics Filtering Fixes")
        self.test_statistics_filtering_fixes()
        
        # Test 4: PRIORITY 2 - Enhanced Validation Fixes
        print("🔍 Testing PRIORITY 2: Enhanced Validation Fixes")
        self.test_enhanced_validation_fixes()
        
        # Test 5: Previously working endpoints
        print("🔍 Testing Previously Working Notification Endpoints")
        self.test_working_notification_endpoints()
        
        # Generate summary
        self.generate_summary()
        
        return True

    def generate_summary(self):
        """Generate test summary"""
        print("=" * 80)
        print("📊 NOTIFICATION SYSTEM FIXES TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result['passed'])
        failed_tests = total_tests - passed_tests
        
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        print()
        
        # Show failed tests
        if failed_tests > 0:
            print("❌ FAILED TESTS:")
            for result in self.test_results:
                if not result['passed']:
                    print(f"  • {result['test']}: {result['details']}")
            print()
        
        # Show critical issues
        critical_issues = [r for r in self.test_results if not r['passed'] and ('SQL' in r['details'] or 'CRITICAL' in r['details'])]
        if critical_issues:
            print("🚨 CRITICAL ISSUES FOUND:")
            for issue in critical_issues:
                print(f"  • {issue['test']}")
                print(f"    {issue['details']}")
            print()
        
        # Show fixed functionality
        fixed_features = [r for r in self.test_results if r['passed'] and 'FIXED' in r['details']]
        if fixed_features:
            print("✅ FIXED FUNCTIONALITY:")
            for feature in fixed_features:
                print(f"  • {feature['test']}")
            print()
        
        # Notification-specific summary
        stats_tests = [r for r in self.test_results if 'statistics' in r['test'].lower()]
        validation_tests = [r for r in self.test_results if 'validation' in r['test'].lower()]
        
        if stats_tests:
            working_stats = sum(1 for r in stats_tests if r['passed'])
            print(f"📊 STATISTICS FILTERING SUMMARY:")
            print(f"  • Total Statistics Tests: {len(stats_tests)}")
            print(f"  • Working Tests: {working_stats}")
            print(f"  • SQL Syntax Error Fix Status: {'VERIFIED - ALL WORKING' if working_stats == len(stats_tests) else 'PARTIAL - SOME STILL FAILING'}")
            print()
        
        if validation_tests:
            working_validation = sum(1 for r in validation_tests if r['passed'])
            print(f"🔍 VALIDATION ORDER SUMMARY:")
            print(f"  • Total Validation Tests: {len(validation_tests)}")
            print(f"  • Working Tests: {working_validation}")
            print(f"  • Validation Order Fix Status: {'VERIFIED - ALL WORKING' if working_validation == len(validation_tests) else 'PARTIAL - SOME STILL FAILING'}")
            print()
        
        # Save results to file
        with open('/app/notification_system_fixes_test_results.json', 'w') as f:
            json.dump({
                'summary': {
                    'total_tests': total_tests,
                    'passed_tests': passed_tests,
                    'failed_tests': failed_tests,
                    'success_rate': f"{(passed_tests/total_tests)*100:.1f}%",
                    'statistics_fix_verified': len(stats_tests) > 0 and working_stats == len(stats_tests) if stats_tests else False,
                    'validation_fix_verified': len(validation_tests) > 0 and working_validation == len(validation_tests) if validation_tests else False
                },
                'test_results': self.test_results
            }, f, indent=2, default=str)
        
        print("📁 Detailed results saved to: /app/notification_system_fixes_test_results.json")
        print("=" * 80)

if __name__ == "__main__":
    tester = NotificationSystemTester()
    tester.run_all_tests()