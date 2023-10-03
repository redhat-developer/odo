import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
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
  @Input() image: Image | undefined;

  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<Image>();
  @Output() saved = new EventEmitter<Image>();

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

  save() {
    const newValue = this.form.value;
    newValue.name = this.image?.name;
    this.telemetry.track("[ui] edit volume");
    this.saved.emit(this.form.value);
  }

  cancel() {
    this.canceled.emit();
  }

  ngOnChanges(changes: SimpleChanges) {
    console.log("changes", changes);
    if (!changes['image']) {
      return;
    }
    const img = changes['image'].currentValue;
    if (img == undefined) {
      this.form.get('name')?.enable();
    } else {
      this.form.reset();
      this.form.patchValue(img);
      this.form.get('name')?.disable();
    }
  }
}
