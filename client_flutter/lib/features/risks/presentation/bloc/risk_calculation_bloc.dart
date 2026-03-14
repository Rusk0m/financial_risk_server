import 'package:client_flutter/features/risks/domain/usecases/calculate_all_risk.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
part 'risk_calculation_event.dart';
part 'risk_calculation_state.dart';

class RiskCalculationBloc extends Bloc<RiskCalculationEvent, RiskCalculationState> {
  final CalculateAllRisks _calculateAllRisks;

  RiskCalculationBloc(this._calculateAllRisks) : super(const RiskCalculationState.initial()) {
    on<CalculateRiskRequested>(_onCalculateRiskRequested);
  }

  Future<void> _onCalculateRiskRequested(
    CalculateRiskRequested event,
    Emitter<RiskCalculationState> emit,
  ) async {
    emit(const RiskCalculationState.loading());
    
    try {
      final result = await _calculateAllRisks(
        CalculateAllRisksParams(
          enterpriseId: event.enterpriseId,
          horizonDays: event.horizonDays,
          confidenceLevel: event.confidenceLevel,
          priceChangePct: event.priceChangePct,
          rateChangePct: event.rateChangePct,
        ),
      );
      
      emit(RiskCalculationState.loaded(result));
    } catch (e) {
      emit(RiskCalculationState.error(e.toString()));
    }
  }
}