package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

func main() {
    databaseURL := "postgresql://postgres:Raiyan786@database-1.cx8e26gwubmj.ap-south-1.rds.amazonaws.com:5432/postgres"
    
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        log.Fatal("Failed to open database:", err)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        log.Fatal("Failed to ping database:", err)
    }
    
    fmt.Println("✅ Database connection successful!")
    
    // Test a simple query
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
    if err != nil {
        log.Fatal("Failed to query database:", err)
    }
    
    fmt.Printf("✅ Found %d tables in public schema\n", count)
}
