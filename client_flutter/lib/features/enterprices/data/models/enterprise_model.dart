import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:json_annotation/json_annotation.dart';

part 'enterprise_model.g.dart';
@JsonSerializable()
class EnterpriseModel {
  final int id;
  final String name;
  final String industry;
  @JsonKey(name: "annual_production_t")
  final double annualProductionT;
  @JsonKey(name: "export_share_percent")
  final double exportSharePercent;
  @JsonKey(name: "main_currency")
  final String mainCurrency;

  EnterpriseModel({
    required this.id,
    required this.name,
    required this.industry,
    required this.annualProductionT,
    required this.exportSharePercent,
    required this.mainCurrency,
  });

 factory EnterpriseModel.fromJson(Map<String, dynamic> json) => 
      _$EnterpriseModelFromJson(json);

  Map<String, dynamic> toJson() => _$EnterpriseModelToJson(this);

  // Преобразование модели → сущности
  Enterprise toEntity() => Enterprise(
        id: id,
        name: name,
        industry: industry,
        annualProductionT: annualProductionT,
        exportSharePercent: exportSharePercent,
        mainCurrency: mainCurrency,
      );

  // Преобразование сущности → модели
  factory EnterpriseModel.fromEntity(Enterprise entity) => EnterpriseModel(
        id: entity.id,
        name: entity.name,
        industry: entity.industry,
        annualProductionT: entity.annualProductionT,
        exportSharePercent: entity.exportSharePercent,
        mainCurrency: entity.mainCurrency,
      );

       static int? _toInt(dynamic value) {
    if (value == null) return null;
    if (value is int) return value;
    if (value is double) return value.toInt();
    if (value is String) return int.tryParse(value);
    return null;
  }

  static double _toDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  static String _toString(dynamic value) {
    if (value == null) return '';
    if (value is String) return value;
    return value.toString();
  }
}
