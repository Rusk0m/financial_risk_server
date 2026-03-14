import 'package:flutter/material.dart';
import 'package:client_flutter/core/constants/app_colors.dart';
import 'package:client_flutter/presentation/widgets/app_bar.dart';

class LoginPage extends StatelessWidget {
  const LoginPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        foregroundColor: AppColors.textPrimary,
        title: const Text(
          'Финансовый Риск-Аналитик',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 400),
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // Логотип
                Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFF0284C7), Color(0xFF0369A1)],
                    ),
                    shape: BoxShape.circle,
                  ),
                  child: const Center(
                    child: Text(
                      'FR',
                      style: TextStyle(
                        fontSize: 32,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 32),
                
                // Заголовок
                const Text(
                  'Вход в систему',
                  style: TextStyle(
                    fontSize: 28,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 8),
                const Center(
                  child: Text(
                  'Анализ финансовых рисков ОАО «Беларуськалий»',
                  style: TextStyle(
                    fontSize: 16,
                    color: AppColors.textSecondary,
                  ),
                ),
                ),
                const SizedBox(height: 32),
                
                // Форма входа
                _buildLoginForm(context),
                
                const SizedBox(height: 24),
                
                // Демо-режим
                TextButton(
                  onPressed: () => _loginAsDemoUser(context),
                  child: const Text(
                    'Войти как демо-пользователь',
                    style: TextStyle(color: AppColors.primary),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildLoginForm(BuildContext context) {
    return Column(
      children: [
        TextFormField(
          decoration: InputDecoration(
            labelText: 'Логин',
            prefixIcon: const Icon(Icons.person),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        const SizedBox(height: 16),
        TextFormField(
          decoration: InputDecoration(
            labelText: 'Пароль',
            prefixIcon: const Icon(Icons.lock),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
          obscureText: true,
        ),
        const SizedBox(height: 24),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () => _login(context),
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
              backgroundColor: AppColors.primary,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
            child: const Text('Войти'),
          ),
        ),
      ],
    );
  }

  void _login(BuildContext context) {
    // В реальном приложении здесь будет вызов API
    // Для демо показываем сообщение об ошибке (т.к. нет реального бэкенда аутентификации)
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Для демо-режима используйте кнопку ниже'),
        backgroundColor: Colors.blue,
      ),
    );
  }

  void _loginAsDemoUser(BuildContext context) {
    // Сохраняем флаг демо-режима
    // В реальном приложении здесь будет сохранение токена
    Navigator.pushNamedAndRemoveUntil(context, '/', (route) => false);
    
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Вход выполнен успешно! Добро пожаловать в демо-режим.'),
        backgroundColor: Colors.green,
      ),
    );
  }
}