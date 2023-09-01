import { Component, OnInit } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { Resource } from 'src/app/api-gen';

@Component({
  selector: 'app-resources',
  templateUrl: './resources.component.html',
  styleUrls: ['./resources.component.css']
})
export class ResourcesComponent implements OnInit {

  forceDisplayForm: boolean = false;
  resources: Resource[] | undefined = [];
  editingResource: Resource | undefined;

  constructor(
    private state: StateService,
    private devstate: DevstateService,
  ) {}

  ngOnInit() {
    const that = this;
    this.state.state.subscribe(async newContent => {
      that.resources = newContent?.resources;
      if (this.resources == null) {
        return
      }
      that.forceDisplayForm = false;
    });
  }

  displayAddForm() {
    this.editingResource = undefined;
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
    if(confirm('You will delete the resource "'+name+'". Continue?')) {
      const result = this.devstate.deleteResource(name);
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

  edit(resource: Resource) {
    this.editingResource = resource;
    this.displayForm();
  }

  onCreated(resource: Resource) {
    const result = this.devstate.addResource(resource);
    result.subscribe({
      next: (value) => {
        this.state.changeDevfileYaml(value);
      },
      error: (error) => {
        alert(error.error.message);
      }
    });
  }
  
  onSaved(resource: Resource) {
    const result = this.devstate.saveResource(resource);
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
