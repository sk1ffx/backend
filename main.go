package main

import (
    "database/sql"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
)

type Course struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Level       string `json:"level"`
    Coss        string `json:"coss"`
}

type Progress struct {
    Username string `json:"username"`
    CourseID string `json:"course_id"`
}

type Message struct {
    ID        int    `json:"id"`
    Username  string `json:"username"`
    CourseID  string `json:"course_id"`
    Message   string `json:"message"`
    CreatedAt string `json:"created_at"`
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
        rows, err := db.Query("SELECT id, name, description, level, coss FROM courses")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var courses []Course
        for rows.Next() {
            var course Course
            if err := rows.Scan(&course.ID, &course.Name, &course.Description, &course.Level, &course.Coss); err != nil {
                continue
            }
            courses = append(courses, course)
        }

        c.JSON(http.StatusOK, courses)
    })

    r.GET("/progress/:username", func(c *gin.Context) {
        username := c.Param("username")
        rows, err := db.Query("SELECT course_id FROM progress WHERE username = $1", username)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var completed []string
        for rows.Next() {
            var courseID string
            if err := rows.Scan(&courseID); err != nil {
                continue
            }
            completed = append(completed, courseID)
        }

        c.JSON(http.StatusOK, completed)
    })

    r.POST("/progress", func(c *gin.Context) {
        var progress Progress
        if err := c.BindJSON(&progress); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
            return
        }

        _, err := db.Exec("INSERT INTO progress (username, course_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
            progress.Username, progress.CourseID)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Progress saved"})
    })

    r.GET("/chat/:course_id", func(c *gin.Context) {
        courseID := c.Param("course_id")

        rows, err := db.Query("SELECT id, username, course_id, message, created_at FROM messages WHERE course_id = $1 ORDER BY created_at ASC", courseID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer rows.Close()

        var messages []Message
        for rows.Next() {
            var msg Message
            if err := rows.Scan(&msg.ID, &msg.Username, &msg.CourseID, &msg.Message, &msg.CreatedAt); err != nil {
                continue
            }
            messages = append(messages, msg)
        }

        c.JSON(http.StatusOK, messages)
    })

    r.POST("/chat", func(c *gin.Context) {
        var msg Message
        if err := c.BindJSON(&msg); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
            return
        }

        _, err := db.Exec("INSERT INTO messages (username, course_id, message) VALUES ($1, $2, $3)",
            msg.Username, msg.CourseID, msg.Message)

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
