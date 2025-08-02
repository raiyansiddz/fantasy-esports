#!/usr/bin/env python3
"""
Focused KYC Database Error Investigation
This script specifically targets the database update failure in ProcessKYC endpoint
"""

import requests
import json
import time

def test_kyc_processing():
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
    
    print("✅ Admin logged in successfully")
    
    # Get user with KYC documents
    response = session.get(f"{base_url}/api/v1/admin/users/405", headers=headers)
    if response.status_code == 200:
        user_data = response.json()
        kyc_docs = user_data.get("kyc_documents", [])
        print(f"✅ Found user 405 with {len(kyc_docs)} KYC documents")
        
        for doc in kyc_docs:
            print(f"   Document ID: {doc.get('id')}, Type: {doc.get('document_type')}, Status: {doc.get('status')}")
    
    # Test processing document ID 2
    print("\n🔍 Testing KYC Document Processing...")
    
    payload = {
        "status": "verified",
        "notes": "Database error investigation test"
    }
    
    print(f"Making request to: {base_url}/api/v1/admin/kyc/documents/2/process")
    print(f"Payload: {json.dumps(payload, indent=2)}")
    
    start_time = time.time()
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/2/process", 
                          json=payload, headers=headers)
    end_time = time.time()
    
    print(f"Response time: {end_time - start_time:.2f} seconds")
    print(f"Status Code: {response.status_code}")
    print(f"Response Headers: {dict(response.headers)}")
    print(f"Response Body: {response.text}")
    
    if response.status_code == 500:
        print("\n❌ DATABASE UPDATE FAILURE CONFIRMED!")
        print("This is the exact error mentioned in the review request.")
        
        # Try to get more details by checking the document exists
        print("\n🔍 Checking if document exists...")
        response = session.get(f"{base_url}/api/v1/admin/users/405", headers=headers)
        if response.status_code == 200:
            user_data = response.json()
            kyc_docs = user_data.get("kyc_documents", [])
            doc_2 = next((doc for doc in kyc_docs if doc.get("id") == 2), None)
            if doc_2:
                print(f"✅ Document 2 exists: {doc_2}")
            else:
                print("❌ Document 2 not found")
    
    # Try with a different document if available
    print("\n🔍 Testing with different document...")
    payload = {
        "status": "rejected",
        "rejection_reason": "Test rejection for database investigation",
        "notes": "Database error investigation test"
    }
    
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/1/process", 
                          json=payload, headers=headers)
    
    print(f"Document 1 processing - Status: {response.status_code}")
    print(f"Response: {response.text}")

if __name__ == "__main__":
    test_kyc_processing()