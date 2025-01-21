package main

import (
	"atd/backend/handlers"
	"atd/backend/utils"
	"atd/config"
	"atd/types"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func main() {

	utils.InitializeLogger()

	config.LoadEnv()

	apiKey := config.GetOpenAIKey()

	client := openai.NewClient(apiKey)

	dbPath := "./data/database/sensitive_data.db"
	log.Printf("Initializing database at: %s", dbPath)
	db := utils.InitializeDB(dbPath)
	utils.PopulateDB(db)
	defer db.Close()

	r := gin.Default()

	store := cookie.NewStore([]byte("secure-secret-key"))
	r.Use(sessions.Sessions("user-session", store))

	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")

		if userID == nil {
			newSessionID := utils.GenerateUniqueID()
			log.Printf("Generated new Session-ID: %s", newSessionID)
			session.Set("userID", newSessionID)
			session.Save()

			c.SetCookie("Session-ID", newSessionID, 3600, "/", "", false, true)
		} else {
			existingSessionID := userID.(string)
			log.Printf("Retrieved existing Session-ID: %s", existingSessionID)

			c.Request.Header.Set("Session-ID", existingSessionID)
		}

		c.Next()
	})

	r.Static("/assets", config.StaticFilePath)
	r.LoadHTMLGlob(config.TemplatePath)

	modelList := config.VulnerabilityModels

	r.GET("/", func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("userID") == nil {
			session.Set("userID", utils.GenerateUniqueID())
			session.Save()
		}

		userID := session.Get("userID").(string)
		baselinePath := filepath.Join("./data/datasets", userID+"_baseline.csv")
		if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
			utils.InitializeBaselineDataset(userID, baselinePath)
		}

		data := types.TemplateData{
			Title:        config.AppName,
			Instructions: "Select a model to view specific instructions.",
			Models:       modelList,
		}
		c.HTML(http.StatusOK, "index.html", data)
	})

	r.POST("/submit", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")
		if userID == nil {
			log.Println("No user session found.")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
			return
		}

		userIDStr := userID.(string)
		baselinePath := filepath.Join("./data/datasets/", userIDStr+"_baseline.csv")

		selectedModel := c.PostForm("model")
		log.Printf("Received submission for model: %s by user: %s", selectedModel, userIDStr)

		var response string
		var err error

		switch selectedModel {
		case "Prompt Injection":
			log.Println("Handling Prompt Injection")
			prompt := c.PostForm("prompt")
			response, err = handlers.PromptInjectionHandler(prompt)
			if err != nil {
				log.Printf("Error in Prompt Injection: %v", err)
				response = "Error: " + err.Error()
			}

		case "Data Leakage":
			log.Println("Handling Data Leakage")
			file, header, fileErr := c.Request.FormFile("file")
			if fileErr == nil {
				content, err := ioutil.ReadAll(file)
				file.Close()
				if err == nil {
					response, err = handlers.ProcessUploadedFile(client, content, header.Filename)
					if err != nil {
						log.Printf("Error processing uploaded file: %v", err)
						response = "Error: " + err.Error()
					}
				} else {
					log.Printf("Error reading file: %v", err)
					response = "Error reading file: " + err.Error()
				}
			} else {
				log.Printf("File upload error: %v", fileErr)
				response = "Please upload a valid file."
			}

		case "Data Poisoning":
			log.Println("Handling Data Poisoning")
			if file, header, fileErr := c.Request.FormFile("file"); fileErr == nil {
				log.Printf("Uploaded file: %s", header.Filename)
				defer file.Close()
				handlers.PoisonDatasetHandler(c, baselinePath)
				return
			} else if fileErr != http.ErrMissingFile {
				log.Printf("Error uploading file: %v", fileErr)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
				return
			}

			prompt := c.PostForm("prompt")
			if prompt != "" {
				handlers.PoisonPromptHandler(c, baselinePath)
				return
			}

			log.Println("No file or prompt provided for Data Poisoning")
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file or prompt provided for Data Poisoning."})
			return

		case "Supply Chain":
			log.Println("Handling Supply Chain")

			session := sessions.Default(c)
			userID := session.Get("userID")
			if userID == nil {
				log.Println("No user session found.")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
				return
			}

			if file, header, fileErr := c.Request.FormFile("modell"); fileErr == nil {
				log.Printf("Uploaded Supply Chain model: %s", header.Filename)
				defer file.Close()

				response, err = handlers.ProcessSupplyChainFile(c, file, "malicious_chatbot.pth")
				if err != nil {
					log.Printf("Error processing Supply Chain model: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process Supply Chain model: " + err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": response})
				return
			} else if fileErr != http.ErrMissingFile {
				log.Printf("File upload error: %v", fileErr)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file: " + fileErr.Error()})
				return
			}

			log.Println("No file or prompt provided for Supply Chain")
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file or prompt provided for Supply Chain."})
			return

		default:
			log.Printf("Unsupported model selected: %s", selectedModel)
			response = "Unsupported model selected."
		}

		c.JSON(http.StatusOK, gin.H{"message": response})
	})

	r.POST("/reset-baseline", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")
		if userID == nil {
			log.Println("No user session found.")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
			return
		}

		userIDStr := userID.(string)
		baselinePath := filepath.Join("./data/datasets/", userIDStr+"_baseline.csv")

		log.Printf("Resetting baseline for user: %s", userIDStr)

		defaultBaselineData := config.DefaultBaselineData
		err := utils.ResetUserBaseline(baselinePath, defaultBaselineData)
		if err != nil {
			log.Printf("[ERROR] Failed to reset baseline dataset for user %s: %v", userIDStr, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset baseline dataset."})
			return
		}

		utils.ClearScoredRows(userIDStr)
		utils.ClearDatasetScores(userIDStr)
		utils.SetIntegrityScore(userIDStr, 1.0)
		utils.ClearDeduplicationStore(userIDStr)

		log.Printf("[INFO] Baseline and session state reset successfully for user: %s", userIDStr)
		c.JSON(http.StatusOK, gin.H{"message": "Baseline and session data reset successfully.", "integrityScore": 1.00})
	})

	r.POST("/submit-flag", func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("userID")
		if userID == nil {
			log.Println("No user session found.")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not initialized."})
			return
		}

		userIDStr := userID.(string)

		var requestData struct {
			Module string `json:"module"`
			Flag   string `json:"flag"`
		}
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload."})
			return
		}

		module := requestData.Module
		flag := requestData.Flag

		correctFlags := map[string]string{
			"module1": "ME{Pr0mp7_1nj3c7I0n}",
			"module2": "ME{Da74_L3aK4ge}",
			"module4": "ME{D4Ta_P01sOn1nG}",
			"module5": "ME{5upp1y_Ch41N}",
		}

		correctFlag, exists := correctFlags[module]
		if !exists {
			log.Printf("Invalid module: %s", module)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid module."})
			return
		}

		if flag == correctFlag {
			log.Printf("Correct flag submitted for %s by user: %s", module, userIDStr)
			utils.LogFlagSubmission(userIDStr, module, flag)

			c.JSON(http.StatusOK, gin.H{"message": "Congratulations! Your flag is correct."})
		} else {
			log.Printf("Incorrect flag submitted for %s by user: %s", module, userIDStr)
			c.JSON(http.StatusOK, gin.H{"message": "Incorrect flag. Please try again."})
		}
	})

	r.GET("/dataset-scores", handlers.GetDatasetScoresHandler)
	r.GET("/current-integrity-score", handlers.GetCurrentIntegrityScoreHandler)
	r.GET("/supply-chain/download", handlers.DownloadModelHandler)
	r.POST("/supply-chain/chat", handlers.ChatWithModelHandler)

	certFile := "./cert.pem"
	keyFile := "./key.pem"
	port := ":8443"

	log.Printf("Starting HTTPS server on port %s", port)
	err := r.RunTLS(port, certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}

}
