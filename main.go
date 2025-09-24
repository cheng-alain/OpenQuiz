package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Question struct {
	ID          int         `json:"id"`
	Question    string      `json:"question"`
	Options     []string    `json:"options"`
	Correct     interface{} `json:"correct"`
	Explanation string      `json:"explanation"`
}

type QCM struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

type Theme struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Icon           string `json:"icon"`
	Difficulty     string `json:"difficulty"`
	QuestionsCount int    `json:"questions_count"`
	File           string `json:"file"`
}

type ThemesList struct {
	Themes []Theme `json:"themes"`
}

type QCMResponse struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
	Total     int        `json:"total"`
}

type Answer struct {
	QuestionID int         `json:"questionId"`
	Answer     interface{} `json:"answer"`
}

type AnswerResult struct {
	Correct       bool        `json:"correct"`
	CorrectAnswer interface{} `json:"correctAnswer"`
	Explanation   string      `json:"explanation"`
}

var qcmData QCM
var themesList ThemesList

func main() {
	loadThemes()

	host := getEnv("HOST", "0.0.0.0")
	port := getEnv("PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	http.Handle("/style.css", http.FileServer(http.Dir("./")))
	http.Handle("/script.js", http.FileServer(http.Dir("./")))

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/api/themes", getThemes)
	http.HandleFunc("/api/qcm", getQCM)
	http.HandleFunc("/api/check", checkAnswer)

	fmt.Printf("QCM server started at http://%s\n", addr)
	fmt.Printf("Local access: http://localhost:%s\n", port)
	fmt.Printf("Network access: http://%s:%s\n", getLocalIP(), port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func loadThemes() {
	data, err := ioutil.ReadFile("themes/themes-list.json")
	if err != nil {
		log.Fatal("Error reading themes-list.json file:", err)
	}

	err = json.Unmarshal(data, &themesList)
	if err != nil {
		log.Fatal("Error parsing themes JSON:", err)
	}

	fmt.Printf("Loaded %d themes\n", len(themesList.Themes))
}

func loadQCMData(themeID string) error {
	var themeFile string
	for _, theme := range themesList.Themes {
		if theme.ID == themeID {
			themeFile = theme.File
			break
		}
	}

	if themeFile == "" {
		return fmt.Errorf("theme not found: %s", themeID)
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("themes/%s", themeFile))
	if err != nil {
		return fmt.Errorf("error reading theme file: %v", err)
	}

	err = json.Unmarshal(data, &qcmData)
	if err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	var validQuestions []Question
	for _, question := range qcmData.Questions {
		questionText := strings.TrimSpace(question.Question)
		if questionText != "" && question.ID > 0 {
			validQuestions = append(validQuestions, question)
		}
	}
	qcmData.Questions = validQuestions

	fmt.Printf("QCM loaded: %s with %d questions\n", qcmData.Title, len(qcmData.Questions))
	return nil
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func getThemes(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(themesList)
}

func getQCM(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	themeID := r.URL.Query().Get("theme")
	if themeID == "" {
		http.Error(w, "Theme parameter is required", http.StatusBadRequest)
		return
	}

	err := loadQCMData(themeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	countParam := r.URL.Query().Get("count")
	randomParam := r.URL.Query().Get("random")

	questions := make([]Question, len(qcmData.Questions))
	copy(questions, qcmData.Questions)

	if randomParam == "true" {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(questions), func(i, j int) {
			questions[i], questions[j] = questions[j], questions[i]
		})
	}

	if countParam != "" {
		count, err := strconv.Atoi(countParam)
		if err == nil && count > 0 && count < len(questions) {
			questions = questions[:count]
		}
	}

	response := QCMResponse{
		Title:     qcmData.Title,
		Questions: questions,
		Total:     len(questions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func checkAnswer(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var answer Answer
	err := json.NewDecoder(r.Body).Decode(&answer)
	if err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	var question Question
	found := false
	for _, q := range qcmData.Questions {
		if q.ID == answer.QuestionID {
			question = q
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	var isCorrect bool

	switch correctAnswer := question.Correct.(type) {
	case float64:
		userAnswer, ok := answer.Answer.(float64)
		if !ok {
			http.Error(w, "Invalid answer format", http.StatusBadRequest)
			return
		}
		isCorrect = int(userAnswer) == int(correctAnswer)

	case []interface{}:
		userAnswers, ok := answer.Answer.([]interface{})
		if !ok {
			http.Error(w, "Invalid answer format", http.StatusBadRequest)
			return
		}

		var correctInts []int
		for _, v := range correctAnswer {
			if num, ok := v.(float64); ok {
				correctInts = append(correctInts, int(num))
			}
		}

		var userInts []int
		for _, v := range userAnswers {
			if num, ok := v.(float64); ok {
				userInts = append(userInts, int(num))
			}
		}

		isCorrect = len(userInts) == len(correctInts)
		if isCorrect {
			for _, userInt := range userInts {
				found := false
				for _, correctInt := range correctInts {
					if userInt == correctInt {
						found = true
						break
					}
				}
				if !found {
					isCorrect = false
					break
				}
			}
		}

	default:
		http.Error(w, "Invalid question format", http.StatusBadRequest)
		return
	}

	result := AnswerResult{
		Correct:       isCorrect,
		CorrectAnswer: question.Correct,
		Explanation:   question.Explanation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
