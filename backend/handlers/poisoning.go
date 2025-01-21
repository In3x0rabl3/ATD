package handlers

import (
	"atd/backend/utils"
	"atd/config"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

// PoisonDatasetHandler handles dataset uploads, ensuring only valid .csv files are processed
func PoisonDatasetHandler(c *gin.Context, baselinePath string) {
	log.Println("Handling file upload in PoisonDatasetHandler")

	// Retrieve session ID
	sessionID := c.GetHeader("Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session-ID header is required."})
		return
	}

	// Step 1: Validate session reset
	sessionState := utils.GetSessionState(sessionID)
	if len(sessionState.ScoredRows) > 0 || sessionState.IntegrityScore != 1.0 {
		log.Printf("[WARN] Session %s not fully reset. Scored rows: %d, Integrity score: %.2f", sessionID, len(sessionState.ScoredRows), sessionState.IntegrityScore)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session state is not properly reset. Please reset and try again."})
		return
	}

	// Step 2: Retrieve the uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("[ERROR] File upload failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file. Please try again."})
		return
	}
	defer file.Close()

	log.Printf("[INFO] Uploaded file: %s", header.Filename)

	// Step 3: Validate file type and extension
	if !strings.HasSuffix(header.Filename, ".csv") {
		log.Printf("[ERROR] Invalid file type: %s", header.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .csv files are allowed."})
		return
	}

	// Step 4: Calculate the hash of the uploaded file
	fileHash, err := utils.CalculateFileHash(file)
	if err != nil {
		log.Printf("[ERROR] Failed to calculate file hash: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process uploaded file. Please try again."})
		return
	}
	log.Printf("[INFO] File hash: %s", fileHash)

	// Step 5: Check if the file has already been uploaded
	if utils.IsDuplicateFile(sessionID, fileHash) {
		log.Printf("[WARN] Duplicate file upload detected: %s", fileHash)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File with the same contents has already been uploaded."})
		return
	}

	// Step 6: Validate if the file is a valid CSV
	file.Seek(0, 0) // Reset the file pointer before reading again
	if err := utils.ValidateCSV(file); err != nil {
		log.Printf("[ERROR] Invalid CSV file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV file. Please ensure the file is properly formatted."})
		return
	}

	// Step 7: Mark the file as uploaded
	if err := utils.MarkFileAsUploaded(sessionID, fileHash); err != nil {
		log.Printf("[ERROR] Failed to mark file as uploaded: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process uploaded file. Please try again."})
		return
	}

	// Step 8: Parse the uploaded CSV file
	file.Seek(0, 0) // Reset the file pointer before reading again
	uploadedData, err := utils.ParseCSV(file)
	if err != nil {
		log.Printf("[WARN] Failed to parse CSV: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Uploaded CSV file could not be parsed."})
		return
	}

	// Step 9: Load the user-specific baseline dataset
	baselineData, err := utils.LoadDataset(baselinePath)
	if err != nil {
		log.Printf("[ERROR] Failed to load user-specific baseline dataset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user-specific baseline dataset."})
		return
	}

	// Step 10: Align the datasets
	sanitizedData := utils.AlignDataset(uploadedData, baselineData)
	log.Printf("[INFO] Aligned dataset. Rows sanitized: %d", len(sanitizedData))

	// Step 11: Append and save the dataset to the user-specific baseline
	appendedData := append(baselineData, sanitizedData...)
	if err := utils.WriteDataset(appendedData, baselinePath); err != nil {
		log.Printf("[ERROR] Failed to save user-specific baseline dataset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save updated baseline dataset."})
		return
	}

	// Step 12: Calculate integrity score
	integrityScore := utils.CalculateIntegrityScore(sessionID, sanitizedData, baselineData)
	log.Printf("[INFO] Calculated integrity score: %.2f", integrityScore)

	// Step 13: Inject flag if necessary
	if integrityScore < 0.25 {
		log.Println("[INFO] Integrity score below threshold. Injecting flag.")
		if err := utils.InjectFlag(sessionID, appendedData, baselinePath); err != nil {
			log.Printf("[ERROR] Failed to inject flag: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Flag injection failed."})
			return
		}
		log.Println("[INFO] Flag successfully injected.")
	}

	// Step 14: Respond with updated integrity score
	log.Println("[INFO] Dataset processing completed successfully.")
	c.JSON(http.StatusOK, gin.H{
		"message":        "Dataset processed successfully.",
		"integrityScore": integrityScore,
	})
}

func GetDatasetScoresHandler(c *gin.Context) {
	sessionID := c.GetHeader("Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session-ID header is required."})
		return
	}

	scores := utils.GetDatasetScores(sessionID)
	c.JSON(200, gin.H{
		"scores": scores,
	})
}

