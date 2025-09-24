let questions = [];
let currentQuestionIndex = 0;
let userAnswers = {};
let score = 0;
let totalQuestions = 0;
let wrongAnswers = [];
let maxQuestionsAvailable = 0;

window.addEventListener('DOMContentLoaded', async function() {
    const maxQuestionsElement = document.getElementById('maxQuestions');
    if (!maxQuestionsElement) {
        console.error('Element maxQuestions not found');
        return;
    }
    
    try {
        const response = await fetch('/api/qcm');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        maxQuestionsAvailable = data.questions.length;
        
        maxQuestionsElement.textContent = `(max: ${maxQuestionsAvailable})`;
        
        const questionCountInput = document.getElementById('questionCount');
        if (questionCountInput) {
            questionCountInput.max = maxQuestionsAvailable;
        }
        
    } catch (error) {
        console.error('Error loading total questions:', error);
        maxQuestionsElement.textContent = '(max: ?)';
    }
});

async function startQuiz() {
    const questionCount = document.getElementById('questionCount').value;
    const randomOrder = document.getElementById('randomOrder').checked;
    
    document.getElementById('controls').style.display = 'none';
    document.getElementById('loading').style.display = 'block';

    try {
        let url = `/api/qcm?count=${questionCount}`;
        if (randomOrder) {
            url += '&random=true';
        }
        
        const response = await fetch(url);
        const data = await response.json();
        
        questions = data.questions;
        totalQuestions = data.total;
        
        document.getElementById('loading').style.display = 'none';
        document.getElementById('quizContainer').style.display = 'block';
        
        resetQuizState();
        displayQuestion();
    } catch (error) {
        console.error('Error:', error);
        alert('Error while loading the quiz');
        resetQuiz();
    }
}

function resetQuizState() {
    currentQuestionIndex = 0;
    userAnswers = {};
    score = 0;
    wrongAnswers = [];
    updateUI();
}

function displayQuestion() {
    if (currentQuestionIndex >= questions.length) {
        showResults();
        return;
    }

    const question = questions[currentQuestionIndex];
    
    document.getElementById('questionText').textContent = question.question;
    document.getElementById('questionCounter').textContent = `Question ${currentQuestionIndex + 1}/${totalQuestions}`;
    
    const instructionElement = document.getElementById('questionInstruction');
    if (Array.isArray(question.correct) && question.correct.length > 1) {
        instructionElement.textContent = `(Select ${question.correct.length} answers)`;
        instructionElement.style.display = 'block';
    } else {
        instructionElement.style.display = 'none';
    }
    
    const optionsContainer = document.getElementById('optionsContainer');
    optionsContainer.innerHTML = '';

    question.options.forEach((option, index) => {
        const optionElement = document.createElement('div');
        optionElement.className = 'option';
        optionElement.innerHTML = `
            <div class="option-letter">${String.fromCharCode(65 + index)}</div>
            <span>${option}</span>
        `;
        optionElement.onclick = () => selectOption(index);
        
        const selectedAnswers = userAnswers[question.id] || [];
        if (selectedAnswers.includes(index)) {
            optionElement.classList.add('selected');
            optionElement.querySelector('.option-letter').classList.add('selected');
        }
        
        optionsContainer.appendChild(optionElement);
    });

    updateUI();
    updateProgress();
}

function selectOption(optionIndex) {
    const question = questions[currentQuestionIndex];
    const isMultipleChoice = Array.isArray(question.correct) && question.correct.length > 1;
    
    if (!userAnswers[question.id]) {
        userAnswers[question.id] = isMultipleChoice ? [] : null;
    }
    
    if (isMultipleChoice) {
        const selectedAnswers = userAnswers[question.id];
        const optionAlreadySelected = selectedAnswers.includes(optionIndex);
        
        if (optionAlreadySelected) {
            userAnswers[question.id] = selectedAnswers.filter(index => index !== optionIndex);
        } else {
            if (selectedAnswers.length < question.correct.length) {
                userAnswers[question.id].push(optionIndex);
            }
        }
    } else {
        userAnswers[question.id] = optionIndex;
    }
    
    const options = document.querySelectorAll('.option');
    options.forEach((option, index) => {
        const optionLetter = option.querySelector('.option-letter');
        option.classList.remove('selected');
        optionLetter.classList.remove('selected');
        
        if (isMultipleChoice) {
            if (userAnswers[question.id].includes(index)) {
                option.classList.add('selected');
                optionLetter.classList.add('selected');
            }
        } else {
            if (userAnswers[question.id] === index) {
                option.classList.add('selected');
                optionLetter.classList.add('selected');
            }
        }
    });
    
    updateUI();
}

