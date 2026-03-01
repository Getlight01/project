import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { 
  LogOut, 
  Plus, 
  Shirt, 
  Sparkles, 
  Heart, 
  Trash2, 
  Download,
  User,
  Palette,
  Footprints,
  Layers
} from 'lucide-react'
import { useAuthStore } from '../store/authStore'
import { itemsApi, looksApi } from '../utils/api'
import type { Item, Look, BodyPart } from '../types'
import ItemUploadModal from '../components/ItemUploadModal'
import LookGenerator from '../components/LookGenerator'
import LookResult from '../components/LookResult'

const BODY_PARTS: { id: BodyPart; label: string; icon: React.ReactNode }[] = [
  { id: 'top', label: 'Верх', icon: <Shirt className="w-5 h-5" /> },
  { id: 'bottom', label: 'Низ', icon: <Download className="w-5 h-5" /> },
  { id: 'outer', label: 'Верхняя одежда', icon: <Layers className="w-5 h-5" /> },
  { id: 'shoes', label: 'Обувь', icon: <Footprints className="w-5 h-5" /> },
  { id: 'accessory', label: 'Аксессуары', icon: <Palette className="w-5 h-5" /> },
]

export default function WorkshopPage() {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const [items, setItems] = useState<Item[]>([])
  const [looks, setLooks] = useState<Look[]>([])
  const [activeTab, setActiveTab] = useState<'wardrobe' | 'generator' | 'looks'>('wardrobe')
  const [selectedPart, setSelectedPart] = useState<BodyPart | 'all'>('all')
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false)
  const [isGenerating, setIsGenerating] = useState(false)
  const [generatedLooks, setGeneratedLooks] = useState<Look[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadItems()
    loadLooks()
  }, [])

  const loadItems = async () => {
    try {
      const response = await itemsApi.getUserItems()
      setItems(response.data.items || [])
    } catch (error) {
      console.error('Failed to load items:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const loadLooks = async () => {
    try {
      const response = await looksApi.getUserLooks()
      setLooks(response.data.looks || [])
    } catch (error) {
      console.error('Failed to load looks:', error)
    }
  }

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  const handleDeleteItem = async (id: string) => {
    if (!confirm('Удалить эту вещь?')) return
    try {
      await itemsApi.deleteItem(id)
      setItems(items.filter(item => item.id !== id))
    } catch (error) {
      alert('Ошибка при удалении')
    }
  }

  const handleGenerateLooks = async (selectedItemIds: string[]) => {
    setIsGenerating(true)
    try {
      const response = await looksApi.generate({ itemIds: selectedItemIds })
      setGeneratedLooks([...(generatedLooks || []), ...(response.data.looks || [])])
      await loadLooks()
    } catch (error: any) {
      alert(error.response?.data?.message || 'Ошибка генерации образов')
    } finally {
      setIsGenerating(false)
    }
  }

  const handleSaveLook = async (lookId: string, isSaved: boolean) => {
    try {
      await looksApi.updateLook(lookId, { isSaved })
      setLooks(looks.map(look => 
        look.id === lookId ? { ...look, isSaved } : look
      ))
      setGeneratedLooks(generatedLooks.map(look => 
        look.id === lookId ? { ...look, isSaved } : look
      ))
    } catch (error) {
      console.error('Failed to save look:', error)
    }
  }

  const handleFavoriteLook = async (lookId: string, isFavorite: boolean) => {
    try {
      await looksApi.updateLook(lookId, { isFavorite })
      setLooks(looks.map(look => 
        look.id === lookId ? { ...look, isFavorite } : look
      ))
      setGeneratedLooks(generatedLooks.map(look => 
        look.id === lookId ? { ...look, isFavorite } : look
      ))
    } catch (error) {
      console.error('Failed to favorite look:', error)
    }
  }

  const filteredItems = selectedPart === 'all' 
    ? items 
    : items.filter(item => item.part === selectedPart)

  const itemsByPart = BODY_PARTS.map(part => ({
    ...part,
    count: items.filter(item => item.part === part.id).length
  }))

  const favoriteLooks = looks.filter(l => l.isFavorite)

  const openUploadModal = () => {
    setIsUploadModalOpen(true)
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-cream flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-olive"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-cream">
      {/* Header */}
      <header className="bg-dark-green text-cream py-4 sticky top-0 z-50">
        <div className="container mx-auto px-6 flex justify-between items-center">
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold">Мастерская</h1>
            <span className="text-cream/60 text-sm hidden md:inline">
              {user?.displayName}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <div className="text-sm text-cream/80 hidden md:block">
              {items.length} вещей • {favoriteLooks.length} избранных образов
            </div>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 text-cream/80 hover:text-cream transition-colors"
            >
              <LogOut className="w-5 h-5" />
              <span className="hidden md:inline">Выйти</span>
            </button>
          </div>
        </div>
      </header>

      {/* Navigation */}
      <nav className="bg-white border-b border-sand/20">
        <div className="container mx-auto px-6">
          <div className="flex gap-1">
            {[
              { id: 'wardrobe', label: 'Гардероб', icon: <Shirt className="w-4 h-4" /> },
              { id: 'generator', label: 'Генератор', icon: <Sparkles className="w-4 h-4" /> },
              { id: 'looks', label: 'Мои образы', icon: <Heart className="w-4 h-4" /> },
            ].map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center gap-2 px-6 py-4 font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'text-olive border-b-2 border-olive'
                    : 'text-olive/60 hover:text-olive'
                }`}
              >
                {tab.icon}
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="container mx-auto px-6 py-8">
        {/* Wardrobe Tab */}
        {activeTab === 'wardrobe' && (
          <div>
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-2xl font-bold text-dark-green">Мой гардероб</h2>
              <button
                onClick={openUploadModal}
                className="btn-primary flex items-center gap-2"
              >
                <Plus className="w-5 h-5" />
                Добавить вещь
              </button>
            </div>

            {/* Filter by body part */}
            <div className="flex flex-wrap gap-2 mb-6">
              <button
                onClick={() => setSelectedPart('all')}
                className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                  selectedPart === 'all'
                    ? 'bg-olive text-cream'
                    : 'bg-white text-olive hover:bg-sand/20'
                }`}
              >
                Все ({items.length})
              </button>
              {itemsByPart.map(part => (
                <button
                  key={part.id}
                  onClick={() => setSelectedPart(part.id)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
                    selectedPart === part.id
                      ? 'bg-olive text-cream'
                      : 'bg-white text-olive hover:bg-sand/20'
                  }`}
                >
                  {part.icon}
                  {part.label} ({part.count})
                </button>
              ))}
            </div>

            {/* Items Grid */}
            {filteredItems.length === 0 ? (
              <div className="card text-center py-16">
                <Shirt className="w-16 h-16 text-sand mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-dark-green mb-2">
                  {selectedPart === 'all' ? 'Гардероб пуст' : 'Нет вещей в этой категории'}
                </h3>
                <p className="text-olive mb-6">
                  {selectedPart === 'all' 
                    ? 'Начните с добавления первой вещи в ваш гардероб'
                    : 'Добавьте вещи в эту категорию'}
                </p>
                <button
                  onClick={openUploadModal}
                  className="btn-primary"
                >
                  Добавить вещь
                </button>
              </div>
            ) : (
              <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
                {filteredItems.map(item => (
                  <div key={item.id} className="card p-4 group relative">
                    <div className="aspect-square rounded-lg overflow-hidden mb-3 bg-sand/10">
                      <img
                        src={item.thumbnailUrl || item.imageUrl}
                        alt={item.part}
                        className="w-full h-full object-cover"
                      />
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm font-medium text-dark-green capitalize">
                        {BODY_PARTS.find(p => p.id === item.part)?.label || item.part}
                      </span>
                      <button
                        onClick={() => handleDeleteItem(item.id)}
                        className="text-terracotta hover:text-terracotta/70 opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Generator Tab */}
        {activeTab === 'generator' && (
          <LookGenerator
            items={items}
            onGenerate={handleGenerateLooks}
            isGenerating={isGenerating}
            generatedLooks={generatedLooks}
            onSave={handleSaveLook}
            onFavorite={handleFavoriteLook}
          />
        )}

        {/* Looks Tab - only show favorited looks */}
        {activeTab === 'looks' && (
          <div>
            <h2 className="text-2xl font-bold text-dark-green mb-6">
              Мои образы
            </h2>
            
            {favoriteLooks.length === 0 ? (
              <div className="card text-center py-16">
                <Heart className="w-16 h-16 text-sand mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-dark-green mb-2">
                  Нет избранных образов
                </h3>
                <p className="text-olive mb-6">
                  Поставьте лайк на понравившийся образ
                </p>
                <button
                  onClick={() => setActiveTab('generator')}
                  className="btn-primary"
                >
                  Перейти к генератору
                </button>
              </div>
            ) : (
              <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                {favoriteLooks.map(look => (
                  <LookResult
                    key={look.id}
                    look={look}
                    onSave={(saved) => handleSaveLook(look.id, saved)}
                    onFavorite={(fav) => handleFavoriteLook(look.id, fav)}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </main>

      {/* Upload Modal */}
      {isUploadModalOpen && (
        <ItemUploadModal
          onClose={() => setIsUploadModalOpen(false)}
          onUpload={loadItems}
          defaultCategory={selectedPart === 'all' ? undefined : selectedPart}
        />
      )}
    </div>
  )
}
