import { Component, EventEmitter, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { PATTERN_COMMAND_ID } from '../patterns';

@Component({
  selector: 'app-command-composite',
  templateUrl: './command-composite.component.html',
  styleUrls: ['./command-composite.component.css']
})
export class CommandCompositeComponent {
  @Output() canceled = new EventEmitter<void>();

  form: FormGroup;
  commandList: string[] = [];

  constructor(
    private devstate: DevstateService,
    private state: StateService,
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMMAND_ID)]),
      parallel: new FormControl(false),
      commands: new FormControl([])
    });

    this.state.state.subscribe(async newContent => {
      const commands = newContent?.commands;
      if (commands == null) {
        return
      }
      this.commandList = commands.map(command => command.name);
    });

  }

  create() {
    const result = this.devstate.addCompositeCommand(this.form.value["name"], this.form.value);
    result.subscribe({
      next: (value) => {
        this.state.changeDevfileYaml(value);
      },
      error: (error) => {
        alert(error.error.message);
      }
    });
   }
 
  cancel() {
    this.canceled.emit();
  }
}
