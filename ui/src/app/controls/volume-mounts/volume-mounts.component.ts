import {Component, EventEmitter, forwardRef, Input, Output} from '@angular/core';
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
import {Volume, VolumeMount} from 'src/app/api-gen';

@Component({
  selector: 'app-volume-mounts',
  templateUrl: './volume-mounts.component.html',
  styleUrls: ['./volume-mounts.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: VolumeMountsComponent
    },
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => VolumeMountsComponent),
      multi: true,
    },
  ]
})
export class VolumeMountsComponent implements ControlValueAccessor, Validator {

  @Input() volumes: string[] = [];
  
  @Output() createNewVolume = new EventEmitter<Volume>();

  form = new FormArray<FormGroup>([]);

  showNewVolume: boolean[] = [];

  onChange = (_: VolumeMount[]) => {};
  onValidatorChange = () => {};

  constructor() {
    this.form.valueChanges.subscribe(value => {
      this.onChange(value);
    });
  }

  writeValue(value: VolumeMount[]) {
    value.forEach(v => this.add(v.name, v.path));
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  newVolumeMount(vol: VolumeMount): FormGroup {
    return new FormGroup({
      name: new FormControl(vol.name, [Validators.required]),
      path: new FormControl(vol.path, [Validators.required]),
    });
  }

  add(name: string, path: string) {
    this.form.push(this.newVolumeMount({name, path}));
  }

  onNameChange(i: number, name: string) {
    this.showNewVolume[i] = name == "!";
  }

  onNewVolumeCreated(i: number, v: Volume) {
    this.volumes.push(v.name);
    this.form.at(i).get('name')?.setValue(v.name);
    this.createNewVolume.next(v);
    this.showNewVolume[i] = false;
    this.onValidatorChange();
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
