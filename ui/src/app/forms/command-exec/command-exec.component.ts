import { Component, EventEmitter, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { Container, WasmGoService } from 'src/app/services/wasm-go.service';
import { PATTERN_COMMAND_ID } from '../patterns';

@Component({
  selector: 'app-command-exec',
  templateUrl: './command-exec.component.html',
  styleUrls: ['./command-exec.component.css']
})
export class CommandExecComponent {
  @Output() canceled = new EventEmitter<void>();

  form: FormGroup;
  containerList: string[] = [];
  showNewContainer: boolean = false;
  containerToCreate: Container | null = null;

  constructor(
    private wasm: WasmGoService,
    private state: StateService,
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMMAND_ID)]),
      component: new FormControl("", [Validators.required]),
      commandLine: new FormControl("", [Validators.required]),
      workingDir: new FormControl("", [Validators.required]),
      hotReloadCapable: new FormControl(false),
    });

    this.state.state.subscribe(async newContent => {
      const containers = newContent?.containers;
      if (containers == null) {
        return
      }
      this.containerList = containers.map(container => container.name);
    });
  }

  create() {
    if (this.containerToCreate != null && 
        this.containerToCreate?.name == this.form.controls["component"].value) {
      const result = this.wasm.addContainer(this.containerToCreate);
      if (result.err != '') {
        alert(result.err);
        return;
      }
    }
    const result = this.wasm.addExecCommand(this.form.value["name"], this.form.value);
    if (result.err != '') {
      alert(result.err);      
    } else {
      this.state.changeDevfileYaml(result.value);
    }
  }

  cancel() {
    this.canceled.emit();
  }

  onProjectsRoot() {
    this.form.controls['workingDir'].setValue('${PROJECTS_ROOT}');
  }

  onCreateNewContainer(v: boolean) {
    this.showNewContainer = v;
  }

  onNewContainerCreated(container: Container) {
    this.containerList.push(container.name);
    this.form.controls["component"].setValue(container.name);
    this.showNewContainer = false;
    this.containerToCreate = container;
  }
}
