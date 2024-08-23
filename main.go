package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Post struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/article", getPosts)
	router.GET("/article/:id", getPostById)
	router.PUT("/article/:id", updatePostById)
	router.POST("/article", addPost)
	router.DELETE("/article/:id", deletePostById)

	router.Run("localhost:8080")
}

func getPosts(context *gin.Context) {
	rows, err := db.Query("SELECT id, title, content, category, status status FROM posts")
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.Status); err != nil {
			context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		posts = append(posts, post)
	}

	context.IndentedJSON(http.StatusOK, posts)
}

func getPostById(context *gin.Context) {
	id := context.Param("id")
	var post Post

	err := db.QueryRow("SELECT id, title, content, category, status FROM posts WHERE id = ?", id).Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			context.IndentedJSON(http.StatusNotFound, gin.H{"error": "post not found"})
		} else {
			context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	context.IndentedJSON(http.StatusOK, post)
}

func addPost(context *gin.Context) {
	var newPost Post
	if err := context.BindJSON(&newPost); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newPost.Title == "" || newPost.Content == "" || newPost.Category == "" || newPost.Status == "" {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing or invalid input"})
		return
	}
	if len(newPost.Title) < 20 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title must be at least 20 characters"})
		return
	}
	if len(newPost.Content) < 200 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Content must be at least 200 characters"})
		return
	}
	if len(newPost.Category) < 3 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category must be at least 3 characters"})
		return
	}
	if newPost.Status != "publish" && newPost.Status != "draft" && newPost.Status != "trash" {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Status must be either publish, draft, or trash"})
		return
	}

	result, err := db.Exec("INSERT INTO posts (title, content, category, status) VALUES (?, ?, ?, ?)", newPost.Title, newPost.Content, newPost.Category, newPost.Status)
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newPost.ID = int(id)
	context.JSON(http.StatusCreated, newPost)
}

func updatePostById(context *gin.Context) {
	id := context.Param("id")

	postID, err := strconv.Atoi(id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var updatedPost Post
	if err := context.BindJSON(&updatedPost); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updatedPost.Title == "" || updatedPost.Content == "" || updatedPost.Category == "" || updatedPost.Status == "" {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "missing or invalid input"})
		return
	}
	if len(updatedPost.Title) < 20 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Title must be at least 20 characters"})
		return
	}
	if len(updatedPost.Content) < 200 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Content must be at least 200 characters"})
		return
	}
	if len(updatedPost.Category) < 3 {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Category must be at least 3 characters"})
		return
	}
	if updatedPost.Status != "publish" && updatedPost.Status != "draft" && updatedPost.Status != "trash" {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Status must be either publish, draft, or trash"})
		return
	}

	result, err := db.Exec("UPDATE posts SET title = ?, content = ?, category = ?, status = ? WHERE id = ?", updatedPost.Title, updatedPost.Content, updatedPost.Category, updatedPost.Status, postID)
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		context.IndentedJSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	updatedPost.ID = int(postID)
	context.IndentedJSON(http.StatusOK, updatedPost)
}

func deletePostById(context *gin.Context) {
	id := context.Param("id")

	result, err := db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		context.IndentedJSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}
