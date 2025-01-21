package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"atd/backend/utils"
	"atd/config"

	"github.com/otiai10/gosseract/v2"
	"github.com/sashabaranov/go-openai"
)

func ProcessUploadedFile(client *openai.Client, content []byte, filename string) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	isImage := strings.HasSuffix(strings.ToLower(filename), ".png") || strings.HasSuffix(strings.ToLower(filename), ".jpg") || strings.HasSuffix(strings.ToLower(filename), ".jpeg")
	if !isImage {
		return "", fmt.Errorf("only image files (PNG, JPG, JPEG) are allowed")
	}

	ocrClient := gosseract.NewClient()
	defer ocrClient.Close()

	err := ocrClient.SetImageFromBytes(content)
	if err != nil {
		return "", fmt.Errorf("error reading image for OCR: %v", err)
	}

	text, err := ocrClient.Text()
	if err != nil {
		return "", fmt.Errorf("error extracting text from image: %v", err)
	}

	log.Printf("Extracted text for analysis: '%s'", text)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(`You are an AI security analyst. 
The following text was extracted from an uploaded image. 
Our application has a simulated SQL injection vulnerability that allows executing any SQL query.

Your task:
- Respond "YES" if the extracted text contains a valid SQL query.
- Respond "NO" otherwise.

Extracted text: "%s"`, text)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a security expert analyzing uploaded content for valid SQL queries.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:        10,
		Temperature:      0.0,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("Error contacting OpenAI API: %v", err)
		return "", fmt.Errorf("error contacting OpenAI API: %v", err)
	}

	if len(resp.Choices) > 0 {
		chatResponse := strings.TrimSpace(resp.Choices[0].Message.Content)
		log.Printf("ChatGPT Full Response: %s", chatResponse)

		if strings.Contains(strings.ToLower(chatResponse), "yes") {
			log.Println("Valid SQL query detected. Executing on the database.")

			query := strings.TrimSpace(strings.TrimPrefix(text, "'"))
			log.Printf("Executing SQL query: %s", query)

			db := utils.InitializeDB(config.GetSensitiveDatasetPath())
			defer db.Close()

			return utils.ExecuteQuery(db, query)
		}
	}

	return fmt.Sprintf("Image '%s' has been successfully uploaded to the database.", filename), nil
}

func DataLeakageHandler(prompt string) (string, error) {
	return "Please upload your photo to the database for further analysis.", nil
}
