import { Component, OnInit } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { Image } from 'src/app/api-gen';

@Component({
  selector: 'app-images',
  templateUrl: './images.component.html',
  styleUrls: ['./images.component.css']
})
export class ImagesComponent implements OnInit {

  forceDisplayForm: boolean = false;
  images: Image[] | undefined = [];
  editingImage: Image | undefined;

  constructor(
    private state: StateService,
    private devstate: DevstateService,
  ) {}

  ngOnInit() {
    const that = this;
    this.state.state.subscribe(async newContent => {
      that.images = newContent?.images;
      if (this.images == null) {
        return
      }
      that.forceDisplayForm = false;
    });
  }

  displayAddForm() {
    this.editingImage = undefined;
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
    if(confirm('You will delete the image "'+name+'". Continue?')) {
      const result = this.devstate.deleteImage(name);
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

  edit(image: Image) {
    this.editingImage = image;
    this.displayForm();
  }

  onCreated(image: Image) {
    const result = this.devstate.addImage(image);
    result.subscribe({
      next: value => {
        this.state.changeDevfileYaml(value);
      },
      error: error => {
        alert(error.error.message);
      }
    });
  }

  onSaved(image: Image) {
    const result = this.devstate.saveImage(image);
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
