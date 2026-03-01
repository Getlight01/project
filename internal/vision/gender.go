package vision

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"fashion-look-generator/internal/models"
)

type GenderDetector struct {






}

func NewGenderDetector() *GenderDetector {
	return &GenderDetector{}
}


func (gd *GenderDetector) DetectGender(items []models.Item) (string, float64, error) {
	if len(items) == 0 {
		return "unisex", 0.0, nil
	}













	/*
	client := &http.Client{Timeout: 10 * time.Second}
	
	for _, item := range items {

		resp, err := client.Post(gd.apiEndpoint+"/analyze", ...)
		if err != nil {
			continue
		}
		
		var result struct {
			Gender    string  `json:"gender"`
			Confidence float64 `json:"confidence"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		...
	}
	*/


	return gd.mockDetectGender(items)
}

func (gd *GenderDetector) mockDetectGender(items []models.Item) (string, float64, error) {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	partCounts := make(map[models.BodyPart]int)
	for _, item := range items {
		partCounts[item.Part]++
	}

	maleScore := 0.0
	femaleScore := 0.0

	for _, item := range items {
		filename := strings.ToLower(filepath.Base(item.ImageURL))

		maleKeywords := []string{"men", "male", "man", "муж", "m_"}
		femaleKeywords := []string{"women", "female", "woman", "жен", "f_", "dress", "skirt"}
		
		for _, kw := range maleKeywords {
			if strings.Contains(filename, kw) {
				maleScore += 0.3
			}
		}
		
		for _, kw := range femaleKeywords {
			if strings.Contains(filename, kw) {
				femaleScore += 0.3
			}
		}
	}

	maleScore += rng.Float64() * 0.2
	femaleScore += rng.Float64() * 0.2

	if maleScore > femaleScore+0.2 {
		return "male", 0.6 + rng.Float64()*0.3, nil
	} else if femaleScore > maleScore+0.2 {
		return "female", 0.6 + rng.Float64()*0.3, nil
	}

	return "unisex", 0.5, nil
}

func (gd *GenderDetector) DetectFromImage(imagePath string) (string, float64, error) {




	/*
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return "", 0, err
	}
	defer client.Close()
	
	image := vision.NewImageFromReader(file)
	annotations, err := client.DetectLabels(ctx, image, nil, 10)
	if err != nil {
		return "", 0, err
	}

	for _, label := range annotations {
		if strings.Contains(strings.ToLower(label.Description), "men") {
			return "male", 0.8, nil
		}
		if strings.Contains(strings.ToLower(label.Description), "women") {
			return "female", 0.8, nil
		}
	}
	*/


	return "unisex", 0.5, nil
}

func (gd *GenderDetector) BatchDetect(items []models.Item) map[string]struct {
	Gender     string
	Confidence float64
} {
	results := make(map[string]struct {
		Gender     string
		Confidence float64
	})

	for _, item := range items {
		gender, confidence, _ := gd.DetectFromImage(item.ImageURL)
		results[item.ID] = struct {
			Gender     string
			Confidence float64
		}{gender, confidence}
	}

	return results
}

func GetConfidenceLevel(confidence float64) string {
	if confidence >= 0.8 {
		return "high"
	} else if confidence >= 0.6 {
		return "medium"
	}
	return "low"
}

func (gd *GenderDetector) SuggestGenderPrompt(currentGender string) string {
	if currentGender == "unisex" || currentGender == "" {
		return "Мы не смогли точно определить пол. Пожалуйста, выберите вручную для лучших рекомендаций."
	}
	return fmt.Sprintf("Определен пол: %s. Нажмите для изменения.", 
		map[string]string{"male": "мужской", "female": "женский"}[currentGender])
}
