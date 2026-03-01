import { useState, useRef } from 'react'
import { X, Upload, Link, Camera } from 'lucide-react'
import { itemsApi } from '../utils/api'
import type { BodyPart } from '../types'

interface ItemUploadModalProps {
  onClose: () => void
  onUpload: () => void
  defaultCategory?: BodyPart
}

const BODY_PARTS: { id: BodyPart; label: string }[] = [
  { id: 'top', label: 'Верх' },
  { id: 'bottom', label: 'Низ' },
  { id: 'outer', label: 'Верхняя одежда' },
  { id: 'shoes', label: 'Обувь' },
  { id: 'accessory', label: 'Аксессуары' },
]

export default function ItemUploadModal({ onClose, onUpload, defaultCategory }: ItemUploadModalProps) {
  const [activeTab, setActiveTab] = useState<'file' | 'link'>('file')
  const [selectedPart, setSelectedPart] = useState<BodyPart>(defaultCategory || 'top')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [linkUrl, setLinkUrl] = useState('')
  const [isUploading, setIsUploading] = useState(false)
  const [preview, setPreview] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setSelectedFile(file)
      const reader = new FileReader()
      reader.onloadend = () => {
        setPreview(reader.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  const handleUpload = async () => {
    if (activeTab === 'file' && !selectedFile) {
      alert('Выберите файл')
      return
    }
    if (activeTab === 'link' && !linkUrl) {
      alert('Введите ссылку')
      return
    }

    setIsUploading(true)
    try {
      if (activeTab === 'file') {
        const formData = new FormData()
        formData.append('file', selectedFile!)
        formData.append('part', selectedPart)
        await itemsApi.upload(formData)
      } else {
        await itemsApi.addLink({ url: linkUrl, part: selectedPart })
      }
      onUpload()
      onClose()
    } catch (error: any) {
      alert(error.response?.data?.message || 'Ошибка загрузки')
    } finally {
      setIsUploading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="card w-full max-w-lg max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-2xl font-bold text-dark-green">Добавить вещь</h2>
          <button
            onClick={onClose}
            className="text-sand hover:text-olive transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex gap-2 mb-6">
          <button
            onClick={() => setActiveTab('file')}
            className={`flex-1 flex items-center justify-center gap-2 py-3 rounded-lg font-medium transition-colors ${
              activeTab === 'file'
                ? 'bg-olive text-cream'
                : 'bg-sand/20 text-olive hover:bg-sand/30'
            }`}
          >
            <Upload className="w-5 h-5" />
            Загрузить файл
          </button>
          <button
            onClick={() => setActiveTab('link')}
            className={`flex-1 flex items-center justify-center gap-2 py-3 rounded-lg font-medium transition-colors ${
              activeTab === 'link'
                ? 'bg-olive text-cream'
                : 'bg-sand/20 text-olive hover:bg-sand/30'
            }`}
          >
            <Link className="w-5 h-5" />
            По ссылке
          </button>
        </div>

        {/* Body Part Selection */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-dark-green mb-3">
            Категория вещи
          </label>
          <div className="grid grid-cols-2 gap-2">
            {BODY_PARTS.map(part => (
              <button
                key={part.id}
                onClick={() => setSelectedPart(part.id)}
                className={`px-4 py-3 rounded-lg font-medium transition-colors text-left ${
                  selectedPart === part.id
                    ? 'bg-olive text-cream'
                    : 'bg-sand/10 text-olive hover:bg-sand/20'
                }`}
              >
                {part.label}
              </button>
            ))}
          </div>
        </div>

        {/* File Upload */}
        {activeTab === 'file' && (
          <div className="mb-6">
            <div
              onClick={() => fileInputRef.current?.click()}
              className="border-2 border-dashed border-sand/30 rounded-xl p-8 text-center cursor-pointer hover:border-olive/50 transition-colors"
            >
              {preview ? (
                <img
                  src={preview}
                  alt="Preview"
                  className="max-h-48 mx-auto rounded-lg"
                />
              ) : (
                <>
                  <Camera className="w-12 h-12 text-sand mx-auto mb-4" />
                  <p className="text-olive font-medium mb-2">
                    Нажмите для выбора файла
                  </p>
                  <p className="text-sand text-sm">
                    или перетащите сюда (JPEG, PNG, WebP)
                  </p>
                </>
              )}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleFileSelect}
                className="hidden"
              />
            </div>
          </div>
        )}

        {/* Link Input */}
        {activeTab === 'link' && (
          <div className="mb-6">
            <label className="block text-sm font-medium text-dark-green mb-2">
              Ссылка на изображение
            </label>
            <input
              type="url"
              value={linkUrl}
              onChange={(e) => setLinkUrl(e.target.value)}
              className="input-field"
              placeholder="https://example.com/image.jpg"
            />
            <p className="text-sand text-sm mt-2">
              Мы скачаем изображение с указанного URL
            </p>
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 btn-secondary"
            disabled={isUploading}
          >
            Отмена
          </button>
          <button
            onClick={handleUpload}
            disabled={isUploading}
            className="flex-1 btn-primary flex justify-center items-center gap-2"
          >
            {isUploading ? (
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-cream"></div>
            ) : (
              'Добавить'
            )}
          </button>
        </div>
      </div>
    </div>
  )
}
