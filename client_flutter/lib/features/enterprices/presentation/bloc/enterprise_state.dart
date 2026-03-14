part of 'enterprise_bloc.dart';


class EnterpriseState {
  final EnterpriseStatus status;
  final List<Enterprise> enterprises;
  final Enterprise? selectedEnterprise;
  final String? error;
  final bool showSuccessMessage;

  const EnterpriseState({
    required this.status,
    required this.enterprises,
    this.selectedEnterprise,
    this.error,
    this.showSuccessMessage = false,
  });

  EnterpriseState.initial()
      : this(
          status: EnterpriseStatus.initial,
          enterprises: [],
        );

  EnterpriseState.loading()
      : this(
          status: EnterpriseStatus.loading,
          enterprises: [],
        );

  EnterpriseState.loaded(List<Enterprise> enterprises)
      : this(
          status: EnterpriseStatus.loaded,
          enterprises: enterprises,
        );

  EnterpriseState.error(String error)
      : this(
          status: EnterpriseStatus.error,
          enterprises: [],
          error: error,
        );

  const EnterpriseState.success(List<Enterprise> enterprises)
      : this(
          status: EnterpriseStatus.success,
          enterprises: enterprises,
          showSuccessMessage: true,
        );

  EnterpriseState copyWith({
    EnterpriseStatus? status,
    List<Enterprise>? enterprises,
    Enterprise? selectedEnterprise,
    String? error,
    bool? showSuccessMessage,
  }) {
    return EnterpriseState(
      status: status ?? this.status,
      enterprises: enterprises ?? this.enterprises,
      selectedEnterprise: selectedEnterprise ?? this.selectedEnterprise,
      error: error ?? this.error,
      showSuccessMessage: showSuccessMessage ?? this.showSuccessMessage,
    );
  }
}

enum EnterpriseStatus {
  initial,
  loading,
  loaded,
  creating,
  updating,
  deleting,
  success,
  error,
}