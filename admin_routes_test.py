#!/usr/bin/env python3
"""
Quick test to check which admin routes are working
"""

import requests
import json

def test_admin_routes():
    base_url = "http://localhost:8001/api/v1"
    
    # First get admin token
    login_response = requests.post(f"{base_url}/admin/login", json={
        "username": "admin",
        "password": "admin123"
    })
    
    if login_response.status_code != 200:
        print("❌ Admin login failed")
        return
    
    token = login_response.json().get('access_token')
    headers = {"Authorization": f"Bearer {token}"}
    
    # Test various admin routes
    test_routes = [
        {"method": "GET", "path": "/admin/users", "name": "Get Users"},
        {"method": "GET", "path": "/admin/kyc/documents", "name": "Get KYC Documents"},
        {"method": "GET", "path": "/admin/config", "name": "Get System Config"},
        {"method": "GET", "path": "/admin/analytics/dashboard", "name": "Analytics Dashboard"},
        {"method": "GET", "path": "/admin/bi/dashboard", "name": "BI Dashboard"},
        {"method": "POST", "path": "/admin/reports/generate", "name": "Generate Report"},
        {"method": "GET", "path": "/admin/reports", "name": "Get Reports"},
    ]
    
    print("Testing admin routes:")
    print("=" * 50)
    
    for route in test_routes:
        try:
            if route["method"] == "GET":
                response = requests.get(f"{base_url}{route['path']}", headers=headers, timeout=5)
            else:
                payload = {
                    "report_type": "financial",
                    "format": "json",
                    "date_from": "2024-01-01",
                    "date_to": "2024-12-31",
                    "description": "Test report"
                }
                response = requests.post(f"{base_url}{route['path']}", headers=headers, json=payload, timeout=5)
            
            status = "✅ WORKING" if response.status_code in [200, 201] else f"❌ {response.status_code}"
            print(f"{status} | {route['name']} ({route['method']} {route['path']})")
            
        except Exception as e:
            print(f"❌ ERROR | {route['name']}: {str(e)}")
    
    print("=" * 50)

if __name__ == "__main__":
    test_admin_routes()