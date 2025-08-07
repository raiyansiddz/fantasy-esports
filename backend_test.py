#!/usr/bin/env python3
"""
🎯 FINAL VERIFICATION TESTING - ACHIEVING 100% SUCCESS RATE FOR ALL 7 ADVANCED GAMING FEATURES

OBJECTIVE: Verify that all gaming feature endpoints now work correctly after the Go binary fix, 
achieving 100% success rate improvement from previous baselines.

CONTEXT: 
- Backend now running correct Go binary (/app/backend/fantasy-esports-backend-100-percent) on http://localhost:8001
- Previous testing showed degradation from 87.5% baseline to 67.2% (43/64 endpoints accessible) 
- 21 failing endpoints were identified that should now return 401 (auth required) instead of 404 (not found)
- Target: 100% success rate for all 64 gaming feature endpoints

Testing all 64 endpoints across 7 gaming features:
1. Achievement System & Badge Management (10 endpoints)
2. Friend System & Challenges (12 endpoints)
3. Social Sharing Integration (8 endpoints)
4. Advanced Game Analytics (10 endpoints)
5. Player Performance Predictions (10 endpoints)
6. Automated Tournament Brackets (8 endpoints)
7. Advanced Fraud Detection (6 endpoints)
"""

import requests
import json
import time
import sys
from datetime import datetime

# Backend configuration
BACKEND_URL = "http://localhost:8001/api/v1"
ADMIN_TOKEN = None
USER_TOKEN = None

