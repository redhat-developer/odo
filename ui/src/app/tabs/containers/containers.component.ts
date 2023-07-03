import { Component, OnInit } from '@angular/core';
import { StateService } from 'src/app/services/state.service';
import { Container, WasmGoService } from 'src/app/services/wasm-go.service';
import { catchError } from 'rxjs/operators'
import { HttpErrorResponse } from '@angular/common/http';
import { throwError } from 'rxjs';

@Component({
  selector: 'app-containers',
  templateUrl: './containers.component.html',
  styleUrls: ['./containers.component.css']
})
export class ContainersComponent implements OnInit {
  
  forceDisplayAdd: boolean = false;
  containers: Container[] | undefined = [];

  constructor(
    private state: StateService,
    private wasm: WasmGoService,
  ) {}

  ngOnInit() {
    const that = this;
    this.state.state.subscribe(async newContent => {
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
      const result = this.wasm.deleteContainer(name);
      if (result.err != '') {
        alert(result.err);
      } else {
        this.state.changeDevfileYaml(result.value);
      }
    }
  }

  onCreated(container: Container) {
    const result = this.wasm.addContainer(container);
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
