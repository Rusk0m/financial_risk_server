import 'package:client_flutter/di/service_locator.dart';
import 'package:client_flutter/di/service_locator.dart' as di;
import 'package:client_flutter/features/reports/presentation/bloc/report_bloc.dart';
import 'package:client_flutter/presentation/router/app_router.dart';
import 'package:client_flutter/presentation/themes/app_theme.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Инициализация зависимостей
  await setupServiceLocator();

  runApp(const FinancialRiskApp());
}

class FinancialRiskApp extends StatelessWidget {
  const FinancialRiskApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Финансовый Риск-Аналитик',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      initialRoute: '/',
      onGenerateRoute: AppRouter.generateRoute,
      builder: (context, child) {
        // Обеспечиваем доступ к размерам экрана
        return MultiBlocProvider(
          providers:[
            BlocProvider(create: (context) => di.sl<ReportBloc>()..add(LoadReportsEvent())),
          ],
          child: MediaQuery(
            data: MediaQuery.of(context).copyWith(textScaleFactor: 1.0),
            child: child!,
          ),
        );
      },
    );
  }
}
