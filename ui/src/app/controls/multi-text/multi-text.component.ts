import {Component, forwardRef, Input} from '@angular/core';
import {
  AbstractControl,
  ControlValueAccessor,
  FormArray,
  FormControl,
  NG_VALIDATORS,
  NG_VALUE_ACCESSOR,
  ValidationErrors,
  Validator,
  Validators
} from '@angular/forms';

@Component({
  selector: 'app-multi-text',
  templateUrl: './multi-text.component.html',
  styleUrls: ['./multi-text.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: MultiTextComponent
    },
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => MultiTextComponent),
      multi: true,
    },
  ]
})
export class MultiTextComponent implements ControlValueAccessor, Validator {

  @Input() dataCyPrefix: string = "";
  @Input() label: string = "";
  @Input() addLabel: string = "";
  @Input() title: string = "";

  onChange = (_: string[]) => {};

  form = new FormArray<FormControl>([]);

  constructor() {
    this.form.valueChanges.subscribe(value => {
      this.onChange(value);
    });
  }

  newText(text: string): FormControl {
    return new FormControl(text, [Validators.required]);
  }

  writeValue(value: string[]) {
    value?.forEach(v => this.addText(v));
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  addText(text: string) {
    this.form.push(this.newText(text));
  }

  removeText(index: number) {
    this.form.removeAt(index);
  }

  /* Validator implementation */
  validate(control: AbstractControl): ValidationErrors | null {
    if (!this.form.valid) {
      return {'internal': true};
    }
    return null;
  }
}
