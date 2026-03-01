package fashion

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"fashion-look-generator/internal/models"

	"github.com/google/uuid"
)

type FashionDB struct {
	Rules        FashionRules          `json:"rules"`
	ColorPalette ColorPalette          `json:"colorPalette"`
	ColorTheory  ColorTheory           `json:"color_theory"`
	Styles       map[string]Style      `json:"styles"`
	Combinations []CombinationRule     `json:"combinations"`
	Textures     TextureRules          `json:"textures"`
	SeasonRules  map[string]SeasonRule `json:"season_rules"`
	Generator    GeneratorRules        `json:"look_generator"`
	Tips         []string              `json:"tips"`
}

type FashionRules struct {
	MaxItemsPerLook   int      `json:"maxItemsPerLook"`
	MinItemsPerLook   int      `json:"minItemsPerLook"`
	AllowSameCategory bool     `json:"allowSameCategory"`
	PreferredSeasons  []string `json:"preferredSeasons"`
}

type ColorPalette struct {
	Complementary map[string][]string `json:"complementary"`
	Analogous     map[string][]string `json:"analogous"`
	Neutral       []string            `json:"neutral"`
	Seasonal      map[string][]string `json:"seasonal"`
	TeenFavorites map[string][]string `json:"teen_favorites"`
}

type ColorTheory struct {
	BaseColors            []string      `json:"base_colors"`
	ColorSchemes          []ColorScheme `json:"color_schemes"`
	ForbiddenCombinations []string      `json:"forbidden_combinations"`
}

type ColorScheme struct {
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Rules       []string   `json:"rules"`
	Examples    [][]string `json:"examples"`
}

type Style struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Occasions   []string `json:"occasions"`
	Colors      []string `json:"colors"`
	Items       []string `json:"items"`
	Features    []string `json:"features,omitempty"`
	Avoid       []string `json:"avoid,omitempty"`
}

type CombinationRule struct {
	Parts      []string `json:"parts"`
	Style      string   `json:"style"`
	ScoreBonus float64  `json:"scoreBonus"`
}

type TextureRules struct {
	Rules           []string `json:"rules"`
	BadCombinations []string `json:"bad_combinations"`
}

type SeasonRule struct {
	AllowedColors []string `json:"allowed_colors"`
	Layering      []string `json:"layering"`
}

type GeneratorRules struct {
	Rules        []string           `json:"rules"`
	ScoreWeights map[string]float64 `json:"score_weights"`
}

type LookGenerator struct {
	db *FashionDB
}

func NewLookGenerator(db *FashionDB) *LookGenerator {
	return &LookGenerator{
		db: db,
	}
}

func LoadFashionDB(path string) (*FashionDB, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return CreateDefaultFashionDB(), nil
		}
		return nil, fmt.Errorf("failed to read fashion db: %w", err)
	}

	var db FashionDB
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to parse fashion db: %w", err)
	}

	return &db, nil
}

func CreateDefaultFashionDB() *FashionDB {
	return &FashionDB{
		Rules: FashionRules{
			MaxItemsPerLook:   5,
			MinItemsPerLook:   2,
			AllowSameCategory: false,
			PreferredSeasons:  []string{"spring", "summer", "autumn", "winter"},
		},
		ColorPalette: ColorPalette{
			Complementary: map[string][]string{
				"blue":  {"orange", "yellow"},
				"red":   {"green", "turquoise"},
				"green": {"red", "pink"},
			},
			Analogous: map[string][]string{
				"blue":  {"navy", "cyan", "purple"},
				"red":   {"pink", "orange", "burgundy"},
				"green": {"olive", "teal", "lime"},
			},
			Neutral: []string{"black", "white", "gray", "beige", "brown", "navy"},
			Seasonal: map[string][]string{
				"spring": {"pastel", "light", "bright"},
				"summer": {"white", "blue", "light"},
				"autumn": {"brown", "orange", "olive"},
				"winter": {"black", "navy", "dark"},
			},
		},
		Styles: map[string]Style{
			"casual": {
				Name:        "casual",
				Description: "Повседневный стиль",
				Occasions:   []string{"work", "travel", "daily"},
				Colors:      []string{"blue", "gray", "white", "beige"},
				Items:       []string{"top", "bottom", "shoes"},
			},
			"streetwear": {
				Name:        "streetwear",
				Description: "Уличный стиль",
				Occasions:   []string{"daily", "travel"},
				Colors:      []string{"black", "gray", "white", "khaki"},
				Items:       []string{"top", "bottom", "shoes", "outer"},
			},
		},
		Combinations: []CombinationRule{
			{Parts: []string{"top", "bottom"}, Style: "casual", ScoreBonus: 0.1},
			{Parts: []string{"top", "bottom", "shoes"}, Style: "casual", ScoreBonus: 0.2},
		},
		Tips: []string{
			"Сочетайте не более 3 основных цветов в образе",
			"Используйте аксессуары для завершения образа",
		},
	}
}

