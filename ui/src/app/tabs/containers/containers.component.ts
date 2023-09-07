import { Component, OnInit } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { Container, Volume } from 'src/app/api-gen';
import { ToCreate } from 'src/app/forms/container/container.component';

@Component({
  selector: 'app-containers',
  templateUrl: './containers.component.html',
  styleUrls: ['./containers.component.css']
})
export class ContainersComponent implements OnInit {
  
  forceDisplayForm: boolean = false;
  containers: Container[] | undefined = [];
  volumeNames: string[] | undefined = [];

  editingContainer: Container | undefined;

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
      that.forceDisplayForm = false;
    });
  }

  displayAddForm() {
    this.editingContainer = undefined;
    this.displayForm();
  }

  displayForm() {
    this.forceDisplayForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  undisplayAddForm() {
    this.forceDisplayForm = false;
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

  edit(container: Container) {
    this.editingContainer = container;
    this.displayForm();
  }

  onCreated(toCreate: ToCreate) {
    const container = toCreate.container;
    this.createVolumes(toCreate.volumes, 0, () => {
      const result = this.devstate.addContainer(container);
      result.subscribe({
        next: value => {
          this.state.changeDevfileYaml(value);
        },
        error: error => {
          alert(error.error.message);
        }
      });  
    });
  }

  onSaved(toCreate: ToCreate) {
    const container = toCreate.container;
    this.createVolumes(toCreate.volumes, 0, () => {
      const result = this.devstate.saveContainer(container);
      result.subscribe({
        next: value => {
          this.state.changeDevfileYaml(value);
        },
        error: error => {
          alert(error.error.message);
        }
      });
    });
  }

  scrollToBottom() {
    window.scrollTo(0,document.body.scrollHeight);
  }
}
