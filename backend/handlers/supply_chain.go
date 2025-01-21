package handlers

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const uploadedModelName = "malicious_chatbot.pth"
const modelScript = "./data/scripts/model.py"
const runModel = "./data/scripts/run.py"

// UploadModelHandler handles the upload and saves the model file locally
func UploadModelHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("modell")
	if err != nil {
		log.Printf("[ERROR] Upload error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload model. Ensure the form field is named 'modell'."})
		return
	}
	defer file.Close()

	log.Printf("[INFO] Uploaded file: %s", header.Filename)

	// Retrieve user session ID
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
		return
	}

	// Save the file locally
	userDir := filepath.Join("/tmp", userID.(string))
	err = os.MkdirAll(userDir, 0755)
	if err != nil {
		log.Printf("[ERROR] Failed to create user directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user directory."})
		return
	}

	modelPath := filepath.Join(userDir, uploadedModelName)
	tempFile, err := os.Create(modelPath)
	if err != nil {
		log.Printf("[ERROR] Error creating file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file."})
		return
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("[ERROR] Error writing to file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file content."})
		return
	}

	log.Printf("[INFO] Model saved locally: %s", modelPath)
	c.JSON(http.StatusOK, gin.H{"message": "Model uploaded successfully!"})
}

// DownloadModelHandler serves the model file for download
func DownloadModelHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
		return
	}

	modelPath := filepath.Join(modelScript)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Printf("[ERROR] Model script not found: %s", modelPath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Model script not found."})
		return
	}

	log.Printf("[INFO] Serving file: %s", modelPath)
	c.File(modelPath)
}

// ProcessSupplyChainFile processes and saves the supply chain file locally
func ProcessSupplyChainFile(c *gin.Context, file multipart.File, filename string) (string, error) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		log.Printf("[ERROR] User session not initialized")
		return "", fmt.Errorf("user session not initialized")
	}

	// Save the file locally
	userDir := filepath.Join("./data/pth-models/", userID.(string))
	err := os.MkdirAll(userDir, 0755)
	if err != nil {
		log.Printf("[ERROR] Failed to create user directory: %v", err)
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	filePath := filepath.Join(userDir, filename)
	tempFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("[ERROR] Failed to create file: %v", err)
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("[ERROR] Failed to write file content: %v", err)
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	log.Printf("[INFO] File saved locally: %s", filePath)
	return fmt.Sprintf("File '%s' uploaded successfully!", filename), nil
}

// ChatWithModelHandler handles local interactions with the uploaded model
func ChatWithModelHandler(c *gin.Context) {
	var requestData struct {
		Prompt string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		log.Printf("[ERROR] Invalid request data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data."})
		return
	}

	prompt := requestData.Prompt
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt cannot be empty."})
		return
	}

	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
		return
	}

	modelPath := filepath.Join("./data/pth-models", userID.(string), uploadedModelName)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Printf("[ERROR] Model file not found: %s", modelPath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Model file not found."})
		return
	}

	// Execute the Python script locally
	cmd := exec.Command("python3", runModel, modelPath, prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Error executing script: %v, output: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  fmt.Sprintf("Failed to execute script: %v", err),
			"output": string(output),
		})
		return
	}

	log.Printf("[INFO] Chat response: %s", string(output))
	c.JSON(http.StatusOK, gin.H{"response": string(output)})
}
