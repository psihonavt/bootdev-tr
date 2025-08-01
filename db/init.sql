-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Question types
CREATE TABLE question_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Insert question types
INSERT INTO question_types (name) VALUES 
    ('multiple_choice'),
    ('fill_blank');

-- Quizzes table (stores course UUID as reference)
CREATE TABLE quizzes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_uuid TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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

-- Sample data - Four quizzes with different themes
INSERT INTO quizzes (course_uuid) VALUES 
    ('european-capitals-uuid'),
    ('rdbms-fundamentals-uuid'), 
    ('meme-knowledge-uuid'),
    ('total-recall-uuid');

-- European Capitals Quiz (Quiz ID 1) - 10 questions
INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) VALUES 
    (1, 1, 'What is the capital of France?', 'Paris is the capital and most populous city of France.', 
     '["Paris", "Lyon", "Marseille", "Nice"]', 'Paris'),
    (1, 1, 'What is the capital of Germany?', 'Berlin is the capital and largest city of Germany.', 
     '["Munich", "Hamburg", "Berlin", "Frankfurt"]', 'Berlin'),
    (1, 2, 'The capital of Italy is _____.', 'Rome has been the capital of Italy since 1871.', '', 'Rome'),
    (1, 1, 'What is the capital of Spain?', 'Madrid is the capital and most populous city of Spain.', 
     '["Barcelona", "Madrid", "Valencia", "Seville"]', 'Madrid'),
    (1, 1, 'What is the capital of Portugal?', 'Lisbon is the capital and largest city of Portugal.', 
     '["Porto", "Braga", "Lisbon", "Coimbra"]', 'Lisbon'),
    (1, 2, 'The capital of Netherlands is _____.', 'Amsterdam is the capital, though The Hague is the seat of government.', '', 'Amsterdam'),
    (1, 1, 'What is the capital of Poland?', 'Warsaw is the capital and largest city of Poland.', 
     '["Krakow", "Warsaw", "Gdansk", "Wroclaw"]', 'Warsaw'),
    (1, 1, 'What is the capital of Sweden?', 'Stockholm is the capital of Sweden.', 
     '["Stockholm", "Gothenburg", "Malmö", "Uppsala"]', 'Stockholm'),
    (1, 2, 'The capital of Greece is _____.', 'Athens is the capital and largest city of Greece.', '', 'Athens'),
    (1, 1, 'What is the capital of Czech Republic?', 'Prague is the capital and largest city of Czech Republic.', 
     '["Brno", "Prague", "Ostrava", "Plzen"]', 'Prague');

-- RDBMS Quiz (Quiz ID 2) - 10 questions
INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) VALUES 
    (2, 1, 'What does SQL stand for?', 'SQL is a domain-specific language used in programming for managing relational databases.', 
     '["Structured Query Language", "Simple Query Language", "Standard Query Language", "Sequential Query Language"]', 'Structured Query Language'),
    (2, 1, 'Which SQL command is used to retrieve data?', 'SELECT is the primary command for querying data in SQL.', 
     '["GET", "FETCH", "SELECT", "RETRIEVE"]', 'SELECT'),
    (2, 2, 'A _____ key uniquely identifies each record in a database table.', 'Primary keys ensure each row can be uniquely identified.', '', 'primary'),
    (2, 1, 'What is a foreign key?', 'A foreign key is a field that refers to the primary key in another table.', 
     '["A key from another country", "A field that refers to the primary key in another table", "An encrypted key", "A backup key"]', 'A field that refers to the primary key in another table'),
    (2, 1, 'Which SQL command is used to add new data?', 'INSERT is used to add new records to a table.', 
     '["ADD", "INSERT", "CREATE", "PUT"]', 'INSERT'),
    (2, 2, 'The process of organizing data to reduce redundancy is called _____.', 'Normalization reduces data duplication and improves integrity.', '', 'normalization'),
    (2, 1, 'What does ACID stand for in database transactions?', 'ACID properties ensure reliable database transactions.', 
     '["Atomic, Consistent, Isolated, Durable", "All, Create, Insert, Delete", "Always, Complete, Individual, Done", "Accurate, Complete, Independent, Definite"]', 'Atomic, Consistent, Isolated, Durable'),
    (2, 1, 'Which SQL clause is used to filter results?', 'WHERE clause filters rows based on specified conditions.', 
     '["FILTER", "WHERE", "HAVING", "IF"]', 'WHERE'),
    (2, 2, 'A database _____ is a collection of related tables.', 'A schema defines the structure and organization of a database.', '', 'schema'),
    (2, 1, 'What is an index in a database?', 'An index improves query performance by creating shortcuts to data.', 
     '["A table of contents", "A data structure that improves query performance", "A backup copy", "A user permission"]', 'A data structure that improves query performance');

