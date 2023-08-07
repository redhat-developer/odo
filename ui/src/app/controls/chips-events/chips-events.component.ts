import { Component, ElementRef, EventEmitter, Input, Output, SimpleChanges } from '@angular/core';
import { MatAutocompleteSelectedEvent } from '@angular/material/autocomplete';
import { MatChipInputEvent } from '@angular/material/chips';
import {COMMA, ENTER} from '@angular/cdk/keycodes';
import { Observable, startWith, map } from 'rxjs';
import { FormControl } from '@angular/forms';

@Component({
  selector: 'app-chips-events',
  templateUrl: './chips-events.component.html',
  styleUrls: ['./chips-events.component.css']
})
export class ChipsEventsComponent {
  @Input() commands : string[] = [];
  @Input() allCommands: string[] = [];
  @Output() updated = new EventEmitter<string[]>();

  separatorKeysCodes: number[] = [ENTER, COMMA];
  commandCtrl = new FormControl('');
  filteredCommands = new (Observable<string[]>);
  
  
  constructor(public commandInput :ElementRef<HTMLInputElement>) {}

  ngOnChanges(changes: SimpleChanges) {
    this.filteredCommands = this.commandCtrl.valueChanges.pipe(
      startWith(null),
      map((cmd: string | null) => (cmd ? this._filter(cmd) : this.allCommands.slice())),
    );
  }

  add(event: MatChipInputEvent): void {
    const value = (event.value || '').trim();

    // Add our command
    if (value) {
      this.commands.push(value);
      this.updated.emit(this.commands);
    }

    // Clear the input value
    event.chipInput!.clear();

    this.commandCtrl.setValue(null);
  }

  remove(command: string): void {
    const index = this.commands.indexOf(command);
    
    if (index >= 0) {
      this.commands.splice(index, 1);
      this.updated.emit(this.commands);
    }
  }

  selected(event: MatAutocompleteSelectedEvent): void {
    this.commands.push(event.option.viewValue);
    this.updated.emit(this.commands);
    this.commandInput.nativeElement.value = '';
    this.commandCtrl.setValue(null);
  }

  private _filter(value: string): string[] {
    const filterValue = value.toLowerCase();

    return this.allCommands.filter(cmd => cmd.toLowerCase().includes(filterValue));
  }
}
