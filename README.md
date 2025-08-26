# QCM Application

A simple and customizable multiple-choice questionnaire (QCM) application built with Go backend and JavaScript frontend.

## Features

- **Dynamic QCM**: Load questions from a JSON file
- **Customizable**: Easily modify questions and answers
- **Real-time validation**: Immediate feedback on answers
- **Responsive design**: Works on desktop and mobile devices
- **Simple setup**: No external dependencies required

## Quick Start

1. Clone or download the project files
2. Navigate to the project directory
3. Run the application:

```bash
go run main.go
```

4. Open your browser and go to `http://localhost:8080`

## Customizing Your QCM

Edit the `qcm.json` file to create your own questionnaire:

```json
{
  "title": "My Custom Quiz",
  "questions": [
    {
      "question": "Your question here?",
      "options": ["Option 1", "Option 2", "Option 3", "Option 4"],
      "answer": 0
    }
  ]
}
```

### JSON Structure

- `title`: The title of your quiz
- `questions`: Array of question objects
  - `question`: The question text
  - `options`: Array of possible answers (4 options recommended)
  - `answer`: Index of the correct answer (0-based)

## Project Structure

```
.
├── main.go        # Go server implementation
├── qcm.json       # Questionnaire data (customize this)
├── index.html     # Main HTML page
├── style.css      # Styling
├── script.js      # Frontend logic
└── README.md      # This file
```