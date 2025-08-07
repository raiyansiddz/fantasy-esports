#!/usr/bin/env python3
"""
Comprehensive FAQ Management Testing
Testing all 7 FAQ-related endpoints to verify 100% success rate
"""

import requests
import json
import time

class ComprehensiveFAQTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.test_results = []
        self.created_faq_sections = []
        self.created_faq_items = []
        
    def log_test(self, test_name: str, success: bool, details: str = ""):
        """Log test results"""
        result = {
            "test": test_name,
            "success": success,
            "details": details,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        self.test_results.append(result)
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status}: {test_name}")
        if details:
            print(f"   Details: {details}")
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

    def test_all_faq_endpoints(self):
        """Test all 7 FAQ-related endpoints"""
        
        # 1. POST /api/v1/admin/content/faq/sections - Create FAQ section
        try:
            section_data = {
                "name": "Comprehensive Test Section",
                "description": "Test section for comprehensive FAQ testing",
                "sort_order": 1
            }
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/sections", json=section_data)
            success = response.status_code == 201
            section_id = None
            if success:
                data = response.json()
                section_id = data.get("data", {}).get("id")
                if section_id:
                    self.created_faq_sections.append(section_id)
            self.log_test("1. POST /admin/content/faq/sections - Create FAQ section", success, 
                         f"Status: {response.status_code}, Section ID: {section_id}")
        except Exception as e:
            self.log_test("1. POST /admin/content/faq/sections - Create FAQ section", False, f"Exception: {str(e)}")

        # 2. GET /api/v1/admin/content/faq/sections - List FAQ sections (NEWLY FIXED)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/faq/sections")
            success = response.status_code == 200
            count = 0
            if success:
                data = response.json()
                sections = data.get("data", [])
                count = len(sections)
            self.log_test("2. GET /admin/content/faq/sections - List FAQ sections (NEWLY FIXED)", success, 
                         f"Status: {response.status_code}, Found {count} sections")
        except Exception as e:
            self.log_test("2. GET /admin/content/faq/sections - List FAQ sections (NEWLY FIXED)", False, f"Exception: {str(e)}")

        # 3. PUT /api/v1/admin/content/faq/sections/{id} - Update FAQ section
        if self.created_faq_sections:
            try:
                section_id = self.created_faq_sections[0]
                update_data = {
                    "name": "Updated Comprehensive Test Section",
                    "description": "Updated test section for comprehensive FAQ testing",
                    "sort_order": 2
                }
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/sections/{section_id}", json=update_data)
                success = response.status_code == 200
                self.log_test("3. PUT /admin/content/faq/sections/{id} - Update FAQ section", success, 
                             f"Status: {response.status_code}, Section ID: {section_id}")
            except Exception as e:
                self.log_test("3. PUT /admin/content/faq/sections/{id} - Update FAQ section", False, f"Exception: {str(e)}")
        else:
            self.log_test("3. PUT /admin/content/faq/sections/{id} - Update FAQ section", False, "No section ID available")

        # 4. POST /api/v1/admin/content/faq/items - Create FAQ item
        if self.created_faq_sections:
            try:
                section_id = self.created_faq_sections[0]
                item_data = {
                    "section_id": section_id,
                    "question": "What is comprehensive FAQ testing?",
                    "answer": "Comprehensive FAQ testing verifies all FAQ management endpoints work correctly.",
                    "sort_order": 1
                }
                response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/items", json=item_data)
                success = response.status_code == 201
                item_id = None
                if success:
                    data = response.json()
                    item_id = data.get("data", {}).get("id")
                    if item_id:
                        self.created_faq_items.append(item_id)
                self.log_test("4. POST /admin/content/faq/items - Create FAQ item", success, 
                             f"Status: {response.status_code}, Item ID: {item_id}")
            except Exception as e:
                self.log_test("4. POST /admin/content/faq/items - Create FAQ item", False, f"Exception: {str(e)}")
        else:
            self.log_test("4. POST /admin/content/faq/items - Create FAQ item", False, "No section ID available")

        # 5. PUT /api/v1/admin/content/faq/items/{id} - Update FAQ item
        if self.created_faq_items:
            try:
                item_id = self.created_faq_items[0]
                # Use correct field structure for update
                update_data = {
                    "question": "What is updated comprehensive FAQ testing?",
                    "answer": "Updated comprehensive FAQ testing verifies all FAQ management endpoints work correctly with proper data validation.",
                    "sort_order": 2
                }
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/items/{item_id}", json=update_data)
                success = response.status_code == 200
                self.log_test("5. PUT /admin/content/faq/items/{id} - Update FAQ item", success, 
                             f"Status: {response.status_code}, Item ID: {item_id}")
            except Exception as e:
                self.log_test("5. PUT /admin/content/faq/items/{id} - Update FAQ item", False, f"Exception: {str(e)}")
        else:
            self.log_test("5. PUT /admin/content/faq/items/{id} - Update FAQ item", False, "No item ID available")

        # 6. GET /api/v1/faq/sections - List FAQ sections (public)
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
            count = 0
            if success:
                data = response.json()
                sections = data.get("data", [])
                count = len(sections)
            self.log_test("6. GET /faq/sections - List FAQ sections (public)", success, 
                         f"Status: {response.status_code}, Found {count} public sections")
        except Exception as e:
            self.log_test("6. GET /faq/sections - List FAQ sections (public)", False, f"Exception: {str(e)}")

        # 7. GET /api/v1/faq/items - List FAQ items (public)
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
            count = 0
            if success:
                data = response.json()
                items = data.get("data", [])
                count = len(items)
            self.log_test("7. GET /faq/items - List FAQ items (public)", success, 
                         f"Status: {response.status_code}, Found {count} public items")
        except Exception as e:
            self.log_test("7. GET /faq/items - List FAQ items (public)", False, f"Exception: {str(e)}")

    def run_comprehensive_test(self):
        """Run comprehensive FAQ testing"""
        print("🚀 Starting Comprehensive FAQ Management Testing")
        print("Testing all 7 FAQ-related endpoints for 100% success rate verification")
        print("=" * 80)
        
        if not self.authenticate_admin():
            print("❌ Admin authentication failed. Cannot test admin endpoints.")
            return
        
        self.test_all_faq_endpoints()
        
        # Generate Summary
        print("\n" + "=" * 80)
        print("📊 COMPREHENSIVE FAQ MANAGEMENT TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results) - 1  # Exclude authentication test
        passed_tests = sum(1 for result in self.test_results[1:] if result["success"])  # Exclude authentication test
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"FAQ Endpoints Tested: {total_tests}")
        print(f"Working: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {success_rate:.1f}%")
        print()
        
        # Show results for each endpoint
        print("📋 ENDPOINT-BY-ENDPOINT RESULTS:")
        for i, result in enumerate(self.test_results[1:], 1):  # Skip authentication
            status = "✅" if result["success"] else "❌"
            print(f"  {status} {result['test']}")
        
        # Show failed tests
        failed_results = [r for r in self.test_results[1:] if not r["success"]]
        if failed_results:
            print(f"\n❌ FAILED ENDPOINTS ({len(failed_results)}):")
            for result in failed_results:
                print(f"  • {result['test']}: {result['details']}")
        else:
            print("\n✅ ALL FAQ ENDPOINTS WORKING!")
        
        # Final assessment
        if success_rate == 100:
            print(f"\n🎉 PERFECT SUCCESS: FAQ Management has 100% success rate ({passed_tests}/{total_tests} endpoints working)!")
            print("   The Content Management System FAQ Management is fully functional.")
        elif success_rate >= 85:
            print(f"\n✅ EXCELLENT: FAQ Management has {success_rate:.1f}% success rate ({passed_tests}/{total_tests} endpoints working)!")
            print("   The Content Management System FAQ Management is mostly functional with minor issues.")
        else:
            print(f"\n⚠️ NEEDS ATTENTION: FAQ Management has {success_rate:.1f}% success rate ({passed_tests}/{total_tests} endpoints working).")
            print("   Some FAQ endpoints still need fixes.")
        
        # Show created resources
        if self.created_faq_sections or self.created_faq_items:
            print(f"\n📝 CREATED TEST RESOURCES:")
            if self.created_faq_sections:
                print(f"  FAQ Sections: {len(self.created_faq_sections)} created (IDs: {self.created_faq_sections})")
            if self.created_faq_items:
                print(f"  FAQ Items: {len(self.created_faq_items)} created (IDs: {self.created_faq_items})")

if __name__ == "__main__":
    tester = ComprehensiveFAQTester()
    tester.run_comprehensive_test()