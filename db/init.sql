-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Courses table
CREATE TABLE courses (
    uuid TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL
);

-- Lessons table
CREATE TABLE lessons (
    uuid TEXT PRIMARY KEY,
    course_uuid TEXT NOT NULL,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (course_uuid) REFERENCES courses(uuid) ON DELETE CASCADE,
    UNIQUE(course_uuid, slug)
);

-- Question types
CREATE TABLE question_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Insert question types
INSERT INTO question_types (name) VALUES 
    ('multiple_choice'),
    ('fill_blank');

-- Quizzes table (one per course)
CREATE TABLE quizzes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_uuid TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (course_uuid) REFERENCES courses(uuid) ON DELETE CASCADE,
    UNIQUE(course_uuid) -- Ensure one quiz per course
);

-- Questions table
CREATE TABLE questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quiz_id INTEGER NOT NULL,
    question_type_id INTEGER NOT NULL,
    question_text TEXT NOT NULL,
    explanation TEXT,
    answer_choices TEXT, -- JSON array for multiple choice, empty for fill_blank
    correct_answer TEXT NOT NULL, -- The correct answer
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    FOREIGN KEY (question_type_id) REFERENCES question_types(id)
);

-- User answers
CREATE TABLE user_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_id INTEGER NOT NULL,
    user_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    answered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
);

-- Sample data
INSERT INTO courses (uuid, title, slug) VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'Introduction to Programming', 'intro-programming');

INSERT INTO lessons (uuid, course_uuid, title, slug, content) VALUES 
    ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'What are Variables?', 'what-are-variables', 'Variables are containers for storing data values. In most programming languages, you declare a variable by giving it a name and optionally specifying its type. Variables can hold different types of data such as numbers, text, or boolean values.'),
    ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'Data Types', 'data-types', 'Programming languages have different data types to represent different kinds of information. Common data types include integers for whole numbers, floats for decimal numbers, strings for text, and booleans for true/false values.');

INSERT INTO quizzes (course_uuid) VALUES ('550e8400-e29b-41d4-a716-446655440001');

INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) VALUES 
    (1, 1, 'What is a variable in programming?', 'A variable is a storage location with an associated name that contains data.', 
     '["A container for storing data values", "A type of loop", "A programming language", "A function"]', 
     'A container for storing data values'),
    (1, 2, 'Complete the sentence: Variables can hold different types of _____ such as numbers, text, or boolean values.', 
     'Variables store data, which can be of various types.', '', 'data');
