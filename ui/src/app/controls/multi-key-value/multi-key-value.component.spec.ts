import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MultiKeyValueComponent } from './multi-key-value.component';

describe('MultiKeyValueComponent', () => {
  let component: MultiKeyValueComponent;
  let fixture: ComponentFixture<MultiKeyValueComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ MultiKeyValueComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MultiKeyValueComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
