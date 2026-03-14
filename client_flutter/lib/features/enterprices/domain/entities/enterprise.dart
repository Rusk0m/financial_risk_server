// СУЩНОСТЬ — чистая бизнес-логика БЕЗ зависимостей от фреймворков
class Enterprise {
  final int id;
  final String name;
  final String industry;
  final double annualProductionT;
  final double exportSharePercent;
  final String mainCurrency;

  Enterprise({
    required this.id,
    required this.name,
    required this.industry,
    required this.annualProductionT,
    required this.exportSharePercent,
    required this.mainCurrency,
  });

  // Бизнес-логика в сущности
  bool get isExportOriented => exportSharePercent > 50;

  double get monthlyProductionT => annualProductionT / 12;

  String get exportRiskLevel {
    if (exportSharePercent > 90) return 'very_high';
    if (exportSharePercent > 75) return 'high';
    if (exportSharePercent > 50) return 'medium';
    return 'low';
  }

  Enterprise copyWith({
    int? id,
    String? name,
    String? industry,
    double? annualProductionT,
    double? exportSharePercent,
    String? mainCurrency,
  }) {
    return Enterprise(
      id: id ?? this.id,
      name: name ?? this.name,
      industry: industry ?? this.industry,
      annualProductionT: annualProductionT ?? this.annualProductionT,
      exportSharePercent: exportSharePercent ?? this.exportSharePercent,
      mainCurrency: mainCurrency ?? this.mainCurrency,
    );
  }

  @override
  List<Object?> get props => [
        id,
        name,
        industry,
        annualProductionT,
        exportSharePercent,
        mainCurrency,
      ];
}
