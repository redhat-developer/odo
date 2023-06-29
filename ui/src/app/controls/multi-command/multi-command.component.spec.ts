import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MultiCommandComponent } from './multi-command.component';

describe('MultiCommandComponent', () => {
  let component: MultiCommandComponent;
  let fixture: ComponentFixture<MultiCommandComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ MultiCommandComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(MultiCommandComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
