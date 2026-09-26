import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { RiskBadgeComponent } from './risk-badge.component';
import { StatusBadgeComponent } from './status-badge.component';
import { TurnaroundReadinessComponent } from './turnaround-readiness.component';
import { Turnaround } from '../../types';

@Component({
  selector: 'app-turnaround-detail-dialog',
  standalone: true,
  imports: [CommonModule, MatButtonModule, MatDialogModule, MatIconModule, RiskBadgeComponent, StatusBadgeComponent, TurnaroundReadinessComponent],
  template: `
    <h2 mat-dialog-title>
      <mat-icon>flight</mat-icon>周转 #{{ data.turnaround.id }} · {{ data.turnaround.flight_no }}
      <app-risk-badge [level]="data.turnaround.risk_level"></app-risk-badge>
      <app-status-badge [value]="data.turnaround.status"></app-status-badge>
    </h2>
    <mat-dialog-content>
      <app-turnaround-readiness [turnaround]="data.turnaround"></app-turnaround-readiness>
      <p class="tip"><mat-icon>info</mat-icon>存在复检过期设备时完全放行会被系统拦截；可选择限制放行，并在运行条件中写清逾期设备的复检条件。</p>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-flat-button (click)="dialog.close()">关闭</button>
    </mat-dialog-actions>
  `,
  styles: [`
    h2 { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-size: 17px; }
    h2 mat-icon { color: #0d8b82; }
    mat-dialog-content { min-width: min(560px, 88vw); }
    .tip { display: flex; align-items: flex-start; gap: 5px; margin: 10px 2px 0; color: #8a5b00; font-size: 12px; }
    .tip mat-icon { font-size: 16px; width: 16px; height: 16px; }
  `],
})
export class TurnaroundDetailDialogComponent {
  constructor(
    public readonly dialog: MatDialogRef<TurnaroundDetailDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public readonly data: { turnaround: Turnaround },
  ) {}
}
