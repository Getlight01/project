package vision

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)












const GigaChatClientID = "879cdfc4-9305-4dee-bcc4-a11b0f3b6ce2"
const GigaChatClientSecret = "ca3d2a0e-7240-40cf-b4e9-2a129fd407ce"
const GigaChatAuthKey = "ODc5Y2RmYzQtOTMwNS00ZGVlLWJjYzQtYTExYjBmM2I2Y2UyOjY5MjhiNmU1LWFkNWItNDU3OC1hZDk1LWI4MjNkMGYzYWI0Ng=="

type GigaChatClient struct {
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

func NewGigaChatClient() *GigaChatClient {
	clientID := os.Getenv("GIGACHAT_CLIENT_ID")
	if clientID == "" {
		clientID = GigaChatClientID
	}

	clientSecret := os.Getenv("GIGACHAT_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = GigaChatClientSecret
	}

	tr := &http.Transport{
		TLSClientConfig: createTLSConfig(),
	}

	return &GigaChatClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}

}

type StyleAnalysis struct {
	ItemType   string   `json:"item_type"`  // "свитер", "футболка", "джинсы"
	Style      string   `json:"style"`      // "оверсайз", "слим", "классический"
	Fit        string   `json:"fit"`        // "свободный", "облегающий", "прямой"
	Colors     []string `json:"colors"`     // ["красный", "тёмный"]
	Season     string   `json:"season"`     // "зима", "лето", "демисезон"
	Occasion   string   `json:"occasion"`   // "повседневный", "формальный"
	Confidence float64  `json:"confidence"` // 0.0 - 1.0
}

func (gc *GigaChatClient) AnalyzeImageStyle(imagePath string) (*StyleAnalysis, error) {

	if err := gc.ensureToken(); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	prompt := `Проанализируй предмет одежды на фото. Ответь строго в формате JSON:
{
  "item_type": "тип вещи (свитер/футболка/рубашка/джинсы/юбка/платье/куртка/пальто/обувь)",
  "style": "стиль (оверсайз/слим/классический/спортивный/кэжуал/формальный)",
  "fit": "посадка (свободный/облегающий/прямой/широкий)",
  "colors": ["основной цвет", "дополнительный цвет"],
  "season": "сезон (зима/весна/лето/осень/демисезон)",
  "occasion": "случай (повседневный/формальный/спортивный/вечерний)"
}`

	requestBody := map[string]interface{}{
		"model": "GigaChat-2",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
				"attachments": []map[string]string{
					{
						"file_name": "clothing.jpg",
						"file_data": imageBase64,
					},
				},
			},
		},
		"temperature": 0.1, // Низкая температура для точности
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://gigachat.devices.sberbank.ru/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+gc.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from GigaChat")
	}

	content := response.Choices[0].Message.Content
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON in response: %s", content)
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	var analysis StyleAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	analysis.Confidence = 0.9 // GigaChat обычно уверен в своих ответах

	return &analysis, nil
}

func (gc *GigaChatClient) ensureToken() error {

	if gc.accessToken != "" && time.Now().Before(gc.tokenExpiry) {
		return nil
	}




	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "GIGACHAT_API_PERS") // Personal scope for individual accounts

	fmt.Printf("[GigaChat] Request body: %s\n", data.Encode())

	req, err := http.NewRequest("POST", "https://ngw.devices.sberbank.ru/api/v2/oauth", strings.NewReader(data.Encode()))

	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	fmt.Printf("[GigaChat] Using AuthKey: %s...%s\n", GigaChatAuthKey[:20], GigaChatAuthKey[len(GigaChatAuthKey)-10:])
	req.Header.Set("Authorization", "Basic "+GigaChatAuthKey)

	tokenClient := &http.Client{
		Transport: gc.httpClient.Transport,
		Timeout:   5 * time.Second,
	}

	resp, err := tokenClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[GigaChat] Token request failed: status=%d\n", resp.StatusCode)
		fmt.Printf("[GigaChat] Response body: %s\n", string(body))
		return fmt.Errorf("token error (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[GigaChat] Token obtained successfully!\n")

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	gc.accessToken = tokenResp.AccessToken
	gc.tokenExpiry = time.Unix(tokenResp.ExpiresAt, 0)

	return nil
}

func generateRqUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func createTLSConfig() *tls.Config {

	certPath := "russian_trusted_root_ca.pem"

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	certData, err := os.ReadFile(certPath)
	if err == nil {
		if ok := rootCAs.AppendCertsFromPEM(certData); ok {
			fmt.Printf("[GigaChat] Russian Trusted Root CA certificate loaded successfully\n")
		} else {
			fmt.Printf("[GigaChat] Failed to append Russian Trusted Root CA certificate\n")
		}
	} else {
		fmt.Printf("[GigaChat] Certificate file not found at %s, using system certificates only\n", certPath)
	}



	return &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: true,
	}

}

type MockStyleAnalyzer struct{}

func NewMockStyleAnalyzer() *MockStyleAnalyzer {
	return &MockStyleAnalyzer{}
}

func (m *MockStyleAnalyzer) AnalyzeImageStyle(imagePath string) (*StyleAnalysis, error) {
	return &StyleAnalysis{
		ItemType:   "свитер",
		Style:      "оверсайз",
		Fit:        "свободный",
		Colors:     []string{"серый", "тёмный"},
		Season:     "зима",
		Occasion:   "повседневный",
		Confidence: 0.85,
	}, nil
}
