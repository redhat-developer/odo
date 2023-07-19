import { CdkDragDrop } from '@angular/cdk/drag-drop';
import { Component } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { DevstateService } from 'src/app/services/devstate.service';
import { Command } from 'src/app/api-gen';
import { SegmentService } from 'ngx-segment-analytics';

@Component({
  selector: 'app-commands',
  templateUrl: './commands.component.html',
  styleUrls: ['./commands.component.css']
})
export class CommandsComponent {
  forceDisplayExecForm: boolean = false;
  forceDisplayApplyForm: boolean = false;
  forceDisplayImageForm: boolean = false;
  forceDisplayCompositeForm: boolean = false;
  enableDragAndDrop: boolean;

  commands: Command[] | undefined = [];

  constructor(
    private state: StateService,
    private devstate: DevstateService,
    private segment: SegmentService
  ) {
    this.enableDragAndDrop = this.state.getDragAndDropEnabled();
  }

  ngOnInit() {
    this.state.state.subscribe(async newContent => {
      this.commands = newContent?.commands;
      if (this.commands == null) {
        return
      }
      this.forceDisplayExecForm = false;
      this.forceDisplayApplyForm = false;
      this.forceDisplayImageForm = false;
      this.forceDisplayCompositeForm = false;
    });
  }

  displayExecForm() {
    this.segment.track("[ui] start create exec command");
    this.forceDisplayExecForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  displayApplyForm() {
    this.segment.track("[ui] start create apply command");
    this.forceDisplayApplyForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  displayImageForm() {
    this.segment.track("[ui] start create image command");
    this.forceDisplayImageForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  displayCompositeForm() {
    this.segment.track("[ui] start create composite command");
    this.forceDisplayCompositeForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  undisplayExecForm() {
    this.forceDisplayExecForm = false;
  }

  undisplayApplyForm() {
    this.forceDisplayApplyForm = false;
  }

  undisplayImageForm() {
    this.forceDisplayImageForm = false;
  }

  undisplayCompositeForm() {
    this.forceDisplayCompositeForm = false;
  }

  drop(event: CdkDragDrop<string>) {
    this.moveCommand(
      event.previousContainer.data,
      event.container.data,
      event.previousIndex,
      event.currentIndex,
    );
  }

  moveCommand(previousKind: string, newKind: string, previousIndex: number, newIndex: number) {
    const result = this.devstate.moveCommand(previousKind, newKind, previousIndex, newIndex);
    result.subscribe({
      next: (value) => {
        this.state.changeDevfileYaml(value);
      },
      error: (error) => {
        alert(error.error.message);
      }
    });
  }

  enableDragAndDropChange() {
    this.state.saveDragAndDropEnabled(this.enableDragAndDrop);
  }

  scrollToBottom() {
    window.scrollTo(0,document.body.scrollHeight);
  }
  
}
