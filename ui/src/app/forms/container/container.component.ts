import { Component, EventEmitter, Input, Output } from '@angular/core';
import { AbstractControl, AsyncValidatorFn, FormControl, FormGroup, ValidationErrors, ValidatorFn, Validators } from '@angular/forms';
import { PATTERN_COMPONENT_ID } from '../patterns';
import { DevstateService } from 'src/app/services/devstate.service';
import { Observable, of, map, catchError } from 'rxjs';
import { Container } from 'src/app/api-gen';
import { SegmentService } from 'ngx-segment-analytics';

@Component({
  selector: 'app-container',
  templateUrl: './container.component.html',
  styleUrls: ['./container.component.css']
})
export class ContainerComponent {
  @Input() cancelable: boolean = false;
  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<Container>();

  form: FormGroup;

  quantityErrMsgMemory = 'Numeric value, with optional unit Ki, Mi, Gi, Ti, Pi, Ei';
  quantityErrMsgCPU = 'Numeric value, with optional unit m, k, M, G, T, P, E';

  constructor(
    private devstate: DevstateService,
    private segment: SegmentService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMPONENT_ID)]),
      image: new FormControl("", [Validators.required]),
      command: new FormControl([]),
      args: new FormControl([]),
      memoryRequest: new FormControl("", null, [this.isQuantity()]),
      memoryLimit: new FormControl("", null, [this.isQuantity()]),
      cpuRequest: new FormControl("", null, [this.isQuantity()]),
      cpuLimit: new FormControl("", null, [this.isQuantity()]),
    })
  }

  create() {
    this.segment.track("[ui] create container");
    this.created.emit(this.form.value);
  }

  cancel() {
    this.canceled.emit();
  }

  isQuantity():  AsyncValidatorFn {
    return (control: AbstractControl): Observable<ValidationErrors | null> => {
      const val = control.value;
      if (val == '') {
        return of(null);
      }
      const valid = this.devstate.isQuantityValid(val);
      return valid.pipe(
        map(() => null),
        catchError(() => of({"isQuantity": false}))
      );
    };
  }   
}
