# OpenQuiz - Interactive Quiz Platform

A modern and customizable quiz application with theme selection, built with Go backend and JavaScript frontend.

## Features

- **Theme-based Quizzes**: Multiple quiz topics with dedicated themes
- **Interactive Interface**: Intuitive navigation and theme selection
- **Configurable Quizzes**: Choose number of questions and random order
- **Real-time Feedback**: Immediate validation and scoring
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Modern UI**: Clean, professional interface with smooth transitions
- **No Dependencies**: Simple setup with no external requirements

## Quick Start

1. Clone or download the project files
2. Navigate to the project directory
3. Run the application:

```bash
go run main.go
```

4. Open your browser and go to `http://localhost:8080`
5. Select a quiz theme from the available topics
6. Configure your quiz settings and start the assessment

## Available Quiz Themes

- **Web Development**: Test your HTML, CSS, JavaScript, and modern web development knowledge
- **General Culture**: Challenge yourself with diverse questions covering various topics

## Adding New Quiz Themes

### 1. Create a Theme File

Create a new JSON file in the `themes/` directory:

```json
{
  "title": "Your Theme Title",
  "description": "Description of your theme",
  "questions": [
    {
      "question": "Your question here?",
      "options": ["Option 1", "Option 2", "Option 3", "Option 4"],
      "answer": 0,
      "explanation": "Optional explanation of the correct answer"
    }
  ]
}
```

### 2. Register the Theme

Add your theme to `themes-list.json`:

```json
{
  "themes": [
    {
      "id": "your-theme-id",
      "title": "Your Theme Title",
      "description": "Brief description for the selection page",
      "icon": "ICON_NAME",
      "difficulty": "Beginner|Intermediate|Advanced",
      "questions_count": 10,
      "file": "your-theme-file.json"
    }
  ]
}
```

### JSON Structure

#### Theme File Structure
- `title`: The theme's display title
- `description`: Brief description of the theme
- `questions`: Array of question objects
  - `question`: The question text
  - `options`: Array of possible answers (4 options recommended)
  - `answer`: Index of the correct answer (0-based)
  - `explanation`: Optional explanation shown after answering

#### Theme List Structure
- `id`: Unique identifier for the theme
- `title`: Display name for theme selection
- `description`: Brief description shown in theme selection
- `icon`: Icon identifier (WEB, GLOBE, etc.)
- `difficulty`: Difficulty level indicator
- `questions_count`: Total number of questions in the theme
- `file`: Filename in the themes/ directory

## Project Structure

```
.
├── main.go                    # Go server implementation
├── themes-list.json           # Theme registry and metadata
├── themes/                    # Quiz theme files directory
│   ├── web-programming.json   # Web development quiz
│   └── general-culture.json   # General culture quiz
├── index.html                 # Main HTML application
├── style.css                  # Styling and theme colors
├── script.js                  # Frontend logic and navigation
└── README.md                  # This documentation
```

## Technical Features

- **Theme Management**: Dynamic loading of quiz themes from JSON files
- **Configurable Quizzes**: Users can select question count and randomization
- **Navigation System**: Smooth transitions between theme selection and quiz
- **Responsive UI**: Modern design that adapts to different screen sizes
- **Score Tracking**: Real-time scoring with detailed feedback
- **Professional Styling**: DevOps-inspired color scheme with tech aesthetics