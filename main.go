package main

import (
    "database/sql"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
)

type Course struct {
    ID   string   `json:"id"`
    Name string  `json:"name"`
    Description string  `json:"description"`
}

func main() {
    connStr := os.Getenv("DATABASE_URL")
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    r := gin.Default()

    r.GET("/courses", func(c *gin.Context) {
        rows, err := db.Query("SELECT id, name, description FROM courses")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var courses []Course
        for rows.Next() {
            var course Course
            if err := rows.Scan(&course.ID, &course.Name, &course.Description); err != nil {
                continue
            }
            courses = append(courses, course)
        }

        c.JSON(http.StatusOK, courses)
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
