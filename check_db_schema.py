#!/usr/bin/env python3
"""
Database Schema Checker for Crown Jewel Manual Scoring System
"""

import psycopg2
import json

# Database connection
DATABASE_URL = "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"

def check_contests_schema():
    """Check contests table schema"""
    try:
        conn = psycopg2.connect(DATABASE_URL)
        cur = conn.cursor()
        
        print("=== CONTESTS TABLE SCHEMA ===")
        cur.execute("""
            SELECT column_name, data_type, is_nullable 
            FROM information_schema.columns 
            WHERE table_name = 'contests' 
            ORDER BY ordinal_position
        """)
        
        for row in cur.fetchall():
            print(f"Column: {row[0]}, Type: {row[1]}, Nullable: {row[2]}")
        
        print("\n=== SAMPLE CONTEST DATA ===")
        cur.execute("""
            SELECT id, name, total_prize_pool, prize_distribution
            FROM contests 
            WHERE match_id = 1 
            LIMIT 3
        """)
        
        for row in cur.fetchall():
            print(f"Contest {row[0]}: {row[1]}, Prize Pool: {row[2]}, Prize Distribution: {row[3]}")
        
        cur.close()
        conn.close()
        
    except Exception as e:
        print(f"Error checking contests schema: {e}")

def check_contest_participants():
    """Check contest_participants for matches"""
    try:
        conn = psycopg2.connect(DATABASE_URL)
        cur = conn.cursor()
        
        for match_id in [1, 2, 20, 21]:
            print(f"\n=== CONTEST PARTICIPANTS FOR MATCH {match_id} ===")
            cur.execute("""
                SELECT COUNT(*)
                FROM contest_participants cp
                JOIN contests c ON cp.contest_id = c.id
                WHERE c.match_id = %s
            """, (match_id,))
            
            count = cur.fetchone()[0]
            print(f"Match {match_id}: {count} contest participants")
            
            if count > 0:
                cur.execute("""
                    SELECT cp.id, cp.contest_id, cp.user_id, cp.rank
                    FROM contest_participants cp
                    JOIN contests c ON cp.contest_id = c.id
                    WHERE c.match_id = %s
                    LIMIT 5
                """, (match_id,))
                
                for row in cur.fetchall():
                    print(f"  Participant {row[0]}: Contest {row[1]}, User {row[2]}, Rank {row[3]}")
        
        cur.close()
        conn.close()
        
    except Exception as e:
        print(f"Error checking contest participants: {e}")

def check_match_status():
    """Check match statuses"""
    try:
        conn = psycopg2.connect(DATABASE_URL)
        cur = conn.cursor()
        
        print("\n=== MATCH STATUSES ===")
        cur.execute("""
            SELECT id, name, status, match_type, best_of
            FROM matches 
            WHERE id IN (1, 2, 20, 21)
            ORDER BY id
        """)
        
        for row in cur.fetchall():
            print(f"Match {row[0]}: {row[1]} - Status: {row[2]}, Type: {row[3]}, Best of: {row[4]}")
        
        cur.close()
        conn.close()
        
    except Exception as e:
        print(f"Error checking match status: {e}")

if __name__ == "__main__":
    print("🔍 Checking Database Schema for Crown Jewel Manual Scoring System")
    check_contests_schema()
    check_contest_participants()
    check_match_status()