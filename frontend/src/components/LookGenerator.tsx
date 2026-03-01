import { useState } from 'react'
import { Sparkles, Check, Shirt, Loader2, Heart } from 'lucide-react'
import type { Item, BodyPart, Look } from '../types'
import LookResult from './LookResult'

interface LookGeneratorProps {
  items: Item[]
  onGenerate: (selectedItemIds: string[]) => void
  isGenerating: boolean
  generatedLooks?: Look[]
  onSave?: (lookId: string, saved: boolean) => void
  onFavorite?: (lookId: string, favorite: boolean) => void
  generationProgress?: string
}

const BODY_PARTS: { id: BodyPart; label: string }[] = [
  { id: 'top', label: 'Верх' },
  { id: 'bottom', label: 'Низ' },
  { id: 'outer', label: 'Верхняя одежда' },
  { id: 'shoes', label: 'Обувь' },
  { id: 'accessory', label: 'Аксессуары' },
]

export default function LookGenerator({ 
  items, 
  onGenerate, 
  isGenerating,
  generatedLooks = [],
  onSave,
  onFavorite,
  generationProgress = 'Генерация образов...'
}: LookGeneratorProps) {
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set())
  const [preferences, setPreferences] = useState({
    style: '',
    season: '',
    occasion: '',
  })

  const itemsByPart = BODY_PARTS.map(part => ({
    ...part,
    items: items.filter(item => item.part === part.id),
  }))

  const toggleItem = (itemId: string) => {
    const newSelected = new Set(selectedItems)
    if (newSelected.has(itemId)) {
      newSelected.delete(itemId)
    } else {
      newSelected.add(itemId)
    }
    setSelectedItems(newSelected)
  }

  const handleGenerate = () => {
    if (selectedItems.size < 2) {
      alert('Выберите минимум 2 вещи для генерации образа')
      return
    }
    // Pass item IDs to generate looks
    onGenerate(Array.from(selectedItems))
  }

  const hasItemsInCategory = itemsByPart.some(cat => cat.items.length > 0)

  if (!hasItemsInCategory) {
    return (
      <div className="card text-center py-16">
        <Shirt className="w-16 h-16 text-sand mx-auto mb-4" />
        <h3 className="text-xl font-semibold text-dark-green mb-2">
          Гардероб пуст
        </h3>
        <p className="text-olive mb-6">
          Добавьте вещи в гардероб, чтобы создавать образы
        </p>
      </div>
    )
  }

  if (isGenerating) {
    return (
      <div className="card text-center py-16">
        <Loader2 className="w-16 h-16 text-olive mx-auto mb-4 animate-spin" />
        <h3 className="text-xl font-semibold text-dark-green mb-2">
          {generationProgress}
        </h3>
        <p className="text-olive mb-4">
          Анализируем вещи и подбираем сочетания...
        </p>
        <div className="w-64 mx-auto bg-sand/20 rounded-full h-2 overflow-hidden">
          <div className="bg-olive h-full rounded-full animate-pulse" style={{ width: '60%' }}></div>
        </div>
        <p className="text-sm text-olive/60 mt-4">
          Это может занять 10-30 секунд
        </p>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-2xl font-bold text-dark-green">Генератор образов</h2>
          <p className="text-olive">
            Выберите вещи для создания сочетаний ({selectedItems.size} выбрано)
          </p>
        </div>
        <button
          onClick={handleGenerate}
          disabled={selectedItems.size < 2}
          className="btn-primary flex items-center gap-2 disabled:opacity-50"
        >
          <Sparkles className="w-5 h-5" />
          Сгенерировать
        </button>
      </div>

      {/* Generated Looks - show after generation */}
      {generatedLooks.length > 0 && (
        <div className="mb-8">
          <h3 className="text-xl font-semibold text-dark-green mb-4 flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-olive" />
            Сгенерированные образы ({generatedLooks.length})
          </h3>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {generatedLooks.map(look => (
              <LookResult
                key={look.id}
                look={look}
                onSave={onSave ? (saved) => onSave(look.id, saved) : () => {}}
                onFavorite={onFavorite ? (fav) => onFavorite(look.id, fav) : () => {}}
              />
            ))}
          </div>
        </div>
      )}

      {/* Preferences */}
      <div className="card mb-6">
        <h3 className="font-semibold text-dark-green mb-4">Предпочтения</h3>
        <div className="grid md:grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-olive mb-2">Стиль</label>
            <select
              value={preferences.style}
              onChange={(e) => setPreferences({ ...preferences, style: e.target.value })}
              className="input-field"
            >
              <option value="">Любой</option>
              <option value="casual">Повседневный</option>
              <option value="formal">Деловой</option>
              <option value="sport">Спортивный</option>
              <option value="elegant">Элегантный</option>
              <option value="y2k">Y2K</option>
              <option value="soft_girl">Soft Girl</option>
              <option value="streetwear">Streetwear</option>
              <option value="techwear">Techwear</option>
              <option value="skater">Skater</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-olive mb-2">Сезон</label>
            <select
              value={preferences.season}
              onChange={(e) => setPreferences({ ...preferences, season: e.target.value })}
              className="input-field"
            >
              <option value="">Любой</option>
              <option value="spring">Весна</option>
              <option value="summer">Лето</option>
              <option value="autumn">Осень</option>
              <option value="winter">Зима</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-olive mb-2">Повод</label>
            <select
              value={preferences.occasion}
              onChange={(e) => setPreferences({ ...preferences, occasion: e.target.value })}
              className="input-field"
            >
              <option value="">Любой</option>
              <option value="work">Работа</option>
              <option value="date">Свидание</option>
              <option value="party">Вечеринка</option>
              <option value="travel">Путешествие</option>
            </select>
          </div>
        </div>
      </div>

      {/* Items by Category */}
      <div className="space-y-6">
        {itemsByPart.map(part => (
          part.items.length > 0 && (
            <div key={part.id} className="card">
              <h3 className="font-semibold text-dark-green mb-4 flex items-center gap-2">
                {part.label}
                <span className="text-sm font-normal text-olive">
                  ({part.items.length})
                </span>
              </h3>
              <div className="grid grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
                {part.items.map(item => {
                  const isSelected = selectedItems.has(item.id)
                  return (
                    <button
                      key={item.id}
                      onClick={() => toggleItem(item.id)}
                      className={`relative rounded-lg overflow-hidden aspect-square transition-all ${
                        isSelected
                          ? 'ring-4 ring-olive ring-offset-2'
                          : 'hover:opacity-80'
                      }`}
                    >
                      <img
                        src={item.thumbnailUrl || item.imageUrl}
                        alt={part.label}
                        className="w-full h-full object-cover"
                      />
                      {isSelected && (
                        <div className="absolute inset-0 bg-olive/50 flex items-center justify-center">
                          <Check className="w-8 h-8 text-cream" />
                        </div>
                      )}
                    </button>
                  )
                })}
              </div>
            </div>
          )
        ))}
      </div>
    </div>
  )
}
