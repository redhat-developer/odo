import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { PATTERN_COMPONENT_ID } from '../patterns';
import { DevstateService } from 'src/app/services/devstate.service';
import { Annotation, Container, Volume } from 'src/app/api-gen';
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
  seeMore: boolean = false;

  constructor(
    private devstate: DevstateService,
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMPONENT_ID)]),
      image: new FormControl("", [Validators.required]),
      command: new FormControl([]),
      args: new FormControl([]),
      env: new FormControl([]),
      volumeMounts: new FormControl([]),
      memoryRequest: new FormControl("", null, [this.devstate.isQuantity()]),
      memoryLimit: new FormControl("", null, [this.devstate.isQuantity()]),
      cpuRequest: new FormControl("", null, [this.devstate.isQuantity()]),
      cpuLimit: new FormControl("", null, [this.devstate.isQuantity()]),
      configureSources: new FormControl(false),
      mountSources: new FormControl(true),
      _specificDir: new FormControl(false),
      sourceMapping: new FormControl(""),
      deployAnnotations: new FormControl([]),
      svcAnnotations: new FormControl([]),
    });

    this.form.valueChanges.subscribe((value: any) => {
      this.updateSourceFields(value);
    });
    this.updateSourceFields(this.form.value);
  }

  updateSourceFields(value: any) {
    const sourceMappingEnabled = value.mountSources && value._specificDir;
    if (!sourceMappingEnabled && !this.form.get('sourceMapping')?.disabled) {
      this.form.get('sourceMapping')?.disable();
      this.form.get('sourceMapping')?.setValue('');
      this.form.get('_specificDir')?.setValue(false);
    }       
    if (sourceMappingEnabled && !this.form.get('sourceMapping')?.enabled ) {
      this.form.get('sourceMapping')?.enable();
    }

    const specificDirEnabled = value.mountSources;
    if (!specificDirEnabled && !this.form.get('_specificDir')?.disabled) {
      this.form.get('_specificDir')?.disable();
    }       
    if (specificDirEnabled && !this.form.get('_specificDir')?.enabled ) {
      this.form.get('_specificDir')?.enable();
    }
  }

  create() {
    this.telemetry.track("[ui] create container");

    const toObject = (o: {name: string, value: string}[]) => {
      return o.reduce((acc: any, val: {name: string, value: string}) => { acc[val.name] = val.value; return acc; }, {});
    };

    const container = this.form.value;
    container.annotation = {
      deployment: toObject(container.deployAnnotations),
      service: toObject(container.svcAnnotations),
    };
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

  more() {
    this.seeMore = true;
  }
  less() {
    this.seeMore = false;
  }
}
