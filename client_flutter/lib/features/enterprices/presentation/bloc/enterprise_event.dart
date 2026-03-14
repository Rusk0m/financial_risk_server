part of 'enterprise_bloc.dart';

 class EnterpriseEvent {}

final class LoadEnterprises extends EnterpriseEvent {}

final class CreateEnterpriseEvent extends EnterpriseEvent {
  final Enterprise enterprise;

  CreateEnterpriseEvent(this.enterprise);
}

final class UpdateEnterpriseEvent extends EnterpriseEvent {
  final int id;
  final Enterprise enterprise;

  UpdateEnterpriseEvent(this.id, this.enterprise);
}

final class DeleteEnterpriseEvent extends EnterpriseEvent {
  final int id;

  DeleteEnterpriseEvent(this.id);
}

final class SelectEnterprise extends EnterpriseEvent {
  final int? enterpriseId;

  SelectEnterprise(this.enterpriseId);
}