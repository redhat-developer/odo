import { Component, EventEmitter, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { Container, DevstateService } from 'src/app/services/devstate.service';
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
    private devstate: DevstateService,
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

    const subcreate = () => {
      const result = this.devstate.addExecCommand(this.form.value["name"], this.form.value);
      result.subscribe({
        next: (value) => {
          this.state.changeDevfileYaml(value);
        },
        error: (error) => {
          alert(error.error.message);
        }
      });
    }

    if (this.containerToCreate != null && 
        this.containerToCreate?.name == this.form.controls["component"].value) {
        const res = this.devstate.addContainer(this.containerToCreate);
        res.subscribe({
          next: () => {
            subcreate();
          },
          error: error => {
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
