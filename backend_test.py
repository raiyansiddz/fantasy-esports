#!/usr/bin/env python3
"""
Backend Test Suite for Notification System Fixes
Testing the 4 specific notification validation scenarios requested.
"""

import requests
import json
import time
import sys
from typing import Dict, Any, List

class NotificationTestSuite:
    def __init__(self):
        self.base_url = "http://localhost:8001/api/v1"
        self.admin_token = None
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = ""):
        """Log test result"""
        status = "✅ PASS" if success else "❌ FAIL"
        result = {
            "test": test_name,
            "success": success,
            "details": details,
            "status": status
        }
        self.test_results.append(result)
        print(f"{status}: {test_name}")
        if details:
            print(f"   Details: {details}")
        print()

    def get_admin_token(self) -> bool:
        """Get admin authentication token"""
        try:
            # Try to login as admin
            login_data = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = requests.post(f"{self.base_url}/admin/login", json=login_data, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                if 'token' in data:
                    self.admin_token = data['token']
                    self.log_test("Admin Authentication", True, "Successfully obtained admin token")
                    return True
                elif 'access_token' in data:
                    self.admin_token = data['access_token']
                    self.log_test("Admin Authentication", True, "Successfully obtained admin token")
                    return True
            
            self.log_test("Admin Authentication", False, f"Status: {response.status_code}, Response: {response.text[:200]}")
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def get_headers(self) -> Dict[str, str]:
        """Get headers with admin authentication"""
        return {
            "Authorization": f"Bearer {self.admin_token}",
            "Content-Type": "application/json"
        }

    def test_sms_validation_invalid_recipient(self):
        """Test SMS validation with recipient 'abc123' - should return error containing 'phone number should start with'"""
        try:
            payload = {
                "channel": "sms",
                "recipient": "abc123",
                "body": "Test SMS message"
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/send",
                json=payload,
                headers=self.get_headers(),
                timeout=10
            )
            
            if response.status_code == 400:
                response_data = response.json()
                error_message = response_data.get('error', '').lower()
                
                if 'phone number should start with' in error_message:
                    self.log_test("SMS Validation - Invalid Recipient (abc123)", True, 
                                f"Correctly returned 400 with expected error message: {response_data.get('error')}")
                    return True
                else:
                    self.log_test("SMS Validation - Invalid Recipient (abc123)", False, 
                                f"Got 400 but wrong error message. Expected 'phone number should start with', got: {response_data.get('error')}")
                    return False
            else:
                self.log_test("SMS Validation - Invalid Recipient (abc123)", False, 
                            f"Expected 400 status, got {response.status_code}. Response: {response.text[:200]}")
                return False
                
        except Exception as e:
            self.log_test("SMS Validation - Invalid Recipient (abc123)", False, f"Exception: {str(e)}")
            return False

    def test_bulk_notification_max_recipients(self):
        """Test bulk notification with 1001 recipients - should return 400 error with 'maximum 1000 recipients allowed'"""
        try:
            # Create 1001 recipients
            recipients = [f"+91987654{str(i).zfill(4)}" for i in range(1001)]
            
            payload = {
                "channel": "sms",
                "recipients": recipients,
                "body": "Test bulk SMS message"
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/bulk",
                json=payload,
                headers=self.get_headers(),
                timeout=15
            )
            
            if response.status_code == 400:
                response_data = response.json()
                error_message = response_data.get('error', '').lower()
                
                if 'maximum 1000 recipients allowed' in error_message:
                    self.log_test("Bulk Notification - Max Recipients (1001)", True, 
                                f"Correctly returned 400 with expected error message: {response_data.get('error')}")
                    return True
                else:
                    self.log_test("Bulk Notification - Max Recipients (1001)", False, 
                                f"Got 400 but wrong error message. Expected 'maximum 1000 recipients allowed', got: {response_data.get('error')}")
                    return False
            else:
                self.log_test("Bulk Notification - Max Recipients (1001)", False, 
                            f"Expected 400 status, got {response.status_code}. Response: {response.text[:200]}")
                return False
                
        except Exception as e:
            self.log_test("Bulk Notification - Max Recipients (1001)", False, f"Exception: {str(e)}")
            return False

    def test_bulk_notification_invalid_recipient_in_list(self):
        """Test bulk notification with invalid recipient in list - should return 400 error with 'invalid recipient'"""
        try:
            # Mix of valid and invalid recipients
            recipients = [
                "+919876543210",  # Valid
                "+919876543211",  # Valid
                "abc123",         # Invalid - should trigger error
                "+919876543212"   # Valid
            ]
            
            payload = {
                "channel": "sms",
                "recipients": recipients,
                "body": "Test bulk SMS message"
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/bulk",
                json=payload,
                headers=self.get_headers(),
                timeout=10
            )
            
            if response.status_code == 400:
                response_data = response.json()
                error_message = response_data.get('error', '').lower()
                
                if 'invalid recipient' in error_message and 'abc123' in error_message:
                    self.log_test("Bulk Notification - Invalid Recipient in List", True, 
                                f"Correctly returned 400 with expected error message: {response_data.get('error')}")
                    return True
                else:
                    self.log_test("Bulk Notification - Invalid Recipient in List", False, 
                                f"Got 400 but wrong error message. Expected 'invalid recipient' with 'abc123', got: {response_data.get('error')}")
                    return False
            else:
                self.log_test("Bulk Notification - Invalid Recipient in List", False, 
                            f"Expected 400 status, got {response.status_code}. Response: {response.text[:200]}")
                return False
                
        except Exception as e:
            self.log_test("Bulk Notification - Invalid Recipient in List", False, f"Exception: {str(e)}")
            return False

    def test_additional_validation_scenarios(self):
        """Test additional notification validation scenarios to ensure system is working properly"""
        additional_tests = []
        
        # Test 1: Valid single SMS notification
        try:
            payload = {
                "channel": "sms",
                "recipient": "+919876543210",
                "body": "Test SMS message"
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/send",
                json=payload,
                headers=self.get_headers(),
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                additional_tests.append(("Valid Single SMS", True, "Successfully sent SMS notification"))
            else:
                additional_tests.append(("Valid Single SMS", False, f"Status: {response.status_code}, Response: {response.text[:100]}"))
                
        except Exception as e:
            additional_tests.append(("Valid Single SMS", False, f"Exception: {str(e)}"))

        # Test 2: Valid bulk SMS notification (small batch)
        try:
            recipients = ["+919876543210", "+919876543211", "+919876543212"]
            payload = {
                "channel": "sms",
                "recipients": recipients,
                "body": "Test bulk SMS message"
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/bulk",
                json=payload,
                headers=self.get_headers(),
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                additional_tests.append(("Valid Bulk SMS (3 recipients)", True, "Successfully sent bulk SMS notification"))
            else:
                additional_tests.append(("Valid Bulk SMS (3 recipients)", False, f"Status: {response.status_code}, Response: {response.text[:100]}"))
                
        except Exception as e:
            additional_tests.append(("Valid Bulk SMS (3 recipients)", False, f"Exception: {str(e)}"))

        # Test 3: Missing body validation
        try:
            payload = {
                "channel": "sms",
                "recipient": "+919876543210"
                # Missing body
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/send",
                json=payload,
                headers=self.get_headers(),
                timeout=10
            )
            
            if response.status_code == 400:
                response_data = response.json()
                error_message = response_data.get('error', '').lower()
                if 'body' in error_message or 'template' in error_message:
                    additional_tests.append(("Missing Body Validation", True, f"Correctly rejected missing body: {response_data.get('error')}"))
                else:
                    additional_tests.append(("Missing Body Validation", False, f"Got 400 but unexpected error: {response_data.get('error')}"))
            else:
                additional_tests.append(("Missing Body Validation", False, f"Expected 400, got {response.status_code}"))
                
        except Exception as e:
            additional_tests.append(("Missing Body Validation", False, f"Exception: {str(e)}"))

        # Test 4: Invalid channel validation
        try:
            payload = {
                "channel": "invalid_channel",
                "recipient": "+919876543210",
                "body": "Test message"
            }
            
            response = requests.post(
                f"{self.base_url}/admin/notify/send",
                json=payload,
                headers=self.get_headers(),
                timeout=10
            )
            
            if response.status_code == 400:
                response_data = response.json()
                error_message = response_data.get('error', '').lower()
                if 'channel' in error_message or 'invalid' in error_message:
                    additional_tests.append(("Invalid Channel Validation", True, f"Correctly rejected invalid channel: {response_data.get('error')}"))
                else:
                    additional_tests.append(("Invalid Channel Validation", False, f"Got 400 but unexpected error: {response_data.get('error')}"))
            else:
                additional_tests.append(("Invalid Channel Validation", False, f"Expected 400, got {response.status_code}"))
                
        except Exception as e:
            additional_tests.append(("Invalid Channel Validation", False, f"Exception: {str(e)}"))

        # Log all additional test results
        for test_name, success, details in additional_tests:
            self.log_test(f"Additional Validation - {test_name}", success, details)
        
        return sum(1 for _, success, _ in additional_tests if success)

    def run_all_tests(self):
        """Run all notification system tests"""
        print("🚀 Starting Notification System Fixes Testing")
        print("=" * 60)
        
        # Get admin token first
        if not self.get_admin_token():
            print("❌ Cannot proceed without admin authentication")
            return False
        
        print("🎯 Testing 4 Specific Notification System Fixes:")
        print("-" * 50)
        
        # Test the 4 specific scenarios requested
        test1_result = self.test_sms_validation_invalid_recipient()
        test2_result = self.test_bulk_notification_max_recipients()
        test3_result = self.test_bulk_notification_invalid_recipient_in_list()
        
        print("🔍 Additional Validation Scenarios:")
        print("-" * 40)
        additional_passed = self.test_additional_validation_scenarios()
        
        # Calculate results
        main_tests_passed = sum([test1_result, test2_result, test3_result])
        total_main_tests = 3
        total_additional_tests = 4
        total_tests = len(self.test_results)
        total_passed = sum(1 for result in self.test_results if result['success'])
        
        print("=" * 60)
        print("📊 TEST RESULTS SUMMARY")
        print("=" * 60)
        
        print(f"🎯 Main Notification Fixes: {main_tests_passed}/{total_main_tests} passed ({main_tests_passed/total_main_tests*100:.1f}%)")
        print(f"🔍 Additional Validations: {additional_passed}/{total_additional_tests} passed ({additional_passed/total_additional_tests*100:.1f}%)")
        print(f"📈 Overall Success Rate: {total_passed}/{total_tests} passed ({total_passed/total_tests*100:.1f}%)")
        
        print("\n📋 Detailed Results:")
        for result in self.test_results:
            print(f"  {result['status']}: {result['test']}")
            if result['details']:
                print(f"      {result['details']}")
        
        # Determine overall status
        if main_tests_passed == total_main_tests:
            print("\n🎉 ALL MAIN NOTIFICATION FIXES ARE WORKING CORRECTLY!")
            return True
        else:
            print(f"\n⚠️  {total_main_tests - main_tests_passed} out of {total_main_tests} main notification fixes need attention")
            return False

def main():
    """Main test execution"""
    test_suite = NotificationTestSuite()
    success = test_suite.run_all_tests()
    
    # Exit with appropriate code
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()