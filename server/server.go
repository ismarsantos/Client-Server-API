package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const QUOTATION_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const REQUEST_TIMEOUT = time.Duration(time.Millisecond * 200)
const PERSIST_TIMEOUT = time.Duration(time.Millisecond * 10)

type USDBRL struct {
	USDBRL Quotation `json:"USDBRL"`
}

type Quotation struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

var db *sql.DB

func handlerQuotation(w http.ResponseWriter, r *http.Request) {
	log.Println("Request quotation started")
	defer log.Println("Request quotation ended")

	ctx := r.Context()

	response, err := requestQuotation(ctx)
	if err != nil {
		log.Println("ERROR: API Request Timeout")
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	err = persistQuotation(ctx, db, response)
	if err != nil {
		log.Println("ERROR: Persist Timeout")
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func requestQuotation(ctx context.Context) (*Quotation, error) {
	ctx, timeout := context.WithTimeout(ctx, REQUEST_TIMEOUT)
	defer timeout()

	req, err := http.NewRequestWithContext(ctx, "GET", QUOTATION_URL, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var data USDBRL
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	return &data.USDBRL, nil
}

func createDatabase() {
	var err error

	db, err = sql.Open("sqlite3", "./quotations_db.sqlite")
	if err != nil {
		panic(err)
	}

	sql := `CREATE TABLE IF NOT EXISTS quotations(
            id INTEGER PRIMARY KEY,
            code TEXT,
            code_in TEXT,
            name TEXT,
            high TEXT,
            low TEXT,
            var_bid TEXT,
            pct_change TEXT,
            bid TEXT,
            ask TEXT,
            timestamp TEXT,
            create_date TEXT,
            persist_date DATETIME DEFAULT CURRENT_TIMESTAMP
          );`
	_, err = db.Exec(sql)

	if err != nil {
		panic(err)
	}
}

func persistQuotation(ctx context.Context, db *sql.DB, e *Quotation) error {
	ctx, timeout := context.WithTimeout(ctx, PERSIST_TIMEOUT)
	defer timeout()

	sql := `INSERT INTO quotations(
            code,
            code_in,
            name,
            high,
            low,
            var_bid,
            pct_change,
            bid,
            ask,
            timestamp,
            create_date
          ) VALUES(?,?,?,?,?,?,?,?,?,?,?)`

	stmt, err := db.PrepareContext(ctx, sql)
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec(e.Code, e.CodeIn, e.Name, e.High, e.Low, e.VarBid, e.PctChange, e.Bid, e.Ask, e.Timestamp, e.CreateDate)
	if err != nil {
		return err
	}

	log.Println("Quotation persisted in database!")

	return nil
}

func main() {
	log.Println("Server application started!")

	createDatabase()

	defer db.Close()

	http.HandleFunc("/cotacao", handlerQuotation)

	http.ListenAndServe(":8080", nil)
}
