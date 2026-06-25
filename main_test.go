package main

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

// 1. ЮНИТ-ТЕСТ: Проверяем базовую логику структуры в памяти
func TestUnit_UserLogStructure(t *testing.T) {
	logEntry := UserLog{
		ID:     99,
		Login:  "test_user",
		Action: "Unit Тестирование",
	}

	if logEntry.Login != "test_user" {
		t.Errorf("Ожидался логин 'test_user', но получили '%s'", logEntry.Login)
	}
}

// 2. ИНТЕГРАЦИОННЫЙ ТЕСТ: Напрямую контактирует с базой в Docker
func TestIntegration_DatabaseInsertAndSelect(t *testing.T) {
	connStr := "host=localhost port=5432 user=myuser password=mypassword dbname=mydb sslmode=disable"
	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Не удалось подключиться к базе для теста: %v", err)
	}
	defer testDB.Close()

	var insertedID int
	err = testDB.QueryRow(
		"INSERT INTO users_log (login, action) VALUES ($1, $2) RETURNING id",
		"integration_test", "Проверка записи напрямую").Scan(&insertedID)

	if err != nil {
		t.Fatalf("Ошибка прямой записи в базу при тесте: %v", err)
	}

	var dbLogin string
	err = testDB.QueryRow("SELECT login FROM users_log WHERE id = $1", insertedID).Scan(&dbLogin)
	if err != nil {
		t.Fatalf("Ошибка чтения из базы при тесте: %v", err)
	}

	if dbLogin != "integration_test" {
		t.Errorf("Данные исказились! Ожидалось 'integration_test', пришло '%s'", dbLogin)
	}

	// Вычищаем тестовый мусор за собой
	_, _ = testDB.Exec("DELETE FROM users_log WHERE id = $1", insertedID)
}
