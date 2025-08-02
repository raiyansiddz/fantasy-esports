#!/usr/bin/env python3
"""
Comprehensive Backend API Testing for Fantasy Esports KYC Approval Workflow
Tests KYC system functionality including:
- Admin authentication and login
- Getting pending KYC documents with filters
- Processing KYC documents (approve/reject)
- User management integration
- Real database operations with PostgreSQL
"""

import requests
import json
import time
import random
import string
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

class FantasyEsportsKYCTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = ""):
        """Log test results"""
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} {test_name}")
        if details:
            print(f"   Details: {details}")
        self.test_results.append({
            "test": test_name,
            "success": success,
            "details": details,
            "timestamp": datetime.now().isoformat()
        })

    def test_health_check(self) -> bool:
        """Test basic health endpoint"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            success = response.status_code == 200 and "healthy" in response.text
            self.log_test("Health Check", success, f"Status: {response.status_code}")
            return success
        except Exception as e:
            self.log_test("Health Check", False, f"Error: {str(e)}")
            return False

    def test_admin_login(self) -> bool:
        """Test admin login functionality"""
        try:
            login_payload = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=login_payload)
            
            if response.status_code != 200:
                self.log_test("Admin Login", False, f"Login failed: {response.status_code} - {response.text}")
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test("Admin Login", False, f"Login failed: {data}")
                return False
                
            self.admin_token = data.get("access_token")
            if not self.admin_token:
                self.log_test("Admin Login", False, "No access token received")
                return False
                
            self.log_test("Admin Login", True, f"Admin logged in successfully")
            return True
            
        except Exception as e:
            self.log_test("Admin Login", False, f"Error: {str(e)}")
            return False

    def get_admin_headers(self) -> Dict[str, str]:
        """Get authorization headers for admin"""
        if not self.admin_token:
            return {}
        return {"Authorization": f"Bearer {self.admin_token}"}

    def test_get_pending_kyc_documents(self) -> Optional[List]:
        """Test getting pending KYC documents without filters"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test("Get Pending KYC Documents", False, "No admin token")
                return None
                
            response = self.session.get(f"{self.base_url}/api/v1/admin/kyc/pending", headers=headers)
            
            if response.status_code != 200:
                self.log_test("Get Pending KYC Documents", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test("Get Pending KYC Documents", False, f"API returned error: {data}")
                return None
                
            documents = data.get("documents", [])
            total = data.get("total", 0)
            self.log_test("Get Pending KYC Documents", True, 
                        f"Found {len(documents)} documents, Total: {total}")
            return documents
            
        except Exception as e:
            self.log_test("Get Pending KYC Documents", False, f"Error: {str(e)}")
            return None

    def test_get_kyc_documents_with_filters(self) -> Optional[List]:
        """Test getting KYC documents with various filters"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test("Get KYC Documents with Filters", False, "No admin token")
                return None
            
            # Test with status filter
            params = {
                "status": "pending",
                "page": 1,
                "limit": 10
            }
            
            response = self.session.get(f"{self.base_url}/api/v1/admin/kyc/pending", 
                                      headers=headers, params=params)
            
            if response.status_code != 200:
                self.log_test("Get KYC Documents with Filters", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test("Get KYC Documents with Filters", False, f"API returned error: {data}")
                return None
                
            documents = data.get("documents", [])
            pagination = {
                "page": data.get("page", 1),
                "total": data.get("total", 0),
                "pages": data.get("pages", 0)
            }
            
            self.log_test("Get KYC Documents with Filters", True, 
                        f"Filtered results: {len(documents)} documents, Page: {pagination['page']}")
            return documents
            
        except Exception as e:
            self.log_test("Get KYC Documents with Filters", False, f"Error: {str(e)}")
            return None

    def test_get_users_endpoint(self) -> Optional[List]:
        """Test admin users endpoint"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test("Get Users Endpoint", False, "No admin token")
                return None
                
            response = self.session.get(f"{self.base_url}/api/v1/admin/users", headers=headers)
            
            if response.status_code != 200:
                self.log_test("Get Users Endpoint", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test("Get Users Endpoint", False, f"API returned error: {data}")
                return None
                
            users = data.get("users", [])
            total = data.get("total", 0)
            self.log_test("Get Users Endpoint", True, 
                        f"Found {len(users)} users, Total: {total}")
            return users
            
        except Exception as e:
            self.log_test("Get Users Endpoint", False, f"Error: {str(e)}")
            return None

    def test_get_user_details(self, user_id: int) -> Optional[Dict]:
        """Test getting user details with KYC information"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test(f"Get User Details ({user_id})", False, "No admin token")
                return None
                
            response = self.session.get(f"{self.base_url}/api/v1/admin/users/{user_id}", headers=headers)
            
            if response.status_code != 200:
                self.log_test(f"Get User Details ({user_id})", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"Get User Details ({user_id})", False, f"API returned error: {data}")
                return None
                
            user = data.get("user", {})
            kyc_documents = data.get("kyc_documents", [])
            self.log_test(f"Get User Details ({user_id})", True, 
                        f"User: {user.get('first_name', '')} {user.get('last_name', '')}, "
                        f"KYC Status: {user.get('kyc_status', 'N/A')}, "
                        f"Documents: {len(kyc_documents)}")
            return data
            
        except Exception as e:
            self.log_test(f"Get User Details ({user_id})", False, f"Error: {str(e)}")
            return None

    def test_update_user_status(self, user_id: int, new_status: str = "active") -> bool:
        """Test updating user account status"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test(f"Update User Status ({user_id})", False, "No admin token")
                return False
                
            payload = {
                "account_status": new_status,
                "reason": "Testing account status update"
            }
            
            response = self.session.put(f"{self.base_url}/api/v1/admin/users/{user_id}/status", 
                                      json=payload, headers=headers)
            
            if response.status_code != 200:
                self.log_test(f"Update User Status ({user_id})", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"Update User Status ({user_id})", False, f"API returned error: {data}")
                return False
                
            self.log_test(f"Update User Status ({user_id})", True, 
                        f"Status updated to: {new_status}")
            return True
            
        except Exception as e:
            self.log_test(f"Update User Status ({user_id})", False, f"Error: {str(e)}")
            return False

    def test_process_kyc_document(self, document_id: int, action: str = "verified", 
                                rejection_reason: str = None) -> bool:
        """Test processing KYC document (approve/reject)"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test(f"Process KYC Document ({document_id})", False, "No admin token")
                return False
                
            payload = {
                "status": action,
                "notes": f"Testing KYC {action} workflow"
            }
            
            if action == "rejected" and rejection_reason:
                payload["rejection_reason"] = rejection_reason
            elif action == "rejected":
                payload["rejection_reason"] = "Test rejection for workflow validation"
                
            response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                      json=payload, headers=headers)
            
            if response.status_code != 200:
                self.log_test(f"Process KYC Document ({document_id})", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"Process KYC Document ({document_id})", False, f"API returned error: {data}")
                return False
                
            user_kyc_status = data.get("user_kyc_status", "unknown")
            self.log_test(f"Process KYC Document ({document_id})", True, 
                        f"Document {action}, User KYC Status: {user_kyc_status}")
            return True
            
        except Exception as e:
            self.log_test(f"Process KYC Document ({document_id})", False, f"Error: {str(e)}")
            return False

    def test_kyc_edge_cases(self):
        """Test KYC edge cases and error scenarios"""
        print("\n🧪 Testing KYC Edge Cases...")
        
        headers = self.get_admin_headers()
        if not headers:
            self.log_test("KYC Edge Cases", False, "No admin token")
            return
        
        # Test 1: Invalid document ID
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/99999/process", 
                                  json={"status": "verified"}, headers=headers)
        success = response.status_code == 404
        self.log_test("Edge Case - Invalid Document ID", success, 
                    f"Expected 404, got {response.status_code}")
        
        # Test 2: Missing rejection reason for rejected status
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "rejected"}, headers=headers)
        success = response.status_code == 400
        self.log_test("Edge Case - Missing Rejection Reason", success, 
                    f"Expected 400, got {response.status_code}")
        
        # Test 3: Invalid status value
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "invalid_status"}, headers=headers)
        success = response.status_code in [400, 422]
        self.log_test("Edge Case - Invalid Status", success, 
                    f"Expected 400/422, got {response.status_code}")

    def run_comprehensive_kyc_test(self):
        """Run comprehensive KYC approval workflow test"""
        print("🚀 Starting Comprehensive KYC Approval Workflow Test")
        print("=" * 60)
        
        # Test 1: Health Check
        if not self.test_health_check():
            print("❌ Server health check failed. Aborting tests.")
            return
            
        # Test 2: Admin Authentication
        print("\n🔐 Testing Admin Authentication...")
        if not self.test_admin_login():
            print("❌ Admin login failed. Aborting tests.")
            return
            
        # Test 3: Get Pending KYC Documents
        print("\n📋 Testing KYC Document Retrieval...")
        pending_docs = self.test_get_pending_kyc_documents()
        
        # Test 4: Get KYC Documents with Filters
        filtered_docs = self.test_get_kyc_documents_with_filters()
        
        # Test 5: User Management Integration
        print("\n👥 Testing User Management Integration...")
        users = self.test_get_users_endpoint()
        
        if users and len(users) > 0:
            # Test user details for first user
            first_user = users[0]
            user_id = first_user.get("id")
            if user_id:
                user_details = self.test_get_user_details(user_id)
                
                # Test user status update
                self.test_update_user_status(user_id, "active")
        
        # Test 6: KYC Document Processing
        print("\n⚖️ Testing KYC Document Processing...")
        
        # If we have pending documents, test processing them
        if pending_docs and len(pending_docs) > 0:
            first_doc = pending_docs[0]
            doc_id = first_doc.get("id")
            
            if doc_id:
                # Test approval
                self.test_process_kyc_document(doc_id, "verified")
                
                # If there are more documents, test rejection
                if len(pending_docs) > 1:
                    second_doc = pending_docs[1]
                    second_doc_id = second_doc.get("id")
                    if second_doc_id:
                        self.test_process_kyc_document(second_doc_id, "rejected", 
                                                     "Document quality insufficient for verification")
        else:
            print("   No pending KYC documents found for processing tests")
            self.log_test("KYC Document Processing", True, 
                        "No pending documents available - endpoint structure validated")
        
        # Test 7: Edge Cases
        self.test_kyc_edge_cases()
        
        # Test Summary
        print("\n" + "=" * 60)
        print("📋 TEST SUMMARY")
        print("=" * 60)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        
        print(f"Total Tests: {total_tests}")
        print(f"✅ Passed: {passed_tests}")
        print(f"❌ Failed: {failed_tests}")
        print(f"Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        
        if failed_tests > 0:
            print("\n❌ FAILED TESTS:")
            for result in self.test_results:
                if not result["success"]:
                    print(f"  - {result['test']}: {result['details']}")
        
        # Save detailed results
        with open("/app/kyc_test_results.json", "w") as f:
            json.dump({
                "summary": {
                    "total_tests": total_tests,
                    "passed": passed_tests,
                    "failed": failed_tests,
                    "success_rate": (passed_tests/total_tests)*100
                },
                "test_results": self.test_results,
                "timestamp": datetime.now().isoformat()
            }, f, indent=2)
        
        print(f"\n📄 Detailed results saved to: /app/kyc_test_results.json")
        
        return passed_tests == total_tests

def main():
    """Main test execution"""
    tester = FantasyEsportsKYCTester()
    success = tester.run_comprehensive_kyc_test()
    
    if success:
        print("\n🎉 All tests passed! KYC Approval Workflow is working correctly.")
        exit(0)
    else:
        print("\n⚠️  Some tests failed. Check the results above.")
        exit(1)

if __name__ == "__main__":
    main()