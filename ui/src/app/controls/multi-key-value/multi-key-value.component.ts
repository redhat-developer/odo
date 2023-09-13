import {Component, forwardRef, Input} from '@angular/core';
import {
  AbstractControl,
  ControlValueAccessor,
  FormArray,
  FormControl,
  FormGroup,
  NG_VALIDATORS,
  NG_VALUE_ACCESSOR,
  ValidationErrors,
  Validator,
  Validators
} from '@angular/forms';

interface KeyValue {
  name: string;
  value: string;
}

@Component({
  selector: 'app-multi-key-value',
  templateUrl: './multi-key-value.component.html',
  styleUrls: ['./multi-key-value.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: MultiKeyValueComponent
    },
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => MultiKeyValueComponent),
      multi: true,
    },
  ]
})
export class MultiKeyValueComponent implements ControlValueAccessor, Validator {

  @Input() dataCyPrefix: string = "";
  @Input() addLabel: string = "";

  form = new FormArray<FormGroup>([]);

  onChange = (_: KeyValue[]) => {};
  onValidatorChange = () => {};

  constructor() {
    this.form.valueChanges.subscribe(value => {
      this.onChange(value);
    });
  }

  writeValue(value: KeyValue[]) {
    value.forEach(v => this.addEntry(v.name, v.value));
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  newKeyValueForm(kv: KeyValue): FormGroup {
    return new FormGroup({
      name: new FormControl(kv.name, [Validators.required]),
      value: new FormControl(kv.value, [Validators.required]),
    });
  }

  addEntry(name: string, value: string) {
    this.form.push(this.newKeyValueForm({name, value}));
  }

  removeEntry(index: number) {
    this.form.removeAt(index);
  }

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
