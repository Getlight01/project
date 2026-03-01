import { useState } from 'react'
import { Heart, Download, Check, User, Shirt } from 'lucide-react'
import type { Look } from '../types'

interface LookResultProps {
  look: Look
  onSave: (saved: boolean) => void
  onFavorite: (favorite: boolean) => void
}

export default function LookResult({ look, onSave, onFavorite }: LookResultProps) {
  const [isSaved, setIsSaved] = useState(look?.isSaved || false)
  const [isFavorite, setIsFavorite] = useState(look?.isFavorite || false)
  const [imageError, setImageError] = useState(false)

  if (!look) {
    return null
  }

  const handleSave = async () => {
    const renderedUrl = look.renderedUrl || ''
    if (!renderedUrl || imageError) {
      alert('Изображение недоступно для скачивания')
      return
    }
    
    try {
      const response = await fetch(renderedUrl)
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `look_${look.id}.jpg`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      setIsSaved(true)
      onSave(true)
    } catch (error) {
      console.error('Failed to download:', error)
      alert('Не удалось скачать изображение')
    }
  }

  const handleFavorite = () => {
    const newFavorite = !isFavorite
    
    if (!newFavorite) {
      if (!confirm('Вы уверены? Образ будет удален из "Моих образов". Это действие нельзя отменить.')) {
        return
      }
    }
    
    setIsFavorite(newFavorite)
    onFavorite(newFavorite)
  }

  const items = look.items || []
  const meta = look.meta || { style: 'casual', season: 'all', colors: [] }
  const renderedUrl = look.renderedUrl || ''
  const modelGender = look.modelGender || 'unisex'

  return (
    <div className="card overflow-hidden">
      <div className="relative aspect-[3/4] bg-sand/10 rounded-lg overflow-hidden mb-4">
        {renderedUrl && !imageError ? (
          <img
            src={renderedUrl}
            alt="Generated look"
            className="w-full h-full object-cover"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-sand/20">
            <div className="text-center p-4">
              <User className="w-16 h-16 text-sand mx-auto mb-2" />
              <p className="text-olive text-sm mb-1">
                Модель: {modelGender === 'male' ? 'Мужчина' : modelGender === 'female' ? 'Женщина' : 'Унисекс'}
              </p>
              <p className="text-olive/60 text-xs">
                {imageError ? 'Ошибка загрузки изображения' : 'Изображение генерируется...'}
              </p>
            </div>
          </div>
        )}
      </div>

      <div className="space-y-3">
        <div className="flex items-center gap-2 text-sm text-olive flex-wrap">
          <span className="bg-sand/20 px-2 py-1 rounded">{meta.style || 'casual'}</span>
          <span className="bg-sand/20 px-2 py-1 rounded">{meta.season || 'all'}</span>
          {meta.colors && meta.colors.length > 0 && (
            <span className="bg-sand/20 px-2 py-1 rounded">
              {meta.colors.slice(0, 3).join(', ')}
            </span>
          )}
        </div>

        {items.length > 0 ? (
          <div className="flex gap-2 overflow-x-auto pb-2">
            {items.map(item => (
              <div
                key={item?.id || Math.random()}
                className="flex-shrink-0 w-16 h-16 rounded-lg overflow-hidden bg-sand/10"
              >
                {item?.thumbnailUrl || item?.imageUrl ? (
                  <img
                    src={item.thumbnailUrl || item.imageUrl}
                    alt={item?.part || 'item'}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center bg-sand/20">
                    <Shirt className="w-6 h-6 text-sand" />
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-olive/60 italic">Нет данных о вещах</p>
        )}

        <div className="flex gap-2 pt-2">
          <button
            onClick={handleSave}
            className={`flex-1 flex items-center justify-center gap-2 py-2 rounded-lg font-medium transition-colors ${
              isSaved ? 'bg-olive text-cream' : 'bg-sand/20 text-olive hover:bg-sand/30'
            }`}
          >
            {isSaved ? <Check className="w-4 h-4" /> : <Download className="w-4 h-4" />}
            {isSaved ? 'Сохранено' : 'Скачать'}
          </button>
          <button
            onClick={handleFavorite}
            className={`flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
              isFavorite ? 'bg-terracotta text-cream' : 'bg-sand/20 text-olive hover:bg-sand/30'
            }`}
          >
            <Heart className={`w-4 h-4 ${isFavorite ? 'fill-current' : ''}`} />
          </button>
        </div>
      </div>
    </div>
  )
}