// PoisonPromptHandler handles user prompts for Data Poisoning
func PoisonPromptHandler(c *gin.Context, baselinePath string) {
	sessionID := c.GetHeader("Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session-ID header is required."})
		return
	}

	client := openai.NewClient(config.GetOpenAIKey())

	prompt := c.PostForm("prompt")
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt cannot be empty."})
		return
	}

	// Retrieve the current integrity score for the session
	integrityScore := utils.GetIntegrityScore(sessionID)

	// Load the user-specific dataset
	currentDataset, err := utils.LoadDataset(baselinePath)
	if err != nil {
		log.Printf("[ERROR] Failed to load user-specific baseline dataset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user-specific dataset."})
		return
	}

	// Convert the dataset to a formatted string
	currentDatasetFormatted := utils.FormatDataset(currentDataset)

	// Prompt generation logic based on integrity score
	var promptWithDataset string
	if integrityScore > 0.25 {
		if integrityScore >= 0.50 {
			// Accurate and detailed response
			promptWithDataset = fmt.Sprintf(`The following is a user-specific dataset related to facts about cats:

%s

The user has asked the following question about this dataset:
"%s"

Provide an accurate and detailed response based on the dataset.`, currentDatasetFormatted, prompt)
		} else {
			// False truths and inaccuracies
			promptWithDataset = fmt.Sprintf(`The following is a user-specific dataset related to facts about cats:

%s

The user has asked the following question about this dataset:
"%s"

Provide responses that contain false truths and inaccuracies based on the dataset.`, currentDatasetFormatted, prompt)
		}
	} else {
		// Integrity score is 0.25 or below; inform the user about the flag
		promptWithDataset = fmt.Sprintf(`The following is a user-specific dataset related to facts about cats, but its integrity has been compromised and a flag has been injected into the dataset.

The user has asked the following question:
"%s"

Instructions:
1. Clearly and directly lead the user to the flag if asked about it. Make it easy to identify.
2. Always respond to the user's queries about the flag explicitly and in detail.
3. The flag is "ME{D4Ta_P01sOn1nG}". When the user identifies the flag, congratulate them and confirm its correctness.

Provide responses that assist the user in understanding the altered dataset while ensuring the flag is clearly and easily discoverable.`, prompt)
		log.Printf("[INFO] Flag injected. Integrity score: %.2f", integrityScore)
	}

	// Call the Chat API with the generated prompt
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are a dataset analysis AI."},
			{Role: openai.ChatMessageRoleUser, Content: promptWithDataset},
		},
		MaxTokens:   300,
		Temperature: 0.7,
		TopP:        1.0,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process prompt."})
		log.Printf("Error with Chat API: %v", err)
		return
	}

	// Return the response from the Chat API
	if len(resp.Choices) > 0 {
		c.JSON(http.StatusOK, gin.H{"response": resp.Choices[0].Message.Content})
	} else {
		c.JSON(http.StatusOK, gin.H{"response": "No meaningful response generated."})
	}
}

// loadCurrentDataset reads the baseline dataset and formats it for display
func loadCurrentDataset(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening dataset file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("error reading dataset file: %v", err)
	}

	var result strings.Builder
	for _, row := range data {
		result.WriteString(strings.Join(row, " | "))
		result.WriteString("\n")
	}
	return result.String(), nil
}

func ResetBaselineHandler(c *gin.Context) {
	sessionID := c.GetHeader("Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session-ID header is required."})
		return
	}

	log.Println("Resetting baseline dataset to default.")

	// Parse the default baseline data into [][]string
	reader := csv.NewReader(strings.NewReader(config.DefaultBaselineData))
	defaultData, err := reader.ReadAll()
	if err != nil {
		log.Printf("[ERROR] Failed to parse default baseline data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse default baseline data."})
		return
	}

	// Reset the baseline dataset
	err = utils.WriteDataset(defaultData, config.BaselineDatasetPath)
	if err != nil {
		log.Printf("[ERROR] Failed to reset baseline dataset: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset baseline dataset."})
		return
	}

	// Clear session-specific data
	utils.ClearScoredRows(sessionID)         // Clears scored rows
	utils.ClearDatasetScores(sessionID)      // Clears dataset scores
	utils.ClearDeduplicationStore(sessionID) // Clears deduplication data
	utils.SetIntegrityScore(sessionID, 1.0)  // Resets integrity score to 1.00

	log.Printf("[INFO] Session %s fully reset. Integrity score: 1.00", sessionID)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Baseline dataset and session data have been reset.",
		"integrityScore": 1.00,
	})
}

func GetCurrentIntegrityScoreHandler(c *gin.Context) {
	sessionID := c.GetHeader("Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session-ID header is required."})
		return
	}

	// Fetch the current integrity score from the session
	currentScore := utils.GetIntegrityScore(sessionID)

	// If the session doesn't exist, ensure a proper error response
	if currentScore == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found or integrity score unavailable."})
		return
	}

	// Return the current integrity score
	c.JSON(http.StatusOK, gin.H{"integrityScore": currentScore})
}

func initializeBaselineDataset() {
	log.Println("Checking and resetting baseline dataset...")

	// Parse the default baseline data from config
	defaultBaselineData := utils.ParseBaselineData(config.DefaultBaselineData)

	// Reset the baseline dataset to default
	err := utils.CreateBaselineDataset(config.BaselineDatasetPath, defaultBaselineData)
	if err != nil {
		log.Printf("Error resetting baseline dataset: %v", err)
	} else {
		log.Println("Baseline dataset reset to default successfully.")
	}

	// Clear the deduplication store
	err = utils.ClearDeduplicationStore("global")
	if err != nil {
		log.Printf("Error clearing deduplication store: %v", err)
	} else {
		log.Println("Deduplication store cleared successfully.")
	}
}
