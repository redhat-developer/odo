import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SelectContainerComponent } from './select-container.component';

describe('SelectContainerComponent', () => {
  let component: SelectContainerComponent;
  let fixture: ComponentFixture<SelectContainerComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ SelectContainerComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(SelectContainerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
