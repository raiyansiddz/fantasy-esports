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

    def test_working_admin_endpoints(self):
        """Test known working admin endpoints with authentication"""
        if not self.admin_token:
            self.log_test(
                "Working Admin Endpoints Test",
                False,
                "Cannot test admin endpoints - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        working_endpoints = [
            {"method": "GET", "path": "/admin/users", "name": "Get Users"},
            {"method": "GET", "path": "/admin/kyc/documents", "name": "Get KYC Documents"},
            {"method": "GET", "path": "/admin/config", "name": "Get System Config"}
        ]
        
        success_count = 0
        total_count = len(working_endpoints)
        
        for endpoint in working_endpoints:
            try:
                if endpoint["method"] == "GET":
                    response = self.session.get(f"{self.api_base}{endpoint['path']}", headers=headers, timeout=10)
                else:
                    response = self.session.post(f"{self.api_base}{endpoint['path']}", headers=headers, json={}, timeout=10)
                
                if response.status_code in [200, 201]:
                    self.log_test(
                        f"Working Admin Endpoint - {endpoint['name']}",
                        True,
                        f"✅ {endpoint['name']} working correctly (status: {response.status_code})",
                        {"status_code": response.status_code, "endpoint": endpoint['path']}
                    )
                    success_count += 1
                else:
                    self.log_test(
                        f"Working Admin Endpoint - {endpoint['name']}",
                        False,
                        f"❌ {endpoint['name']} returned unexpected status: {response.status_code}",
                        {"status_code": response.status_code, "endpoint": endpoint['path'], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Working Admin Endpoint - {endpoint['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "endpoint": endpoint['path']}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Working Admin Endpoints Summary",
            overall_success,
            f"{'✅ All working admin endpoints functional' if overall_success else f'❌ {total_count - success_count}/{total_count} admin endpoints failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

    def test_analytics_endpoints(self):
        """Test analytics endpoints - expected to return 404 due to route registration issues"""
        if not self.admin_token:
            self.log_test(
                "Analytics Endpoints Test",
                False,
                "Cannot test analytics endpoints - no admin token available",
                {"admin_token": None}
            )
            return False

        headers = {"Authorization": f"Bearer {self.admin_token}"}
        
        analytics_endpoints = [
            {"method": "GET", "path": "/admin/analytics/dashboard", "name": "Analytics Dashboard"},
            {"method": "GET", "path": "/admin/analytics/users", "name": "User Analytics"},
            {"method": "GET", "path": "/admin/bi/dashboard", "name": "BI Dashboard"},
            {"method": "GET", "path": "/admin/bi/kpis", "name": "KPI Metrics"},
            {"method": "POST", "path": "/admin/reports/generate", "name": "Generate Report"},
            {"method": "GET", "path": "/admin/reports", "name": "Get Reports"}
        ]
        
        expected_404_count = 0
        total_count = len(analytics_endpoints)
        
        for endpoint in analytics_endpoints:
            try:
                if endpoint["method"] == "GET":
                    response = self.session.get(f"{self.api_base}{endpoint['path']}", headers=headers, timeout=10)
                else:
                    # For POST endpoints, send minimal valid payload
                    payload = {
                        "report_type": "financial",
                        "format": "json",
                        "date_from": "2024-01-01",
                        "date_to": "2024-12-31",
                        "description": "Test report"
                    }
                    response = self.session.post(f"{self.api_base}{endpoint['path']}", headers=headers, json=payload, timeout=10)
                
                if response.status_code == 404:
                    self.log_test(
                        f"Analytics Endpoint - {endpoint['name']}",
                        True,  # This is expected behavior
                        f"✅ EXPECTED: {endpoint['name']} returns 404 (route not registered)",
                        {"status_code": 404, "endpoint": endpoint['path'], "expected": True}
                    )
                    expected_404_count += 1
                elif response.status_code in [200, 201]:
                    self.log_test(
                        f"Analytics Endpoint - {endpoint['name']}",
                        False,  # This means the route is actually working
                        f"❌ UNEXPECTED: {endpoint['name']} is working (status: {response.status_code}) - route registration issue may be fixed",
                        {"status_code": response.status_code, "endpoint": endpoint['path'], "expected": False}
                    )
                elif response.status_code == 401:
                    self.log_test(
                        f"Analytics Endpoint - {endpoint['name']}",
                        False,
                        f"❌ UNEXPECTED: {endpoint['name']} returns 401 (auth issue, but route exists)",
                        {"status_code": 401, "endpoint": endpoint['path'], "expected": False}
                    )
                else:
                    self.log_test(
                        f"Analytics Endpoint - {endpoint['name']}",
                        False,
                        f"❌ UNEXPECTED: {endpoint['name']} returns {response.status_code}",
                        {"status_code": response.status_code, "endpoint": endpoint['path'], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Analytics Endpoint - {endpoint['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "endpoint": endpoint['path']}
                )
        
        # Success means we got the expected 404s
        overall_success = expected_404_count == total_count
        self.log_test(
            "Analytics Endpoints Summary",
            overall_success,
            f"{'✅ All analytics endpoints return expected 404s (route registration issue confirmed)' if overall_success else f'❌ {total_count - expected_404_count}/{total_count} analytics endpoints are unexpectedly working'} (Expected 404s: {expected_404_count}/{total_count})",
            {"expected_404_count": expected_404_count, "total_count": total_count}
        )
        
        return overall_success

    def test_basic_user_endpoints(self):
        """Test basic public user endpoints that should work"""
        user_endpoints = [
            {"method": "GET", "path": "/games", "name": "Get Games"},
            {"method": "GET", "path": "/tournaments", "name": "Get Tournaments"}
        ]
        
        success_count = 0
        total_count = len(user_endpoints)
        
        for endpoint in user_endpoints:
            try:
                response = self.session.get(f"{self.api_base}{endpoint['path']}", timeout=10)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get('success'):
                        self.log_test(
                            f"Basic User Endpoint - {endpoint['name']}",
                            True,
                            f"✅ {endpoint['name']} working correctly",
                            {"status_code": 200, "endpoint": endpoint['path']}
                        )
                        success_count += 1
                    else:
                        self.log_test(
                            f"Basic User Endpoint - {endpoint['name']}",
                            False,
                            f"❌ {endpoint['name']} returned success=false",
                            {"status_code": 200, "endpoint": endpoint['path'], "success": False}
                        )
                else:
                    self.log_test(
                        f"Basic User Endpoint - {endpoint['name']}",
                        False,
                        f"❌ {endpoint['name']} returned status: {response.status_code}",
                        {"status_code": response.status_code, "endpoint": endpoint['path'], "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Basic User Endpoint - {endpoint['name']}",
                    False,
                    f"Request failed: {str(e)}",
                    {"error": str(e), "endpoint": endpoint['path']}
                )
        
        overall_success = success_count == total_count
        self.log_test(
            "Basic User Endpoints Summary",
            overall_success,
            f"{'✅ All basic user endpoints working' if overall_success else f'❌ {total_count - success_count}/{total_count} user endpoints failed'} (Success rate: {success_count}/{total_count})",
            {"success_count": success_count, "total_count": total_count}
        )
        
        return overall_success

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