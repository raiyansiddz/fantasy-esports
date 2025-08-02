#!/usr/bin/env python3
"""
Comprehensive Backend API Testing for Fantasy Esports Referral System
Tests all referral system functionality including:
- User registration with referral codes
- Referral code application and validation
- Referral completion logic (deposit/contest triggers)
- Tier-based reward system
- Referral statistics and leaderboard
- Edge cases and error handling
"""

import requests
import json
import time
import random
import string
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple

class FantasyEsportsAPITester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.users = {}  # Store user data for testing
        self.test_results = []
        
    def log_test(self, test_name: str, success: bool, details: str = ""):
        """Log test results"""
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} {test_name}")
        if details:
            print(f"   Details: {details}")
        self.test_results.append({
            "test": test_name,
            "success": success,
            "details": details,
            "timestamp": datetime.now().isoformat()
        })
        
    def generate_test_mobile(self) -> str:
        """Generate a unique test mobile number"""
        # Generate Indian mobile number in format +91[6-9]XXXXXXXXX
        first_digit = random.choice(['6', '7', '8', '9'])
        remaining_digits = ''.join(random.choices(string.digits, k=9))
        return f"+91{first_digit}{remaining_digits}"
        
    def generate_test_email(self) -> str:
        """Generate a unique test email"""
        return f"test_{random.randint(1000, 9999)}@example.com"

    def test_health_check(self) -> bool:
        """Test basic health endpoint"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            success = response.status_code == 200 and "healthy" in response.text
            self.log_test("Health Check", success, f"Status: {response.status_code}")
            return success
        except Exception as e:
            self.log_test("Health Check", False, f"Error: {str(e)}")
            return False

    def register_user(self, mobile: str, profile_data: Dict, referral_code: Optional[str] = None) -> Optional[Dict]:
        """Register a new user with optional referral code"""
        try:
            # Step 1: Verify mobile (get OTP session)
            verify_payload = {
                "mobile": mobile,
                "country_code": "+91",
                "device_id": f"test_device_{random.randint(1000, 9999)}",
                "app_version": "1.0.0",
                "platform": "android"
            }
            
            if referral_code:
                verify_payload["referral_code"] = referral_code
                
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-mobile", json=verify_payload)
            
            if response.status_code != 200:
                self.log_test(f"User Registration - Mobile Verify ({mobile})", False, 
                            f"Mobile verify failed: {response.status_code} - {response.text}")
                return None
                
            mobile_data = response.json()
            session_id = mobile_data.get("session_id")
            
            if not session_id:
                self.log_test(f"User Registration - Mobile Verify ({mobile})", False, "No session ID received")
                return None
            
            # Step 2: Verify OTP (complete registration)
            otp_payload = {
                "session_id": session_id,
                "otp": "123456",  # Development OTP
                "device_info": {
                    "platform": "android",
                    "device_id": f"test_device_{random.randint(1000, 9999)}",
                    "app_version": "1.0.0"
                },
                "profile_data": profile_data
            }
            
            if referral_code:
                otp_payload["referral_code"] = referral_code
                
            response = self.session.post(f"{self.base_url}/api/v1/auth/verify-otp", json=otp_payload)
            
            if response.status_code != 200:
                self.log_test(f"User Registration - OTP Verify ({mobile})", False, 
                            f"OTP verify failed: {response.status_code} - {response.text}")
                return None
                
            user_data = response.json()
            
            if not user_data.get("success"):
                self.log_test(f"User Registration - OTP Verify ({mobile})", False, 
                            f"Registration failed: {user_data}")
                return None
                
            # Store user data
            user_info = {
                "mobile": mobile,
                "access_token": user_data.get("access_token"),
                "user": user_data.get("user"),
                "referral_code": user_data.get("user", {}).get("referral_code"),
                "referred_by": referral_code
            }
            
            self.users[mobile] = user_info
            self.log_test(f"User Registration ({mobile})", True, 
                        f"User ID: {user_info['user']['id']}, Referral Code: {user_info['referral_code']}")
            return user_info
            
        except Exception as e:
            self.log_test(f"User Registration ({mobile})", False, f"Error: {str(e)}")
            return None

    def get_auth_headers(self, mobile: str) -> Dict[str, str]:
        """Get authorization headers for a user"""
        user = self.users.get(mobile)
        if not user or not user.get("access_token"):
            return {}
        return {"Authorization": f"Bearer {user['access_token']}"}

    def test_referral_stats(self, mobile: str) -> Optional[Dict]:
        """Test getting referral statistics for a user"""
        try:
            headers = self.get_auth_headers(mobile)
            if not headers:
                self.log_test(f"Referral Stats ({mobile})", False, "No auth token")
                return None
                
            response = self.session.get(f"{self.base_url}/api/v1/referrals/my-stats", headers=headers)
            
            if response.status_code != 200:
                self.log_test(f"Referral Stats ({mobile})", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"Referral Stats ({mobile})", False, f"API returned error: {data}")
                return None
                
            stats = data.get("referral_stats", {})
            self.log_test(f"Referral Stats ({mobile})", True, 
                        f"Total: {stats.get('total_referrals', 0)}, "
                        f"Successful: {stats.get('successful_referrals', 0)}, "
                        f"Earnings: ₹{stats.get('total_earnings', 0)}, "
                        f"Tier: {stats.get('current_tier', 'N/A')}")
            return stats
            
        except Exception as e:
            self.log_test(f"Referral Stats ({mobile})", False, f"Error: {str(e)}")
            return None

    def test_referral_history(self, mobile: str) -> Optional[List]:
        """Test getting referral history for a user"""
        try:
            headers = self.get_auth_headers(mobile)
            if not headers:
                self.log_test(f"Referral History ({mobile})", False, "No auth token")
                return None
                
            response = self.session.get(f"{self.base_url}/api/v1/referrals/history", headers=headers)
            
            if response.status_code != 200:
                self.log_test(f"Referral History ({mobile})", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"Referral History ({mobile})", False, f"API returned error: {data}")
                return None
                
            referrals = data.get("referrals", [])
            self.log_test(f"Referral History ({mobile})", True, 
                        f"Found {len(referrals)} referral records")
            return referrals
            
        except Exception as e:
            self.log_test(f"Referral History ({mobile})", False, f"Error: {str(e)}")
            return None

    def test_apply_referral_code(self, mobile: str, referral_code: str) -> bool:
        """Test applying a referral code after registration"""
        try:
            headers = self.get_auth_headers(mobile)
            if not headers:
                self.log_test(f"Apply Referral Code ({mobile})", False, "No auth token")
                return False
                
            payload = {"referral_code": referral_code}
            response = self.session.post(f"{self.base_url}/api/v1/referrals/apply", 
                                       json=payload, headers=headers)
            
            success = response.status_code == 200
            if success:
                data = response.json()
                success = data.get("success", False)
                
            self.log_test(f"Apply Referral Code ({mobile})", success, 
                        f"Code: {referral_code}, Response: {response.status_code}")
            return success
            
        except Exception as e:
            self.log_test(f"Apply Referral Code ({mobile})", False, f"Error: {str(e)}")
            return False

    def test_wallet_deposit(self, mobile: str, amount: float = 100.0) -> bool:
        """Test wallet deposit (triggers referral completion)"""
        try:
            headers = self.get_auth_headers(mobile)
            if not headers:
                self.log_test(f"Wallet Deposit ({mobile})", False, "No auth token")
                return False
                
            payload = {
                "amount": amount,
                "payment_method": "upi",
                "return_url": "https://example.com/return"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/wallet/deposit", 
                                       json=payload, headers=headers)
            
            success = response.status_code == 200
            if success:
                data = response.json()
                success = data.get("success", False)
                
            self.log_test(f"Wallet Deposit ({mobile})", success, 
                        f"Amount: ₹{amount}, Response: {response.status_code}")
            return success
            
        except Exception as e:
            self.log_test(f"Wallet Deposit ({mobile})", False, f"Error: {str(e)}")
            return False

    def test_wallet_balance(self, mobile: str) -> Optional[Dict]:
        """Test getting wallet balance"""
        try:
            headers = self.get_auth_headers(mobile)
            if not headers:
                self.log_test(f"Wallet Balance ({mobile})", False, "No auth token")
                return None
                
            response = self.session.get(f"{self.base_url}/api/v1/wallet/balance", headers=headers)
            
            if response.status_code != 200:
                self.log_test(f"Wallet Balance ({mobile})", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test(f"Wallet Balance ({mobile})", False, f"API returned error: {data}")
                return None
                
            balance = data.get("balance", {})
            self.log_test(f"Wallet Balance ({mobile})", True, 
                        f"Total: ₹{balance.get('total_balance', 0)}, "
                        f"Bonus: ₹{balance.get('bonus_balance', 0)}, "
                        f"Deposit: ₹{balance.get('deposit_balance', 0)}")
            return balance
            
        except Exception as e:
            self.log_test(f"Wallet Balance ({mobile})", False, f"Error: {str(e)}")
            return None

    def test_referral_leaderboard(self) -> Optional[List]:
        """Test referral leaderboard"""
        try:
            # Use any authenticated user for this test
            if not self.users:
                self.log_test("Referral Leaderboard", False, "No authenticated users available")
                return None
                
            mobile = list(self.users.keys())[0]
            headers = self.get_auth_headers(mobile)
            
            response = self.session.get(f"{self.base_url}/api/v1/referrals/leaderboard", headers=headers)
            
            if response.status_code != 200:
                self.log_test("Referral Leaderboard", False, 
                            f"Failed: {response.status_code} - {response.text}")
                return None
                
            data = response.json()
            if not data.get("success"):
                self.log_test("Referral Leaderboard", False, f"API returned error: {data}")
                return None
                
            leaderboard = data.get("leaderboard", [])
            self.log_test("Referral Leaderboard", True, 
                        f"Found {len(leaderboard)} entries")
            return leaderboard
            
        except Exception as e:
            self.log_test("Referral Leaderboard", False, f"Error: {str(e)}")
            return None

    def test_edge_cases(self):
        """Test edge cases and error scenarios"""
        print("\n🧪 Testing Edge Cases...")
        
        # Test 1: Invalid referral code
        mobile1 = self.generate_test_mobile()
        profile1 = {
            "first_name": "EdgeCase",
            "last_name": "User1",
            "email": self.generate_test_email(),
            "date_of_birth": "1995-01-01T00:00:00Z",
            "state": "Maharashtra"
        }
        
        user1 = self.register_user(mobile1, profile1, "INVALID_CODE")
        if user1:
            self.log_test("Edge Case - Invalid Referral Code", False, 
                        "Should have failed with invalid referral code")
        else:
            self.log_test("Edge Case - Invalid Referral Code", True, 
                        "Correctly rejected invalid referral code")
        
        # Test 2: Self-referral attempt
        mobile2 = self.generate_test_mobile()
        profile2 = {
            "first_name": "EdgeCase",
            "last_name": "User2",
            "email": self.generate_test_email(),
            "date_of_birth": "1995-01-01T00:00:00Z",
            "state": "Maharashtra"
        }
        
        user2 = self.register_user(mobile2, profile2)
        if user2:
            # Try to apply own referral code
            success = self.test_apply_referral_code(mobile2, user2["referral_code"])
            self.log_test("Edge Case - Self Referral", not success, 
                        "Self-referral should be rejected")

    def run_comprehensive_referral_test(self):
        """Run comprehensive referral system test"""
        print("🚀 Starting Comprehensive Referral System Test")
        print("=" * 60)
        
        # Test 1: Health Check
        if not self.test_health_check():
            print("❌ Server health check failed. Aborting tests.")
            return
            
        print("\n📝 Testing User Registration and Referral Flow...")
        
        # Test 2: Register referrer user (User A)
        mobile_a = self.generate_test_mobile()
        profile_a = {
            "first_name": "Referrer",
            "last_name": "UserA",
            "email": self.generate_test_email(),
            "date_of_birth": "1990-01-01T00:00:00Z",
            "state": "Karnataka"
        }
        
        user_a = self.register_user(mobile_a, profile_a)
        if not user_a:
            print("❌ Failed to register referrer user. Aborting tests.")
            return
            
        referral_code_a = user_a["referral_code"]
        print(f"✅ Referrer User A registered with code: {referral_code_a}")
        
        # Test 3: Register referred user (User B) with A's referral code
        mobile_b = self.generate_test_mobile()
        profile_b = {
            "first_name": "Referred",
            "last_name": "UserB",
            "email": self.generate_test_email(),
            "date_of_birth": "1992-01-01T00:00:00Z",
            "state": "Maharashtra"
        }
        
        user_b = self.register_user(mobile_b, profile_b, referral_code_a)
        if not user_b:
            print("❌ Failed to register referred user. Aborting tests.")
            return
            
        print(f"✅ Referred User B registered with referral code: {referral_code_a}")
        
        # Test 4: Check initial referral stats for User A
        print("\n📊 Testing Referral Statistics...")
        stats_a_initial = self.test_referral_stats(mobile_a)
        
        # Test 5: Check referral history for User A
        history_a = self.test_referral_history(mobile_a)
        
        # Test 6: Check wallet balances before deposit
        print("\n💰 Testing Wallet Operations...")
        balance_a_before = self.test_wallet_balance(mobile_a)
        balance_b_before = self.test_wallet_balance(mobile_b)
        
        # Test 7: User B makes a deposit (should trigger referral completion)
        print("\n🏦 Testing Deposit and Referral Completion...")
        deposit_success = self.test_wallet_deposit(mobile_b, 500.0)
        
        if deposit_success:
            # Wait a moment for referral processing
            time.sleep(2)
            
            # Test 8: Check wallet balances after deposit
            balance_a_after = self.test_wallet_balance(mobile_a)
            balance_b_after = self.test_wallet_balance(mobile_b)
            
            # Test 9: Check updated referral stats for User A
            stats_a_after = self.test_referral_stats(mobile_a)
            
            # Verify referral completion
            if (stats_a_after and stats_a_initial and 
                stats_a_after.get("successful_referrals", 0) > stats_a_initial.get("successful_referrals", 0)):
                self.log_test("Referral Completion", True, 
                            "Referral successfully completed after deposit")
            else:
                self.log_test("Referral Completion", False, 
                            "Referral not completed after deposit")
        
        # Test 10: Test multiple referrals for tier progression
        print("\n🏆 Testing Multiple Referrals and Tier System...")
        
        # Register multiple users with User A's referral code
        for i in range(3):
            mobile_ref = self.generate_test_mobile()
            profile_ref = {
                "first_name": f"Referred{i+2}",
                "last_name": f"User{chr(67+i)}",  # C, D, E
                "email": self.generate_test_email(),
                "date_of_birth": f"199{3+i}-01-01T00:00:00Z",
                "state": "Tamil Nadu"
            }
            
            user_ref = self.register_user(mobile_ref, profile_ref, referral_code_a)
            if user_ref:
                # Make a small deposit to complete referral
                self.test_wallet_deposit(mobile_ref, 100.0)
                time.sleep(1)  # Brief pause between operations
        
        # Test 11: Check final referral stats and tier
        print("\n📈 Final Referral Statistics...")
        final_stats = self.test_referral_stats(mobile_a)
        final_balance = self.test_wallet_balance(mobile_a)
        
        # Test 12: Test referral leaderboard
        print("\n🏅 Testing Referral Leaderboard...")
        leaderboard = self.test_referral_leaderboard()
        
        # Test 13: Edge cases
        self.test_edge_cases()
        
        # Test Summary
        print("\n" + "=" * 60)
        print("📋 TEST SUMMARY")
        print("=" * 60)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        
        print(f"Total Tests: {total_tests}")
        print(f"✅ Passed: {passed_tests}")
        print(f"❌ Failed: {failed_tests}")
        print(f"Success Rate: {(passed_tests/total_tests)*100:.1f}%")
        
        if failed_tests > 0:
            print("\n❌ FAILED TESTS:")
            for result in self.test_results:
                if not result["success"]:
                    print(f"  - {result['test']}: {result['details']}")
        
        # Save detailed results
        with open("/app/referral_test_results.json", "w") as f:
            json.dump({
                "summary": {
                    "total_tests": total_tests,
                    "passed": passed_tests,
                    "failed": failed_tests,
                    "success_rate": (passed_tests/total_tests)*100
                },
                "test_results": self.test_results,
                "users_created": len(self.users),
                "timestamp": datetime.now().isoformat()
            }, f, indent=2)
        
        print(f"\n📄 Detailed results saved to: /app/referral_test_results.json")
        
        return passed_tests == total_tests

def main():
    """Main test execution"""
    tester = FantasyEsportsAPITester()
    success = tester.run_comprehensive_referral_test()
    
    if success:
        print("\n🎉 All tests passed! Referral system is working correctly.")
        exit(0)
    else:
        print("\n⚠️  Some tests failed. Check the results above.")
        exit(1)

if __name__ == "__main__":
    main()