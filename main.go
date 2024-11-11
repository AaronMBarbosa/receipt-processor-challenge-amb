package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	// Add this import for debugging
	"math"

	"github.com/google/uuid"
)

type Receipt struct {
	ID           string `json:"id"`
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
	Points       int    `json:"-"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

var receipts = make(map[string]Receipt)

func processReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var receipt Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a unique ID and calculate points
	receipt.ID = uuid.New().String()
	receipt.Points = calculatePoints(receipt)
	receipts[receipt.ID] = receipt

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": receipt.ID})
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract the ID from the URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 || pathParts[1] != "receipts" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	id := pathParts[2]

	// Look up the receipt by ID
	receipt, exists := receipts[id]
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"points": receipt.Points})
}

func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: Points for each alphanumeric character in retailer name
	retailerPoints := 0
	for _, char := range receipt.Retailer {
		if isAlphanumeric(char) {
			retailerPoints++
		}
	}
	points += retailerPoints
	//log.Printf("Retailer points: %d", retailerPoints)

	// Rule 2: 50 points if total is a round dollar amount
	if isRoundDollar(receipt.Total) {
		points += 50
		//log.Printf("Added 50 points for round dollar total")
	}

	// Rule 3: 25 points if total is multiple of 0.25
	if isMultipleOfQuarter(receipt.Total) {
		points += 25
		//log.Printf("Added 25 points for total as multiple of 0.25")
	}

	// Rule 4: 5 points for every two items
	itemPairPoints := (len(receipt.Items) / 2) * 5
	points += itemPairPoints
	//log.Printf("Item pair points: %d", itemPairPoints)

	// Rule 5: Additional points for item description length
	itemDescriptionPoints := 0
	for _, item := range receipt.Items {
		desc := strings.TrimSpace(item.ShortDescription)
		descLength := len(desc)
		//log.Printf("Item '%s' has trimmed length %d", desc, descLength)
		if descLength%3 == 0 {
			price, _ := parsePrice(item.Price)
			itemPoints := int(math.Ceil(price * 0.2))
			itemDescriptionPoints += itemPoints
			//log.Printf("Item description points for '%s': %d", item.ShortDescription, itemPoints)
		}
	}
	points += itemDescriptionPoints

	// Rule 6: 6 points if the purchase day is odd
	if isOddDay(receipt.PurchaseDate) {
		points += 6
		//log.Printf("Added 6 points for odd purchase day")
	}

	// Rule 7: 10 points if the time is between 2:00pm and 4:00pm
	if isInAfternoonRange(receipt.PurchaseTime) {
		points += 10
		//log.Printf("Added 10 points for purchase time between 2:00pm and 4:00pm")
	}

	//log.Printf("Total calculated points: %d", points)
	return points
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func isRoundDollar(total string) bool {
	return strings.HasSuffix(total, ".00")
}

func isMultipleOfQuarter(total string) bool {
	price, _ := parsePrice(total)
	return int(price*100)%25 == 0
}

func isOddDay(date string) bool {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	return parsedDate.Day()%2 != 0
}

func isInAfternoonRange(timeStr string) bool {
	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return false
	}
	return parsedTime.Hour() == 14
}

func parsePrice(priceStr string) (float64, error) {
	var price float64
	_, err := fmt.Sscanf(priceStr, "%f", &price)
	return price, err
}

func main() {
	http.HandleFunc("/receipts/process", processReceipt)
	http.HandleFunc("/receipts/", getPoints)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
