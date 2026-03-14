// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'risk_result_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

RiskResultModel _$RiskResultModelFromJson(Map<String, dynamic> json) =>
    RiskResultModel(
      id: (json['id'] as num).toInt(),
      enterpriseId: (json['enterpriseId'] as num).toInt(),
      riskType: json['riskType'] as String,
      calculationDate: DateTime.parse(json['calculationDate'] as String),
      horizonDays: (json['horizonDays'] as num).toInt(),
      confidenceLevel: (json['confidenceLevel'] as num).toDouble(),
      exposureAmount: (json['exposureAmount'] as num).toDouble(),
      varValue: (json['varValue'] as num).toDouble(),
      stressTestLoss: (json['stressTestLoss'] as num).toDouble(),
      riskLevel: json['riskLevel'] as String,
      recommendations: (json['recommendations'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      scenarioType: json['scenarioType'] as String,
    );

Map<String, dynamic> _$RiskResultModelToJson(RiskResultModel instance) =>
    <String, dynamic>{
      'id': instance.id,
      'enterpriseId': instance.enterpriseId,
      'riskType': instance.riskType,
      'calculationDate': instance.calculationDate.toIso8601String(),
      'horizonDays': instance.horizonDays,
      'confidenceLevel': instance.confidenceLevel,
      'exposureAmount': instance.exposureAmount,
      'varValue': instance.varValue,
      'stressTestLoss': instance.stressTestLoss,
      'riskLevel': instance.riskLevel,
      'recommendations': instance.recommendations,
      'scenarioType': instance.scenarioType,
    };
