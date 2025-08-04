#!/usr/bin/env python3
"""
Notification System Fixes Testing Script for Fantasy Esports Platform
Testing the specific notification system fixes mentioned in the review request:

FOCUS: Test the two specific issues that were previously failing:
1. Statistics Issue Fix: GET /api/v1/admin/stats/notifications with filtering
2. Edge Cases & Validation Fix: Enhanced validation for notification endpoints

Expected Results:
- All validation should return proper 400 status codes with descriptive error messages
- Statistics filtering should work without SQL errors
- Overall success rate should improve to 100% or very close to it
"""

import requests
import json
import sys
import time
from typing import Dict, Any, List, Optional

class NotificationFixesTester:
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
        if response_data and len(str(response_data)) < 300:
            print(f"      Response: {json.dumps(response_data, indent=2)}")
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
                        f"Admin login successful. Token obtained.",
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

    def test_notification_stats_filtering(self):
        """Test Issue 1: Statistics filtering with query parameters"""
        if not self.admin_token:
            self.log_test(
                "Notification Stats Filtering",
                False,
                "Cannot test - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test various combinations of channel and provider filters
        test_cases = [
            {"params": {}, "name": "No filters"},
            {"params": {"channel": "sms"}, "name": "SMS channel filter"},
            {"params": {"channel": "email"}, "name": "Email channel filter"},
            {"params": {"provider": "fast2sms"}, "name": "Fast2SMS provider filter"},
            {"params": {"provider": "smtp"}, "name": "SMTP provider filter"},
            {"params": {"channel": "sms", "provider": "fast2sms"}, "name": "SMS + Fast2SMS filter"},
            {"params": {"channel": "email", "provider": "smtp"}, "name": "Email + SMTP filter"},
            {"params": {"days": "7"}, "name": "7 days filter"},
            {"params": {"days": "30"}, "name": "30 days filter"},
            {"params": {"channel": "sms", "provider": "fast2sms", "days": "7"}, "name": "Combined filters"},
        ]
        
        success_count = 0
        total_count = len(test_cases)
        
        for test_case in test_cases:
            try:
                response = self.session.get(
                    f"{self.api_base}/admin/stats/notifications", 
                    headers=headers, 
                    params=test_case["params"],
                    timeout=10
                )
                
                if response.status_code == 200:
                    try:
                        data = response.json()
                        if data.get('success') and 'stats' in data:
                            stats = data['stats']
                            # Validate stats structure
                            required_fields = ['total_sent', 'total_delivered', 'total_failed', 'total_pending', 'delivery_rate', 'failure_rate']
                            has_all_fields = all(field in stats for field in required_fields)
                            
                            if has_all_fields:
                                self.log_test(
                                    f"Stats Filtering - {test_case['name']}",
                                    True,
                                    f"✅ FIXED: Statistics endpoint working with {test_case['name']} - SQL query fix successful",
                                    {"status_code": 200, "params": test_case["params"], "stats_fields": list(stats.keys())}
                                )
                                success_count += 1
                            else:
                                self.log_test(
                                    f"Stats Filtering - {test_case['name']}",
                                    False,
                                    f"Response missing required stats fields. Got: {list(stats.keys())}",
                                    {"status_code": 200, "params": test_case["params"], "missing_fields": [f for f in required_fields if f not in stats]}
                                )
                        else:
                            self.log_test(
                                f"Stats Filtering - {test_case['name']}",
                                False,
                                f"Response missing success or stats field",
                                {"status_code": 200, "params": test_case["params"], "response_keys": list(data.keys())}
                            )
                    except json.JSONDecodeError as e:
                        self.log_test(
                            f"Stats Filtering - {test_case['name']}",
                            False,
                            f"Invalid JSON response: {str(e)}",
                            {"status_code": 200, "params": test_case["params"], "json_error": str(e)}
                        )
                elif response.status_code == 500:
                    self.log_test(
                        f"Stats Filtering - {test_case['name']}",
                        False,
                        f"❌ SQL ERROR: Statistics endpoint still failing with 500 - SQL query fix not working",
                        {"status_code": 500, "params": test_case["params"], "response": response.text[:200]}
                    )
                else:
                    self.log_test(
                        f"Stats Filtering - {test_case['name']}",
                        False,
                        f"Unexpected status code: {response.status_code}",
                        {"status_code": response.status_code, "params": test_case["params"], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Stats Filtering - {test_case['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "params": test_case["params"]}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Statistics Filtering Summary",
            overall_success,
            f"{'✅ All statistics filtering tests passed - SQL query fix verified' if overall_success else f'❌ {total_count - success_count}/{total_count} statistics filtering tests failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def test_send_notification_validation(self):
        """Test Issue 2: Enhanced validation for send notification endpoint"""
        if not self.admin_token:
            self.log_test(
                "Send Notification Validation",
                False,
                "Cannot test - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test cases for validation errors (should return 400)
        validation_test_cases = [
            {
                "payload": {},
                "name": "Missing required fields",
                "expected_error": "recipient"
            },
            {
                "payload": {"channel": "sms"},
                "name": "Missing recipient",
                "expected_error": "recipient"
            },
            {
                "payload": {"recipient": "+919876543210"},
                "name": "Missing channel",
                "expected_error": "channel"
            },
            {
                "payload": {"channel": "invalid_channel", "recipient": "+919876543210"},
                "name": "Invalid channel",
                "expected_error": "Invalid channel"
            },
            {
                "payload": {"channel": "sms", "recipient": "123"},
                "name": "Invalid SMS recipient (too short)",
                "expected_error": "invalid phone number"
            },
            {
                "payload": {"channel": "sms", "recipient": "abc123"},
                "name": "Invalid SMS recipient (bad format)",
                "expected_error": "phone number should start with"
            },
            {
                "payload": {"channel": "email", "recipient": "invalid-email"},
                "name": "Invalid email recipient (no @)",
                "expected_error": "invalid email format"
            },
            {
                "payload": {"channel": "email", "recipient": "test@invalid"},
                "name": "Invalid email recipient (no .)",
                "expected_error": "invalid email format"
            },
            {
                "payload": {"channel": "push", "recipient": "short"},
                "name": "Invalid push token (too short)",
                "expected_error": "invalid push token"
            },
            {
                "payload": {"channel": "sms", "recipient": "+919876543210"},
                "name": "Missing template_id and body",
                "expected_error": "Either template_id or body must be provided"
            }
        ]
        
        success_count = 0
        total_count = len(validation_test_cases)
        
        for test_case in validation_test_cases:
            try:
                response = self.session.post(
                    f"{self.api_base}/admin/notify/send",
                    headers=headers,
                    json=test_case["payload"],
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_message = data.get('error', '').lower()
                        expected_error = test_case['expected_error'].lower()
                        
                        if expected_error in error_message:
                            self.log_test(
                                f"Send Validation - {test_case['name']}",
                                True,
                                f"✅ FIXED: Proper validation error returned for {test_case['name']}",
                                {"status_code": 400, "error": data.get('error'), "payload": test_case["payload"]}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Send Validation - {test_case['name']}",
                                False,
                                f"Wrong error message. Expected '{expected_error}' in '{error_message}'",
                                {"status_code": 400, "error": data.get('error'), "expected": expected_error}
                            )
                    except json.JSONDecodeError:
                        self.log_test(
                            f"Send Validation - {test_case['name']}",
                            False,
                            f"400 status but invalid JSON response",
                            {"status_code": 400, "response": response.text[:200]}
                        )
                else:
                    self.log_test(
                        f"Send Validation - {test_case['name']}",
                        False,
                        f"❌ VALIDATION NOT WORKING: Expected 400, got {response.status_code}",
                        {"status_code": response.status_code, "payload": test_case["payload"], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Send Validation - {test_case['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "payload": test_case["payload"]}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Send Notification Validation Summary",
            overall_success,
            f"{'✅ All send notification validation tests passed - Enhanced validation working' if overall_success else f'❌ {total_count - success_count}/{total_count} validation tests failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def test_bulk_notification_validation(self):
        """Test Issue 2: Enhanced validation for bulk notification endpoint"""
        if not self.admin_token:
            self.log_test(
                "Bulk Notification Validation",
                False,
                "Cannot test - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test cases for bulk validation errors
        bulk_validation_test_cases = [
            {
                "payload": {"channel": "sms"},
                "name": "Empty recipients and no user filter",
                "expected_error": "Recipients list or user filter is required"
            },
            {
                "payload": {"channel": "sms", "recipients": ["+" + "1234567890" * 100] * 1001},
                "name": "Too many recipients (>1000)",
                "expected_error": "Maximum 1000 recipients allowed"
            },
            {
                "payload": {"channel": "sms", "recipients": ["+919876543210", "invalid-phone"]},
                "name": "Invalid recipient in bulk list",
                "expected_error": "Invalid recipient"
            },
            {
                "payload": {"channel": "email", "recipients": ["test@example.com", "invalid-email"]},
                "name": "Invalid email in bulk list",
                "expected_error": "Invalid recipient"
            }
        ]
        
        success_count = 0
        total_count = len(bulk_validation_test_cases)
        
        for test_case in bulk_validation_test_cases:
            try:
                response = self.session.post(
                    f"{self.api_base}/admin/notify/bulk",
                    headers=headers,
                    json=test_case["payload"],
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_message = data.get('error', '').lower()
                        expected_error = test_case['expected_error'].lower()
                        
                        if expected_error in error_message:
                            self.log_test(
                                f"Bulk Validation - {test_case['name']}",
                                True,
                                f"✅ FIXED: Proper validation error returned for {test_case['name']}",
                                {"status_code": 400, "error": data.get('error')}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Bulk Validation - {test_case['name']}",
                                False,
                                f"Wrong error message. Expected '{expected_error}' in '{error_message}'",
                                {"status_code": 400, "error": data.get('error'), "expected": expected_error}
                            )
                    except json.JSONDecodeError:
                        self.log_test(
                            f"Bulk Validation - {test_case['name']}",
                            False,
                            f"400 status but invalid JSON response",
                            {"status_code": 400, "response": response.text[:200]}
                        )
                else:
                    self.log_test(
                        f"Bulk Validation - {test_case['name']}",
                        False,
                        f"❌ VALIDATION NOT WORKING: Expected 400, got {response.status_code}",
                        {"status_code": response.status_code, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Bulk Validation - {test_case['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e)}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Bulk Notification Validation Summary",
            overall_success,
            f"{'✅ All bulk notification validation tests passed' if overall_success else f'❌ {total_count - success_count}/{total_count} bulk validation tests failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def test_template_validation(self):
        """Test Issue 2: Enhanced validation for template creation endpoint"""
        if not self.admin_token:
            self.log_test(
                "Template Validation",
                False,
                "Cannot test - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test cases for template validation errors
        template_validation_test_cases = [
            {
                "payload": {},
                "name": "Missing required fields",
                "expected_error": "name is required"
            },
            {
                "payload": {"name": "Test Template"},
                "name": "Missing body",
                "expected_error": "body is required"
            },
            {
                "payload": {"name": "Test Template", "body": "Test body"},
                "name": "Missing channel",
                "expected_error": "channel"
            },
            {
                "payload": {"name": "Test Template", "body": "Test body", "channel": "invalid_channel"},
                "name": "Invalid channel",
                "expected_error": "Invalid channel"
            },
            {
                "payload": {"name": "Test Template", "body": "Test body", "channel": "sms", "provider": "invalid_provider"},
                "name": "Invalid provider",
                "expected_error": "Invalid provider"
            },
            {
                "payload": {"name": "A" * 201, "body": "Test body", "channel": "sms", "provider": "fast2sms"},
                "name": "Name too long (>200 chars)",
                "expected_error": "must not exceed 200 characters"
            },
            {
                "payload": {"name": "Test Template", "body": "Test body", "channel": "email", "provider": "smtp", "subject": "A" * 501},
                "name": "Subject too long (>500 chars)",
                "expected_error": "must not exceed 500 characters"
            }
        ]
        
        success_count = 0
        total_count = len(template_validation_test_cases)
        
        for test_case in template_validation_test_cases:
            try:
                response = self.session.post(
                    f"{self.api_base}/admin/templates",
                    headers=headers,
                    json=test_case["payload"],
                    timeout=10
                )
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_message = data.get('error', '').lower()
                        expected_error = test_case['expected_error'].lower()
                        
                        if expected_error in error_message:
                            self.log_test(
                                f"Template Validation - {test_case['name']}",
                                True,
                                f"✅ FIXED: Proper validation error returned for {test_case['name']}",
                                {"status_code": 400, "error": data.get('error')}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Template Validation - {test_case['name']}",
                                False,
                                f"Wrong error message. Expected '{expected_error}' in '{error_message}'",
                                {"status_code": 400, "error": data.get('error'), "expected": expected_error}
                            )
                    except json.JSONDecodeError:
                        self.log_test(
                            f"Template Validation - {test_case['name']}",
                            False,
                            f"400 status but invalid JSON response",
                            {"status_code": 400, "response": response.text[:200]}
                        )
                else:
                    self.log_test(
                        f"Template Validation - {test_case['name']}",
                        False,
                        f"❌ VALIDATION NOT WORKING: Expected 400, got {response.status_code}",
                        {"status_code": response.status_code, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Template Validation - {test_case['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e)}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Template Validation Summary",
            overall_success,
            f"{'✅ All template validation tests passed' if overall_success else f'❌ {total_count - success_count}/{total_count} template validation tests failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def test_valid_notification_scenarios(self):
        """Test that valid notification requests work correctly"""
        if not self.admin_token:
            self.log_test(
                "Valid Notification Scenarios",
                False,
                "Cannot test - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        # Test valid scenarios (should return 200 or 201)
        valid_test_cases = [
            {
                "endpoint": "/admin/notify/send",
                "payload": {"channel": "sms", "recipient": "+919876543210", "body": "Test SMS message"},
                "name": "Valid SMS notification"
            },
            {
                "endpoint": "/admin/notify/send",
                "payload": {"channel": "email", "recipient": "test@example.com", "subject": "Test", "body": "Test email message"},
                "name": "Valid email notification"
            },
            {
                "endpoint": "/admin/templates",
                "payload": {"name": "Test Template", "channel": "sms", "provider": "fast2sms", "body": "Hello {name}!"},
                "name": "Valid template creation"
            }
        ]
        
        success_count = 0
        total_count = len(valid_test_cases)
        
        for test_case in valid_test_cases:
            try:
                response = self.session.post(
                    f"{self.api_base}{test_case['endpoint']}",
                    headers=headers,
                    json=test_case["payload"],
                    timeout=10
                )
                
                if response.status_code in [200, 201]:
                    try:
                        data = response.json()
                        if data.get('success') or 'id' in data:
                            self.log_test(
                                f"Valid Scenario - {test_case['name']}",
                                True,
                                f"✅ Valid request processed successfully",
                                {"status_code": response.status_code, "endpoint": test_case['endpoint']}
                            )
                            success_count += 1
                        else:
                            self.log_test(
                                f"Valid Scenario - {test_case['name']}",
                                False,
                                f"Success status but response indicates failure",
                                {"status_code": response.status_code, "response": data}
                            )
                    except json.JSONDecodeError:
                        self.log_test(
                            f"Valid Scenario - {test_case['name']}",
                            False,
                            f"Success status but invalid JSON response",
                            {"status_code": response.status_code, "response": response.text[:200]}
                        )
                else:
                    self.log_test(
                        f"Valid Scenario - {test_case['name']}",
                        False,
                        f"Valid request failed with status {response.status_code}",
                        {"status_code": response.status_code, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Valid Scenario - {test_case['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e)}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Valid Notification Scenarios Summary",
            overall_success,
            f"{'✅ All valid notification scenarios work correctly' if overall_success else f'❌ {total_count - success_count}/{total_count} valid scenarios failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def run_all_tests(self):
        """Run all notification system tests and generate summary"""
        print("=" * 80)
        print("🧪 NOTIFICATION SYSTEM FIXES TESTING")
        print("Testing notification system fixes as requested in review")
        print("Focus: Statistics filtering fix + Enhanced validation fixes")
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
        
        # Test 3: Statistics filtering (Issue 1)
        print("🔍 Testing Issue 1: Statistics Filtering Fix")
        self.test_notification_stats_filtering()
        
        # Test 4: Enhanced validation (Issue 2)
        print("🔍 Testing Issue 2: Enhanced Validation Fixes")
        self.test_send_notification_validation()
        self.test_bulk_notification_validation()
        self.test_template_validation()
        
        # Test 5: Valid scenarios still work
        print("🔍 Testing Valid Notification Scenarios")
        self.test_valid_notification_scenarios()
        
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
        
        # Issue-specific analysis
        stats_tests = [r for r in self.test_results if 'stats' in r['test'].lower() and 'filtering' in r['test'].lower()]
        validation_tests = [r for r in self.test_results if 'validation' in r['test'].lower()]
        
        if stats_tests:
            stats_passed = sum(1 for r in stats_tests if r['passed'])
            print(f"📊 ISSUE 1 - Statistics Filtering Fix:")
            print(f"  • Tests: {len(stats_tests)}")
            print(f"  • Passed: {stats_passed}")
            print(f"  • Status: {'✅ FIXED' if stats_passed == len(stats_tests) else '❌ STILL BROKEN'}")
            print()
        
        if validation_tests:
            validation_passed = sum(1 for r in validation_tests if r['passed'])
            print(f"🔒 ISSUE 2 - Enhanced Validation Fix:")
            print(f"  • Tests: {len(validation_tests)}")
            print(f"  • Passed: {validation_passed}")
            print(f"  • Status: {'✅ FIXED' if validation_passed == len(validation_tests) else '❌ STILL BROKEN'}")
            print()
        
        # Show failed tests
        if failed_tests > 0:
            print("❌ FAILED TESTS:")
            for result in self.test_results:
                if not result['passed']:
                    print(f"  • {result['test']}: {result['details']}")
            print()
        
        # Overall assessment
        success_rate = (passed_tests/total_tests)*100
        if success_rate >= 95:
            print("🎉 OVERALL ASSESSMENT: EXCELLENT - Notification system fixes are working correctly!")
        elif success_rate >= 80:
            print("✅ OVERALL ASSESSMENT: GOOD - Most notification system fixes are working")
        elif success_rate >= 60:
            print("⚠️ OVERALL ASSESSMENT: PARTIAL - Some notification system fixes need attention")
        else:
            print("❌ OVERALL ASSESSMENT: POOR - Notification system fixes are not working properly")
        
        # Save results to file
        with open('/app/notification_fixes_test_results.json', 'w') as f:
            json.dump({
                'summary': {
                    'total_tests': total_tests,
                    'passed_tests': passed_tests,
                    'failed_tests': failed_tests,
                    'success_rate': f"{success_rate:.1f}%",
                    'issue_1_stats_fix': len(stats_tests) > 0 and sum(1 for r in stats_tests if r['passed']) == len(stats_tests),
                    'issue_2_validation_fix': len(validation_tests) > 0 and sum(1 for r in validation_tests if r['passed']) == len(validation_tests)
                },
                'test_results': self.test_results
            }, f, indent=2, default=str)
        
        print("📁 Detailed results saved to: /app/notification_fixes_test_results.json")
        print("=" * 80)

if __name__ == "__main__":
    tester = NotificationFixesTester()
    tester.run_all_tests()