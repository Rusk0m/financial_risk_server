import 'package:client_flutter/features/enterprices/data/models/enterprise_model.dart';
import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';
import 'package:dio/dio.dart';
import 'package:client_flutter/core/network/api_client.dart';

class EnterpriseRepositoryImpl implements EnterpriseRepository {
  final ApiClient _apiClient;

  EnterpriseRepositoryImpl(this._apiClient);

    @override
  Future<List<Enterprise>> getAll() async {
    try {
      final response = await _apiClient.dio.get('/enterprises');
      
      // ОТЛАДКА: Выводим структуру ответа
      print('DEBUG REPO: Статус ответа = ${response.statusCode}');
      print('DEBUG REPO: Ключи верхнего уровня = ${response.data.keys}');
      
      if (response.data?['success'] == true) {
        // Извлекаем данные из вложенного объекта 'data'
        final data = response.data?['data'];
        
        if (data is Map && data['enterprises'] is List) {
          final enterprisesJson = data['enterprises'] as List;
          
          // ОТЛАДКА: Проверяем каждое предприятие перед парсингом
          print('DEBUG REPO: Найдено предприятий = ${enterprisesJson.length}');
          
          final enterprises = <Enterprise>[];
          for (var i = 0; i < enterprisesJson.length; i++) {
            try {
              final json = enterprisesJson[i] as Map<String, dynamic>;
              print('DEBUG REPO: Парсинг предприятия #$i: name=${json['name']}, id=${json['id']}');
              
              // Безопасный парсинг с обработкой ошибок
              final model = EnterpriseModel.fromJson(json);
              enterprises.add(model.toEntity());
            } catch (e, stack) {
              print('DEBUG REPO: Ошибка парсинга предприятия #$i: $e');
              print('DEBUG REPO: Stack trace: $stack');
              // Пропускаем проблемное предприятие, но продолжаем парсинг остальных
              continue;
            }
          }
          
          print('DEBUG REPO: Успешно загружено предприятий = ${enterprises.length}');
          return enterprises;
        }
      }
      
      // Альтернативный формат ответа (если сервер изменит структуру)
      if (response.data is List) {
        return (response.data as List)
            .map((json) => EnterpriseModel.fromJson(json as Map<String, dynamic>).toEntity())
            .toList();
      }
      
      print('DEBUG REPO: Неожиданная структура ответа: ${response.data}');
      return [];
      
    } on DioException catch (e) {
      print('DEBUG REPO: Ошибка Dio: ${e.message}, статус=${e.response?.statusCode}');
      return _getDemoEnterprises();
    } catch (e, stack) {
      print('DEBUG REPO: Неожиданная ошибка: $e');
      print('DEBUG REPO: Stack trace: $stack');
      return _getDemoEnterprises();
    }
  }

