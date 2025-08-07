#!/usr/bin/env python3
"""
Comprehensive Content Management System Testing for GoLang Fantasy Esports Backend
Testing CMS functionality including banners, email templates, marketing campaigns, 
SEO content, FAQ management, legal documents, and analytics tracking
"""

import requests
import json
import time
import uuid
from typing import Dict, Any, Optional, Tuple, List

class ContentManagementTester:
    def __init__(self, base_url: str = "http://localhost:8001"):
        self.base_url = base_url
        self.session = requests.Session()
        self.admin_token = None
        self.user_token = None
        self.test_results = []
        self.created_resources = {
            "banners": [],
            "email_templates": [],
            "campaigns": [],
            "seo_content": [],
            "faq_sections": [],
            "faq_items": [],
            "legal_documents": []
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
            # Try to authenticate as admin
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

    def test_database_setup_verification(self) -> bool:
        """Test if CMS database tables were created successfully"""
        try:
            # Test by trying to access admin endpoints that would fail if tables don't exist
            endpoints_to_test = [
                "/api/v1/admin/content/banners",
                "/api/v1/admin/content/email-templates",
                "/api/v1/admin/content/campaigns",
                "/api/v1/admin/content/seo",
                "/api/v1/admin/content/faq/sections",
                "/api/v1/admin/content/legal"
            ]
            
            successful_endpoints = 0
            total_endpoints = len(endpoints_to_test)
            
            for endpoint in endpoints_to_test:
                try:
                    response = self.session.get(f"{self.base_url}{endpoint}")
                    # 200 (success) or 401 (auth required) means endpoint exists
                    # 404 means endpoint/table doesn't exist
                    if response.status_code in [200, 401]:
                        successful_endpoints += 1
                except:
                    pass
            
            success = successful_endpoints == total_endpoints
            details = f"Found {successful_endpoints}/{total_endpoints} CMS endpoints accessible"
            
            if success:
                details += " - Database tables appear to be created successfully"
            else:
                details += " - Some CMS endpoints are missing (possible database setup issue)"
            
            self.log_test("Database Setup Verification", success, details)
            return success
            
        except Exception as e:
            self.log_test("Database Setup Verification", False, f"Exception: {str(e)}")
            return False

    # ADMIN BANNER MANAGEMENT TESTS
    def test_admin_banner_create(self) -> bool:
        """Test admin banner creation"""
        if not self.admin_token:
            self.log_test("Admin Banner Create", False, "No admin token available")
            return False
            
        try:
            banner_data = {
                "title": "Welcome to Fantasy Esports",
                "content": "Join the ultimate fantasy esports experience!",
                "banner_type": "promotional",
                "position": "header",
                "priority": 1,
                "is_active": True,
                "start_date": "2025-01-01T00:00:00Z",
                "end_date": "2025-12-31T23:59:59Z",
                "target_audience": "all_users",
                "click_url": "https://fantasy-esports.com/signup",
                "image_url": "https://example.com/banner.jpg"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/banners", json=banner_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    banner_id = data["data"].get("id")
                    self.created_resources["banners"].append(banner_id)
                    details += f" - Banner created successfully with ID: {banner_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin Banner Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Banner Create", False, f"Exception: {str(e)}")
            return False

    def test_admin_banner_list(self) -> bool:
        """Test admin banner listing with filters"""
        if not self.admin_token:
            self.log_test("Admin Banner List", False, "No admin token available")
            return False
            
        try:
            # Test basic listing
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/banners")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    banners = data.get("data", [])
                    details += f" - Found {len(banners)} banners"
                    
                    # Test with filters
                    filter_response = self.session.get(
                        f"{self.base_url}/api/v1/admin/content/banners?banner_type=promotional&is_active=true"
                    )
                    if filter_response.status_code == 200:
                        filter_data = filter_response.json()
                        filtered_banners = filter_data.get("data", [])
                        details += f", {len(filtered_banners)} promotional active banners"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Admin Banner List",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Banner List", False, f"Exception: {str(e)}")
            return False

    def test_admin_banner_update(self) -> bool:
        """Test admin banner update"""
        if not self.admin_token or not self.created_resources["banners"]:
            self.log_test("Admin Banner Update", False, "No admin token or banner ID available")
            return False
            
        try:
            banner_id = self.created_resources["banners"][0]
            update_data = {
                "title": "Updated Fantasy Esports Banner",
                "content": "Updated content for the banner",
                "priority": 2,
                "is_active": False
            }
            
            response = self.session.put(
                f"{self.base_url}/api/v1/admin/content/banners/{banner_id}", 
                json=update_data
            )
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Banner updated successfully"
                else:
                    success = False
                    details += " - Update failed"
            
            self.log_test(
                "Admin Banner Update",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Banner Update", False, f"Exception: {str(e)}")
            return False

    def test_admin_banner_toggle(self) -> bool:
        """Test admin banner status toggle"""
        if not self.admin_token or not self.created_resources["banners"]:
            self.log_test("Admin Banner Toggle", False, "No admin token or banner ID available")
            return False
            
        try:
            banner_id = self.created_resources["banners"][0]
            
            response = self.session.patch(f"{self.base_url}/api/v1/admin/content/banners/{banner_id}/toggle")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    new_status = data.get("data", {}).get("is_active")
                    details += f" - Banner status toggled to: {new_status}"
                else:
                    success = False
                    details += " - Toggle failed"
            
            self.log_test(
                "Admin Banner Toggle",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Banner Toggle", False, f"Exception: {str(e)}")
            return False

    # EMAIL TEMPLATE TESTS
    def test_admin_email_template_create(self) -> bool:
        """Test admin email template creation"""
        if not self.admin_token:
            self.log_test("Admin Email Template Create", False, "No admin token available")
            return False
            
        try:
            template_data = {
                "name": "Welcome Email",
                "subject": "Welcome to Fantasy Esports!",
                "template_type": "welcome",
                "html_content": "<h1>Welcome {{.FirstName}}!</h1><p>Thanks for joining Fantasy Esports.</p>",
                "text_content": "Welcome {{.FirstName}}! Thanks for joining Fantasy Esports.",
                "variables": ["FirstName", "Email"],
                "is_active": True,
                "description": "Welcome email template for new users"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/email-templates", json=template_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    template_id = data["data"].get("id")
                    self.created_resources["email_templates"].append(template_id)
                    details += f" - Email template created successfully with ID: {template_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin Email Template Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Email Template Create", False, f"Exception: {str(e)}")
            return False

    def test_admin_email_template_list(self) -> bool:
        """Test admin email template listing"""
        if not self.admin_token:
            self.log_test("Admin Email Template List", False, "No admin token available")
            return False
            
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/email-templates")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    templates = data.get("data", [])
                    details += f" - Found {len(templates)} email templates"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Admin Email Template List",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Email Template List", False, f"Exception: {str(e)}")
            return False

    # MARKETING CAMPAIGN TESTS
    def test_admin_campaign_create(self) -> bool:
        """Test admin marketing campaign creation"""
        if not self.admin_token:
            self.log_test("Admin Campaign Create", False, "No admin token available")
            return False
            
        try:
            campaign_data = {
                "name": "New Year Promotion 2025",
                "campaign_type": "promotional",
                "status": "draft",
                "start_date": "2025-01-01T00:00:00Z",
                "end_date": "2025-01-31T23:59:59Z",
                "target_audience": "active_users",
                "budget": 10000.0,
                "description": "New Year promotional campaign for 2025",
                "channels": ["email", "push", "banner"],
                "goals": ["increase_engagement", "boost_deposits"]
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/campaigns", json=campaign_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    campaign_id = data["data"].get("id")
                    self.created_resources["campaigns"].append(campaign_id)
                    details += f" - Campaign created successfully with ID: {campaign_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin Campaign Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Campaign Create", False, f"Exception: {str(e)}")
            return False

    def test_admin_campaign_status_update(self) -> bool:
        """Test admin campaign status update"""
        if not self.admin_token or not self.created_resources["campaigns"]:
            self.log_test("Admin Campaign Status Update", False, "No admin token or campaign ID available")
            return False
            
        try:
            campaign_id = self.created_resources["campaigns"][0]
            
            response = self.session.patch(
                f"{self.base_url}/api/v1/admin/content/campaigns/{campaign_id}/status",
                json={"status": "active"}
            )
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    new_status = data.get("data", {}).get("status")
                    details += f" - Campaign status updated to: {new_status}"
                else:
                    success = False
                    details += " - Status update failed"
            
            self.log_test(
                "Admin Campaign Status Update",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Campaign Status Update", False, f"Exception: {str(e)}")
            return False

    # SEO CONTENT TESTS
    def test_admin_seo_create(self) -> bool:
        """Test admin SEO content creation"""
        if not self.admin_token:
            self.log_test("Admin SEO Create", False, "No admin token available")
            return False
            
        try:
            seo_data = {
                "page_slug": "home",
                "title": "Fantasy Esports - Ultimate Gaming Experience",
                "meta_description": "Join the ultimate fantasy esports platform. Create teams, compete in tournaments, and win real money.",
                "meta_keywords": ["fantasy esports", "gaming", "tournaments", "esports betting"],
                "og_title": "Fantasy Esports - Ultimate Gaming Experience",
                "og_description": "Join the ultimate fantasy esports platform",
                "og_image": "https://example.com/og-image.jpg",
                "canonical_url": "https://fantasy-esports.com/",
                "schema_markup": {"@type": "WebSite", "name": "Fantasy Esports"},
                "is_active": True
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=seo_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    seo_id = data["data"].get("id")
                    self.created_resources["seo_content"].append(seo_id)
                    details += f" - SEO content created successfully with ID: {seo_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin SEO Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin SEO Create", False, f"Exception: {str(e)}")
            return False

    def test_admin_seo_list(self) -> bool:
        """Test admin SEO content listing"""
        if not self.admin_token:
            self.log_test("Admin SEO List", False, "No admin token available")
            return False
            
        try:
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/seo")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    seo_items = data.get("data", [])
                    details += f" - Found {len(seo_items)} SEO content items"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Admin SEO List",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin SEO List", False, f"Exception: {str(e)}")
            return False

    # FAQ MANAGEMENT TESTS
    def test_admin_faq_section_create(self) -> bool:
        """Test admin FAQ section creation"""
        if not self.admin_token:
            self.log_test("Admin FAQ Section Create", False, "No admin token available")
            return False
            
        try:
            section_data = {
                "title": "Getting Started",
                "description": "Basic questions about using Fantasy Esports",
                "display_order": 1,
                "is_active": True,
                "icon": "question-circle"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/sections", json=section_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    section_id = data["data"].get("id")
                    self.created_resources["faq_sections"].append(section_id)
                    details += f" - FAQ section created successfully with ID: {section_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin FAQ Section Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin FAQ Section Create", False, f"Exception: {str(e)}")
            return False

    def test_admin_faq_item_create(self) -> bool:
        """Test admin FAQ item creation"""
        if not self.admin_token or not self.created_resources["faq_sections"]:
            self.log_test("Admin FAQ Item Create", False, "No admin token or FAQ section ID available")
            return False
            
        try:
            section_id = self.created_resources["faq_sections"][0]
            item_data = {
                "section_id": section_id,
                "question": "How do I create my first fantasy team?",
                "answer": "To create your first fantasy team, go to the 'Create Team' section, select your game, choose your players within the budget, and submit your team.",
                "display_order": 1,
                "is_active": True,
                "tags": ["team creation", "getting started"]
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/faq/items", json=item_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    item_id = data["data"].get("id")
                    self.created_resources["faq_items"].append(item_id)
                    details += f" - FAQ item created successfully with ID: {item_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin FAQ Item Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin FAQ Item Create", False, f"Exception: {str(e)}")
            return False

    # LEGAL DOCUMENT TESTS
    def test_admin_legal_create(self) -> bool:
        """Test admin legal document creation"""
        if not self.admin_token:
            self.log_test("Admin Legal Create", False, "No admin token available")
            return False
            
        try:
            legal_data = {
                "document_type": "terms",
                "title": "Terms of Service",
                "content": "These terms of service govern your use of Fantasy Esports platform...",
                "version": "1.0",
                "is_active": True,
                "effective_date": "2025-01-01T00:00:00Z",
                "language": "en",
                "metadata": {
                    "last_reviewed": "2025-01-01",
                    "review_frequency": "quarterly"
                }
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/legal", json=legal_data)
            
            success = response.status_code == 201
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success") and "data" in data:
                    legal_id = data["data"].get("id")
                    self.created_resources["legal_documents"].append(legal_id)
                    details += f" - Legal document created successfully with ID: {legal_id}"
                else:
                    success = False
                    details += " - Response missing expected data structure"
            
            self.log_test(
                "Admin Legal Create",
                success,
                details,
                response.json() if response.status_code in [200, 201] else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Legal Create", False, f"Exception: {str(e)}")
            return False

    def test_admin_legal_publish(self) -> bool:
        """Test admin legal document publish"""
        if not self.admin_token or not self.created_resources["legal_documents"]:
            self.log_test("Admin Legal Publish", False, "No admin token or legal document ID available")
            return False
            
        try:
            legal_id = self.created_resources["legal_documents"][0]
            
            response = self.session.patch(f"{self.base_url}/api/v1/admin/content/legal/{legal_id}/publish")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Legal document published successfully"
                else:
                    success = False
                    details += " - Publish failed"
            
            self.log_test(
                "Admin Legal Publish",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Admin Legal Publish", False, f"Exception: {str(e)}")
            return False

    # PUBLIC API TESTS
    def test_public_banners_active(self) -> bool:
        """Test public active banners endpoint"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/banners/active")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    banners = data.get("data", [])
                    details += f" - Found {len(banners)} active banners"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Public Banners Active",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Public Banners Active", False, f"Exception: {str(e)}")
            return False

    def test_public_seo_by_slug(self) -> bool:
        """Test public SEO content by slug endpoint"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/seo/home")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    seo_data = data.get("data", {})
                    details += f" - SEO data retrieved for slug 'home'"
                    if seo_data.get("title"):
                        details += f" - Title: {seo_data['title'][:50]}..."
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Public SEO by Slug",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Public SEO by Slug", False, f"Exception: {str(e)}")
            return False

    def test_public_faq_sections(self) -> bool:
        """Test public FAQ sections endpoint"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/faq/sections")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    sections = data.get("data", [])
                    details += f" - Found {len(sections)} FAQ sections"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Public FAQ Sections",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Public FAQ Sections", False, f"Exception: {str(e)}")
            return False

    def test_public_legal_document(self) -> bool:
        """Test public legal document endpoint"""
        try:
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.get(f"{self.base_url}/api/v1/legal/terms")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    legal_doc = data.get("data", {})
                    details += f" - Legal document retrieved"
                    if legal_doc.get("title"):
                        details += f" - Title: {legal_doc['title']}"
                else:
                    success = False
                    details += " - Response missing success field"
            
            self.log_test(
                "Public Legal Document",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Public Legal Document", False, f"Exception: {str(e)}")
            return False

    # ANALYTICS TRACKING TESTS
    def test_banner_click_tracking(self) -> bool:
        """Test banner click tracking"""
        if not self.created_resources["banners"]:
            self.log_test("Banner Click Tracking", False, "No banner ID available for testing")
            return False
            
        try:
            banner_id = self.created_resources["banners"][0]
            
            # Remove admin auth for public endpoint
            original_headers = self.session.headers.copy()
            if 'Authorization' in self.session.headers:
                del self.session.headers['Authorization']
            
            response = self.session.post(f"{self.base_url}/api/v1/banners/{banner_id}/click")
            
            # Restore admin headers
            self.session.headers.clear()
            self.session.headers.update(original_headers)
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    details += " - Banner click tracked successfully"
                else:
                    success = False
                    details += " - Click tracking failed"
            
            self.log_test(
                "Banner Click Tracking",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Banner Click Tracking", False, f"Exception: {str(e)}")
            return False

    def test_content_analytics(self) -> bool:
        """Test content analytics endpoint"""
        if not self.admin_token or not self.created_resources["banners"]:
            self.log_test("Content Analytics", False, "No admin token or banner ID available")
            return False
            
        try:
            banner_id = self.created_resources["banners"][0]
            
            response = self.session.get(f"{self.base_url}/api/v1/admin/content/analytics/banner/{banner_id}")
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                data = response.json()
                if data.get("success"):
                    analytics = data.get("data", {})
                    details += f" - Analytics retrieved for banner {banner_id}"
                    if "views" in analytics:
                        details += f" - Views: {analytics.get('views', 0)}"
                    if "clicks" in analytics:
                        details += f", Clicks: {analytics.get('clicks', 0)}"
                else:
                    success = False
                    details += " - Analytics retrieval failed"
            
            self.log_test(
                "Content Analytics",
                success,
                details,
                response.json() if success else response.text
            )
            return success
            
        except Exception as e:
            self.log_test("Content Analytics", False, f"Exception: {str(e)}")
            return False

    # VALIDATION TESTS
    def test_validation_errors(self) -> bool:
        """Test validation error handling"""
        if not self.admin_token:
            self.log_test("Validation Errors", False, "No admin token available")
            return False
            
        validation_tests_passed = 0
        total_validation_tests = 0
        
        # Test 1: Banner creation with missing required fields
        total_validation_tests += 1
        try:
            invalid_banner_data = {
                "content": "Missing title field"
                # Missing required 'title' field
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/banners", json=invalid_banner_data)
            
            if response.status_code == 400:
                validation_tests_passed += 1
                self.log_test("Validation - Missing Banner Title", True, "Correctly rejected missing title")
            else:
                self.log_test("Validation - Missing Banner Title", False, f"Expected 400, got {response.status_code}")
        except Exception as e:
            self.log_test("Validation - Missing Banner Title", False, f"Exception: {str(e)}")
        
        # Test 2: SEO content with invalid slug format
        total_validation_tests += 1
        try:
            invalid_seo_data = {
                "page_slug": "invalid slug with spaces",  # Invalid slug format
                "title": "Test Title",
                "meta_description": "Test description"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/seo", json=invalid_seo_data)
            
            if response.status_code == 400:
                validation_tests_passed += 1
                self.log_test("Validation - Invalid SEO Slug", True, "Correctly rejected invalid slug format")
            else:
                self.log_test("Validation - Invalid SEO Slug", False, f"Expected 400, got {response.status_code}")
        except Exception as e:
            self.log_test("Validation - Invalid SEO Slug", False, f"Exception: {str(e)}")
        
        # Test 3: Legal document with invalid type
        total_validation_tests += 1
        try:
            invalid_legal_data = {
                "document_type": "invalid_type",  # Invalid document type
                "title": "Test Document",
                "content": "Test content"
            }
            
            response = self.session.post(f"{self.base_url}/api/v1/admin/content/legal", json=invalid_legal_data)
            
            if response.status_code == 400:
                validation_tests_passed += 1
                self.log_test("Validation - Invalid Legal Type", True, "Correctly rejected invalid document type")
            else:
                self.log_test("Validation - Invalid Legal Type", False, f"Expected 400, got {response.status_code}")
        except Exception as e:
            self.log_test("Validation - Invalid Legal Type", False, f"Exception: {str(e)}")
        
        success = validation_tests_passed == total_validation_tests
        self.log_test(
            "Validation Errors - Overall",
            success,
            f"Passed {validation_tests_passed}/{total_validation_tests} validation tests"
        )
        
        return success

    def test_authorization_middleware(self) -> bool:
        """Test authorization middleware for admin endpoints"""
        auth_tests_passed = 0
        total_auth_tests = 0
        
        # Remove admin auth
        original_headers = self.session.headers.copy()
        if 'Authorization' in self.session.headers:
            del self.session.headers['Authorization']
        
        admin_endpoints = [
            "/api/v1/admin/content/banners",
            "/api/v1/admin/content/email-templates",
            "/api/v1/admin/content/campaigns",
            "/api/v1/admin/content/seo",
            "/api/v1/admin/content/faq/sections",
            "/api/v1/admin/content/legal"
        ]
        
        for endpoint in admin_endpoints:
            total_auth_tests += 1
            try:
                response = self.session.get(f"{self.base_url}{endpoint}")
                
                if response.status_code == 401:
                    auth_tests_passed += 1
                    self.log_test(f"Auth Middleware - {endpoint}", True, "Correctly returned 401 for unauthorized access")
                else:
                    self.log_test(f"Auth Middleware - {endpoint}", False, f"Expected 401, got {response.status_code}")
            except Exception as e:
                self.log_test(f"Auth Middleware - {endpoint}", False, f"Exception: {str(e)}")
        
        # Restore admin headers
        self.session.headers.clear()
        self.session.headers.update(original_headers)
        
        success = auth_tests_passed == total_auth_tests
        self.log_test(
            "Authorization Middleware - Overall",
            success,
            f"Passed {auth_tests_passed}/{total_auth_tests} authorization tests"
        )
        
        return success

    def run_comprehensive_cms_tests(self):
        """Run all CMS tests"""
        print("🚀 Starting Comprehensive Content Management System Testing")
        print("=" * 80)
        
        # Test 1: Health Check
        if not self.test_health_check():
            print("❌ Backend is not healthy. Stopping tests.")
            return
        
        # Test 2: Database Setup Verification
        self.test_database_setup_verification()
        
        # Test 3: Admin Authentication
        if not self.authenticate_admin():
            print("❌ Admin authentication failed. Cannot test admin endpoints.")
            return
        
        # Test 4: Admin Banner Management
        print("\n📋 Testing Admin Banner Management...")
        self.test_admin_banner_create()
        self.test_admin_banner_list()
        self.test_admin_banner_update()
        self.test_admin_banner_toggle()
        
        # Test 5: Email Template Management
        print("\n📧 Testing Email Template Management...")
        self.test_admin_email_template_create()
        self.test_admin_email_template_list()
        
        # Test 6: Marketing Campaign Management
        print("\n📈 Testing Marketing Campaign Management...")
        self.test_admin_campaign_create()
        self.test_admin_campaign_status_update()
        
        # Test 7: SEO Content Management
        print("\n🔍 Testing SEO Content Management...")
        self.test_admin_seo_create()
        self.test_admin_seo_list()
        
        # Test 8: FAQ Management
        print("\n❓ Testing FAQ Management...")
        self.test_admin_faq_section_create()
        self.test_admin_faq_item_create()
        
        # Test 9: Legal Document Management
        print("\n📄 Testing Legal Document Management...")
        self.test_admin_legal_create()
        self.test_admin_legal_publish()
        
        # Test 10: Public API Endpoints
        print("\n🌐 Testing Public API Endpoints...")
        self.test_public_banners_active()
        self.test_public_seo_by_slug()
        self.test_public_faq_sections()
        self.test_public_legal_document()
        
        # Test 11: Analytics Tracking
        print("\n📊 Testing Analytics Tracking...")
        self.test_banner_click_tracking()
        self.test_content_analytics()
        
        # Test 12: Validation & Error Handling
        print("\n✅ Testing Validation & Error Handling...")
        self.test_validation_errors()
        self.test_authorization_middleware()
        
        # Generate Summary
        self.generate_summary()

    def generate_summary(self):
        """Generate test summary"""
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
            "Health & Setup": [],
            "Authentication": [],
            "Banner Management": [],
            "Email Templates": [],
            "Marketing Campaigns": [],
            "SEO Content": [],
            "FAQ Management": [],
            "Legal Documents": [],
            "Public APIs": [],
            "Analytics": [],
            "Validation & Security": []
        }
        
        for result in self.test_results:
            test_name = result["test"]
            if "Health" in test_name or "Database Setup" in test_name:
                categories["Health & Setup"].append(result)
            elif "Authentication" in test_name:
                categories["Authentication"].append(result)
            elif "Banner" in test_name:
                categories["Banner Management"].append(result)
            elif "Email Template" in test_name:
                categories["Email Templates"].append(result)
            elif "Campaign" in test_name:
                categories["Marketing Campaigns"].append(result)
            elif "SEO" in test_name:
                categories["SEO Content"].append(result)
            elif "FAQ" in test_name:
                categories["FAQ Management"].append(result)
            elif "Legal" in test_name:
                categories["Legal Documents"].append(result)
            elif "Public" in test_name:
                categories["Public APIs"].append(result)
            elif "Analytics" in test_name or "Click Tracking" in test_name:
                categories["Analytics"].append(result)
            elif "Validation" in test_name or "Auth Middleware" in test_name:
                categories["Validation & Security"].append(result)
        
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
        if success_rate >= 90:
            print("🎉 EXCELLENT: Content Management System is working excellently!")
        elif success_rate >= 75:
            print("✅ GOOD: Content Management System is working well with minor issues.")
        elif success_rate >= 50:
            print("⚠️  MODERATE: Content Management System has some issues that need attention.")
        else:
            print("❌ CRITICAL: Content Management System has significant issues requiring immediate attention.")
        
        # Key functionality assessment
        admin_tests = [r for r in self.test_results if "Admin" in r["test"]]
        public_tests = [r for r in self.test_results if "Public" in r["test"]]
        
        admin_success = sum(1 for r in admin_tests if r["success"]) / len(admin_tests) * 100 if admin_tests else 0
        public_success = sum(1 for r in public_tests if r["success"]) / len(public_tests) * 100 if public_tests else 0
        
        print(f"\nAdmin Content Management: {admin_success:.1f}% functional")
        print(f"Public Content APIs: {public_success:.1f}% functional")
        
        if admin_success >= 75 and public_success >= 75:
            print("\n🚀 READY FOR PRODUCTION: Core CMS functionality is working!")
        else:
            print("\n⚠️  NEEDS WORK: Core CMS functionality requires fixes before production.")
        
        # Show created resources for cleanup reference
        print(f"\n📝 CREATED TEST RESOURCES:")
        for resource_type, ids in self.created_resources.items():
            if ids:
                print(f"  {resource_type}: {len(ids)} items created")

if __name__ == "__main__":
    tester = ContentManagementTester()
    tester.run_comprehensive_cms_tests()