func (lg *LookGenerator) GenerateLooks(
	items []models.Item,
	userGender string,
	preferences *models.LookPreferences,
	modelPhotos []models.ModelPhoto,
) ([]models.Look, error) {

	if len(items) < lg.db.Rules.MinItemsPerLook {
		return nil, fmt.Errorf("need at least %d items", lg.db.Rules.MinItemsPerLook)
	}

	analyzedItems := lg.analyzeItems(items)

	itemsByPart := make(map[models.BodyPart][]AnalyzedItem)
	for _, item := range analyzedItems {
		itemsByPart[item.Item.Part] = append(itemsByPart[item.Item.Part], item)
	}

	combinations := lg.generateSmartCombinations(itemsByPart, preferences)

	var looks []models.Look
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i, combo := range combinations {
		if i >= 5 { // Max 5 looks per generation
			break
		}

		score := lg.scoreCombination(combo, preferences)

		modelPhoto := lg.selectModelPhoto(modelPhotos, userGender, rng)

		lookItems := make([]models.Item, len(combo))
		for j, ai := range combo {
			lookItems[j] = ai.Item
		}

		look := models.Look{
			ID:           generateID(),
			Items:        lookItems,
			ModelGender:  modelPhoto.Gender,
			ModelPhotoID: modelPhoto.ID,
			RenderedURL:  "",
			Score:        score,
			IsFavorite:   false,
			IsSaved:      false,
			CreatedAt:    time.Now(),
			Meta: models.LookMeta{
				Colors: lg.extractColorsFromAnalyzed(combo),
				Style:  lg.determineStyle(combo, preferences),
				Season: lg.determineSeason(preferences),
			},
		}

		looks = append(looks, look)
	}

	sort.Slice(looks, func(i, j int) bool {
		return looks[i].Score > looks[j].Score
	})

	return looks, nil
}

type AnalyzedItem struct {
	Item   models.Item
	Colors []string
	Style  string
}

func (lg *LookGenerator) analyzeItems(items []models.Item) []AnalyzedItem {
	fmt.Printf("[analyzeItems] Starting analysis of %d items\n", len(items))

	analyzed := make([]AnalyzedItem, len(items))

	for i, item := range items {

		ai := AnalyzedItem{
			Item:   item,
			Colors: []string{"neutral"},
			Style:  "casual",
		}

		switch item.Part {
		case "top":
			ai.Colors = []string{"neutral", "blue"}
		case "bottom":
			ai.Colors = []string{"neutral", "black"}
		case "outer":
			ai.Colors = []string{"neutral", "black"}
		case "shoes":
			ai.Colors = []string{"neutral", "black"}
		case "accessory":
			ai.Colors = []string{"neutral", "gold"}
		}

		fmt.Printf("[analyzeItems] Item %s: part=%s, colors=%v, style=%s\n",
			item.ID, item.Part, ai.Colors, ai.Style)

		analyzed[i] = ai
	}

	fmt.Printf("[analyzeItems] Completed analysis of %d items\n", len(analyzed))
	return analyzed
}

func (lg *LookGenerator) getImagePath(url string) string {


	if len(url) > 0 && url[0] == '/' {
		return "storage" + url
	}
	return url
}

