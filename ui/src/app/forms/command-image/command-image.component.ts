import { Component, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { PATTERN_COMMAND_ID } from '../patterns';
import { Command, Image } from 'src/app/api-gen';
import { TelemetryService } from 'src/app/services/telemetry.service';

@Component({
  selector: 'app-command-image',
  templateUrl: './command-image.component.html',
  styleUrls: ['./command-image.component.css']
})
export class CommandImageComponent {
  @Input() command: Command | undefined;

  @Output() canceled = new EventEmitter<void>();

  form: FormGroup;
  imageList: string[] = [];
  showNewImage: boolean = false;
  imageToCreate: Image | null = null;

  constructor(
    private devstate: DevstateService,
    private state: StateService,
    private telemetry: TelemetryService
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
    this.telemetry.track("[ui] create image command");
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


  ngOnChanges(changes: SimpleChanges) {
    if (!changes['command']) {
      return;
    }
    const cmd = changes['command'].currentValue;
    if (cmd == undefined) {
      this.form.get('name')?.enable();
    } else {
      this.form.reset();
      this.form.patchValue(cmd);
      this.form.patchValue(cmd.image);
      this.form.get('name')?.disable();
    }
  }

  save() {
    this.telemetry.track("[ui] update image command");
    const subcreate = () => {
      if (this.command == undefined) {
        return;
      }
      const result = this.devstate.updateApplyCommand(this.command.name, this.form.value);
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
        next: (value) => {
          this.state.changeDevfileYaml(value);
          subcreate();
        },
        error: (error) => {
          alert(error.error.message);
        }
      });        
    } else {
      subcreate();
    }
  }
}
