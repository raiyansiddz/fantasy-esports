#!/usr/bin/env python3
"""
Debug script to check Crown Jewel Manual Scoring System database state
"""

import psycopg2
import json

# Database connection details from backend/.env
DATABASE_URL = "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"

def check_crown_jewel_state():
    try:
        # Connect to database
        conn = psycopg2.connect(DATABASE_URL)
        cursor = conn.cursor()
        
        print("🔍 CROWN JEWEL MANUAL SCORING SYSTEM DATABASE STATE")
        print("=" * 60)
        
        # Check matches that are being tested
        test_matches = [1, 2, 20, 21]
        
        for match_id in test_matches:
            print(f"\n📊 MATCH {match_id} STATE:")
            
            # Get match details
            cursor.execute("SELECT id, status, winner_team_id FROM matches WHERE id = %s", (match_id,))
            match = cursor.fetchone()
            
            if match:
                print(f"  Status: {match[1]}, Winner: {match[2]}")
            else:
                print(f"  ❌ Match {match_id} not found")
                continue
            
            # Check contests for this match
            cursor.execute("SELECT id, total_prize_pool, status FROM contests WHERE match_id = %s", (match_id,))
            contests = cursor.fetchall()
            
            print(f"  Contests: {len(contests)}")
            for contest in contests:
                print(f"    Contest {contest[0]}: Prize Pool ${contest[1]}, Status: {contest[2]}")
            
            # Check contest participants
            cursor.execute("""
                SELECT COUNT(*) 
                FROM contest_participants cp
                JOIN contests c ON cp.contest_id = c.id
                WHERE c.match_id = %s
            """, (match_id,))
            participant_count = cursor.fetchone()[0]
            
            print(f"  Contest Participants: {participant_count}")
            
            # Check user teams
            cursor.execute("SELECT COUNT(*) FROM user_teams WHERE match_id = %s", (match_id,))
            team_count = cursor.fetchone()[0]
            
            print(f"  User Teams: {team_count}")
            
            # If there are participants, show some details
            if participant_count > 0:
                cursor.execute("""
                    SELECT cp.contest_id, cp.user_id, cp.team_id, cp.rank
                    FROM contest_participants cp
                    JOIN contests c ON cp.contest_id = c.id
                    WHERE c.match_id = %s
                    LIMIT 5
                """, (match_id,))
                participants = cursor.fetchall()
                
                print(f"  Sample Participants:")
                for p in participants:
                    print(f"    Contest {p[0]}: User {p[1]}, Team {p[2]}, Rank {p[3]}")
        
        # Check for any database constraints or issues
        print(f"\n🔧 DATABASE CONSTRAINT CHECKS:")
        
        # Check for orphaned contest participants
        cursor.execute("""
            SELECT COUNT(*) 
            FROM contest_participants cp
            LEFT JOIN contests c ON cp.contest_id = c.id
            WHERE c.id IS NULL
        """)
        orphaned_participants = cursor.fetchone()[0]
        print(f"  Orphaned contest participants: {orphaned_participants}")
        
        # Check for contests without matches
        cursor.execute("""
            SELECT COUNT(*) 
            FROM contests c
            LEFT JOIN matches m ON c.match_id = m.id
            WHERE m.id IS NULL
        """)
        orphaned_contests = cursor.fetchone()[0]
        print(f"  Orphaned contests: {orphaned_contests}")
        
        # Check for user teams without matches
        cursor.execute("""
            SELECT COUNT(*) 
            FROM user_teams ut
            LEFT JOIN matches m ON ut.match_id = m.id
            WHERE m.id IS NULL
        """)
        orphaned_teams = cursor.fetchone()[0]
        print(f"  Orphaned user teams: {orphaned_teams}")
        
        cursor.close()
        conn.close()
        
    except Exception as e:
        print(f"❌ Database check failed: {e}")

if __name__ == "__main__":
    check_crown_jewel_state()