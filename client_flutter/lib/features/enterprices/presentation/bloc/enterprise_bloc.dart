import 'package:client_flutter/features/enterprices/domain/entities/enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/create_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/delete_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/get_all_enterprises.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/get_by_id_enterprise.dart';
import 'package:client_flutter/features/enterprices/domain/usecases/update_enterprise.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

part 'enterprise_event.dart';
part 'enterprise_state.dart';

class EnterpriseBloc extends Bloc<EnterpriseEvent, EnterpriseState> {
  final GetAllEnterprises _getAllEnterprises;
  final GetByIdEnterprise _getByIdEnterprise;
  final CreateEnterprise _createEnterprise;
  final UpdateEnterprise _updateEnterprise;
  final DeleteEnterprise _deleteEnterprise;

  EnterpriseBloc({
    required GetAllEnterprises getAllEnterprises,
    required GetByIdEnterprise getByIdEnterprise,
    required CreateEnterprise createEnterprise,
    required UpdateEnterprise updateEnterprise,
    required DeleteEnterprise deleteEnterprise,
  })  : _getAllEnterprises = getAllEnterprises,
        _getByIdEnterprise = getByIdEnterprise,
        _createEnterprise = createEnterprise,
        _updateEnterprise = updateEnterprise,
        _deleteEnterprise = deleteEnterprise,
        super(EnterpriseState.initial()) {
    on<LoadEnterprises>(_onLoadEnterprises);
    on<CreateEnterpriseEvent>(_onCreateEnterprise);
    on<UpdateEnterpriseEvent>(_onUpdateEnterprise);
    on<DeleteEnterpriseEvent>(_onDeleteEnterprise);
    on<SelectEnterprise>(_onSelectEnterprise);
  }

  Future<void> _onLoadEnterprises(
    LoadEnterprises event,
    Emitter<EnterpriseState> emit,
  ) async {
    emit(EnterpriseState.loading());
    
    try {
      final enterprises = await _getAllEnterprises();
      emit(EnterpriseState.loaded(enterprises));
    } catch (e) {
      emit(EnterpriseState.error(e.toString()));
      // Автоматически скрываем ошибку через 5 секунд
      await Future.delayed(const Duration(seconds: 5));
      if (state.status == EnterpriseStatus.error) {
        emit(state.copyWith(status: EnterpriseStatus.loaded, error: null));
      }
    }
  }

  Future<void> _onCreateEnterprise(
    CreateEnterpriseEvent event,
    Emitter<EnterpriseState> emit,
  ) async {
    emit(state.copyWith(status: EnterpriseStatus.creating));
    
    try {
      final newEnterprise = await _createEnterprise(event.enterprise);
      
      // Обновляем список предприятий
      final updatedEnterprises = List<Enterprise>.from(state.enterprises)..add(newEnterprise);
      
      // Показываем сообщение об успехе
      emit(EnterpriseState.success(updatedEnterprises));
      
      // Через 3 секунды убираем сообщение
      await Future.delayed(const Duration(seconds: 3));
      if (state.showSuccessMessage) {
        emit(state.copyWith(status: EnterpriseStatus.loaded, showSuccessMessage: false));
      }
    } catch (e) {
      emit(state.copyWith(
        status: EnterpriseStatus.error,
        error: 'Ошибка создания: ${e.toString()}',
      ));
    }
  }

  Future<void> _onUpdateEnterprise(
    UpdateEnterpriseEvent event,
    Emitter<EnterpriseState> emit,
  ) async {
    emit(state.copyWith(status: EnterpriseStatus.updating));
    
    try {
      final updatedEnterprise = await _updateEnterprise(event.id, event.enterprise);
      
      // Обновляем предприятие в списке
      final updatedEnterprises = state.enterprises.map((enterprise) {
        return enterprise.id == event.id ? updatedEnterprise : enterprise;
      }).toList();
      
      // Обновляем выбранное предприятие, если оно было открыто
      final updatedSelected = state.selectedEnterprise?.id == event.id
          ? updatedEnterprise
          : state.selectedEnterprise;
      
      emit(state.copyWith(
        status: EnterpriseStatus.success,
        enterprises: updatedEnterprises,
        selectedEnterprise: updatedSelected,
        showSuccessMessage: true,
      ));
      
      // Через 3 секунды убираем сообщение
      await Future.delayed(const Duration(seconds: 3));
      if (state.showSuccessMessage) {
        emit(state.copyWith(status: EnterpriseStatus.loaded, showSuccessMessage: false));
      }
    } catch (e) {
      emit(state.copyWith(
        status: EnterpriseStatus.error,
        error: 'Ошибка обновления: ${e.toString()}',
      ));
    }
  }

  Future<void> _onDeleteEnterprise(
    DeleteEnterpriseEvent event,
    Emitter<EnterpriseState> emit,
  ) async {
    emit(state.copyWith(status: EnterpriseStatus.deleting));
    
    try {
      await _deleteEnterprise(event.id);
      
      // Удаляем предприятие из списка
      final updatedEnterprises = state.enterprises
          .where((enterprise) => enterprise.id != event.id)
          .toList();
      
      // Если удаляемое предприятие было выбрано, сбрасываем выбор
      final updatedSelected = state.selectedEnterprise?.id == event.id
          ? null
          : state.selectedEnterprise;
      
      emit(state.copyWith(
        status: EnterpriseStatus.success,
        enterprises: updatedEnterprises,
        selectedEnterprise: updatedSelected,
        showSuccessMessage: true,
      ));
      
      // Через 3 секунды убираем сообщение
      await Future.delayed(const Duration(seconds: 3));
      if (state.showSuccessMessage) {
        emit(state.copyWith(status: EnterpriseStatus.loaded, showSuccessMessage: false));
      }
    } catch (e) {
      emit(state.copyWith(
        status: EnterpriseStatus.error,
        error: 'Ошибка удаления: ${e.toString()}',
      ));
    }
  }

  Future<void> _onSelectEnterprise(
    SelectEnterprise event,
    Emitter<EnterpriseState> emit,
  ) async {
    if (event.enterpriseId == null) {
      // Сбрасываем выбор
      emit(state.copyWith(selectedEnterprise: null));
      return;
    }
    
    try {
      // Загружаем детали предприятия
      final enterprise = await _getByIdEnterprise(event.enterpriseId!);
      emit(state.copyWith(selectedEnterprise: enterprise));
    } catch (e) {
      emit(state.copyWith(
        error: 'Ошибка загрузки деталей: ${e.toString()}',
      ));
    }
  }
}