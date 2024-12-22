# Image Upload and Compression Service

A backend service built with Hertz framework that handles image uploads, performs compression, and serves the compressed images.

## Features

- Image upload endpoint with automatic compression
- Supports multiple image formats (JPG, PNG, GIF, BMP, WebP)
- Automatic compression to ensure files are under 1MB
- Static file serving for uploaded images
- CORS support for cross-origin requests

## Project Structure

```
.
├── README.md
├── backend
│   ├── main.go           # Main server implementation
│   ├── parse_response.py # Helper script for parsing responses
│   └── go.mod           # Go module definition
└── go.mod               # Root Go module
```

## API Documentation

### Health Check
- **GET** `/ping`
- Response: `{"message": "pong"}`

### Upload Image
- **POST** `/upload`
- Content-Type: `multipart/form-data`
- Form field: `image`
- Supported formats: JPG, JPEG, PNG, GIF, BMP, WebP
- Response:
  ```json
  {
    "message": "Image uploaded and compressed successfully",
    "original_size": 1234567,
    "compressed_size": 123456,
    "filename": "timestamp.jpg",
    "url": "http://localhost:8888/uploads/timestamp.jpg"
  }
  ```

### Access Uploaded Images
- **GET** `/uploads/{filename}`
- Returns the compressed image file

## Setup Instructions

1. Install dependencies:
   ```bash
   # Install libvips for image processing
   apt-get update && apt-get install -y libvips-dev

   # Install Go dependencies
   go mod tidy
   ```

2. Run the server:
   ```bash
   cd backend
   go run main.go
   ```

The server will start on port 8888.

## Dependencies

- [Hertz](https://github.com/cloudwego/hertz) - HTTP framework
- [bimg](https://github.com/h2non/bimg) - Image processing library
- libvips - Image processing system (system dependency)

## Development

The project uses Go modules for dependency management. Make sure to run `go mod tidy` after adding new imports.

## Testing

You can test the API using curl:

```bash
# Health check
curl http://localhost:8888/ping

# Upload image
curl -X POST -F "image=@test.jpg" http://localhost:8888/upload

# Access uploaded image
curl http://localhost:8888/uploads/filename.jpg
```
