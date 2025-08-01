#!/usr/bin/env python3
"""
Fix script to create SYSTEM_ADMIN user and test the Fantasy Points Engine
"""

import psycopg2
import json

# Database connection details from backend/.env
DATABASE_URL = "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"

def fix_database():
    try:
        # Connect to database
        conn = psycopg2.connect(DATABASE_URL)
        cursor = conn.cursor()
        
        print("✅ Database connection successful")
        
        # Create SYSTEM_ADMIN user if it doesn't exist
        cursor.execute("""
            INSERT INTO users (mobile, email, first_name, last_name, is_verified, is_active, account_status, kyc_status, referral_code) 
            VALUES ('SYSTEM_ADMIN', 'system@fantasy-esports.com', 'System', 'Administrator', true, true, 'active', 'verified', 'SYS_ADMIN')
            ON CONFLICT (mobile) DO NOTHING
            RETURNING id;
        """)
        
        result = cursor.fetchone()
        if result:
            print(f"✅ Created SYSTEM_ADMIN user with ID: {result[0]}")
        else:
            # User already exists, get the ID
            cursor.execute("SELECT id FROM users WHERE mobile = 'SYSTEM_ADMIN'")
            existing_user = cursor.fetchone()
            if existing_user:
                print(f"✅ SYSTEM_ADMIN user already exists with ID: {existing_user[0]}")
            else:
                print("❌ Failed to create or find SYSTEM_ADMIN user")
                return
        
        # Commit the changes
        conn.commit()
        
        # Test the insert query that was failing
        print("\n🧪 Testing match event insert...")
        try:
            cursor.execute("SELECT id FROM users WHERE mobile = 'SYSTEM_ADMIN'")
            system_user_id = cursor.fetchone()[0]
            
            # Test insert (rollback after test)
            cursor.execute("BEGIN")
            cursor.execute("""
                INSERT INTO match_events (match_id, player_id, event_type, points, round_number, 
                                         description, additional_data, created_by, created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
                RETURNING id
            """, (1, 1, 'kill', 2.0, 5, 'Entry frag', '{}', system_user_id))
            
            event_id = cursor.fetchone()[0]
            print(f"✅ Match event insert successful! Event ID: {event_id}")
            
            # Rollback the test insert
            cursor.execute("ROLLBACK")
            print("✅ Test insert rolled back (not actually saved)")
            
        except Exception as e:
            cursor.execute("ROLLBACK")
            print(f"❌ Match event insert failed: {e}")
        
        cursor.close()
        conn.close()
        
        print("\n🎯 Database fix completed! The Fantasy Points Engine should now work.")
        
    except Exception as e:
        print(f"❌ Database fix failed: {e}")

if __name__ == "__main__":
    fix_database()