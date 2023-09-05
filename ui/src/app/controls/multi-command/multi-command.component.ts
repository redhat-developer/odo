import {Component, forwardRef, Input} from '@angular/core';
import {
  AbstractControl,
  ControlValueAccessor,
  FormArray,
  FormControl,
  FormGroup,
  NG_VALIDATORS,
  NG_VALUE_ACCESSOR, ValidationErrors, Validator,
  Validators
} from '@angular/forms';

@Component({
  selector: 'app-multi-command',
  templateUrl: './multi-command.component.html',
  styleUrls: ['./multi-command.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: MultiCommandComponent
    },
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => MultiCommandComponent),
      multi: true,
    },
  ]
})
export class MultiCommandComponent implements ControlValueAccessor, Validator {

  @Input() addLabel: string = "";
  @Input() commandList: string[] = [];
  @Input() title: string = "";

  onChange = (_: string[]) => {};

  form = new FormArray<FormControl>([]);

  constructor() {
    this.form.valueChanges.subscribe(value => {
      this.onChange(value);
    });
  }

  writeValue(value: string[]) {
    value.forEach(v => this.addCommand(v));
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  newCommand(cmdName : string) {
    return new FormControl(cmdName, [Validators.required]);
  }

  addCommand(cmdName: string) {
    this.form.push(this.newCommand(cmdName));
  }

  /* Validator implementation */
  validate(control: AbstractControl): ValidationErrors | null {
    if (!this.form.valid) {
      return {'internal': true};
    }
    return null;
  }
}
