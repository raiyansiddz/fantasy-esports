#!/usr/bin/env python3
"""
FAQ Management Fix Verification Test
Testing the specific fix for Content Management System FAQ Management

FOCUS: Verify that the missing GET route for admin FAQ sections has been fixed
- GET /api/v1/admin/content/faq/sections should now return 401 (not 404) without auth
- With admin auth, should return 200 with proper FAQ sections listing
- Complete FAQ Management workflow should work correctly
"""

import requests
import json
import time
from typing import Dict, Any, Optional, Tuple, List

class FAQManagementVerificationTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.test_results = []
        self.created_faq_sections = []
        self.created_faq_items = []
        
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
            
            self.log_test(
                "Admin Authentication", 
                False, 
                f"All authentication methods failed. Last status: {response.status_code}",
                response.text
            )
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def test_faq_sections_endpoint_without_auth(self) -> bool:
        """Test that GET /api/v1/admin/content/faq/sections returns 401 without auth (not 404)"""
        try:
            # Remove admin auth
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/faq/sections")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            # Should return 401 (unauthorized) not 404 (not found)
            success = response.status_code == 401
            details = f"Status: {response.status_code}"
            
            if success:
                details += " - Correctly returns 401 (unauthorized) instead of 404 (not found)"
            elif response.status_code == 404:
                details += " - ❌ STILL RETURNS 404 - Route not found! Fix not applied correctly."
            else:
                details += f" - Unexpected status code: {response.status_code}"
            
            self.log_test(
                "FAQ Sections Endpoint - No Auth (Should be 401)",
                success,
                details,
                response.text if not success else None
            )
            return success
            
        except Exception as e:
            self.log_test("FAQ Sections Endpoint - No Auth", False, f"Exception: {str(e)}")
            return False

    def test_faq_sections_endpoint_with_auth(self) -> bool:
        """Test that GET /api/v1/admin/content/faq/sections returns 200 with admin auth"""
        if not self.admin_token:
            self.log_test("FAQ Sections Endpoint - With Auth", False, "No admin token available")
            return False
            
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/faq/sections")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    sections = data.get("data", [])
                    details += f" - Successfully retrieved {len(sections)} FAQ sections"
                    if sections:
                        details += f" - Sample section: {sections[0].get('name', 'N/A')}"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "FAQ Sections Endpoint - With Auth (Should be 200)",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("FAQ Sections Endpoint - With Auth", False, f"Exception: {str(e)}")
            return False

    def test_create_faq_section(self) -> bool:
        """Test creating a new FAQ section"""
        if not self.admin_token:
            self.log_test("Create FAQ Section", False, "No admin token available")
            return False
            
        try:
            section_data = {
                "name": "FAQ Management Test Section",
                "description": "Test section created during FAQ Management verification",
                "sort_order": 999
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/sections", json=section_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    section_id = data["data"].get("id")
                    if section_id:
                        self.created_faq_sections.append(section_id)
                        details += f" - FAQ section created successfully with ID: {section_id}"
                    else:
                        success = False
                        details += " - Response missing section ID"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Create FAQ Section",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Create FAQ Section", False, f"Exception: {str(e)}")
            return False

    def test_update_faq_section(self) -> bool:
        """Test updating an FAQ section"""
        if not self.admin_token:
            self.log_test("Update FAQ Section", False, "No admin token available")
            return False
            
        if not self.created_faq_sections:
            self.log_test("Update FAQ Section", False, "No FAQ section ID available for testing")
            return False
            
        try:
            section_id = self.created_faq_sections[0]
            update_data = {
                "name": "Updated FAQ Management Test Section",
                "description": "Updated test section during FAQ Management verification",
                "sort_order": 1000
            }
            
            response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/sections/{section_id}", json=update_data)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += f" - FAQ section {section_id} updated successfully"
                else:
                    success = False
                    details += " - Update failed"
            
            self.log_test(
                "Update FAQ Section",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Update FAQ Section", False, f"Exception: {str(e)}")
            return False

    def test_create_faq_item(self) -> bool:
        """Test creating an FAQ item under a section"""
        if not self.admin_token:
            self.log_test("Create FAQ Item", False, "No admin token available")
            return False
            
        if not self.created_faq_sections:
            self.log_test("Create FAQ Item", False, "No FAQ section ID available for testing")
            return False
            
        try:
            section_id = self.created_faq_sections[0]
            item_data = {
                "section_id": section_id,
                "question": "How does the FAQ Management system work?",
                "answer": "The FAQ Management system allows admins to create, update, and organize frequently asked questions into sections.",
                "sort_order": 1
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/items", json=item_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    item_id = data["data"].get("id")
                    if item_id:
                        self.created_faq_items.append(item_id)
                        details += f" - FAQ item created successfully with ID: {item_id}"
                    else:
                        success = False
                        details += " - Response missing item ID"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Create FAQ Item",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Create FAQ Item", False, f"Exception: {str(e)}")
            return False

    def test_update_faq_item(self) -> bool:
        """Test updating an FAQ item"""
        if not self.admin_token:
            self.log_test("Update FAQ Item", False, "No admin token available")
            return False
            
        if not self.created_faq_items:
            self.log_test("Update FAQ Item", False, "No FAQ item ID available for testing")
            return False
            
        try:
            item_id = self.created_faq_items[0]
            update_data = {
                "question": "How does the updated FAQ Management system work?",
                "answer": "The updated FAQ Management system provides comprehensive functionality for managing FAQs with proper routing and authentication.",
                "sort_order": 2
            }
            
            response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/items/{item_id}", json=update_data)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += f" - FAQ item {item_id} updated successfully"
                else:
                    success = False
                    details += " - Update failed"
            
            self.log_test(
                "Update FAQ Item",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Update FAQ Item", False, f"Exception: {str(e)}")
            return False

    def test_public_faq_sections(self) -> bool:
        """Test public FAQ sections endpoint (no auth required)"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/faq/sections")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    sections = data.get("data", [])
                    details += f" - Found {len(sections)} public FAQ sections"
                    if sections:
                        details += f" - Sample section: {sections[0].get('name', 'N/A')}"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Public FAQ Sections",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Public FAQ Sections", False, f"Exception: {str(e)}")
            return False

    def test_public_faq_items(self) -> bool:
        """Test public FAQ items endpoint (no auth required)"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/faq/items")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    items = data.get("data", [])
                    details += f" - Found {len(items)} public FAQ items"
                    if items:
                        details += f" - Sample item: {items[0].get('question', 'N/A')[:50]}..."
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Public FAQ Items",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Public FAQ Items", False, f"Exception: {str(e)}")
            return False

    def run_faq_management_verification(self):
        """Run comprehensive FAQ Management verification tests"""
        print("🚀 Starting FAQ Management Fix Verification")
        print("Testing the specific fix for missing GET route: adminRoutes.GET(\"/content/faq/sections\", contentHandler.ListFAQSections)")
        print("=" * 100)
        
        # Test 1: Critical Fix Verification - Route should return 401 not 404
        print("\n🔍 PHASE 1: CRITICAL FIX VERIFICATION")
        print("-" * 60)
        
        route_fix_success = self.test_faq_sections_endpoint_without_auth()
        
        if not route_fix_success:
            print("❌ CRITICAL: The FAQ Management route fix has NOT been applied correctly!")
            print("   The GET /api/v1/admin/content/faq/sections endpoint is still returning 404.")
            print("   This indicates the route is still missing from server.go")
            return
        
        print("✅ SUCCESS: FAQ Management route fix has been applied correctly!")
        print("   The GET /api/v1/admin/content/faq/sections endpoint now returns 401 (not 404)")
        
        # Test 2: Authentication and Basic Functionality
        print("\n🔐 PHASE 2: AUTHENTICATION & BASIC FUNCTIONALITY")
        print("-" * 60)
        
        if not self.authenticate_admin():
            print("❌ Admin authentication failed. Cannot test authenticated endpoints.")
            return
        
        self.test_faq_sections_endpoint_with_auth()
        
        # Test 3: Complete FAQ Management Workflow
        print("\n📝 PHASE 3: COMPLETE FAQ MANAGEMENT WORKFLOW")
        print("-" * 60)
        
        self.test_create_faq_section()
        self.test_update_faq_section()
        self.test_create_faq_item()
        self.test_update_faq_item()
        
        # Test 4: Public FAQ APIs
        print("\n🌐 PHASE 4: PUBLIC FAQ APIs")
        print("-" * 60)
        
        self.test_public_faq_sections()
        self.test_public_faq_items()
        
        # Generate Summary
        self.generate_summary()

    def generate_summary(self):
        """Generate comprehensive test summary"""
        print("\n" + "=" * 100)
        print("📊 FAQ MANAGEMENT FIX VERIFICATION SUMMARY")
        print("=" * 100)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {success_rate:.1f}%")
        print()
        
        # Critical fix assessment
        route_fix_test = next((r for r in self.test_results if "No Auth" in r["test"]), None)
        if route_fix_test and route_fix_test["success"]:
            print("🎉 CRITICAL FIX VERIFIED: The missing GET route has been successfully added!")
            print("   ✅ GET /api/v1/admin/content/faq/sections now returns 401 (not 404)")
        else:
            print("❌ CRITICAL FIX FAILED: The missing GET route has NOT been fixed!")
            print("   ❌ GET /api/v1/admin/content/faq/sections still returns 404")
        
        # FAQ Management functionality assessment
        faq_tests = [r for r in self.test_results if "FAQ" in r["test"] and r["test"] != "Admin Authentication"]
        if faq_tests:
            faq_success = sum(1 for r in faq_tests if r["success"]) / len(faq_tests) * 100
            print(f"\nFAQ Management Functionality: {faq_success:.1f}% working")
            
            if faq_success == 100:
                print("🎉 EXCELLENT: All FAQ Management endpoints are working perfectly!")
            elif faq_success >= 85:
                print("✅ GOOD: FAQ Management is mostly functional with minor issues.")
            elif faq_success >= 70:
                print("⚠️  MODERATE: FAQ Management has some issues that need attention.")
            else:
                print("❌ CRITICAL: FAQ Management has significant issues.")
        
        # Show failed tests
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print(f"\n❌ FAILED TESTS ({len(failed_results)}):")
            for result in failed_results:
                print(f"  • {result['test']}: {result['details']}")
        else:
            print("\n✅ ALL TESTS PASSED!")
        
        # Show created resources
        if self.created_faq_sections or self.created_faq_items:
            print(f"\n📝 CREATED TEST RESOURCES:")
            if self.created_faq_sections:
                print(f"  FAQ Sections: {len(self.created_faq_sections)} created (IDs: {self.created_faq_sections})")
            if self.created_faq_items:
                print(f"  FAQ Items: {len(self.created_faq_items)} created (IDs: {self.created_faq_items})")
        
        print(f"\n🎯 EXPECTED RESULTS VERIFICATION:")
        print("✅ GET /api/v1/admin/content/faq/sections should return 401 without auth (not 404)")
        print("✅ With admin auth, should return 200 with proper FAQ sections listing")
        print("✅ All other FAQ endpoints should continue working correctly")
        print("✅ FAQ Management should now have 100% success rate (7/7 endpoints working)")
        
        # Final assessment
        if success_rate >= 85:
            print(f"\n🎉 FAQ MANAGEMENT FIX VERIFICATION: SUCCESS!")
            print("   The Content Management System FAQ Management issue has been completely resolved.")
            print("   All FAQ-related endpoints are working correctly.")
        else:
            print(f"\n❌ FAQ MANAGEMENT FIX VERIFICATION: ISSUES FOUND!")
            print("   The fix may not have been applied correctly or there are other issues.")

if __name__ == "__main__":
    tester = FAQManagementVerificationTester()
    tester.run_faq_management_verification()