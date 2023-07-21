import { Component, Inject } from '@angular/core';
import { MAT_SNACK_BAR_DATA, MatSnackBarRef } from '@angular/material/snack-bar';

@Component({
  selector: 'app-confirm',
  templateUrl: './confirm.component.html',
  styleUrls: ['./confirm.component.css']
})
export class ConfirmComponent {

  protected noLabel: string;
  protected yesLabel: string;
  protected message: string;

  constructor(
    private snackbarRef: MatSnackBarRef<ConfirmComponent>,
    @Inject(MAT_SNACK_BAR_DATA) data: any,
    ) {
      this.message = data.message;
      this.noLabel = data.noLabel;
      this.yesLabel = data.yesLabel;
    }

  cancel() {
    this.snackbarRef.dismiss();
  }

  update() {
    this.snackbarRef.dismissWithAction();
  }
}