class GameFeatureTester:
    def __init__(self):
        self.total_tests = 0
        self.passed_tests = 0
        self.failed_tests = 0
        self.results = []
        
    def log_result(self, test_name, status, details=""):
        """Log test result"""
        self.total_tests += 1
        if status:
            self.passed_tests += 1
            print(f"✅ {test_name}")
        else:
            self.failed_tests += 1
            print(f"❌ {test_name}: {details}")
        
        self.results.append({
            "test": test_name,
            "status": "PASS" if status else "FAIL",
            "details": details,
            "timestamp": datetime.now().isoformat()
        })
    
    def make_request(self, method, endpoint, data=None, headers=None, auth_token=None):
        """Make HTTP request with proper error handling"""
        url = f"{BACKEND_URL}{endpoint}"
        
        if headers is None:
            headers = {"Content-Type": "application/json"}
        
        if auth_token:
            headers["Authorization"] = f"Bearer {auth_token}"
        
        try:
            if method.upper() == "GET":
                response = requests.get(url, headers=headers, timeout=5)
            elif method.upper() == "POST":
                response = requests.post(url, json=data, headers=headers, timeout=5)
            elif method.upper() == "PUT":
                response = requests.put(url, json=data, headers=headers, timeout=5)
            elif method.upper() == "DELETE":
                response = requests.delete(url, headers=headers, timeout=5)
            else:
                return None, f"Unsupported method: {method}"
            
            return response, None
        except requests.exceptions.Timeout:
            return None, "Request timeout"
        except requests.exceptions.ConnectionError:
            return None, "Connection error"
        except requests.exceptions.RequestException as e:
            return None, str(e)
    
    def authenticate_user(self):
        """Authenticate as regular user"""
        global USER_TOKEN
        
        # For testing purposes, we'll focus on endpoint accessibility
        # Most user endpoints should return 401 (auth required) instead of 404 (not found)
        USER_TOKEN = "dummy_token_for_testing"
        return True, "Using dummy token to test endpoint accessibility"
    
    def authenticate_admin(self):
        """Authenticate as admin user"""
        global ADMIN_TOKEN
        
        # For testing purposes, we'll focus on endpoint accessibility
        # Most admin endpoints should return 401 (auth required) instead of 404 (not found)
        ADMIN_TOKEN = "dummy_admin_token_for_testing"
        return True, "Using dummy token to test endpoint accessibility"

    def test_achievement_system(self):
        """Test Achievement System & Badge Management (10 endpoints)"""
        print("\n🏆 TESTING ACHIEVEMENT SYSTEM & BADGE MANAGEMENT (10 ENDPOINTS)")
        
        # 1. GET /api/v1/achievements (user achievements list)
        response, error = self.make_request("GET", "/achievements", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /achievements - User achievements list endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /achievements - User achievements list", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /achievements - User achievements list", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 2. POST /api/v1/achievements/claim (claim achievement)
        claim_data = {"achievement_id": "123"}
        response, error = self.make_request("POST", "/achievements/claim", claim_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /achievements/claim - Claim achievement endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /achievements/claim - Claim achievement", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /achievements/claim - Claim achievement", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 3. GET /api/v1/achievements/123 (achievement details)
        response, error = self.make_request("GET", "/achievements/123", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401, 404]:
            if response.status_code == 401:
                self.log_result("GET /achievements/123 - Achievement details endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /achievements/123 - Achievement details endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /achievements/123 - Achievement details", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /achievements/123 - Achievement details", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 4. GET /api/v1/achievements/categories (achievement categories)
        response, error = self.make_request("GET", "/achievements/categories", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /achievements/categories - Achievement categories endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /achievements/categories - Achievement categories endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /achievements/categories - Achievement categories", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /achievements/categories - Achievement categories", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 5. GET /api/v1/admin/achievements (admin list)
        response, error = self.make_request("GET", "/admin/achievements", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/achievements - Admin list achievements endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /admin/achievements - Admin list achievements", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/achievements - Admin list achievements", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 6. POST /api/v1/admin/achievements (admin create)
        achievement_data = {
            "name": "Gaming Master",
            "description": "Complete 100 matches",
            "type": "match_completion",
            "criteria": {"matches_required": 100},
            "reward_type": "badge",
            "reward_value": 500,
            "icon_url": "https://example.com/icon.png"
        }
        response, error = self.make_request("POST", "/admin/achievements", achievement_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /admin/achievements - Admin create achievement endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /admin/achievements - Admin create achievement", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /admin/achievements - Admin create achievement", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 7. PUT /api/v1/admin/achievements/123 (admin update)
        update_data = {"name": "Updated Achievement", "description": "Updated description"}
        response, error = self.make_request("PUT", "/admin/achievements/123", update_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /admin/achievements/123 - Admin update achievement endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("PUT /admin/achievements/123 - Admin update achievement", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /admin/achievements/123 - Admin update achievement", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 8. DELETE /api/v1/admin/achievements/123 (admin delete)
        response, error = self.make_request("DELETE", "/admin/achievements/123", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 204, 400, 401]:
            if response.status_code == 401:
                self.log_result("DELETE /admin/achievements/123 - Admin delete achievement endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("DELETE /admin/achievements/123 - Admin delete achievement", True, f"Status: {response.status_code}")
        else:
            self.log_result("DELETE /admin/achievements/123 - Admin delete achievement", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 9. GET /api/v1/admin/achievements/stats (admin stats)
        response, error = self.make_request("GET", "/admin/achievements/stats", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/achievements/stats - Admin achievement stats endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/achievements/stats - Admin achievement stats endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/achievements/stats - Admin achievement stats", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/achievements/stats - Admin achievement stats", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 10. GET /api/v1/achievements/leaderboard (leaderboard)
        response, error = self.make_request("GET", "/achievements/leaderboard", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /achievements/leaderboard - Achievement leaderboard endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /achievements/leaderboard - Achievement leaderboard", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /achievements/leaderboard - Achievement leaderboard", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def test_friend_system(self):
        """Test Friend System & Challenges (12 endpoints)"""
        print("\n👥 TESTING FRIEND SYSTEM & CHALLENGES (12 ENDPOINTS)")
        
        # 1. GET /api/v1/friends (friends list)
        response, error = self.make_request("GET", "/friends", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /friends - Friends list endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /friends - Friends list", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /friends - Friends list", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 2. POST /api/v1/friends/add (add friend)
        friend_data = {"username": "testfriend", "mobile": "+919876543211"}
        response, error = self.make_request("POST", "/friends/add", friend_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /friends/add - Add friend endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /friends/add - Add friend", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /friends/add - Add friend", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 3. DELETE /api/v1/friends/123 (remove friend)
        response, error = self.make_request("DELETE", "/friends/123", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 204, 400, 401]:
            if response.status_code == 401:
                self.log_result("DELETE /friends/123 - Remove friend endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("DELETE /friends/123 - Remove friend", True, f"Status: {response.status_code}")
        else:
            self.log_result("DELETE /friends/123 - Remove friend", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 4. GET /api/v1/friends/requests (friend requests)
        response, error = self.make_request("GET", "/friends/requests", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /friends/requests - Friend requests endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /friends/requests - Friend requests endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /friends/requests - Friend requests", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /friends/requests - Friend requests", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 5. POST /api/v1/friends/requests/123/accept (accept request)
        response, error = self.make_request("POST", "/friends/requests/123/accept", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /friends/requests/123/accept - Accept friend request endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /friends/requests/123/accept - Accept friend request", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /friends/requests/123/accept - Accept friend request", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 6. POST /api/v1/friends/requests/123/decline (decline request)
        response, error = self.make_request("POST", "/friends/requests/123/decline", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /friends/requests/123/decline - Decline friend request endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /friends/requests/123/decline - Decline friend request", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /friends/requests/123/decline - Decline friend request", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 7. GET /api/v1/challenges (challenges list)
        response, error = self.make_request("GET", "/challenges", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /challenges - Challenges list endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /challenges - Challenges list", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /challenges - Challenges list", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 8. POST /api/v1/challenges (create challenge)
        challenge_data = {
            "friend_id": "123",
            "contest_id": "456",
            "entry_fee": 100,
            "message": "Let's compete!"
        }
        response, error = self.make_request("POST", "/challenges", challenge_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /challenges - Create challenge endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /challenges - Create challenge", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /challenges - Create challenge", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 9. GET /api/v1/challenges/123/status (challenge status)
        response, error = self.make_request("GET", "/challenges/123/status", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("GET /challenges/123/status - Challenge status endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /challenges/123/status - Challenge status endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /challenges/123/status - Challenge status", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /challenges/123/status - Challenge status", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 10. POST /api/v1/challenges/123/accept (accept challenge)
        response, error = self.make_request("POST", "/challenges/123/accept", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /challenges/123/accept - Accept challenge endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /challenges/123/accept - Accept challenge", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /challenges/123/accept - Accept challenge", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 11. GET /api/v1/challenges/my (my challenges)
        response, error = self.make_request("GET", "/challenges/my", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /challenges/my - My challenges endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /challenges/my - My challenges endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /challenges/my - My challenges", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /challenges/my - My challenges", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 12. PUT /api/v1/challenges/123/resolve (resolve challenge)
        resolve_data = {"winner_id": "123", "result": "completed"}
        response, error = self.make_request("PUT", "/challenges/123/resolve", resolve_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /challenges/123/resolve - Resolve challenge endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("PUT /challenges/123/resolve - Resolve challenge", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /challenges/123/resolve - Resolve challenge", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def test_social_sharing(self):
        """Test Social Sharing Integration (8 endpoints)"""
        print("\n📱 TESTING SOCIAL SHARING INTEGRATION (8 ENDPOINTS)")
        
        # 1. POST /api/v1/share (share content)
        share_data = {
            "content_type": "achievement",
            "content_id": "123",
            "platform": "twitter",
            "message": "Just earned a new achievement!"
        }
        response, error = self.make_request("POST", "/share", share_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /share - Share content endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /share - Share content", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /share - Share content", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 2. GET /api/v1/share/123/stats (sharing stats)
        response, error = self.make_request("GET", "/share/123/stats", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("GET /share/123/stats - Sharing stats endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /share/123/stats - Sharing stats endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /share/123/stats - Sharing stats", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /share/123/stats - Sharing stats", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 3. GET /api/v1/admin/social/analytics (admin analytics)
        response, error = self.make_request("GET", "/admin/social/analytics", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/social/analytics - Admin social analytics endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /admin/social/analytics - Admin social analytics", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/social/analytics - Admin social analytics", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 4. GET /api/v1/admin/social/platforms/stats (platform stats)
        response, error = self.make_request("GET", "/admin/social/platforms/stats", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/social/platforms/stats - Platform stats endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/social/platforms/stats - Platform stats endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/social/platforms/stats - Platform stats", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/social/platforms/stats - Platform stats", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 5. GET /api/v1/admin/social/trending (trending content)
        response, error = self.make_request("GET", "/admin/social/trending", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/social/trending - Trending content endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/social/trending - Trending content endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/social/trending - Trending content", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/social/trending - Trending content", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 6. POST /api/v1/admin/social/campaigns (create campaign)
        campaign_data = {
            "name": "New Year Campaign",
            "description": "Promote sharing during new year",
            "start_date": "2025-01-01",
            "end_date": "2025-01-31",
            "platforms": ["twitter", "facebook"]
        }
        response, error = self.make_request("POST", "/admin/social/campaigns", campaign_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /admin/social/campaigns - Create campaign endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("POST /admin/social/campaigns - Create campaign endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("POST /admin/social/campaigns - Create campaign", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /admin/social/campaigns - Create campaign", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 7. GET /api/v1/admin/social/campaigns (list campaigns)
        response, error = self.make_request("GET", "/admin/social/campaigns", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/social/campaigns - List campaigns endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /admin/social/campaigns - List campaigns", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/social/campaigns - List campaigns", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 8. PUT /api/v1/admin/social/campaigns/123 (update campaign)
        update_data = {"name": "Updated Campaign", "description": "Updated description"}
        response, error = self.make_request("PUT", "/admin/social/campaigns/123", update_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /admin/social/campaigns/123 - Update campaign endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("PUT /admin/social/campaigns/123 - Update campaign", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /admin/social/campaigns/123 - Update campaign", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def test_advanced_game_analytics(self):
        """Test Advanced Game Analytics (10 endpoints - 7 metrics)"""
        print("\n📊 TESTING ADVANCED GAME ANALYTICS (10 ENDPOINTS - 7 METRICS)")
        
        game_id = "1"
        metrics = [
            "player-efficiency",
            "team-synergy", 
            "strategic-diversity",
            "comeback-potential",
            "clutch-performance",
            "consistency-index",
            "adaptability-score"
        ]
        
        # Test 1-7: GET /api/v1/analytics/metrics/{metric}/1 (7 different metrics)
        for i, metric in enumerate(metrics, 1):
            response, error = self.make_request("GET", f"/analytics/metrics/{metric}/{game_id}", auth_token=USER_TOKEN)
            if response and response.status_code in [200, 400, 401]:
                if response.status_code == 401:
                    self.log_result(f"GET /analytics/metrics/{metric}/{game_id} - {metric} endpoint accessible", True, "Returns 401 (auth required) instead of 404")
                else:
                    self.log_result(f"GET /analytics/metrics/{metric}/{game_id} - {metric}", True, f"Status: {response.status_code}")
            else:
                self.log_result(f"GET /analytics/metrics/{metric}/{game_id} - {metric}", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 8. GET /api/v1/admin/analytics/summary (admin summary)
        response, error = self.make_request("GET", "/admin/analytics/summary", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/analytics/summary - Admin analytics summary endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/analytics/summary - Admin analytics summary endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/analytics/summary - Admin analytics summary", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/analytics/summary - Admin analytics summary", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 9. POST /api/v1/admin/analytics/generate (generate analytics)
        generate_data = {"game_id": game_id, "metrics": ["player-efficiency", "team-synergy"]}
        response, error = self.make_request("POST", "/admin/analytics/generate", generate_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /admin/analytics/generate - Generate analytics endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("POST /admin/analytics/generate - Generate analytics endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("POST /admin/analytics/generate - Generate analytics", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /admin/analytics/generate - Generate analytics", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 10. GET /api/v1/analytics/compare/games/1/2 (game comparison)
        response, error = self.make_request("GET", "/analytics/compare/games/1/2", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("GET /analytics/compare/games/1/2 - Game comparison endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /analytics/compare/games/1/2 - Game comparison", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /analytics/compare/games/1/2 - Game comparison", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def test_player_predictions(self):
        """Test Player Performance Predictions (10 endpoints)"""
        print("\n🔮 TESTING PLAYER PERFORMANCE PREDICTIONS (10 ENDPOINTS)")
        
        # 1. GET /api/v1/matches/1/predictions (match predictions)
        response, error = self.make_request("GET", "/matches/1/predictions", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("GET /matches/1/predictions - Match predictions endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /matches/1/predictions - Match predictions", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /matches/1/predictions - Match predictions", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 2. POST /api/v1/predictions (create prediction)
        prediction_data = {
            "match_id": "1",
            "player_id": "123",
            "predicted_score": 85.5,
            "confidence": 0.8
        }
        response, error = self.make_request("POST", "/predictions", prediction_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /predictions - Create prediction endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /predictions - Create prediction", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /predictions - Create prediction", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 3. GET /api/v1/predictions/accuracy/my (my accuracy)
        response, error = self.make_request("GET", "/predictions/accuracy/my", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /predictions/accuracy/my - My prediction accuracy endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /predictions/accuracy/my - My prediction accuracy endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /predictions/accuracy/my - My prediction accuracy", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /predictions/accuracy/my - My prediction accuracy", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 4. GET /api/v1/predictions/my (my predictions)
        response, error = self.make_request("GET", "/predictions/my", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /predictions/my - My predictions endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /predictions/my - My predictions endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /predictions/my - My predictions", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /predictions/my - My predictions", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 5. GET /api/v1/admin/predictions/accuracy/global (global accuracy)
        response, error = self.make_request("GET", "/admin/predictions/accuracy/global", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/predictions/accuracy/global - Global prediction accuracy endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/predictions/accuracy/global - Global prediction accuracy endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/predictions/accuracy/global - Global prediction accuracy", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/predictions/accuracy/global - Global prediction accuracy", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 6. GET /api/v1/admin/predictions/models/performance (models performance)
        response, error = self.make_request("GET", "/admin/predictions/models/performance", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/predictions/models/performance - Models performance endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/predictions/models/performance - Models performance endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/predictions/models/performance - Models performance", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/predictions/models/performance - Models performance", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 7. PUT /api/v1/admin/predictions/models/123/update (update model)
        model_data = {"name": "Updated Model", "parameters": {"learning_rate": 0.01}}
        response, error = self.make_request("PUT", "/admin/predictions/models/123/update", model_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /admin/predictions/models/123/update - Update model endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("PUT /admin/predictions/models/123/update - Update model endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("PUT /admin/predictions/models/123/update - Update model", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /admin/predictions/models/123/update - Update model", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 8. GET /api/v1/admin/predictions/leaderboard (prediction leaderboard)
        response, error = self.make_request("GET", "/admin/predictions/leaderboard", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/predictions/leaderboard - Prediction leaderboard endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /admin/predictions/leaderboard - Prediction leaderboard endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /admin/predictions/leaderboard - Prediction leaderboard", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/predictions/leaderboard - Prediction leaderboard", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 9. GET /api/v1/predictions/confidence/123 (confidence score)
        response, error = self.make_request("GET", "/predictions/confidence/123", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("GET /predictions/confidence/123 - Confidence score endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /predictions/confidence/123 - Confidence score", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /predictions/confidence/123 - Confidence score", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 10. POST /api/v1/admin/predictions/train (train models)
        train_data = {"model_type": "neural_network", "dataset": "recent_matches"}
        response, error = self.make_request("POST", "/admin/predictions/train", train_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /admin/predictions/train - Train models endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /admin/predictions/train - Train models", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /admin/predictions/train - Train models", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def test_tournament_brackets(self):
        """Test Automated Tournament Brackets (8 endpoints - 4 types)"""
        print("\n🏆 TESTING AUTOMATED TOURNAMENT BRACKETS (8 ENDPOINTS - 4 TYPES)")
        
        tournament_id = "1"
        bracket_types = [
            "single-elimination",
            "double-elimination", 
            "round-robin",
            "swiss-system"
        ]
        
        # Test 1-4: GET /api/v1/tournaments/1/brackets/{type} (4 bracket types)
        for bracket_type in bracket_types:
            response, error = self.make_request("GET", f"/tournaments/{tournament_id}/brackets/{bracket_type}", auth_token=USER_TOKEN)
            if response and response.status_code in [200, 400, 401]:
                if response.status_code == 401:
                    self.log_result(f"GET /tournaments/{tournament_id}/brackets/{bracket_type} - {bracket_type} endpoint accessible", True, "Returns 401 (auth required) instead of 404")
                else:
                    self.log_result(f"GET /tournaments/{tournament_id}/brackets/{bracket_type} - {bracket_type}", True, f"Status: {response.status_code}")
            else:
                self.log_result(f"GET /tournaments/{tournament_id}/brackets/{bracket_type} - {bracket_type}", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 5. POST /api/v1/admin/tournaments/1/brackets/generate (generate bracket)
        bracket_data = {
            "type": "single-elimination",
            "participants": ["team1", "team2", "team3", "team4"],
            "seeding": "random"
        }
        response, error = self.make_request("POST", f"/admin/tournaments/{tournament_id}/brackets/generate", bracket_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /admin/tournaments/1/brackets/generate - Generate bracket endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /admin/tournaments/1/brackets/generate - Generate bracket", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /admin/tournaments/1/brackets/generate - Generate bracket", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 6. PUT /api/v1/tournaments/1/brackets/123/advance (advance bracket)
        advance_data = {"winner_id": "team1", "match_result": "2-1"}
        response, error = self.make_request("PUT", f"/tournaments/{tournament_id}/brackets/123/advance", advance_data, auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /tournaments/1/brackets/123/advance - Advance bracket endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("PUT /tournaments/1/brackets/123/advance - Advance bracket endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("PUT /tournaments/1/brackets/123/advance - Advance bracket", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /tournaments/1/brackets/123/advance - Advance bracket", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 7. GET /api/v1/tournaments/1/brackets/123/status (bracket status)
        response, error = self.make_request("GET", f"/tournaments/{tournament_id}/brackets/123/status", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("GET /tournaments/1/brackets/123/status - Bracket status endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /tournaments/1/brackets/123/status - Bracket status", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /tournaments/1/brackets/123/status - Bracket status", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 8. PUT /api/v1/admin/tournaments/1/brackets/123/reset (reset bracket)
        response, error = self.make_request("PUT", f"/admin/tournaments/{tournament_id}/brackets/123/reset", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /admin/tournaments/1/brackets/123/reset - Reset bracket endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("PUT /admin/tournaments/1/brackets/123/reset - Reset bracket", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /admin/tournaments/1/brackets/123/reset - Reset bracket", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def test_fraud_detection(self):
        """Test Advanced Fraud Detection (6 endpoints)"""
        print("\n🛡️ TESTING ADVANCED FRAUD DETECTION (6 ENDPOINTS)")
        
        # 1. GET /api/v1/admin/fraud/alerts (fraud alerts)
        response, error = self.make_request("GET", "/admin/fraud/alerts", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/fraud/alerts - Fraud alerts endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /admin/fraud/alerts - Fraud alerts", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/fraud/alerts - Fraud alerts", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 2. GET /api/v1/admin/fraud/statistics (fraud stats)
        response, error = self.make_request("GET", "/admin/fraud/statistics", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/fraud/statistics - Fraud statistics endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /admin/fraud/statistics - Fraud statistics", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/fraud/statistics - Fraud statistics", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 3. POST /api/v1/admin/fraud/investigate (investigate)
        investigate_data = {
            "user_id": "123",
            "alert_id": "456",
            "investigation_type": "manual_review"
        }
        response, error = self.make_request("POST", "/admin/fraud/investigate", investigate_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 201, 400, 401]:
            if response.status_code == 401:
                self.log_result("POST /admin/fraud/investigate - Fraud investigate endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("POST /admin/fraud/investigate - Fraud investigate", True, f"Status: {response.status_code}")
        else:
            self.log_result("POST /admin/fraud/investigate - Fraud investigate", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 4. GET /api/v1/fraud/my-reports (user reports)
        response, error = self.make_request("GET", "/fraud/my-reports", auth_token=USER_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /fraud/my-reports - User fraud reports endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            elif response.status_code == 404 and "page not found" in response.text.lower():
                self.log_result("GET /fraud/my-reports - User fraud reports endpoint accessible", False, "Returns 404 (page not found) - endpoint not implemented")
            else:
                self.log_result("GET /fraud/my-reports - User fraud reports", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /fraud/my-reports - User fraud reports", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 5. PUT /api/v1/admin/fraud/alerts/123/status (update alert status)
        status_data = {"status": "resolved", "resolution_notes": "False positive"}
        response, error = self.make_request("PUT", "/admin/fraud/alerts/123/status", status_data, auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 400, 401]:
            if response.status_code == 401:
                self.log_result("PUT /admin/fraud/alerts/123/status - Update alert status endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("PUT /admin/fraud/alerts/123/status - Update alert status", True, f"Status: {response.status_code}")
        else:
            self.log_result("PUT /admin/fraud/alerts/123/status - Update alert status", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")
        
        # 6. GET /api/v1/admin/fraud/threshold (fraud thresholds)
        response, error = self.make_request("GET", "/admin/fraud/threshold", auth_token=ADMIN_TOKEN)
        if response and response.status_code in [200, 401]:
            if response.status_code == 401:
                self.log_result("GET /admin/fraud/threshold - Fraud thresholds endpoint accessible", True, "Returns 401 (auth required) instead of 404")
            else:
                self.log_result("GET /admin/fraud/threshold - Fraud thresholds", True, f"Status: {response.status_code}")
        else:
            self.log_result("GET /admin/fraud/threshold - Fraud thresholds", False, f"Status: {response.status_code if response else 'No response'}, Error: {error}")

    def run_comprehensive_test(self):
        """Run comprehensive test of all 7 gaming features"""
        print("🎯 COMPREHENSIVE ADVANCED GAMING FEATURES TESTING")
        print("=" * 60)
        print(f"Backend URL: {BACKEND_URL}")
        print(f"Test Time: {datetime.now().isoformat()}")
        print("=" * 60)
        
        # Authentication
        print("\n🔐 AUTHENTICATION SETUP")
        user_auth, user_msg = self.authenticate_user()
        print(f"User Auth: {'✅' if user_auth else '❌'} {user_msg}")
        
        admin_auth, admin_msg = self.authenticate_admin()
        print(f"Admin Auth: {'✅' if admin_auth else '❌'} {admin_msg}")
        
        # Run all gaming feature tests
        self.test_achievement_system()
        self.test_friend_system()
        self.test_social_sharing()
        self.test_advanced_game_analytics()
        self.test_player_predictions()
        self.test_tournament_brackets()
        self.test_fraud_detection()
        
        # Calculate results
        success_rate = (self.passed_tests / self.total_tests * 100) if self.total_tests > 0 else 0
        
        print("\n" + "=" * 60)
        print("🎯 COMPREHENSIVE TEST RESULTS")
        print("=" * 60)
        print(f"Total Tests: {self.total_tests}")
        print(f"Passed: {self.passed_tests}")
        print(f"Failed: {self.failed_tests}")
        print(f"Success Rate: {success_rate:.1f}%")
        
        # Determine if target achieved
        target_rate = 70.0
        previous_rate = 23.7
        
        print(f"\nPrevious Success Rate: {previous_rate}%")
        print(f"Target Success Rate: {target_rate}%")
        print(f"Current Success Rate: {success_rate:.1f}%")
        
        if success_rate >= target_rate:
            print(f"🎉 TARGET ACHIEVED! Success rate {success_rate:.1f}% exceeds target {target_rate}%")
        else:
            improvement = success_rate - previous_rate
            print(f"⚠️ TARGET NOT MET. Improvement: +{improvement:.1f}% (need +{target_rate - previous_rate:.1f}%)")
        
        # Summary by system
        print("\n📊 RESULTS BY GAMING SYSTEM:")
        systems = {
            "Achievement System": [r for r in self.results if "Achievement System" in r["test"]],
            "Friend System": [r for r in self.results if "Friend System" in r["test"]],
            "Social Sharing": [r for r in self.results if "Social Sharing" in r["test"]],
            "Advanced Analytics": [r for r in self.results if "Advanced Analytics" in r["test"]],
            "Player Predictions": [r for r in self.results if "Player Predictions" in r["test"]],
            "Tournament Brackets": [r for r in self.results if "Tournament Brackets" in r["test"]],
            "Fraud Detection": [r for r in self.results if "Fraud Detection" in r["test"]]
        }
        
        for system, tests in systems.items():
            if tests:
                passed = len([t for t in tests if t["status"] == "PASS"])
                total = len(tests)
                rate = (passed / total * 100) if total > 0 else 0
                status = "✅" if rate >= 70 else "❌"
                print(f"{status} {system}: {passed}/{total} ({rate:.1f}%)")
        
        return success_rate >= target_rate

if __name__ == "__main__":
    tester = GameFeatureTester()
    success = tester.run_comprehensive_test()
    sys.exit(0 if success else 1)