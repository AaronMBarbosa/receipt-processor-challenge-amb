package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessReceiptEndpoint(t *testing.T) {
	// Create a shared router
	router := http.NewServeMux()
	router.HandleFunc("/receipts/process", processReceipt)
	router.HandleFunc("/receipts/", getPoints)

	// Example 1: Simple Receipt
	receipt1 := Receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Items: []Item{
			{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
			{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
			{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
			{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
			{ShortDescription: "Klarbrunn 12-PK 12 FL OZ", Price: "12.00"},
		},
		Total: "35.35",
	}
	expectedPoints1 := 28

	runProcessReceiptTest(t, router, receipt1, expectedPoints1)

	// Example 2: Receipt with Round Dollar Total
	receipt2 := Receipt{
		Retailer:     "M&M Corner Market",
		PurchaseDate: "2022-03-20",
		PurchaseTime: "14:33",
		Items: []Item{
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
		},
		Total: "9.00",
	}
	expectedPoints2 := 109

	runProcessReceiptTest(t, router, receipt2, expectedPoints2)

	// Example 3: Receipt with Complex Item Descriptions
	receipt3 := Receipt{
		Retailer:     "Walmart",
		PurchaseDate: "2023-05-15",
		PurchaseTime: "15:45",
		Items: []Item{
			{ShortDescription: "Organic Honey", Price: "10.00"},
			{ShortDescription: "Granola Bars - Mixed", Price: "3.33"},
			{ShortDescription: "Sparkling Water", Price: "1.25"},
		},
		Total: "14.58",
	}
	expectedPoints3 := 19
	runProcessReceiptTest(t, router, receipt3, expectedPoints3)
}

// Helper function to run individual receipt tests
func runProcessReceiptTest(t *testing.T, router *http.ServeMux, receipt Receipt, expectedPoints int) {
	// Convert receipt to JSON
	payload, _ := json.Marshal(receipt)

	// Create a new HTTP request for /receipts/process
	req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Use httptest to create a ResponseRecorder
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse the response to get the receipt ID
	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	receiptID := response["id"]

	// Test the /receipts/{id}/points endpoint with the returned receipt ID
	req, err = http.NewRequest("GET", "/receipts/"+receiptID+"/points", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new ResponseRecorder for the GET request
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code for the GET request
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GET /receipts/%s/points returned wrong status code: got %v want %v", receiptID, status, http.StatusOK)
	}

	// Parse the response for points
	var pointsResponse map[string]int
	if err := json.NewDecoder(rr.Body).Decode(&pointsResponse); err != nil {
		t.Errorf("Failed to decode points response: %v", err)
	}
	actualPoints := pointsResponse["points"]

	// Compare the expected points with the actual points
	if actualPoints != expectedPoints {
		t.Errorf("Points mismatch for receipt %s: got %v want %v", receiptID, actualPoints, expectedPoints)
	}
}
