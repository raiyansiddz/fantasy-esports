#!/usr/bin/env python3
"""
Test KYC Fix - Verify the JSONB issue is resolved
"""

import requests
import json
import time

def test_kyc_fix():
    base_url = "http://localhost:8001"
    
    # Login as admin
    login_payload = {
        "username": "admin",
        "password": "admin123"
    }
    
    session = requests.Session()
    response = session.post(f"{base_url}/api/v1/admin/login", json=login_payload)
    
    if response.status_code != 200:
        print(f"❌ Login failed: {response.status_code}")
        return False
    
    data = response.json()
    admin_token = data.get("access_token")
    headers = {"Authorization": f"Bearer {admin_token}"}
    
    print("✅ Admin logged in successfully")
    
    # Test 1: Process document with notes (should now work)
    print("\n🔍 Test 1: Processing with notes (after fix)")
    payload = {
        "status": "verified",
        "notes": "Document verified successfully after fixing JSONB issue"
    }
    
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/2/process", 
                          json=payload, headers=headers)
    
    print(f"Status: {response.status_code}")
    print(f"Response: {response.text}")
    
    if response.status_code == 200:
        print("✅ SUCCESS: Document processed with notes!")
        data = response.json()
        print(f"   User KYC Status: {data.get('user_kyc_status')}")
        print(f"   Document Status: {data.get('new_status')}")
        return True
    else:
        print("❌ FAILED: Still getting error")
        return False

def test_comprehensive_kyc_scenarios():
    """Test all KYC scenarios mentioned in the review request"""
    base_url = "http://localhost:8001"
    
    # Login as admin
    login_payload = {
        "username": "admin",
        "password": "admin123"
    }
    
    session = requests.Session()
    response = session.post(f"{base_url}/api/v1/admin/login", json=login_payload)
    
    if response.status_code != 200:
        print(f"❌ Login failed: {response.status_code}")
        return
    
    data = response.json()
    admin_token = data.get("access_token")
    headers = {"Authorization": f"Bearer {admin_token}"}
    
    print("\n🧪 COMPREHENSIVE KYC TESTING")
    print("=" * 50)
    
    # Test 1: Valid document verification
    print("\n1. Valid Document Verification")
    payload = {
        "status": "verified",
        "notes": "Document meets all verification criteria"
    }
    
    # Find a pending document
    users_response = session.get(f"{base_url}/api/v1/admin/users", headers=headers)
    if users_response.status_code == 200:
        users = users_response.json().get("users", [])
        for user in users:
            user_id = user.get("id")
            user_response = session.get(f"{base_url}/api/v1/admin/users/{user_id}", headers=headers)
            if user_response.status_code == 200:
                user_data = user_response.json()
                kyc_docs = user_data.get("kyc_documents", [])
                pending_doc = next((doc for doc in kyc_docs if doc.get("status") == "pending"), None)
                if pending_doc:
                    doc_id = pending_doc.get("id")
                    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/{doc_id}/process", 
                                         json=payload, headers=headers)
                    print(f"   Status: {response.status_code}")
                    if response.status_code == 200:
                        print("   ✅ Document verified successfully")
                    else:
                        print(f"   ❌ Failed: {response.text}")
                    break
    
    # Test 2: Document rejection with reason
    print("\n2. Document Rejection with Reason")
    payload = {
        "status": "rejected",
        "rejection_reason": "Document image quality is too poor for verification",
        "notes": "Requested user to resubmit with better quality image"
    }
    
    # Find another document to reject
    for user in users:
        user_id = user.get("id")
        user_response = session.get(f"{base_url}/api/v1/admin/users/{user_id}", headers=headers)
        if user_response.status_code == 200:
            user_data = user_response.json()
            kyc_docs = user_data.get("kyc_documents", [])
            pending_doc = next((doc for doc in kyc_docs if doc.get("status") == "pending"), None)
            if pending_doc:
                doc_id = pending_doc.get("id")
                response = session.put(f"{base_url}/api/v1/admin/kyc/documents/{doc_id}/process", 
                                     json=payload, headers=headers)
                print(f"   Status: {response.status_code}")
                if response.status_code == 200:
                    print("   ✅ Document rejected successfully")
                else:
                    print(f"   ❌ Failed: {response.text}")
                break
    
    # Test 3: Invalid document ID
    print("\n3. Invalid Document ID")
    payload = {"status": "verified"}
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/99999/process", 
                          json=payload, headers=headers)
    print(f"   Status: {response.status_code}")
    if response.status_code == 404:
        print("   ✅ Correctly handled invalid document ID")
    else:
        print(f"   ❌ Unexpected response: {response.text}")
    
    # Test 4: Missing rejection reason
    print("\n4. Missing Rejection Reason")
    payload = {"status": "rejected"}
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/1/process", 
                          json=payload, headers=headers)
    print(f"   Status: {response.status_code}")
    if response.status_code == 400:
        print("   ✅ Correctly required rejection reason")
    else:
        print(f"   ❌ Unexpected response: {response.text}")
    
    # Test 5: Duplicate processing
    print("\n5. Duplicate Processing Prevention")
    payload = {"status": "verified"}
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/1/process", 
                          json=payload, headers=headers)
    print(f"   Status: {response.status_code}")
    if response.status_code == 400:
        data = response.json()
        if data.get("code") == "ALREADY_PROCESSED":
            print("   ✅ Correctly prevented duplicate processing")
        else:
            print(f"   ❌ Wrong error code: {data.get('code')}")
    else:
        print(f"   ❌ Unexpected response: {response.text}")

if __name__ == "__main__":
    # Test the fix first
    success = test_kyc_fix()
    
    if success:
        print("\n🎉 FIX CONFIRMED! Now running comprehensive tests...")
        test_comprehensive_kyc_scenarios()
    else:
        print("\n❌ Fix did not work. Need further investigation.")