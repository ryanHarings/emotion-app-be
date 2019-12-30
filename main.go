package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

func repeatHandler(r int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var buffer bytes.Buffer
		for i := 0; i < r; i++ {
			buffer.WriteString("Hello from Go!\n")
		}
		c.String(http.StatusOK, buffer.String())
	}
}

func sendEmotions() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Emotion struct {
			ID    int    `json:"id"`
			Value string `json:"value"`
			Color string `json:"color"`
		}
		/*emotions := [2]emotion struct{
			{id: 1, value: "Sad"},
			{id: 2, value: "Happy"}
		}*/

		emotions := []Emotion{{1, "Sad", "#4C8EE6"}, Emotion{2, "Happy", "#1FBF34"}, Emotion{3, "Anxious", "#999999"}}

		c.JSON(http.StatusOK, emotions)
	}
}

/*func sendEmotions(c *gin.Context) {
	emotionOb := [
		{id: 1, value: "Sad", color: "#4C8EE6"},
		{id: 2, value: "Happy", color: "#1FBF34"}
	]
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H {emotionOb})
 }*/

func dbFunc(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := db.Exec("CREATE TABLE IF NOT EXISTS ticks (tick timestamp)"); err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error creating database table: %q", err))
			return
		}

		if _, err := db.Exec("INSERT INTO ticks VALUES (now())"); err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error incrementing tick: %q", err))
			return
		}

		rows, err := db.Query("SELECT tick FROM ticks")
		if err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error reading ticks: %q", err))
			return
		}

		defer rows.Close()
		for rows.Next() {
			var tick time.Time
			if err := rows.Scan(&tick); err != nil {
				c.String(http.StatusInternalServerError,
					fmt.Sprintf("Error scanning ticks: %q", err))
				return
			}
			c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", tick.String()))
		}
	}
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	tStr := os.Getenv("REPEAT")
	repeat, err := strconv.Atoi(tStr)
	if err != nil {
		log.Printf("Error converting $REPEAT to an int: %q - Using default\n", err)
		repeat = 5
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	// emotions, err := strconv.Atoi(tStr)
	// if err != nil {
	// 	log.Printf("Get request failed - Using default\n", err)
	// }

	router := gin.New()
	// router.Use(gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://ryanharings.github.io/emotion-app"},
		AllowMethods:     []string{"PUT", "GET", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		// c.HTML(http.StatusOK, "index.tmpl.html", nil)
		c.String(http.StatusOK, "Hi Ryan")
	})

	router.GET("/repeat", repeatHandler(repeat))

	router.GET("/db", dbFunc(db))

	router.GET("/emotions", sendEmotions())

	// router.OPTIONS("/emotions", preflight)

	router.Run(":" + port)
}

// func preflight(c *gin.Context) {
// 	c.Header("Access-Control-Allow-Origin", "*")
// 	c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
// 	c.JSON(http.StatusOK, struct{}{})
// }
