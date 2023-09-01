import {Component, EventEmitter, forwardRef, Input, Output} from '@angular/core';
import {
  AbstractControl,
  ControlValueAccessor,
  FormArray, FormControl,
  FormGroup, NG_VALIDATORS,
  NG_VALUE_ACCESSOR,
  ValidationErrors, Validator, Validators
} from '@angular/forms';

@Component({
  selector: 'app-select-container',
  templateUrl: './select-container.component.html',
  styleUrls: ['./select-container.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: SelectContainerComponent
    },
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => SelectContainerComponent),
      multi: true,
    },
  ]
})
export class SelectContainerComponent implements ControlValueAccessor, Validator {

  @Input() containers: string[] = [];
  @Input() label: string = "";
  @Output() createNew = new EventEmitter<boolean>();

  formCtrl: FormControl;

  onChange = (_: string) => {};

  constructor() {
    this.formCtrl = new FormControl('', [Validators.required]);
  }

  writeValue(value: string) {
    this.formCtrl.setValue(value);
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  onSelectChange(v: string) {
    if (v != "!") {
      this.onChange(v);
    }
    this.createNew.emit(v == "!");
  }

  /* Validator implementation */
  validate(control: AbstractControl): ValidationErrors | null {
    if (!this.formCtrl.valid) {
      return {'internal': true};
    }
    return null;
  }
}
