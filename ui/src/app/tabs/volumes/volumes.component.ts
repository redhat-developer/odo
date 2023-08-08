import { Component } from '@angular/core';
import { Volume } from 'src/app/api-gen';
import { DevstateService } from 'src/app/services/devstate.service';
import { StateService } from 'src/app/services/state.service';

@Component({
  selector: 'app-volumes',
  templateUrl: './volumes.component.html',
  styleUrls: ['./volumes.component.css']
})
export class VolumesComponent {

  forceDisplayAdd: boolean = false;
  volumes: Volume[] | undefined = [];

  constructor(
    private state: StateService,
    private devstate: DevstateService,
  ) {}

  ngOnInit() {
    const that = this;
    this.state.state.subscribe(async newContent => {
      that.volumes = newContent?.volumes;
      if (this.volumes == null) {
        return
      }
      that.forceDisplayAdd = false;
    });
  }

  displayAddForm() {
    this.forceDisplayAdd = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  undisplayAddForm() {
    this.forceDisplayAdd = false;
  }

  delete(name: string) {
    if(confirm('You will delete the volume "'+name+'". Continue?')) {
      const result = this.devstate.deleteVolume(name);
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

  onCreated(volume: Volume) {
    const result = this.devstate.addVolume(volume);
    result.subscribe({
      next: value => {
        this.state.changeDevfileYaml(value);
      },
      error: error => {
        alert(error.error.message);
      }
    });
  }

  scrollToBottom() {
    window.scrollTo(0,document.body.scrollHeight);
  }
}
