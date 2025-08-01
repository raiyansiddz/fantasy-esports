#!/usr/bin/env python3
"""
Debug script to check database connection and SYSTEM_ADMIN user
"""

import psycopg2
import json

# Database connection details from backend/.env
DATABASE_URL = "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"

def check_database():
    try:
        # Connect to database
        conn = psycopg2.connect(DATABASE_URL)
        cursor = conn.cursor()
        
        print("✅ Database connection successful")
        
        # Check if SYSTEM_ADMIN user exists
        cursor.execute("SELECT id, mobile, first_name, last_name FROM users WHERE mobile = 'SYSTEM_ADMIN';")
        system_user = cursor.fetchone()
        
        if system_user:
            print(f"✅ SYSTEM_ADMIN user found: ID={system_user[0]}, Mobile={system_user[1]}, Name={system_user[2]} {system_user[3]}")
        else:
            print("❌ SYSTEM_ADMIN user not found")
            
        # Check if admin user exists
        cursor.execute("SELECT id, username, role FROM admin_users WHERE username = 'admin';")
        admin_user = cursor.fetchone()
        
        if admin_user:
            print(f"✅ Admin user found: ID={admin_user[0]}, Username={admin_user[1]}, Role={admin_user[2]}")
        else:
            print("❌ Admin user not found")
            
        # Check match_events table structure
        cursor.execute("""
            SELECT column_name, data_type, is_nullable, column_default 
            FROM information_schema.columns 
            WHERE table_name = 'match_events' 
            ORDER BY ordinal_position;
        """)
        columns = cursor.fetchall()
        
        print("\n📋 match_events table structure:")
        for col in columns:
            print(f"  {col[0]}: {col[1]} (nullable: {col[2]}, default: {col[3]})")
            
        # Check foreign key constraints on match_events
        cursor.execute("""
            SELECT 
                tc.constraint_name, 
                tc.table_name, 
                kcu.column_name, 
                ccu.table_name AS foreign_table_name,
                ccu.column_name AS foreign_column_name 
            FROM 
                information_schema.table_constraints AS tc 
                JOIN information_schema.key_column_usage AS kcu
                  ON tc.constraint_name = kcu.constraint_name
                  AND tc.table_schema = kcu.table_schema
                JOIN information_schema.constraint_column_usage AS ccu
                  ON ccu.constraint_name = tc.constraint_name
                  AND ccu.table_schema = tc.table_schema
            WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name='match_events';
        """)
        constraints = cursor.fetchall()
        
        print("\n🔗 match_events foreign key constraints:")
        for constraint in constraints:
            print(f"  {constraint[0]}: {constraint[2]} -> {constraint[3]}.{constraint[4]}")
            
        # Test the exact query that's failing
        print("\n🧪 Testing the exact insert query...")
        try:
            cursor.execute("SELECT id FROM users WHERE mobile = 'SYSTEM_ADMIN'")
            system_user_id = cursor.fetchone()
            if system_user_id:
                print(f"✅ SYSTEM_ADMIN user ID: {system_user_id[0]}")
                
                # Test the insert query (without actually inserting)
                cursor.execute("""
                    EXPLAIN INSERT INTO match_events (match_id, player_id, event_type, points, round_number, 
                                                     description, additional_data, created_by, created_at)
                    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
                """, ('1', 1, 'kill', 2.0, 5, 'Entry frag', '{}', system_user_id[0]))
                
                explain_result = cursor.fetchall()
                print("✅ Insert query plan successful:")
                for row in explain_result:
                    print(f"  {row[0]}")
            else:
                print("❌ SYSTEM_ADMIN user ID not found")
                
        except Exception as e:
            print(f"❌ Insert query test failed: {e}")
        
        cursor.close()
        conn.close()
        
    except Exception as e:
        print(f"❌ Database connection failed: {e}")

if __name__ == "__main__":
    check_database()