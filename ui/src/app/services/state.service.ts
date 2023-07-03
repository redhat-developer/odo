import { Injectable } from '@angular/core';

import { BehaviorSubject } from 'rxjs';

import { ResultValue } from './devstate.service';

@Injectable({
  providedIn: 'root'
})
export class StateService {

  private _state = new BehaviorSubject<ResultValue | null>(null);
  public state = this._state.asObservable(); 

  changeDevfileYaml(newValue: ResultValue) {
    this._state.next(newValue);
  }

  getDragAndDropEnabled(): boolean {
    return localStorage.getItem("dragAndDropEnabled") == "true";
  }

  saveDragAndDropEnabled(enabled: boolean) {
    return localStorage.setItem("dragAndDropEnabled", enabled ? "true" : "false");
  }
}
