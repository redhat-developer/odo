import { CdkDragDrop } from '@angular/cdk/drag-drop';
import { Component } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { Command, WasmGoService } from 'src/app/services/wasm-go.service';

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
    private wasm: WasmGoService,
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
    this.forceDisplayExecForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  displayApplyForm() {
    this.forceDisplayApplyForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  displayImageForm() {
    this.forceDisplayImageForm = true;
    setTimeout(() => {
      this.scrollToBottom();      
    }, 0);
  }

  displayCompositeForm() {
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
    const result = this.wasm.moveCommand(previousKind, newKind, previousIndex, newIndex);
    if (result.err != '') {
      alert(result.err);
    } else {
      this.state.changeDevfileYaml(result.value);
    }
  }

  enableDragAndDropChange() {
    this.state.saveDragAndDropEnabled(this.enableDragAndDrop);
  }

  scrollToBottom() {
    window.scrollTo(0,document.body.scrollHeight);
  }
  
}
