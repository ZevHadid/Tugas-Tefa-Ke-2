package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Province struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

var db *sql.DB

func ConnectDB() {
	var err error

	err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	user := os.Getenv("DB_USER")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s@tcp(%s:%s)/%s", user, host, port, dbname)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
}

func FetchProvinces() ([]Province, error) {
	resp, err := http.Get("https://emsifa.github.io/api-wilayah-indonesia/api/provinces.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiProvinces []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(body, &apiProvinces); err != nil {
		return nil, err
	}

	var provinces []Province
	for _, p := range apiProvinces {
		id, err := strconv.Atoi(p.ID)
		if err != nil {
			return nil, err
		}
		provinces = append(provinces, Province{ID: id, Code: p.ID, Name: p.Name})
	}

	return provinces, nil
}

func UpdateDB() error {
	provinces, err := FetchProvinces()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM provinces")
	if err != nil {
		return err
	}

	for _, p := range provinces {
		_, err := db.Exec("INSERT INTO provinces (id, code, name) VALUES (?, ?, ?)", p.ID, p.Code, p.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetProvincesHandler(w http.ResponseWriter, r *http.Request) {
	if err := UpdateDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, code, name FROM provinces")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var provinces []Province
	for rows.Next() {
		var p Province
		if err := rows.Scan(&p.ID, &p.Code, &p.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		provinces = append(provinces, p)
	}

	response := struct {
		Status  string     `json:"status"`
		Code    int        `json:"code"`
		Message string     `json:"message"`
		Data    []Province `json:"data"`
	}{
		Status:  "success",
		Code:    200,
		Message: "Successfully fetched provinces",
		Data:    provinces,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	ConnectDB()

	http.HandleFunc("/", GetProvincesHandler)

	http.ListenAndServe(":8080", nil)
}
