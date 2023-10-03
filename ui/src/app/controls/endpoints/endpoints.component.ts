import { Component, forwardRef } from '@angular/core';
import { AbstractControl, ControlValueAccessor, FormArray, FormControl, FormGroup, NG_VALIDATORS, NG_VALUE_ACCESSOR, ValidationErrors, Validator, Validators } from '@angular/forms';
import { Endpoint } from 'src/app/api-gen';

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

  newEndpoint(ep: Endpoint): FormGroup {
    return new FormGroup({
      name: new FormControl(ep.name, [Validators.required]),
      targetPort: new FormControl(ep.targetPort, [Validators.required, Validators.pattern("^[0-9]*$")]),
      exposure: new FormControl(ep.exposure),
      path: new FormControl(ep.path),
      protocol: new FormControl(ep.protocol),
      secure: new FormControl(ep.secure),
    });
  }

  addEndpoint() {
    this.form.push(this.newEndpoint({
      name: '',
      targetPort: 0,
    }));
  }

  removeEndpoint(index: number) {
    this.form.removeAt(index);
  }

  /* ControlValueAccessor implementation */
  writeValue(value: Endpoint[]) {
    value.forEach(ep => {
      this.form.push(this.newEndpoint(ep));
    });
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
