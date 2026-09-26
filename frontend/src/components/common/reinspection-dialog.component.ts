import { CommonModule } from '@angular/common';
import { Component, Inject, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar } from '@angular/material/snack-bar';
import { reinspectionRegisterApi } from '../../api/ground-unit.api';
import { GroundUnit } from '../../types';
import { parseHttpError, useHttp } from '../../utils/request';

export interface ReinspectionDialogData {
  unit: GroundUnit;
}

@Component({
  selector: 'app-reinspection-dialog',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule, MatButtonModule, MatDialogModule, MatFormFieldModule,
    MatIconModule, MatInputModule, MatSelectModule,
  ],
  template: `
    <h2 mat-dialog-title><mat-icon>fact_check</mat-icon>设备复检登记 · {{ data.unit.unit_code }}</h2>
    <mat-dialog-content>
      <p class="hint">复检只更新有效期锚点与审计记录，不改变设备当前状态“{{ data.unit.state }}”。</p>
      <form [formGroup]="form" (ngSubmit)="submit()" id="reinspection-form">
        <mat-form-field appearance="outline"><mat-label>复检结论</mat-label><mat-select formControlName="result"><mat-option value="passed">复检通过（刷新有效期）</mat-option><mat-option value="failed">复检未通过（不刷新有效期）</mat-option></mat-select></mat-form-field>
        <mat-form-field appearance="outline"><mat-label>复检证据编号 / 文件名</mat-label><input matInput formControlName="evidence" placeholder="多个证据用逗号分隔，如 reinspect-TUG-017.jpg"></mat-form-field>
        <mat-form-field appearance="outline"><mat-label>复检说明</mat-label><textarea matInput rows="3" formControlName="remark" placeholder="可记录复检项目、现场情况"></textarea></mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button (click)="dialog.close(false)" [disabled]="saving">取消</button>
      <button mat-flat-button type="submit" form="reinspection-form" [disabled]="form.invalid || saving"><mat-icon>save</mat-icon>{{ saving ? '提交中…' : '保存复检' }}</button>
    </mat-dialog-actions>
  `,
  styles: [`
    h2 { display: flex; align-items: center; gap: 8px; font-size: 17px; }
    h2 mat-icon { color: #0d8b82; }
    mat-dialog-content { min-width: min(440px, 80vw); }
    .hint { margin: 4px 0 12px; padding: 8px 10px; color: #5d7077; background: #eef6f5; border-left: 3px solid #0e9187; border-radius: 3px; font-size: 12px; }
    mat-form-field { width: 100%; }
  `],
})
export class ReinspectionDialogComponent {
  readonly dialog = inject(MatDialogRef<ReinspectionDialogComponent>);
  private readonly fb = inject(FormBuilder);
  private readonly http = useHttp();
  private readonly snack = inject(MatSnackBar);
  readonly form = this.fb.nonNullable.group({
    result: ['passed' as 'passed' | 'failed', Validators.required],
    evidence: ['', Validators.required],
    remark: [''],
  });
  saving = false;

  constructor(@Inject(MAT_DIALOG_DATA) public readonly data: ReinspectionDialogData) {}

  submit(): void {
    if (this.form.invalid) return;
    const value = this.form.getRawValue();
    this.saving = true;
    reinspectionRegisterApi(this.http, this.data.unit.id, {
      result: value.result,
      evidence: value.evidence.split(',').map(item => item.trim()).filter(Boolean),
      remark: value.remark.trim(),
    }).subscribe({
      next: record => {
        this.snack.open('复检记录已保存，设备状态保持不变', '关闭', { duration: 2500 });
        this.dialog.close(record);
      },
      error: error => {
        this.saving = false;
        this.snack.open(parseHttpError(error), '关闭', { duration: 4500 });
      },
    });
  }
}
