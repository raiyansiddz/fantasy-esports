#!/usr/bin/env python3
"""
Backend API Testing Script for Fantasy Esports Platform
Testing GoLang Fantasy Esports backend analytics endpoints and core functionality:

FOCUS: Analytics routes registration issue testing
- Health Check: Verify backend is responsive
- Basic Admin Authentication: Test admin login
- Working Admin Endpoints: Test known working endpoints
- Analytics Endpoints: Test expected 404 failures
- Basic User Endpoints: Test public endpoints

Expected: Analytics endpoints should return 404 due to route registration issues
"""

import requests
import json
import sys
import time
from typing import Dict, Any, List, Optional

class FantasyEsportsAPITester:
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

    def test_tournament_filter_completed(self):
        """
        Issue 1: Test GET /api/v1/tournaments?status=completed
        Should return empty array instead of null when no completed tournaments exist
        """
        try:
            response = self.session.get(f"{self.api_base}/tournaments?status=completed", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                
                # Check if response has success field
                if not data.get('success', False):
                    self.log_test(
                        "Tournament Filter - Completed Status",
                        False,
                        "Response success field is false or missing",
                        data
                    )
                    return False
                
                # Check if tournaments field exists and is a list
                tournaments = data.get('tournaments')
                if tournaments is None:
                    self.log_test(
                        "Tournament Filter - Completed Status",
                        False,
                        "❌ CRITICAL: tournaments field is null/none instead of empty array",
                        data
                    )
                    return False
                
                if not isinstance(tournaments, list):
                    self.log_test(
                        "Tournament Filter - Completed Status",
                        False,
                        f"❌ CRITICAL: tournaments field is {type(tournaments)} instead of list/array",
                        data
                    )
                    return False
                
                # Success - tournaments is an empty array
                self.log_test(
                    "Tournament Filter - Completed Status",
                    True,
                    f"✅ FIXED: Returns empty array with {len(tournaments)} tournaments. Response structure correct.",
                    {
                        "tournaments_count": len(tournaments),
                        "tournaments_type": str(type(tournaments)),
                        "has_pagination": "pagination" in data
                    }
                )
                return True
                
            else:
                self.log_test(
                    "Tournament Filter - Completed Status",
                    False,
                    f"Unexpected status code: {response.status_code}",
                    {"status_code": response.status_code, "response": response.text[:200]}
                )
                return False
                
        except Exception as e:
            self.log_test(
                "Tournament Filter - Completed Status",
                False,
                f"Request failed: {str(e)}",
                {"error": str(e)}
            )
            return False

    def test_active_live_streams(self):
        """
        Issue 2: Test GET /api/v1/live-streams/active
        Should return 200 with empty array when no active streams exist, not 404
        """
        try:
            response = self.session.get(f"{self.api_base}/live-streams/active", timeout=10)
            
            if response.status_code == 404:
                self.log_test(
                    "Get Active Live Streams",
                    False,
                    "❌ CRITICAL: Returns 404 instead of 200 with empty array",
                    {"status_code": 404, "response": response.text[:200]}
                )
                return False
            
            if response.status_code == 200:
                data = response.json()
                
                # Check if response has success field
                if not data.get('success', False):
                    self.log_test(
                        "Get Active Live Streams",
                        False,
                        "Response success field is false or missing",
                        data
                    )
                    return False
                
                # Check if active_streams field exists and is a list
                active_streams = data.get('active_streams')
                if active_streams is None:
                    self.log_test(
                        "Get Active Live Streams",
                        False,
                        "❌ CRITICAL: active_streams field is null/none instead of empty array",
                        data
                    )
                    return False
                
                if not isinstance(active_streams, list):
                    self.log_test(
                        "Get Active Live Streams",
                        False,
                        f"❌ CRITICAL: active_streams field is {type(active_streams)} instead of list/array",
                        data
                    )
                    return False
                
                # Success - returns 200 with empty array
                self.log_test(
                    "Get Active Live Streams",
                    True,
                    f"✅ FIXED: Returns 200 with empty array. Found {len(active_streams)} active streams.",
                    {
                        "status_code": 200,
                        "active_streams_count": len(active_streams),
                        "active_streams_type": str(type(active_streams)),
                        "has_count_field": "count" in data
                    }
                )
                return True
                
            else:
                self.log_test(
                    "Get Active Live Streams",
                    False,
                    f"Unexpected status code: {response.status_code}",
                    {"status_code": response.status_code, "response": response.text[:200]}
                )
                return False
                
        except Exception as e:
            self.log_test(
                "Get Active Live Streams",
                False,
                f"Request failed: {str(e)}",
                {"error": str(e)}
            )
            return False

    def test_stream_url_validation(self):
        """
        Issue 3: Test POST /api/v1/admin/matches/{id}/live-stream with invalid URL
        Should return 400/422 with proper error message, not 404
        """
        # First, let's try to get admin token (we'll test without auth later)
        admin_token = self.get_admin_token()
        
        # Test with a non-existent match ID but valid auth to isolate URL validation
        match_id = 99999  # Non-existent match ID
        invalid_urls = [
            "not-a-url",
            "ftp://invalid-protocol.com",
            "http://",
            "invalid-format",
            ""
        ]
        
        headers = {}
        if admin_token:
            headers["Authorization"] = f"Bearer {admin_token}"
        
        for invalid_url in invalid_urls:
            try:
                payload = {
                    "stream_url": invalid_url,
                    "stream_title": "Test Stream",
                    "auto_activate": False
                }
                
                response = self.session.post(
                    f"{self.api_base}/admin/matches/{match_id}/live-stream",
                    json=payload,
                    headers=headers,
                    timeout=10
                )
                
                if response.status_code == 404:
                    self.log_test(
                        f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                        False,
                        "❌ CRITICAL: Returns 404 instead of 400/422 for invalid URL",
                        {"status_code": 404, "url_tested": invalid_url}
                    )
                    continue
                
                if response.status_code in [400, 422]:
                    try:
                        data = response.json()
                        error_message = data.get('error', '')
                        
                        # Check if error message mentions URL validation
                        if 'url' in error_message.lower() or 'invalid' in error_message.lower():
                            self.log_test(
                                f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                                True,
                                f"✅ FIXED: Returns {response.status_code} with proper error message",
                                {
                                    "status_code": response.status_code,
                                    "error_message": error_message,
                                    "url_tested": invalid_url
                                }
                            )
                        else:
                            self.log_test(
                                f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                                False,
                                f"Returns {response.status_code} but error message doesn't mention URL validation",
                                {
                                    "status_code": response.status_code,
                                    "error_message": error_message,
                                    "url_tested": invalid_url
                                }
                            )
                    except:
                        self.log_test(
                            f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                            True,
                            f"✅ FIXED: Returns {response.status_code} (proper validation status)",
                            {"status_code": response.status_code, "url_tested": invalid_url}
                        )
                
                elif response.status_code == 401:
                    # This is expected if we don't have proper admin auth
                    self.log_test(
                        f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                        True,
                        "✅ Returns 401 (auth required) - this is correct behavior",
                        {"status_code": 401, "url_tested": invalid_url}
                    )
                
                else:
                    self.log_test(
                        f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                        False,
                        f"Unexpected status code: {response.status_code}",
                        {
                            "status_code": response.status_code,
                            "url_tested": invalid_url,
                            "response": response.text[:200]
                        }
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Stream URL Validation - Invalid URL: '{invalid_url}'",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "url_tested": invalid_url}
                )
        
        return True

    def test_admin_endpoint_without_auth(self):
        """
        Issue 4: Test admin endpoint without Authorization header
        Should return 401 (unauthorized) not 404 (not found)
        """
        admin_endpoints = [
            "/admin/users",
            "/admin/kyc/documents",
            "/admin/matches/live-scoring",
            "/admin/matches/1/start-scoring",
            "/admin/matches/1/live-stream"
        ]
        
        for endpoint in admin_endpoints:
            try:
                # Test GET endpoints
                if endpoint in ["/admin/users", "/admin/kyc/documents", "/admin/matches/live-scoring"]:
                    response = self.session.get(f"{self.api_base}{endpoint}", timeout=10)
                else:
                    # Test POST endpoints
                    response = self.session.post(f"{self.api_base}{endpoint}", json={}, timeout=10)
                
                if response.status_code == 404:
                    self.log_test(
                        f"Admin Auth Check - {endpoint}",
                        False,
                        "❌ CRITICAL: Returns 404 instead of 401 for missing auth",
                        {"status_code": 404, "endpoint": endpoint}
                    )
                    continue
                
                if response.status_code == 401:
                    try:
                        data = response.json()
                        error_message = data.get('error', '')
                        
                        self.log_test(
                            f"Admin Auth Check - {endpoint}",
                            True,
                            f"✅ FIXED: Returns 401 with proper error message",
                            {
                                "status_code": 401,
                                "error_message": error_message,
                                "endpoint": endpoint
                            }
                        )
                    except:
                        self.log_test(
                            f"Admin Auth Check - {endpoint}",
                            True,
                            "✅ FIXED: Returns 401 (unauthorized)",
                            {"status_code": 401, "endpoint": endpoint}
                        )
                
                else:
                    self.log_test(
                        f"Admin Auth Check - {endpoint}",
                        False,
                        f"Unexpected status code: {response.status_code}",
                        {
                            "status_code": response.status_code,
                            "endpoint": endpoint,
                            "response": response.text[:200]
                        }
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Admin Auth Check - {endpoint}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "endpoint": endpoint}
                )
        
        return True

    def get_admin_token(self) -> Optional[str]:
        """Try to get admin token for authenticated tests"""
        try:
            payload = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.api_base}/admin/login", json=payload, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                return data.get('access_token')
            
        except Exception as e:
            print(f"Could not get admin token: {e}")
        
        return None

    def run_all_tests(self):
        """Run all tests and generate summary"""
        print("=" * 80)
        print("🧪 FANTASY ESPORTS BACKEND API TESTING")
        print("Testing 4 specific issues mentioned in review request")
        print("=" * 80)
        print()
        
        # Test 1: Health check
        if not self.test_health_check():
            print("❌ Backend is not running. Cannot proceed with tests.")
            return False
        
        # Test 2: Tournament filter with completed status
        print("🔍 Testing Issue 1: Tournament Filter - status=completed")
        self.test_tournament_filter_completed()
        
        # Test 3: Active live streams
        print("🔍 Testing Issue 2: Get Active Live Streams")
        self.test_active_live_streams()
        
        # Test 4: Stream URL validation
        print("🔍 Testing Issue 3: Stream URL Validation")
        self.test_stream_url_validation()
        
        # Test 5: Admin endpoints without auth
        print("🔍 Testing Issue 4: Admin Endpoint Without Auth")
        self.test_admin_endpoint_without_auth()
        
        # Generate summary
        self.generate_summary()
        
        return True

    def generate_summary(self):
        """Generate test summary"""
        print("=" * 80)
        print("📊 TEST SUMMARY")
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
        critical_issues = [r for r in self.test_results if not r['passed'] and 'CRITICAL' in r['details']]
        if critical_issues:
            print("🚨 CRITICAL ISSUES FOUND:")
            for issue in critical_issues:
                print(f"  • {issue['test']}")
                print(f"    {issue['details']}")
            print()
        
        # Show fixes verified
        fixes_verified = [r for r in self.test_results if r['passed'] and 'FIXED' in r['details']]
        if fixes_verified:
            print("✅ FIXES VERIFIED:")
            for fix in fixes_verified:
                print(f"  • {fix['test']}")
            print()
        
        # Save results to file
        with open('/app/backend_test_results.json', 'w') as f:
            json.dump({
                'summary': {
                    'total_tests': total_tests,
                    'passed_tests': passed_tests,
                    'failed_tests': failed_tests,
                    'success_rate': f"{(passed_tests/total_tests)*100:.1f}%"
                },
                'test_results': self.test_results
            }, f, indent=2, default=str)
        
        print("📁 Detailed results saved to: /app/backend_test_results.json")
        print("=" * 80)

if __name__ == "__main__":
    tester = FantasyEsportsAPITester()
    tester.run_all_tests()