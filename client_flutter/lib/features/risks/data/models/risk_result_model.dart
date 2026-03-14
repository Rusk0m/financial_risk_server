import 'package:json_annotation/json_annotation.dart';

part 'risk_result_model.g.dart';

@JsonSerializable()
class RiskResultModel {
  final int id;
  final int enterpriseId;
  final String riskType;
  final DateTime calculationDate;
  final int horizonDays;
  final double confidenceLevel;
  final double exposureAmount;
  final double varValue;
  final double stressTestLoss;
  final String riskLevel;
  final List<String> recommendations;
  final String scenarioType;

  RiskResultModel({
    required this.id,
    required this.enterpriseId,
    required this.riskType,
    required this.calculationDate,
    required this.horizonDays,
    required this.confidenceLevel,
    required this.exposureAmount,
    required this.varValue,
    required this.stressTestLoss,
    required this.riskLevel,
    required this.recommendations,
    required this.scenarioType,
  });

  factory RiskResultModel.fromJson(Map<String, dynamic> json) => 
      _$RiskResultModelFromJson(json);

  Map<String, dynamic> toJson() => _$RiskResultModelToJson(this);
}