import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { PATTERN_COMPONENT_ID } from '../patterns';
import { DevstateService } from 'src/app/services/devstate.service';
import { Container, Volume } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';

export interface ToCreate {
  container: Container;
  volumes: Volume[];
}

@Component({
  selector: 'app-container',
  templateUrl: './container.component.html',
  styleUrls: ['./container.component.css']
})
export class ContainerComponent {
  @Input() volumeNames: string[] = [];
  @Input() cancelable: boolean = false;
  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<ToCreate>();

  form: FormGroup;

  quantityErrMsgMemory = 'Numeric value, with optional unit Ki, Mi, Gi, Ti, Pi, Ei';
  quantityErrMsgCPU = 'Numeric value, with optional unit m, k, M, G, T, P, E';

  volumesToCreate: Volume[] = [];

  constructor(
    private devstate: DevstateService,
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMPONENT_ID)]),
      image: new FormControl("", [Validators.required]),
      command: new FormControl([]),
      args: new FormControl([]),
      memoryRequest: new FormControl("", null, [this.devstate.isQuantity()]),
      memoryLimit: new FormControl("", null, [this.devstate.isQuantity()]),
      cpuRequest: new FormControl("", null, [this.devstate.isQuantity()]),
      cpuLimit: new FormControl("", null, [this.devstate.isQuantity()]),
      volumeMounts: new FormControl([]),
    })
  }

  create() {
    this.telemetry.track("[ui] create container");
    this.created.emit({
      container: this.form.value,
      volumes: this.volumesToCreate,
    });
  }

  cancel() {
    this.canceled.emit();
  }

  onCreateNewVolume(v: Volume) {
    this.volumesToCreate.push(v);
  }
}
