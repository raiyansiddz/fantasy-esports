#!/usr/bin/env python3
"""
Test specific admin routes to find where route registration breaks
"""

import requests

def test_specific_routes():
    base_url = "http://localhost:8001/api/v1"
    
    # Get admin token
    login_response = requests.post(f"{base_url}/admin/login", json={
        "username": "admin",
        "password": "admin123"
    })
    
    token = login_response.json().get('access_token')
    headers = {"Authorization": f"Bearer {token}"}
    
    # Test routes in order they appear in server.go
    test_routes = [
        "/admin/users",                    # Line 180 - WORKING
        "/admin/kyc/documents",            # Line 185 - NOT WORKING
        "/admin/matches/live-scoring",     # Line 189
        "/admin/transactions",             # Line 217
        "/admin/config",                   # Line 221 - WORKING
        "/admin/analytics/dashboard",      # Line 225 - NOT WORKING
    ]
    
    print("Testing routes in registration order:")
    print("=" * 60)
    
    for route in test_routes:
        try:
            response = requests.get(f"{base_url}{route}", headers=headers, timeout=5)
            status = "✅ WORKING" if response.status_code in [200, 201] else f"❌ {response.status_code}"
            print(f"{status} | {route}")
        except Exception as e:
            print(f"❌ ERROR | {route}: {str(e)}")

if __name__ == "__main__":
    test_specific_routes()