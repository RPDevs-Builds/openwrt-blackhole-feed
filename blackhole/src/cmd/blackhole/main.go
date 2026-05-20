package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	ipAddr       = flag.String("ip", "0.0.0.0", "IP address to bind to")
	port         = flag.String("port", "8080", "Port to listen on")
	logFile      = flag.String("log", "blackhole.log", "Log file path")
	rootDir      = flag.String("root", "blackhole_root", "Root directory for mirroring")
	contentDir   = flag.String("content", "blackhole_content", "Directory containing real content to serve")
	pixelEnable  = flag.Bool("pixel-enable", true, "Enable tracking pixel response")
	pixelFile    = flag.String("pixel-file", "", "Custom tracking pixel file path")
	pixelHex     = flag.String("pixel-hex", "", "Custom tracking pixel hex string")
	logLevel     = flag.String("log-level", "info", "Log level (debug, info, error)")
	logMaxSizeMB = flag.Int("log-max-size", 10, "Maximum log file size in MB before rotation")
)

const version = "0.1.0"

var defaultTrackingPixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00,
	0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
	0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
}

var activePixel []byte
var activePixelType = "image/gif"

type RequestLog struct {
	Timestamp  string              `json:"timestamp"`
	RemoteAddr string              `json:"remote_addr"`
	Method     string              `json:"method"`
	URL        string              `json:"url"`
	Headers    map[string][]string `json:"headers"`
}

func initPixel() {
	if !*pixelEnable {
		return
	}
	if *pixelFile != "" {
		data, err := os.ReadFile(*pixelFile)
		if err == nil {
			activePixel = data
			ext := strings.ToLower(filepath.Ext(*pixelFile))
			switch ext {
			case ".png":
				activePixelType = "image/png"
			case ".jpg", ".jpeg":
				activePixelType = "image/jpeg"
			default:
				activePixelType = "application/octet-stream"
			}
			return
		}
		log.Printf("Failed to read custom pixel file: %v. Falling back to default/hex.", err)
	}
	if *pixelHex != "" {
		data, err := hex.DecodeString(strings.ReplaceAll(*pixelHex, " ", ""))
		if err == nil {
			activePixel = data
			activePixelType = "application/octet-stream" // default if not specified
			return
		}
		log.Printf("Failed to decode custom pixel hex: %v. Falling back to default.", err)
	}
	activePixel = defaultTrackingPixel
	activePixelType = "image/gif"
}

func main() {
	flag.Parse()

	// Ensure root directory exists
	if err := os.MkdirAll(*rootDir, 0755); err != nil {
		log.Fatalf("Failed to create root directory: %v", err)
	}
	// Ensure content directory exists
	if err := os.MkdirAll(*contentDir, 0755); err != nil {
		log.Fatalf("Failed to create content directory: %v", err)
	}

	initPixel()

	http.HandleFunc("/", handleRequest)

	addr := fmt.Sprintf("%s:%s", *ipAddr, *port)
	fmt.Printf("Blackhole server v%s listening on %s\n", version, addr)
	fmt.Printf("Logging to: %s (Level: %s, Max Size: %dMB)\n", *logFile, *logLevel, *logMaxSizeMB)
	fmt.Printf("Mirroring to: %s/\n", *rootDir)
	fmt.Printf("Serving content from: %s/\n", *contentDir)
	if *pixelEnable {
		fmt.Printf("Tracking pixel enabled: Type %s, Length %d bytes\n", activePixelType, len(activePixel))
	} else {
		fmt.Printf("Tracking pixel disabled.\n")
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 1. Log the request
	go logRequest(r)

	// 2. Check if real content exists and serve it
	cleanPath := path.Clean("/" + r.URL.Path)
	contentPath := filepath.Join(*contentDir, cleanPath)

	info, err := os.Stat(contentPath)
	if err == nil {
		// If it's a file and has content
		if !info.IsDir() && info.Size() > 0 {
			http.ServeFile(w, r, contentPath)
			return
		}
		// If it's a directory, check for index.html
		if info.IsDir() {
			idxPath := filepath.Join(contentPath, "index.html")
			idxInfo, idxErr := os.Stat(idxPath)
			if idxErr == nil && !idxInfo.IsDir() && idxInfo.Size() > 0 {
				http.ServeFile(w, r, contentPath)
				return
			}
		}
	}

	// 3. Mirror the path to filesystem
	go mirrorPath(r.URL.Path)

	// 4. Respond with tracking pixel if enabled
	if *pixelEnable {
		w.Header().Set("Content-Type", activePixelType)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		w.Write(activePixel)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func checkLogRotation() {
	if *logMaxSizeMB <= 0 {
		return
	}
	info, err := os.Stat(*logFile)
	if err != nil {
		return
	}
	if info.Size() > int64(*logMaxSizeMB*1024*1024) {
		backupName := fmt.Sprintf("%s.%s", *logFile, time.Now().Format("20060102-150405"))
		os.Rename(*logFile, backupName)
	}
}

func logRequest(r *http.Request) {
	checkLogRotation()

	entry := RequestLog{
		Timestamp:  time.Now().Format(time.RFC3339),
		RemoteAddr: r.RemoteAddr,
		Method:     r.Method,
		URL:        r.URL.String(),
		Headers:    r.Header,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		if *logLevel == "debug" {
			log.Printf("Error marshaling log entry: %v", err)
		}
		return
	}

	f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		if *logLevel == "error" || *logLevel == "info" || *logLevel == "debug" {
			log.Printf("Error opening log file: %v", err)
		}
		return
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		if *logLevel == "error" || *logLevel == "info" || *logLevel == "debug" {
			log.Printf("Error writing to log file: %v", err)
		}
	}
}

func mirrorPath(rawPath string) {
	// Sanitize and clean path
	cleanPath := path.Clean("/" + rawPath)
	if cleanPath == "/" {
		return
	}

	targetPath := filepath.Join(*rootDir, cleanPath)

	// Logic: If it has an extension, treat as file. Else treat as directory.
	ext := path.Ext(cleanPath)
	if ext != "" {
		// It's a file. Create the parent directories first.
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			if *logLevel == "error" || *logLevel == "debug" {
				log.Printf("Error creating directory %s: %v", dir, err)
			}
			return
		}

		// Create the empty file
		f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			if *logLevel == "error" || *logLevel == "debug" {
				log.Printf("Error creating file %s: %v", targetPath, err)
			}
			return
		}
		f.Close()
	} else {
		// It's a directory. Create it (and parents).
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			if *logLevel == "error" || *logLevel == "debug" {
				log.Printf("Error creating directory %s: %v", targetPath, err)
			}
		}
	}
}
