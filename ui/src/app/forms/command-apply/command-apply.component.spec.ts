import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CommandApplyComponent } from './command-apply.component';

describe('CommandApplyComponent', () => {
  let component: CommandApplyComponent;
  let fixture: ComponentFixture<CommandApplyComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ CommandApplyComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(CommandApplyComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
