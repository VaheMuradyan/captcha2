package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/nfnt/resize"
)

var ctx = context.Background()
var rdb *redis.Client

type ShapeType int

const (
	Circle ShapeType = iota
	Square
	Triangle
)

type Shape struct {
	Type     ShapeType `json:"type"`
	Row      int       `json:"row"`
	Col      int       `json:"col"`
	Position int       `json:"position"`
}

type CaptchaData struct {
	Shapes   []Shape `json:"shapes"`
	Sequence string  `json:"sequence"` // Add this field
}

func generateUniqueKey() string {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(10000)
	return fmt.Sprintf("captcha:%d:%d", timestamp, random)
}

// Ֆունկցիա բանալու յունիքությունը ստուգելու համար
func getUniqueRedisKey(baseKey string) string {
	counter := 1
	newKey := baseKey

	// Ստուգում ենք արդյոք բանալին գոյություն ունի
	for {
		exists, err := rdb.Exists(ctx, newKey).Result()
		if err != nil {
			log.Printf("Error checking key existence: %v", err)
			return baseKey // Սխալի դեպքում վերադարձնում ենք սկզբնական բանալին
		}

		if exists == 0 {
			// Բանալին գոյություն չունի, կարող ենք օգտագործել
			break
		}

		// Եթե բանալին գոյություն ունի, ավելացնում ենք հերթական համար
		newKey = fmt.Sprintf("%s:%d", baseKey, counter)
		counter++
	}

	return newKey
}

func drawShape(gc *draw2dimg.GraphicContext, shapeType ShapeType, x, y float64, size float64) {
	gc.SetFillColor(color.Transparent) // Թափանցիկ լցոնում
	gc.SetStrokeColor(color.White)     // Սև եզրագծեր
	gc.SetLineWidth(2)

	switch shapeType {
	case Circle:
		gc.BeginPath()
		gc.ArcTo(x+size/2, y+size/2, size/3, size/3, 0, 2*3.14159)
		gc.Close()
		gc.FillStroke()
	case Square:
		margin := size / 4
		gc.BeginPath()
		gc.MoveTo(x+margin, y+margin)
		gc.LineTo(x+size-margin, y+margin)
		gc.LineTo(x+size-margin, y+size-margin)
		gc.LineTo(x+margin, y+size-margin)
		gc.Close()
		gc.FillStroke()
	case Triangle:
		margin := size / 4
		gc.BeginPath()
		gc.MoveTo(x+size/2, y+margin)
		gc.LineTo(x+size-margin, y+size-margin)
		gc.LineTo(x+margin, y+size-margin)
		gc.Close()
		gc.FillStroke()
	}
}
func generateCaptcha() (CaptchaData, []byte) {
	width, height := 300, 200

	// Load background image
	backgroundFile, err := os.Open("image/background.png") // Your image path here
	if err != nil {
		log.Printf("Error loading background: %v", err)
		// Fallback to white background if image loading fails
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		gc := draw2dimg.NewGraphicContext(img)
		gc.SetFillColor(color.White)
		gc.Clear()
		return createCaptchaWithBackground(img)
	}
	defer backgroundFile.Close()

	// Decode the background image
	background, _, err := image.Decode(backgroundFile)
	if err != nil {
		log.Printf("Error decoding background: %v", err)
		// Fallback to white background
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		gc := draw2dimg.NewGraphicContext(img)
		gc.SetFillColor(color.White)
		gc.Clear()
		return createCaptchaWithBackground(img)
	}

	// Resize background if needed
	resized := resize.Resize(uint(width), uint(height), background, resize.Lanczos3)

	// Convert to RGBA
	bounds := resized.Bounds()
	img := image.NewRGBA(bounds)
	for y := 0; y < bounds.Max.Y; y++ {
		for x := 0; x < bounds.Max.X; x++ {
			img.Set(x, y, resized.At(x, y))
		}
	}

	return createCaptchaWithBackground(img)
}

