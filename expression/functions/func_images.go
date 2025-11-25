package functions

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// Image dimension cache to avoid re-reading files
	imageDimCache = make(map[string][2]int)
	imageDimMutex = sync.RWMutex{}

	// BaseDir is set by the renderer to enable resolving relative image paths
	BaseDir string
)

// ImageWidth returns the width of an image file in pixels
func ImageWidth(filePath string) (string, error) {
	filePath = strings.Trim(filePath, "'\"")

	// Check for external URLs
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return "", fmt.Errorf("image-width: external URLs not yet supported (tried %s)", filePath)
	}

	width, _, err := getImageDimensions(filePath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%dpx", width), nil
}

// ImageHeight returns the height of an image file in pixels
func ImageHeight(filePath string) (string, error) {
	filePath = strings.Trim(filePath, "'\"")

	// Check for external URLs
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return "", fmt.Errorf("image-height: external URLs not yet supported (tried %s)", filePath)
	}

	_, height, err := getImageDimensions(filePath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%dpx", height), nil
}

// ImageSize returns both width and height as a space-separated string
func ImageSize(filePath string) (string, error) {
	filePath = strings.Trim(filePath, "'\"")

	// Check for external URLs
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return "", fmt.Errorf("image-size: external URLs not yet supported (tried %s)", filePath)
	}

	width, height, err := getImageDimensions(filePath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%dpx %dpx", width, height), nil
}

// getImageDimensions reads an image file and returns its dimensions
func getImageDimensions(filePath string) (int, int, error) {
	// Check cache first
	imageDimMutex.RLock()
	if dims, ok := imageDimCache[filePath]; ok {
		imageDimMutex.RUnlock()
		return dims[0], dims[1], nil
	}
	imageDimMutex.RUnlock()

	// Resolve the file path relative to BaseDir if it's not absolute
	resolvedPath := filePath
	if BaseDir != "" && !filepath.IsAbs(filePath) {
		resolvedPath = filepath.Join(BaseDir, filePath)
	}

	// Try to open the file
	file, err := os.Open(resolvedPath)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot open image file %s: %w", filePath, err)
	}
	defer file.Close()

	// Decode image config (this is fast and doesn't load the whole image)
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot decode image %s: %w", filePath, err)
	}

	// Cache the result
	imageDimMutex.Lock()
	imageDimCache[filePath] = [2]int{config.Width, config.Height}
	imageDimMutex.Unlock()

	return config.Width, config.Height, nil
}

// ClearImageDimCache clears the image dimension cache (useful for testing)
func ClearImageDimCache() {
	imageDimMutex.Lock()
	defer imageDimMutex.Unlock()
	imageDimCache = make(map[string][2]int)
}
