import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { PATTERN_COMMAND_ID } from '../patterns';
import { Command, Container, Volume } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';
import { ToCreate } from '../container/container.component';

@Component({
  selector: 'app-command-exec',
  templateUrl: './command-exec.component.html',
  styleUrls: ['./command-exec.component.css']
})
export class CommandExecComponent {
  @Input() command: Command | undefined;

  @Output() canceled = new EventEmitter<void>();

  form: FormGroup;
  containerList: string[] = [];
  showNewContainer: boolean = false;
  containerToCreate: Container | null = null;
  volumesToCreate: Volume[] = [];
  volumeNames: string[] | undefined = [];

  constructor(
    private devstate: DevstateService,
    private state: StateService,
    private telemetry: TelemetryService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMMAND_ID)]),
      component: new FormControl("", [Validators.required]),
      commandLine: new FormControl("", [Validators.required]),
      workingDir: new FormControl("", [Validators.required]),
      hotReloadCapable: new FormControl(false),
    });

    this.state.state.subscribe(async newContent => {
      this.volumeNames = newContent?.volumes.map((v: Volume) => v.name);
      const containers = newContent?.containers;
      if (containers == null) {
        return
      }
      this.containerList = containers.map(container => container.name);
    });
  }

  createVolumes(volumes: Volume[], i: number, next: () => any) {
    if (volumes.length == i) {
      next();
      return;
    }
    const res = this.devstate.addVolume(volumes[i]);
      res.subscribe({
        next: value => {
          this.createVolumes(volumes, i+1, next);
        },
        error: error => {
          alert(error.error.message);
        }
      });
  }

  create() {
    this.telemetry.track("[ui] create exec command");
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

    this.createVolumes(this.volumesToCreate, 0, () => {
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
    });
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

  onNewContainerCreated(toCreate: ToCreate) {
    const container = toCreate.container;
    this.containerList.push(container.name);
    this.form.controls["component"].setValue(container.name);
    this.showNewContainer = false;
    this.containerToCreate = container;
    this.volumesToCreate.push(...toCreate.volumes);
  }

  ngOnChanges(changes: SimpleChanges) {
    if (!changes['command']) {
      return;
    }
    const cmd = changes['command'].currentValue;
    if (cmd == undefined) {
      this.form.get('name')?.enable();
    } else {
      this.form.reset();
      this.form.patchValue(cmd);
      this.form.patchValue(cmd.exec);
      this.form.get('name')?.disable();
    }
  }

  save() {
    this.telemetry.track("[ui] update exec command");
    const subcreate = () => {
      if (this.command == undefined) {
        return;
      }
      const result = this.devstate.updateExecCommand(this.command.name, this.form.value);
      result.subscribe({
        next: (value) => {
          this.state.changeDevfileYaml(value);
        },
        error: (error) => {
          alert(error.error.message);
        }
      });
    }

    this.createVolumes(this.volumesToCreate, 0, () => {
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
    });
  }
}
