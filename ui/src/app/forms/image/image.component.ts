import { Component, EventEmitter, Input, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { PATTERN_COMPONENT_ID } from '../patterns';
import { Image } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';

@Component({
  selector: 'app-image',
  templateUrl: './image.component.html',
  styleUrls: ['./image.component.css']
})
export class ImageComponent {
  @Input() cancelable: boolean = false;
  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<Image>();

  form: FormGroup;

  constructor(
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMPONENT_ID)]),
      imageName: new FormControl("", [Validators.required]),
      args: new FormControl([]),
      buildContext: new FormControl(""),
      rootRequired: new FormControl(false),
      uri: new FormControl("", [Validators.required]),
      autoBuild: new FormControl("undefined"),
    })
  }

  create() {
    this.telemetry.track("[ui] create image");
    this.created.emit(this.form.value);
  }

  cancel() {
    this.canceled.emit();
  }
}
