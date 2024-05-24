package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Quotation struct {
	Bid string `json:"bid"`
}

const QUOTATION_URL = "http://localhost:8080/cotacao"
const REQUEST_TIMEOUT = time.Duration(time.Millisecond * 300)

func requestQuotation(ctx context.Context) (*Quotation, error) {
	log.Println("Requesting quotation...")

	ctx, cancel := context.WithTimeout(ctx, REQUEST_TIMEOUT)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", QUOTATION_URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusGatewayTimeout {
		return nil, errors.New("quotation request timeout")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data Quotation
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	log.Println("Quotation Requested")

	return &data, nil
}

func saveQuotationFile(quotation Quotation) error {
	log.Println("Quotation in 'cotacao.txt' file...")

	f, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(f, "Dólar: %s\n", quotation.Bid)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("Init quotation request...")
	ctx := context.Background()

	quotation, err := requestQuotation(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = saveQuotationFile(*quotation)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("finished!")
}
