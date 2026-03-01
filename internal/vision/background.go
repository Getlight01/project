package vision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type BackgroundRemover struct {
	pythonPath string
	scriptPath string
}

func NewBackgroundRemover() *BackgroundRemover {
	return &BackgroundRemover{
		pythonPath: "python", // или "python3" на Linux/Mac
		scriptPath: "scripts/remove_bg.py",
	}
}


func (br *BackgroundRemover) RemoveBackground(inputPath string, outputDir string) (string, error) {

	if _, err := os.Stat(br.scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Python script not found: %s", br.scriptPath)
	}

	filename := filepath.Base(inputPath)
	nameWithoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
	outputPath := filepath.Join(outputDir, nameWithoutExt+"_nobg.png")

	cmd := exec.Command(br.pythonPath, br.scriptPath, inputPath, outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("rembg failed: %v\nOutput: %s", err, string(output))
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output file not created")
	}

	return outputPath, nil
}

func (br *BackgroundRemover) IsAvailable() bool {

	cmd := exec.Command(br.pythonPath, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}

	cmd = exec.Command(br.pythonPath, "-c", "import rembg")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func (br *BackgroundRemover) GetInstallInstructions() string {
	return `
=== УСТАНОВКА REMBG ===

1. Установите Python 3.9+ с https://python.org
   (Важно: отметьте "Add Python to PATH" при установке)

2. Откройте PowerShell и выполните:
   pip install rembg pillow

3. Проверьте установку:
   python scripts/remove_bg.py

Если команда "python" не найдена, попробуйте "py" или "python3"
`
}
