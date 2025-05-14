package main

import (
	"database/sql"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Course struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func main() {
	connStr := os.Getenv("postgresql://neondb_owner:npg_K2FJuC6zpiNO@ep-wandering-tree-a28tfjaf-pooler.eu-central-1.aws.neon.tech/neondb?sslmode=require") // Подключение через переменную окружения

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := gin.Default()

	r.GET("/courses", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, title FROM courses")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var courses []Course
		for rows.Next() {
			var course Course
			if err := rows.Scan(&course.ID, &course.Title); err != nil {
				continue
			}
			courses = append(courses, course)
		}

		c.JSON(http.StatusOK, courses)
	})

	r.Run() // PORT будет из переменной окружения PORT
}
