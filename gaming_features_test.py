#!/usr/bin/env python3
"""
Comprehensive Gaming Features Testing for GoLang Fantasy Esports Backend
Testing all 7 advanced gaming features as requested in the review

Features to test:
1. Achievement System & Badges
2. Friend System & Challenges  
3. Social Sharing Integration
4. Advanced Game Analytics
5. Automated Tournament Brackets
6. Player Performance Predictions
7. Advanced Fraud Detection
"""

import requests
import json
import time
import uuid
from typing import Dict, Any, Optional, Tuple, List

class GamingFeaturesTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        self.created_resources = {
            "achievements": [],
            "friends": [],
            "challenges": [],
            "shares": [],
            "brackets": [],
            "predictions": [],
            "fraud_reports": []
        }
        
    def log_test(self, test_name: str, success: bool, details: str = "", response_data: Any = None):
        """Log test results"""
        result = {
            "test": test_name,
            "success": success,
            "details": details,
            "response_data": response_data,
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
        }
        self.test_results.append(result)
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status}: {test_name}")
        if details:
            print(f"   Details: {details}")
        if not success and response_data:
            print(f"   Response: {response_data}")
        print()

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user"""
        try:
            # Try to get admin token using standard admin credentials
            login_data = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/v1/admin/login",
                json=login_data,
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                if 'access_token' in data:
                    self.admin_token = data['access_token']
                    self.session.headers.update({'Authorization': f'Bearer {self.admin_token}'})
                    self.log_test("Admin Authentication", True, f"Successfully authenticated as admin")
                    return True
                    
            self.log_test("Admin Authentication", False, f"Failed to authenticate admin: {response.status_code} - {response.text}")
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception during admin auth: {str(e)}")
            return False

    def authenticate_user(self) -> bool:
        """Authenticate as regular user"""
        try:
            # Step 1: Verify mobile number
            mobile_data = {
                "mobile": "+919876543210",
                "country_code": "+91"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/v1/auth/verify-mobile",
                json=mobile_data,
                timeout=10
            )
            
            if response.status_code != 200:
                self.log_test("User Mobile Verification", False, f"Mobile verification failed: {response.status_code}")
                return False
            
            # Step 2: Verify OTP
            otp_data = {
                "mobile": "+919876543210",
                "otp": "123456",
                "name": "Test User",
                "email": "testuser@example.com"
            }
            
            response = self.session.post(
                f"{self.base_url}/api/v1/auth/verify-otp",
                json=otp_data,
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                if 'token' in data:
                    self.user_token = data['token']
                    # Create a separate session for user requests
                    self.user_session = requests.Session()
                    self.user_session.headers.update({'Authorization': f'Bearer {self.user_token}'})
                    self.log_test("User Authentication", True, "Successfully authenticated as user")
                    return True
                    
            self.log_test("User Authentication", False, f"Failed to authenticate user: {response.status_code} - {response.text}")
            return False
            
        except Exception as e:
            self.log_test("User Authentication", False, f"Exception during user auth: {str(e)}")
            return False

    # FEATURE 1: ACHIEVEMENT SYSTEM & BADGES
    def test_achievement_system(self):
        """Test Achievement System & Badges functionality"""
        print("🏆 TESTING FEATURE 1: ACHIEVEMENT SYSTEM & BADGES")
        print("=" * 60)
        
        # Test admin achievement creation
        achievement_data = {
            "name": "First Victory",
            "description": "Win your first contest",
            "badge_icon": "https://example.com/victory-badge.png",
            "category": "contest",
            "trigger_criteria": {
                "type": "contest_win",
                "count": 1
            },
            "reward_type": "bonus",
            "reward_amount": 100,
            "is_active": True
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/api/v1/admin/achievements",
                json=achievement_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                if 'id' in data:
                    achievement_id = data['id']
                    self.created_resources["achievements"].append(achievement_id)
                    self.log_test("Create Achievement", True, f"Created achievement with ID: {achievement_id}")
                else:
                    self.log_test("Create Achievement", True, "Achievement created successfully")
            else:
                self.log_test("Create Achievement", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Create Achievement", False, f"Exception: {str(e)}")

        # Test get all achievements (admin)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/achievements", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("List Admin Achievements", True, f"Retrieved achievements list")
            else:
                self.log_test("List Admin Achievements", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("List Admin Achievements", False, f"Exception: {str(e)}")

        # Test user achievements endpoints
        if hasattr(self, 'user_session'):
            try:
                response = self.user_session.get(f"{self.base_url}/api/v1/achievements", timeout=10)
                
                if response.status_code == 200:
                    data = response.json()
                    self.log_test("Get User Achievements", True, f"Retrieved user achievements")
                else:
                    self.log_test("Get User Achievements", False, f"Status: {response.status_code}", response.text)
                    
            except Exception as e:
                self.log_test("Get User Achievements", False, f"Exception: {str(e)}")

            try:
                response = self.user_session.get(f"{self.base_url}/api/v1/achievements/my", timeout=10)
                
                if response.status_code == 200:
                    data = response.json()
                    self.log_test("Get My Achievements", True, f"Retrieved my achievements")
                else:
                    self.log_test("Get My Achievements", False, f"Status: {response.status_code}", response.text)
                    
            except Exception as e:
                self.log_test("Get My Achievements", False, f"Exception: {str(e)}")

    # FEATURE 2: FRIEND SYSTEM & CHALLENGES
    def test_friend_system(self):
        """Test Friend System & Challenges functionality"""
        print("👥 TESTING FEATURE 2: FRIEND SYSTEM & CHALLENGES")
        print("=" * 60)
        
        if not hasattr(self, 'user_session'):
            self.log_test("Friend System Setup", False, "User authentication required")
            return

        # Test add friend request
        friend_data = {
            "friend_mobile": "+919876543211",
            "message": "Let's be friends!"
        }
        
        try:
            response = self.user_session.post(
                f"{self.base_url}/api/v1/friends",
                json=friend_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                self.log_test("Add Friend Request", True, f"Friend request sent successfully")
            else:
                self.log_test("Add Friend Request", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Add Friend Request", False, f"Exception: {str(e)}")

        # Test get friends list
        try:
            response = self.user_session.get(f"{self.base_url}/api/v1/friends", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Friends List", True, f"Retrieved friends list")
            else:
                self.log_test("Get Friends List", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Friends List", False, f"Exception: {str(e)}")

        # Test create challenge
        challenge_data = {
            "friend_id": "test-friend-id",
            "contest_id": "test-contest-id",
            "entry_fee": 50,
            "prize_amount": 100,
            "message": "Challenge accepted!"
        }
        
        try:
            response = self.user_session.post(
                f"{self.base_url}/api/v1/challenges",
                json=challenge_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                self.log_test("Create Challenge", True, f"Challenge created successfully")
            else:
                self.log_test("Create Challenge", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Create Challenge", False, f"Exception: {str(e)}")

        # Test get challenges
        try:
            response = self.user_session.get(f"{self.base_url}/api/v1/challenges", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Challenges", True, f"Retrieved challenges list")
            else:
                self.log_test("Get Challenges", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Challenges", False, f"Exception: {str(e)}")

        # Test friend activities
        try:
            response = self.user_session.get(f"{self.base_url}/api/v1/friends/activities", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Friend Activities", True, f"Retrieved friend activities")
            else:
                self.log_test("Get Friend Activities", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Friend Activities", False, f"Exception: {str(e)}")

    # FEATURE 3: SOCIAL SHARING INTEGRATION
    def test_social_sharing(self):
        """Test Social Sharing Integration functionality"""
        print("📱 TESTING FEATURE 3: SOCIAL SHARING INTEGRATION")
        print("=" * 60)
        
        if not hasattr(self, 'user_session'):
            self.log_test("Social Sharing Setup", False, "User authentication required")
            return

        # Test create share record
        share_data = {
            "content_type": "team",
            "content_id": "test-team-id",
            "platform": "twitter",
            "share_url": "https://twitter.com/share?text=Check%20out%20my%20team"
        }
        
        try:
            response = self.user_session.post(
                f"{self.base_url}/api/v1/share",
                json=share_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                if 'id' in data:
                    share_id = data['id']
                    self.created_resources["shares"].append(share_id)
                self.log_test("Create Share Record", True, f"Share record created successfully")
            else:
                self.log_test("Create Share Record", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Create Share Record", False, f"Exception: {str(e)}")

        # Test generate team share URLs
        try:
            response = self.user_session.get(
                f"{self.base_url}/api/v1/share/teams/test-team-id/urls",
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Generate Team Share URLs", True, f"Generated team share URLs")
            else:
                self.log_test("Generate Team Share URLs", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Generate Team Share URLs", False, f"Exception: {str(e)}")

        # Test generate contest win share URLs
        try:
            response = self.user_session.get(
                f"{self.base_url}/api/v1/share/contests/test-contest-id/urls",
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Generate Contest Win URLs", True, f"Generated contest win share URLs")
            else:
                self.log_test("Generate Contest Win URLs", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Generate Contest Win URLs", False, f"Exception: {str(e)}")

        # Test generate achievement share URLs
        try:
            response = self.user_session.get(
                f"{self.base_url}/api/v1/share/achievements/test-achievement-id/urls",
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Generate Achievement URLs", True, f"Generated achievement share URLs")
            else:
                self.log_test("Generate Achievement URLs", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Generate Achievement URLs", False, f"Exception: {str(e)}")

        # Test get user shares
        try:
            response = self.user_session.get(f"{self.base_url}/api/v1/share/my", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get User Shares", True, f"Retrieved user shares")
            else:
                self.log_test("Get User Shares", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get User Shares", False, f"Exception: {str(e)}")

        # Test share analytics (admin)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/social/analytics", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Share Analytics", True, f"Retrieved share analytics")
            else:
                self.log_test("Get Share Analytics", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Share Analytics", False, f"Exception: {str(e)}")

    # FEATURE 4: ADVANCED GAME ANALYTICS
    def test_advanced_analytics(self):
        """Test Advanced Game Analytics functionality"""
        print("📊 TESTING FEATURE 4: ADVANCED GAME ANALYTICS")
        print("=" * 60)
        
        # Test advanced game metrics
        try:
            response = self.session.get(
                f"{self.base_url}/api/v1/admin/games/test-game-id/advanced-metrics",
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Advanced Game Metrics", True, f"Retrieved advanced metrics")
            else:
                self.log_test("Get Advanced Game Metrics", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Advanced Game Metrics", False, f"Exception: {str(e)}")

        # Test metrics history
        try:
            response = self.session.get(
                f"{self.base_url}/api/v1/admin/games/test-game-id/metrics-history",
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Metrics History", True, f"Retrieved metrics history")
            else:
                self.log_test("Get Metrics History", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Metrics History", False, f"Exception: {str(e)}")

        # Test game comparison
        try:
            params = {
                "game_ids": "test-game-1,test-game-2",
                "metrics": "player_efficiency,team_synergy"
            }
            response = self.session.get(
                f"{self.base_url}/api/v1/admin/games/compare",
                params=params,
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Compare Games", True, f"Retrieved game comparison")
            else:
                self.log_test("Compare Games", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Compare Games", False, f"Exception: {str(e)}")

    # FEATURE 5: AUTOMATED TOURNAMENT BRACKETS
    def test_tournament_brackets(self):
        """Test Automated Tournament Brackets functionality"""
        print("🏆 TESTING FEATURE 5: AUTOMATED TOURNAMENT BRACKETS")
        print("=" * 60)
        
        # Test create tournament bracket
        bracket_data = {
            "tournament_id": "test-tournament-id",
            "bracket_type": "single_elimination",
            "teams": ["team1", "team2", "team3", "team4"],
            "settings": {
                "seeded": True,
                "randomize": False
            }
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/api/v1/admin/tournaments/brackets",
                json=bracket_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                if 'id' in data:
                    bracket_id = data['id']
                    self.created_resources["brackets"].append(bracket_id)
                self.log_test("Create Tournament Bracket", True, f"Tournament bracket created successfully")
            else:
                self.log_test("Create Tournament Bracket", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Create Tournament Bracket", False, f"Exception: {str(e)}")

        # Test get bracket types
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/brackets/types", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Bracket Types", True, f"Retrieved bracket types")
            else:
                self.log_test("Get Bracket Types", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Bracket Types", False, f"Exception: {str(e)}")

        # Test get tournament brackets
        try:
            response = self.session.get(
                f"{self.base_url}/api/v1/admin/tournaments/test-tournament-id/brackets",
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Tournament Brackets", True, f"Retrieved tournament brackets")
            else:
                self.log_test("Get Tournament Brackets", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Tournament Brackets", False, f"Exception: {str(e)}")

        # Test bracket advancement (if we have a bracket ID)
        if self.created_resources["brackets"]:
            bracket_id = self.created_resources["brackets"][0]
            advance_data = {
                "match_id": "test-match-id",
                "winner_team_id": "team1"
            }
            
            try:
                response = self.session.put(
                    f"{self.base_url}/api/v1/admin/brackets/{bracket_id}/advance",
                    json=advance_data,
                    timeout=10
                )
                
                if response.status_code == 200:
                    data = response.json()
                    self.log_test("Advance Bracket", True, f"Bracket advanced successfully")
                else:
                    self.log_test("Advance Bracket", False, f"Status: {response.status_code}", response.text)
                    
            except Exception as e:
                self.log_test("Advance Bracket", False, f"Exception: {str(e)}")

    # FEATURE 6: PLAYER PERFORMANCE PREDICTIONS
    def test_player_predictions(self):
        """Test Player Performance Predictions functionality"""
        print("🔮 TESTING FEATURE 6: PLAYER PERFORMANCE PREDICTIONS")
        print("=" * 60)
        
        # Test generate match predictions (admin)
        prediction_data = {
            "factors": {
                "recent_form": 0.8,
                "head_to_head": 0.6,
                "team_strength": 0.9,
                "map_performance": 0.7,
                "team_morale": 0.8
            }
        }
        
        try:
            response = self.session.post(
                f"{self.base_url}/api/v1/admin/matches/test-match-id/generate-predictions",
                json=prediction_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                self.log_test("Generate Match Predictions", True, f"Match predictions generated successfully")
            else:
                self.log_test("Generate Match Predictions", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Generate Match Predictions", False, f"Exception: {str(e)}")

        # Test get match predictions (user)
        if hasattr(self, 'user_session'):
            try:
                response = self.user_session.get(
                    f"{self.base_url}/api/v1/matches/test-match-id/predictions",
                    timeout=10
                )
                
                if response.status_code == 200:
                    data = response.json()
                    self.log_test("Get Match Predictions", True, f"Retrieved match predictions")
                else:
                    self.log_test("Get Match Predictions", False, f"Status: {response.status_code}", response.text)
                    
            except Exception as e:
                self.log_test("Get Match Predictions", False, f"Exception: {str(e)}")

        # Test update prediction accuracy (admin)
        accuracy_data = {
            "actual_results": {
                "team1_score": 16,
                "team2_score": 14,
                "winner": "team1"
            },
            "prediction_accuracy": 0.85
        }
        
        try:
            response = self.session.put(
                f"{self.base_url}/api/v1/admin/matches/test-match-id/update-accuracy",
                json=accuracy_data,
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Update Prediction Accuracy", True, f"Prediction accuracy updated successfully")
            else:
                self.log_test("Update Prediction Accuracy", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Update Prediction Accuracy", False, f"Exception: {str(e)}")

        # Test get prediction analytics (admin)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/predictions/analytics", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Prediction Analytics", True, f"Retrieved prediction analytics")
            else:
                self.log_test("Get Prediction Analytics", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Prediction Analytics", False, f"Exception: {str(e)}")

    # FEATURE 7: ADVANCED FRAUD DETECTION
    def test_fraud_detection(self):
        """Test Advanced Fraud Detection functionality"""
        print("🛡️ TESTING FEATURE 7: ADVANCED FRAUD DETECTION")
        print("=" * 60)
        
        # Test report suspicious activity (public endpoint)
        fraud_report_data = {
            "user_id": "test-user-id",
            "activity_type": "multiple_accounts",
            "description": "User appears to have multiple accounts with same device fingerprint",
            "evidence": {
                "device_fingerprint": "abc123def456",
                "ip_addresses": ["192.168.1.1", "192.168.1.2"],
                "account_creation_times": ["2025-01-01T10:00:00Z", "2025-01-01T10:05:00Z"]
            }
        }
        
        # Create a session without auth for public endpoint
        public_session = requests.Session()
        
        try:
            response = public_session.post(
                f"{self.base_url}/api/v1/fraud/report",
                json=fraud_report_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                self.log_test("Report Suspicious Activity", True, f"Fraud report submitted successfully")
            else:
                self.log_test("Report Suspicious Activity", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Report Suspicious Activity", False, f"Exception: {str(e)}")

        # Test fraud webhook (public endpoint)
        webhook_data = {
            "event_type": "fraud_detected",
            "user_id": "test-user-id",
            "fraud_type": "bot_behavior",
            "confidence_score": 0.95,
            "timestamp": "2025-01-01T10:00:00Z"
        }
        
        try:
            response = public_session.post(
                f"{self.base_url}/api/v1/fraud/webhook",
                json=webhook_data,
                timeout=10
            )
            
            if response.status_code in [200, 201]:
                data = response.json()
                self.log_test("Fraud Detection Webhook", True, f"Fraud webhook processed successfully")
            else:
                self.log_test("Fraud Detection Webhook", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Fraud Detection Webhook", False, f"Exception: {str(e)}")

        # Test get fraud alerts (admin)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/alerts", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Fraud Alerts", True, f"Retrieved fraud alerts")
            else:
                self.log_test("Get Fraud Alerts", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Fraud Alerts", False, f"Exception: {str(e)}")

        # Test update alert status (admin)
        alert_update_data = {
            "status": "reviewed",
            "notes": "Alert reviewed and marked as false positive"
        }
        
        try:
            response = self.session.put(
                f"{self.base_url}/api/v1/admin/fraud/alerts/test-alert-id",
                json=alert_update_data,
                timeout=10
            )
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Update Alert Status", True, f"Alert status updated successfully")
            else:
                self.log_test("Update Alert Status", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Update Alert Status", False, f"Exception: {str(e)}")

        # Test get fraud statistics (admin)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/fraud/statistics", timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                self.log_test("Get Fraud Statistics", True, f"Retrieved fraud statistics")
            else:
                self.log_test("Get Fraud Statistics", False, f"Status: {response.status_code}", response.text)
                
        except Exception as e:
            self.log_test("Get Fraud Statistics", False, f"Exception: {str(e)}")

    def run_comprehensive_test(self):
        """Run comprehensive test of all 7 gaming features"""
        print("🎮 COMPREHENSIVE GAMING FEATURES TESTING")
        print("=" * 80)
        print("Testing all 7 advanced gaming features in Go-based Fantasy Esports backend")
        print("=" * 80)
        print()
        
        # Phase 1: Authentication
        print("🔐 PHASE 1: AUTHENTICATION")
        print("-" * 40)
        admin_auth_success = self.authenticate_admin()
        user_auth_success = self.authenticate_user()
        print()
        
        if not admin_auth_success:
            print("❌ Admin authentication failed. Some tests will be skipped.")
        if not user_auth_success:
            print("❌ User authentication failed. Some user tests will be skipped.")
        print()
        
        # Phase 2: Test all 7 gaming features
        print("🎮 PHASE 2: GAMING FEATURES TESTING")
        print("-" * 40)
        
        self.test_achievement_system()
        self.test_friend_system()
        self.test_social_sharing()
        self.test_advanced_analytics()
        self.test_tournament_brackets()
        self.test_player_predictions()
        self.test_fraud_detection()
        
        # Phase 3: Generate summary
        self.generate_summary()

    def generate_summary(self):
        """Generate comprehensive test summary"""
        print("📊 COMPREHENSIVE TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result['success'])
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests}")
        print(f"Failed: {failed_tests}")
        print(f"Success Rate: {success_rate:.1f}%")
        print()
        
        # Group results by feature
        features = {
            "Achievement System": [],
            "Friend System": [],
            "Social Sharing": [],
            "Advanced Analytics": [],
            "Tournament Brackets": [],
            "Player Predictions": [],
            "Fraud Detection": []
        }
        
        for result in self.test_results:
            test_name = result['test']
            if any(keyword in test_name.lower() for keyword in ['achievement', 'badge']):
                features["Achievement System"].append(result)
            elif any(keyword in test_name.lower() for keyword in ['friend', 'challenge']):
                features["Friend System"].append(result)
            elif any(keyword in test_name.lower() for keyword in ['share', 'social']):
                features["Social Sharing"].append(result)
            elif any(keyword in test_name.lower() for keyword in ['analytics', 'metrics']):
                features["Advanced Analytics"].append(result)
            elif any(keyword in test_name.lower() for keyword in ['bracket', 'tournament']):
                features["Tournament Brackets"].append(result)
            elif any(keyword in test_name.lower() for keyword in ['prediction']):
                features["Player Predictions"].append(result)
            elif any(keyword in test_name.lower() for keyword in ['fraud', 'alert']):
                features["Fraud Detection"].append(result)
        
        print("📋 FEATURE-WISE RESULTS:")
        print("-" * 40)
        
        for feature_name, results in features.items():
            if results:
                feature_passed = sum(1 for r in results if r['success'])
                feature_total = len(results)
                feature_rate = (feature_passed / feature_total * 100) if feature_total > 0 else 0
                status = "✅" if feature_rate >= 70 else "⚠️" if feature_rate >= 50 else "❌"
                print(f"{status} {feature_name}: {feature_passed}/{feature_total} ({feature_rate:.1f}%)")
        
        print()
        print("🔍 FAILED TESTS DETAILS:")
        print("-" * 40)
        
        failed_results = [r for r in self.test_results if not r['success']]
        if failed_results:
            for result in failed_results:
                print(f"❌ {result['test']}: {result['details']}")
        else:
            print("🎉 No failed tests!")
        
        print()
        print("=" * 80)
        print("🎮 GAMING FEATURES TESTING COMPLETED")
        print("=" * 80)

if __name__ == "__main__":
    tester = GamingFeaturesTester()
    tester.run_comprehensive_test()