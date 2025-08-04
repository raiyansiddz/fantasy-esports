#!/usr/bin/env python3
"""
Backend API Testing Script for Fantasy Esports Platform - Notification System Fixes
Testing the notification system fixes that were just implemented:

PRIMARY FOCUS - Test These Critical Fixes:

Issue 1: Statistics Filtering Fix
- Test ALL statistics filtering scenarios to ensure SQL syntax errors are completely resolved
- Test GET /admin/stats/notifications with various filter combinations

Issue 2: Enhanced Validation Fixes  
- Test that recipient validation now happens BEFORE template_id/body validation
- Test single and bulk notification validation scenarios

Expected: 100% success rate on notification system functionality
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
        if response_data and not passed:
            print(f"      Response: {json.dumps(response_data, indent=2)[:300]}...")
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

    def test_statistics_filtering_fix(self):
        """Test Issue 1: Statistics Filtering Fix - ALL scenarios should return 200 instead of 500 SQL errors"""
        if not self.admin_token:
            self.log_test(
                "Statistics Filtering Fix",
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
            {"name": "Combined channel=sms&provider=fast2sms&days=7", "params": {"channel": "sms", "provider": "fast2sms", "days": "7"}},
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
                        if data.get('success'):
                            self.log_test(
                                f"Statistics Filter - {scenario['name']}",
                                True,
                                f"✅ FIXED: Statistics filtering working correctly (status: 200, success: true)",
                                {"status_code": 200, "params": scenario["params"], "has_success": True}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Statistics Filter - {scenario['name']}",
                                False,
                                f"❌ Returns 200 but success=false",
                                {"status_code": 200, "params": scenario["params"], "success": data.get('success')}
                            )
                    except Exception as json_err:
                        self.log_test(
                            f"Statistics Filter - {scenario['name']}",
                            False,
                            f"❌ Returns 200 but invalid JSON: {str(json_err)}",
                            {"status_code": 200, "params": scenario["params"], "json_error": str(json_err)}
                        )
                elif response.status_code == 500:
                    # Check if it's the SQL syntax error
                    try:
                        error_data = response.json()
                        error_msg = error_data.get('error', '')
                        if 'syntax error' in error_msg.lower() or 'pq:' in error_msg:
                            self.log_test(
                                f"Statistics Filter - {scenario['name']}",
                                False,
                                f"❌ STILL BROKEN: SQL syntax error not fixed - {error_msg[:100]}",
                                {"status_code": 500, "params": scenario["params"], "sql_error": True}
                            )
                        else:
                            self.log_test(
                                f"Statistics Filter - {scenario['name']}",
                                False,
                                f"❌ Server error (non-SQL): {error_msg[:100]}",
                                {"status_code": 500, "params": scenario["params"], "error": error_msg}
                            )
                    except:
                        self.log_test(
                            f"Statistics Filter - {scenario['name']}",
                            False,
                            f"❌ Server error with unparseable response",
                            {"status_code": 500, "params": scenario["params"], "response": response.text[:200]}
                        )
                else:
                    self.log_test(
                        f"Statistics Filter - {scenario['name']}",
                        False,
                        f"❌ Unexpected status: {response.status_code}",
                        {"status_code": response.status_code, "params": scenario["params"], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Statistics Filter - {scenario['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "params": scenario["params"]}
                )
        
        # Overall success for Issue 1
        overall_success = success_count == total_count
        self.log_test(
            "Issue 1 - Statistics Filtering Fix Summary",
            overall_success,
            f"{'✅ ALL statistics filtering tests PASSED - SQL syntax errors FIXED' if overall_success else f'❌ {total_count - success_count}/{total_count} statistics filtering tests still failing'} (Success rate: {success_count}/{total_count})",
            {"working_count": success_count, "total_count": total_count, "sql_fix_verified": overall_success}
        )
        
        return overall_success

    def test_enhanced_validation_fix(self):
        """Test Issue 2: Enhanced Validation Fix - Recipient validation should happen BEFORE template validation"""
        if not self.admin_token:
            self.log_test(
                "Enhanced Validation Fix",
                False,
                "Cannot test validation endpoints - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test Single Notification Validation (POST /notify/send)
        single_validation_tests = [
            {
                "name": "Invalid SMS recipient (too short)",
                "payload": {
                    "channel": "sms",
                    "recipient": "123",
                    "body": "Test message"
                },
                "expected_error": "invalid phone number"
            },
            {
                "name": "Invalid SMS recipient (bad format)",
                "payload": {
                    "channel": "sms", 
                    "recipient": "abc123",
                    "body": "Test message"
                },
                "expected_error": "phone number should start with"
            },
            {
                "name": "Invalid email recipient (no @)",
                "payload": {
                    "channel": "email",
                    "recipient": "test.com",
                    "body": "Test message"
                },
                "expected_error": "invalid email format"
            },
            {
                "name": "Invalid email recipient (no .)",
                "payload": {
                    "channel": "email",
                    "recipient": "test@com",
                    "body": "Test message"
                },
                "expected_error": "invalid email format"
            },
            {
                "name": "Invalid push token (too short)",
                "payload": {
                    "channel": "push",
                    "recipient": "abc",
                    "body": "Test message"
                },
                "expected_error": "invalid push token"
            }
        ]
        
        single_success_count = 0
        
        for test in single_validation_tests:
            try:
                response = self.session.post(
                    f"{self.api_base}/admin/notify/send",
                    headers=headers,
                    json=test["payload"],
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_msg = data.get('error', '').lower()
                        
                        if test["expected_error"].lower() in error_msg:
                            self.log_test(
                                f"Single Validation - {test['name']}",
                                True,
                                f"✅ FIXED: Correct recipient validation error returned",
                                {"status_code": 400, "expected_error": test["expected_error"], "got_expected": True}
                            )
                            single_success_count += 1
                        elif "template" in error_msg or "body" in error_msg:
                            self.log_test(
                                f"Single Validation - {test['name']}",
                                False,
                                f"❌ STILL BROKEN: Template/body validation returned instead of recipient validation",
                                {"status_code": 400, "expected_error": test["expected_error"], "got_template_error": True, "error": error_msg}
                            )
                        else:
                            self.log_test(
                                f"Single Validation - {test['name']}",
                                False,
                                f"❌ Unexpected validation error: {error_msg}",
                                {"status_code": 400, "expected_error": test["expected_error"], "unexpected_error": error_msg}
                            )
                    except Exception as json_err:
                        self.log_test(
                            f"Single Validation - {test['name']}",
                            False,
                            f"❌ Invalid JSON response: {str(json_err)}",
                            {"status_code": 400, "json_error": str(json_err)}
                        )
                else:
                    self.log_test(
                        f"Single Validation - {test['name']}",
                        False,
                        f"❌ Expected 400 status, got {response.status_code}",
                        {"status_code": response.status_code, "expected": 400, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Single Validation - {test['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "payload": test["payload"]}
                )
        
        # Test Bulk Notification Validation (POST /notify/bulk)
        bulk_validation_tests = [
            {
                "name": "Too many recipients (>1000)",
                "payload": {
                    "channel": "sms",
                    "recipients": [f"+91987654{i:04d}" for i in range(1001)],  # 1001 recipients
                    "body": "Test message"
                },
                "expected_error": "maximum 1000 recipients allowed"
            },
            {
                "name": "Invalid recipient in bulk list",
                "payload": {
                    "channel": "sms",
                    "recipients": ["+919876543210", "123", "+919876543211"],  # One invalid
                    "body": "Test message"
                },
                "expected_error": "invalid recipient"
            },
            {
                "name": "Invalid email in bulk list",
                "payload": {
                    "channel": "email",
                    "recipients": ["test@example.com", "invalid.email", "test2@example.com"],  # One invalid
                    "body": "Test message"
                },
                "expected_error": "invalid recipient"
            }
        ]
        
        bulk_success_count = 0
        
        for test in bulk_validation_tests:
            try:
                response = self.session.post(
                    f"{self.api_base}/admin/notify/bulk",
                    headers=headers,
                    json=test["payload"],
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_msg = data.get('error', '').lower()
                        
                        if test["expected_error"].lower() in error_msg:
                            self.log_test(
                                f"Bulk Validation - {test['name']}",
                                True,
                                f"✅ FIXED: Correct bulk validation error returned",
                                {"status_code": 400, "expected_error": test["expected_error"], "got_expected": True}
                            )
                            bulk_success_count += 1
                        elif "template" in error_msg or "body" in error_msg:
                            self.log_test(
                                f"Bulk Validation - {test['name']}",
                                False,
                                f"❌ STILL BROKEN: Template/body validation returned instead of recipient validation",
                                {"status_code": 400, "expected_error": test["expected_error"], "got_template_error": True, "error": error_msg}
                            )
                        else:
                            self.log_test(
                                f"Bulk Validation - {test['name']}",
                                False,
                                f"❌ Unexpected validation error: {error_msg}",
                                {"status_code": 400, "expected_error": test["expected_error"], "unexpected_error": error_msg}
                            )
                    except Exception as json_err:
                        self.log_test(
                            f"Bulk Validation - {test['name']}",
                            False,
                            f"❌ Invalid JSON response: {str(json_err)}",
                            {"status_code": 400, "json_error": str(json_err)}
                        )
                else:
                    self.log_test(
                        f"Bulk Validation - {test['name']}",
                        False,
                        f"❌ Expected 400 status, got {response.status_code}",
                        {"status_code": response.status_code, "expected": 400, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Bulk Validation - {test['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "payload": test["payload"]}
                )
        
        # Overall success for Issue 2
        total_validation_tests = len(single_validation_tests) + len(bulk_validation_tests)
        total_validation_success = single_success_count + bulk_success_count
        overall_validation_success = total_validation_success == total_validation_tests
        
        self.log_test(
            "Issue 2 - Enhanced Validation Fix Summary",
            overall_validation_success,
            f"{'✅ ALL validation tests PASSED - Recipient validation happens BEFORE template validation' if overall_validation_success else f'❌ {total_validation_tests - total_validation_success}/{total_validation_tests} validation tests still failing'} (Single: {single_success_count}/{len(single_validation_tests)}, Bulk: {bulk_success_count}/{len(bulk_validation_tests)})",
            {"single_success": single_success_count, "bulk_success": bulk_success_count, "total_success": total_validation_success, "total_tests": total_validation_tests, "validation_fix_verified": overall_validation_success}
        )
        
        return overall_validation_success

    def test_working_notification_endpoints(self):
        """Test that previously working notification functionality remains intact"""
        if not self.admin_token:
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        working_endpoints = [
            {
                "method": "GET",
                "path": "/admin/templates",
                "name": "Get Templates"
            },
            {
                "method": "GET", 
                "path": "/admin/config/notifications",
                "params": {"provider": "fast2sms", "channel": "sms"},
                "name": "Get Notification Config"
            }
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
                    self.log_test(
                        f"Working Endpoint - {endpoint['name']}",
                        True,
                        f"✅ {endpoint['name']} still working correctly (status: {response.status_code})",
                        {"status_code": response.status_code, "endpoint": endpoint['path']}
                    )
                    success_count += 1
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
            "Previously Working Functionality Summary",
            overall_success,
            f"{'✅ All previously working endpoints still functional' if overall_success else f'❌ {total_count - success_count}/{total_count} previously working endpoints now broken'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def run_all_tests(self):
        """Run all notification system fix tests and generate summary"""
        print("=" * 80)
        print("🧪 NOTIFICATION SYSTEM FIXES TESTING")
        print("Testing the notification system fixes that were just implemented")
        print("Focus: Statistics Filtering Fix + Enhanced Validation Fix")
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
        
        # Test 3: Issue 1 - Statistics Filtering Fix
        print("🔍 Testing Issue 1 - Statistics Filtering Fix")
        stats_fix_success = self.test_statistics_filtering_fix()
        
        # Test 4: Issue 2 - Enhanced Validation Fix
        print("🔍 Testing Issue 2 - Enhanced Validation Fix")
        validation_fix_success = self.test_enhanced_validation_fix()
        
        # Test 5: Previously working functionality
        print("🔍 Testing Previously Working Functionality")
        working_functionality_success = self.test_working_notification_endpoints()
        
        # Generate summary
        self.generate_summary(stats_fix_success, validation_fix_success, working_functionality_success)
        
        return stats_fix_success and validation_fix_success

    def generate_summary(self, stats_fix_success: bool, validation_fix_success: bool, working_functionality_success: bool):
        """Generate comprehensive test summary"""
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
        
        # Critical Issues Summary
        print("🎯 CRITICAL FIXES STATUS:")
        print(f"  • Issue 1 - Statistics Filtering Fix: {'✅ FIXED' if stats_fix_success else '❌ STILL BROKEN'}")
        print(f"  • Issue 2 - Enhanced Validation Fix: {'✅ FIXED' if validation_fix_success else '❌ STILL BROKEN'}")
        print(f"  • Previously Working Functionality: {'✅ INTACT' if working_functionality_success else '❌ BROKEN'}")
        print()
        
        # Show failed tests if any
        if failed_tests > 0:
            print("❌ FAILED TESTS:")
            for result in self.test_results:
                if not result['passed']:
                    print(f"  • {result['test']}: {result['details']}")
            print()
        
        # Show critical issues
        critical_issues = [r for r in self.test_results if not r['passed'] and ('STILL BROKEN' in r['details'] or 'SQL syntax error' in r['details'])]
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
        
        # Final verdict
        both_fixes_working = stats_fix_success and validation_fix_success
        print(f"🎯 FINAL VERDICT:")
        if both_fixes_working:
            print("✅ NOTIFICATION SYSTEM FIXES SUCCESSFULLY IMPLEMENTED!")
            print("   Both critical issues have been resolved:")
            print("   - Statistics filtering SQL syntax errors FIXED")
            print("   - Enhanced validation order FIXED")
        else:
            print("❌ NOTIFICATION SYSTEM FIXES NOT FULLY IMPLEMENTED")
            if not stats_fix_success:
                print("   - Statistics filtering still has SQL syntax errors")
            if not validation_fix_success:
                print("   - Validation order still incorrect (template validation before recipient validation)")
        print()
        
        # Save results to file
        with open('/app/notification_fixes_test_results.json', 'w') as f:
            json.dump({
                'summary': {
                    'total_tests': total_tests,
                    'passed_tests': passed_tests,
                    'failed_tests': failed_tests,
                    'success_rate': f"{(passed_tests/total_tests)*100:.1f}%",
                    'stats_fix_success': stats_fix_success,
                    'validation_fix_success': validation_fix_success,
                    'working_functionality_success': working_functionality_success,
                    'both_critical_fixes_working': both_fixes_working
                },
                'test_results': self.test_results
            }, f, indent=2, default=str)
        
        print("📁 Detailed results saved to: /app/notification_fixes_test_results.json")
        print("=" * 80)

if __name__ == "__main__":
    tester = NotificationSystemTester()
    tester.run_all_tests()