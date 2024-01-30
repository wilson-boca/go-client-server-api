package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Start() {
	db, _ = sql.Open("sqlite3", "./quotations.db")
	defer db.Close()
	createTables()

	http.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", nil)
}

type Dolar struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Response struct {
	Bid string `json:"bid"`
}

func createTables() {
	createQuotationsTableSQL := `CREATE TABLE IF NOT EXISTS quotations (		
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"currency" TEXT,
		"value" FLOAT
	  );`
	statement, err := db.Prepare(createQuotationsTableSQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
}

func insertQuotation(ctx context.Context, currency string, value float32) error {
	log.Println("Inserting quotation record ...")
	insertQuotationSQL := `INSERT INTO quotations(currency, value) VALUES (?, ?)`
	statement, err := db.Prepare(insertQuotationSQL)
	if err != nil {
		return err
	}
	_, err = statement.ExecContext(ctx, currency, value)
	if err != nil {
		return err
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {

	log.Println("Request iniciada")
	defer log.Println("Request finalizada")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	// O desafio pede para setar esse valor para 10 milisegundos, mas na minha maquina nao foi suficiente
	ctxDb, cancelDb := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	defer cancelDb()

	dolar, err := GetCotation(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar: %v\n", err)
		return
	}
	res, _ := json.Marshal(dolar)
	fmt.Println(string(res))
	value, err := strconv.ParseFloat(dolar.Usdbrl.Bid, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao converter para float: %v\n", err)
		return
	}
	float := float32(value)
	err = insertQuotation(ctxDb, dolar.Usdbrl.Code, float)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar no banco: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"bid":` + dolar.Usdbrl.Bid + `}`))
}

func GetCotation(ctx context.Context) (*Dolar, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var d Dolar
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
