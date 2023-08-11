import { Component, Input } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

@Component({
  selector: 'app-multi-text',
  templateUrl: './multi-text.component.html',
  styleUrls: ['./multi-text.component.css'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      multi: true,
      useExisting: MultiTextComponent
    }
  ]
})
export class MultiTextComponent implements ControlValueAccessor {

  @Input() label: string = "";
  @Input() addLabel: string = "";
  @Input() title: string = "";

  onChange = (_: string[]) => {};

  texts: string[] = [];

  writeValue(value: any) {
    this.texts = value;
  }

  registerOnChange(onChange: any) {
    this.onChange = onChange;
  }

  registerOnTouched(_: any) {}

  addText() {
    this.texts.push("");
    this.onChange(this.texts);
  }

  onTextChange(i: number, e: Event) {
    const target = e.target as HTMLInputElement;
    this.texts[i] = target.value;
    this.onChange(this.texts);
  }
}
