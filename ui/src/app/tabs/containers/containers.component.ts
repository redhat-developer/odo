import { Component, OnInit } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { Container, Volume } from 'src/app/api-gen';

@Component({
  selector: 'app-containers',
  templateUrl: './containers.component.html',
  styleUrls: ['./containers.component.css']
})
export class ContainersComponent implements OnInit {
  
  forceDisplayAdd: boolean = false;
  containers: Container[] | undefined = [];
  volumeNames: string[] | undefined = [];

  constructor(
    private state: StateService,
    private devstate: DevstateService,
  ) {}

  ngOnInit() {
    const that = this;
    this.state.state.subscribe(async newContent => {
      this.volumeNames = newContent?.volumes.map((v: Volume) => v.name);
      that.containers = newContent?.containers;
      if (this.containers == null) {
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
    if(confirm('You will delete the container "'+name+'". Continue?')) {
      const result = this.devstate.deleteContainer(name);
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

  onCreated(container: Container) {
    const result = this.devstate.addContainer(container);
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
