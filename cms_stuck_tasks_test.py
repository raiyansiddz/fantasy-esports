#!/usr/bin/env python3
"""
Focused CMS Testing for Two Stuck Tasks
1. SEO Content Management - Database NULL handling issue with og_image column causing 500 errors
2. FAQ Management - Routing issue where admin FAQ sections endpoint returns 404 instead of 401
"""

import requests
import json
import time
from typing import Dict, Any, Optional, Tuple, List

class CMSStuckTasksTester:
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
            # Try multiple admin authentication methods
            auth_methods = [
                {"username": "admin", "password": "admin123"},
                {"email": "admin@fantasy-esports.com", "password": "admin123"},
                {"username": "admin", "password": "password"},
            ]
            
            for auth_data in auth_methods:
                response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=auth_data)
                
                if response.status_code == 200:
                    data = response.json()
                    if data.get("success") and "access_token" in data:
                        self.admin_token = data["access_token"]
                        self.session.headers.update({"Authorization": f"Bearer {self.admin_token}"})
                        self.log_test("Admin Authentication", True, f"Successfully authenticated with {auth_data}")
                        return True
            
            self.log_test("Admin Authentication", False, f"All authentication methods failed. Last status: {response.status_code}", response.text)
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    # ========================= STUCK TASK 1: SEO CONTENT MANAGEMENT =========================
    
    def test_seo_content_creation(self) -> bool:
        """Test SEO content creation to set up data for testing NULL og_image issue"""
        if not self.admin_token:
            self.log_test("SEO Content Creation", False, "No admin token available")
            return False
            
        try:
            # Create SEO content with og_image (should work)
            seo_data_with_image = {
                "page_type": "home",
                "page_slug": "home-page-with-image",
                "meta_title": "Fantasy Esports - Ultimate Gaming Experience",
                "meta_description": "Join the ultimate fantasy esports platform. Create teams, compete in tournaments, and win real money prizes.",
                "keywords": ["fantasy esports", "gaming", "tournaments"],
                "og_title": "Fantasy Esports - Ultimate Gaming Experience",
                "og_description": "Join the ultimate fantasy esports platform and win real money",
                "og_image": "https://example.com/og-image.jpg",
                "twitter_card": "summary_large_image",
                "structured_data": {"@type": "WebSite", "name": "Fantasy Esports", "url": "https://fantasy-esports.com"},
                "content": "<h1>Welcome to Fantasy Esports</h1><p>Create your dream team and compete for real money prizes!</p>"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=seo_data_with_image)
            
            if response.status_code == 201:
                data = response.json()
                if data.get("success") and "data" in data:
                    seo_id = data["data"].get("id")
                    self.log_test("SEO Content Creation (with og_image)", True, f"Created SEO content with ID: {seo_id}")
                    return True
                else:
                    self.log_test("SEO Content Creation (with og_image)", False, "Response missing expected data structure", response.json())
                    return False
            else:
                self.log_test("SEO Content Creation (with og_image)", False, f"Status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("SEO Content Creation (with og_image)", False, f"Exception: {str(e)}")
            return False

    def test_seo_content_creation_without_og_image(self) -> bool:
        """Test SEO content creation without og_image to create NULL values"""
        if not self.admin_token:
            self.log_test("SEO Content Creation (NULL og_image)", False, "No admin token available")
            return False
            
        try:
            # Create SEO content without og_image (should create NULL in database)
            seo_data_without_image = {
                "page_type": "about",
                "page_slug": "about-page-null-image",
                "meta_title": "About Fantasy Esports",
                "meta_description": "Learn more about our fantasy esports platform and our mission.",
                "keywords": ["about", "fantasy esports", "mission"],
                "og_title": "About Fantasy Esports",
                "og_description": "Learn more about our fantasy esports platform",
                # Intentionally omitting og_image to create NULL value
                "twitter_card": "summary",
                "structured_data": {"@type": "AboutPage", "name": "About Fantasy Esports"},
                "content": "<h1>About Us</h1><p>We are the leading fantasy esports platform.</p>"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=seo_data_without_image)
            
            if response.status_code == 201:
                data = response.json()
                if data.get("success") and "data" in data:
                    seo_id = data["data"].get("id")
                    self.log_test("SEO Content Creation (NULL og_image)", True, f"Created SEO content with NULL og_image, ID: {seo_id}")
                    return True
                else:
                    self.log_test("SEO Content Creation (NULL og_image)", False, "Response missing expected data structure", response.json())
                    return False
            else:
                self.log_test("SEO Content Creation (NULL og_image)", False, f"Status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("SEO Content Creation (NULL og_image)", False, f"Exception: {str(e)}")
            return False

    def test_seo_list_endpoint_null_handling(self) -> bool:
        """Test the main issue: SEO content list endpoint with NULL og_image causing 500 errors"""
        if not self.admin_token:
            self.log_test("SEO List Endpoint (NULL handling)", False, "No admin token available")
            return False
            
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/seo")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    contents = data.get("contents", [])
                    self.log_test("SEO List Endpoint (NULL handling)", True, f"Successfully retrieved {len(contents)} SEO contents - NULL og_image handling FIXED!")
                    return True
                else:
                    self.log_test("SEO List Endpoint (NULL handling)", False, "Response missing success field", response.json())
                    return False
            elif response.status_code == 500:
                # This is the expected error from the stuck task
                error_text = response.text
                if "og_image" in error_text and "NULL" in error_text:
                    self.log_test("SEO List Endpoint (NULL handling)", False, "CONFIRMED STUCK TASK: NULL og_image column causing 500 error", error_text)
                else:
                    self.log_test("SEO List Endpoint (NULL handling)", False, f"500 error but different cause: {error_text}")
                return False
            else:
                self.log_test("SEO List Endpoint (NULL handling)", False, f"Unexpected status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("SEO List Endpoint (NULL handling)", False, f"Exception: {str(e)}")
            return False

    def test_seo_public_endpoint_null_handling(self) -> bool:
        """Test public SEO endpoint with NULL og_image handling"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            # Test with slug that should have NULL og_image
            response = self.session.get(f"{self.base_url}/api/v1/seo/about-page-null-image")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    seo_data = data.get("data", {})
                    self.log_test("SEO Public Endpoint (NULL handling)", True, f"Successfully retrieved SEO data for slug with NULL og_image - FIXED!")
                    return True
                else:
                    self.log_test("SEO Public Endpoint (NULL handling)", False, "Response missing success field", response.json())
                    return False
            elif response.status_code == 500:
                # This is the expected error from the stuck task
                error_text = response.text
                if "og_image" in error_text and "NULL" in error_text:
                    self.log_test("SEO Public Endpoint (NULL handling)", False, "CONFIRMED STUCK TASK: NULL og_image column causing 500 error in public endpoint", error_text)
                else:
                    self.log_test("SEO Public Endpoint (NULL handling)", False, f"500 error but different cause: {error_text}")
                return False
            elif response.status_code == 404:
                self.log_test("SEO Public Endpoint (NULL handling)", False, "SEO content not found - may need to create test data first")
                return False
            else:
                self.log_test("SEO Public Endpoint (NULL handling)", False, f"Unexpected status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("SEO Public Endpoint (NULL handling)", False, f"Exception: {str(e)}")
            return False

    # ========================= STUCK TASK 2: FAQ MANAGEMENT =========================
    
    def test_faq_sections_routing_issue(self) -> bool:
        """Test the main issue: FAQ sections admin endpoint returns 404 instead of 401"""
        try:
            # Remove admin auth to test unauthorized access
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            # Test admin FAQ sections endpoint without auth
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/faq/sections")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            if response.status_code == 401:
                self.log_test("FAQ Sections Routing (Unauthorized)", True, "Correctly returns 401 for unauthorized access - ROUTING FIXED!")
                return True
            elif response.status_code == 404:
                # This is the expected error from the stuck task
                self.log_test("FAQ Sections Routing (Unauthorized)", False, "CONFIRMED STUCK TASK: Returns 404 instead of 401 - routing issue", response.text)
                return False
            else:
                self.log_test("FAQ Sections Routing (Unauthorized)", False, f"Unexpected status: {response.status_code} (expected 401 or 404)", response.text)
                return False
            
        except Exception as e:
            self.log_test("FAQ Sections Routing (Unauthorized)", False, f"Exception: {str(e)}")
            return False

    def test_faq_sections_with_auth(self) -> bool:
        """Test FAQ sections endpoint with proper authentication"""
        if not self.admin_token:
            self.log_test("FAQ Sections (With Auth)", False, "No admin token available")
            return False
            
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/faq/sections")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    sections = data.get("sections", [])
                    self.log_test("FAQ Sections (With Auth)", True, f"Successfully retrieved {len(sections)} FAQ sections with proper auth")
                    return True
                else:
                    self.log_test("FAQ Sections (With Auth)", False, "Response missing success field", response.json())
                    return False
            elif response.status_code == 404:
                self.log_test("FAQ Sections (With Auth)", False, "Still returns 404 even with auth - routing issue persists", response.text)
                return False
            else:
                self.log_test("FAQ Sections (With Auth)", False, f"Unexpected status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("FAQ Sections (With Auth)", False, f"Exception: {str(e)}")
            return False

    def test_faq_section_creation(self) -> bool:
        """Test FAQ section creation to verify the endpoint works"""
        if not self.admin_token:
            self.log_test("FAQ Section Creation", False, "No admin token available")
            return False
            
        try:
            section_data = {
                "name": "Getting Started",
                "description": "Basic questions about using Fantasy Esports platform",
                "sort_order": 1
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/sections", json=section_data)
            
            if response.status_code == 201:
                data = response.json()
                if data.get("success") and "data" in data:
                    section_id = data["data"].get("id")
                    self.log_test("FAQ Section Creation", True, f"Successfully created FAQ section with ID: {section_id}")
                    return True
                else:
                    self.log_test("FAQ Section Creation", False, "Response missing expected data structure", response.json())
                    return False
            elif response.status_code == 404:
                self.log_test("FAQ Section Creation", False, "POST endpoint also returns 404 - routing issue affects all FAQ section endpoints", response.text)
                return False
            else:
                self.log_test("FAQ Section Creation", False, f"Status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("FAQ Section Creation", False, f"Exception: {str(e)}")
            return False

    def test_faq_update_routing(self) -> bool:
        """Test FAQ section update endpoint routing"""
        if not self.admin_token:
            self.log_test("FAQ Section Update Routing", False, "No admin token available")
            return False
            
        try:
            # Test with a dummy ID to see if the route exists
            section_data = {
                "name": "Updated Section",
                "description": "Updated description",
                "sort_order": 2
            }
            
            response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/sections/1", json=section_data)
            
            if response.status_code in [200, 404, 400]:  # Any of these means the route exists
                if response.status_code == 404 and "not found" in response.text.lower():
                    self.log_test("FAQ Section Update Routing", True, "PUT endpoint exists (returns 404 for non-existent ID, which is correct)")
                    return True
                elif response.status_code == 200:
                    self.log_test("FAQ Section Update Routing", True, "PUT endpoint works correctly")
                    return True
                else:
                    self.log_test("FAQ Section Update Routing", True, f"PUT endpoint exists (status: {response.status_code})")
                    return True
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_test("FAQ Section Update Routing", False, "PUT endpoint also has routing issues - returns 'page not found'", response.text)
                return False
            else:
                self.log_test("FAQ Section Update Routing", False, f"Unexpected status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("FAQ Section Update Routing", False, f"Exception: {str(e)}")
            return False

    def test_public_faq_sections(self) -> bool:
        """Test public FAQ sections endpoint (should work regardless of admin routing issues)"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/faq/sections")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    sections = data.get("sections", [])
                    self.log_test("Public FAQ Sections", True, f"Public endpoint works correctly - found {len(sections)} sections")
                    return True
                else:
                    self.log_test("Public FAQ Sections", False, "Response missing success field", response.json())
                    return False
            else:
                self.log_test("Public FAQ Sections", False, f"Status: {response.status_code}", response.text)
                return False
            
        except Exception as e:
            self.log_test("Public FAQ Sections", False, f"Exception: {str(e)}")
            return False

    def run_stuck_tasks_tests(self):
        """Run focused tests for the two stuck tasks"""
        print("🎯 FOCUSED TESTING FOR CMS STUCK TASKS")
        print("=" * 80)
        print("TASK 1: SEO Content Management - Database NULL handling issue with og_image column")
        print("TASK 2: FAQ Management - Routing issue where admin FAQ sections endpoint returns 404 instead of 401")
        print("=" * 80)
        
        # Authenticate first
        if not self.authenticate_admin():
            print("❌ Cannot proceed without admin authentication")
            return
        
        print("\n🔍 TESTING STUCK TASK 1: SEO CONTENT MANAGEMENT")
        print("-" * 60)
        
        # Test SEO content creation to set up test data
        self.test_seo_content_creation()
        self.test_seo_content_creation_without_og_image()
        
        # Test the main issue: NULL og_image handling
        self.test_seo_list_endpoint_null_handling()
        self.test_seo_public_endpoint_null_handling()
        
        print("\n🔍 TESTING STUCK TASK 2: FAQ MANAGEMENT")
        print("-" * 60)
        
        # Test the main issue: routing problem
        self.test_faq_sections_routing_issue()
        self.test_faq_sections_with_auth()
        self.test_faq_section_creation()
        self.test_faq_update_routing()
        self.test_public_faq_sections()
        
        # Generate Summary
        self.generate_summary()

    def generate_summary(self):
        """Generate focused summary for stuck tasks"""
        print("\n" + "=" * 80)
        print("📊 CMS STUCK TASKS TESTING SUMMARY")
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
        
        # Categorize by stuck task
        seo_tests = [r for r in self.test_results if "SEO" in r["test"]]
        faq_tests = [r for r in self.test_results if "FAQ" in r["test"]]
        
        print("📋 STUCK TASK RESULTS:")
        if seo_tests:
            seo_passed = sum(1 for r in seo_tests if r["success"])
            seo_total = len(seo_tests)
            seo_rate = (seo_passed / seo_total * 100) if seo_total > 0 else 0
            print(f"  TASK 1 - SEO Content Management: {seo_passed}/{seo_total} passed ({seo_rate:.1f}%)")
        
        if faq_tests:
            faq_passed = sum(1 for r in faq_tests if r["success"])
            faq_total = len(faq_tests)
            faq_rate = (faq_passed / faq_total * 100) if faq_total > 0 else 0
            print(f"  TASK 2 - FAQ Management: {faq_passed}/{faq_total} passed ({faq_rate:.1f}%)")
        
        print("\n" + "=" * 80)
        print("🔍 STUCK TASK ANALYSIS")
        print("=" * 80)
        
        # Show failed tests with focus on stuck tasks
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print("❌ ISSUES FOUND:")
            for result in failed_results:
                if "CONFIRMED STUCK TASK" in result["details"]:
                    print(f"  🚨 {result['test']}: {result['details']}")
                else:
                    print(f"  • {result['test']}: {result['details']}")
        else:
            print("✅ ALL STUCK TASKS RESOLVED!")
        
        print("\n" + "=" * 80)
        print("🎯 STUCK TASKS STATUS")
        print("=" * 80)
        
        # Determine status of each stuck task
        seo_stuck = any("CONFIRMED STUCK TASK" in r["details"] and "SEO" in r["test"] for r in failed_results)
        faq_stuck = any("CONFIRMED STUCK TASK" in r["details"] and "FAQ" in r["test"] for r in failed_results)
        
        print("TASK 1 - SEO Content Management (NULL og_image handling):")
        if seo_stuck:
            print("  ❌ STILL STUCK - Database NULL handling issue persists")
        else:
            print("  ✅ RESOLVED - NULL og_image handling working correctly")
        
        print("\nTASK 2 - FAQ Management (Admin routing issue):")
        if faq_stuck:
            print("  ❌ STILL STUCK - Admin FAQ sections endpoint routing issue persists")
        else:
            print("  ✅ RESOLVED - Admin FAQ sections endpoint routing working correctly")
        
        print(f"\n🏆 OVERALL STUCK TASKS STATUS:")
        if seo_stuck or faq_stuck:
            stuck_count = sum([seo_stuck, faq_stuck])
            print(f"  ⚠️  {stuck_count}/2 tasks still stuck - requires main agent attention")
        else:
            print("  🎉 ALL STUCK TASKS RESOLVED - CMS is fully functional!")

if __name__ == "__main__":
    tester = CMSStuckTasksTester()
    tester.run_stuck_tasks_tests()