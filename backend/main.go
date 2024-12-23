package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/h2non/bimg"
)

// isImageFile checks if the file has an image extension
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}
	return validExtensions[ext]
}

// compressImage compresses the image to ensure it's under 1MB
func compressImage(imageData []byte) ([]byte, error) {
	img := bimg.NewImage(imageData)
	
	// Get original size in bytes
	size := len(imageData)
	maxSize := 1024 * 1024 // 1MB in bytes
	
	if size <= maxSize {
		return imageData, nil // No compression needed
	}
	
	// Start with 80% quality
	quality := 80
	
	// Try compression with decreasing quality until size is under 1MB
	for quality >= 20 {
		options := bimg.Options{
			Quality: quality,
		}
		
		compressed, err := img.Process(options)
		if err != nil {
			return nil, fmt.Errorf("compression failed: %v", err)
		}
		
		if len(compressed) <= maxSize {
			return compressed, nil
		}
		
		quality -= 10
	}
	
	// If still too large, try reducing dimensions
	options := bimg.Options{
		Quality: 70,
		Width:   800, // Reduce width to 800px max
	}
	
	return img.Process(options)
}

// handleImageUpload handles the image upload request
func handleImageUpload(ctx context.Context, c *app.RequestContext) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"error": "Failed to get image file from request",
		})
		return
	}
	if !isImageFile(fileHeader.Filename) {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"error": "Uploaded file is not a valid image",
		})
		return
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to open uploaded file",
		})
		return
	}
	defer file.Close()

	// Read the file into memory
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to read uploaded file",
		})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir, err := filepath.Abs("uploads")
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to get absolute path for uploads directory",
		})
		return
	}
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to create uploads directory",
		})
		return
	}

	// Compress the image
	compressed, err := compressImage(buffer.Bytes())
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"error": fmt.Sprintf("Failed to compress image: %v", err),
		})
		return
	}

	// Generate unique filename
	timestamp := time.Now().UnixNano()
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%d%s", timestamp, ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Save the compressed image
	if err := os.WriteFile(filePath, compressed, 0644); err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to save compressed image",
		})
		return
	}

	// Return the file information
	c.JSON(consts.StatusOK, map[string]interface{}{
		"message": "Image uploaded and compressed successfully",
		"original_size": fileHeader.Size,
		"compressed_size": len(compressed),
		"filename": filename,
		"url": func() string {
			publicURL := os.Getenv("PUBLIC_URL")
			fmt.Printf("Debug: PUBLIC_URL=%s\n", publicURL)
			if publicURL == "" {
				publicURL = "http://localhost:8888"
			}
			return fmt.Sprintf("%s/uploads/%s", strings.TrimRight(publicURL, "/"), filename)
		}(),
	})
}

func main() {
	h := server.Default(
		server.WithHostPorts(":8888"),
		server.WithMaxRequestBodySize(20*1024*1024), // Allow up to 20MB uploads
	)

	// Setup CORS middleware
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}
		c.Next(ctx)
	})

	// Basic health check endpoint
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"message": "pong",
		})
	})

	// Image upload endpoint
	h.POST("/upload", handleImageUpload)

	// Serve static files from uploads directory
	uploadsPath, err := filepath.Abs("uploads")
	if err != nil {
		panic(err)
	}
	h.StaticFS("/uploads", &app.FS{Root: uploadsPath, PathRewrite: app.NewPathSlashesStripper(1)})

	h.Spin()
}
