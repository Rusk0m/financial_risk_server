// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'enterprise_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

EnterpriseModel _$EnterpriseModelFromJson(Map<String, dynamic> json) =>
    EnterpriseModel(
      id: (json['id'] as num).toInt(),
      name: json['name'] as String,
      industry: json['industry'] as String,
      annualProductionT: (json['annual_production_t'] as num).toDouble(),
      exportSharePercent: (json['export_share_percent'] as num).toDouble(),
      mainCurrency: json['main_currency'] as String,
    );

Map<String, dynamic> _$EnterpriseModelToJson(EnterpriseModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'industry': instance.industry,
      'annual_production_t': instance.annualProductionT,
      'export_share_percent': instance.exportSharePercent,
      'main_currency': instance.mainCurrency,
    };
