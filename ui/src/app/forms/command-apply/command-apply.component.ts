import { Component, EventEmitter, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { ClusterResource, WasmGoService } from 'src/app/services/wasm-go.service';
import { PATTERN_COMMAND_ID } from '../patterns';

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
  resourceToCreate: ClusterResource | null = null;

  constructor(
    private wasm: WasmGoService,
    private state: StateService,
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
    if (this.resourceToCreate != null && 
      this.resourceToCreate?.name == this.form.controls["component"].value) {
      const result = this.wasm.addResource(this.resourceToCreate);
      if (result.err != '') {
        alert(result.err);
        return;
      }
    }

    const result = this.wasm.addApplyCommand(this.form.value["name"], this.form.value);
    if (result.err != '') {
      alert(result.err);      
    } else {
      this.state.changeDevfileYaml(result.value);
    }
  }

  cancel() {
    this.canceled.emit();
  }

  onCreateNewContainer(v: boolean) {
    this.showNewResource = v;
  }

  onNewResourceCreated(resource: ClusterResource) {
    this.resourceList.push(resource.name);
    this.form.controls["component"].setValue(resource.name);
    this.showNewResource = false;
    this.resourceToCreate = resource;
  }
}
