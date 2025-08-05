#!/usr/bin/env python3
"""
Initialize Payment Gateway Configurations
This script sets up the default payment gateway configurations for testing
"""

import psycopg2
import os
from urllib.parse import urlparse

def init_payment_gateways():
    # Parse database URL from environment
    database_url = "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"
    
    try:
        # Connect to database
        conn = psycopg2.connect(database_url)
        cursor = conn.cursor()
        
        # Insert Razorpay configuration
        razorpay_query = """
        INSERT INTO payment_gateway_configs (gateway, key1, key2, client_version, is_live, enabled, currency)
        VALUES (%s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (gateway) DO UPDATE SET
            key1 = EXCLUDED.key1,
            key2 = EXCLUDED.key2,
            client_version = EXCLUDED.client_version,
            is_live = EXCLUDED.is_live,
            enabled = EXCLUDED.enabled,
            currency = EXCLUDED.currency,
            updated_at = CURRENT_TIMESTAMP
        """
        
        cursor.execute(razorpay_query, (
            'razorpay',
            'rzp_test_SvOV4KyH7o0FSg',  # Test key from review request
            'test_secret_key_12345',     # Test secret
            '',                          # No client version for Razorpay
            False,                       # Test environment
            True,                        # Enabled
            'INR'                        # Currency
        ))
        
        # Insert PhonePe configuration
        phonepe_query = """
        INSERT INTO payment_gateway_configs (gateway, key1, key2, client_version, is_live, enabled, currency)
        VALUES (%s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (gateway) DO UPDATE SET
            key1 = EXCLUDED.key1,
            key2 = EXCLUDED.key2,
            client_version = EXCLUDED.client_version,
            is_live = EXCLUDED.is_live,
            enabled = EXCLUDED.enabled,
            currency = EXCLUDED.currency,
            updated_at = CURRENT_TIMESTAMP
        """
        
        cursor.execute(phonepe_query, (
            'phonepe',
            'TEST-M22RDIMXCYCLN_25080',   # Test merchant ID from review request
            'test_client_secret_12345',   # Test secret
            '1',                          # Client version
            False,                        # Test environment
            True,                         # Enabled
            'INR'                         # Currency
        ))
        
        # Commit changes
        conn.commit()
        
        print("✅ Payment gateway configurations initialized successfully!")
        print("✅ Razorpay: rzp_test_SvOV4KyH7o0FSg (enabled)")
        print("✅ PhonePe: TEST-M22RDIMXCYCLN_25080 (enabled)")
        
        # Verify the configurations
        cursor.execute("SELECT gateway, key1, enabled FROM payment_gateway_configs ORDER BY gateway")
        configs = cursor.fetchall()
        
        print("\n📋 Current Gateway Configurations:")
        for config in configs:
            gateway, key1, enabled = config
            status = "enabled" if enabled else "disabled"
            print(f"  • {gateway}: {key1} ({status})")
        
        cursor.close()
        conn.close()
        
    except Exception as e:
        print(f"❌ Error initializing payment gateways: {e}")
        return False
    
    return True

if __name__ == "__main__":
    init_payment_gateways()