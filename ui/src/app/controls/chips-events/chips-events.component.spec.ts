import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ChipsEventsComponent } from './chips-events.component';

describe('ChipsEventsComponent', () => {
  let component: ChipsEventsComponent;
  let fixture: ComponentFixture<ChipsEventsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ ChipsEventsComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ChipsEventsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
