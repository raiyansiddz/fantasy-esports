#!/usr/bin/env python3
"""
Enhanced URL Validation Test - Testing valid URLs for stream endpoint
This tests the enhanced URL validation mentioned in the review request
"""

import requests
import json
import sys
import time

class EnhancedURLValidationTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.api_base = f"{base_url}/api/v1"
        self.session = requests.Session()
        self.test_results = []
        
    def log_test(self, test_name: str, passed: bool, details: str, response_data: dict = None):
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

    def get_admin_token(self):
        """Get admin token for authenticated tests"""
        try:
            payload = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.api_base}/admin/login", json=payload, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                return data.get('access_token')
            else:
                print(f"Admin login failed: {response.status_code} - {response.text}")
                
        except Exception as e:
            print(f"Could not get admin token: {e}")
        
        return None

    def test_valid_stream_urls(self):
        """Test valid streaming URLs that should be accepted"""
        admin_token = self.get_admin_token()
        
        if not admin_token:
            self.log_test(
                "Valid URL Test Setup",
                False,
                "Could not get admin token for testing valid URLs",
                {"error": "No admin token"}
            )
            return False
        
        headers = {"Authorization": f"Bearer {admin_token}"}
        
        # Test with a valid match ID (we'll use 1 as it's commonly used in tests)
        match_id = 1
        
        valid_urls = [
            "https://youtube.com/watch?v=12345",
            "https://twitch.tv/streamer",
            "https://example.com/live/stream",
            "https://streaming.example.com/channel/123",
            "http://localhost:3000/stream/test",
            "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
            "https://www.twitch.tv/ninja"
        ]
        
        for valid_url in valid_urls:
            try:
                payload = {
                    "stream_url": valid_url,
                    "stream_title": "Test Stream",
                    "auto_activate": False
                }
                
                response = self.session.post(
                    f"{self.api_base}/admin/matches/{match_id}/live-stream",
                    json=payload,
                    headers=headers,
                    timeout=10
                )
                
                # Valid URLs should either:
                # 1. Return 200 (success)
                # 2. Return 404 if match doesn't exist (but URL validation passed)
                # 3. Return some other business logic error (but not URL validation error)
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_message = data.get('error', '')
                        
                        # If it's a URL validation error, that's a problem
                        if 'url' in error_message.lower() and ('invalid' in error_message.lower() or 'scheme' in error_message.lower()):
                            self.log_test(
                                f"Valid URL Test - '{valid_url}'",
                                False,
                                f"❌ ISSUE: Valid URL rejected with validation error: {error_message}",
                                {
                                    "status_code": response.status_code,
                                    "error_message": error_message,
                                    "url_tested": valid_url
                                }
                            )
                        else:
                            # Some other business logic error (e.g., match not found, etc.)
                            self.log_test(
                                f"Valid URL Test - '{valid_url}'",
                                True,
                                f"✅ URL validation passed (business logic error: {error_message})",
                                {
                                    "status_code": response.status_code,
                                    "error_message": error_message,
                                    "url_tested": valid_url
                                }
                            )
                    except:
                        self.log_test(
                            f"Valid URL Test - '{valid_url}'",
                            True,
                            f"✅ URL validation passed (status 400 but not URL validation error)",
                            {"status_code": response.status_code, "url_tested": valid_url}
                        )
                
                elif response.status_code in [200, 201]:
                    self.log_test(
                        f"Valid URL Test - '{valid_url}'",
                        True,
                        f"✅ Valid URL accepted successfully",
                        {"status_code": response.status_code, "url_tested": valid_url}
                    )
                
                elif response.status_code == 404:
                    self.log_test(
                        f"Valid URL Test - '{valid_url}'",
                        True,
                        f"✅ URL validation passed (match not found, but URL is valid)",
                        {"status_code": response.status_code, "url_tested": valid_url}
                    )
                
                else:
                    # Other status codes - check if it's URL validation related
                    try:
                        data = response.json()
                        error_message = data.get('error', '')
                        
                        if 'url' in error_message.lower() and ('invalid' in error_message.lower() or 'scheme' in error_message.lower()):
                            self.log_test(
                                f"Valid URL Test - '{valid_url}'",
                                False,
                                f"❌ ISSUE: Valid URL rejected: {error_message}",
                                {
                                    "status_code": response.status_code,
                                    "error_message": error_message,
                                    "url_tested": valid_url
                                }
                            )
                        else:
                            self.log_test(
                                f"Valid URL Test - '{valid_url}'",
                                True,
                                f"✅ URL validation passed (other error: {error_message})",
                                {
                                    "status_code": response.status_code,
                                    "error_message": error_message,
                                    "url_tested": valid_url
                                }
                            )
                    except:
                        self.log_test(
                            f"Valid URL Test - '{valid_url}'",
                            True,
                            f"✅ URL validation passed (status {response.status_code})",
                            {"status_code": response.status_code, "url_tested": valid_url}
                        )
                        
            except Exception as e:
                self.log_test(
                    f"Valid URL Test - '{valid_url}'",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "url_tested": valid_url}
                )
        
        return True

    def test_edge_case_urls(self):
        """Test edge case URLs that should be handled properly"""
        admin_token = self.get_admin_token()
        
        if not admin_token:
            return False
        
        headers = {"Authorization": f"Bearer {admin_token}"}
        match_id = 1
        
        edge_case_urls = [
            "http://example.com/random",  # No streaming keywords
            "https://example.com/video",  # Generic video URL
            "https://example.com/watch",  # Generic watch URL
        ]
        
        for url in edge_case_urls:
            try:
                payload = {
                    "stream_url": url,
                    "stream_title": "Test Stream",
                    "auto_activate": False
                }
                
                response = self.session.post(
                    f"{self.api_base}/admin/matches/{match_id}/live-stream",
                    json=payload,
                    headers=headers,
                    timeout=10
                )
                
                # These URLs might be rejected based on business logic (no streaming keywords)
                # but should not cause server errors
                
                if response.status_code == 400:
                    try:
                        data = response.json()
                        error_message = data.get('error', '')
                        
                        self.log_test(
                            f"Edge Case URL Test - '{url}'",
                            True,
                            f"✅ Edge case handled properly: {error_message}",
                            {
                                "status_code": response.status_code,
                                "error_message": error_message,
                                "url_tested": url
                            }
                        )
                    except:
                        self.log_test(
                            f"Edge Case URL Test - '{url}'",
                            True,
                            f"✅ Edge case handled (status 400)",
                            {"status_code": response.status_code, "url_tested": url}
                        )
                
                elif response.status_code in [200, 201, 404]:
                    self.log_test(
                        f"Edge Case URL Test - '{url}'",
                        True,
                        f"✅ Edge case handled properly (status {response.status_code})",
                        {"status_code": response.status_code, "url_tested": url}
                    )
                
                else:
                    self.log_test(
                        f"Edge Case URL Test - '{url}'",
                        False,
                        f"❌ Unexpected status code: {response.status_code}",
                        {
                            "status_code": response.status_code,
                            "url_tested": url,
                            "response": response.text[:200]
                        }
                    )
                        
            except Exception as e:
                self.log_test(
                    f"Edge Case URL Test - '{url}'",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "url_tested": url}
                )
        
        return True

    def run_enhanced_tests(self):
        """Run enhanced URL validation tests"""
        print("=" * 80)
        print("🧪 ENHANCED URL VALIDATION TESTING")
        print("Testing valid URLs and edge cases for stream endpoint")
        print("=" * 80)
        print()
        
        # Test valid URLs
        print("🔍 Testing Valid Streaming URLs")
        self.test_valid_stream_urls()
        
        # Test edge case URLs
        print("🔍 Testing Edge Case URLs")
        self.test_edge_case_urls()
        
        # Generate summary
        self.generate_summary()
        
        return True

    def generate_summary(self):
        """Generate test summary"""
        print("=" * 80)
        print("📊 ENHANCED URL VALIDATION TEST SUMMARY")
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
        
        # Show successful validations
        successful_validations = [r for r in self.test_results if r['passed']]
        if successful_validations:
            print("✅ SUCCESSFUL VALIDATIONS:")
            for validation in successful_validations:
                print(f"  • {validation['test']}")
            print()
        
        print("=" * 80)

if __name__ == "__main__":
    tester = EnhancedURLValidationTester()
    tester.run_enhanced_tests()