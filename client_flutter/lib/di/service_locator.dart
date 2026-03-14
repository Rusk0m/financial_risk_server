import 'package:client_flutter/core/network/api_client.dart';
import 'package:client_flutter/features/enterprices/data/repositories/enterprice_repository_impl.dart';
import 'package:client_flutter/features/enterprices/domain/repositories/enterprise_repository.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/create_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/delete_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/get_all_enterprises.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/get_by_id_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/update_enterprise.dart';
import 'package:client_flutter/features/enterprices/presentation/bloc/enterprise_bloc.dart';
import 'package:client_flutter/features/reports/data/repositories/report_repository_impl.dart';
import 'package:client_flutter/features/reports/domain/repositories/report_repository.dart';
import 'package:client_flutter/features/reports/domain/usecases/delete_report.dart';
import 'package:client_flutter/features/reports/domain/usecases/download_report.dart';
import 'package:client_flutter/features/reports/domain/usecases/get_all_report.dart';
import 'package:client_flutter/features/reports/domain/usecases/upload_report.dart';
import 'package:client_flutter/features/reports/presentation/bloc/report_bloc.dart';
import 'package:client_flutter/features/risks/data/repositories/risk_repository_impl.dart';
import 'package:client_flutter/features/risks/domain/repositories/risk_repository.dart';
import 'package:client_flutter/features/risks/domain/usecases/calculate_all_risk.dart';
import 'package:client_flutter/features/risks/presentation/bloc/risk_calculation_bloc.dart';
import 'package:get_it/get_it.dart';

final sl = GetIt.instance;

Future<void> setupServiceLocator() async {
  // Core
  // Core
  sl.registerLazySingleton<ApiClient>(() => ApiClient());

  // Enterprise Feature
  sl.registerLazySingleton<EnterpriseRepository>(
    () => EnterpriseRepositoryImpl(sl()),
  );
  sl.registerLazySingleton<GetAllEnterprises>(() => GetAllEnterprises(sl()));
  sl.registerLazySingleton<CreateEnterprise>(() => CreateEnterprise(sl()));
  sl.registerLazySingleton<UpdateEnterprise>(() => UpdateEnterprise(sl()));
  sl.registerLazySingleton<DeleteEnterprise>(() => DeleteEnterprise(sl()));
  sl.registerLazySingleton<GetByIdEnterprise>(() => GetByIdEnterprise(sl()));

  // Risk Feature
  sl.registerLazySingleton<RiskRepository>(() => RiskRepositoryImpl(sl()));
  sl.registerLazySingleton<CalculateAllRisks>(() => CalculateAllRisks(sl()));

  // Reports Feature
  sl.registerLazySingleton<ReportRepository>(() => ReportRepositoryImpl(sl()));

  sl.registerLazySingleton<UploadReport>(() => UploadReport(sl()));
  sl.registerLazySingleton<GetAllReports>(() => GetAllReports(sl()));
  sl.registerLazySingleton<DownloadReport>(() => DownloadReport(sl()));
  sl.registerLazySingleton<DeleteReport>(() => DeleteReport(sl()));

  // BLoCs
  sl.registerFactory<ReportBloc>(
    () => ReportBloc(
      upload: sl(),
      getAll: sl(),
      download: sl(),
      delete: sl(),
    ),
  );
  // BLoCs
  sl.registerFactory<EnterpriseBloc>(
    () => EnterpriseBloc(
      getAllEnterprises: sl(),
      createEnterprise: sl(),
      getByIdEnterprise: sl(),
      updateEnterprise: sl(),
      deleteEnterprise: sl(),
    ),
  );
  sl.registerFactory<RiskCalculationBloc>(() => RiskCalculationBloc(sl()));
}
