#!/usr/bin/env python3
"""
KYC Database Fix Test
This script tests the fix for the JSONB type mismatch issue
"""

import requests
import json

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
        return
    
    data = response.json()
    admin_token = data.get("access_token")
    headers = {"Authorization": f"Bearer {admin_token}"}
    
    print("✅ Admin logged in successfully")
    
    # Test 1: Process with notes (this should fail before fix)
    print("\n🔍 Test 1: Processing with notes (JSONB issue)")
    payload = {
        "status": "verified",
        "notes": "This is a string that should be converted to JSONB"
    }
    
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/2/process", 
                          json=payload, headers=headers)
    
    print(f"Status: {response.status_code}")
    print(f"Response: {response.text}")
    
    if response.status_code == 500:
        print("❌ CONFIRMED: JSONB type mismatch error")
        print("The issue is that 'notes' (string) is being inserted into 'additional_data' (JSONB) column")
        print("Solution: Convert string to proper JSONB format or use correct column mapping")
    
    # Test 2: Process without notes
    print("\n🔍 Test 2: Processing without notes")
    payload = {
        "status": "verified"
    }
    
    response = session.put(f"{base_url}/api/v1/admin/kyc/documents/1/process", 
                          json=payload, headers=headers)
    
    print(f"Status: {response.status_code}")
    print(f"Response: {response.text}")

if __name__ == "__main__":
    test_kyc_fix()