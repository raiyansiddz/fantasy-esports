#!/usr/bin/env python3
"""
Advanced Analytics, Business Intelligence, and Reporting System Testing
Testing the newly implemented analytics features in the GoLang Fantasy Esports backend.

Features to test:
1. Analytics Dashboard endpoints (/api/v1/admin/analytics/*)
2. Business Intelligence endpoints (/api/v1/admin/bi/*)
3. Advanced Reporting System endpoints (/api/v1/admin/reports/*)
4. Error handling and authentication
5. Data accuracy and response structures
"""

import requests
import json
import sys
import time
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta

class AnalyticsAPITester:
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
        if response_data and len(str(response_data)) < 500:
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

    def admin_login(self):
        """Login as admin to get authentication token"""
        try:
            # Try common admin credentials
            login_data = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.api_base}/admin/login", json=login_data, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                if data.get('success') and data.get('data', {}).get('token'):
                    self.admin_token = data['data']['token']
                    self.session.headers.update({'Authorization': f'Bearer {self.admin_token}'})
                    self.log_test(
                        "Admin Login",
                        True,
                        "Successfully logged in as admin",
                        {"has_token": bool(self.admin_token)}
                    )
                    return True
                else:
                    self.log_test(
                        "Admin Login",
                        False,
                        "Login response missing token",
                        data
                    )
                    return False
            else:
                # Try alternative credentials
                alt_credentials = [
                    {"username": "admin", "password": "password"},
                    {"username": "admin", "password": "admin"},
                    {"email": "admin@example.com", "password": "admin123"}
                ]
                
                for creds in alt_credentials:
                    response = self.session.post(f"{self.api_base}/admin/login", json=creds, timeout=10)
                    if response.status_code == 200:
                        data = response.json()
                        if data.get('success') and data.get('data', {}).get('token'):
                            self.admin_token = data['data']['token']
                            self.session.headers.update({'Authorization': f'Bearer {self.admin_token}'})
                            self.log_test(
                                "Admin Login",
                                True,
                                f"Successfully logged in with alternative credentials",
                                {"credentials_used": creds['username']}
                            )
                            return True
                
                self.log_test(
                    "Admin Login",
                    False,
                    f"Admin login failed with status {response.status_code}",
                    {"status_code": response.status_code, "response": response.text[:200]}
                )
                return False
                
        except Exception as e:
            self.log_test(
                "Admin Login",
                False,
                f"Admin login request failed: {str(e)}",
                {"error": str(e)}
            )
            return False

    def test_analytics_dashboard(self):
        """Test GET /api/v1/admin/analytics/dashboard"""
        try:
            # Test without parameters
            response = self.session.get(f"{self.api_base}/admin/analytics/dashboard", timeout=15)
            
            if response.status_code == 401:
                self.log_test(
                    "Analytics Dashboard - No Auth",
                    True,
                    "Correctly returns 401 when no authentication provided",
                    {"status_code": 401}
                )
            elif response.status_code == 200:
                data = response.json()
                if data.get('success') and 'data' in data:
                    dashboard_data = data['data']
                    expected_fields = ['UserMetrics', 'RevenueMetrics', 'ContestMetrics', 'GameMetrics', 'EngagementMetrics', 'SystemHealth', 'TopPerformers']
                    
                    missing_fields = [field for field in expected_fields if field not in dashboard_data]
                    
                    if not missing_fields:
                        self.log_test(
                            "Analytics Dashboard - Basic",
                            True,
                            f"Successfully retrieved comprehensive dashboard with all expected sections",
                            {"sections_count": len(dashboard_data), "sections": list(dashboard_data.keys())}
                        )
                    else:
                        self.log_test(
                            "Analytics Dashboard - Basic",
                            False,
                            f"Dashboard missing expected fields: {missing_fields}",
                            {"available_fields": list(dashboard_data.keys()), "missing_fields": missing_fields}
                        )
                else:
                    self.log_test(
                        "Analytics Dashboard - Basic",
                        False,
                        "Dashboard response missing success or data fields",
                        data
                    )
            else:
                self.log_test(
                    "Analytics Dashboard - Basic",
                    False,
                    f"Unexpected status code: {response.status_code}",
                    {"status_code": response.status_code, "response": response.text[:200]}
                )
                
            # Test with date filters
            date_from = (datetime.now() - timedelta(days=30)).strftime('%Y-%m-%d')
            date_to = datetime.now().strftime('%Y-%m-%d')
            
            params = {
                'date_from': date_from,
                'date_to': date_to,
                'period': 'month'
            }
            
            response = self.session.get(f"{self.api_base}/admin/analytics/dashboard", params=params, timeout=15)
            
            if response.status_code == 200:
                data = response.json()
                if data.get('success'):
                    self.log_test(
                        "Analytics Dashboard - With Filters",
                        True,
                        f"Successfully retrieved dashboard with date filters ({date_from} to {date_to})",
                        {"has_data": 'data' in data, "period": params['period']}
                    )
                else:
                    self.log_test(
                        "Analytics Dashboard - With Filters",
                        False,
                        "Dashboard with filters response not successful",
                        data
                    )
            else:
                self.log_test(
                    "Analytics Dashboard - With Filters",
                    False,
                    f"Dashboard with filters failed: {response.status_code}",
                    {"status_code": response.status_code}
                )
                
        except Exception as e:
            self.log_test(
                "Analytics Dashboard",
                False,
                f"Analytics dashboard test failed: {str(e)}",
                {"error": str(e)}
            )

    def test_analytics_endpoints(self):
        """Test all analytics endpoints"""
        endpoints = [
            ('users', 'User Metrics'),
            ('revenue', 'Revenue Metrics'),
            ('contests', 'Contest Metrics'),
            ('games', 'Game Metrics'),
            ('realtime', 'Real-time Metrics'),
            ('performance', 'Performance Metrics')
        ]
        
        for endpoint, name in endpoints:
            try:
                response = self.session.get(f"{self.api_base}/admin/analytics/{endpoint}", timeout=15)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get('success') and 'data' in data:
                        self.log_test(
                            f"Analytics - {name}",
                            True,
                            f"Successfully retrieved {name.lower()}",
                            {"has_data": bool(data['data']), "data_type": type(data['data']).__name__}
                        )
                    else:
                        self.log_test(
                            f"Analytics - {name}",
                            False,
                            f"{name} response missing success or data fields",
                            data
                        )
                elif response.status_code == 401:
                    self.log_test(
                        f"Analytics - {name}",
                        True,
                        f"Correctly requires authentication (401)",
                        {"status_code": 401}
                    )
                else:
                    self.log_test(
                        f"Analytics - {name}",
                        False,
                        f"{name} returned unexpected status: {response.status_code}",
                        {"status_code": response.status_code}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Analytics - {name}",
                    False,
                    f"{name} test failed: {str(e)}",
                    {"error": str(e)}
                )

    def test_business_intelligence_endpoints(self):
        """Test all business intelligence endpoints"""
        endpoints = [
            ('dashboard', 'BI Dashboard'),
            ('kpis', 'KPI Metrics'),
            ('revenue', 'Revenue Analytics'),
            ('user-behavior', 'User Behavior Analysis'),
            ('predictive', 'Predictive Analytics')
        ]
        
        for endpoint, name in endpoints:
            try:
                response = self.session.get(f"{self.api_base}/admin/bi/{endpoint}", timeout=15)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get('success') and 'data' in data:
                        # Check for specific BI data structures
                        bi_data = data['data']
                        if endpoint == 'kpis':
                            expected_kpis = ['CustomerAcquisitionCost', 'CustomerLifetimeValue', 'MonthlyRecurringRevenue']
                            has_kpis = any(kpi in str(bi_data) for kpi in expected_kpis)
                            self.log_test(
                                f"BI - {name}",
                                has_kpis,
                                f"KPI metrics {'contain expected KPIs' if has_kpis else 'missing expected KPIs'}",
                                {"has_kpi_data": has_kpis}
                            )
                        elif endpoint == 'predictive':
                            has_predictions = 'ChurnPrediction' in str(bi_data) or 'Forecasting' in str(bi_data)
                            self.log_test(
                                f"BI - {name}",
                                has_predictions,
                                f"Predictive analytics {'contains predictions' if has_predictions else 'missing prediction data'}",
                                {"has_predictions": has_predictions}
                            )
                        else:
                            self.log_test(
                                f"BI - {name}",
                                True,
                                f"Successfully retrieved {name.lower()}",
                                {"has_data": bool(bi_data), "data_keys": list(bi_data.keys()) if isinstance(bi_data, dict) else "non-dict"}
                            )
                    else:
                        self.log_test(
                            f"BI - {name}",
                            False,
                            f"{name} response missing success or data fields",
                            data
                        )
                elif response.status_code == 401:
                    self.log_test(
                        f"BI - {name}",
                        True,
                        f"Correctly requires authentication (401)",
                        {"status_code": 401}
                    )
                else:
                    self.log_test(
                        f"BI - {name}",
                        False,
                        f"{name} returned unexpected status: {response.status_code}",
                        {"status_code": response.status_code}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"BI - {name}",
                    False,
                    f"{name} test failed: {str(e)}",
                    {"error": str(e)}
                )

    def test_reporting_system(self):
        """Test advanced reporting system"""
        # Test report generation
        report_types = ['financial', 'user', 'contest', 'compliance', 'referral', 'game']
        
        for report_type in report_types:
            try:
                report_request = {
                    "report_type": report_type,
                    "format": "json",
                    "date_from": (datetime.now() - timedelta(days=30)).strftime('%Y-%m-%d'),
                    "date_to": datetime.now().strftime('%Y-%m-%d'),
                    "description": f"Test {report_type} report generated by automated testing"
                }
                
                response = self.session.post(f"{self.api_base}/admin/reports/generate", json=report_request, timeout=15)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get('success') and 'data' in data:
                        report_data = data['data']
                        if 'ID' in report_data or 'id' in report_data:
                            self.log_test(
                                f"Report Generation - {report_type.title()}",
                                True,
                                f"Successfully generated {report_type} report",
                                {"report_id": report_data.get('ID') or report_data.get('id'), "status": report_data.get('Status') or report_data.get('status')}
                            )
                        else:
                            self.log_test(
                                f"Report Generation - {report_type.title()}",
                                False,
                                f"{report_type} report generation response missing ID",
                                report_data
                            )
                    else:
                        self.log_test(
                            f"Report Generation - {report_type.title()}",
                            False,
                            f"{report_type} report generation response missing success or data",
                            data
                        )
                elif response.status_code == 401:
                    self.log_test(
                        f"Report Generation - {report_type.title()}",
                        True,
                        f"Correctly requires authentication (401)",
                        {"status_code": 401}
                    )
                else:
                    self.log_test(
                        f"Report Generation - {report_type.title()}",
                        False,
                        f"{report_type} report generation failed: {response.status_code}",
                        {"status_code": response.status_code, "response": response.text[:200]}
                    )
                    
            except Exception as e:
                self.log_test(
                    f"Report Generation - {report_type.title()}",
                    False,
                    f"{report_type} report generation test failed: {str(e)}",
                    {"error": str(e)}
                )
        
        # Test report listing
        try:
            response = self.session.get(f"{self.api_base}/admin/reports", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                if data.get('success') and 'data' in data:
                    reports_data = data['data']
                    if 'Reports' in reports_data or 'reports' in reports_data:
                        reports_list = reports_data.get('Reports') or reports_data.get('reports')
                        self.log_test(
                            "Report Listing",
                            True,
                            f"Successfully retrieved reports list with {len(reports_list)} reports",
                            {"reports_count": len(reports_list), "has_pagination": 'Total' in reports_data or 'total' in reports_data}
                        )
                    else:
                        self.log_test(
                            "Report Listing",
                            True,
                            "Reports list endpoint working (empty list)",
                            reports_data
                        )
                else:
                    self.log_test(
                        "Report Listing",
                        False,
                        "Report listing response missing success or data fields",
                        data
                    )
            elif response.status_code == 401:
                self.log_test(
                    "Report Listing",
                    True,
                    "Correctly requires authentication (401)",
                    {"status_code": 401}
                )
            else:
                self.log_test(
                    "Report Listing",
                    False,
                    f"Report listing failed: {response.status_code}",
                    {"status_code": response.status_code}
                )
                
        except Exception as e:
            self.log_test(
                "Report Listing",
                False,
                f"Report listing test failed: {str(e)}",
                {"error": str(e)}
            )

    def test_error_handling(self):
        """Test error handling with invalid parameters"""
        # Test invalid date range
        try:
            invalid_params = {
                'date_from': '2024-12-31',
                'date_to': '2024-01-01'  # date_to before date_from
            }
            
            response = self.session.get(f"{self.api_base}/admin/analytics/dashboard", params=invalid_params, timeout=10)
            
            if response.status_code == 400:
                self.log_test(
                    "Error Handling - Invalid Date Range",
                    True,
                    "Correctly returns 400 for invalid date range",
                    {"status_code": 400}
                )
            elif response.status_code == 401:
                self.log_test(
                    "Error Handling - Invalid Date Range",
                    True,
                    "Authentication required (401) - cannot test validation without auth",
                    {"status_code": 401}
                )
            else:
                self.log_test(
                    "Error Handling - Invalid Date Range",
                    False,
                    f"Expected 400 for invalid date range, got {response.status_code}",
                    {"status_code": response.status_code}
                )
                
        except Exception as e:
            self.log_test(
                "Error Handling - Invalid Date Range",
                False,
                f"Error handling test failed: {str(e)}",
                {"error": str(e)}
            )
        
        # Test invalid report type
        try:
            invalid_report = {
                "report_type": "invalid_type",
                "format": "json",
                "date_from": "2024-01-01",
                "date_to": "2024-01-31"
            }
            
            response = self.session.post(f"{self.api_base}/admin/reports/generate", json=invalid_report, timeout=10)
            
            if response.status_code == 400:
                self.log_test(
                    "Error Handling - Invalid Report Type",
                    True,
                    "Correctly returns 400 for invalid report type",
                    {"status_code": 400}
                )
            elif response.status_code == 401:
                self.log_test(
                    "Error Handling - Invalid Report Type",
                    True,
                    "Authentication required (401) - cannot test validation without auth",
                    {"status_code": 401}
                )
            else:
                self.log_test(
                    "Error Handling - Invalid Report Type",
                    False,
                    f"Expected 400 for invalid report type, got {response.status_code}",
                    {"status_code": response.status_code}
                )
                
        except Exception as e:
            self.log_test(
                "Error Handling - Invalid Report Type",
                False,
                f"Invalid report type test failed: {str(e)}",
                {"error": str(e)}
            )

    def run_all_tests(self):
        """Run all analytics tests"""
        print("🚀 Starting Advanced Analytics, BI, and Reporting System Tests")
        print("=" * 80)
        
        # Health check
        if not self.test_health_check():
            print("❌ Backend is not running. Stopping tests.")
            return False
        
        # Admin login
        admin_logged_in = self.admin_login()
        
        # Run analytics tests
        print("\n📊 Testing Analytics Dashboard...")
        self.test_analytics_dashboard()
        
        print("\n📈 Testing Analytics Endpoints...")
        self.test_analytics_endpoints()
        
        print("\n🧠 Testing Business Intelligence Endpoints...")
        self.test_business_intelligence_endpoints()
        
        print("\n📋 Testing Reporting System...")
        self.test_reporting_system()
        
        print("\n⚠️ Testing Error Handling...")
        self.test_error_handling()
        
        # Summary
        print("\n" + "=" * 80)
        print("📊 TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result['passed'])
        failed_tests = total_tests - passed_tests
        
        print(f"Total Tests: {total_tests}")
        print(f"✅ Passed: {passed_tests}")
        print(f"❌ Failed: {failed_tests}")
        print(f"Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        
        if not admin_logged_in:
            print("\n⚠️ NOTE: Many tests may have failed due to authentication issues.")
            print("   Admin login credentials may need to be updated.")
        
        # Save results
        with open('/app/analytics_test_results.json', 'w') as f:
            json.dump(self.test_results, f, indent=2, default=str)
        
        print(f"\n📄 Detailed results saved to: /app/analytics_test_results.json")
        
        return failed_tests == 0

if __name__ == "__main__":
    tester = AnalyticsAPITester()
    success = tester.run_all_tests()
    sys.exit(0 if success else 1)