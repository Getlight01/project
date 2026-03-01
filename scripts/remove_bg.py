#!/usr/bin/env python3
"""
Скрипт для удаления фона с изображения одежды
Использует библиотеку rembg (бесплатно, локально)

Установка:
    pip install rembg pillow

Использование:
    python remove_bg.py input.jpg output.png
"""

import sys
import os

def check_dependencies():
    """Проверяет установлены ли необходимые библиотеки"""
    try:
        from rembg import remove
        from PIL import Image
        return True
    except ImportError as e:
        print(f"ERROR: Missing dependency: {e}")
        print("Install with: pip install rembg pillow")
        return False

def remove_background(input_path, output_path):
    """
    Удаляет фон с изображения
    
    Args:
        input_path: путь к входному изображению
        output_path: путь для сохранения (рекомендуется .png для прозрачности)
    """
    try:
        from rembg import remove
        from PIL import Image
        
        # Проверяем существование файла
        if not os.path.exists(input_path):
            print(f"ERROR: File not found: {input_path}")
            return False
        
        # Открываем изображение
        print(f"Processing: {input_path}")
        input_image = Image.open(input_path)
        
        # Удаляем фон
        print("Removing background...")
        output_image = remove(input_image)
        
        # Сохраняем результат
        output_image.save(output_path)
        print(f"Saved to: {output_path}")
        
        return True
        
    except Exception as e:
        print(f"ERROR: {e}")
        return False

def main():
    # Проверяем аргументы
    if len(sys.argv) != 3:
        print("Usage: python remove_bg.py <input.jpg> <output.png>")
        print("Example: python remove_bg.py shirt.jpg shirt_nobg.png")
        sys.exit(1)
    
    input_path = sys.argv[1]
    output_path = sys.argv[2]
    
    # Проверяем зависимости
    if not check_dependencies():
        sys.exit(1)
    
    # Обрабатываем изображение
    success = remove_background(input_path, output_path)
    
    if success:
        print("SUCCESS")
        sys.exit(0)
    else:
        print("FAILED")
        sys.exit(1)

if __name__ == "__main__":
    main()