func createCaptchaWithBackground(img *image.RGBA) (CaptchaData, []byte) {
	gc := draw2dimg.NewGraphicContext(img)
	width, height := float64(img.Bounds().Max.X), float64(img.Bounds().Max.Y)

	cellWidth := width / 6
	cellHeight := height / 4

	// Generate grid positions
	positions := make([][]int, 0)
	for i := 0; i < 4; i++ {
		for j := 0; j < 6; j++ {
			positions = append(positions, []int{i, j})
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	// Create and shuffle shape types for random order
	shapeTypes := []ShapeType{Circle, Square, Triangle}
	rand.Shuffle(len(shapeTypes), func(i, j int) {
		shapeTypes[i], shapeTypes[j] = shapeTypes[j], shapeTypes[i]
	})

	// Create shapes array
	shapes := make([]Shape, 3)

	// Create sequence description
	sequence := make([]string, 3)

	// Draw shapes with clear outline
	gc.SetStrokeColor(color.RGBA{0, 0, 0, 255}) // Black outline
	gc.SetLineWidth(2)

	for i := 0; i < 3; i++ {
		pos := positions[i]
		shapes[i] = Shape{
			Type:     shapeTypes[i],
			Row:      pos[0],
			Col:      pos[1],
			Position: i + 1,
		}

		// Add shape name to sequence
		switch shapeTypes[i] {
		case Circle:
			sequence[i] = "Circle"
		case Square:
			sequence[i] = "Square"
		case Triangle:
			sequence[i] = "Triangle"
		}

		x := float64(pos[1]) * cellWidth
		y := float64(pos[0]) * cellHeight
		drawShape(gc, shapeTypes[i], x, y, math.Min(cellWidth, cellHeight))
	}

	var imgBytes []byte
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	imgBytes = buf.Bytes()

	// Return both the shapes data and the sequence
	return CaptchaData{
		Shapes:   shapes,
		Sequence: strings.Join(sequence, " → "),
	}, imgBytes
}

func main() {
	// Redis կապի ստուգում
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0, // default DB
	})

	// Ստուգում ենք Redis-ի կապը
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis error: %v", err)
	} else {
		log.Printf("Redis connected! Response: %v", pong)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		// Այստեղ պետք է ավելացնել X-Captcha-Key
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Captcha-Key") // Փոխված տող
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Captcha-Key") // Ավելացված տող

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/api/captcha", func(c *gin.Context) {
		captchaData, imgBytes := generateCaptcha()

		// Ստեղծում ենք յունիք բանալի
		baseKey := generateUniqueKey()
		redisKey := getUniqueRedisKey(baseKey)

		captchaJSON, _ := json.Marshal(captchaData)
		rdb.Set(ctx, redisKey, string(captchaJSON), 5*time.Minute)

		c.Header("X-Captcha-Key", redisKey)
		c.Header("Content-Type", "image/png")
		c.Data(http.StatusOK, "image/png", imgBytes)
	})

	r.POST("/api/verify", func(c *gin.Context) {

		// Ստանում ենք բանալին request-ից
		captchaKey := c.GetHeader("X-Captcha-Key")
		if captchaKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Captcha key is required"})
			return
		}

		var userSequence [][]int
		if err := c.BindJSON(&userSequence); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid":     false,
				"message":   "Invalid input format",
				"userInput": userSequence,
			})
			return
		}

		storedCaptchaJSON, err := rdb.Get(ctx, captchaKey).Result()
		if err == redis.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"valid":     false,
				"message":   "Captcha expired",
				"userInput": userSequence,
			})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"valid":     false,
				"message":   "Server error",
				"userInput": userSequence,
			})
			return
		}

		var storedCaptcha CaptchaData
		json.Unmarshal([]byte(storedCaptchaJSON), &storedCaptcha)

		if len(userSequence) != 3 {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid":          false,
				"message":        "Must select exactly 3 shapes",
				"userInput":      userSequence,
				"expectedLength": 3,
			})
			return
		}

		correct := true
		for i, shape := range storedCaptcha.Shapes {
			if shape.Row != userSequence[i][0] || shape.Col != userSequence[i][1] {
				correct = false
				break
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"valid": correct,
			"message": map[bool]string{
				true:  "Sequence verified successfully",
				false: "Incorrect sequence",
			}[correct],
			"userInput": userSequence,
		})
		rdb.Del(ctx, captchaKey)
	})

	r.GET("/api/sequence", func(c *gin.Context) {

		// Ստանում ենք բանալին query-ից
		captchaKey := c.GetHeader("X-Captcha-Key")
		if captchaKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Captcha key is required",
			})
			return
		}
		// Ստանում ենք պահպանված տվյալները Redis-ից
		storedCaptchaJSON, err := rdb.Get(ctx, captchaKey).Result()
		if err == redis.Nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No captcha data found",
			})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get sequence",
			})
			return
		}

		var storedCaptcha CaptchaData
		if err := json.Unmarshal([]byte(storedCaptchaJSON), &storedCaptcha); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse sequence data",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sequence": storedCaptcha.Sequence,
		})
	})

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
