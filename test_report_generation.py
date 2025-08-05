#!/usr/bin/env python3
"""
Test report generation with correct request format
"""

import requests
import json
import time

def test_report_generation():
    base_url = "http://localhost:8001"
    
    # Authenticate as admin
    login_data = {
        "username": "admin",
        "password": "admin123"
    }
    
    session = requests.Session()
    response = session.post(f"{base_url}/api/v1/admin/login", json=login_data)
    
    if response.status_code != 200:
        print(f"❌ Admin authentication failed: {response.status_code}")
        return
    
    data = response.json()
    if not data.get("success") or "access_token" not in data:
        print(f"❌ Admin authentication failed: {data}")
        return
    
    admin_token = data["access_token"]
    session.headers.update({"Authorization": f"Bearer {admin_token}"})
    print("✅ Admin authenticated successfully")
    
    # Test report generation with correct format
    report_data = {
        "report_type": "user",
        "format": "json",
        "date_from": "2024-08-01T00:00:00Z",
        "date_to": "2024-08-31T23:59:59Z",
        "description": "Test user report for analytics dashboard testing",
        "filters": {
            "kyc_status": "verified"
        }
    }
    
    print(f"Testing report generation with data: {json.dumps(report_data, indent=2)}")
    
    response = session.post(
        f"{base_url}/api/v1/admin/reports/generate", 
        json=report_data
    )
    
    print(f"Response status: {response.status_code}")
    print(f"Response body: {response.text}")
    
    if response.status_code == 200:
        data = response.json()
        if data.get("success"):
            report_info = data.get("data", {})
            print(f"✅ Report generated successfully: ID {report_info.get('id')}")
            return True
        else:
            print(f"❌ Report generation failed: {data.get('message', 'Unknown error')}")
    else:
        print(f"❌ Report generation failed with status {response.status_code}")
    
    return False

if __name__ == "__main__":
    test_report_generation()