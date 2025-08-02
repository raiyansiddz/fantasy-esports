#!/usr/bin/env python3
"""
Simple KYC Test - Test the fix
"""

import requests
import json

def test_kyc_simple():
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
    
    # Test with notes
    print("\n🔍 Testing KYC processing with notes...")
    payload = {
        "status": "verified",
        "notes": "Test notes after fix"
    }
    
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/2/process", 
                          json=payload, headers=headers)
    
    print(f"Status: {response.status_code}")
    print(f"Response: {response.text}")
    
    if response.status_code == 200:
        print("✅ SUCCESS: KYC processing with notes works!")
    elif response.status_code == 400:
        data = response.json()
        if data.get("code") == "ALREADY_PROCESSED":
            print("✅ Document already processed (expected)")
        else:
            print(f"❌ Validation error: {data}")
    else:
        print("❌ Still failing")

if __name__ == "__main__":
    test_kyc_simple()