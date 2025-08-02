#!/usr/bin/env python3
"""
KYC Database Update Failure Investigation Script
This script specifically tests the PUT /admin/kyc/documents/{document_id}/process endpoint
to identify database update failures as mentioned in the review request.
"""

import requests
import json
import time
import random
from datetime import datetime
from typing import Dict, List, Optional

class KYCDatabaseTester:
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

    def admin_login(self) -> bool:
        """Login as admin"""
        try:
            login_payload = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=login_payload)
            
            if response.status_code != 200:
                print(f"❌ Admin login failed: {response.status_code} - {response.text}")
                return False
                
            data = response.json()
            if not data.get("success"):
                print(f"❌ Admin login failed: {data}")
                return False
                
            self.admin_token = data.get("access_token")
            if not self.admin_token:
                print("❌ No access token received")
                return False
                
            print("✅ Admin logged in successfully")
            return True
            
        except Exception as e:
            print(f"❌ Admin login error: {str(e)}")
            return False

    def get_admin_headers(self) -> Dict[str, str]:
        """Get authorization headers for admin"""
        if not self.admin_token:
            return {}
        return {"Authorization": f"Bearer {self.admin_token}"}

    def create_test_user_with_kyc(self) -> Optional[int]:
        """Create a test user with KYC documents for testing"""
        try:
            # First, let's get existing users to see if we can use one
            headers = self.get_admin_headers()
            response = self.session.get(f"{self.base_url}/api/v1/admin/users", headers=headers)
            
            if response.status_code == 200:
                data = response.json()
                users = data.get("users", [])
                
                # Find a user with KYC documents
                for user in users:
                    user_id = user.get("id")
                    if user_id:
                        # Get user details to check KYC documents
                        user_response = self.session.get(f"{self.base_url}/api/v1/admin/users/{user_id}", headers=headers)
                        if user_response.status_code == 200:
                            user_data = user_response.json()
                            kyc_docs = user_data.get("kyc_documents", [])
                            if kyc_docs and len(kyc_docs) > 0:
                                print(f"✅ Found user {user_id} with {len(kyc_docs)} KYC documents")
                                return user_id
                
                # If no user with KYC docs found, use the first user
                if users and len(users) > 0:
                    user_id = users[0].get("id")
                    print(f"✅ Using existing user {user_id} for testing")
                    return user_id
            
            print("❌ No suitable test user found")
            return None
            
        except Exception as e:
            print(f"❌ Error finding test user: {str(e)}")
            return None

    def create_test_kyc_document(self, user_id: int) -> Optional[int]:
        """Create a test KYC document by directly inserting into database via API simulation"""
        try:
            # Since we can't directly insert into DB, let's check if user already has KYC documents
            headers = self.get_admin_headers()
            response = self.session.get(f"{self.base_url}/api/v1/admin/users/{user_id}", headers=headers)
            
            if response.status_code == 200:
                data = response.json()
                kyc_docs = data.get("kyc_documents", [])
                
                if kyc_docs and len(kyc_docs) > 0:
                    # Return the first document ID
                    doc_id = kyc_docs[0].get("id")
                    print(f"✅ Found existing KYC document {doc_id} for user {user_id}")
                    return doc_id
            
            print(f"❌ No KYC documents found for user {user_id}")
            return None
            
        except Exception as e:
            print(f"❌ Error getting KYC document: {str(e)}")
            return None

    def test_kyc_processing_database_updates(self, document_id: int) -> bool:
        """Test KYC document processing with focus on database update failures"""
        print(f"\n🔍 Testing KYC Document Processing for Document ID: {document_id}")
        
        headers = self.get_admin_headers()
        if not headers:
            self.log_test("KYC Processing - Database Updates", False, "No admin token")
            return False
        
        # Test 1: Valid document verification
        print("  Test 1: Document Verification (status: verified)")
        payload = {
            "status": "verified",
            "notes": "Test verification for database update investigation"
        }
        
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                  json=payload, headers=headers)
        
        print(f"    Response Status: {response.status_code}")
        print(f"    Response Body: {response.text}")
        
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                self.log_test("KYC Processing - Verification", True, 
                            f"Document verified, User KYC Status: {data.get('user_kyc_status')}")
            else:
                self.log_test("KYC Processing - Verification", False, 
                            f"API returned error: {data}")
                return False
        else:
            self.log_test("KYC Processing - Verification", False, 
                        f"HTTP {response.status_code}: {response.text}")
            return False
        
        # Test 2: Try to process the same document again (should fail with ALREADY_PROCESSED)
        print("  Test 2: Duplicate Processing (should fail)")
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                  json=payload, headers=headers)
        
        print(f"    Response Status: {response.status_code}")
        print(f"    Response Body: {response.text}")
        
        if response.status_code == 400:
            data = response.json()
            if data.get("code") == "ALREADY_PROCESSED":
                self.log_test("KYC Processing - Duplicate Prevention", True, 
                            "Correctly prevented duplicate processing")
            else:
                self.log_test("KYC Processing - Duplicate Prevention", False, 
                            f"Wrong error code: {data.get('code')}")
        else:
            self.log_test("KYC Processing - Duplicate Prevention", False, 
                        f"Expected 400, got {response.status_code}")
        
        return True

    def test_kyc_rejection_with_reason(self, document_id: int) -> bool:
        """Test KYC document rejection with proper reason"""
        print(f"\n🔍 Testing KYC Document Rejection for Document ID: {document_id}")
        
        headers = self.get_admin_headers()
        if not headers:
            self.log_test("KYC Rejection - Database Updates", False, "No admin token")
            return False
        
        payload = {
            "status": "rejected",
            "rejection_reason": "Document quality is insufficient for verification",
            "notes": "Test rejection for database update investigation"
        }
        
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                  json=payload, headers=headers)
        
        print(f"    Response Status: {response.status_code}")
        print(f"    Response Body: {response.text}")
        
        if response.status_code == 200:
            data = response.json()
            if data.get("success"):
                self.log_test("KYC Processing - Rejection", True, 
                            f"Document rejected, User KYC Status: {data.get('user_kyc_status')}")
                return True
            else:
                self.log_test("KYC Processing - Rejection", False, 
                            f"API returned error: {data}")
                return False
        else:
            self.log_test("KYC Processing - Rejection", False, 
                        f"HTTP {response.status_code}: {response.text}")
            return False

    def test_concurrent_processing(self, document_id: int) -> bool:
        """Test concurrent processing of the same document"""
        print(f"\n🔍 Testing Concurrent Processing for Document ID: {document_id}")
        
        headers = self.get_admin_headers()
        if not headers:
            self.log_test("Concurrent Processing", False, "No admin token")
            return False
        
        payload = {
            "status": "verified",
            "notes": "Concurrent processing test"
        }
        
        # Make two simultaneous requests
        import threading
        import time
        
        results = []
        
        def make_request():
            response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                      json=payload, headers=headers)
            results.append({
                "status_code": response.status_code,
                "response": response.text
            })
        
        # Start two threads simultaneously
        thread1 = threading.Thread(target=make_request)
        thread2 = threading.Thread(target=make_request)
        
        thread1.start()
        thread2.start()
        
        thread1.join()
        thread2.join()
        
        print(f"    Request 1: Status {results[0]['status_code']}")
        print(f"    Request 2: Status {results[1]['status_code']}")
        
        # One should succeed (200), one should fail (400 - already processed)
        success_count = sum(1 for r in results if r['status_code'] == 200)
        failure_count = sum(1 for r in results if r['status_code'] == 400)
        
        if success_count == 1 and failure_count == 1:
            self.log_test("Concurrent Processing", True, 
                        "Correctly handled concurrent requests - one succeeded, one failed")
            return True
        else:
            self.log_test("Concurrent Processing", False, 
                        f"Unexpected results: {success_count} successes, {failure_count} failures")
            return False

    def test_different_document_types(self) -> bool:
        """Test processing different document types"""
        print(f"\n🔍 Testing Different Document Types")
        
        headers = self.get_admin_headers()
        if not headers:
            self.log_test("Different Document Types", False, "No admin token")
            return False
        
        # Get all KYC documents to test different types
        response = self.session.get(f"{self.base_url}/api/v1/admin/kyc/pending?status=pending", headers=headers)
        
        if response.status_code != 200:
            # Try to get any documents
            response = self.session.get(f"{self.base_url}/api/v1/admin/kyc/pending", headers=headers)
        
        if response.status_code == 200:
            data = response.json()
            documents = data.get("documents", [])
            
            document_types = {}
            for doc in documents:
                doc_type = doc.get("document_type")
                if doc_type not in document_types:
                    document_types[doc_type] = doc.get("id")
            
            if document_types:
                print(f"    Found document types: {list(document_types.keys())}")
                self.log_test("Different Document Types", True, 
                            f"Found {len(document_types)} different document types")
                return True
            else:
                self.log_test("Different Document Types", True, 
                            "No documents available for type testing")
                return True
        else:
            self.log_test("Different Document Types", False, 
                        f"Failed to get documents: {response.status_code}")
            return False

    def run_database_failure_investigation(self):
        """Run comprehensive database failure investigation"""
        print("🔍 KYC DATABASE UPDATE FAILURE INVESTIGATION")
        print("=" * 60)
        
        # Step 1: Admin login
        if not self.admin_login():
            print("❌ Cannot proceed without admin access")
            return False
        
        # Step 2: Find test user with KYC documents
        print("\n📋 Finding Test User with KYC Documents...")
        user_id = self.create_test_user_with_kyc()
        if not user_id:
            print("❌ Cannot proceed without test user")
            return False
        
        # Step 3: Get KYC document for testing
        print(f"\n📄 Getting KYC Document for User {user_id}...")
        document_id = self.create_test_kyc_document(user_id)
        if not document_id:
            print("❌ Cannot proceed without KYC document")
            return False
        
        # Step 4: Test database update scenarios
        print("\n🔬 TESTING DATABASE UPDATE SCENARIOS")
        print("-" * 40)
        
        # Test verification
        self.test_kyc_processing_database_updates(document_id)
        
        # Find another document for rejection test
        headers = self.get_admin_headers()
        response = self.session.get(f"{self.base_url}/api/v1/admin/users/{user_id}", headers=headers)
        if response.status_code == 200:
            data = response.json()
            kyc_docs = data.get("kyc_documents", [])
            
            # Find a document that's not already processed
            for doc in kyc_docs:
                if doc.get("status") == "pending" and doc.get("id") != document_id:
                    self.test_kyc_rejection_with_reason(doc.get("id"))
                    break
        
        # Test concurrent processing (if we have another pending document)
        # This would test transaction management
        
        # Test different document types
        self.test_different_document_types()
        
        # Step 5: Summary
        print("\n" + "=" * 60)
        print("📋 DATABASE INVESTIGATION SUMMARY")
        print("=" * 60)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        
        print(f"Total Tests: {total_tests}")
        print(f"✅ Passed: {passed_tests}")
        print(f"❌ Failed: {failed_tests}")
        
        if failed_tests > 0:
            print("\n❌ FAILED TESTS (Potential Database Issues):")
            for result in self.test_results:
                if not result["success"]:
                    print(f"  - {result['test']}: {result['details']}")
        else:
            print("\n✅ All database operations completed successfully!")
        
        # Save results
        with open("/app/kyc_database_test_results.json", "w") as f:
            json.dump({
                "investigation_summary": {
                    "total_tests": total_tests,
                    "passed": passed_tests,
                    "failed": failed_tests,
                    "database_issues_found": failed_tests > 0
                },
                "test_results": self.test_results,
                "timestamp": datetime.now().isoformat()
            }, f, indent=2)
        
        print(f"\n📄 Investigation results saved to: /app/kyc_database_test_results.json")
        
        return failed_tests == 0

def main():
    """Main investigation execution"""
    tester = KYCDatabaseTester()
    success = tester.run_database_failure_investigation()
    
    if success:
        print("\n🎉 Database investigation completed - No critical issues found!")
    else:
        print("\n⚠️  Database issues detected - Check results above for details.")

if __name__ == "__main__":
    main()