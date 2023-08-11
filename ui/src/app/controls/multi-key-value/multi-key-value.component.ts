import { Component, Input } from '@angular/core';
import { NG_VALUE_ACCESSOR } from '@angular/forms';

interface KeyValue {
  name: string;
  value: string;
}

@Component({
  selector: 'app-multi-key-value',
  templateUrl: './multi-key-value.component.html',
  styleUrls: ['./multi-key-value.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: MultiKeyValueComponent
    }
  ]
})
export class MultiKeyValueComponent {

  @Input() dataCyPrefix: string = "";
  @Input() addLabel: string = "";

  onChange = (_: KeyValue[]) => {};

  entries: KeyValue[] = [];

  writeValue(value: KeyValue[]) {
    this.entries = value;
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  addEntry() {
    this.entries.push({name: "", value: ""});
    this.onChange(this.entries);
  }

  onKeyChange(i: number, e: Event) {
    const target = e.target as HTMLInputElement;
    this.entries[i].name = target.value;
    this.onChange(this.entries);
  }

  onValueChange(i: number, e: Event) {
    const target = e.target as HTMLInputElement;
    this.entries[i].value = target.value;
    this.onChange(this.entries);
  }
}