async function nextQuestion() {
    const question = questions[currentQuestionIndex];
    
    if (userAnswers[question.id] !== undefined) {
        try {
            const response = await fetch('/api/check', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    questionId: question.id,
                    answer: userAnswers[question.id]
                })
            });
            
            const result = await response.json();
            
            if (result.correct) {
                score++;
            } else {
                const question = questions[currentQuestionIndex];
                const isMultipleChoice = Array.isArray(question.correct) && question.correct.length > 1;
                
                let userAnswerText, correctAnswerText;
                
                if (isMultipleChoice) {
                    const userSelectedIndexes = userAnswers[question.id] || [];
                    const correctIndexes = question.correct;
                    
                    userAnswerText = userSelectedIndexes.length > 0 
                        ? userSelectedIndexes.map(i => question.options[i]).join(', ')
                        : 'No answer';
                    correctAnswerText = correctIndexes.map(i => question.options[i]).join(', ');
                } else {
                    userAnswerText = userAnswers[question.id] !== undefined 
                        ? question.options[userAnswers[question.id]]
                        : 'No answer';
                    correctAnswerText = question.options[result.correctAnswer];
                }
                
                wrongAnswers.push({
                    question: question.question,
                    userAnswer: userAnswerText,
                    correctAnswer: correctAnswerText,
                    options: question.options,
                    isMultipleChoice: isMultipleChoice,
                    explanation: result.explanation || ''
                });
            }
            
            showCorrection(result);
            
            setTimeout(() => {
                currentQuestionIndex++;
                displayQuestion();
            }, 1000);
            
        } catch (error) {
            console.error('Error:', error);
            currentQuestionIndex++;
            displayQuestion();
        }
    }
}

function showCorrection(result) {
    const question = questions[currentQuestionIndex];
    const isMultipleChoice = Array.isArray(question.correct) && question.correct.length > 1;
    const options = document.querySelectorAll('.option');
    
    options.forEach((option, index) => {
        const optionLetter = option.querySelector('.option-letter');
        
        if (isMultipleChoice) {
            const correctAnswers = question.correct;
            const userAnswers_current = userAnswers[question.id] || [];
            
            if (correctAnswers.includes(index)) {
                option.classList.add('correct');
                optionLetter.classList.add('correct');
            } else if (userAnswers_current.includes(index)) {
                option.classList.add('incorrect');
                optionLetter.classList.add('incorrect');
            }
        } else {
            if (index === result.correctAnswer) {
                option.classList.add('correct');
                optionLetter.classList.add('correct');
            } else if (index === userAnswers[question.id] && !result.correct) {
                option.classList.add('incorrect');
                optionLetter.classList.add('incorrect');
            }
        }
        
        option.onclick = null;
    });
    
    updateScore();
}

function updateUI() {
    const question = questions[currentQuestionIndex];
    const isMultipleChoice = Array.isArray(question?.correct) && question.correct.length > 1;
    let hasAnswer = false;
    if (question) {
        if (isMultipleChoice) {
            const selectedAnswers = userAnswers[question.id] || [];
            hasAnswer = selectedAnswers.length === question.correct.length;
        } else {
            hasAnswer = userAnswers[question.id] !== undefined;
        }
    }
    
    const nextBtn = document.getElementById('nextBtn');
    if (nextBtn) nextBtn.disabled = !hasAnswer;
}

function updateProgress() {
    const progress = ((currentQuestionIndex + 1) / totalQuestions) * 100;
    document.getElementById('progressFill').style.width = progress + '%';
}

function updateScore() {
    document.getElementById('scoreDisplay').textContent = `Score: ${score}/${totalQuestions}`;
}

function showResults() {
    document.getElementById('quizContainer').style.display = 'none';
    document.getElementById('results').style.display = 'block';
    
    const percentage = Math.round((score / totalQuestions) * 100);
    document.getElementById('finalScore').textContent = `${score}/${totalQuestions} (${percentage}%)`;
    
    let message = '';
    if (percentage >= 80) {
        message = '🏆 Excellent! You mastered the subject perfectly!';
    } else if (percentage >= 60) {
        message = '👍 Well done! A few revisions and you\'ll be perfect!';
    } else if (percentage >= 40) {
        message = '📚 Not bad, but there\'s still work to do!';
    } else {
        message = '💪 Don\'t get discouraged, keep learning!';
    }
    
    document.getElementById('resultMessage').textContent = message;
    
    displayWrongAnswers();
}

function displayWrongAnswers() {
    const wrongAnswersSection = document.getElementById('wrongAnswersSection');
    const wrongAnswersContainer = document.getElementById('wrongAnswersContainer');
    
    if (wrongAnswers.length === 0) {
        wrongAnswersSection.style.display = 'none';
    } else {
        wrongAnswersSection.style.display = 'block';
        wrongAnswersContainer.innerHTML = '';
        
        wrongAnswers.forEach((wrong, index) => {
            const wrongQuestionDiv = document.createElement('div');
            wrongQuestionDiv.className = 'wrong-question';
            wrongQuestionDiv.innerHTML = `
                <div class="wrong-question-text">${index + 1}. ${wrong.question}</div>
                <div class="answer-comparison">
                    <div class="user-answer">
                        <span class="answer-label">Your answer:</span>
                        <span>${wrong.userAnswer}</span>
                    </div>
                    <div class="correct-answer">
                        <span class="answer-label">Correct answer:</span>
                        <span>${wrong.correctAnswer}</span>
                    </div>
                </div>
                ${wrong.explanation ? `<div class="explanation"><strong>Explanation:</strong> ${wrong.explanation}</div>` : ''}
            `;
            wrongAnswersContainer.appendChild(wrongQuestionDiv);
        });
    }
}

function goHome() {
    if (confirm('Are you sure you want to quit the quiz?')) {
        resetQuiz();
    }
}

function resetQuiz() {
    document.getElementById('results').style.display = 'none';
    document.getElementById('quizContainer').style.display = 'none';
    document.getElementById('controls').style.display = 'block';
    
    questions = [];
    currentQuestionIndex = 0;
    userAnswers = {};
    score = 0;
    totalQuestions = 0;
    wrongAnswers = [];
}