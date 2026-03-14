import 'package:client_flutter/features/enterprices/presentation/pages/enterprise_dashboard.dart';
import 'package:client_flutter/features/reports/presentation/pages/reports_list_page.dart';
import 'package:client_flutter/features/reports/presentation/pages/reports_upload_page.dart';
import 'package:flutter/material.dart';
import 'package:client_flutter/features/risks/presentation/pages/risk_calculation_page.dart';

class AppRouter {
  static Route<dynamic> generateRoute(RouteSettings settings) {
    switch (settings.name) {
      case '/':
        // Главная страница — дашборд предприятия
        return MaterialPageRoute(builder: (_) => const EnterpriseDashboard());
      case '/reports':
        return MaterialPageRoute(builder: (_) => const ReportsListPage());
      case '/risks/calculate':
        return MaterialPageRoute(builder: (_) => const RiskCalculationPage());
      // В методе generateRoute добавить:
      case '/reports/upload':
        return MaterialPageRoute(builder: (_) => const ReportUploadPage());
      default:
        return MaterialPageRoute(
          builder: (_) => Scaffold(
            body: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.error_outline, size: 64, color: Colors.grey[400]),
                  const SizedBox(height: 16),
                  Text(
                    '404 - Страница не найдена',
                    style: TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                      color: Colors.grey[600],
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
    }
  }
}