func (lg *LookGenerator) generateSmartCombinations(
	itemsByPart map[models.BodyPart][]AnalyzedItem,
	preferences *models.LookPreferences,
) [][]AnalyzedItem {

	var combinations [][]AnalyzedItem

	parts := make([]models.BodyPart, 0)
	for part, items := range itemsByPart {
		if len(items) > 0 {
			parts = append(parts, part)
		}
	}

	tops, hasTops := itemsByPart["top"]
	bottoms, hasBottoms := itemsByPart["bottom"]
	shoes, hasShoes := itemsByPart["shoes"]

	if !hasTops || !hasBottoms {
		return combinations // Need at least top and bottom
	}

	maxItems := 4
	if lg.db.Rules.MaxItemsPerLook > 0 {
		maxItems = lg.db.Rules.MaxItemsPerLook
	}

	for _, top := range tops {
		for _, bottom := range bottoms {

			combo := []AnalyzedItem{top, bottom}

			if hasShoes && len(combo) < maxItems {
				for _, shoe := range shoes {

					if lg.areColorsCompatible(top.Colors, bottom.Colors, shoe.Colors) {
						newCombo := append(combo, shoe)

						if outers, hasOuter := itemsByPart["outer"]; hasOuter && len(newCombo) < maxItems {
							for _, outer := range outers {
								if lg.areColorsCompatible(top.Colors, bottom.Colors, outer.Colors) {
									finalCombo := append(newCombo, outer)
									combinations = append(combinations, finalCombo)
								}
							}
						} else {
							combinations = append(combinations, newCombo)
						}
					}
				}
			} else {

				combinations = append(combinations, combo)
			}

			if len(combinations) >= 20 {
				break
			}
		}
		if len(combinations) >= 20 {
			break
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(combinations), func(i, j int) {
		combinations[i], combinations[j] = combinations[j], combinations[i]
	})

	if len(combinations) > 10 {
		combinations = combinations[:10]
	}

	return combinations
}

func (lg *LookGenerator) areColorsCompatible(colors1, colors2, colors3 []string) bool {

	allColors := append(append(colors1, colors2...), colors3...)

	for _, forbidden := range lg.db.ColorTheory.ForbiddenCombinations {

		if containsAny(allColors, forbidden) {
			return false
		}
	}

	uniqueColors := make(map[string]bool)
	for _, c := range allColors {
		uniqueColors[c] = true
	}

	return len(uniqueColors) <= 3
}

func (lg *LookGenerator) scoreCombination(items []AnalyzedItem, preferences *models.LookPreferences) float64 {
	weights := lg.db.Generator.ScoreWeights
	if len(weights) == 0 {
		weights = map[string]float64{
			"color_match":     0.45,
			"style_match":     0.35,
			"texture_harmony": 0.1,
			"season_match":    0.1,
		}
	}

	score := 0.0

	colorScore := lg.scoreColorMatch(items)
	score += colorScore * weights["color_match"]

	styleScore := lg.scoreStyleMatch(items, preferences)
	score += styleScore * weights["style_match"]

	textureScore := 0.5 // default
	score += textureScore * weights["texture_harmony"]

	seasonScore := 0.5 // default
	if preferences != nil && preferences.Season != "" {
		seasonScore = 0.8
	}
	score += seasonScore * weights["season_match"]

	hasTop, hasBottom, hasShoes := false, false, false
	for _, item := range items {
		switch item.Item.Part {
		case "top":
			hasTop = true
		case "bottom":
			hasBottom = true
		case "shoes":
			hasShoes = true
		}
	}

	if hasTop && hasBottom && hasShoes {
		score += 0.1 // Bonus for complete outfit
	}

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

func (lg *LookGenerator) scoreColorMatch(items []AnalyzedItem) float64 {
	if len(items) < 2 {
		return 0.5
	}

	allColors := []string{}
	for _, item := range items {
		allColors = append(allColors, item.Colors...)
	}

	neutralCount := 0
	for _, c := range allColors {
		if contains(lg.db.ColorPalette.Neutral, c) {
			neutralCount++
		}
	}

	uniqueColors := make(map[string]bool)
	for _, c := range allColors {
		uniqueColors[c] = true
	}

	numUnique := len(uniqueColors)
	if numUnique >= 2 && numUnique <= 3 {
		return 0.9
	} else if numUnique == 1 {
		return 0.7 // monochrome
	} else if numUnique > 4 {
		return 0.4 // too many colors
	}

	return 0.6
}

func (lg *LookGenerator) scoreStyleMatch(items []AnalyzedItem, preferences *models.LookPreferences) float64 {
	if preferences == nil || preferences.Style == "" {
		return 0.6 // no preference
	}

	matchingCount := 0
	for _, item := range items {
		if item.Style == preferences.Style {
			matchingCount++
		}
	}

	if matchingCount == len(items) {
		return 1.0 // perfect match
	} else if matchingCount > 0 {
		return 0.7 // partial match
	}

	return 0.4 // no match
}

func (lg *LookGenerator) selectModelPhoto(
	photos []models.ModelPhoto,
	userGender string,
	rng *rand.Rand,
) models.ModelPhoto {

	var candidates []models.ModelPhoto

	if userGender == "male" || userGender == "female" {
		for _, photo := range photos {
			if photo.Gender == userGender {
				candidates = append(candidates, photo)
			}
		}
	}

	if len(candidates) == 0 {
		candidates = photos
	}

	if len(candidates) == 0 {
		return models.ModelPhoto{
			ID:       "default",
			Gender:   "unisex",
			ImageURL: "/models/male_1.png",
			Pose:     "standing",
		}
	}

	return candidates[rng.Intn(len(candidates))]
}

func (lg *LookGenerator) extractColorsFromAnalyzed(items []AnalyzedItem) []string {
	colors := []string{}
	for _, item := range items {
		colors = append(colors, item.Colors...)
	}

	seen := make(map[string]bool)
	unique := []string{}
	for _, c := range colors {
		if !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}

	return unique
}

func (lg *LookGenerator) determineStyle(items []AnalyzedItem, preferences *models.LookPreferences) string {
	if preferences != nil && preferences.Style != "" {
		return preferences.Style
	}

	styleCounts := make(map[string]int)
	for _, item := range items {
		styleCounts[item.Style]++
	}

	maxCount := 0
	dominantStyle := "casual"
	for style, count := range styleCounts {
		if count > maxCount {
			maxCount = count
			dominantStyle = style
		}
	}

	return dominantStyle
}

func (lg *LookGenerator) determineSeason(preferences *models.LookPreferences) string {
	if preferences != nil && preferences.Season != "" {
		return preferences.Season
	}

	month := time.Now().Month()
	switch month {
	case 3, 4, 5:
		return "spring"
	case 6, 7, 8:
		return "summer"
	case 9, 10, 11:
		return "autumn"
	default:
		return "winter"
	}
}


func generateID() string {
	return uuid.New().String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsAny(slice []string, substr string) bool {
	for _, s := range slice {
		if s == substr {
			return true
		}
	}
	return false
}
