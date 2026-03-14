import 'dart:io';
import 'package:client_flutter/features/reports/domain/usecases/get_all_report.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:meta/meta.dart';

import 'package:client_flutter/features/reports/domain/entities/report.dart';
import 'package:client_flutter/features/reports/domain/usecases/upload_report.dart';
import 'package:client_flutter/features/reports/domain/usecases/delete_report.dart';
import 'package:client_flutter/features/reports/domain/usecases/download_report.dart';

part 'report_event.dart';
part 'report_state.dart';

class ReportBloc extends Bloc<ReportEvent, ReportState> {
  final GetAllReports _getAll;
  final UploadReport _upload;
  final DeleteReport _delete;
  final DownloadReport _download;

  ReportBloc({
    required GetAllReports getAll,
    required UploadReport upload,
    required DeleteReport delete,
    required DownloadReport download,
  })  : _getAll = getAll,
        _upload = upload,
        _delete = delete,
        _download = download,
        super(ReportState.initial()) {
    
    on<LoadReportsEvent>(_onLoad);
    on<UploadReportEvent>(_onUpload);
    on<DeleteReportEvent>(_onDelete);
    on<DownloadReportEvent>(_onDownload);
    on<FilterReportsEvent>(_onFilter);
  }

  Future<void> _onLoad(LoadReportsEvent event, Emitter<ReportState> emit) async {
    emit(state.copyWith(status: ReportStatus.loading));
    try {
      final reports = await _getAll(type: event.type, status: event.status);
      emit(ReportState.loaded(reports, type: event.type, status: event.status));
    } catch (e) {
      emit(ReportState.error('Ошибка загрузки: $e'));
    }
  }

  Future<void> _onUpload(UploadReportEvent event, Emitter<ReportState> emit) async {
    emit(state.copyWith(uploadStatus: UploadStatus.uploading, uploadProgress: 0.0));
    try {
      await _upload(
        file: event.file,
        type: event.type,
        description: event.description,
        periodStart: event.periodStart,
        periodEnd: event.periodEnd,
      );
      emit(state.copyWith(uploadStatus: UploadStatus.success));
      // Обновляем список
      add(LoadReportsEvent(type: state.filterType, status: state.filterStatus));
    } catch (e) {
      emit(state.copyWith(
        uploadStatus: UploadStatus.error,
        uploadError: e.toString().replaceAll('Exception: ', ''),
      ));
    }
  }

  Future<void> _onDelete(DeleteReportEvent event, Emitter<ReportState> emit) async {
    try {
      await _delete(event.id);
      add(LoadReportsEvent(type: state.filterType, status: state.filterStatus));
    } catch (e) {
      emit(ReportState.error('Ошибка удаления: $e'));
    }
  }

  Future<void> _onDownload(DownloadReportEvent event, Emitter<ReportState> emit) async {
    try {
      await _download(event.id);
    } catch (e) {
      emit(ReportState.error('Ошибка скачивания: $e'));
    }
  }

  Future<void> _onFilter(FilterReportsEvent event, Emitter<ReportState> emit) async {
    emit(state.copyWith(status: ReportStatus.loading));
    try {
      final reports = await _getAll(type: event.type, status: event.status);
      emit(ReportState.loaded(reports, type: event.type, status: event.status));
    } catch (e) {
      emit(ReportState.error('Ошибка фильтрации: $e'));
    }
  }
}