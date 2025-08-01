#!/usr/bin/env python3
"""
Advanced debug script to test the exact match event insert scenario
"""

import psycopg2
import json

# Database connection details from backend/.env
DATABASE_URL = "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"

def debug_match_event():
    try:
        # Connect to database
        conn = psycopg2.connect(DATABASE_URL)
        cursor = conn.cursor()
        
        print("✅ Database connection successful")
        
        # Check SYSTEM_ADMIN user
        cursor.execute("SELECT id FROM users WHERE mobile = 'SYSTEM_ADMIN'")
        system_user = cursor.fetchone()
        if system_user:
            system_user_id = system_user[0]
            print(f"✅ SYSTEM_ADMIN user found with ID: {system_user_id}")
        else:
            print("❌ SYSTEM_ADMIN user not found")
            return
            
        # Check if match ID 1 exists
        cursor.execute("SELECT id, name, status FROM matches WHERE id = 1")
        match = cursor.fetchone()
        if match:
            print(f"✅ Match 1 found: {match[1]} (status: {match[2]})")
        else:
            print("❌ Match 1 not found")
            return
            
        # Check if player ID 1 exists
        cursor.execute("SELECT id, name, team_id FROM players WHERE id = 1")
        player = cursor.fetchone()
        if player:
            print(f"✅ Player 1 found: {player[1]} (team: {player[2]})")
        else:
            print("❌ Player 1 not found")
            return
            
        # Test the exact insert with proper parameter binding
        print("\n🧪 Testing match event insert with exact parameters...")
        try:
            cursor.execute("BEGIN")
            
            # Use the exact same query as in the GoLang code
            cursor.execute("""
                INSERT INTO match_events (match_id, player_id, event_type, points, round_number, 
                                         description, additional_data, created_by, created_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, NOW())
                RETURNING id
            """, (1, 1, 'kill', 2.0, 5, 'Entry frag', '{}', system_user_id))
            
            event_id = cursor.fetchone()[0]
            print(f"✅ Match event insert successful! Event ID: {event_id}")
            
            # Check the inserted event
            cursor.execute("""
                SELECT me.id, me.event_type, me.points, p.name, u.mobile 
                FROM match_events me 
                JOIN players p ON me.player_id = p.id 
                JOIN users u ON me.created_by = u.id 
                WHERE me.id = %s
            """, (event_id,))
            
            event_details = cursor.fetchone()
            print(f"✅ Event details: ID={event_details[0]}, Type={event_details[1]}, Points={event_details[2]}, Player={event_details[3]}, CreatedBy={event_details[4]}")
            
            # Rollback the test insert
            cursor.execute("ROLLBACK")
            print("✅ Test insert rolled back (not actually saved)")
            
        except Exception as e:
            cursor.execute("ROLLBACK")
            print(f"❌ Match event insert failed: {e}")
            print(f"Error details: {type(e).__name__}: {str(e)}")
            
        # Check if there are any other constraints or issues
        print("\n🔍 Checking for potential issues...")
        
        # Check if match_id should be string or int
        cursor.execute("SELECT data_type FROM information_schema.columns WHERE table_name = 'match_events' AND column_name = 'match_id'")
        match_id_type = cursor.fetchone()[0]
        print(f"📋 match_id column type: {match_id_type}")
        
        # Check if there are any triggers on match_events
        cursor.execute("""
            SELECT trigger_name, event_manipulation, action_statement 
            FROM information_schema.triggers 
            WHERE event_object_table = 'match_events'
        """)
        triggers = cursor.fetchall()
        if triggers:
            print("🔧 Triggers on match_events table:")
            for trigger in triggers:
                print(f"  {trigger[0]}: {trigger[1]} - {trigger[2]}")
        else:
            print("✅ No triggers on match_events table")
            
        cursor.close()
        conn.close()
        
    except Exception as e:
        print(f"❌ Debug failed: {e}")

if __name__ == "__main__":
    debug_match_event()