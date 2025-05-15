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
}

type Progress struct {
    Username string `json:"username"`
    CourseID string `json:"course_id"`
}

func main() {
    connStr := os.Getenv("DATABASE_URL")
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    r := gin.Default()

    // Получение списка курсов
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

    // Получение пройденных курсов для пользователя
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

    // Отметить курс как завершённый
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

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
