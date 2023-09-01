import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { PATTERN_COMPONENT_ID } from '../patterns';
import { Resource } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';

@Component({
  selector: 'app-resource',
  templateUrl: './resource.component.html',
  styleUrls: ['./resource.component.css']
})
export class ResourceComponent {
  @Input() cancelable: boolean = false;
  @Input() resource: Resource | undefined;
  
  @Output() canceled = new EventEmitter<void>();
  @Output() created = new EventEmitter<Resource>();
  @Output() saved = new EventEmitter<Resource>();

  form: FormGroup;
  uriOrInlined: string = 'uri';

  constructor(
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMPONENT_ID)]),
      _choice: new FormControl("uri"),
      uri: new FormControl("", [Validators.required]),
      inlined: new FormControl("", []),
      deployByDefault: new FormControl("undefined"),
    })
  }

  changeUriOrInlined(value: string) {
    this.uriOrInlined = value;
    if (this.uriOrInlined == 'uri') {
      this.form.controls['inlined'].removeValidators(Validators.required);
      this.form.controls['inlined'].setValue('');
      
      this.form.controls['uri']?.addValidators(Validators.required);
    } else if (this.uriOrInlined == 'inlined') {
      this.form.controls['uri']?.removeValidators(Validators.required);
      this.form.controls['uri'].setValue('');

      this.form.controls['inlined']?.setValidators(Validators.required);
    }
    this.form.controls['uri'].updateValueAndValidity()
    this.form.controls['inlined'].updateValueAndValidity()
  }

  create() {
    this.telemetry.track("[ui] create resource");
    this.created.emit(this.form.value);
  }

  save() {
    const newValue = this.form.value;
    newValue.name = this.resource?.name;
    this.telemetry.track("[ui] edit resource");
    this.saved.emit(this.form.value);
  }

  cancel() {
    this.canceled.emit();    
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!changes['resource']) {
      return;
    }
    const res = changes['resource'].currentValue;
    if (res == undefined) {
      this.form.get('name')?.enable();
    } else {
      this.form.reset();
      this.form.patchValue(res);
      if (res['inlined']) {
        this.form.get('_choice')?.setValue('inlined');
        this.changeUriOrInlined('inlined');
      }
      this.form.get('name')?.disable();
    }
  }
}
