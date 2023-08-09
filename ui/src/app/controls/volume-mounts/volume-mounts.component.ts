import { Component, Input } from '@angular/core';
import { NG_VALUE_ACCESSOR } from '@angular/forms';
import { VolumeMount } from 'src/app/api-gen';

@Component({
  selector: 'app-volume-mounts',
  templateUrl: './volume-mounts.component.html',
  styleUrls: ['./volume-mounts.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: VolumeMountsComponent
    }
  ]
})
export class VolumeMountsComponent {

  @Input() volumes: string[] = [];
  
  volumeMounts: VolumeMount[] = [];

  onChange = (_: VolumeMount[]) => {};

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
    this.volumeMounts[i].name = name;
    this.onChange(this.volumeMounts);
  }
}
