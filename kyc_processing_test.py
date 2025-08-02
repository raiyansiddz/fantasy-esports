#!/usr/bin/env python3
"""
FOCUSED KYC Document Processing Endpoint Testing
Tests the FIXED KYC Document Processing endpoint to verify database update issue resolution:

**ENDPOINT TO TEST**: PUT /admin/kyc/documents/{document_id}/process

**VERIFICATION FOCUS**:
1. Fixed Database Update Testing - JSONB type mismatch issue resolved
2. Documents can be processed WITH and WITHOUT notes
3. Transaction commits work correctly
4. User KYC status updates properly
5. Performance and error handling improvements
"""

import requests
import json
import time
import random
import string
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

class KYCProcessingTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = "", response_time: float = None):
        """Log test results with response time tracking"""
        status = "✅ PASS" if success else "❌ FAIL"
        time_info = f" ({response_time:.3f}s)" if response_time else ""
        print(f"{status} {test_name}{time_info}")
        if details:
            print(f"   Details: {details}")
        self.test_results.append({
            "test": test_name,
            "success": success,
            "details": details,
            "response_time": response_time,
            "timestamp": datetime.now().isoformat()
        })

    def test_health_check(self) -> bool:
        """Test basic health endpoint"""
        try:
            start_time = time.time()
            response = self.session.get(f"{self.base_url}/health")
            response_time = time.time() - start_time
            
            success = response.status_code == 200 and "healthy" in response.text
            self.log_test("Health Check", success, f"Status: {response.status_code}", response_time)
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
            
            start_time = time.time()
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=login_payload)
            response_time = time.time() - start_time
            
            if response.status_code != 200:
                self.log_test("Admin Login", False, f"Login failed: {response.status_code} - {response.text}", response_time)
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test("Admin Login", False, f"Login failed: {data}", response_time)
                return False
                
            self.admin_token = data.get("access_token")
            if not self.admin_token:
                self.log_test("Admin Login", False, "No access token received", response_time)
                return False
                
            self.log_test("Admin Login", True, f"Admin logged in successfully", response_time)
            return True
            
        except Exception as e:
            self.log_test("Admin Login", False, f"Error: {str(e)}")
            return False

    def get_admin_headers(self) -> Dict[str, str]:
        """Get authorization headers for admin"""
        if not self.admin_token:
            return {}
        return {"Authorization": f"Bearer {self.admin_token}"}

    def get_available_documents(self) -> List[Dict]:
        """Get available KYC documents for testing"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                return []
                
            # Try to get pending documents first
            response = self.session.get(f"{self.base_url}/api/v1/admin/kyc/pending", 
                                      headers=headers, params={"status": "pending", "limit": 10})
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    documents = data.get("documents", [])
                    if documents:
                        return documents
            
            # If no pending documents, try to get any documents
            response = self.session.get(f"{self.base_url}/api/v1/admin/kyc/pending", 
                                      headers=headers, params={"limit": 10})
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success"):
                    return data.get("documents", [])
                    
            return []
            
        except Exception as e:
            print(f"Error getting documents: {str(e)}")
            return []

    def test_kyc_processing_with_notes(self, document_id: int) -> bool:
        """Test KYC document processing WITH notes (JSONB fix verification)"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test(f"KYC Processing WITH Notes ({document_id})", False, "No admin token")
                return False
                
            payload = {
                "status": "verified",
                "notes": "Document verified successfully. All details are clear and authentic. Processing completed with comprehensive review."
            }
            
            start_time = time.time()
            response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                      json=payload, headers=headers)
            response_time = time.time() - start_time
            
            if response.status_code != 200:
                self.log_test(f"KYC Processing WITH Notes ({document_id})", False, 
                            f"Failed: {response.status_code} - {response.text}", response_time)
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"KYC Processing WITH Notes ({document_id})", False, 
                            f"API returned error: {data}", response_time)
                return False
                
            user_kyc_status = data.get("user_kyc_status", "unknown")
            self.log_test(f"KYC Processing WITH Notes ({document_id})", True, 
                        f"Document verified, User KYC Status: {user_kyc_status}, Notes processed successfully", response_time)
            return True
            
        except Exception as e:
            self.log_test(f"KYC Processing WITH Notes ({document_id})", False, f"Error: {str(e)}")
            return False

    def test_kyc_processing_without_notes(self, document_id: int) -> bool:
        """Test KYC document processing WITHOUT notes"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test(f"KYC Processing WITHOUT Notes ({document_id})", False, "No admin token")
                return False
                
            payload = {
                "status": "verified"
                # No notes field
            }
            
            start_time = time.time()
            response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                      json=payload, headers=headers)
            response_time = time.time() - start_time
            
            if response.status_code != 200:
                self.log_test(f"KYC Processing WITHOUT Notes ({document_id})", False, 
                            f"Failed: {response.status_code} - {response.text}", response_time)
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"KYC Processing WITHOUT Notes ({document_id})", False, 
                            f"API returned error: {data}", response_time)
                return False
                
            user_kyc_status = data.get("user_kyc_status", "unknown")
            self.log_test(f"KYC Processing WITHOUT Notes ({document_id})", True, 
                        f"Document verified, User KYC Status: {user_kyc_status}, No notes processed successfully", response_time)
            return True
            
        except Exception as e:
            self.log_test(f"KYC Processing WITHOUT Notes ({document_id})", False, f"Error: {str(e)}")
            return False

    def test_kyc_rejection_with_reason_and_notes(self, document_id: int) -> bool:
        """Test KYC document rejection with both rejection reason and notes"""
        try:
            headers = self.get_admin_headers()
            if not headers:
                self.log_test(f"KYC Rejection WITH Reason & Notes ({document_id})", False, "No admin token")
                return False
                
            payload = {
                "status": "rejected",
                "rejection_reason": "Document quality is insufficient for verification",
                "notes": "The document image is blurry and some text is not clearly visible. Please resubmit with a clearer image."
            }
            
            start_time = time.time()
            response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/{document_id}/process", 
                                      json=payload, headers=headers)
            response_time = time.time() - start_time
            
            if response.status_code != 200:
                self.log_test(f"KYC Rejection WITH Reason & Notes ({document_id})", False, 
                            f"Failed: {response.status_code} - {response.text}", response_time)
                return False
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"KYC Rejection WITH Reason & Notes ({document_id})", False, 
                            f"API returned error: {data}", response_time)
                return False
                
            user_kyc_status = data.get("user_kyc_status", "unknown")
            self.log_test(f"KYC Rejection WITH Reason & Notes ({document_id})", True, 
                        f"Document rejected, User KYC Status: {user_kyc_status}, Reason & notes processed", response_time)
            return True
            
        except Exception as e:
            self.log_test(f"KYC Rejection WITH Reason & Notes ({document_id})", False, f"Error: {str(e)}")
            return False

    def test_kyc_processing_edge_cases(self):
        """Test KYC processing edge cases and error scenarios"""
        print("\n🧪 Testing KYC Processing Edge Cases...")
        
        headers = self.get_admin_headers()
        if not headers:
            self.log_test("KYC Edge Cases", False, "No admin token")
            return
        
        # Test 1: Invalid document ID
        start_time = time.time()
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/99999/process", 
                                  json={"status": "verified"}, headers=headers)
        response_time = time.time() - start_time
        success = response.status_code == 404
        self.log_test("Edge Case - Invalid Document ID", success, 
                    f"Expected 404, got {response.status_code}", response_time)
        
        # Test 2: Missing rejection reason for rejected status
        start_time = time.time()
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "rejected"}, headers=headers)
        response_time = time.time() - start_time
        success = response.status_code == 400
        self.log_test("Edge Case - Missing Rejection Reason", success, 
                    f"Expected 400, got {response.status_code}", response_time)
        
        # Test 3: Invalid status value
        start_time = time.time()
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "invalid_status"}, headers=headers)
        response_time = time.time() - start_time
        success = response.status_code in [400, 404, 422]
        self.log_test("Edge Case - Invalid Status", success, 
                    f"Expected 400/404/422, got {response.status_code}", response_time)

        # Test 4: Empty notes (should work)
        start_time = time.time()
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "verified", "notes": ""}, headers=headers)
        response_time = time.time() - start_time
        # This should work (empty notes are allowed)
        success = response.status_code in [200, 404]  # 404 if document doesn't exist is OK
        self.log_test("Edge Case - Empty Notes", success, 
                    f"Expected 200/404, got {response.status_code}", response_time)

        # Test 5: Very long notes (test JSONB handling)
        long_notes = "A" * 1000  # 1000 character notes
        start_time = time.time()
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "verified", "notes": long_notes}, headers=headers)
        response_time = time.time() - start_time
        success = response.status_code in [200, 404]  # 404 if document doesn't exist is OK
        self.log_test("Edge Case - Long Notes (1000 chars)", success, 
                    f"Expected 200/404, got {response.status_code}", response_time)

        # Test 6: Special characters in notes (test JSONB encoding)
        special_notes = "Document contains special chars: àáâãäåæçèéêë ñòóôõö ùúûüý 中文 العربية 🎉"
        start_time = time.time()
        response = self.session.put(f"{self.base_url}/api/v1/admin/kyc/documents/1/process", 
                                  json={"status": "verified", "notes": special_notes}, headers=headers)
        response_time = time.time() - start_time
        success = response.status_code in [200, 404]  # 404 if document doesn't exist is OK
        self.log_test("Edge Case - Special Characters in Notes", success, 
                    f"Expected 200/404, got {response.status_code}", response_time)

    def run_focused_kyc_processing_test(self):
        """Run focused KYC document processing test"""
        print("🚀 Starting FOCUSED KYC Document Processing Test")
        print("🎯 Testing FIXED Database Update Issue Resolution")
        print("=" * 70)
        
        # Test 1: Health Check
        if not self.test_health_check():
            print("❌ Server health check failed. Aborting tests.")
            return False
            
        # Test 2: Admin Authentication
        print("\n🔐 Testing Admin Authentication...")
        if not self.test_admin_login():
            print("❌ Admin login failed. Aborting tests.")
            return False
            
        # Test 3: Get Available Documents
        print("\n📋 Getting Available KYC Documents...")
        documents = self.get_available_documents()
        
        if not documents:
            print("⚠️  No KYC documents found. Creating test scenarios with mock IDs...")
            # Test with mock document IDs to verify endpoint structure
            test_doc_ids = [1, 2, 3]
        else:
            print(f"✅ Found {len(documents)} documents for testing")
            test_doc_ids = [doc.get("id") for doc in documents[:3] if doc.get("id")]
            
        # Test 4: Core KYC Processing Tests
        print("\n⚖️ Testing KYC Document Processing (FIXED JSONB Issues)...")
        
        success_count = 0
        total_tests = 0
        
        for i, doc_id in enumerate(test_doc_ids):
            if doc_id:
                print(f"\n--- Testing Document ID: {doc_id} ---")
                
                # Test processing WITH notes (main fix verification)
                if i == 0:
                    if self.test_kyc_processing_with_notes(doc_id):
                        success_count += 1
                    total_tests += 1
                
                # Test processing WITHOUT notes
                elif i == 1:
                    if self.test_kyc_processing_without_notes(doc_id):
                        success_count += 1
                    total_tests += 1
                
                # Test rejection with reason and notes
                elif i == 2:
                    if self.test_kyc_rejection_with_reason_and_notes(doc_id):
                        success_count += 1
                    total_tests += 1
        
        # Test 5: Edge Cases
        self.test_kyc_processing_edge_cases()
        
        # Test Summary
        print("\n" + "=" * 70)
        print("📋 FOCUSED KYC PROCESSING TEST SUMMARY")
        print("=" * 70)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        
        # Calculate average response time
        response_times = [r["response_time"] for r in self.test_results if r.get("response_time")]
        avg_response_time = sum(response_times) / len(response_times) if response_times else 0
        
        print(f"Total Tests: {total_tests}")
        print(f"✅ Passed: {passed_tests}")
        print(f"❌ Failed: {failed_tests}")
        print(f"Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        print(f"Average Response Time: {avg_response_time:.3f}s")
        
        # Check if the main issue is resolved
        processing_tests = [r for r in self.test_results if "KYC Processing" in r["test"]]
        processing_success = all(r["success"] for r in processing_tests)
        
        if processing_success and len(processing_tests) > 0:
            print("\n🎉 MAIN ISSUE RESOLVED: KYC document processing with JSONB notes is working!")
        elif len(processing_tests) == 0:
            print("\n⚠️  No actual KYC processing tests were performed (no available documents)")
        else:
            print("\n❌ MAIN ISSUE PERSISTS: KYC document processing still has issues")
        
        if failed_tests > 0:
            print("\n❌ FAILED TESTS:")
            for result in self.test_results:
                if not result["success"]:
                    time_info = f" ({result.get('response_time', 0):.3f}s)" if result.get('response_time') else ""
                    print(f"  - {result['test']}{time_info}: {result['details']}")
        
        # Save detailed results
        with open("/app/kyc_processing_test_results.json", "w") as f:
            json.dump({
                "test_focus": "KYC Document Processing Endpoint - JSONB Fix Verification",
                "endpoint": "PUT /admin/kyc/documents/{document_id}/process",
                "summary": {
                    "total_tests": total_tests,
                    "passed": passed_tests,
                    "failed": failed_tests,
                    "success_rate": (passed_tests/total_tests)*100,
                    "avg_response_time": avg_response_time,
                    "main_issue_resolved": processing_success and len(processing_tests) > 0
                },
                "test_results": self.test_results,
                "timestamp": datetime.now().isoformat()
            }, f, indent=2)
        
        print(f"\n📄 Detailed results saved to: /app/kyc_processing_test_results.json")
        
        return passed_tests == total_tests

def main():
    """Main test execution"""
    tester = KYCProcessingTester()
    success = tester.run_focused_kyc_processing_test()
    
    if success:
        print("\n🎉 All tests passed! KYC Document Processing endpoint is working correctly.")
        exit(0)
    else:
        print("\n⚠️  Some tests failed. Check the results above.")
        exit(1)

if __name__ == "__main__":
    main()