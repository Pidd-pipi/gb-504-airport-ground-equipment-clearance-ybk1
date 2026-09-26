import { CommonModule } from '@angular/common';
import { Component, EventEmitter, Input, OnChanges, Output, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { turnaroundReadinessApi } from '../../api/turnaround.api';
import { StatusBadgeComponent } from './status-badge.component';
import { Turnaround, TurnaroundReadiness } from '../../types';
import { parseHttpError, useHttp } from '../../utils/request';

@Component({
  selector: 'app-turnaround-readiness',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatIconModule, MatProgressBarModule, StatusBadgeComponent],
  template: `
    <section class="readiness" *ngIf="turnaround as row">
      <header>
        <span><mat-icon>rule</mat-icon>周转放行条件</span>
        <button mat-button type="button" (click)="load()"><mat-icon>refresh</mat-icon>刷新</button>
      </header>
      <mat-progress-bar *ngIf="loading()" mode="indeterminate"></mat-progress-bar>
      <p class="error" *ngIf="error()">{{ error() }}</p>
      <ng-container *ngIf="data() as readiness">
        <div class="summary-line">
          <app-status-badge [value]="readiness.clearance_state"></app-status-badge>
          <span [class.blocked]="!readiness.ready_for_decision">待处理检查 {{ readiness.pending_checks }}</span>
          <span [class.blocked]="readiness.failed_checks > 0">未通过 {{ readiness.failed_checks }}</span>
          <span [class.blocked]="readiness.expired_unit_codes.length" class="expired">复检过期 {{ readiness.expired_unit_codes.length }}</span>
        </div>
        <div class="full-clearance" [class.ok]="readiness.ready_for_full_clearance" [class.no]="!readiness.ready_for_full_clearance">
          <mat-icon>{{ readiness.ready_for_full_clearance ? 'check_circle' : 'block' }}</mat-icon>
          <strong>{{ readiness.ready_for_full_clearance ? '满足完全放行条件' : '完全放行被拦截' }}</strong>
        </div>
        <h4>投入设备复检有效期（{{ row.risk_level === 'high' || row.risk_level === 'critical' ? '高/严重风险航班 8 小时' : '普通风险航班 24 小时' }}）</h4>
        <ul class="unit-list">
          <li *ngFor="let unit of readiness.unit_validities" [class.expired]="unit.expired">
            <mat-icon>{{ unit.expired ? 'error' : 'verified' }}</mat-icon>
            <span class="unit-copy"><strong>{{ unit.unit_code }}</strong><small>有效期自 {{ unit.validity_from ? (unit.validity_from | date:'MM-dd HH:mm') : '无检查记录' }} · {{ unit.window_hours }}h 窗口</small></span>
            <app-status-badge [value]="unit.unit_state"></app-status-badge>
            <span class="tag" *ngIf="unit.expired">复检过期</span>
          </li>
        </ul>
        <ul class="blockers" *ngIf="readiness.blockers.length">
          <li *ngFor="let blocker of readiness.blockers">{{ blocker }}</li>
        </ul>
      </ng-container>
    </section>
  `,
  styles: [`
    .readiness { border: 1px solid #d7e1e3; border-radius: 5px; background: #f8fbfb; }
    header { display: flex; align-items: center; justify-content: space-between; padding: 8px 12px; border-bottom: 1px solid #e0e7e8; font-weight: 600; color: #273940; }
    header > span { display: flex; align-items: center; gap: 6px; }
    header mat-icon { color: #0d8b82; font-size: 19px; width: 19px; height: 19px; }
    .summary-line { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 10px 12px 4px; font-size: 12px; color: #576a71; }
    .summary-line .blocked { color: #b42318; font-weight: 600; }
    .summary-line .expired { color: #b42318; }
    .full-clearance { display: flex; align-items: center; gap: 7px; margin: 8px 12px; padding: 8px 10px; border-radius: 4px; font-size: 13px; }
    .full-clearance.ok { color: #16835a; background: #e7f6ed; }
    .full-clearance.no { color: #b42318; background: #fdecea; }
    h4 { margin: 10px 12px 4px; font-size: 12px; color: #43565e; }
    .unit-list { list-style: none; margin: 0; padding: 0 12px; }
    .unit-list li { display: flex; align-items: center; gap: 8px; padding: 7px 0; border-bottom: 1px dashed #e3e9ea; font-size: 12px; }
    .unit-list li:last-child { border-bottom: 0; }
    .unit-list mat-icon { font-size: 17px; width: 17px; height: 17px; color: #16835a; }
    .unit-list li.expired mat-icon { color: #b42318; }
    .unit-copy { display: flex; flex-direction: column; flex: 1; }
    .unit-copy small { color: #7d8c92; }
    .tag { padding: 1px 7px; border-radius: 4px; background: #fdecea; color: #b42318; font-size: 11px; }
    .blockers { margin: 4px 12px 12px 26px; padding: 0; color: #a63b32; font-size: 12px; }
    .error { padding: 10px 12px; margin: 0; color: #b42318; font-size: 12px; }
  `],
})
export class TurnaroundReadinessComponent implements OnChanges {
  @Input({ required: true }) turnaround: Turnaround | null = null;
  @Output() dataChange = new EventEmitter<TurnaroundReadiness>();
  private readonly http = useHttp();
  readonly data = signal<TurnaroundReadiness | null>(null);
  readonly loading = signal(false);
  readonly error = signal('');

  ngOnChanges(): void { this.load(); }

  load(): void {
    if (!this.turnaround) return;
    this.data.set(null);
    this.loading.set(true);
    this.error.set('');
    turnaroundReadinessApi(this.http, this.turnaround.id).subscribe({
      next: result => { this.data.set(result); this.dataChange.emit(result); this.loading.set(false); },
      error: err => { this.error.set(parseHttpError(err)); this.loading.set(false); },
    });
  }
}
