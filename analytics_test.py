#!/usr/bin/env python3
"""
Comprehensive Analytics Dashboard System Testing for GoLang Fantasy Esports Backend
Testing Analytics Dashboard, Business Intelligence Dashboard, and Reporting System endpoints
"""

import requests
import json
import time
from typing import Dict, Any, Optional, List

class AnalyticsDashboardTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
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

    def test_health_check(self) -> bool:
        """Test basic health check endpoint"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            success = response.status_code == 200 and "healthy" in response.text
            
            self.log_test(
                "Health Check - Backend Connectivity",
                success,
                f"Status: {response.status_code}, Response: {response.text[:100]}",
                response.json() if success else response.text
            )
            return success
        except Exception as e:
            self.log_test("Health Check - Backend Connectivity", False, f"Exception: {str(e)}")
            return False

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user"""
        try:
            # Try to authenticate as admin
            login_data = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=login_data)
            
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
                f"Status: {response.status_code}",
                response.text
            )
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def test_analytics_dashboard_endpoint(self) -> bool:
        """Test GET /api/v1/admin/analytics/dashboard endpoint"""
        if not self.admin_token:
            self.log_test("Analytics Dashboard Endpoint", False, "No admin token available")
            return False
            
        try:
            # Test GET /api/v1/admin/analytics/dashboard
            response = self.session.get(f"{self.base_url}/api/v1/admin/analytics/dashboard")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    dashboard_data = data.get("data", {})
                    
                    # Check for expected dashboard components
                    expected_components = [
                        "user_metrics", "revenue_metrics", "contest_metrics", 
                        "game_metrics", "engagement_metrics", "system_health", "top_performers"
                    ]
                    
                    found_components = []
                    for component in expected_components:
                        if component in dashboard_data:
                            found_components.append(component)
                    
                    details += f" - Dashboard loaded successfully"
                    details += f" - Found components: {found_components}"
                    details += f" - Expected components: {len(expected_components)}, Found: {len(found_components)}"
                    
                    if len(found_components) >= len(expected_components) * 0.7:  # At least 70% of expected components
                        details += " - Comprehensive analytics data available"
                    else:
                        success = False
                        details += " - Missing critical dashboard components"
                else:
                    success = False
                    details += f" - Request failed: {data.get('message', 'Unknown error')}"
            else:
                if response.status_code == 404:
                    details += " - Endpoint not found (404 error)"
                elif response.status_code == 401:
                    details += " - Authentication required (401 error)"
                else:
                    details += f" - Unexpected error"
            
            self.log_test(
                "Analytics Dashboard Endpoint - GET /admin/analytics/dashboard",
                success,
                details,
                response.json() if response.status_code == 200 else response.text[:200]
            )
            return success
            
        except Exception as e:
            self.log_test("Analytics Dashboard Endpoint - GET /admin/analytics/dashboard", False, f"Exception: {str(e)}")
            return False

    def test_bi_dashboard_endpoint(self) -> bool:
        """Test GET /api/v1/admin/bi/dashboard endpoint"""
        if not self.admin_token:
            self.log_test("BI Dashboard Endpoint", False, "No admin token available")
            return False
            
        try:
            # Test GET /api/v1/admin/bi/dashboard
            response = self.session.get(f"{self.base_url}/api/v1/admin/bi/dashboard")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    bi_data = data.get("data", {})
                    
                    # Check for expected BI dashboard components
                    expected_components = [
                        "kpi_metrics", "revenue_analytics", "user_behavior_analysis", 
                        "predictive_analytics", "competitive_analysis", "business_insights"
                    ]
                    
                    found_components = []
                    for component in expected_components:
                        if component in bi_data:
                            found_components.append(component)
                    
                    details += f" - BI Dashboard loaded successfully"
                    details += f" - Found components: {found_components}"
                    details += f" - Expected components: {len(expected_components)}, Found: {len(found_components)}"
                    
                    if len(found_components) >= len(expected_components) * 0.7:  # At least 70% of expected components
                        details += " - Comprehensive BI data available"
                    else:
                        success = False
                        details += " - Missing critical BI dashboard components"
                else:
                    success = False
                    details += f" - Request failed: {data.get('message', 'Unknown error')}"
            else:
                if response.status_code == 404:
                    details += " - Endpoint not found (404 error)"
                elif response.status_code == 401:
                    details += " - Authentication required (401 error)"
                else:
                    details += f" - Unexpected error"
            
            self.log_test(
                "BI Dashboard Endpoint - GET /admin/bi/dashboard",
                success,
                details,
                response.json() if response.status_code == 200 else response.text[:200]
            )
            return success
            
        except Exception as e:
            self.log_test("BI Dashboard Endpoint - GET /admin/bi/dashboard", False, f"Exception: {str(e)}")
            return False

    def test_report_generation_endpoint(self) -> bool:
        """Test POST /api/v1/admin/reports/generate endpoint"""
        if not self.admin_token:
            self.log_test("Report Generation Endpoint", False, "No admin token available")
            return False
            
        try:
            # Test POST /api/v1/admin/reports/generate with valid report types
            valid_report_types = ["user", "financial", "contest", "game", "performance", "compliance", "referral"]
            
            for report_type in valid_report_types:
                report_data = {
                    "report_type": report_type,
                    "format": "json",
                    "date_from": "2024-08-01T00:00:00Z",
                    "date_to": "2024-08-31T23:59:59Z",
                    "description": f"Automated test report for {report_type} analytics",
                    "filters": {}
                }
                
                response = self.session.post(
                    f"{self.base_url}/api/v1/admin/reports/generate", 
                    json=report_data
                )
                
                success = response.status_code == 200
                details = f"Report Type: {report_type}, Status: {response.status_code}"
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        report_info = data.get("data", {})
                        report_id = report_info.get("id")
                        details += f" - Report generated successfully (ID: {report_id})"
                        
                        # Validate report structure
                        required_fields = ["id", "title", "report_type", "status"]
                        missing_fields = [f for f in required_fields if f not in report_info]
                        
                        if missing_fields:
                            success = False
                            details += f" - Missing required fields: {missing_fields}"
                        else:
                            details += f" - All required fields present"
                    else:
                        success = False
                        details += f" - Report generation failed: {data.get('message', 'Unknown error')}"
                else:
                    if response.status_code == 404:
                        details += " - Endpoint not found (404 error)"
                    elif response.status_code == 401:
                        details += " - Authentication required (401 error)"
                    elif response.status_code == 400:
                        details += " - Bad request (validation error)"
                    else:
                        details += f" - Unexpected error"
                
                self.log_test(
                    f"Report Generation - {report_type.title()} Report",
                    success,
                    details,
                    response.json() if response.status_code == 200 else response.text[:200]
                )
                
                # If any report type fails, we'll continue but note the failure
                if not success:
                    break
            
            return success
            
        except Exception as e:
            self.log_test("Report Generation Endpoint - POST /admin/reports/generate", False, f"Exception: {str(e)}")
            return False

    def test_reports_list_endpoint(self) -> bool:
        """Test GET /api/v1/admin/reports endpoint"""
        if not self.admin_token:
            self.log_test("Reports List Endpoint", False, "No admin token available")
            return False
            
        try:
            # Test GET /api/v1/admin/reports
            response = self.session.get(f"{self.base_url}/api/v1/admin/reports")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    reports_data = data.get("data", {})
                    
                    # Check for pagination and reports list
                    if isinstance(reports_data, dict):
                        reports = reports_data.get("reports", [])
                        pagination = reports_data.get("pagination", {})
                        
                        details += f" - Reports list retrieved successfully"
                        details += f" - Found {len(reports)} reports"
                        details += f" - Pagination: Page {pagination.get('page', 1)} of {pagination.get('pages', 1)}"
                        details += f" - Total reports: {pagination.get('total', 0)}"
                        
                        # Validate report structure if reports exist
                        if reports:
                            sample_report = reports[0]
                            required_fields = ["id", "title", "report_type", "status", "created_at"]
                            missing_fields = [f for f in required_fields if f not in sample_report]
                            
                            if missing_fields:
                                success = False
                                details += f" - Missing required fields in reports: {missing_fields}"
                            else:
                                details += f" - Report structure is valid"
                        else:
                            details += " - No reports found (empty list)"
                    else:
                        success = False
                        details += " - Invalid response structure"
                else:
                    success = False
                    details += f" - Request failed: {data.get('message', 'Unknown error')}"
            else:
                if response.status_code == 404:
                    details += " - Endpoint not found (404 error)"
                elif response.status_code == 401:
                    details += " - Authentication required (401 error)"
                else:
                    details += f" - Unexpected error"
            
            self.log_test(
                "Reports List Endpoint - GET /admin/reports",
                success,
                details,
                response.json() if response.status_code == 200 else response.text[:200]
            )
            return success
            
        except Exception as e:
            self.log_test("Reports List Endpoint - GET /admin/reports", False, f"Exception: {str(e)}")
            return False

    def test_authentication_enforcement(self) -> bool:
        """Test that endpoints properly enforce authentication"""
        try:
            # Save current headers
            original_headers = self.session.headers.copy()
            
            # Remove Authorization header
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            endpoints_to_test = [
                "/api/v1/admin/analytics/dashboard",
                "/api/v1/admin/bi/dashboard",
                "/api/v1/admin/reports"
            ]
            
            auth_tests_passed = 0
            total_auth_tests = len(endpoints_to_test)
            
            for endpoint in endpoints_to_test:
                response = self.session.get(f"{self.base_url}{endpoint}")
                
                if response.status_code == 401:
                    auth_tests_passed += 1
                    self.log_test(
                        f"Authentication Enforcement - {endpoint}",
                        True,
                        "Correctly rejected unauthenticated request (401)"
                    )
                else:
                    self.log_test(
                        f"Authentication Enforcement - {endpoint}",
                        False,
                        f"Expected 401, got {response.status_code}"
                    )
            
            # Restore original headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = auth_tests_passed == total_auth_tests
            self.log_test(
                "Authentication Enforcement - Overall",
                success,
                f"Passed {auth_tests_passed}/{total_auth_tests} authentication tests"
            )
            
            return success
            
        except Exception as e:
            self.log_test("Authentication Enforcement", False, f"Exception: {str(e)}")
            return False

    def test_analytics_endpoints_with_filters(self) -> bool:
        """Test analytics endpoints with various filter parameters"""
        if not self.admin_token:
            self.log_test("Analytics Endpoints with Filters", False, "No admin token available")
            return False
            
        try:
            # Test analytics dashboard with filters
            filter_params = {
                "date_from": "2024-01-01",
                "date_to": "2024-12-31",
                "period": "month",
                "game_id": "1"
            }
            
            response = self.session.get(
                f"{self.base_url}/api/v1/admin/analytics/dashboard",
                params=filter_params
            )
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Analytics dashboard with filters working"
                    details += f" - Filters applied: {filter_params}"
                else:
                    success = False
                    details += f" - Request failed: {data.get('message', 'Unknown error')}"
            else:
                details += " - Failed to apply filters"
            
            self.log_test(
                "Analytics Dashboard with Filters",
                success,
                details,
                response.json() if response.status_code == 200 else response.text[:200]
            )
            
            # Test BI dashboard with filters
            bi_filter_params = {
                "date_from": "2024-01-01",
                "date_to": "2024-12-31",
                "user_segment": "premium",
                "confidence_level": "0.95"
            }
            
            bi_response = self.session.get(
                f"{self.base_url}/api/v1/admin/bi/dashboard",
                params=bi_filter_params
            )
            
            bi_success = bi_response.status_code == 200
            bi_details = f"Status: {bi_response.status_code}"
            
            if bi_success:
                bi_data = bi_response.json()
                if bi_data.get("success"):
                    bi_details += " - BI dashboard with filters working"
                    bi_details += f" - Filters applied: {bi_filter_params}"
                else:
                    bi_success = False
                    bi_details += f" - Request failed: {bi_data.get('message', 'Unknown error')}"
            else:
                bi_details += " - Failed to apply BI filters"
            
            self.log_test(
                "BI Dashboard with Filters",
                bi_success,
                bi_details,
                bi_response.json() if bi_response.status_code == 200 else bi_response.text[:200]
            )
            
            return success and bi_success
            
        except Exception as e:
            self.log_test("Analytics Endpoints with Filters", False, f"Exception: {str(e)}")
            return False

    def run_comprehensive_tests(self):
        """Run all analytics dashboard tests"""
        print("🚀 Starting Comprehensive Analytics Dashboard System Testing")
        print("=" * 80)
        
        # Test 1: Health Check
        if not self.test_health_check():
            print("❌ Backend is not healthy. Stopping tests.")
            return
        
        # Test 2: Admin Authentication
        if not self.authenticate_admin():
            print("❌ Admin authentication failed. Cannot test admin endpoints.")
            return
        
        # Test 3: Analytics Dashboard Endpoint
        self.test_analytics_dashboard_endpoint()
        
        # Test 4: Business Intelligence Dashboard Endpoint
        self.test_bi_dashboard_endpoint()
        
        # Test 5: Report Generation Endpoint
        self.test_report_generation_endpoint()
        
        # Test 6: Reports List Endpoint
        self.test_reports_list_endpoint()
        
        # Test 7: Authentication Enforcement
        self.test_authentication_enforcement()
        
        # Test 8: Analytics Endpoints with Filters
        self.test_analytics_endpoints_with_filters()
        
        # Generate Summary
        self.generate_summary()

    def generate_summary(self):
        """Generate test summary"""
        print("\n" + "=" * 80)
        print("📊 ANALYTICS DASHBOARD SYSTEM TEST SUMMARY")
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
        
        # Categorize results
        categories = {
            "Health & Connectivity": [],
            "Authentication": [],
            "Analytics Dashboard": [],
            "Business Intelligence": [],
            "Reporting System": [],
            "Security & Validation": []
        }
        
        for result in self.test_results:
            test_name = result["test"]
            if "Health" in test_name:
                categories["Health & Connectivity"].append(result)
            elif "Admin Authentication" in test_name:
                categories["Authentication"].append(result)
            elif "Analytics Dashboard" in test_name or "Analytics Endpoints" in test_name:
                categories["Analytics Dashboard"].append(result)
            elif "BI Dashboard" in test_name:
                categories["Business Intelligence"].append(result)
            elif "Report" in test_name:
                categories["Reporting System"].append(result)
            elif "Authentication Enforcement" in test_name:
                categories["Security & Validation"].append(result)
        
        for category, results in categories.items():
            if results:
                passed = sum(1 for r in results if r["success"])
                total = len(results)
                print(f"{category}: {passed}/{total} passed")
        
        print("\n" + "=" * 80)
        print("🔍 DETAILED FINDINGS")
        print("=" * 80)
        
        # Show failed tests
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print("❌ FAILED TESTS:")
            for result in failed_results:
                print(f"  • {result['test']}: {result['details']}")
        else:
            print("✅ ALL TESTS PASSED!")
        
        print("\n" + "=" * 80)
        print("🎯 ANALYTICS DASHBOARD SYSTEM STATUS")
        print("=" * 80)
        
        # Overall assessment
        if success_rate >= 90:
            print("🎉 EXCELLENT: Analytics Dashboard System is working excellently!")
        elif success_rate >= 75:
            print("✅ GOOD: Analytics Dashboard System is working well with minor issues.")
        elif success_rate >= 50:
            print("⚠️  MODERATE: Analytics Dashboard System has some issues that need attention.")
        else:
            print("❌ CRITICAL: Analytics Dashboard System has significant issues requiring immediate attention.")
        
        # Key functionality assessment
        analytics_tests = [r for r in self.test_results if "Analytics Dashboard" in r["test"]]
        bi_tests = [r for r in self.test_results if "BI Dashboard" in r["test"]]
        reporting_tests = [r for r in self.test_results if "Report" in r["test"]]
        
        analytics_success = sum(1 for r in analytics_tests if r["success"]) / len(analytics_tests) * 100 if analytics_tests else 0
        bi_success = sum(1 for r in bi_tests if r["success"]) / len(bi_tests) * 100 if bi_tests else 0
        reporting_success = sum(1 for r in reporting_tests if r["success"]) / len(reporting_tests) * 100 if reporting_tests else 0
        
        print(f"\nAnalytics Dashboard: {analytics_success:.1f}% functional")
        print(f"Business Intelligence: {bi_success:.1f}% functional")
        print(f"Reporting System: {reporting_success:.1f}% functional")
        
        if analytics_success >= 75 and bi_success >= 75 and reporting_success >= 75:
            print("\n🚀 READY FOR PRODUCTION: Core analytics functionality is working!")
        else:
            print("\n⚠️  NEEDS WORK: Core analytics functionality requires fixes before production.")

if __name__ == "__main__":
    tester = AnalyticsDashboardTester()
    tester.run_comprehensive_tests()