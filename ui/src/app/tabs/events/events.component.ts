import { Component } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { Events, WasmGoService } from 'src/app/services/wasm-go.service';

@Component({
  selector: 'app-events',
  templateUrl: './events.component.html',
  styleUrls: ['./events.component.css']
})
export class EventsComponent {
  
  events: Events | undefined;
  allCommands: string[] | undefined;

  constructor(
    private state: StateService,
    private wasm: WasmGoService,
  ) {}

  ngOnInit() {
    this.state.state.subscribe(async newContent => {
      this.events = newContent?.events;
      this.allCommands = newContent?.commands.map(c => c.name);
    });
  }

  onUpdate(event: "preStart" | "postStart" | "preStop" | "postStop", commands: string[]) {
    const result = this.wasm.updateEvents(event, commands);
    if (result.err != '') {
      alert(result.err);
    } else {
      this.state.changeDevfileYaml(result.value);
    }
  }
}
