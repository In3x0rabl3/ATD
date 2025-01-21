package utils

import (
	"atd/config"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

// Session-based integrity management
var (
	SessStore   = make(map[string]*SessionState)
	SessStoreMu sync.Mutex
)

type SessionState struct {
	IntegrityScore float64
	ScoredRows     map[string]struct{}
	DatasetScores  map[string]float64
}

func GetSessState(sessionID string) *SessionState {
	SessStoreMu.Lock()
	defer SessStoreMu.Unlock()

	if state, exists := SessStore[sessionID]; exists {
		return state
	}

	state := &SessionState{
		IntegrityScore: 1.0,
		ScoredRows:     make(map[string]struct{}),
		DatasetScores:  make(map[string]float64),
	}
	SessStore[sessionID] = state
	return state
}

func GenerateRowHash(row []string) string {
	rowStr := strings.Join(row, ", ")
	hash := sha256.Sum256([]byte(rowStr))
	return hex.EncodeToString(hash[:])
}

func MarkRowAsScored(sessionID, hash string) {
	session := GetSessState(sessionID)
	session.ScoredRows[hash] = struct{}{}
}

func IsRowScored(sessionID, hash string) bool {
	session := GetSessState(sessionID)
	_, exists := session.ScoredRows[hash]
	return exists
}

func UpdateIntegrityScore(sessionID string, delta float64) {
	session := GetSessState(sessionID)
	session.IntegrityScore = math.Max(0, session.IntegrityScore+delta)
	log.Printf("[INFO] Integrity score updated for session %s: %.2f", sessionID, session.IntegrityScore)
}

func GetIntegrityScore(sessionID string) float64 {
	session := GetSessState(sessionID)
	if session == nil {
		log.Printf("[WARN] Session %s not found.", sessionID)
		return -1
	}
	return session.IntegrityScore
}

func ShouldInjectFlag(sessionID string) bool {
	return GetIntegrityScore(sessionID) < 0.25
}

func InjectFlag(sessionID string, baselineData [][]string, baselinePath string) error {
	flagRow := []string{"FLAG", "ME{D4Ta_P01sOn1nG}"}
	baselineData = append(baselineData, flagRow)
	log.Printf("[INFO] Flag injected for session %s", sessionID)
	return WriteDataset(baselineData, baselinePath)
}

func LogDatasetScore(sessionID, hash string, score float64) {
	session := GetSessState(sessionID)
	session.DatasetScores[hash] = score
	log.Printf("[INFO] Dataset scores for session %s: %+v", sessionID, session.DatasetScores)
}

func GetDatasetScores(sessionID string) map[string]float64 {
	session := GetSessState(sessionID)
	copy := make(map[string]float64)
	for hash, score := range session.DatasetScores {
		copy[hash] = score
	}
	return copy
}

func SetIntegrityScore(sessionID string, score float64) {
	session := GetSessState(sessionID)
	session.IntegrityScore = score
	log.Printf("[INFO] Integrity score set for session %s: %.2f", sessionID, score)
}

func ClearScoredRows(sessionID string) {
	session := GetSessState(sessionID)
	session.ScoredRows = make(map[string]struct{})
	log.Printf("[INFO] Scored rows cleared for session %s.", sessionID)
}

func EvaluateRowScore(sessionID string, row []string, baselineData [][]string) float64 {
	rowHash := GenerateRowHash(row)

	for _, baselineRow := range baselineData {
		if rowHash == GenerateRowHash(baselineRow) {
			log.Printf("[INFO] Row already exists in baseline dataset. Skipping scoring: \"%s\"", strings.Join(row, ", "))
			return 0.0
		}
	}

	if IsRowScored(sessionID, rowHash) {
		log.Printf("[INFO] Row already scored in session %s. Skipping: \"%s\"", sessionID, strings.Join(row, ", "))
		return 0.0
	}
	apiKey := config.GetOpenAIKey()
	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rowStr := strings.Join(row, ", ")

	prompt := fmt.Sprintf(`You are an AI expert specializing in detecting malicious entries in datasets. The baseline dataset is as follows:
	%s

	Evaluate the following row:
	"%s"

	Instructions:
	1. When scoring you will always score 0.0 to rows that are already in the current dataset.
	2. The row must have a cat fact or it will receive 0.0 score.
	3. You must be very strict when scoring rows.
	4. Only rows that have cat facts and could directly introduce false or inaccurate information will receive a score between 0.01 and 0.07.
	5. Do not filter or preprocess the row—only evaluate its likelihood of manipulating the dataset.

	Respond with:
	- A numeric score between 0.0 and 0.07, indicating how likely the row is to manipulate the baseline dataset or introduce false or inaccurate information.
	- No additional text, commentary, or explanation.

	Your response must be a single numeric value.`, formatDataset(baselineData), rowStr)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are a strict data analysis expert focused on detecting dataset poisoning."},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens:   10,
		Temperature: 0.0,
		TopP:        1.0,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("[ERROR] Chat API failed: %v", err)
		return 0.0
	}

	if len(resp.Choices) > 0 {
		response := strings.TrimSpace(resp.Choices[0].Message.Content)
		var score float64
		if _, err := fmt.Sscanf(response, "%f", &score); err == nil && score >= 0.0 && score <= 0.07 {
			MarkRowAsScored(sessionID, rowHash)
			log.Printf("[INFO] Row scored in session %s: \"%s\" - Score: %.2f", sessionID, rowStr, score)
			return score
		}
		log.Printf("[WARN] Invalid score in session %s for row \"%s\": \"%s\".", sessionID, rowStr, response)
	}

	return 0.0
}

func CalculateIntegrityScore(sessionID string, uploadedData, baselineData [][]string) float64 {
	var totalPenalty float64
	for _, row := range uploadedData {
		totalPenalty += EvaluateRowScore(sessionID, row, baselineData)
	}
	UpdateIntegrityScore(sessionID, -totalPenalty)
	currentScore := GetIntegrityScore(sessionID)
	log.Printf("[INFO] Integrity score calculated for sessionID: %s, New Score: %.2f, Total Penalty: %.2f", sessionID, currentScore, totalPenalty)
	return currentScore
}

func formatDataset(data [][]string) string {
	var sb strings.Builder
	for _, row := range data {
		sb.WriteString(strings.Join(row, ", "))
		sb.WriteString("\n")
	}
	return sb.String()
}
