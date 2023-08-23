import { Component, Input, forwardRef } from '@angular/core';
import { AbstractControl, ControlValueAccessor, NG_VALIDATORS, NG_VALUE_ACCESSOR, ValidationErrors, Validator, Validators } from '@angular/forms';

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

  onChange = (_: KeyValue[]) => {};
  onValidatorChange = () => {};

  entries: KeyValue[] = [];

  writeValue(value: KeyValue[]) {
    this.entries = value;
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  addEntry() {
    this.entries.push({name: "", value: ""});
    this.onChange(this.entries);
  }

  onKeyChange(i: number, e: Event) {
    const target = e.target as HTMLInputElement;
    this.entries[i].name = target.value;
    this.onChange(this.entries);
  }

  onValueChange(i: number, e: Event) {
    const target = e.target as HTMLInputElement;
    this.entries[i].value = target.value;
    this.onChange(this.entries);
  }

  /* Validator implementation */
  validate(control: AbstractControl): ValidationErrors | null {
    for (let i=0; i<this.entries.length; i++) {
      const entry = this.entries[i];
      if (entry.name == "" || entry.value == "") {
        return {'internal': true};
      }
    }
    return null;
  }

  registerOnValidatorChange?(onValidatorChange: () => void): void {
    this.onValidatorChange = onValidatorChange;
  }
}
