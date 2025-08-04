#!/usr/bin/env python3
"""
Focused Notification System Test - Debug Authentication Issues
"""

import requests
import json

def test_admin_endpoints():
    base_url = "http://localhost:8001"
    api_base = f"{base_url}/api/v1"
    
    # Get admin token
    login_response = requests.post(f"{api_base}/admin/login", json={
        "username": "admin",
        "password": "admin123"
    })
    
    if login_response.status_code != 200:
        print("❌ Admin login failed")
        return
        
    token = login_response.json()['access_token']
    headers = {"Authorization": f"Bearer {token}"}
    
    print(f"✅ Admin token obtained: {token[:50]}...")
    
    # Test various admin endpoints to see which ones work
    endpoints_to_test = [
        {"method": "GET", "path": "/admin/users", "name": "Get Users"},
        {"method": "GET", "path": "/admin/config", "name": "Get System Config"},
        {"method": "GET", "path": "/admin/templates", "name": "Get Templates"},
        {"method": "GET", "path": "/admin/config/notifications?provider=fast2sms&channel=sms", "name": "Get Notification Config"},
        {"method": "GET", "path": "/admin/stats/notifications", "name": "Get Notification Stats"},
        {"method": "POST", "path": "/admin/templates", "name": "Create Template", "data": {
            "name": "Test Template",
            "channel": "sms", 
            "provider": "fast2sms",
            "body": "Test message"
        }},
        {"method": "POST", "path": "/admin/notify/send", "name": "Send Notification", "data": {
            "channel": "sms",
            "recipient": "+919876543210",
            "body": "Test message"
        }}
    ]
    
    for endpoint in endpoints_to_test:
        try:
            if endpoint['method'] == 'GET':
                response = requests.get(f"{api_base}{endpoint['path']}", headers=headers, timeout=10)
            else:
                data = endpoint.get('data', {})
                response = requests.post(f"{api_base}{endpoint['path']}", json=data, headers=headers, timeout=10)
            
            print(f"{endpoint['name']}: {response.status_code}")
            if response.status_code not in [200, 201]:
                try:
                    error_data = response.json()
                    print(f"  Error: {error_data.get('error', 'Unknown error')}")
                except:
                    print(f"  Error: {response.text[:100]}")
            else:
                try:
                    data = response.json()
                    if 'templates' in data:
                        print(f"  Success: Found {len(data['templates'])} templates")
                    elif 'config' in data:
                        print(f"  Success: Found {len(data['config'])} config items")
                    elif 'stats' in data:
                        print(f"  Success: Stats retrieved")
                    elif 'id' in data:
                        print(f"  Success: Created with ID {data['id']}")
                    else:
                        print(f"  Success: {str(data)[:100]}")
                except:
                    print(f"  Success: Response received")
                    
        except Exception as e:
            print(f"{endpoint['name']}: ERROR - {str(e)}")

if __name__ == "__main__":
    test_admin_endpoints()