-- Meme Knowledge Quiz (Quiz ID 3) - 10 questions
INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) VALUES 
    (3, 1, 'Which meme features a dog surrounded by fire saying "This is fine"?', 'This meme represents staying calm in chaotic situations.', 
     '["Grumpy Cat", "Distracted Boyfriend", "This is Fine Dog", "Drake Pointing"]', 'This is Fine Dog'),
    (3, 2, 'Complete the meme: "One does not simply _____ into Mordor"', 'This Boromir meme is from Lord of the Rings.', '', 'walk'),
    (3, 1, 'What does "Based" mean in internet slang?', 'Based means being authentic and not caring about others opinions.', 
     '["Fake", "Authentic and not caring about others opinions", "Located at a base", "Mathematical foundation"]', 'Authentic and not caring about others opinions'),
    (3, 1, 'Which meme shows a man looking back at another woman while his girlfriend looks disapproving?', 'This meme represents being tempted by alternatives.', 
     '["Distracted Boyfriend", "Hide the Pain Harold", "Woman Yelling at Cat", "Drake Pointing"]', 'Distracted Boyfriend'),
    (3, 2, 'Complete the phrase: "It''s over 9000!" comes from _____.', 'This famous line is from the Dragon Ball Z anime.', '', 'Dragon Ball Z'),
    (3, 1, 'What does "sus" mean in Among Us and internet culture?', 'Sus is short for suspicious, popularized by Among Us.', 
     '["Super", "Suspicious", "Support", "System"]', 'Suspicious'),
    (3, 1, 'Which meme features a cat at a dinner table with vegetables?', 'Woman Yelling at Cat became popular for expressing disagreements.', 
     '["Keyboard Cat", "Nyan Cat", "Woman Yelling at Cat", "Business Cat"]', 'Woman Yelling at Cat'),
    (3, 2, 'The "_____ Guy" meme features a man with a forced smile hiding pain.', 'Hide the Pain Harold represents masking discomfort with a smile.', '', 'Hide the Pain'),
    (3, 1, 'What does "POV" stand for in TikTok and social media?', 'POV means Point of View, used to set up scenarios.', 
     '["Point of View", "Power of Victory", "Post of Value", "Part of Video"]', 'Point of View'),
    (3, 2, 'Complete the meme: "But wait, there''s _____!"', 'This phrase became popular from infomercials and is used ironically.', '', 'more');

-- Total Recall Quiz (Quiz ID 4) - 15 questions
INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) VALUES 
    (4, 1, 'What''s the name of Arnold''s character who may or may not be dreaming?', 'Douglas Quaid is our possibly-dreaming protagonist.', 
     '["Douglas Quaid", "John Matrix", "Dutch Schaefer", "Ben Richards"]', 'Douglas Quaid'),
    (4, 1, 'On which planet does most of the action take place?', 'Mars is where all the mutant fun happens!', 
     '["Venus", "Mars", "Jupiter", "Uranus"]', 'Mars'),
    (4, 1, 'What does Quaid originally do for a living?', 'He''s just a regular construction worker... or is he?', 
     '["Construction worker", "Secret agent", "Taxi driver", "Chef"]', 'Construction worker'),
    (4, 1, 'What''s the name of the memory implantation company?', 'Rekall - where your wildest dreams become fake memories!', 
     '["Rekall", "MemCorp", "DreamTech", "MindBenders"]', 'Rekall'),
    (4, 1, 'Who plays Quaid''s wife Lori?', 'Sharon Stone before she became famous for crossing her legs.', 
     '["Sharon Stone", "Jamie Lee Curtis", "Sigourney Weaver", "Linda Hamilton"]', 'Sharon Stone'),
    (4, 1, 'What''s special about the three-breasted woman?', 'She''s got... well, three breasts. Mars really changed the dress code.', 
     '["She has three arms", "She has three breasts", "She has three eyes", "She has three heads"]', 'She has three breasts'),
    (4, 1, 'What happens when you remove your helmet on Mars?', 'Your eyes pop out like a cartoon character - very practical!', 
     '["You turn blue", "Your eyes bulge out", "You grow gills", "You become invisible"]', 'Your eyes bulge out'),
    (4, 1, 'Who''s the leader of the Mars resistance?', 'Kuato - the guy who really puts the "guts" in leadership.', 
     '["Kuato", "Cohaagen", "Benny", "Richter"]', 'Kuato'),
    (4, 1, 'Where does Kuato literally live?', 'In his brother''s stomach - talk about family closeness!', 
     '["In a cave", "In his brother''s stomach", "Underground", "In a spaceship"]', 'In his brother''s stomach'),
    (4, 1, 'What does Benny drive around Mars?', 'A taxi cab - because even on Mars, you need reliable transportation.', 
     '["A hover car", "A taxi cab", "A motorcycle", "A bulldozer"]', 'A taxi cab'),
    (4, 1, 'What famous Arnold line does he say before shooting someone?', 'Consider that a divorce - harsh but effective!', 
     '["I''ll be back", "Consider that a divorce", "Get to the chopper", "Come with me if you want to live"]', 'Consider that a divorce'),
    (4, 1, 'What color is the sky on Mars in the movie?', 'Red, like a permanent sunset (or a really bad air quality day).', 
     '["Blue", "Purple", "Green", "Red"]', 'Red'),
    (4, 1, 'How does Quaid disguise himself?', 'As a fat woman - because nothing says "subtle" like Arnold in drag.', 
     '["As a robot", "As a fat woman", "As an alien", "As a child"]', 'As a fat woman'),
    (4, 1, 'What starts the Mars atmosphere machine?', 'An alien reactor - because aliens always leave the best toys behind.', 
     '["A nuclear bomb", "An alien reactor", "Solar panels", "Wind power"]', 'An alien reactor'),
    (4, 1, 'What''s the twist about whether it''s all a dream?', 'The movie never tells us - it''s the ultimate "choose your own adventure."', 
     '["It''s definitely real", "It''s definitely a dream", "We never find out for sure", "It''s a simulation"]', 'We never find out for sure');
