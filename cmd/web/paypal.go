package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// PayPalAccessTokenResponse represents the OAuth2 token response from PayPal.
type PayPalAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// OrderRequest defines the structure for creating a PayPal order.
type OrderRequest struct {
	Intent        string         `json:"intent"`
	PurchaseUnits []PurchaseUnit `json:"purchase_units"`
}

type PurchaseUnit struct {
	Amount Amount `json:"amount"`
	Items  []Item `json:"items"`
}

type Amount struct {
	CurrencyCode string    `json:"currency_code"`
	Value        string    `json:"value"`
	Breakdown    Breakdown `json:"breakdown"`
}

type Breakdown struct {
	ItemTotal ItemTotal `json:"item_total"`
}

type ItemTotal struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

type Item struct {
	Name        string     `json:"name"`
	UnitAmount  UnitAmount `json:"unit_amount"`
	Quantity    string     `json:"quantity"`
	Description string     `json:"description"`
	Sku         string     `json:"sku"`
}

type UnitAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value        string `json:"value"`
}

// getPayPalAccessToken retrieves an access token from PayPal's OAuth2 API.
func getPayPalAccessToken(clientID, clientSecret string) (string, error) {
	url := "https://api.sandbox.paypal.com/v1/oauth2/token"
	data := "grant_type=client_credentials"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get access token: %s", body)
	}

	var tokenResp PayPalAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// createPayPalOrder sends a request to PayPal to create a new order.
func createPayPalOrder(accessToken string) (map[string]interface{}, error) {
	orderRequest := OrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []PurchaseUnit{
			{
				Amount: Amount{
					CurrencyCode: "USD",
					Value:        "100",
					Breakdown: Breakdown{
						ItemTotal: ItemTotal{
							CurrencyCode: "USD",
							Value:        "100",
						},
					},
				},
				Items: []Item{
					{
						Name: "T-Shirt",
						UnitAmount: UnitAmount{
							CurrencyCode: "USD",
							Value:        "100",
						},
						Quantity:    "1",
						Description: "Super Fresh Shirt",
						Sku:         "sku01",
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(orderRequest)
	if err != nil {
		return nil, err
	}

	url := "https://api.sandbox.paypal.com/v2/checkout/orders"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create order: %s", body)
	}

	var orderData map[string]interface{}
	if err := json.Unmarshal(body, &orderData); err != nil {
		return nil, err
	}

	return orderData, nil
}

// capturePayPalOrder captures a previously created PayPal order.
func CapturePayPalOrder(accessToken, orderID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.sandbox.paypal.com/v2/checkout/orders/%s/capture", orderID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Prefer", "return=minimal")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to capture order: %s", body)
	}

	var captureData map[string]interface{}
	if err := json.Unmarshal(body, &captureData); err != nil {
		return nil, err
	}

	return captureData, nil
}

// createOrderHandler handles the order creation request from the frontend.
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	// clientID := os.Getenv("PAYPAL_CLIENT_ID")
	// clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	clientID := "AdX1qGwJOf1KqCJmjKQWq2auv62FRtVIQzVuKfk0eX9NQ4UnNfhrq5e1VaEZ08g3FJu5sYoD-JjMqPgo"
	clientSecret := "EKU5oZLWq0ELhWRoX_6IwRFH6yg_QmU-Uhm55VhiCyafAsmsUwwiExbIdPmgjnwx-U7AN7ipPaskkxv_"

	if clientID == "" || clientSecret == "" {
		http.Error(w, "PayPal client credentials not set", http.StatusInternalServerError)
		return
	}

	accessToken, err := getPayPalAccessToken(clientID, clientSecret)
	if err != nil {
		http.Error(w, "Failed to get access token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	orderData, err := createPayPalOrder(accessToken)
	if err != nil {
		http.Error(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderData)
}

// captureOrderHandler handles the order capture request from the frontend.
func CaptureOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderID")

	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		http.Error(w, "PayPal client credentials not set", http.StatusInternalServerError)
		return
	}

	accessToken, err := getPayPalAccessToken(clientID, clientSecret)
	if err != nil {
		http.Error(w, "Failed to get access token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	captureData, err := CapturePayPalOrder(accessToken, orderID)
	if err != nil {
		http.Error(w, "Failed to capture order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(captureData)
}
