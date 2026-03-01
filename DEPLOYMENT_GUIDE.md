# Руководство по деплою на бесплатный хостинг

## Вариант 1: Render.com (рекомендуется)

### Что нужно:
- Аккаунт на [render.com](https://render.com)
- Репозиторий на GitHub

### Шаги:

#### 1. Подготовка проекта

Добавьте `render.yaml` в корень проекта:

```
yaml
services:
  # Backend (Go сервер)
  - type: web
    name: fashion-look-backend
    env: go
    region: oregon
    buildCommand: go build -o server ./cmd/server
    startCommand: ./server
    envVars:
      - key: PORT
        value: 10000
      - key: JWT_SECRET
        generateValue: true
      - key: DATABASE_URL
        value: fashion.db
      - key: STORAGE_PATH
        value: /tmp/storage

  # Frontend (статика)
  - type: web
    name: fashion-look-frontend
    region: oregon
    buildCommand: cd frontend && npm install && npm run build
    staticPublishPath: frontend/dist
    envVars:
      - key: VITE_API_URL
        value: https://fashion-look-backend.onrender.com/api
```

#### 2. Деплой

1. Войдите на render.com через GitHub
2. Нажмите "New Web Service"
3. Выберите ваш репозиторий
4. Render автоматически определит Go и Node.js

---

## Вариант 2: Railway + Vercel

### Backend на Railway

1. Зарегистрируйтесь на [railway.app](https://railway.app)
2. Создайте новый проект → "Deploy from GitHub repo"
3. Настройте переменные окружения:
   - `PORT` = 8080
   - `JWT_SECRET` = ваш_секретный_ключ
   - `DATABASE_URL` = fashion.db
   - `STORAGE_PATH` = /app/storage

### Frontend на Vercel

1. Зарегистрируйтесь на [vercel.com](https://vercel.com)
2. Импортируйте репозиторий
3. Настройте:
   - Build Command: `cd frontend && npm install && npm run build`
   - Output Directory: `frontend/dist`
   - Environment Variables: `VITE_API_URL` = URL_вашего_railway_приложения

---

## Вариант 3: Fly.io

### Установка flyctl

```
bash
# Windows
iwr https://fly.io/install.ps1 | iex

# Mac
brew install flyctl
```

### Деплой

```
bash
fly auth login
fly launch
# Ответьте на вопросы:
# - App name: your-app-name
# - Region: closest to you
# - Database: No (используем SQLite в persistence volume)

# Добавьте volume для хранения
fly volume create storage --size 1

# Деплой
fly deploy
```

---

## Вариант 4: Равotá (для России)

### VPS

1. Зарегистрируйтесь на [reg.ru](https://reg.ru) или [rcloud.ru](https://rcloud.ru)
2. Арендуйте минимальный VPS (от 100₽/мес)
3. Установите Docker:

```
bash
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker
```

4. Создайте Dockerfile:

```
dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server .
COPY --from=frontend-builder /app/dist ./frontend/dist
COPY --from=frontend-builder /app/storage ./storage
EXPOSE 8080
CMD ["./server"]
```

5. Соберите и запустите:

```
bash
docker build -t fashion-look .
docker run -d -p 8080:8080 -v $(pwd)/storage:/app/storage fashion-look
```

---

## Вариант 5: Простой статический хостинг

### Frontend без сервера

Если вам нужен только frontend (без генерации образов):

1. Соберите frontend:
```
bash
cd frontend
npm run build
```

2. Загрузите папку `frontend/dist` на:
   - **Netlify** (бесплатно)
   - **Vercel** (бесплатно) 
   - **GitHub Pages** (бесплатно)

---

## Рекомендации

| Платформа | Бесплатный лимит | Подходит для |
|-----------|------------------|--------------|
| Render | 750 часов/мес | Backend + Frontend |
| Railway | $5/мес | Backend |
| Vercel | 100GB/мес | Frontend |
| Fly.io | 3 приложения | Docker |

---

## Настройка для продакшена

1. Измените JWT_SECRET на безопасный:
```
bash
openssl rand -hex 32
```

2. Для SQLite добавьтеvolume в Docker или используйте облачное хранилище

3. Настройте HTTPS (все платформы предоставляют бесплатный SSL)

---

## Частые проблемы

### Ошибка 502 Bad Gateway
- Проверьте PORT в переменных окружения
- Убедитесь, что сервер запускается

### База данных не сохраняется
- Добавьте volume для SQLite
- Или используйте PostgreSQL на Railway

### Картинки не загружаются
- Проверьте права на запись в storage
- Используйте внешнее хранилище (S3, Cloudinary)
