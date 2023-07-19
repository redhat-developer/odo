import { Component, EventEmitter, Output } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { PATTERN_COMMAND_ID } from '../patterns';
import { Image } from 'src/app/api-gen';
import { SegmentService } from 'ngx-segment-analytics';

@Component({
  selector: 'app-command-image',
  templateUrl: './command-image.component.html',
  styleUrls: ['./command-image.component.css']
})
export class CommandImageComponent {
  @Output() canceled = new EventEmitter<void>();

  form: FormGroup;
  imageList: string[] = [];
  showNewImage: boolean = false;
  imageToCreate: Image | null = null;

  constructor(
    private devstate: DevstateService,
    private state: StateService,
    private segment: SegmentService
  ) {
    this.form = new FormGroup({
      name: new FormControl("", [Validators.required, Validators.pattern(PATTERN_COMMAND_ID)]),
      component: new FormControl("", [Validators.required]),
    });

    this.state.state.subscribe(async newContent => {
      const images = newContent?.images;
      if (images == null) {
        return
      }
      this.imageList = images.map(image => image.name);
    });
  }

  create() {
    this.segment.track("[ui] create image command");
    const subcreate = () => {
      const result = this.devstate.addApplyCommand(this.form.value["name"], this.form.value);
      result.subscribe({
        next: (value) => {
          this.state.changeDevfileYaml(value);
        },
        error: (error) => {
          alert(error.error.message);
        }
      });
    }

    if (this.imageToCreate != null && 
      this.imageToCreate?.name == this.form.controls["component"].value) {
        const result = this.devstate.addImage(this.imageToCreate);
        result.subscribe({
          next: () => {
            subcreate();
          },
          error: error => {
            alert(error.error.message);
          }
        });
    } else {
      subcreate();
    }
  }

  cancel() {
    this.canceled.emit();
  }

  onCreateNewImage(v: boolean) {
    this.showNewImage = v;
  }

  onNewImageCreated(image: Image) {
    this.imageList.push(image.name);
    this.form.controls["component"].setValue(image.name);
    this.showNewImage = false;
    this.imageToCreate = image;
  }
}
