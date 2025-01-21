package utils

import (
	"atd/config"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type SessUtilState struct {
	UploadedFileHashes map[string]bool
	IntegrityScore     float64
	ScoredRows         map[string]struct{}
}

var (
	sessionStore   = make(map[string]*SessUtilState)
	sessionStoreMu sync.Mutex
)

func GetSessionState(sessionID string) *SessUtilState {
	sessionStoreMu.Lock()
	defer sessionStoreMu.Unlock()

	if state, exists := sessionStore[sessionID]; exists {
		return state
	}
	state := &SessUtilState{
		UploadedFileHashes: make(map[string]bool),
		IntegrityScore:     1.0,
		ScoredRows:         make(map[string]struct{}), // Initialize ScoredRows
	}
	sessionStore[sessionID] = state
	return state
}

func InitializeBaselineDataset(sessionID string, baselinePath string) {
	log.Println("Checking and resetting baseline dataset...")

	defaultBaselineData := ParseBaselineData(config.DefaultBaselineData)

	err := CreateBaselineDataset(baselinePath, defaultBaselineData)
	if err != nil {
		log.Printf("Error resetting baseline dataset: %v", err)
	} else {
		log.Println("Baseline dataset reset to default successfully.")
	}

	err = ClearDeduplicationStore(sessionID)
	if err != nil {
		log.Printf("Error clearing deduplication store: %v", err)
	} else {
		log.Println("Deduplication store cleared successfully.")
	}
}

func CreateBaselineDataset(path string, defaultData [][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range defaultData {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func ParseBaselineData(data string) [][]string {
	var result [][]string
	for _, line := range strings.Split(data, "\n") {
		result = append(result, strings.Split(line, ","))
	}
	return result
}

func ParseCSV(file io.Reader) ([][]string, error) {
	reader := csv.NewReader(file)
	var result [][]string

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[WARN] Skipping row due to parsing error: %v", err)
			continue
		}
		result = append(result, record)
	}
	return result, nil
}

func LoadDataset(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func WriteDataset(data [][]string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.WriteAll(data)
}

func FormatDataset(data [][]string) string {
	var builder strings.Builder
	for _, row := range data {
		builder.WriteString(strings.Join(row, " | "))
		builder.WriteString("\n")
	}
	return builder.String()
}

func AlignDataset(uploadedData, baselineData [][]string) [][]string {
	if len(baselineData) == 0 {
		log.Printf("[WARN] Baseline dataset is empty; cannot align data.")
		return uploadedData
	}

	expectedFields := len(baselineData[0])
	var alignedData [][]string

	for _, row := range uploadedData {
		if len(row) < expectedFields {
			paddedRow := append(row, make([]string, expectedFields-len(row))...)
			alignedData = append(alignedData, paddedRow)
		} else if len(row) > expectedFields {
			truncatedRow := row[:expectedFields]
			alignedData = append(alignedData, truncatedRow)
		} else {
			alignedData = append(alignedData, row)
		}
	}

	log.Printf("[INFO] Uploaded data aligned to baseline format. Rows processed: %d", len(alignedData))
	return alignedData
}

func CalculateFileHash(file io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func LoadCSV(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file (%s): %v", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var data []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV file (%s): %v", filePath, err)
		}
		data = append(data, strings.Join(record, " "))
	}
	return data, nil
}

func SaveCSV(filePath string, data []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file (%s): %v", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range data {
		if err := writer.Write([]string{record}); err != nil {
			return fmt.Errorf("failed to write to CSV file (%s): %v", filePath, err)
		}
	}
	log.Printf("File (%s) saved successfully.", filePath)
	return nil
}

func ValidateCSV(file io.Reader) error {
	_, err := csv.NewReader(file).ReadAll()
	return err
}

func ClearDeduplicationStore(sessionID string) error {
	state := GetSessionState(sessionID)
	state.UploadedFileHashes = make(map[string]bool)
	log.Println("[INFO] Deduplication store reset successfully.")
	return nil
}

func IsDuplicateFile(sessionID, fileHash string) bool {
	state := GetSessionState(sessionID)
	_, exists := state.UploadedFileHashes[fileHash]
	return exists
}

func MarkFileAsUploaded(sessionID, fileHash string) error {
	state := GetSessionState(sessionID)
	state.UploadedFileHashes[fileHash] = true
	return nil
}

func RemoveDuplicateRows(uploadedData, baselineData [][]string) [][]string {
	existingRows := make(map[string]bool)
	for _, row := range baselineData {
		existingRows[strings.Join(row, ",")] = true
	}

	uniqueRows := [][]string{}
	for _, row := range uploadedData {
		rowKey := strings.Join(row, ",")
		if !existingRows[rowKey] {
			existingRows[rowKey] = true
			uniqueRows = append(uniqueRows, row)
		}
	}

	return uniqueRows
}

func ClearDatasetScores(sessionID string) {
	session := GetSessState(sessionID)
	session.DatasetScores = make(map[string]float64)
	log.Printf("[INFO] Dataset scores cleared for session %s.", sessionID)
}

func ResetUserBaseline(baselinePath, defaultData string) error {
	data := ParseBaselineData(defaultData)
	return WriteDataset(data, baselinePath)
}
