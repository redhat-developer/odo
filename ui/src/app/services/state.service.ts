import { Injectable } from '@angular/core';

import { BehaviorSubject } from 'rxjs';
import { DevfileContent } from '../api-gen';

@Injectable({
  providedIn: 'root'
})
export class StateService {

  private savedDevfile: string = "";

  private _state = new BehaviorSubject<DevfileContent | null>(null);
  public state = this._state.asObservable(); 

  private _modified = new BehaviorSubject<boolean | null>(null);
  public modified = this._modified.asObservable(); 

  changeDevfileYaml(newValue: DevfileContent, fromDisk: boolean = false) {
    this._state.next(newValue);

    if (fromDisk) {
      this.savedDevfile = newValue.content;
    }
    if (this.savedDevfile == "") {
      this.savedDevfile = newValue.content;
    }
    if (this.savedDevfile == newValue.content) {
      this._modified.next(false);
    } else {
      this._modified.next(true);
    }    
  }

  isUpdated(devfile: string): boolean {
    return devfile != this.savedDevfile;
  }

  getDragAndDropEnabled(): boolean {
    return localStorage.getItem("dragAndDropEnabled") == "true";
  }

  saveDragAndDropEnabled(enabled: boolean) {
    return localStorage.setItem("dragAndDropEnabled", enabled ? "true" : "false");
  }
}
