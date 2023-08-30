import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { Volume } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';
import { PATTERN_COMPONENT_ID } from '../patterns';
import { DevstateService } from 'src/app/services/devstate.service';

@Component({
  selector: 'app-volume',
  templateUrl: './volume.component.html',
  styleUrls: ['./volume.component.css']
})
export class VolumeComponent {
  @Input() cancelable: boolean = false;
  @Input() volume: Volume | undefined;
  
  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<Volume>();
  @Output() saved = new EventEmitter<Volume>();

  form: FormGroup;

  constructor(
    private devstate: DevstateService,
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMPONENT_ID)]),
      size: new FormControl("", null, [this.devstate.isQuantity()]),
      ephemeral: new FormControl(false),
    })
  }

  create() {
    this.telemetry.track("[ui] create volume");
    this.created.emit(this.form.value);
  }

  save() {
    const newValue = this.form.value;
    newValue.name = this.volume?.name;
    this.telemetry.track("[ui] edit volume");
    this.saved.emit(this.form.value);
  }

  cancel() {
    this.canceled.emit();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!changes['volume']) {
      return;
    }
    const vol = changes['volume'].currentValue;
    if (vol == undefined) {
      this.form.get('name')?.enable();
    } else {
      this.form.reset();
      this.form.patchValue(vol);
      this.form.get('name')?.disable();
    }
  }
}
