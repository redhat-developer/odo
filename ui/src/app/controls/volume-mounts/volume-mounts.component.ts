import { Component, EventEmitter, Input, Output, forwardRef } from '@angular/core';
import { AbstractControl, NG_VALIDATORS, NG_VALUE_ACCESSOR, ValidationErrors, Validator } from '@angular/forms';
import { Volume, VolumeMount } from 'src/app/api-gen';

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
export class VolumeMountsComponent implements Validator {

  @Input() volumes: string[] = [];
  
  @Output() createNewVolume = new EventEmitter<Volume>();

  volumeMounts: VolumeMount[] = [];
  showNewVolume: boolean[] = [];

  onChange = (_: VolumeMount[]) => {};
  onValidatorChange = () => {};

  writeValue(value: any) {
    this.volumeMounts = value;
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  add() {
    this.volumeMounts.push({name: "", path: ""});
    this.onChange(this.volumeMounts);  
  }

  onPathChange(i: number, e: Event) {
    const target = e.target as HTMLInputElement;
    this.volumeMounts[i].path = target.value;
    this.onChange(this.volumeMounts);
  }

  onNameChange(i: number, name: string) {
    if (name != "!") {
      this.volumeMounts[i].name = name;
      this.onChange(this.volumeMounts);
    } 

    this.showNewVolume[i] = name == "!";
  }

  onNewVolumeCreated(i: number, v: Volume) {
    this.volumes.push(v.name);
    this.volumeMounts[i].name = v.name;
    this.createNewVolume.next(v);
    this.showNewVolume[i] = false;
    this.onValidatorChange();
  }

  /* Validator implementation */
  validate(control: AbstractControl): ValidationErrors | null {
    for (let i=0; i<this.volumeMounts.length; i++) {
      const vm = this.volumeMounts[i];
      if (vm.name == "" || vm.path == "") {
        return {'internal': true};
      }
    }
    return null;
  }

  registerOnValidatorChange?(onValidatorChange: () => void): void {
    this.onValidatorChange = onValidatorChange;
  }
}
