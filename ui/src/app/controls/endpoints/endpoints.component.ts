import { Component, forwardRef } from '@angular/core';
import { AbstractControl, ControlValueAccessor, FormArray, FormControl, FormGroup, NG_VALIDATORS, NG_VALUE_ACCESSOR, ValidationErrors, Validator, Validators } from '@angular/forms';

interface Endpoint {

}

@Component({
  selector: 'app-endpoints',
  templateUrl: './endpoints.component.html',
  styleUrls: ['./endpoints.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: EndpointsComponent
    },
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => EndpointsComponent),
      multi: true,
    },
  ]
})
export class EndpointsComponent implements ControlValueAccessor, Validator {

  onChange = (_: Endpoint[]) => {};
  onValidatorChange = () => {};

  form = new FormArray<FormGroup>([]);

  constructor() {
    this.form.valueChanges.subscribe(value => {
      this.onChange(value);
    });
  }

  newEndpoint(): FormGroup {
    return new FormGroup({
      name: new FormControl("", [Validators.required]),
      targetPort: new FormControl("", [Validators.required, Validators.pattern("^[0-9]*$")]),
      exposure: new FormControl(""),
      path: new FormControl(""),
      protocol: new FormControl(""),
      secure: new FormControl(false),
    });
  }

  addEndpoint() {
    this.form.push(this.newEndpoint());
  }

  /* ControlValueAccessor implementation */
  writeValue(value: Endpoint[]) {
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  /* Validator implementation */
  validate(control: AbstractControl): ValidationErrors | null {
    if (!this.form.valid) {
      return {'internal': true};
    }
    return null;
  }

  registerOnValidatorChange?(onValidatorChange: () => void): void {
    this.onValidatorChange = onValidatorChange;
  }
}
