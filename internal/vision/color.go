package vision

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)











const GoogleVisionAPIKey = "AIzaSyВАШ_КЛЮЧ_ЗДЕСЬ"

type GoogleVisionClient struct {
	apiKey string
	client *http.Client
}

func NewGoogleVisionClient() *GoogleVisionClient {
	apiKey := os.Getenv("GOOGLE_VISION_API_KEY")
	if apiKey == "" {
		apiKey = GoogleVisionAPIKey
	}

	return &GoogleVisionClient{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

type ColorInfo struct {
	Color         string  `json:"color"`
	Score         float64 `json:"score"`
	PixelFraction float64 `json:"pixelFraction"`
}

type ImageProperties struct {
	DominantColors []ColorInfo `json:"dominantColors"`
}

type VisionResponse struct {
	Responses []struct {
		ImageProperties struct {
			DominantColors struct {
				Colors []struct {
					Color struct {
						Red   float64 `json:"red"`
						Green float64 `json:"green"`
						Blue  float64 `json:"blue"`
					} `json:"color"`
					Score         float64 `json:"score"`
					PixelFraction float64 `json:"pixelFraction"`
				} `json:"colors"`
			} `json:"dominantColors"`
		} `json:"imagePropertiesAnnotation"`
		LabelAnnotations []struct {
			Description string  `json:"description"`
			Score       float64 `json:"score"`
		} `json:"labelAnnotations"`
	} `json:"responses"`
}

func (c *GoogleVisionClient) AnalyzeImage(imagePath string) (*ImageProperties, []string, error) {

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(imageData)

	requestBody := map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"image": map[string]interface{}{
					"content": encoded,
				},
				"features": []map[string]interface{}{
					{
						"type":       "IMAGE_PROPERTIES",
						"maxResults": 10,
					},
					{
						"type":       "LABEL_DETECTION",
						"maxResults": 10,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", c.apiKey)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("API error: %s", string(body))
	}

	var visionResp VisionResponse
	if err := json.Unmarshal(body, &visionResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(visionResp.Responses) == 0 {
		return nil, nil, fmt.Errorf("no response from Vision API")
	}

	response := visionResp.Responses[0]

	var colors []ColorInfo
	for _, c := range response.ImageProperties.DominantColors.Colors {
		colorName := rgbToColorName(c.Color.Red, c.Color.Green, c.Color.Blue)
		colors = append(colors, ColorInfo{
			Color:         colorName,
			Score:         c.Score,
			PixelFraction: c.PixelFraction,
		})
	}

	var labels []string
	for _, label := range response.LabelAnnotations {
		if label.Score > 0.7 { // Только уверенные метки
			labels = append(labels, label.Description)
		}
	}

	return &ImageProperties{DominantColors: colors}, labels, nil
}

func (c *GoogleVisionClient) ExtractDominantColors(imagePath string) ([]string, error) {
	props, _, err := c.AnalyzeImage(imagePath)
	if err != nil {
		return nil, err
	}

	var colors []string
	for _, c := range props.DominantColors {
		if c.Score > 0.1 { // Только значимые цвета
			colors = append(colors, c.Color)
		}
	}

	if len(colors) == 0 {
		return []string{"neutral"}, nil
	}

	return colors, nil
}

func DetectClothingType(labels []string) string {
	clothingTypes := map[string][]string{
		"top":       {"shirt", "t-shirt", "blouse", "sweater", "hoodie", "jacket", "coat"},
		"bottom":    {"pants", "jeans", "trousers", "skirt", "shorts"},
		"shoes":     {"shoes", "sneakers", "boots", "sandals", "footwear"},
		"outer":     {"jacket", "coat", "blazer", "cardigan"},
		"dress":     {"dress", "gown"},
		"accessory": {"bag", "handbag", "purse", "belt", "scarf", "hat"},
	}

	for _, label := range labels {
		labelLower := toLower(label)
		for itemType, keywords := range clothingTypes {
			for _, keyword := range keywords {
				if contains(labelLower, keyword) {
					return itemType
				}
			}
		}
	}

	return "unknown"
}

func rgbToColorName(r, g, b float64) string {

	red, green, blue := int(r), int(g), int(b)

	if red > 200 && green > 200 && blue > 200 {
		return "white"
	}
	if red < 50 && green < 50 && blue < 50 {
		return "black"
	}
	if abs(red-green) < 20 && abs(green-blue) < 20 && abs(red-blue) < 20 {
		if red > 150 {
			return "gray"
		}
		return "darkgray"
	}

	if red > green && red > blue {
		if green > 150 && blue < 100 {
			return "yellow"
		}
		if blue > 150 && green < 100 {
			return "magenta"
		}
		if green > 100 && blue > 100 {
			return "pink"
		}
		return "red"
	}
	if green > red && green > blue {
		if red > 150 && blue < 100 {
			return "yellow"
		}
		if blue > 150 {
			return "cyan"
		}
		return "green"
	}
	if blue > red && blue > green {
		if red > 150 {
			return "magenta"
		}
		if green > 150 {
			return "cyan"
		}
		return "blue"
	}

	return "neutral"
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

type MockColorAnalyzer struct{}

func NewMockColorAnalyzer() *MockColorAnalyzer {
	return &MockColorAnalyzer{}
}

func (m *MockColorAnalyzer) ExtractDominantColors(imagePath string) ([]string, error) {

	colors := []string{"blue", "white", "gray"}
	return colors, nil
}
