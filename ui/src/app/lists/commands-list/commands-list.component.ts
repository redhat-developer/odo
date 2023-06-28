import { Component, EventEmitter, Input } from '@angular/core';
import { MatCheckboxChange } from '@angular/material/checkbox';
import { MatSnackBar } from '@angular/material/snack-bar';
import { StateService } from 'src/app/services/state.service';
import { Command, WasmGoService } from 'src/app/services/wasm-go.service';

@Component({
  selector: 'app-commands-list',
  templateUrl: './commands-list.component.html',
  styleUrls: ['./commands-list.component.css']
})
export class CommandsListComponent {
  @Input() commands: Command[] | undefined;
  @Input() kind: string = "";
  @Input() dragDisabled: boolean = true;

  constructor(
    private wasm: WasmGoService,
    private state: StateService,
  ) {}

  toggleDefault(event: MatCheckboxChange, command: string, group: string) {
    if (event.checked) {
      this.setDefault(command, group);
    } else {
      this.unsetDefault(command);
    }
  }

  setDefault(command: string, group: string) {
    const result = this.wasm.setDefaultCommand(command, group);
    if (result.err != '') {
      alert(result.err);
    } else {
      this.state.changeDevfileYaml(result.value);
    }
  }

  unsetDefault(command: string) {
    const result = this.wasm.unsetDefaultCommand(command);
    if (result.err != '') {
      alert(result.err);
    } else {
      this.state.changeDevfileYaml(result.value);
    }
  }

  getCommandsByKind(commands: Command[] | undefined, kind: string ): Command[] | undefined {
    return commands?.filter((c: Command) => c.group == kind);
  }

  delete(command: string) {
    if(confirm('You will delete the command "'+command+'". Continue?')) {
      const result = this.wasm.deleteCommand(command);
      if (result.err != '') {
        alert(result.err);
      } else {
        this.state.changeDevfileYaml(result.value);
      }
    }
  }
}
