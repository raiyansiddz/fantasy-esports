#!/usr/bin/env python3
"""
Comprehensive Content Management System Testing for GoLang Fantasy Esports Backend
Testing all CMS endpoints including banners, email templates, marketing campaigns, 
SEO content, FAQ management, legal documents, and content analytics
"""

import requests
import json
import time
import uuid
from typing import Dict, Any, Optional, Tuple
from datetime import datetime, timedelta

class ContentManagementTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.test_results = []
        self.created_resources = {
            'banners': [],
            'email_templates': [],
            'campaigns': [],
            'seo_content': [],
            'faq_sections': [],
            'faq_items': [],
            'legal_documents': []
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

    def test_health_check(self) -> bool:
        """Test basic health check endpoint"""
        try:
            response = self.session.get(f"{self.base_url}/health")
            success = response.status_code == 200 and "healthy" in response.text
            
            self.log_test(
                "Health Check - Backend Connectivity",
                success,
                f"Status: {response.status_code}, Response: {response.text[:100]}",
                response.json() if success else response.text
            )
            return success
        except Exception as e:
            self.log_test("Health Check - Backend Connectivity", False, f"Exception: {str(e)}")
            return False

    def authenticate_admin(self) -> bool:
        """Authenticate as admin user"""
        try:
            login_data = {
                "username": "admin",
                "password": "admin123"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/login", json=login_data)
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "access_token" in data:
                    self.admin_token = data["access_token"]
                    self.session.headers.update({"Authorization": f"Bearer {self.admin_token}"})
                    self.log_test("Admin Authentication", True, "Successfully authenticated as admin")
                    return True
            
            self.log_test(
                "Admin Authentication", 
                False, 
                f"Status: {response.status_code}",
                response.text
            )
            return False
            
        except Exception as e:
            self.log_test("Admin Authentication", False, f"Exception: {str(e)}")
            return False

    def test_database_setup(self) -> bool:
        """Test if content management tables are created and have sample data"""
        try:
            # Test if we can access email templates (should have sample data)
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/email-templates")
            
            if response.status_code == 200:
                data = response.json()
                if data.get("success") and "templates" in data:
                    templates = data["templates"]
                    success = len(templates) > 0
                    details = f"Found {len(templates)} email templates in database"
                    if success:
                        details += " - Sample data exists"
                    else:
                        details += " - No sample data found"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                success = False
                details = f"Status: {response.status_code}"
                
            self.log_test(
                "Database Setup - Content Tables & Sample Data",
                success,
                details,
                data if response.status_code == 200 else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Database Setup - Content Tables & Sample Data", False, f"Exception: {str(e)}")
            return False

    # ========================= BANNER MANAGEMENT TESTS =========================

    def test_banner_management(self) -> bool:
        """Test complete banner management functionality"""
        all_tests_passed = True
        
        # Test 1: Create Banner
        banner_data = {
            "title": "Test Promotional Banner",
            "description": "This is a test banner for CMS testing",
            "image_url": "https://example.com/banner.jpg",
            "link_url": "https://example.com/promotion",
            "position": "top",
            "type": "promotion",
            "priority": 80,
            "start_date": datetime.now().isoformat(),
            "end_date": (datetime.now() + timedelta(days=30)).isoformat(),
            "target_roles": {"users": ["all"]},
            "metadata": {"campaign": "test_campaign"}
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/banners", json=banner_data)
            success = response.status_code == 201
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    banner_id = data["data"]["id"]
                    self.created_resources['banners'].append(banner_id)
                    details = f"Banner created successfully with ID: {banner_id}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Banner Management - Create Banner", success, details, 
                         response.json() if response.status_code == 201 else response.text)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Banner Management - Create Banner", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: List Banners with Filters
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/banners?position=top&type=promotion")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "banners" in data:
                    banners = data["banners"]
                    details = f"Found {len(banners)} banners with filters"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Banner Management - List with Filters", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Banner Management - List with Filters", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 3: Get Banner Details
        if self.created_resources['banners']:
            try:
                banner_id = self.created_resources['banners'][0]
                response = self.session.get(f"{self.base_url}/api/v1/admin/content/banners/{banner_id}")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success") and "data" in data:
                        banner = data["data"]
                        details = f"Retrieved banner details: {banner.get('title', 'Unknown')}"
                    else:
                        success = False
                        details = "Invalid response structure"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Banner Management - Get Banner Details", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Banner Management - Get Banner Details", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 4: Update Banner
        if self.created_resources['banners']:
            try:
                banner_id = self.created_resources['banners'][0]
                update_data = banner_data.copy()
                update_data["title"] = "Updated Test Banner"
                update_data["priority"] = 90
                
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/banners/{banner_id}", json=update_data)
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "Banner updated successfully"
                    else:
                        success = False
                        details = "Update failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Banner Management - Update Banner", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Banner Management - Update Banner", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 5: Toggle Banner Status
        if self.created_resources['banners']:
            try:
                banner_id = self.created_resources['banners'][0]
                response = self.session.patch(f"{self.base_url}/api/v1/admin/content/banners/{banner_id}/toggle")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "Banner status toggled successfully"
                    else:
                        success = False
                        details = "Toggle failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Banner Management - Toggle Status", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Banner Management - Toggle Status", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    def test_public_banner_endpoints(self) -> bool:
        """Test public banner endpoints (no auth required)"""
        all_tests_passed = True
        
        # Test 1: Get Active Banners
        try:
            # Remove auth header temporarily for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/banners/active")
            
            # Restore headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    banners = data["data"]
                    details = f"Found {len(banners)} active banners"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Public Banners - Get Active Banners", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Public Banners - Get Active Banners", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: Track Banner Click
        if self.created_resources['banners']:
            try:
                banner_id = self.created_resources['banners'][0]
                
                # Remove auth header temporarily for public endpoint
                original_headers = self.session.headers.copy()
                if 'Authorization' in self.session.headers:
                    del self.session.headers['Authorization']
                
                response = self.session.post(f"{self.base_url}/api/v1/banners/{banner_id}/click")
                
                # Restore headers
                self.session.headers.clear()
                self.session.headers.update(original_headers)
                
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "Banner click tracked successfully"
                    else:
                        success = False
                        details = "Click tracking failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Public Banners - Track Click", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Public Banners - Track Click", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    # ========================= EMAIL TEMPLATE TESTS =========================

    def test_email_template_management(self) -> bool:
        """Test email template management functionality"""
        all_tests_passed = True
        
        # Test 1: Create Email Template
        template_data = {
            "name": "Test Welcome Template",
            "description": "Test template for new user welcome",
            "subject": "Welcome to Fantasy Esports - {{name}}!",
            "html_content": "<h1>Welcome {{name}}!</h1><p>Thanks for joining our platform. Your email is {{email}}.</p>",
            "text_content": "Welcome {{name}}! Thanks for joining our platform. Your email is {{email}}.",
            "category": "welcome",
            "variables": {"name": "User Name", "email": "User Email"}
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/email-templates", json=template_data)
            success = response.status_code == 201
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    template_id = data["data"]["id"]
                    self.created_resources['email_templates'].append(template_id)
                    details = f"Email template created successfully with ID: {template_id}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Email Templates - Create Template", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Email Templates - Create Template", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: List Email Templates with Filters
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/email-templates?category=welcome&active=true")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "templates" in data:
                    templates = data["templates"]
                    details = f"Found {len(templates)} welcome templates"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Email Templates - List with Filters", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Email Templates - List with Filters", False, f"Exception: {str(e)}")
            all_tests_passed = False

        return all_tests_passed

    # ========================= MARKETING CAMPAIGN TESTS =========================

    def test_marketing_campaign_management(self) -> bool:
        """Test marketing campaign management functionality"""
        all_tests_passed = True
        
        # Test 1: Create Marketing Campaign
        campaign_data = {
            "name": "Test Deposit Bonus Campaign",
            "subject": "Get 100% Bonus on Your First Deposit!",
            "email_template": "deposit_bonus_template",
            "target_segment": "new_users",
            "target_criteria": {"registration_days": 7},
            "scheduled_at": (datetime.now() + timedelta(hours=1)).isoformat()
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/campaigns", json=campaign_data)
            success = response.status_code == 201
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    campaign_id = data["data"]["id"]
                    self.created_resources['campaigns'].append(campaign_id)
                    details = f"Marketing campaign created successfully with ID: {campaign_id}"
                    if "estimated_recipients" in data:
                        details += f", Estimated recipients: {data['estimated_recipients']}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Marketing Campaigns - Create Campaign", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Marketing Campaigns - Create Campaign", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: List Marketing Campaigns
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/campaigns?status=draft")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "campaigns" in data:
                    campaigns = data["campaigns"]
                    details = f"Found {len(campaigns)} draft campaigns"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Marketing Campaigns - List Campaigns", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Marketing Campaigns - List Campaigns", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 3: Update Campaign Status
        if self.created_resources['campaigns']:
            try:
                campaign_id = self.created_resources['campaigns'][0]
                status_data = {"status": "scheduled"}
                
                response = self.session.patch(f"{self.base_url}/api/v1/admin/content/campaigns/{campaign_id}/status", json=status_data)
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "Campaign status updated to scheduled"
                    else:
                        success = False
                        details = "Status update failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Marketing Campaigns - Update Status", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Marketing Campaigns - Update Status", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    # ========================= SEO CONTENT TESTS =========================

    def test_seo_content_management(self) -> bool:
        """Test SEO content management functionality"""
        all_tests_passed = True
        
        # Test 1: Create SEO Content
        seo_data = {
            "page_type": "tournament",
            "page_slug": "test-tournament-seo",
            "meta_title": "Test Tournament - Fantasy Esports",
            "meta_description": "Join the ultimate test tournament on our fantasy esports platform. Create teams and win big!",
            "keywords": ["tournament", "fantasy", "esports", "test"],
            "og_title": "Test Tournament - Fantasy Esports",
            "og_description": "Join the ultimate test tournament on our fantasy esports platform.",
            "og_image": "https://example.com/tournament-og.jpg",
            "twitter_card": "summary_large_image",
            "structured_data": {"@type": "Tournament", "name": "Test Tournament"},
            "content": "<h1>Test Tournament</h1><p>This is a test tournament page.</p>"
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=seo_data)
            success = response.status_code == 201
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    seo_id = data["data"]["id"]
                    self.created_resources['seo_content'].append(seo_id)
                    details = f"SEO content created successfully with ID: {seo_id}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("SEO Content - Create SEO Content", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("SEO Content - Create SEO Content", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: List SEO Content
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/seo?page_type=tournament")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "contents" in data:
                    contents = data["contents"]
                    details = f"Found {len(contents)} tournament SEO contents"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("SEO Content - List SEO Content", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("SEO Content - List SEO Content", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 3: Get SEO Content by ID
        if self.created_resources['seo_content']:
            try:
                seo_id = self.created_resources['seo_content'][0]
                response = self.session.get(f"{self.base_url}/api/v1/admin/content/seo/{seo_id}")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success") and "data" in data:
                        seo_content = data["data"]
                        details = f"Retrieved SEO content: {seo_content.get('meta_title', 'Unknown')}"
                    else:
                        success = False
                        details = "Invalid response structure"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("SEO Content - Get by ID", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("SEO Content - Get by ID", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 4: Update SEO Content
        if self.created_resources['seo_content']:
            try:
                seo_id = self.created_resources['seo_content'][0]
                update_data = seo_data.copy()
                update_data["meta_title"] = "Updated Test Tournament - Fantasy Esports"
                
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/seo/{seo_id}", json=update_data)
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "SEO content updated successfully"
                    else:
                        success = False
                        details = "Update failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("SEO Content - Update SEO Content", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("SEO Content - Update SEO Content", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 5: Delete SEO Content
        if self.created_resources['seo_content']:
            try:
                seo_id = self.created_resources['seo_content'][0]
                response = self.session.delete(f"{self.base_url}/api/v1/admin/content/seo/{seo_id}")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "SEO content deleted successfully"
                        # Remove from created resources since it's deleted
                        self.created_resources['seo_content'].remove(seo_id)
                    else:
                        success = False
                        details = "Delete failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("SEO Content - Delete SEO Content", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("SEO Content - Delete SEO Content", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    def test_public_seo_endpoints(self) -> bool:
        """Test public SEO endpoints (no auth required)"""
        all_tests_passed = True
        
        # Test: Get SEO Content by Slug (using existing sample data)
        try:
            # Remove auth header temporarily for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/seo/home")
            
            # Restore headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    seo_content = data["data"]
                    details = f"Retrieved SEO content by slug: {seo_content.get('meta_title', 'Unknown')}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Public SEO - Get by Slug", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Public SEO - Get by Slug", False, f"Exception: {str(e)}")
            all_tests_passed = False

        return all_tests_passed

    # ========================= FAQ MANAGEMENT TESTS =========================

    def test_faq_management(self) -> bool:
        """Test FAQ management functionality"""
        all_tests_passed = True
        
        # Test 1: Create FAQ Section
        section_data = {
            "name": "Test FAQ Section",
            "description": "This is a test FAQ section for CMS testing",
            "sort_order": 10
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/sections", json=section_data)
            success = response.status_code == 201
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    section_id = data["data"]["id"]
                    self.created_resources['faq_sections'].append(section_id)
                    details = f"FAQ section created successfully with ID: {section_id}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("FAQ Management - Create Section", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("FAQ Management - Create Section", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: Update FAQ Section
        if self.created_resources['faq_sections']:
            try:
                section_id = self.created_resources['faq_sections'][0]
                update_data = section_data.copy()
                update_data["name"] = "Updated Test FAQ Section"
                update_data["sort_order"] = 15
                
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/sections/{section_id}", json=update_data)
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "FAQ section updated successfully"
                    else:
                        success = False
                        details = "Update failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("FAQ Management - Update Section", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("FAQ Management - Update Section", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 3: Create FAQ Item
        if self.created_resources['faq_sections']:
            item_data = {
                "section_id": self.created_resources['faq_sections'][0],
                "question": "How do I test the CMS system?",
                "answer": "You can test the CMS system by running comprehensive tests on all endpoints including banners, email templates, campaigns, SEO content, FAQ management, and legal documents.",
                "sort_order": 1,
                "tags": ["testing", "cms", "help"]
            }
            
            try:
                response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/items", json=item_data)
                success = response.status_code == 201
                
                if success:
                    data = response.json()
                    if data.get("success") and "data" in data:
                        item_id = data["data"]["id"]
                        self.created_resources['faq_items'].append(item_id)
                        details = f"FAQ item created successfully with ID: {item_id}"
                    else:
                        success = False
                        details = "Invalid response structure"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("FAQ Management - Create Item", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("FAQ Management - Create Item", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 4: Update FAQ Item
        if self.created_resources['faq_items']:
            try:
                item_id = self.created_resources['faq_items'][0]
                update_data = {
                    "section_id": self.created_resources['faq_sections'][0],
                    "question": "How do I test the updated CMS system?",
                    "answer": "You can test the updated CMS system by running comprehensive tests on all endpoints and verifying that updates work correctly.",
                    "sort_order": 2,
                    "tags": ["testing", "cms", "help", "updated"]
                }
                
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/faq/items/{item_id}", json=update_data)
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "FAQ item updated successfully"
                    else:
                        success = False
                        details = "Update failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("FAQ Management - Update Item", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("FAQ Management - Update Item", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    def test_public_faq_endpoints(self) -> bool:
        """Test public FAQ endpoints (no auth required)"""
        all_tests_passed = True
        
        # Remove auth header temporarily for public endpoints
        original_headers = self.session.headers.copy()
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']
        
        # Test 1: List FAQ Sections (Public)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/faq/sections")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "sections" in data:
                    sections = data["sections"]
                    details = f"Found {len(sections)} public FAQ sections"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Public FAQ - List Sections", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Public FAQ - List Sections", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: List FAQ Items (Public)
        try:
            response = self.session.get(f"{self.base_url}/api/v1/faq/items")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "items" in data:
                    items = data["items"]
                    details = f"Found {len(items)} public FAQ items"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Public FAQ - List Items", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Public FAQ - List Items", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 3: Track FAQ View
        if self.created_resources['faq_items']:
            try:
                item_id = self.created_resources['faq_items'][0]
                response = self.session.post(f"{self.base_url}/api/v1/faq/items/{item_id}/view")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "FAQ view tracked successfully"
                    else:
                        success = False
                        details = "View tracking failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Public FAQ - Track View", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Public FAQ - Track View", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 4: Track FAQ Like
        if self.created_resources['faq_items']:
            try:
                item_id = self.created_resources['faq_items'][0]
                response = self.session.post(f"{self.base_url}/api/v1/faq/items/{item_id}/like")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "FAQ like tracked successfully"
                    else:
                        success = False
                        details = "Like tracking failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Public FAQ - Track Like", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Public FAQ - Track Like", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Restore headers
        self.session.headers.clear()
        self.session.headers.update(original_headers)

        return all_tests_passed

    # ========================= LEGAL DOCUMENT TESTS =========================

    def test_legal_document_management(self) -> bool:
        """Test legal document management functionality"""
        all_tests_passed = True
        
        # Test 1: Create Legal Document
        legal_data = {
            "document_type": "cookie",
            "title": "Test Cookie Policy",
            "content": "This is a test cookie policy for the CMS testing. We use cookies to enhance your experience on our platform...",
            "version": "1.1",
            "effective_date": datetime.now().isoformat(),
            "metadata": {"test": True, "version_notes": "Test version for CMS"}
        }
        
        try:
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/legal", json=legal_data)
            success = response.status_code == 201
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    legal_id = data["data"]["id"]
                    self.created_resources['legal_documents'].append(legal_id)
                    details = f"Legal document created successfully with ID: {legal_id}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Legal Documents - Create Document", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Legal Documents - Create Document", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: List Legal Documents
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/legal?type=cookie")
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "documents" in data:
                    documents = data["documents"]
                    details = f"Found {len(documents)} cookie documents"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Legal Documents - List Documents", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Legal Documents - List Documents", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 3: Update Legal Document
        if self.created_resources['legal_documents']:
            try:
                legal_id = self.created_resources['legal_documents'][0]
                update_data = legal_data.copy()
                update_data["title"] = "Updated Test Cookie Policy"
                update_data["version"] = "1.2"
                
                response = self.session.put(f"{self.base_url}/api/v1/admin/content/legal/{legal_id}", json=update_data)
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "Legal document updated successfully"
                    else:
                        success = False
                        details = "Update failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Legal Documents - Update Document", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Legal Documents - Update Document", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 4: Publish Legal Document
        if self.created_resources['legal_documents']:
            try:
                legal_id = self.created_resources['legal_documents'][0]
                response = self.session.patch(f"{self.base_url}/api/v1/admin/content/legal/{legal_id}/publish")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success"):
                        details = "Legal document published successfully"
                    else:
                        success = False
                        details = "Publish failed"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Legal Documents - Publish Document", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Legal Documents - Publish Document", False, f"Exception: {str(e)}")
                all_tests_passed = False

        # Test 5: Delete Legal Document (should fail for published documents)
        if self.created_resources['legal_documents']:
            try:
                legal_id = self.created_resources['legal_documents'][0]
                response = self.session.delete(f"{self.base_url}/api/v1/admin/content/legal/{legal_id}")
                
                # This should fail because we published the document
                success = response.status_code == 404 or response.status_code == 400
                
                if success:
                    details = "Correctly prevented deletion of published document"
                else:
                    details = f"Status: {response.status_code} - Should not allow deletion of published documents"
                    
                self.log_test("Legal Documents - Delete Published Document", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Legal Documents - Delete Published Document", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    def test_public_legal_endpoints(self) -> bool:
        """Test public legal document endpoints (no auth required)"""
        all_tests_passed = True
        
        # Test: Get Active Legal Document (using existing sample data)
        try:
            # Remove auth header temporarily for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/legal/terms")
            
            # Restore headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    legal_doc = data["data"]
                    details = f"Retrieved active legal document: {legal_doc.get('title', 'Unknown')}"
                else:
                    success = False
                    details = "Invalid response structure"
            else:
                details = f"Status: {response.status_code}"
                
            self.log_test("Public Legal - Get Active Document", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Public Legal - Get Active Document", False, f"Exception: {str(e)}")
            all_tests_passed = False

        return all_tests_passed

    # ========================= CONTENT ANALYTICS TESTS =========================

    def test_content_analytics(self) -> bool:
        """Test content analytics functionality"""
        all_tests_passed = True
        
        # Test: Get Content Analytics
        if self.created_resources['banners']:
            try:
                banner_id = self.created_resources['banners'][0]
                response = self.session.get(f"{self.base_url}/api/v1/admin/content/analytics/banner/{banner_id}?days=7")
                success = response.status_code == 200
                
                if success:
                    data = response.json()
                    if data.get("success") and "data" in data:
                        analytics = data["data"]
                        details = f"Retrieved analytics for banner {banner_id}: {len(analytics)} data points"
                    else:
                        success = False
                        details = "Invalid response structure"
                else:
                    details = f"Status: {response.status_code}"
                    
                self.log_test("Content Analytics - Get Banner Analytics", success, details)
                all_tests_passed &= success
                
            except Exception as e:
                self.log_test("Content Analytics - Get Banner Analytics", False, f"Exception: {str(e)}")
                all_tests_passed = False

        return all_tests_passed

    # ========================= VALIDATION TESTS =========================

    def test_request_validation(self) -> bool:
        """Test request validation for all endpoints"""
        all_tests_passed = True
        
        # Test 1: Banner validation - Invalid position
        try:
            invalid_banner = {
                "title": "Test Banner",
                "image_url": "https://example.com/banner.jpg",
                "position": "invalid_position",  # Invalid
                "type": "promotion",
                "priority": 50,
                "start_date": datetime.now().isoformat(),
                "end_date": (datetime.now() + timedelta(days=30)).isoformat()
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/banners", json=invalid_banner)
            success = response.status_code == 400
            
            details = f"Status: {response.status_code} - Correctly rejected invalid position"
            self.log_test("Validation - Invalid Banner Position", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Validation - Invalid Banner Position", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 2: Email template validation - Invalid category
        try:
            invalid_template = {
                "name": "Test Template",
                "subject": "Test Subject",
                "html_content": "<p>Test</p>",
                "category": "invalid_category"  # Invalid
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/email-templates", json=invalid_template)
            success = response.status_code == 400
            
            details = f"Status: {response.status_code} - Correctly rejected invalid category"
            self.log_test("Validation - Invalid Template Category", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Validation - Invalid Template Category", False, f"Exception: {str(e)}")
            all_tests_passed = False

        # Test 3: Legal document validation - Invalid document type
        try:
            invalid_legal = {
                "document_type": "invalid_type",  # Invalid
                "title": "Test Document",
                "content": "Test content",
                "version": "1.0",
                "effective_date": datetime.now().isoformat()
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/legal", json=invalid_legal)
            success = response.status_code == 400
            
            details = f"Status: {response.status_code} - Correctly rejected invalid document type"
            self.log_test("Validation - Invalid Legal Document Type", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Validation - Invalid Legal Document Type", False, f"Exception: {str(e)}")
            all_tests_passed = False

        return all_tests_passed

    def test_authorization(self) -> bool:
        """Test authorization for admin endpoints"""
        all_tests_passed = True
        
        # Test: Access admin endpoint without auth
        try:
            # Remove authorization header temporarily
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/banners")
            
            # Restore headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 401
            details = f"Status: {response.status_code} - Correctly rejected unauthorized access"
            
            self.log_test("Authorization - Admin Endpoint Without Auth", success, details)
            all_tests_passed &= success
            
        except Exception as e:
            self.log_test("Authorization - Admin Endpoint Without Auth", False, f"Exception: {str(e)}")
            all_tests_passed = False

        return all_tests_passed

    # ========================= CLEANUP =========================

    def cleanup_test_data(self):
        """Clean up test data created during testing"""
        print("\n🧹 Cleaning up test data...")
        
        # Delete created banners
        for banner_id in self.created_resources['banners']:
            try:
                response = self.session.delete(f"{self.base_url}/api/v1/admin/content/banners/{banner_id}")
                if response.status_code == 200:
                    print(f"   ✅ Deleted banner {banner_id}")
                else:
                    print(f"   ⚠️  Failed to delete banner {banner_id}: {response.status_code}")
            except Exception as e:
                print(f"   ❌ Error deleting banner {banner_id}: {str(e)}")

        # Note: SEO content was already deleted during testing
        # Note: Legal documents cannot be deleted once published (by design)
        
        print("🧹 Cleanup completed\n")

    # ========================= MAIN TEST RUNNER =========================

    def run_comprehensive_tests(self):
        """Run all content management system tests"""
        print("🚀 Starting Comprehensive Content Management System Testing")
        print("=" * 80)
        
        # Test 1: Health Check
        if not self.test_health_check():
            print("❌ Backend is not healthy. Stopping tests.")
            return
        
        # Test 2: Admin Authentication
        if not self.authenticate_admin():
            print("❌ Admin authentication failed. Cannot test admin endpoints.")
            return
        
        # Test 3: Database Setup
        self.test_database_setup()
        
        # Test 4: Banner Management
        print("\n📋 Testing Banner Management...")
        self.test_banner_management()
        self.test_public_banner_endpoints()
        
        # Test 5: Email Template Management
        print("\n📧 Testing Email Template Management...")
        self.test_email_template_management()
        
        # Test 6: Marketing Campaign Management
        print("\n📢 Testing Marketing Campaign Management...")
        self.test_marketing_campaign_management()
        
        # Test 7: SEO Content Management
        print("\n🔍 Testing SEO Content Management...")
        self.test_seo_content_management()
        self.test_public_seo_endpoints()
        
        # Test 8: FAQ Management
        print("\n❓ Testing FAQ Management...")
        self.test_faq_management()
        self.test_public_faq_endpoints()
        
        # Test 9: Legal Document Management
        print("\n📜 Testing Legal Document Management...")
        self.test_legal_document_management()
        self.test_public_legal_endpoints()
        
        # Test 10: Content Analytics
        print("\n📊 Testing Content Analytics...")
        self.test_content_analytics()
        
        # Test 11: Validation & Authorization
        print("\n🔒 Testing Validation & Authorization...")
        self.test_request_validation()
        self.test_authorization()
        
        # Cleanup
        self.cleanup_test_data()
        
        # Generate Summary
        self.generate_summary()

    def generate_summary(self):
        """Generate comprehensive test summary"""
        print("\n" + "=" * 80)
        print("📊 CONTENT MANAGEMENT SYSTEM TEST SUMMARY")
        print("=" * 80)
        
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result["success"])
        failed_tests = total_tests - passed_tests
        success_rate = (passed_tests / total_tests * 100) if total_tests > 0 else 0
        
        print(f"Total Tests: {total_tests}")
        print(f"Passed: {passed_tests} ✅")
        print(f"Failed: {failed_tests} ❌")
        print(f"Success Rate: {success_rate:.1f}%")
        print()
        
        # Categorize results
        categories = {
            "Health & Authentication": [],
            "Database Setup": [],
            "Banner Management": [],
            "Email Templates": [],
            "Marketing Campaigns": [],
            "SEO Content": [],
            "FAQ Management": [],
            "Legal Documents": [],
            "Content Analytics": [],
            "Validation & Authorization": []
        }
        
        for result in self.test_results:
            test_name = result["test"]
            if "Health" in test_name or "Authentication" in test_name:
                categories["Health & Authentication"].append(result)
            elif "Database" in test_name:
                categories["Database Setup"].append(result)
            elif "Banner" in test_name:
                categories["Banner Management"].append(result)
            elif "Email" in test_name or "Template" in test_name:
                categories["Email Templates"].append(result)
            elif "Campaign" in test_name:
                categories["Marketing Campaigns"].append(result)
            elif "SEO" in test_name:
                categories["SEO Content"].append(result)
            elif "FAQ" in test_name:
                categories["FAQ Management"].append(result)
            elif "Legal" in test_name:
                categories["Legal Documents"].append(result)
            elif "Analytics" in test_name:
                categories["Content Analytics"].append(result)
            elif "Validation" in test_name or "Authorization" in test_name:
                categories["Validation & Authorization"].append(result)
        
        for category, results in categories.items():
            if results:
                passed = sum(1 for r in results if r["success"])
                total = len(results)
                print(f"{category}: {passed}/{total} passed")
        
        print("\n" + "=" * 80)
        print("🔍 DETAILED FINDINGS")
        print("=" * 80)
        
        # Show failed tests
        failed_results = [r for r in self.test_results if not r["success"]]
        if failed_results:
            print("❌ FAILED TESTS:")
            for result in failed_results:
                print(f"  • {result['test']}: {result['details']}")
        else:
            print("✅ ALL TESTS PASSED!")
        
        print("\n" + "=" * 80)
        print("🎯 CONTENT MANAGEMENT SYSTEM STATUS")
        print("=" * 80)
        
        # Overall assessment
        if success_rate >= 95:
            print("🎉 EXCELLENT: Content Management System is working excellently!")
        elif success_rate >= 85:
            print("✅ GOOD: Content Management System is working well with minor issues.")
        elif success_rate >= 70:
            print("⚠️  MODERATE: Content Management System has some issues that need attention.")
        else:
            print("❌ CRITICAL: Content Management System has significant issues requiring immediate attention.")
        
        # Feature assessment
        admin_tests = [r for r in self.test_results if "admin" in r["test"].lower() or any(cat in r["test"] for cat in ["Banner", "Email", "Campaign", "SEO", "FAQ", "Legal"])]
        public_tests = [r for r in self.test_results if "Public" in r["test"]]
        
        admin_success = sum(1 for r in admin_tests if r["success"]) / len(admin_tests) * 100 if admin_tests else 0
        public_success = sum(1 for r in public_tests if r["success"]) / len(public_tests) * 100 if public_tests else 0
        
        print(f"\nAdmin Content Management: {admin_success:.1f}% functional")
        print(f"Public Content APIs: {public_success:.1f}% functional")
        
        if admin_success >= 85 and public_success >= 85:
            print("\n🚀 READY FOR PRODUCTION: Content Management System is working!")
        else:
            print("\n⚠️  NEEDS WORK: Content Management System requires fixes before production.")
        
        # Key features summary
        print(f"\n📋 KEY FEATURES TESTED:")
        print(f"✅ Banner Management (Admin & Public)")
        print(f"✅ Email Template Management")
        print(f"✅ Marketing Campaign Management")
        print(f"✅ SEO Content Management (Admin & Public)")
        print(f"✅ FAQ Management (Admin & Public)")
        print(f"✅ Legal Document Management (Admin & Public)")
        print(f"✅ Content Analytics")
        print(f"✅ Request Validation")
        print(f"✅ Authorization & Authentication")

if __name__ == "__main__":
    tester = ContentManagementTester()
    tester.run_comprehensive_tests()