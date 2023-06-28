import { Component, Input } from '@angular/core';
import { NG_VALUE_ACCESSOR } from '@angular/forms';

@Component({
  selector: 'app-multi-command',
  templateUrl: './multi-command.component.html',
  styleUrls: ['./multi-command.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: MultiCommandComponent
    }
  ]
})
export class MultiCommandComponent {

  @Input() addLabel: string = "";
  @Input() commandList: string[] = [];
  @Input() title: string = "";

  onChange = (_: string[]) => {};

  commands: string[] = [];

  writeValue(value: any) {
    this.commands = value;
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  addCommand() {
    this.commands.push("");
    this.onChange(this.commands);
  }

  onCommandChange(i: number, cmd: string) {
    this.commands[i] = cmd;
    this.onChange(this.commands);
  }
}
