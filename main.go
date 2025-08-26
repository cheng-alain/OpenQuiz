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

// Structure pour une question QCM
type Question struct {
	ID       int         `json:"id"`
	Question string      `json:"question"`
	Options  []string    `json:"options"`
	Correct  interface{} `json:"correct"` // Peut être int ou []int
}

// Structure pour le QCM complet
type QCM struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

// Structure pour la réponse de l'API
type QCMResponse struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
	Total     int        `json:"total"`
}

// Structure pour vérifier une réponse
type Answer struct {
	QuestionID int         `json:"questionId"`
	Answer     interface{} `json:"answer"` // Peut être int ou []int
}

// Structure pour la réponse de vérification
type AnswerResult struct {
	Correct       bool        `json:"correct"`
	CorrectAnswer interface{} `json:"correctAnswer"` // Peut être int ou []int
}

var qcmData QCM

func main() {
	// Charger les données QCM depuis le fichier JSON
	loadQCMData()

	// Servir les fichiers statiques (CSS, JS)
	http.Handle("/style.css", http.FileServer(http.Dir("./")))
	http.Handle("/script.js", http.FileServer(http.Dir("./")))

	// Configuration des routes
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

	// Paramètres de requête
	countParam := r.URL.Query().Get("count")
	randomParam := r.URL.Query().Get("random")

	// Créer une copie des questions pour éviter de modifier l'original
	questions := make([]Question, len(qcmData.Questions))
	copy(questions, qcmData.Questions)

	// Mélanger les questions SEULEMENT si explicitement demandé
	if randomParam == "true" {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(questions), func(i, j int) {
			questions[i], questions[j] = questions[j], questions[i]
		})
	}

	// Limiter le nombre de questions
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

	// Trouver la question correspondante
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

	// Vérifier la réponse selon le type (simple ou multiple)
	var isCorrect bool

	switch correctAnswer := question.Correct.(type) {
	case float64:
		// Réponse simple
		userAnswer, ok := answer.Answer.(float64)
		if !ok {
			http.Error(w, "Format de réponse invalide", http.StatusBadRequest)
			return
		}
		isCorrect = int(userAnswer) == int(correctAnswer)

	case []interface{}:
		// Réponses multiples
		userAnswers, ok := answer.Answer.([]interface{})
		if !ok {
			http.Error(w, "Format de réponse invalide", http.StatusBadRequest)
			return
		}

		// Convertir les réponses correctes en slice d'int
		var correctInts []int
		for _, v := range correctAnswer {
			if num, ok := v.(float64); ok {
				correctInts = append(correctInts, int(num))
			}
		}

		// Convertir les réponses utilisateur en slice d'int
		var userInts []int
		for _, v := range userAnswers {
			if num, ok := v.(float64); ok {
				userInts = append(userInts, int(num))
			}
		}

		// Vérifier si les réponses correspondent exactement
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
