import { Component, EventEmitter, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { PATTERN_COMMAND_ID } from '../patterns';
import { Resource } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';

@Component({
  selector: 'app-command-apply',
  templateUrl: './command-apply.component.html',
  styleUrls: ['./command-apply.component.css']
})
export class CommandApplyComponent {
  @Output() canceled = new EventEmitter<void>();

  form: FormGroup;
  resourceList: string[] = [];
  showNewResource: boolean = false;
  resourceToCreate: Resource | null = null;

  constructor(
    private devstate: DevstateService,
    private state: StateService,
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMMAND_ID)]),
      component: new FormControl("", [Validators.required]),
    });

    this.state.state.subscribe(async newContent => {
      const resources = newContent?.resources;
      if (resources == null) {
        return
      }
      this.resourceList = resources.map(resource => resource.name);
    });
  }

  create() {
    this.telemetry.track("[ui] create apply command");
    const subcreate = () => {
      const result = this.devstate.addApplyCommand(this.form.value["name"], this.form.value);
      result.subscribe({
        next: (value) => {
          this.state.changeDevfileYaml(value);
        },
        error: (error) => {
          alert(error.error.message);
        }
      });  
    }

    if (this.resourceToCreate != null && 
      this.resourceToCreate?.name == this.form.controls["component"].value) {
      const result = this.devstate.addResource(this.resourceToCreate);
      result.subscribe({
        next: (value) => {
          this.state.changeDevfileYaml(value);
          subcreate();
        },
        error: (error) => {
          alert(error.error.message);
        }
      });        
    } else {
      subcreate();
    }
  }

  cancel() {
    this.canceled.emit();
  }

  onCreateNewContainer(v: boolean) {
    this.showNewResource = v;
  }

  onNewResourceCreated(resource: Resource) {
    this.resourceList.push(resource.name);
    this.form.controls["component"].setValue(resource.name);
    this.showNewResource = false;
    this.resourceToCreate = resource;
  }
}
