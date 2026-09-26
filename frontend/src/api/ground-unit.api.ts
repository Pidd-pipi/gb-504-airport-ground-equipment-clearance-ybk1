import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { ApiResponse, GroundUnit, GroundUnitSummary, PageResult, ReinspectionRecord, UnitState } from '../types';
import { API_BASE, extractData } from '../utils/request';

export interface GroundUnitPayload {
  unit_code: string;
  name: string;
  unit_type: string;
  stand: string;
  state: UnitState;
  notes: string;
}

export function groundUnitListApi(http: HttpClient, page = 1, pageSize = 50, state = '', search = ''): Observable<PageResult<GroundUnit>> {
  let params = new HttpParams().set('page', page).set('page_size', pageSize);
  if (state) params = params.set('state', state);
  if (search) params = params.set('search', search);
  return http.get<ApiResponse<PageResult<GroundUnit>>>(`${API_BASE}/v1/ground-units`, { params }).pipe(map(extractData));
}

export function groundUnitSummaryApi(http: HttpClient): Observable<GroundUnitSummary> {
  return http.get<ApiResponse<GroundUnitSummary>>(`${API_BASE}/v1/ground-units/summary`).pipe(map(extractData));
}

export function groundUnitCreateApi(http: HttpClient, payload: GroundUnitPayload): Observable<GroundUnit> {
  return http.post<ApiResponse<GroundUnit>>(`${API_BASE}/v1/ground-units`, payload).pipe(map(extractData));
}

export function groundUnitStateApi(http: HttpClient, id: number, state: UnitState, notes: string, version: number): Observable<GroundUnit> {
  return http.patch<ApiResponse<GroundUnit>>(`${API_BASE}/v1/ground-units/${id}/state`, { state, notes, version }).pipe(map(extractData));
}

export interface ReinspectionPayload {
  result: 'passed' | 'failed';
  evidence: string[];
  remark: string;
}

export function reinspectionRegisterApi(http: HttpClient, groundUnitId: number, payload: ReinspectionPayload): Observable<ReinspectionRecord> {
  return http.post<ApiResponse<ReinspectionRecord>>(`${API_BASE}/v1/ground-units/${groundUnitId}/reinspections`, payload).pipe(map(extractData));
}

export function reinspectionListApi(http: HttpClient, groundUnitId: number): Observable<ReinspectionRecord[]> {
  return http.get<ApiResponse<ReinspectionRecord[]>>(`${API_BASE}/v1/ground-units/${groundUnitId}/reinspections`).pipe(map(extractData));
}
