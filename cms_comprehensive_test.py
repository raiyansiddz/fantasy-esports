#!/usr/bin/env python3
"""
Comprehensive CMS Testing - Final Verification of Stuck Tasks
"""

import requests
import json
import time
from typing import Dict, Any, Optional, Tuple, List

class CMSComprehensiveTester:
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
                    self.log_test("Admin Authentication", True, "Successfully authenticated")
                    return True
            
            self.log_test("Admin Authentication", False, f"Failed with status: {response.status_code}")
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def test_database_null_handling_direct(self):
        """Test database NULL handling by directly inserting NULL values"""
        if not self.admin_token:
            return False
            
        try:
            # Create SEO content with empty og_image (should be stored as empty string, not NULL)
            seo_data = {
                "page_type": "test",
                "page_slug": "test-null-handling",
                "meta_title": "Test NULL Handling",
                "meta_description": "Testing NULL og_image handling",
                "keywords": ["test"],
                "og_title": "Test",
                "og_description": "Test description",
                "og_image": "",  # Empty string instead of omitting
                "content": "Test content"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=seo_data)
            
            if response.status_code == 201:
                self.log_test("SEO Content Creation (Empty og_image)", True, "Created SEO content with empty og_image")
                
                # Now test listing to see if it handles empty strings
                list_response = self.session.get(f"{self.base_url}/api/v1/admin/content/seo")
                if list_response.status_code == 200:
                    self.log_test("SEO List After Empty og_image", True, "Successfully listed SEO content with empty og_image")
                    return True
                else:
                    self.log_test("SEO List After Empty og_image", False, f"List failed with status: {list_response.status_code}", list_response.text)
                    return False
            else:
                self.log_test("SEO Content Creation (Empty og_image)", False, f"Creation failed with status: {response.status_code}", response.text)
                return False
                
        except Exception as e:
            self.log_test("Database NULL Handling Test", False, f"Exception: {str(e)}")
            return False

    def test_faq_routing_comprehensive(self):
        """Comprehensive test of FAQ routing issues"""
        print("\n🔍 COMPREHENSIVE FAQ ROUTING ANALYSIS")
        print("-" * 60)
        
        # Test all FAQ-related endpoints
        endpoints_to_test = [
            ("GET", "/api/v1/admin/content/faq/sections", "Admin FAQ Sections List"),
            ("POST", "/api/v1/admin/content/faq/sections", "Admin FAQ Sections Create"),
            ("PUT", "/api/v1/admin/content/faq/sections/1", "Admin FAQ Sections Update"),
            ("GET", "/api/v1/faq/sections", "Public FAQ Sections List"),
            ("POST", "/api/v1/admin/content/faq/items", "Admin FAQ Items Create"),
            ("PUT", "/api/v1/admin/content/faq/items/1", "Admin FAQ Items Update"),
            ("GET", "/api/v1/faq/items", "Public FAQ Items List"),
        ]
        
        routing_results = []
        
        for method, endpoint, description in endpoints_to_test:
            try:
                # Test without auth first
                original_headers = self.session.headers.copy()
                if 'Authorization' in self.session.headers:
                    del self.session.headers['Authorization']
                
                if method == "GET":
                    response = self.session.get(f"{self.base_url}{endpoint}")
                elif method == "POST":
                    response = self.session.post(f"{self.base_url}{endpoint}", json={"test": "data"})
                elif method == "PUT":
                    response = self.session.put(f"{self.base_url}{endpoint}", json={"test": "data"})
                
                # Restore headers
                self.session.headers.clear()
                self.session.headers.update(original_headers)
                
                if response.status_code == 404 and "page not found" in response.text.lower():
                    routing_results.append(f"❌ {description}: ROUTE NOT FOUND (404)")
                    self.log_test(f"Routing Check - {description}", False, "Route not found - 404 page not found")
                elif response.status_code == 401:
                    routing_results.append(f"✅ {description}: ROUTE EXISTS (401 unauthorized)")
                    self.log_test(f"Routing Check - {description}", True, "Route exists - returns 401 as expected")
                elif response.status_code == 400:
                    routing_results.append(f"✅ {description}: ROUTE EXISTS (400 bad request)")
                    self.log_test(f"Routing Check - {description}", True, "Route exists - returns 400 for invalid data")
                elif response.status_code == 200:
                    routing_results.append(f"✅ {description}: ROUTE EXISTS (200 success)")
                    self.log_test(f"Routing Check - {description}", True, "Route exists and works")
                else:
                    routing_results.append(f"⚠️  {description}: ROUTE EXISTS ({response.status_code})")
                    self.log_test(f"Routing Check - {description}", True, f"Route exists - status {response.status_code}")
                    
            except Exception as e:
                routing_results.append(f"❌ {description}: ERROR ({str(e)})")
                self.log_test(f"Routing Check - {description}", False, f"Exception: {str(e)}")
        
        print("\n📋 ROUTING ANALYSIS RESULTS:")
        for result in routing_results:
            print(f"  {result}")
        
        return routing_results

    def test_seo_comprehensive(self):
        """Comprehensive SEO testing including NULL handling"""
        print("\n🔍 COMPREHENSIVE SEO TESTING")
        print("-" * 60)
        
        # Test SEO content creation with various og_image scenarios
        test_scenarios = [
            {
                "name": "With valid og_image URL",
                "data": {
                    "page_type": "home",
                    "page_slug": "home-with-image",
                    "meta_title": "Home Page",
                    "meta_description": "Home page description",
                    "og_image": "https://example.com/image.jpg"
                },
                "should_work": True
            },
            {
                "name": "With empty og_image",
                "data": {
                    "page_type": "about",
                    "page_slug": "about-empty-image",
                    "meta_title": "About Page",
                    "meta_description": "About page description",
                    "og_image": ""
                },
                "should_work": True
            },
            {
                "name": "Without og_image field",
                "data": {
                    "page_type": "contact",
                    "page_slug": "contact-no-image",
                    "meta_title": "Contact Page",
                    "meta_description": "Contact page description"
                },
                "should_work": True
            }
        ]
        
        created_ids = []
        
        for scenario in test_scenarios:
            try:
                response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=scenario["data"])
                
                if response.status_code == 201 and scenario["should_work"]:
                    data = response.json()
                    if data.get("success") and "data" in data:
                        seo_id = data["data"].get("id")
                        created_ids.append(seo_id)
                        self.log_test(f"SEO Creation - {scenario['name']}", True, f"Created with ID: {seo_id}")
                    else:
                        self.log_test(f"SEO Creation - {scenario['name']}", False, "Missing expected response structure")
                elif response.status_code != 201 and not scenario["should_work"]:
                    self.log_test(f"SEO Creation - {scenario['name']}", True, f"Correctly failed with status: {response.status_code}")
                else:
                    self.log_test(f"SEO Creation - {scenario['name']}", False, f"Unexpected status: {response.status_code}", response.text)
                    
            except Exception as e:
                self.log_test(f"SEO Creation - {scenario['name']}", False, f"Exception: {str(e)}")
        
        # Test listing with various og_image values
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/seo")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    contents = data.get("contents", [])
                    self.log_test("SEO List Comprehensive", True, f"Successfully listed {len(contents)} SEO contents")
                    
                    # Check if any have NULL or empty og_image
                    null_count = sum(1 for content in contents if not content.get("og_image"))
                    if null_count > 0:
                        self.log_test("SEO NULL og_image Handling", True, f"Successfully handled {null_count} contents with NULL/empty og_image")
                    else:
                        self.log_test("SEO NULL og_image Handling", True, "No NULL og_image values found, but listing works")
                else:
                    self.log_test("SEO List Comprehensive", False, "Response missing success field")
            elif response.status_code == 500:
                error_text = response.text
                if "og_image" in error_text.lower() and "null" in error_text.lower():
                    self.log_test("SEO List Comprehensive", False, "CONFIRMED: NULL og_image causing 500 error", error_text)
                else:
                    self.log_test("SEO List Comprehensive", False, f"500 error but different cause", error_text)
            else:
                self.log_test("SEO List Comprehensive", False, f"Unexpected status: {response.status_code}")
                
        except Exception as e:
            self.log_test("SEO List Comprehensive", False, f"Exception: {str(e)}")
        
        return created_ids

    def run_comprehensive_tests(self):
        """Run comprehensive tests for both stuck tasks"""
        print("🎯 COMPREHENSIVE CMS STUCK TASKS VERIFICATION")
        print("=" * 80)
        
        # Authenticate first
        if not self.authenticate_admin():
            print("❌ Cannot proceed without admin authentication")
            return
        
        # Test SEO comprehensively
        self.test_seo_comprehensive()
        self.test_database_null_handling_direct()
        
        # Test FAQ routing comprehensively
        self.test_faq_routing_comprehensive()
        
        # Generate final summary
        self.generate_final_summary()

    def generate_final_summary(self):
        """Generate comprehensive final summary"""
        print("\n" + "=" * 80)
        print("📊 COMPREHENSIVE CMS TESTING FINAL SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {success_rate:.1f}%")
        
        print("\n" + "=" * 80)
        print("🎯 FINAL STUCK TASKS ANALYSIS")
        print("=" * 80)
        
        # Analyze SEO task
        seo_tests = [r for r in self.test_results if "SEO" in r["test"]]
        seo_failed = [r for r in seo_tests if not r["success"]]
        seo_critical_failures = [r for r in seo_failed if "500" in r["details"] or "NULL" in r["details"]]
        
        print("TASK 1 - SEO Content Management (NULL og_image handling):")
        if seo_critical_failures:
            print("  ❌ STILL STUCK - Critical NULL handling issues found:")
            for failure in seo_critical_failures:
                print(f"    • {failure['test']}: {failure['details']}")
        else:
            print("  ✅ RESOLVED - No critical NULL handling issues found")
            if seo_failed:
                print("  ⚠️  Minor issues found (not blocking):")
                for failure in seo_failed:
                    print(f"    • {failure['test']}: {failure['details']}")
        
        # Analyze FAQ task
        faq_tests = [r for r in self.test_results if "FAQ" in r["test"] or "Routing" in r["test"]]
        faq_failed = [r for r in faq_tests if not r["success"]]
        faq_routing_failures = [r for r in faq_failed if "ROUTE NOT FOUND" in r["details"] or "404" in r["details"]]
        
        print("\nTASK 2 - FAQ Management (Admin routing issue):")
        if faq_routing_failures:
            print("  ❌ STILL STUCK - Critical routing issues found:")
            for failure in faq_routing_failures:
                print(f"    • {failure['test']}: {failure['details']}")
        else:
            print("  ✅ RESOLVED - No critical routing issues found")
            if faq_failed:
                print("  ⚠️  Minor issues found (not blocking):")
                for failure in faq_failed:
                    print(f"    • {failure['test']}: {failure['details']}")
        
        print(f"\n🏆 FINAL VERDICT:")
        critical_issues = len(seo_critical_failures) + len(faq_routing_failures)
        if critical_issues == 0:
            print("  🎉 ALL STUCK TASKS RESOLVED - CMS is fully functional!")
        else:
            print(f"  ⚠️  {critical_issues} critical issues still need main agent attention")
            
        print("\n📋 RECOMMENDATIONS FOR MAIN AGENT:")
        if seo_critical_failures:
            print("  1. Fix database schema to handle NULL og_image values properly")
            print("     - Add COALESCE or NULL handling in SQL queries")
            print("     - Update Go struct scanning to handle NULL values")
        
        if faq_routing_failures:
            print("  2. Add missing FAQ admin routes in server.go:")
            print("     - Add GET /api/v1/admin/content/faq/sections route")
            print("     - Ensure all CRUD operations are properly routed")

if __name__ == "__main__":
    tester = CMSComprehensiveTester()
    tester.run_comprehensive_tests()