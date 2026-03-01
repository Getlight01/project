import axios from 'axios'
import { useAuthStore } from '../store/authStore'

// Use relative API path - will work both locally (via Vite proxy) and production (via Go)
const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api

export const authApi = {
  register: (data: { email: string; password: string; displayName: string }) =>
    api.post('/register', data),
  
  login: (data: { email: string; password: string }) =>
    api.post('/login', data),
}

export const itemsApi = {
  upload: (formData: FormData) =>
    api.post('/upload-item', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
  
  addLink: (data: { url: string; part: string }) =>
    api.post('/add-item-link', data),
  
  getUserItems: () =>
    api.get('/user-items'),
  
  deleteItem: (id: string) =>
    api.delete(`/item/${id}`),
}

export const looksApi = {
  generate: (data: { itemIds: string[]; preferences?: object }) =>
    api.post('/generate-looks', data),
  
  getLook: (id: string) =>
    api.get(`/look/${id}`),
  
  getUserLooks: () =>
    api.get('/user-looks'),
  
  updateLook: (id: string, data: { isFavorite?: boolean; isSaved?: boolean }) =>
    api.patch(`/look/${id}`, data),
  
  deleteLook: (id: string) =>
    api.delete(`/look/${id}`),
}

export const modelsApi = {
  getModels: (gender?: string) =>
    api.get('/models', { params: { gender } }),
}
