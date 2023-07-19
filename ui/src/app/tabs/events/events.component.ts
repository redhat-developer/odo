import { Component } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { Events } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';

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
    private devstate: DevstateService,
    private telemetry: TelemetryService
  ) {}

  ngOnInit() {
    this.state.state.subscribe(async newContent => {
      this.events = newContent?.events;
      this.allCommands = newContent?.commands?.map(c => c.name);
    });
  }

  onUpdate(event: "preStart" | "postStart" | "preStop" | "postStop", commands: string[]) {
    this.telemetry.track("[ui] add "+event+" event");
    const result = this.devstate.updateEvents(event, commands);
    result.subscribe({
      next: (value) => {
        this.state.changeDevfileYaml(value);
      },
      error: (error) => {
        alert(error.error.message);
      }
    });
  }
}