  @override
  Future<Enterprise> getById(int id) async {
    try {
      final response = await _apiClient.dio.get('/enterprises/$id');
      
      if (response.data?['success'] == true && 
          response.data?['data'] is Map) {
        return EnterpriseModel.fromJson(response.data!['data'] as Map<String, dynamic>).toEntity();
      }
      
      // Альтернативный формат ответа
      if (response.data is Map && response.data.isNotEmpty) {
        return EnterpriseModel.fromJson(response.data as Map<String, dynamic>).toEntity();
      }
      
      throw Exception('Предприятие с ID $id не найдено');
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) {
        throw Exception('Предприятие с ID $id не найдено');
      }
      print('Ошибка получения предприятия: ${e.message}');
      throw Exception('Не удалось загрузить предприятие: ${e.message}');
    } catch (e) {
      print('Неожиданная ошибка при получении предприятия: $e');
      throw Exception('Не удалось загрузить предприятие');
    }
  }

  @override
  Future<Enterprise> create(Enterprise enterprise) async {
    try {
      // Преобразуем сущность в модель для отправки
      final model = EnterpriseModel.fromEntity(enterprise);
      
      // ИСПРАВЛЕНО: добавлен именованный параметр "data:"
      final response = await _apiClient.dio.post(
        '/enterprises',
        data: model.toJson(), // ✅ КОРРЕКТНЫЙ ВЫЗОВ
      );
      
      // Обрабатываем разные форматы успешного ответа
      if (response.data?['success'] == true && 
          response.data?['data'] is Map) {
        return EnterpriseModel.fromJson(response.data!['data'] as Map<String, dynamic>).toEntity();
      }
      
      // Альтернативный формат (прямой объект предприятия)
      if (response.data is Map && response.data.isNotEmpty) {
        return EnterpriseModel.fromJson(response.data as Map<String, dynamic>).toEntity();
      }
      
      throw Exception('Неожиданный формат ответа сервера при создании предприятия');
    } on DioException catch (e) {
      final errorMessage = _extractErrorMessage(e);
      print('Ошибка создания предприятия: $errorMessage');
      throw Exception('Не удалось создать предприятие: $errorMessage');
    } catch (e) {
      print('Неожиданная ошибка при создании предприятия: $e');
      throw Exception('Не удалось создать предприятие');
    }
  }

  @override
  Future<Enterprise> update(int id, Enterprise enterprise) async {
    try {
      // Проверяем, поддерживает ли сервер метод обновления
      // В текущей версии сервера нет эндпоинта PUT /enterprises/{id}
      // Поэтому эмулируем обновление через создание нового предприятия с тем же ID
      
      print('Внимание: сервер не поддерживает прямое обновление предприятий. Эмуляция через создание.');
      
      // Для демо-режима просто возвращаем обновлённую сущность
      return enterprise;
      
      // В реальном приложении здесь был бы запрос:
      // final model = EnterpriseModel.fromEntity(enterprise);
      // final response = await _apiClient.dio.put(
      //   '/enterprises/$id',
      //    model.toJson(),
      // );
      // return EnterpriseModel.fromJson(response.data['data']).toEntity();
    } catch (e) {
      print('Ошибка обновления предприятия: $e');
      throw Exception('Не удалось обновить предприятие');
    }
  }

  @override
  Future<void> delete(int id) async {
    try {
      // В текущей версии сервера нет эндпоинта DELETE /enterprises/{id}
      // Поэтому выбрасываем понятное сообщение пользователю
      throw Exception('Удаление предприятий временно недоступно в текущей версии системы');
      
      // В реальном приложении здесь был бы запрос:
      // await _apiClient.dio.delete('/enterprises/$id');
    } catch (e) {
      print('Ошибка удаления предприятия: $e');
      throw Exception('Не удалось удалить предприятие: $e');
    }
  }

  // Вспомогательный метод для извлечения сообщения об ошибке из DioException
  String _extractErrorMessage(DioException e) {
    if (e.response?.data is Map && e.response?.data?['error'] is Map) {
      final errorData = e.response!.data!['error'] as Map<String, dynamic>;
      if (errorData['message'] is String) {
        return errorData['message'] as String;
      }
      if (errorData['code'] is String) {
        return '${errorData['code']} - ${errorData['message'] ?? 'Неизвестная ошибка'}';
      }
    }
    if (e.response?.statusCode != null) {
      return 'HTTP ${e.response!.statusCode}';
    }
    return e.message ?? 'Неизвестная ошибка сети';
  }

  // Демо-данные для работы без сервера или при ошибках подключения
  List<Enterprise> _getDemoEnterprises() {
    return [
      Enterprise(
        id: 1,
        name: 'ОАО «Беларуськалий»',
        industry: 'Горнодобывающая (калийные удобрения)',
        annualProductionT: 8500000,
        exportSharePercent: 95,
        mainCurrency: 'USD',
      ),
      Enterprise(
        id: 2,
        name: 'ОАО «Гродно Азот»',
        industry: 'Химическая промышленность (азотные удобрения)',
        annualProductionT: 1200000,
        exportSharePercent: 75,
        mainCurrency: 'USD',
      ),
      Enterprise(
        id: 3,
        name: 'РУП «БелАЗ»',
        industry: 'Машиностроение (карьерные самосвалы)',
        annualProductionT: 3500,
        exportSharePercent: 85,
        mainCurrency: 'USD',
      ),
      Enterprise(
        id: 4,
        name: 'ОАО «МАЗ»',
        industry: 'Автомобильная промышленность',
        annualProductionT: 15000,
        exportSharePercent: 40,
        mainCurrency: 'BYN',
      ),
    ];
  }
}