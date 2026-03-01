export interface User {
  id: string
  email: string
  displayName: string
  gender: 'male' | 'female' | 'unisex' | null
  createdAt: string
}

export interface Item {
  id: string
  userId: string
  part: BodyPart
  imageUrl: string
  thumbnailUrl: string
  sourceUrl?: string
  uploadedAt: string
}

export type BodyPart = 'top' | 'bottom' | 'outer' | 'shoes' | 'accessory'

export interface Look {
  id: string
  userId: string
  items: Item[]
  modelGender: 'male' | 'female'
  modelPhotoId: string
  renderedUrl: string
  score: number
  isFavorite: boolean
  isSaved: boolean
  createdAt: string
  meta: {
    colors: string[]
    style: string
    season: string
  }
}

export interface ModelPhoto {
  id: string
  gender: 'male' | 'female'
  imageUrl: string
  pose: string
}

export interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
}

export interface GenerateLooksRequest {
  itemIds: string[]
  preferences?: {
    style?: string
    season?: string
    occasion?: string
  }
}

export interface ApiError {
  error: string
  message: string
}
