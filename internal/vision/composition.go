package vision

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"

	"fashion-look-generator/internal/models"

	"github.com/disintegration/imaging"
)

type ImageComposer struct {
	basePath  string
	bgRemover *BackgroundRemover
}

func NewImageComposer(basePath string) *ImageComposer {
	return &ImageComposer{
		basePath:  basePath,
		bgRemover: NewBackgroundRemover(),
	}
}

func (ic *ImageComposer) ComposeLook(
	look *models.Look,
	modelPhoto models.ModelPhoto,
	items []models.Item,
	storagePath string,
) (string, error) {
	return ic.simpleCompose(look, modelPhoto, items, storagePath)
}

func (ic *ImageComposer) simpleCompose(
	look *models.Look,
	modelPhoto models.ModelPhoto,
	items []models.Item,
	storagePath string,
) (string, error) {
	fmt.Printf("[ComposeLook] Starting for look %s, model: %s, items: %d\n", look.ID, modelPhoto.ImageURL, len(items))

	modelRelativePath := modelPhoto.ImageURL
	if len(modelRelativePath) > 0 && modelRelativePath[0] == '/' {
		modelRelativePath = modelRelativePath[1:]
	}
	modelPath := filepath.Join(storagePath, modelRelativePath)

	fmt.Printf("[ComposeLook] Model path: %s\n", modelPath)

	modelImg, err := imaging.Open(modelPath)
	if err != nil {
		fmt.Printf("[ComposeLook] Model not found at %s: %v, creating placeholder\n", modelPath, err)
		modelImg = image.NewRGBA(image.Rect(0, 0, 512, 768))
	} else {
		fmt.Printf("[ComposeLook] Model loaded successfully\n")
	}

	resultBounds := modelImg.Bounds()
	resultImg := image.NewRGBA(resultBounds)
	draw.Draw(resultImg, resultBounds, modelImg, image.Point{}, draw.Src)

	positions := map[models.BodyPart]image.Rectangle{
		"top":       image.Rect(150, 100, 350, 300),
		"bottom":    image.Rect(150, 300, 350, 550),
		"outer":     image.Rect(130, 80, 370, 400),
		"shoes":     image.Rect(180, 550, 320, 700),
		"accessory": image.Rect(200, 50, 300, 150),
	}

	for _, item := range items {
		itemRelativePath := item.ImageURL
		if len(itemRelativePath) > 0 && itemRelativePath[0] == '/' {
			itemRelativePath = itemRelativePath[1:]
		}
		itemPath := filepath.Join(storagePath, itemRelativePath)
		fmt.Printf("[ComposeLook] Loading item %s from: %s\n", item.ID, itemPath)

		itemImg, err := imaging.Open(itemPath)
		if err != nil {
			fmt.Printf("[ComposeLook] Failed to load item %s: %v\n", item.ID, err)
			continue
		}

		pos, exists := positions[item.Part]
		if !exists {
			continue
		}

		fitted := imaging.Fit(itemImg, pos.Dx(), pos.Dy(), imaging.Lanczos)

		itemBounds := fitted.Bounds()
		offsetX := pos.Min.X + (pos.Dx()-itemBounds.Dx())/2
		offsetY := pos.Min.Y + (pos.Dy()-itemBounds.Dy())/2

		draw.Draw(resultImg,
			image.Rect(offsetX, offsetY, offsetX+itemBounds.Dx(), offsetY+itemBounds.Dy()),
			fitted,
			image.Point{},
			draw.Over)
	}

	outputName := fmt.Sprintf("look_%s.jpg", look.ID)
	outputPath := filepath.Join(storagePath, "looks", outputName)
	fmt.Printf("[ComposeLook] Saving result to: %s\n", outputPath)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	if err := jpeg.Encode(outputFile, resultImg, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	fmt.Printf("[ComposeLook] Successfully saved to %s\n", outputPath)

	return "/looks/" + outputName, nil
}

func (ic *ImageComposer) CreatePlaceholder(lookID string, storagePath string) (string, error) {
	width, height := 512, 768
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, image.White)
		}
	}

	outputName := fmt.Sprintf("look_%s_placeholder.jpg", lookID)
	outputPath := filepath.Join(storagePath, "looks", outputName)

	os.MkdirAll(filepath.Dir(outputPath), 0755)

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 85}); err != nil {
		return "", err
	}

	return "/looks/" + outputName, nil
}

func (ic *ImageComposer) PreprocessItem(itemPath string, storagePath string) (string, error) {
	if !ic.bgRemover.IsAvailable() {
		return itemPath, nil
	}

	outputDir := filepath.Join(storagePath, "temp")
	outputPath, err := ic.bgRemover.RemoveBackground(itemPath, outputDir)
	if err != nil {
		return itemPath, nil
	}

	return outputPath, nil
}
