package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "your_neon_database_url_here") // замените на свою строку подключения
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Не удалось подключиться к базе:", err)
	}

	router := gin.Default()

	// маршруты
	router.GET("/courses", GetCourses)
	router.GET("/progress/:username", GetProgress)
	router.POST("/progress", MarkCourseCompleted)

	router.Run(":8080") // или другой порт
}

// структура курса
type Course struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GET /courses
func GetCourses(c *gin.Context) {
	rows, err := db.Query("SELECT id, name FROM courses")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		return
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var course Course
		if err := rows.Scan(&course.ID, &course.Name); err != nil {
			continue
		}
		courses = append(courses, course)
	}

	c.JSON(http.StatusOK, courses)
}

// структура прогресса
type ProgressEntry struct {
	Username string `json:"username"`
	CourseID string `json:"course_id"`
}

// GET /progress/:username
func GetProgress(c *gin.Context) {
	username := c.Param("username")

	rows, err := db.Query("SELECT course_id FROM progress WHERE username = $1 AND completed = true", username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		return
	}
	defer rows.Close()

	var completedCourses []string
	for rows.Next() {
		var courseID string
		if err := rows.Scan(&courseID); err == nil {
			completedCourses = append(completedCourses, courseID)
		}
	}

	c.JSON(http.StatusOK, completedCourses)
}

// POST /progress
func MarkCourseCompleted(c *gin.Context) {
	var entry ProgressEntry
	if err := c.BindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
		return
	}

	_, err := db.Exec(`
		INSERT INTO progress (username, course_id, completed)
		VALUES ($1, $2, true)
		ON CONFLICT (username, course_id) DO UPDATE SET completed = true
	`, entry.Username, entry.CourseID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить прогресс"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
