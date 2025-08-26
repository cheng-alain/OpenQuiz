package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Question struct {
	ID       int         `json:"id"`
	Question string      `json:"question"`
	Options  []string    `json:"options"`
	Correct  interface{} `json:"correct"`
}

type QCM struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
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
}

var qcmData QCM

func main() {
	loadQCMData()

	http.Handle("/style.css", http.FileServer(http.Dir("./")))
	http.Handle("/script.js", http.FileServer(http.Dir("./")))

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/api/qcm", getQCM)
	http.HandleFunc("/api/check", checkAnswer)

	fmt.Println("Serveur QCM démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadQCMData() {
	data, err := ioutil.ReadFile("qcm.json")
	if err != nil {
		log.Fatal("Erreur lors de la lecture du fichier qcm.json:", err)
	}

	err = json.Unmarshal(data, &qcmData)
	if err != nil {
		log.Fatal("Erreur lors du parsing JSON:", err)
	}

	fmt.Printf("QCM chargé: %s avec %d questions\n", qcmData.Title, len(qcmData.Questions))
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func getQCM(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
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
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var answer Answer
	err := json.NewDecoder(r.Body).Decode(&answer)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
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
		http.Error(w, "Question non trouvée", http.StatusNotFound)
		return
	}

	var isCorrect bool

	switch correctAnswer := question.Correct.(type) {
	case float64:
		userAnswer, ok := answer.Answer.(float64)
		if !ok {
			http.Error(w, "Format de réponse invalide", http.StatusBadRequest)
			return
		}
		isCorrect = int(userAnswer) == int(correctAnswer)

	case []interface{}:
		userAnswers, ok := answer.Answer.([]interface{})
		if !ok {
			http.Error(w, "Format de réponse invalide", http.StatusBadRequest)
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
		http.Error(w, "Format de question invalide", http.StatusBadRequest)
		return
	}

	result := AnswerResult{
		Correct:       isCorrect,
		CorrectAnswer: question.Correct,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
