import { Link } from 'react-router-dom'
import { Sparkles, Shirt, Camera, Palette } from 'lucide-react'






export default function LandingPage() {
  return (
    <div className="min-h-screen bg-cream">
      {/* Header */}
      <header className="bg-dark-green text-cream py-4">
        <div className="container mx-auto px-6 flex justify-between items-center">
          <h1 className="text-2xl font-bold">Fashion Look Generator</h1>
          <div className="space-x-4">
            <Link 
              to="/login" 
              className="text-cream hover:text-sand transition-colors"
            >
              Войти
            </Link>
            <Link 
              to="/register" 
              className="bg-olive text-cream px-4 py-2 rounded-lg hover:bg-sand transition-colors"
            >
              Регистрация
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="py-20 px-6">
        <div className="container mx-auto max-w-6xl">
          <div className="grid md:grid-cols-2 gap-12 items-center">
            <div>
              <h2 className="text-5xl font-bold text-dark-green mb-6 leading-tight">
                Создавайте идеальные образы из вашего гардероба
              </h2>
              <p className="text-xl text-olive mb-8">
                Загружайте фотографии своих вещей, и наша система автоматически 
                подберет стильные сочетания, адаптированные под ваш тип.
              </p>
              <div className="flex gap-4">
                <Link 
                  to="/register" 
                  className="btn-primary text-lg inline-block"
                >
                  Начать бесплатно
                </Link>
                <Link 
                  to="/login" 
                  className="btn-secondary text-lg inline-block"
                >
                  Уже есть аккаунт
                </Link>
              </div>
            </div>
            
            {/* Hero Illustration */}
            <div className="rounded-2xl overflow-hidden shadow-xl">
              <img 
                src="/landing.jpg" 
                alt="Fashion Look Generator" 
                className="w-full h-auto object-cover"
              />
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 bg-white">
        <div className="container mx-auto px-6 max-w-6xl">
          <h3 className="text-3xl font-bold text-dark-green text-center mb-12">
            Как это работает
          </h3>
          <div className="grid md:grid-cols-4 gap-8">
            <FeatureCard 
              icon={<Camera className="w-10 h-10" />}
              title="Загрузите вещи"
              description="Фотографируйте или добавляйте ссылки на вашу одежду"
            />
            <FeatureCard 
              icon={<Sparkles className="w-10 h-10" />}
              title="Авто-анализ"
              description="Система определяет пол и анализирует стиль"
            />
            <FeatureCard 
              icon={<Palette className="w-10 h-10" />}
              title="Умный подбор"
              description="Алгоритм создает гармоничные сочетания"
            />
            <FeatureCard 
              icon={<Shirt className="w-10 h-10" />}
              title="Визуализация"
              description="Смотрите, как образы смотрятся на моделях"
            />
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-16 bg-olive text-cream">
        <div className="container mx-auto px-6 text-center">
          <h3 className="text-3xl font-bold mb-4">
            Готовы обновить свой гардероб?
          </h3>
          <p className="text-xl mb-8 opacity-90">
            Присоединяйтесь к тысячам пользователей, которые уже используют Fashion Look Generator
          </p>
          <Link 
            to="/register" 
            className="bg-cream text-dark-green px-8 py-4 rounded-lg font-semibold text-lg hover:bg-sand transition-colors inline-block"
          >
            Создать аккаунт
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-dark-green text-cream/60 py-8">
        <div className="container mx-auto px-6 text-center">
          <p>&copy; 2024 Fashion Look Generator. Все права защищены.</p>
        </div>
      </footer>
    </div>
  )
}

function FeatureCard({ icon, title, description }: { 
  icon: React.ReactNode
  title: string
  description: string 
}) {
  return (
    <div className="card text-center hover:shadow-lg transition-shadow">
      <div className="text-olive mb-4 flex justify-center">{icon}</div>
      <h4 className="text-xl font-semibold text-dark-green mb-2">{title}</h4>
      <p className="text-olive/80">{description}</p>
    </div>
  )
}
