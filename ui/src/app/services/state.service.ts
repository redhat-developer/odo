import { Injectable } from '@angular/core';

import { BehaviorSubject } from 'rxjs';

import { ResultValue } from './wasm-go.service';

@Injectable({
  providedIn: 'root'
})
export class StateService {

  private _state = new BehaviorSubject<ResultValue | null>(null);
  public state = this._state.asObservable(); 

  changeDevfileYaml(newValue: ResultValue) {
    localStorage.setItem("devfile", newValue.content);
    this._state.next(newValue);
  }

  resetDevfile() {
    localStorage.removeItem('devfile');
  }

  getDevfile(): string | null {
    return localStorage.getItem("devfile");
  }

  getDragAndDropEnabled(): boolean {
    return localStorage.getItem("dragAndDropEnabled") == "true";
  }

  saveDragAndDropEnabled(enabled: boolean) {
    return localStorage.setItem("dragAndDropEnabled", enabled ? "true" : "false");
  }
}
