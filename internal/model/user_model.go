package model

import (
	"time"
)

type User struct {
	ID        int64       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Bio       string    `json:"bio"`
	Skills    []string  `json:"skills"` // skills를 배열로 처리
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateUserTable() error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) NOT NULL,
		email VARCHAR(50) NOT NULL,
		password VARCHAR(255) NOT NULL,
		bio VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)
	`
	_, err := DB.Exec(createTableQuery)
	return err
}

func CreateSkillsTable() error {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS skills (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT,
		skill VARCHAR(100),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)
	`
	_, err := DB.Exec(createTableQuery)
	return err
}

func InsertUser(user *User) (int64, error) {
	query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"

	result, err := DB.Exec(query, user.Username, user.Email, user.Password)
	if err != nil {
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func FindUserByUserName(username string) (*User, error) {
	query := "SELECT id, username, email, password, created_at, updated_at FROM users WHERE username = ?"

	var user User
	err := DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// find skills
	skills, err := GetSkillsByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	user.Skills = skills

	return &user, nil
}

func FindUserByEmail(email string) (*User, error) {
	query := "SELECT id, username, email, created_at, updated_at FROM users WHERE email = ?"

	var user User
	err := DB.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// find skills
	skills, err := GetSkillsByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	user.Skills = skills

	return &user, nil
}

func UpdateUser(user *User) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	query := "UPDATE users SET username = ?, email = ?, bio = ? WHERE id = ?"
	_, err = DB.Exec(query, user.Username, user.Email, user.Bio, user.ID)
	if err != nil {
		return err
	}

	// Delete Skills, Insert Skills
	err = DeleteAllSkill(user.ID)
	if err != nil {
		return err
	}
	if len(user.Skills) != 0 {
		for _, skill := range user.Skills {
			err = InsertSkill(user.ID, skill)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func InsertSkill(userID int64, skill string) error {
	query := "INSERT INTO skills (user_id, skill) VALUES (?, ?)"
	_, err := DB.Exec(query, userID, skill)
	return err
}

func DeleteSkill(userID int64, skill string) error {
	query := "DELETE FROM skills WHERE user_id = ? AND skill = ?"
	_, err := DB.Exec(query, userID, skill)
	return err
}

func DeleteAllSkill(userID int64) error {
	query := "DELETE FROM skills WHERE user_id = ?"
	_, err := DB.Exec(query, userID)
	return err
}

func GetSkillsByUserID(userID int64) ([]string, error) {
	query := "SELECT skill FROM skills WHERE user_id = ?"
	
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []string
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	return skills, nil
}
