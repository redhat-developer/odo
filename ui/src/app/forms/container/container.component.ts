import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
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
  @Input() container: Container | undefined;

  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<ToCreate>();
  @Output() saved = new EventEmitter<ToCreate>();

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
      endpoints: new FormControl([]),
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

  toObject(o: {name: string, value: string}[]) {
    if (o == null) {
      return {};
    }
    return o.reduce((acc: any, val: {name: string, value: string}) => { acc[val.name] = val.value; return acc; }, {});
  };

  fromObject(o: any) {
    if (o == null) {
      return [];
    }
    return Object.keys(o).map(k => ({ name: k, value: o[k]}));
  }

  create() {
    this.telemetry.track("[ui] create container");

    const container = this.form.value;
    container.annotation = {
      deployment: this.toObject(container.deployAnnotations),
      service: this.toObject(container.svcAnnotations),
    };
    this.created.emit({
      container: this.form.value,
      volumes: this.volumesToCreate,
    });
  }

  save() {
    this.telemetry.track("[ui] edit container");
    const newValue = this.form.value;
    newValue.name = this.container?.name;
    newValue.annotation = {
      deployment: this.toObject(newValue.deployAnnotations),
      service: this.toObject(newValue.svcAnnotations),
    };
    this.saved.emit({
      container: newValue,
      volumes: this.volumesToCreate,
    });
  }

  cancel() {
    this.canceled.emit();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!changes['container']) {
      return;
    }
    const container = changes['container'].currentValue;
    if (container == undefined) {
      this.form.get('name')?.enable();
    } else {
      this.form.reset();
      this.form.patchValue(container);
      this.form.get('name')?.disable();
      if (this.form.get('sourceMapping')?.value != '') {
        this.form.get('_specificDir')?.setValue(true);
      }
      this.form.get('deployAnnotations')?.setValue(this.fromObject(container.annotation.deployment));
      this.form.get('svcAnnotations')?.setValue(this.fromObject(container.annotation.service));
    }
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
